package main

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// DBInit initializes the database and creates the tables
func DBInit(config EnvConfig) (*gorm.DB, error) {

	db, err := gorm.Open(sqlite.Open(config.DBUrl), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database:", err)
	}

	// Migrate the schema
	err = db.AutoMigrate(&Factor{}, &Home{}, &HomeFactorRating{}, &Shape{}, &ShapeType{}, &ShapeKind{}, &ImageOverlay{}, &ChatType{}, &Chat{}, &ChatResult{}, &Theme{}, &FractalSearch{}, &Point{}, &Message{}, &FractalSearchResultGroup{})
	if err != nil {
		log.Fatal("failed to migrate database:", err)
	}

	InitShapeTypes(db)
	InitShapeKinds(db)
	return db, nil
}

func InitShapeTypes(db *gorm.DB) error {
	// delete all existing shapes
	err := db.Exec("DELETE FROM shape_types")
	if err.Error != nil {
		log.Fatal("failed to delete shape types:", err.Error)
	}

	shapes := []ShapeType{
		{
			ID:   1,
			Name: "area",
		},
	}
	for _, shape := range shapes {
		err := db.Where(ShapeType{ID: shape.ID}).FirstOrCreate(&shape).Error
		if err != nil {
			log.Fatal("failed to create shape:", err)
		}
	}
	return nil
}

func InitShapeKinds(db *gorm.DB) error {
	// delete all existing shapes
	err := db.Exec("DELETE FROM shape_kinds")
	if err.Error != nil {
		log.Fatal("failed to delete shape_kinds:", err.Error)
	}
	shapes := []ShapeKind{
		{
			ID:   1,
			Name: "warning",
		}, {
			ID:   2,
			Name: "noGo",
		}, {
			ID:   3,
			Name: "good",
		},
	}
	for _, shape := range shapes {
		err := db.Where(ShapeKind{ID: shape.ID}).FirstOrCreate(&shape).Error
		if err != nil {
			log.Fatal("failed to create shape:", err)
		}
	}
	return nil
}

// Get all Shapes
func GetShapes(db *gorm.DB) []Shape {
	var shapes []Shape
	err := db.Find(&shapes)
	if err.Error != nil {
		log.Fatal("failed to get shapes:", err.Error)
	}
	return shapes
}

func GetShape(db *gorm.DB, shapeId uint) Shape {
	var shape Shape
	err := db.First(&shape, shapeId)
	if err.Error != nil {
		log.Fatal("failed to get shape:", err.Error)
	}
	return shape
}

func DeleteShape(db *gorm.DB, shapeId uint) Shape {
	shape := GetShape(db, shapeId)
	err := db.Delete(&shape)
	if err.Error != nil {
		log.Fatal("failed to delete shape:", err.Error)
	}
	return shape
}

func GetHome(db *gorm.DB, id uint) (*Home, error) {
	var home Home
	err := db.First(&home, id)
	if err.Error != nil {
		return nil, err.Error
	}
	return &home, nil
}

func GetHomes(db *gorm.DB) []Home {
	var homes []Home
	err := db.Find(&homes)
	if err.Error != nil {
		log.Fatal("failed to get homes:", err.Error)
	}
	return homes
}

func GetShapeTypes(db *gorm.DB) ShapeMeta {
	var shapeTypes []ShapeType
	err := db.Find(&shapeTypes)
	if err.Error != nil {
		log.Fatal("failed to get shape types:", err.Error)
	}

	var shapeKinds []ShapeKind
	err2 := db.Find(&shapeKinds)
	if err2.Error != nil {
		log.Fatal("failed to get shape types:", err2.Error)
	}

	return ShapeMeta{
		types: shapeTypes,
		kinds: shapeKinds,
	}
}

func GetFactors(db *gorm.DB) []Factor {
	var factors []Factor
	err := db.Find(&factors)
	if err.Error != nil {
		log.Printf("failed to get factors: %v", err.Error)
		factors = []Factor{}
	}
	return factors
}

func GetActiveTheme(db *gorm.DB, themeIdOverride uint) Theme {
	if themeIdOverride == 0 {
		var theme Theme
		err := db.First(&theme)
		if err.Error != nil {
			log.Printf("failed to get active theme: %v", err.Error)
			return Theme{}
		}
		return theme
	} else {
		var theme Theme
		err := db.First(&theme, themeIdOverride)
		if err.Error != nil {
			log.Printf("failed to get active theme: %v", err.Error)
			return Theme{}
		}
		return theme
	}
}

func SaveTheme(db *gorm.DB, theme Theme) (*Theme, error) {
	err := db.Save(&theme)
	if err.Error != nil {
		log.Printf("failed to save theme: %v", err.Error)
		return nil, err.Error
	}
	return &theme, nil
}

func GetThemes(db *gorm.DB) []Theme {
	var themes []Theme
	err := db.Find(&themes)
	if err.Error != nil {
		log.Printf("failed to get themes: %v", err.Error)
		themes = []Theme{}
	}
	return themes
}

