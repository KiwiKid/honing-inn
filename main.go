package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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

	// Channel to listen for interrupt signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Start the server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on %s: %v\n", portStr, err)
		}
	}()
	log.Printf("Server started on %s", portStr)

	// Block until we receive an interrupt signal
	<-stop
	log.Println("Shutting down server...")

	// Create a context with timeout to allow graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt to gracefully shutdown the server
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}

type EnvConfig struct {
	BaseURL             string
	DBUrl               string
	HuggingFaceAPIToken string
	ImageDir            string
}

func GetEnvConfig() EnvConfig {

	config := EnvConfig{
		BaseURL:             os.Getenv("BASE_URL"),
		DBUrl:               os.Getenv("DATABASE_URL"),
		HuggingFaceAPIToken: os.Getenv("HUGGINGFACE_API_TOKEN"),
		ImageDir:            os.Getenv("IMAGE_DIR"),
	}

	if len(config.DBUrl) == 0 {
		log.Fatalf("DATABASE_URL not set")
	}

	if len(config.ImageDir) == 0 {
		log.Fatalf("IMAGE_DIR not set")
	}

	if len(config.HuggingFaceAPIToken) == 0 {
		log.Printf("HUGGINGFACE_API_TOKEN not set")
	}
	return config
}

const themeId uint = 1

func main() {

	envConfig := GetEnvConfig()

	// Initialize the database connection
	db, err := DBInit(envConfig)
	if err != nil {
		log.Fatal("ERROR: failed to connect to database:", err)
	}
	defer func() {
		sqlDB, err := db.DB()
		if err != nil {
			log.Fatalf("ERROR: could not get database object from Gorm: %v", err)
		}
		if err := sqlDB.Close(); err != nil {
			log.Fatalf("ERROR: failed to close database: %v", err)
		}
		log.Println("Database connection closed")
	}()

	osmClient := NewOSMClient()

	// Create a new router
	r := chi.NewRouter()

	r.Get("/mapmanager", mapManagerHandler(db))
	r.Get("/set-theme", setThemeHandler(db))
	r.Post("/set-theme/{themeId:[0-9]+}", setThemeHandler(db))

	r.Get("/theme", themeEditHandler(db))
	r.Post("/theme", themeEditHandler(db))

	// Define routes with method-specific handlers
	r.Get("/shapes", shapeHandler(db))
	r.Get("/shapes/{shapeId:[0-9]+}", specificShapeHandler(db))
	r.Delete("/shapes/{shapeId:[0-9]+}", specificShapeHandler(db))
	r.Post("/image-overlay/{imageOverlayId:[0-9]+}/key", imageOverlayKeyHandler(db))

	r.Get("/image-overlay", imageOverlayHandler(db, envConfig))
	r.Post("/image-overlay", imageOverlayHandler(db, envConfig))
	r.Delete("/image-overlay/{imageOverlayId:[0-9]+}", imageOverlayHandler(db, envConfig))

	r.Get("/images/{imageID}", imageHandler(envConfig.ImageDir))

	r.Get("/health", healthHandler())

	r.Post("/shapes", shapeHandler(db))

	r.Get("/homes-rating", getHomeFactorRating(db))
	r.Post("/homes-rating", createHomeFactorRating(db))

	r.Post("/factors", createFactor(db))

	r.Get("/chatlist", chatListHandler(db))

	r.Get("/homes/{homeId:[0-9]+}", singleHomeHandler(db))
	r.Delete("/homes/{homeId:[0-9]+}", singleHomeHandler(db))

	r.Get("/homes", homeHandler(db))
	r.Post("/homes", homeHandler(db))
	r.Post("/homes/url", homeUrlHandler(db))

	r.Get("/", mapHandler(db))
	r.Get("/controls", mapControlsHandler(db))
	r.Get("/address", geocodeAddressHandler(osmClient))

	r.Get("/process", mapProcessView(db, envConfig))
	r.Post("/process", mapProcessHandler(db, envConfig))

	r.Get("/delete-all", deleteHandler(db))
	r.Delete("/delete-all", deleteHandler(db))

	r.Get("/factors", factorHandler(db))
	r.Post("/factors", factorHandler(db))
	r.Delete("/factors/{factorId:[0-9]+}", factorHandler(db))

	r.Get("/chat", chatHandler(db))
	r.Post("/chat", chatHandler(db))
	r.Delete("/chat/{chatId:[0-9]+}", chatHandler(db))

	r.Get("/chattype", chatTypeHandler(db))
	r.Post("/chattype", chatTypeHandler(db))
	r.Delete("/chattype/{chatTypeId:[0-9]+}", chatTypeHandler(db))

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

func getThemeID(r *http.Request) (uint, error) {
	if cookie, err := r.Cookie("themeId"); err == nil {
		cookieInt, err := strconv.Atoi(cookie.Value)
		if err == nil {
			return uint(cookieInt), nil
		} else {
			return 0, err
		}

	}
	return 0, errors.New("No theme cookie")
}

func mapControlsHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			themeId, err := getThemeID(r)
			if err != nil {
				w.Header().Add("HX-Redirect", "/set-theme")
			}

			meta := GetPointMeta(db, themeId)

			mapControls := mapControls(meta, "navigate")
			mapControls.Render(GetContext(r), w)
			return
		default:
			warn := warning("Method not allowed")
			warn.Render(GetContext(r), w)
			return
		}
	}
}

func mapManagerHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		themeId, err := getThemeID(r)
		if err != nil {
			w.Header().Add("HX-Redirect", "/set-theme")
		}
		meta := GetPointMeta(db, themeId)

		mapManager := mapManager(meta)
		mapManager.Render(GetContext(r), w)
		return

	}
}

func setThemeHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			{
				allowEditing := false
				themes := GetThemes(db)

				themeId, err := getThemeID(r)
				if err != nil {
					tSet := setThemeContainer(themes, 0, allowEditing)
					tSet.Render(GetContext(r), w)
					return
				} else {
					tSet := setThemeContainer(themes, themeId, allowEditing)
					tSet.Render(GetContext(r), w)
					return
				}
			}
		case "POST":
			{
				themeIdStr := chi.URLParam(r, "themeId")

				// (ensure we're just setting a number)
				themeId, err := strconv.Atoi(themeIdStr)
				if err != nil {
					warning := warning("Invalid theme ID")
					warning.Render(GetContext(r), w)
					return
				}
				cookie := &http.Cookie{
					Name:   "themeId",
					Value:  fmt.Sprintf("%d", themeId),
					Path:   "/",
					MaxAge: 60 * 60 * 24 * 365, // Store for 1 year
				}
				http.SetCookie(w, cookie)

				w.Header().Add("HX-Redirect", "/set-theme")
				return

			}
		}
	}
}

func themeEditHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			if err := r.ParseForm(); err != nil {
				warning := warning("themeEditHandler - Unable to parse form data")
				warning.Render(GetContext(r), w)
				return
			}
			id := r.FormValue("id")
			name := r.FormValue("themeName")

			themeToSave := Theme{
				Name: name,
			}

			if len(id) > 0 {
				themeId, err := strconv.Atoi(id)
				if err != nil {
					warning := warning("Invalid theme ID")
					warning.Render(GetContext(r), w)
					return
				}
				themeToSave.ID = uint(themeId)

			}

			theme, err := SaveTheme(db, themeToSave)
			if err != nil {
				warning := warning(fmt.Sprintf("Failed to save theme - %s", err))
				warning.Render(GetContext(r), w)
				return
			}

			themeEd := editTheme(*theme)
			themeEd.Render(GetContext(r), w)
			return

		case "GET":
			themeIdStr := chi.URLParam(r, "themeId")

			themeId := 0

			if len(themeIdStr) == 0 {
				themeIdRes, err := strconv.Atoi(themeIdStr)
				if err != nil {
					warning := warning("Invalid themeId")
					warning.Render(GetContext(r), w)
					return
				}

				themeId = themeIdRes
			}

			theme := GetActiveTheme(db, uint(themeId))
			themes := GetThemes(db)
			themeEdit := setTheme(themes, theme.ID, true)
			themeEdit.Render(GetContext(r), w)
			return

		}
	}
}

func getThemeIDOrRedirect(w http.ResponseWriter, r *http.Request) (uint, error) {
	themeId, err := getThemeID(r)
	if err != nil {
		w.Header().Add("HX-Redirect", "/set-theme")
		return 0, err
	}
	return themeId, nil
}

func imageHandler(imageDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the image ID from the URL parameter
		imageID := chi.URLParam(r, "imageID")

		// Define the directory where images are stored

		// Construct the full file path using the image ID
		imagePath := filepath.Join(imageDir, fmt.Sprintf("%s.png", imageID))

		// Check if the file exists
		if _, err := os.Stat(imagePath); os.IsNotExist(err) {
			warning := warning(fmt.Sprintf("Image not found - %s", imagePath))
			warning.Render(GetContext(r), w)
			return
		}

		// Serve the file
		http.ServeFile(w, r, imagePath)
	}
}

func imageOverlayKeyHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			if err := r.ParseForm(); err != nil {
				warning := warning("imageOverlayKeyHandler - Unable to parse form data")
				warning.Render(GetContext(r), w)
				return
			}

			fileInput, _, err := r.FormFile("fileInput") // Changed to use FormFile for file upload
			if err != nil {
				warning := warning(fmt.Sprintf("imageOverlayHandler - CREATE  - Unable to parse file input - %+v", err))
				warning.Render(GetContext(r), w)
				return
			}
			defer fileInput.Close()

			var buf bytes.Buffer
			if _, err := io.Copy(&buf, fileInput); err != nil {
				warning := warning("Failed to read file content")
				warning.Render(GetContext(r), w)
				return
			}

			encodedFile := base64.StdEncoding.EncodeToString(buf.Bytes())

			imgId := chi.URLParam(r, "imageOverlayId")
			imgIdInt, err := strconv.Atoi(imgId)
			if err != nil {
				warning := warning(fmt.Sprintf("Invalid image ID - %s", imgId))
				warning.Render(GetContext(r), w)
				return
			}

			img := GetImgOverlay(db, imgIdInt)

			img = ImageOverlay{
				KeyImage: encodedFile,
			}
			msg := "Created new key image overlay"

			result, err := SaveImgOverlay(db, img)
			if err != nil {
				warning := warning(fmt.Sprintf("Failed to save image overlay - %+v", err))
				warning.Render(GetContext(r), w)
				return
			}

			success := imageOverlayEdit(*result, msg)
			success.Render(GetContext(r), w)
		}
	}

}

func chatTypeHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":

			themeId, err := getThemeIDOrRedirect(w, r)
			if err != nil {
				return
			}

			chatTypes, err := GetChatTypes(db, themeId)
			if err != nil {
				warning := warning(fmt.Sprintf("Failed to get chat types - %s", err))
				warning.Render(GetContext(r), w)
				return
			}

			chatTypeList := chatTypeList(chatTypes, ChatMeta{
				SelectedChatID: 0,
				ThemeID:        themeId,
				HomeID:         0,
			})
			chatTypeList.Render(GetContext(r), w)
			return
		case "POST":
			if err := r.ParseForm(); err != nil {
				warning := warning("chatTypeHandler - Unable to parse form data")
				warning.Render(GetContext(r), w)
				return
			}

			chatType := ChatType{
				Name:    r.FormValue("name"),
				Prompt:  r.FormValue("prompt"),
				ThemeID: themeId,
			}

			chatTypeIDStr := r.FormValue("chatTypeID")
			if len(chatTypeIDStr) == 0 {

				chat, createErr := CreateChatType(db, chatType)
				if createErr != nil {
					warning := warning(fmt.Sprintf("Failed to create chat type - %s", createErr))
					warning.Render(GetContext(r), w)
					return
				}

				success := success(fmt.Sprintf("Created chat type - %s", chat.Name))
				success.Render(GetContext(r), w)
				return
			}

			chatTypeID, err := strconv.Atoi(chatTypeIDStr)
			if err != nil {
				warning := warning("Invalid chatTypeID")
				warning.Render(GetContext(r), w)
				return
			}

			chatType.ID = uint(chatTypeID)
			chat, updateErr := UpdateChatType(db, chatType)
			if updateErr != nil {
				warning := warning(fmt.Sprintf("Failed to update chat type - %s", updateErr))
				warning.Render(GetContext(r), w)
				return
			}

			success := success(fmt.Sprintf("Updated chat type - %s", chat.Name))
			success.Render(GetContext(r), w)
			return

		case http.MethodDelete:
			// Delete chat
			chatTypeIdStr := chi.URLParam(r, "chatTypeId")
			if chatTypeIdStr == "" {
				warning := warning("chatTypeId ID not provided")
				warning.Render(GetContext(r), w)
				return
			}

			id, err := strconv.Atoi(chatTypeIdStr)
			if err != nil {
				warning := warning("Invalid chatId ID")
				warning.Render(GetContext(r), w)
				return
			}

			chat, dErr := DeleteChatType(db, uint(id))
			if dErr != nil {
				warning := warning("Failed to delete factor")
				warning.Render(GetContext(r), w)
				return
			}

			reload := success(fmt.Sprintf("Deleted chat type %s", chat.Name))
			reload.Render(GetContext(r), w)
			return
		default:
			warning := warning("Method not allowed")
			warning.Render(GetContext(r), w)
			return

		}
	}
}

func chatListHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Print("chatListHandler START \n\n")
		switch r.Method {
		case "POST":
			if err := r.ParseForm(); err != nil {
				warning := warning("chatListHandler - Unable to parse form data")
				warning.Render(GetContext(r), w)
			}

		case "GET":

			selectedChatIdStr := r.URL.Query().Get("selectedChatId")
			var selectedChatId int
			var err error
			if len(selectedChatIdStr) == 0 {
				selectedChatId = 0
			} else {

				selectedChatId, err = strconv.Atoi(selectedChatIdStr)
				if err != nil {
					warning := warning("Invalid selectedChatId ID")
					warning.Render(GetContext(r), w)
					return
				}
			}

			themeId, err := getThemeIDOrRedirect(w, r)
			if err != nil {
				return
			}

			chatTypes, err := GetChatTypes(db, uint(themeId))
			if err != nil {
				warning := warning(fmt.Sprintf("Failed to get chatTypes - %s", err))
				warning.Render(GetContext(r), w)
				return
			}

			homeIdStr := r.URL.Query().Get("homeId")
			if homeIdStr == "" {
				log.Printf("!!EMPTY CHATS 1 !!")
				chat := emptyChat([]Chat{}, chatTypes, ChatMeta{
					SelectedChatID: uint(selectedChatId),
					ThemeID:        themeId,
					HomeID:         0,
				}, true)
				chat.Render(GetContext(r), w)
				return
			}

			homeId, err := strconv.Atoi(homeIdStr)
			if err != nil {
				warning := warning("Invalid homeIdStr ID")
				warning.Render(GetContext(r), w)
				return
			}

			var meta = ChatMeta{
				ThemeID:        themeId,
				HomeID:         uint(homeId),
				SelectedChatID: uint(selectedChatId),
			}

			var chatTypeId uint64
			chatTypeIDStr := r.URL.Query().Get("chatTypeId")
			if chatTypeIDStr != "" {
				log.Printf("chatListHandler - CHAT TYPE ID PROVIDED %s", chatTypeIDStr)
				chatTypeId, err = strconv.ParseUint(chatTypeIDStr, 10, 32)
				if err != nil {
					warning := warning("Invalid chatTypeId ID")
					warning.Render(GetContext(r), w)
					return
				}

				meta.ChatTypeID = uint(chatTypeId)

			} else {
				log.Printf("chatListHandler - CHAT TYPE ID NOT PROVIDED %s", chatTypeIDStr)

				chatTypeId = 0
			}
			chats, err := GetChats(db, themeId, uint(homeId), uint(chatTypeId))
			if err != nil {
				warning := warning(fmt.Sprintf("Failed to get chats - %s", err))
				warning.Render(GetContext(r), w)
				return
			}

			/*(if len(chats) == 0 {

				log.Printf("No chats found for themeId %d homeId [%d] and chatTypeId [%d]", themeId, homeId, chatTypeId)
				chatTypeId = 0

				ec := emptyChat(chatTypes, ChatMeta{
					SelectedChatID: uint(selectedChatId),
					ThemeID:        themeId,
					ChatTypeID:     uint(chatTypeId),
					HomeID:         uint(homeId),
				}, true)
				ec.Render(GetContext(r), w)
				return
			}*/

			//log.Printf("chatListHandler - chats %v\n\n meta %+v chatTypeId %d ", chats, meta, chatTypeId)
			chatList := chatList(chats, meta, uint(chatTypeId))
			chatList.Render(GetContext(r), w)
			return

		default:
			warning := warning("Method not allowed")
			warning.Render(GetContext(r), w)
			return
		}
	}
}

