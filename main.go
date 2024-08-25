package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/nfnt/resize"
	"gorm.io/gorm"
)

// startServer starts the HTTP server with graceful shutdown.

// startServer starts the HTTP server with graceful shutdown.
func startServer(r *chi.Mux, portStr string) {
	server := &http.Server{
		Addr:    portStr,
		Handler: r,
	}

	// Start the server in a goroutine
	log.Printf("Starting server on %s", portStr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not listen on %s: %v\n", portStr, err)
	} else {
		log.Printf("Server started on %s", portStr)
	}
}

type EnvConfig struct {
	BaseURL             string
	DBUrl               string
	HuggingFaceAPIToken string
}

func GetEnvConfig() EnvConfig {

	config := EnvConfig{
		BaseURL:             os.Getenv("BASE_URL"),
		DBUrl:               os.Getenv("DATABASE_URL"),
		HuggingFaceAPIToken: os.Getenv("HUGGINGFACE_API_TOKEN"),
	}

	if len(config.DBUrl) == 0 {
		log.Fatalf("DATABASE_URL not set")
	}

	if len(config.HuggingFaceAPIToken) == 0 {
		log.Printf("HUGGINGFACE_API_TOKEN not set")
	}
	return config
}

func main() {

	envConfig := GetEnvConfig()

	// Initialize the database connection
	db, err := DBInit(envConfig)
	if err != nil {
		log.Fatal("ERROR: failed to connect to database:", err)
	}

	// Create a new router
	r := chi.NewRouter()

	// Define routes with method-specific handlers
	r.Get("/shapes", shapeHandler(db))
	r.Get("/shapes/{shapeId:[0-9]+}", specificShapeHandler(db))
	r.Delete("/shapes/{shapeId:[0-9]+}", specificShapeHandler(db))

	r.Get("/health", healthHandler())

	r.Post("/shapes", shapeHandler(db))

	r.Post("/homes-rating", createHomeFactorRating(db))
	r.Post("/factors", createFactor(db))
	r.Get("/homes/{homeId:[0-9]+}", singleHomeHandler(db))
	r.Delete("/homes/{homeId:[0-9]+}", singleHomeHandler(db))

	r.Get("/homes", homeHandler(db))
	r.Post("/homes", homeHandler(db)) // Assuming POST for creating a home
	r.Get("/", mapHandler(db))

	r.Get("/process", mapProcessView(db, envConfig))
	r.Post("/process", mapProcessHandler(db, envConfig))

	r.Get("/delete-all", deleteHandler(db))
	r.Delete("/delete-all", deleteHandler(db))

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = ":8080"
	} else {
		port = ":" + port
	}
	portStr := fmt.Sprintf(":%s", port)
	log.Printf("Using port (%s)", portStr)

	// Start the HTTP server
	startServer(r, port)

}

func mapHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			meta := MapMeta{
				Lat:         -43.53937676715642,
				Lng:         172.55882263183597,
				Zoom:        13,
				Mode:        "---",
				ProcessMode: false,
			}

			lat := r.URL.Query().Get("lat")
			lng := r.URL.Query().Get("lng")
			if lat != "" && lng != "" {
				meta.Lat, _ = strconv.ParseFloat(lat, 64)
				meta.Lng, _ = strconv.ParseFloat(lng, 64)
			}

			zoom := r.URL.Query().Get("zoom")
			if zoom != "" {
				meta.Zoom, _ = strconv.Atoi(zoom)
			}

			mode := r.URL.Query().Get("mode")
			if mode != "" {
				meta.Mode = mode
			}

			mapComp := mapper(meta)
			mapComp.Render(GetContext(r), w)
			log.Printf("home render done")

			return
		default:
			log.Printf("home render WARNING")

			warn := warning("Method not allowed")
			warn.Render(GetContext(r), w)
			return
		}

	}

}

func healthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		success := success("Healthy")
		success.Render(GetContext(r), w)
	}
}

type MapProcessRequest struct {
	StartPoint [2]float64 `json:"start_point"` // [lat, lng]
	Zoom       int        `json:"zoom"`
	Width      int        `json:"width"`
	Height     int        `json:"height"`
	GridHeight int        `json:"grid_height"`
	GridWidth  int        `json:"grid_width"`
}

type MapProcessingMeta struct {
	GridHeight int
	GridWidth  int
}