func DeleteFactor(db *gorm.DB, id uint) (Factor, error) {
	var factor Factor
	err := db.First(&factor, id)
	if err.Error != nil {
		return factor, err.Error
	}

	delErr := db.Delete(&factor)
	if delErr != nil {
		log.Printf("failed to delete factor: %v", delErr)
		return factor, delErr.Error
	}
	return factor, nil
}

func GetImgOverlay(db *gorm.DB, id int) ImageOverlay {
	var overlay ImageOverlay
	err := db.First(&overlay, id)
	if err.Error != nil {
		log.Printf("failed to get overlay: %v", err.Error)
	}
	return overlay
}

func GetImgOverlays(db *gorm.DB) []ImageOverlay {
	var overlays []ImageOverlay
	err := db.Find(&overlays)
	if err.Error != nil {
		log.Printf("failed to get overlays: %v", err.Error)
		overlays = []ImageOverlay{}
	}
	return overlays
}

func SaveImgOverlay(db *gorm.DB, overlay ImageOverlay) (*ImageOverlay, error) {
	err := db.Save(&overlay)
	if err.Error != nil {
		log.Printf("failed to save overlay: %v", err.Error)
		return nil, err.Error
	}
	return &overlay, nil
}

func DeleteImgOverlay(db *gorm.DB, id int) *ImageOverlay {
	overlay := GetImgOverlay(db, id)
	err := db.Delete(&overlay)
	if err.Error != nil {
		log.Printf("failed to delete overlay: %v", err.Error)
		return nil
	}
	return &overlay
}

func CreateChatType(db *gorm.DB, chatType ChatType) (*ChatType, error) {
	err := db.Create(&chatType)
	if err.Error != nil {
		return nil, err.Error
	}
	return &chatType, nil
}

func UpdateChatType(db *gorm.DB, chatType ChatType) (*ChatType, error) {
	err := db.Save(&chatType)
	if err.Error != nil {
		return nil, err.Error
	}
	return &chatType, nil
}

func GetChatType(db *gorm.DB, id uint) (*ChatType, error) {
	var chatType ChatType
	err := db.First(&chatType, id)
	if err.Error != nil {
		return nil, err.Error
	}
	return &chatType, nil
}

func DeleteChat(db *gorm.DB, id uint) (*Chat, error) {
	chat := Chat{ID: id}
	err := db.Delete(&chat)
	if err.Error != nil {
		return nil, err.Error
	}
	return &chat, nil
}

func GetChatTypes(db *gorm.DB, themeId uint) ([]ChatType, error) {
	var chatTypes []ChatType
	err := db.Find(&chatTypes).Where("theme_id = ?", themeId)
	if err.Error != nil {
		return nil, err.Error
	}
	return chatTypes, nil
}

func DeleteChatType(db *gorm.DB, id uint) (*ChatType, error) {
	chatType := ChatType{ID: id}
	err := db.Delete(&chatType)
	if err.Error != nil {
		return nil, err.Error
	}
	if err.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &chatType, nil
}

func GetChats(db *gorm.DB, themeId uint, homeId uint, chatTypeId uint) ([]Chat, error) {
	var chats []Chat
	log.Printf("themeId: %v, homeId: %v chatTypeID: %v", themeId, homeId, chatTypeId)
	if chatTypeId == 0 {
		err := db.Preload("Results").Where("theme_id = ? AND home_id = ?", themeId, homeId).Find(&chats).Error
		if err != nil {
			return nil, err
		}
	} else {
		err := db.Preload("Results").Where("theme_id = ? AND home_id = ? AND chat_type = ?", themeId, homeId, chatTypeId).Find(&chats).Error
		if err != nil {
			return nil, err
		}
	}
	return chats, nil
}

func GetChat(db *gorm.DB, id uint) (*Chat, error) {
	var chat Chat
	err := db.Preload("Results").First(&chat, id)
	if err.Error != nil {
		return nil, err.Error
	}
	return &chat, nil
}

func CreateShape(db *gorm.DB, shape Shape) Shape {
	err := db.Create(&shape)
	if err.Error != nil {
		log.Fatal("failed to create shape:", err.Error)
	}
	return shape
}

func UpdateShape(db *gorm.DB, shape Shape) Shape {
	err := db.Save(&shape)
	if err.Error != nil {
		log.Fatal("failed to update shape:", err.Error)
	}
	return shape
}

type HomeFactorAndRating struct {
	*HomeFactorRating
	Factor
}

