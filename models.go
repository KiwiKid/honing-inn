package main

import (
	"gorm.io/gorm"
)

// Factor represents something like "Near bus lines", "Has backyard", etc.
type Factor struct {
	ID           uint   `gorm:"primaryKey"`
	Title        string `json:"title"`
	DisplayOrder int    `json:"display_order"`
}

// Home represents a home with specific attributes.
type Home struct {
	ID        uint    `gorm:"primaryKey"` // Primary key
	Lat       float64 `gorm:"not null"`   // Required field (non-nullable)
	Lng       float64 `gorm:"not null"`   // Required field (non-nullable)
	PointType string  `gorm:"default:null"`
	Title     string  `gorm:"default:null"`
	Url       string  `gorm:"default:null"`
	Notes     string  `gorm:"default:null" form:"notes"` // Optional field (nullable, form tag for form binding)
}

// HomeFactorRating represents a rating for a specific factor of a home.
type HomeFactorRating struct {
	ID       uint `gorm:"primaryKey"`
	Stars    int  `json:"stars" validate:"min=1,max=5"`
	FactorID uint `json:"factor_id"`
	HomeID   uint `json:"home_id"`
}

// Shape represents a custom area that can be added to the map.
type Shape struct {
	ID         uint   `gorm:"primaryKey"`
	ShapeData  string `json:"shape_data"`
	ShapeTitle string `json:"shape_title"`
	ShapeType  string `json:"shape_type"`
	ShapeKind  string `json:"shape_kind"`
}

type ShapeType struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `json:"name"`
}

type ShapeKind struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `json:"name"`
}

type ShapeMeta struct {
	types []ShapeType
	kinds []ShapeKind
}

type MapMeta struct {
	Lat         float64
	Lng         float64
	Zoom        int
	Mode        string
	ProcessMode bool
}

type PointMeta struct {
	types   []PointTypes
	icons   []PointIcons
	factors []Factor
}

type PointTypes struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `json:"name"`
}

type PointIcons struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `json:"name"`
}

type ImageOverlay struct {
	ID        uint    `gorm:"primaryKey"`
	Name      string  `json:"name"`
	FileName  string  `json:"fileName"`
	Bounds    string  `json:"imgBounds"`
	File      string  `json:"fileInput"`
	KeyImage  string  `json:"keyImage"`
	Opacity   float64 `json:"opacity"`
	SourceUrl string  `json:"sourceUrl"`
}

func GetPointMeta(db *gorm.DB) PointMeta {

	allFactors := GetFactors(db)

	return PointMeta{
		types: []PointTypes{
			{ID: 1, Name: "Home"},
			{ID: 2, Name: "RedFlag"},
		},
		icons: []PointIcons{
			{ID: 1, Name: "Home"},
			{ID: 2, Name: "Shape"},
		},
		factors: allFactors,
	}
}