func mapProcessView(db *gorm.DB, envConfig EnvConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		meta := MapMeta{
			Lat:         -43.53937676715642,
			Lng:         172.55882263183597,
			Zoom:        16,
			Mode:        "---",
			ProcessMode: true,
		}

		lat := r.URL.Query().Get("lat")
		lng := r.URL.Query().Get("lng")
		if lat != "" && lng != "" {
			meta.Lat, _ = strconv.ParseFloat(lat, 64)
			meta.Lng, _ = strconv.ParseFloat(lng, 64)
		}

		zoom := r.URL.Query().Get("zoom")
		if zoom != "" {
			meta.Zoom, _ = strconv.Atoi(zoom)
		}

		mode := r.URL.Query().Get("mode")
		if mode != "" {
			meta.Mode = mode
		}

		processMeta := MapProcessingMeta{
			GridHeight: 10,
			GridWidth:  10,
		}

		gridHeight := r.URL.Query().Get("grid_height")
		if gridHeight != "" {
			processMeta.GridHeight, _ = strconv.Atoi(gridHeight)
		}

		gridWidth := r.URL.Query().Get("grid_width")
		if gridWidth != "" {
			processMeta.GridWidth, _ = strconv.Atoi(gridWidth)
		}

		mapComp := mapperProcessView(meta, processMeta)
		mapComp.Render(GetContext(r), w)

	}
}

type MapRequest struct {
	CurrentRow int    `json:current_row"`
	CurrentCol int    `json:current_col"`
	ImageData  string `json:"image_data"` // Changed to string for Base64 encoded data
}

func mapProcessHandler(db *gorm.DB, envConfig EnvConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req MapRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Printf("MAP PROCESS HANDLER %+v", req)

		// Pass the image data to the processor
		result, err := classifyImageWithHuggingFace(req, envConfig)
		if err != nil {
			warning := fmt.Sprintf("Error classifying image: %+v", err)
			log.Printf(warning)

			http.Error(w, warning, http.StatusInternalServerError)
			return
		}

		// Log the save action (dummy save)
		log.Print(fmt.Sprintf("FAKE FAKE FAKE MAP IMAGE saved %+v", result))

		http.ResponseWriter.Header(w).Set("Content-Type", "application/json")
		http.ResponseWriter.Write(w, result)

		return
		// Send the classification result back to the client
		//	w.Header().Set("Content-Type", "application/json")
		//	w.Write(result)

	}
}

func classifyImageWithHuggingFace(request MapRequest, envConfig EnvConfig) ([]byte, error) {
	// Prepare the multipart form request with the image data
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "image.png")
	if err != nil {
		return nil, err
	}

	// Write image data to the form file
	_, err = io.Copy(part, bytes.NewReader([]byte(request.ImageData)))
	if err != nil {
		return nil, err
	}
	writer.Close()

	const HuggingFaceAPIURL = "https://api-inference.huggingface.co/models/{model_name}"
	req, err := http.NewRequest("POST", HuggingFaceAPIURL, &body)
	if err != nil {
		log.Printf("Error creating request to Hugging Face API: ", err)
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+envConfig.HuggingFaceAPIToken)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error creating http client to Hugging Face API: ", err)

		return nil, err
	}
	defer resp.Body.Close()

	// Read and return the response from Hugging Face API
	return io.ReadAll(resp.Body)
}

func deleteHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			shapes := GetShapes(db)

			homes := GetHomes(db)

			deleteForm := deleteAllForm(shapes, homes)
			deleteForm.Render(GetContext(r), w)
			return
		case "DELETE":
			DeleteAll(db)

			success := success("Deleted all shapes and homes")
			success.Render(GetContext(r), w)
			return
		}
	}
}

func specificShapeHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			shapeIdStr := chi.URLParam(r, "shapeId")

			// Convert shapeIdStr to the appropriate type
			shapeId, err := strconv.Atoi(shapeIdStr)
			if err != nil {
				http.Error(w, "Invalid shapeId", http.StatusBadRequest)
				return
			}

			shape := GetShape(db, uint(shapeId))

			shapeTypes := GetShapeTypes(db)

			areaShape := editShapeForm(shape, shapeTypes, "")
			areaShape.Render(GetContext(r), w)

			return
		case "DELETE":
			shapeIdStr := chi.URLParam(r, "shapeId")

			// Convert shapeIdStr to the appropriate type
			shapeId, err := strconv.Atoi(shapeIdStr)
			if err != nil {
				http.Error(w, "Invalid shapeId", http.StatusBadRequest)
				return
			}

			shape := DeleteShape(db, uint(shapeId))

			warning := refreshButton("Refresh", fmt.Sprintf("Deleted shape %d (%s)", shape.ID, shape.ShapeTitle))
			warning.Render(GetContext(r), w)

		}
	}
}

