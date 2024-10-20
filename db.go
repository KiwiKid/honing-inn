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
	err = db.AutoMigrate(&Factor{}, &Home{}, &HomeFactorRating{}, &Shape{}, &ShapeType{}, &ShapeKind{}, &ImageOverlay{}, &ChatType{}, &Chat{}, &ChatResult{})
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
	return &chatType, nil
}

func GetChats(db *gorm.DB, themeId uint, homeId uint, chatTypeId uint) ([]Chat, error) {
	var chats []Chat
	log.Printf("themeId: %v, homeId: %v", themeId, homeId, chatTypeId)
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