func imageOverlayHandler(db *gorm.DB, envConfig EnvConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			viewMode := r.URL.Query().Get("viewMode")

			switch viewMode {
			case "controls":
				imageOverLays := GetImgOverlays(db)

				overlay := imageOverlayControls(imageOverLays)
				overlay.Render(GetContext(r), w)
				return
				// not used:
			case "popup":
				overlay := imageOverlayPopup()
				overlay.Render(GetContext(r), w)
				return
			case "edit":
				imageId := r.URL.Query().Get("imageId")
				if imageId == "" {
					warning := warning("Image ID not provided")
					warning.Render(GetContext(r), w)
					return
				}
				imageIdInt, err := strconv.Atoi(imageId)
				if err != nil {
					warning := warning(fmt.Sprintf("Invalid image ID - %s", imageId))
					warning.Render(GetContext(r), w)
					return
				}

				imgOverlay := GetImgOverlay(db, imageIdInt)

				overlay := imageOverlayEdit(imgOverlay, "")
				overlay.Render(GetContext(r), w)
				return
			default:
				warning := warning(fmt.Sprintf("viewMode not allowed (controls, popup, edit) got [%s]", viewMode))
				warning.Render(GetContext(r), w)
			}
		case "POST":

			if err := r.ParseForm(); err != nil {
				warning := warning("imageOverlayHandler - Unable to parse form data")
				warning.Render(GetContext(r), w)
				return
			}

			var img ImageOverlay
			var msg string
			ID := r.FormValue("ID")

			if len(ID) == 0 {
				fileInput, _, err := r.FormFile("fileInput") // Changed to use FormFile for file upload
				if err != nil {
					warning := warning(fmt.Sprintf("imageOverlayHandler - CREATE  - Unable to parse file input - %+v", err))
					warning.Render(GetContext(r), w)
					return
				}
				defer fileInput.Close()

				var buf bytes.Buffer
				if _, err := io.Copy(&buf, fileInput); err != nil {
					warning := warning("Failed to read file content")
					warning.Render(GetContext(r), w)
					return
				}
				encodedFile := base64.StdEncoding.EncodeToString(buf.Bytes())
				imgName := uuid.New().String()
				img = ImageOverlay{
					File:     encodedFile,
					FileName: imgName,
				}

				saveErr := SaveImage(envConfig.ImageDir, buf.Bytes(), imgName)
				if saveErr != nil {
					warning := warning(fmt.Sprintf("Failed to save image overlay file - %+v", saveErr))
					warning.Render(GetContext(r), w)
					return
				}

				msg = fmt.Sprintf("Created new image overlay and file - %s", imgName)
			} else {
				imgBounds := r.FormValue("imgBounds")

				if len(imgBounds) == 0 {
					warning := warning("imgBounds not provided")
					warning.Render(GetContext(r), w)
					return
				}
				name := r.FormValue("imgName")
				if len(name) == 0 {
					warning := warning("name not provided")
					warning.Render(GetContext(r), w)
					return
				}
				imgOpacity := r.FormValue("imgOpacity")
				if len(imgOpacity) == 0 {
					warning := warning("imgOpacity not provided")
					warning.Render(GetContext(r), w)
					return
				}
				imgOpacityFloat, err := strconv.ParseFloat(imgOpacity, 64)
				if err != nil {
					warning := warning("Invalid imgOpacity value")
					warning.Render(GetContext(r), w)
					return
				}

				idInt, err := strconv.Atoi(ID)
				if err != nil {
					warning := warning("Invalid ID value")
					warning.Render(GetContext(r), w)
					return
				}

				img = GetImgOverlay(db, idInt)

				imgSourceUrl := r.FormValue("imgSourceUrl")
				if len(imgSourceUrl) != 0 {
					img.SourceUrl = imgSourceUrl
				}

				img.Bounds = imgBounds
				img.Name = name
				img.Opacity = imgOpacityFloat

				msg = fmt.Sprintf("Updated image overlay %s", img.Name)
			}

			result, err := SaveImgOverlay(db, img)
			if err != nil {
				warning := warning(fmt.Sprintf("Failed to save image overlay - %+v", err))
				warning.Render(GetContext(r), w)
				return
			}

			success := imageOverlayEdit(*result, msg)
			success.Render(GetContext(r), w)
			return
		case "DELETE":
			imageOverlayId := chi.URLParam(r, "imageOverlayId")

			// Convert shapeIdStr to the appropriate type
			imageOverlayIdInt, err := strconv.Atoi(imageOverlayId)
			if err != nil {
				warning := warning(fmt.Sprintf("Invalid image overlay ID - %s", imageOverlayId))
				warning.Render(GetContext(r), w)
				return
			}

			imgOverlay := DeleteImgOverlay(db, int(imageOverlayIdInt))
			if imgOverlay == nil {
				warning := warning(fmt.Sprintf("Failed to delete image overlay - %s", imageOverlayId))
				warning.Render(GetContext(r), w)
				return
			}

			imageDelete := DeleteImage(envConfig.ImageDir, imgOverlay.FileName)
			if imageDelete != nil {
				warning := warning(fmt.Sprintf("Failed to delete image overlay file - %s", imgOverlay.FileName))
				warning.Render(GetContext(r), w)
				return
			}

			w.Header().Add("HX-Refresh", "true")
			success := success(fmt.Sprintf("Deleted image overlay - %s", imgOverlay.Name))
			success.Render(GetContext(r), w)
			return

		default:
			warning := warning("Method not allowed")
			warning.Render(GetContext(r), w)
			return
		}
	}

	// Extract form fields

}

func healthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("💊 Health check  💊  at %s", time.Now().Format(time.RFC3339))
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
			warning := warning(fmt.Sprintf("Error decoding request: %+v", err))
			warning.Render(GetContext(r), w)
			return
		}
		log.Printf("MAP PROCESS HANDLER %+v", req)

		// Pass the image data to the processor
		result, err := classifyImageWithHuggingFace(req, envConfig)
		if err != nil {
			warning := warning(fmt.Sprintf("Error classifying image: %+v", err))
			warning.Render(GetContext(r), w)
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
		log.Printf("Error creating request to Hugging Face API: %v", err)
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+envConfig.HuggingFaceAPIToken)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error creating http client to Hugging Face API: %v", err)

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
					imgOverlays := GetImgOverlays(db)

					shapeList := shapeList(shapes, shapeTypes, homes, imgOverlays)
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

				var resultShape Shape
				id := r.FormValue("ID")
				if id != "" {
					idInt, err := strconv.Atoi(id)
					if err != nil {
						warn := warning("Invalid ID value")
						warn.Render(GetContext(r), w)
						return
					}
					shape.ID = uint(idInt)
					resultShape = UpdateShape(db, shape)
				} else {
					resultShape = CreateShape(db, shape)

				}

				shapeTypes := GetShapeTypes(db)

				w.Header().Set("hx-Refresh", "true")
				areaShape := editShapeForm(resultShape, shapeTypes, "created new area")
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
					http.Error(w, "add-area-point Invalid home ID", http.StatusBadRequest)
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
				warning := warning(fmt.Sprintf("singleHomeHandler Invalid home ID - %s", homeIdStr))
				warning.Render(GetContext(r), w)
				return
			}

			home, err := GetHome(db, uint(homeId))
			if err != nil {
				warning := warning(fmt.Sprintf("Failed to get home - %s", err))
				warning.Render(GetContext(r), w)
				return
			}

			themeId, err := getThemeIDOrRedirect(w, r)
			if err != nil {
				return
			}
			pointMeta := GetPointMeta(db, themeId)

			viewMode := r.URL.Query().Get("viewMode")

			log.Printf("\n\nviewviewviewview mode %s", viewMode)
			switch viewMode {
			case "view":
				log.Printf("===============home - view")

				ratings := GetHomeRatings(db, home.ID)
				for _, rating := range ratings {
					log.Printf("rating %+v", rating)
				}

				homeForm := homeView(*home, "", pointMeta, ratings)
				homeForm.Render(GetContext(r), w)

				chats, err := GetChats(db, 1, home.ID, 0)
				if err != nil {
					warning := warning(fmt.Sprintf("Failed to get chats - %s", err))
					warning.Render(GetContext(r), w)
					return
				}

				meta := ChatMeta{
					SelectedChatID: 0,
					ThemeID:        themeId,
					ChatTypeID:     0,
					HomeID:         home.ID,
				}
				var chatTypeId int
				chatTypeIDStr := r.URL.Query().Get("chatTypeId")
				if chatTypeIDStr != "" {
					log.Printf("CHAT TYPE ID PROVIDED")
					chatTypeId, err = strconv.Atoi(chatTypeIDStr)
					if err != nil {
						warning := warning("Invalid chatTypeId ID")
						warning.Render(GetContext(r), w)
						return
					}
				} else {
					log.Printf("CHAT TYPE ID NOT PROVIDED (%d existing chats)", len(chats))

					chatTypes, err := GetChatTypes(db, uint(themeId))
					if err != nil {
						warning := warning(fmt.Sprintf("Failed to get chatTypes - %s", err))
						warning.Render(GetContext(r), w)
						return
					}

					chatTypeId = 0
					log.Printf("!!EMPTY CHATS 2 !!")

					ec := emptyChat(chats, chatTypes, meta, true)
					ec.Render(GetContext(r), w)
					return
				}

				log.Printf("singleHomeHandler - chats %v\n\n meta %+v chatTypeId %d ", chats, meta, chatTypeId)
				chatlist := chatList(chats, meta, uint(chatTypeId))
				chatlist.Render(GetContext(r), w)
				return
			case "edit":
				log.Printf("=============home - edit")

				ratings := GetHomeRatings(db, home.ID)

				homeForm := homeEditForm(*home, "", pointMeta, ratings)
				homeForm.Render(GetContext(r), w)
				return
			default:
				warn := warning(fmt.Sprintf("viewMode not allowed [%s] (edit,view)", viewMode))
				warn.Render(GetContext(r), w)
				return
			}

		case "DELETE":
			idStr := chi.URLParam(r, "homeId")
			id, err := strconv.Atoi(idStr)
			if err != nil {
				warning := warning(fmt.Sprintf("Invalid home ID - %s", idStr))
				warning.Render(GetContext(r), w)
				return
			}

			var home = Home{ID: uint(id)}

			result := db.Delete(&home)
			if result.Error != nil {
				warning := warning(fmt.Sprintf("Failed to delete home - %s", result.Error.Error()))
				warning.Render(GetContext(r), w)
				return
			}

			success := success("Home deleted")
			success.Render(GetContext(r), w)
			return
		}
	}
}

func getHomeFactorRating(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			homeId := r.URL.Query().Get("homeId")
			homeIdInt, err := strconv.Atoi(homeId)
			if err != nil {
				warning := warning("getHomeFactorRating Invalid home ID")
				warning.Render(GetContext(r), w)

				return
			}

			homeFactorVoteList := GetHomeRatings(db, uint(homeIdInt))

			home, err := GetHome(db, uint(homeIdInt))
			if err != nil {
				warning := warning("Failed to get home")
				warning.Render(GetContext(r), w)
				return
			}

			homeFactorVoteListComp := factorVoteList(homeFactorVoteList, *home, "")
			homeFactorVoteListComp.Render(GetContext(r), w)
			return
		}
	}
}

func homeHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Print("==== Home Handler ==== " + r.Method)
		switch r.Method {
		case "GET":
			viewMode := r.URL.Query().Get("viewMode")
			log.Printf("homeHandler - viewMode %s", viewMode)
			switch viewMode {
			case "list":
				homes := GetHomes(db)
				pointList := pointListTable(homes, "")
				pointList.Render(GetContext(r), w)
				return
			case "view":
			default:
				lat := r.URL.Query().Get("lat")
				lng := r.URL.Query().Get("lng")

				suburb := r.URL.Query().Get("suburb")
				postcode := r.URL.Query().Get("postcode")
				state := r.URL.Query().Get("state")
				country := r.URL.Query().Get("country")
				road := r.URL.Query().Get("road")
				houseNumber := r.URL.Query().Get("houseNumber")
				displayName := r.URL.Query().Get("displayName")

				addressInfo := &AddressInitInfo{
					Suburb:      suburb,
					Postcode:    postcode,
					Road:        road,
					State:       state,
					Country:     country,
					HouseNumber: houseNumber,
					DisplayName: displayName,
					Lat:         lat,
					Lng:         lng,
				}

				themeId, err := getThemeID(r)
				if err != nil {
					w.Header().Add("HX-Redirect", "/set-theme")
				}

				pointMeta := GetPointMeta(db, themeId)

				homeForm := homeForm(pointMeta, *addressInfo, "")
				homeForm.Render(GetContext(r), w)
				return
			}
		case "DELETE":
			idStr := r.URL.Query().Get("id")
			id, err := strconv.Atoi(idStr)
			if err != nil {
				http.Error(w, "homeHandler - DELETE - Invalid home ID", http.StatusBadRequest)
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
				warning := warning(fmt.Sprintf("Unable to parse form data - POST %+v", err))
				warning.Render(GetContext(r), w)
				return
			}

			// Extract form fields
			latStr := r.FormValue("lat")
			lngStr := r.FormValue("lng")
			title := r.FormValue("title")
			pointType := r.FormValue("pointType")
			notes := r.FormValue("notes")
			url := r.FormValue("url")
			imageUrl := r.FormValue("imageUrl")

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

			cAddress := r.FormValue("displayName")
			if len(cAddress) == 0 && len(title) > 0 {
				cAddress = cleanAddress(title)
			}

			cSuburb := r.FormValue("suburb")

			// Create a Home object with form data
			home := Home{
				Lat:          lat,
				Lng:          lng,
				PointType:    pointType,
				Title:        title,
				CleanAddress: cAddress,
				CleanSuburb:  cSuburb,
				Notes:        notes,
				Url:          url,
				ImageUrl:     imageUrl,
				Postcode:     r.FormValue("postcode"),
				State:        r.FormValue("state"),
				Country:      r.FormValue("country"),
				Road:         r.FormValue("road"),
				HouseNumber:  r.FormValue("houseNumber"),
				DisplayName:  r.FormValue("displayName"),
			}

			removeRequestAt := r.FormValue("removeRequestAt")
			if len(removeRequestAt) != 0 {
				removeRequestAtTime, err := time.Parse("2006-01-02", removeRequestAt)
				if err != nil {
					warning := warning(fmt.Sprintf("Invalid removeRequestAt value - %+v", err))
					warning.Render(GetContext(r), w)
					return
				} else {
					home.RemoveRequestAt = removeRequestAtTime
				}
			}

			idStr := r.FormValue("ID")
			if len(idStr) != 0 {
				id, err := strconv.Atoi(idStr)
				if err != nil {
					http.Error(w, "Invalid home ID", http.StatusBadRequest)
					return
				} else {
					home.ID = uint(id)
				}
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

			themeId, err := getThemeID(r)
			if err != nil {
				w.Header().Add("HX-Redirect", "/set-theme")
			}
			pointMeta := GetPointMeta(db, themeId)

			viewMode := r.URL.Query().Get("viewMode")

			switch viewMode {
			case "view":
				if len(idStr) == 0 {
					homes := GetHomes(db)
					pointList := pointListTable(homes, "")

					pointList.Render(GetContext(r), w)
					return
				}

				ratings := GetHomeRatings(db, uint(home.ID))

				log.Printf("===============home - view")
				homeForm := homeView(home, msg, pointMeta, ratings)
				homeForm.Render(GetContext(r), w)
				return
			case "edit":

				log.Printf("=============home - edit")
				ratings := GetHomeRatings(db, home.ID)

				homeForm := homeEditForm(home, msg, pointMeta, ratings)
				homeForm.Render(GetContext(r), w)
				return
			default:

				log.Printf("home - default")
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

func homeUrlHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			if err := r.ParseForm(); err != nil {
				warn := warning("Unable to parse form data")
				warn.Render(GetContext(r), w)
				return
			}

			// Extract form fields
			url := r.FormValue("url")

			if len(url) == 0 {
				urlIn := urlInput("", "", false, "no URL provided")
				urlIn.Render(GetContext(r), w)
				return
			}

			log.Print("homeUrlHandler homeUrlHandler URL: " + url)

			metaTags, err := GetWebMeta(url)
			if err != nil {
				msg := fmt.Sprintf("Failed to get Info (enter address below) %v", err)
				urlfield := urlInput(url, "", false, msg)
				//warn := warningWithDetail("Could not get web meta - enter address and suburb manually", msg)
				urlfield.Render(GetContext(r), w)
				return
			}

			success := populatedMetaFields(metaTags)
			success.Render(GetContext(r), w)
			return
		default:
			warn := warning("Method not allowed")
			warn.Render(GetContext(r), w)
			return
		}
	}
}

