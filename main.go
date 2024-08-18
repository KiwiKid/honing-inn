package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
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
	}
}

func main() {
	// Initialize the database connection
	db, err := DBInit()
	if err != nil {
		log.Fatal("ERROR: failed to connect to database:", err)
	}

	// Create a new router
	r := chi.NewRouter()

	// Define routes with method-specific handlers
	r.Get("/shapes", shapeHandler(db))
	r.Get("/shapes/{shapeId:[0-9]+}", specificShapeHandler(db))
	r.Delete("/shapes/{shapeId:[0-9]+}", specificShapeHandler(db))

	r.Post("/shapes", shapeHandler(db))

	r.Post("/homes-rating", createHomeFactorRating(db))
	r.Post("/factors", createFactor(db))
	r.Get("/homes/{homeId:[0-9]+}", singleHomeHandler(db))
	r.Get("/homes", homeHandler(db))
	r.Post("/homes", homeHandler(db)) // Assuming POST for creating a home
	r.Get("/", mapHandler(db))

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

			homes := GetHomes(db)
			shapes := GetShapes(db)
			log.Printf("Homes: %v Shapes: %v", len(homes), len(shapes))

			meta := MapMeta{
				Lat:  -43.53937676715642,
				Lng:  172.55882263183597,
				Zoom: 13,
				Mode: "---",
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

			log.Printf("%+v", shapes)
			mapComp := mapper(meta, homes, shapes)
			mapComp.Render(GetContext(r), w)
			return
		default:
			warn := warning("Method not allowed")
			warn.Render(GetContext(r), w)
			return
		}
	}
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
			}
		case "all":
			{
				shapeTypes := GetShapeTypes(db)
				shapes := GetShapes(db)
				homes := GetHomes(db)

				shapeList := shapeList(shapes, shapeTypes, homes)
				shapeList.Render(GetContext(r), w)

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

			homeForm := homeEditForm(home, "")
			homeForm.Render(GetContext(r), w)
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

			homeForm := homeForm(lat, lng)
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
			notes := r.FormValue("notes")
			url := r.FormValue("url")

			// Convert latitude and longitude to float64
			lat, err := strconv.ParseFloat(latStr, 64)
			if err != nil {
				http.Error(w, "Invalid latitude value", http.StatusBadRequest)
				return
			}

			lng, err := strconv.ParseFloat(lngStr, 64)
			if err != nil {
				http.Error(w, "Invalid longitude value", http.StatusBadRequest)
				return
			}

			// Create a Home object with form data
			home := Home{
				Lat:   lat,
				Lng:   lng,
				Title: title,
				Notes: notes,
				Url:   url,
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

			homeForm := homeEditForm(home, msg)
			homeForm.Render(GetContext(r), w)
			return
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
