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
	err = db.AutoMigrate(&Factor{}, &Home{}, &HomeFactorRating{}, &Shape{}, &ShapeType{}, &ShapeKind{}, &ImageOverlay{})
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

func GetHome(db *gorm.DB, id int) Home {
	var home Home
	err := db.First(&home, id)
	if err.Error != nil {
		log.Fatal("failed to get home:", err.Error)
	}
	return home
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
		log.Printf("failed to get factors:", err.Error)
		factors = []Factor{}
	}
	return factors
}

func GetImgOverlay(db *gorm.DB, id int) ImageOverlay {
	var overlay ImageOverlay
	err := db.First(&overlay, id)
	if err.Error != nil {
		log.Printf("failed to get overlay:", err.Error)
	}
	return overlay
}

func GetImgOverlays(db *gorm.DB) []ImageOverlay {
	var overlays []ImageOverlay
	err := db.Find(&overlays)
	if err.Error != nil {
		log.Printf("failed to get overlays:", err.Error)
		overlays = []ImageOverlay{}
	}
	return overlays
}

func SaveImgOverlay(db *gorm.DB, overlay ImageOverlay) (*ImageOverlay, error) {
	err := db.Save(&overlay)
	if err.Error != nil {
		log.Printf("failed to save overlay:", err.Error)
		return nil, err.Error
	}
	return &overlay, nil
}

func DeleteImgOverlay(db *gorm.DB, id int) *ImageOverlay {
	overlay := GetImgOverlay(db, id)
	err := db.Delete(&overlay)
	if err.Error != nil {
		log.Printf("failed to delete overlay:", err.Error)
		return nil
	}
	return &overlay
}

func CreateShape(db *gorm.DB, shape Shape) Shape {
	err := db.Create(&shape)
	if err.Error != nil {
		log.Fatal("failed to create shape:", err.Error)
	}
	return shape
}

type HomeFactorAndRating struct {
	HomeFactorRating
	Factor
}

func GetHomeRatings(db *gorm.DB, homeId uint) []HomeFactorAndRating {
	var ratings []HomeFactorRating
	err := db.Where("home_id = ?", homeId).Find(&ratings)
	if err.Error != nil {
		log.Fatal("failed to get ratings:", err.Error)
	}

	ratingIds := make([]uint, len(ratings))
	for i, rating := range ratings {
		ratingIds[i] = rating.FactorID
	}

	var factors []Factor
	err2 := db.Find(&factors, "id IN (?)", ratingIds)
	if err2.Error != nil {
		log.Fatal("failed to get factors:", err2.Error)
	}

	var ratingsWithFactors []HomeFactorAndRating
	for _, rating := range ratings {
		for _, factor := range factors {
			if rating.FactorID == factor.ID {
				ratingsWithFactors = append(ratingsWithFactors, HomeFactorAndRating{
					HomeFactorRating: rating,
					Factor:           factor,
				})
			}
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