// createHomeFactorRating handler with db dependency injection
// createHomeFactorRatingForm handler with db dependency injection
func createHomeFactorRating(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse form data
		if err := r.ParseForm(); err != nil {
			warn := warning("Failed to parse form data")
			warn.Render(GetContext(r), w)
			return
		}

		// Convert form values to the appropriate types
		stars, err := strconv.Atoi(r.FormValue("stars"))
		if err != nil || stars < 1 || stars > 5 {
			warn := warning("Invalid stars value")
			warn.Render(GetContext(r), w)
			return
		}

		factorID, err := strconv.ParseUint(r.FormValue("factorId"), 10, 32)
		if err != nil {
			warn := warning("Invalid factor_id value")
			warn.Render(GetContext(r), w)
			return
		}

		homeID, err := strconv.ParseUint(r.FormValue("homeId"), 10, 32)
		if err != nil {
			warn := warning("Invalid home_id value")
			warn.Render(GetContext(r), w)
			return
		}

		// Create the HomeFactorRating instance
		rating := HomeFactorRating{
			Stars:    stars,
			FactorID: uint(factorID),
			HomeID:   uint(homeID),
		}

		// Save to the database
		result := db.Create(&rating)
		if result.Error != nil {
			warn := warning("Failed to create home factor rating")
			warn.Render(GetContext(r), w)
			return
		}

		homeFactorVoteList := GetHomeRatings(db, uint(homeID))

		home, err := GetHome(db, uint(homeID))
		if err != nil {
			warn := warning("Failed to get home")
			warn.Render(GetContext(r), w)
			return
		}

		homeFactorVoteListComp := factorVoteList(homeFactorVoteList, *home, fmt.Sprintf("Home factor rating created - %d * for %d", stars, factorID))
		homeFactorVoteListComp.Render(GetContext(r), w)
		return
		// Render success message
		//success := success()
		//uccess.Render(GetContext(r), w)
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

func factorHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			factors := GetFactors(db)

			factorList := factorList(factors)
			factorList.Render(GetContext(r), w)
			return
		case "POST":
			if err := r.ParseForm(); err != nil {
				warn := warning("Unable to parse form data")
				warn.Render(GetContext(r), w)
				return
			}

			// Extract form fields
			title := r.FormValue("title")
			displayOrder := r.FormValue("displayOrder")

			id := r.FormValue("ID")
			if len(id) == 0 {
				factor := Factor{
					Title:        title,
					DisplayOrder: 99999,
					ID:           0,
				}

				// Save the Factor object to the database
				result := db.Save(&factor)
				if result.Error != nil {
					warn := warning("Failed to save factor")
					warn.Render(GetContext(r), w)
					return
				}

				success := factorListLoad()
				success.Render(GetContext(r), w)
				return

			}
			intInt, err := strconv.Atoi(id)
			if err != nil {
				warn := warning("Invalid id value")
				warn.Render(GetContext(r), w)
				return
			}
			displayOrderInt, err := strconv.Atoi(displayOrder)
			if err != nil {
				displayOrderInt = 99999
			}

			// Create a Factor object with form data
			factor := Factor{
				Title:        title,
				DisplayOrder: displayOrderInt,
				ID:           uint(intInt),
			}

			// Save the Factor object to the database
			result := db.Save(&factor)
			if result.Error != nil {
				warn := warning("Failed to update factor")
				warn.Render(GetContext(r), w)
				return
			}

			success := success("Factor updated")
			success.Render(GetContext(r), w)
			return
		case "DELETE":
			idStr := chi.URLParam(r, "factorId")
			if idStr == "" {
				warning := warning("Factor ID not provided")
				warning.Render(GetContext(r), w)
				return
			}

			id, err := strconv.Atoi(idStr)
			if err != nil {
				warning := warning("Invalid factor ID")
				warning.Render(GetContext(r), w)
				return
			}

			_, dErr := DeleteFactor(db, uint(id))
			if dErr != nil {
				warning := warning("Failed to delete factor")
				warning.Render(GetContext(r), w)
				return
			}

			success := factorListLoad()
			success.Render(GetContext(r), w)
			return

		default:
			warn := warning("Method not allowed")
			warn.Render(GetContext(r), w)
			return
		}
	}
}