func GetHomeRatings(db *gorm.DB, homeId uint) []HomeFactorAndRating {

	factors := GetFactors(db)

	var ratings []HomeFactorRating
	err := db.Where("home_id = ?", homeId).Find(&ratings)
	if err.Error != nil {
		log.Fatal("failed to get ratings:", err.Error)
	}

	ratingMap := make(map[uint]HomeFactorRating)
	for _, rating := range ratings {
		ratingMap[rating.FactorID] = rating
	}

	var ratingsWithFactors []HomeFactorAndRating
	for _, factor := range factors {
		if rating, exists := ratingMap[factor.ID]; exists {
			// Factor has a rating
			ratingsWithFactors = append(ratingsWithFactors, HomeFactorAndRating{
				HomeFactorRating: &rating,
				Factor:           factor,
			})
		} else {
			// Factor does not have a rating
			ratingsWithFactors = append(ratingsWithFactors, HomeFactorAndRating{
				HomeFactorRating: nil,
				Factor:           factor,
			})
		}
	}
	return ratingsWithFactors
}

func DeleteAll(db *gorm.DB) {
	db.Exec("DELETE FROM shapes")
	db.Exec("DELETE FROM shape_types")
	db.Exec("DELETE FROM shape_kinds")
	db.Exec("DELETE FROM homes")
	db.Exec("DELETE FROM factors")
	db.Exec("DELETE FROM home_factor_ratings")
}

func CreateFractalSearch(db *gorm.DB, search FractalSearch) (*FractalSearch, error) {
	err := db.Create(&search)
	if err.Error != nil {
		return nil, err.Error
	}
	return &search, nil
}

func GetFractalSearch(db *gorm.DB, id uint) (*FractalSearch, error) {
	var search FractalSearch
	err := db.First(&search, id)
	if err.Error != nil {
		return nil, err.Error
	}
	return &search, nil
}

func GetFractalSearchFull(db *gorm.DB, id uint) (*FractalSearchFull, error) {
	var search FractalSearchFull
	fs, err := GetFractalSearch(db, id)
	if err != nil {
		return nil, err
	}

	search.FractalSearch = *fs

	points, err := GetPoints(db, id)
	if err != nil {
		return nil, err
	}
	search.Points = points

	messages, err := GetMessages(db, id)
	if err != nil {
		return nil, err
	}
	search.Messages = messages

	return &search, nil
}

func GetFractalSearches(db *gorm.DB, status string) ([]FractalSearch, error) {
	var searches []FractalSearch
	var err *gorm.DB
	if len(status) > 0 {
		err = db.Find(&searches).Where("status = ?", status)
	} else {
		err = db.Find(&searches)
	}

	if err.Error != nil {
		return nil, err.Error
	}
	return searches, nil
}

func CreatePoint(db *gorm.DB, point Point) (*Point, error) {
	err := db.Create(&point)
	if err.Error != nil {
		return nil, err.Error
	}
	return &point, nil
}

func CreateMessage(db *gorm.DB, message Message) (*Message, error) {
	err := db.Create(&message)
	if err.Error != nil {
		return nil, err.Error
	}
	return &message, nil
}

func GetPoints(db *gorm.DB, fsId uint) ([]Point, error) {
	var points []Point
	err := db.Where("fractal_search_id = ?", fsId).Find(&points).Error
	if err != nil {
		return nil, err
	}
	log.Printf("GetPoints: %d", fsId)
	log.Printf("GetPoints: %v", points)
	return points, nil
}

func CreateFractalSearchResultGroup(db *gorm.DB, resultGroup FractalSearchResultGroup) (*FractalSearchResultGroup, error) {
	err := db.Create(&resultGroup)
	if err.Error != nil {
		return nil, err.Error
	}
	return &resultGroup, nil
}

func GetMessages(db *gorm.DB, fsId uint) ([]Message, error) {
	var messages []Message
	err := db.Where("fractal_search_id = ?", fsId).Find(&messages).Error
	if err != nil {
		return nil, err
	}
	return messages, nil
}

func DeletePoints(db *gorm.DB, id uint) error {
	err := db.Exec("DELETE FROM points WHERE fractal_search_id = ?", id)
	if err.Error != nil {
		return err.Error
	}
	return nil
}

func DeleteMessages(db *gorm.DB, id uint) error {
	err := db.Exec("DELETE FROM messages WHERE fractal_search_id = ?", id)
	if err.Error != nil {
		return err.Error
	}
	return nil
}

func UpdatePoint(db *gorm.DB, point Point) (*Point, error) {
	err := db.Save(&point)
	if err.Error != nil {
		return nil, err.Error
	}
	return &point, nil
}

func GetPoint(db *gorm.DB, id uint) (*Point, error) {
	var point Point
	err := db.First(&point, id)
	if err.Error != nil {
		return nil, err.Error
	}
	return &point, nil
}

func DeleteFractalSearch(db *gorm.DB, id uint) (*FractalSearch, error) {
	search := FractalSearch{ID: id}
	sdRes := db.Delete(&search)
	if sdRes.Error != nil {
		return nil, sdRes.Error
	}

	err := DeletePoints(db, id)
	if err != nil {
		return nil, err
	}

	err = DeleteMessages(db, id)
	if err != nil {
		return nil, err
	}

	return &search, nil
}
