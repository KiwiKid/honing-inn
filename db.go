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
	err = db.AutoMigrate(&Factor{}, &Home{}, &HomeFactorRating{}, &Shape{}, &ShapeType{}, &ShapeKind{})
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
			Name: "no-go",
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

func CreateShape(db *gorm.DB, shape Shape) Shape {
	err := db.Create(&shape)
	if err.Error != nil {
		log.Fatal("failed to create shape:", err.Error)
	}
	return shape
}

func DeleteAll(db *gorm.DB) {
	db.Exec("DELETE FROM shapes")
	db.Exec("DELETE FROM shape_types")
	db.Exec("DELETE FROM shape_kinds")
	db.Exec("DELETE FROM homes")
	db.Exec("DELETE FROM factors")
	db.Exec("DELETE FROM home_factor_ratings")
}