func shapeHandler(db *gorm.DB) http.HandlerFunc {
	log.Printf("shapeHandler")

	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("shapeHandler")

		switch r.Method {
		case "GET":
			mode := r.URL.Query().Get("mode")

			switch mode {
			case "area":
				shapeTypes := GetShapeTypes(db)
				areaShape := addShapeForm(shapeTypes)
				areaShape.Render(GetContext(r), w)
				return
			case "all":
				{
					shapeTypes := GetShapeTypes(db)
					shapes := GetShapes(db)
					homes := GetHomes(db)

					shapeList := shapeList(shapes, shapeTypes, homes)
					shapeList.Render(GetContext(r), w)
					return
				}
			default:
				warn := warning("mode not allowed for shapeHandler")
				warn.Render(GetContext(r), w)
				return
			}

			shapes := GetShapes(db)
			shapeTypes := GetShapeTypes(db)

			log.Printf("Shapes: %v ShapeTypes: %v", shapes, shapeTypes)

			mapComp := shapePopupManage(shapes, shapeTypes.types)
			mapComp.Render(GetContext(r), w)
			return
		case "POST":
			log.Printf("shapeHandler POST request received")
			if err := r.ParseForm(); err != nil {
				warn := warning("MUnable to parse form data")
				warn.Render(GetContext(r), w)
				return
			}

			updateMode := r.URL.Query().Get("updateMode")
			switch updateMode {
			case "create-area":
				shape := Shape{}

				shapeData := r.FormValue("shapeData")
				if shapeData == "" {
					warn := warning("No latlngs provided for create-area")
					warn.Render(GetContext(r), w)
					return
				} else {
					shape.ShapeData = shapeData
				}

				shapeTitle := r.FormValue("shapeTitle")
				if shapeTitle != "" {
					shape.ShapeTitle = shapeTitle
				}

				shapeType := r.FormValue("shapeType")
				if shapeType == "" {
					warn := warning("Shape type not provided for create-area")
					warn.Render(GetContext(r), w)
					return
				} else {
					shape.ShapeType = shapeType
				}

				shapeKind := r.FormValue("shapeKind")
				if shapeKind == "" {
					warn := warning("Shape kind not provided for create-area")
					warn.Render(GetContext(r), w)
					return
				} else {
					shape.ShapeKind = shapeKind
				}

				createdShape := CreateShape(db, shape)

				log.Printf("shapeHandler CreateShape \n\n%+v", createdShape)

				shapeTypes := GetShapeTypes(db)

				areaShape := editShapeForm(shape, shapeTypes, "created new area")
				areaShape.Render(GetContext(r), w)
				return

			case "add-area-point":
				idStr := r.FormValue("ID")
				if idStr == "" {
					warn := warning("Shape ID not provided for add-area-point")
					warn.Render(GetContext(r), w)
					return
				}
				id, err := strconv.Atoi(idStr)
				if err != nil {
					http.Error(w, "Invalid home ID", http.StatusBadRequest)
					return
				}

				shape := GetShape(db, uint(id))

				shapeData := r.FormValue("shapeData")
				if shapeData != "" {
					shape.ShapeData = shapeData
				}
				shapeTitle := r.FormValue("shapeTitle")
				if shapeTitle != "" {
					shape.ShapeTitle = shapeTitle
				}
				shape.ShapeTitle = shapeTitle
				shapeType := r.FormValue("shapeType")
				if shapeType != "" {
					shape.ShapeType = shapeType
				}

				shapeKind := r.FormValue("shapeKind")
				if shapeKind != "" {
					shape.ShapeKind = shapeKind
				}

				db.Save(&shape)

				areaShape := areaShape(shape)
				areaShape.Render(GetContext(r), w)
				return
			default:
				warn := warning(fmt.Sprintf("updateMode not allowed [%s]", updateMode))
				warn.Render(GetContext(r), w)
				return
			}

		default:
			warn := warning("Method not allowed")
			warn.Render(GetContext(r), w)
			return
		}
	}
}

// createFactor handler with db dependency injection
func createFactor(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var factor Factor
		if err := json.NewDecoder(r.Body).Decode(&factor); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		db.Create(&factor)
		json.NewEncoder(w).Encode(factor)
	}
}

func singleHomeHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			log.Printf("GET  singleHomeHandler request received")
			homeIdStr := chi.URLParam(r, "homeId")

			// Convert homeIdStr to the appropriate type
			homeId, err := strconv.Atoi(homeIdStr)
			if err != nil {
				http.Error(w, "Invalid homeId", http.StatusBadRequest)
				return
			}

			home := GetHome(db, homeId)

			pointMeta := GetPointMeta()

			viewMode := r.URL.Query().Get("viewMode")

			log.Printf("\n\nviewviewviewview mode %s", viewMode)
			switch viewMode {
			case "view":
				homeForm := homeView(home, "", pointMeta)
				homeForm.Render(GetContext(r), w)
				return
			case "edit":
				homeForm := homeEditForm(home, "", pointMeta)
				homeForm.Render(GetContext(r), w)
				return
			default:
				warn := warning("viewMode not allowed (edit,view)")
				warn.Render(GetContext(r), w)
				return
			}

			return
		case "DELETE":
			idStr := chi.URLParam(r, "homeId")
			id, err := strconv.Atoi(idStr)
			if err != nil {
				http.Error(w, "Invalid home ID", http.StatusBadRequest)
				return
			}

			var home = Home{ID: uint(id)}

			result := db.Delete(&home)
			if result.Error != nil {
				http.Error(w, "Failed to delete home", http.StatusInternalServerError)
				return
			}

			success := success("Home deleted")
			success.Render(GetContext(r), w)
			return
		}
	}
}

// createHome handler with db dependency injection
func homeHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			lat := r.URL.Query().Get("lat")
			lng := r.URL.Query().Get("lng")

			pointMeta := GetPointMeta()

			homeForm := homeForm(pointMeta, lat, lng, "")
			homeForm.Render(GetContext(r), w)
			return

		case "DELETE":
			idStr := r.URL.Query().Get("id")
			id, err := strconv.Atoi(idStr)
			if err != nil {
				http.Error(w, "Invalid home ID", http.StatusBadRequest)
				return
			}

			var home = Home{ID: uint(id)}

			result := db.Delete(&home)
			if result.Error != nil {
				http.Error(w, "Failed to delete home", http.StatusInternalServerError)
				return
			}

			success := success("Home deleted")
			success.Render(GetContext(r), w)
			return

		case "POST":
			log.Printf("POST request received")
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Unable to parse form data", http.StatusBadRequest)
				return
			}

			// Extract form fields
			latStr := r.FormValue("lat")
			lngStr := r.FormValue("lng")
			title := r.FormValue("title")
			pointType := r.FormValue("pointType")
			notes := r.FormValue("notes")
			url := r.FormValue("url")

			// Convert latitude and longitude to float64
			lat, err := strconv.ParseFloat(latStr, 64)
			if err != nil {
				warning := warning("Invalid latitude value")
				warning.Render(GetContext(r), w)
				return
			}

			lng, err := strconv.ParseFloat(lngStr, 64)
			if err != nil {
				warning := warning("Invalid longitude value")
				warning.Render(GetContext(r), w)
				return
			}

			// Create a Home object with form data
			home := Home{
				Lat:       lat,
				Lng:       lng,
				PointType: pointType,
				Title:     title,
				Notes:     notes,
				Url:       url,
			}

			var msg string
			result := db.Save(&home)
			if result.Error != nil {
				msg = "Failed to save home: " + result.Error.Error()
			} else if len(title) == 0 {
				msg = "Created Point"
			} else {
				msg = "Updated Point"
			}

			pointMeta := GetPointMeta()

			viewMode := r.URL.Query().Get("viewMode")

			log.Printf("\n\nviewviewviewview mode %s", viewMode)
			switch viewMode {
			case "view":
				homeForm := homeView(home, msg, pointMeta)
				homeForm.Render(GetContext(r), w)
				return
			case "edit":
				homeForm := homeEditForm(home, msg, pointMeta)
				homeForm.Render(GetContext(r), w)
				return
			default:
				warn := warning("viewMode not allowed (edit,view)")
				warn.Render(GetContext(r), w)
				return
			}

		default:
			warn := warning("Method not allowed")
			warn.Render(GetContext(r), w)
			return
		}
	}
}

// createHomeFactorRating handler with db dependency injection
func createHomeFactorRating(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var rating HomeFactorRating
		if err := json.NewDecoder(r.Body).Decode(&rating); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		db.Create(&rating)
		json.NewEncoder(w).Encode(rating)
	}
}

func parseQueryFloat(value string) float64 {
	if value == "" {
		return 0.0
	}
	parsedValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0.0
	}
	return parsedValue
}

func parseQueryInt(value string) int {
	if value == "" {
		return 0
	}
	parsedValue, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return parsedValue
}

func processMap(startPoint [2]float64, zoom, width, height int) image.Image {
	// Logic for initializing the map (using Leaflet in JS frontend)
	// This is a placeholder where the JS frontend will navigate the map.

	// Create a dummy map image for the example (replace with real map generation logic)
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Optionally resize the image to match the desired dimensions
	resizedImg := resize.Resize(uint(width), uint(height), img, resize.Lanczos3)

	return resizedImg
}
