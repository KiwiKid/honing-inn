package main

import (
	"time"

	"github.com/a-h/templ"
	"gorm.io/gorm"
)

// Factor represents something like "Near bus lines", "Has backyard", etc.
type Factor struct {
	ID           uint   `gorm:"primaryKey"`
	Title        string `json:"title"`
	DisplayOrder int    `json:"display_order"`
}

type Theme struct {
	ID                   uint   `gorm:"primaryKey"`
	Name                 string `json:"name"`
	Description          string `json:"description"`
	StartSystemPrompt    string `json:"start_system_prompt"`
	StartGeoSystemPrompt string `json:"start_geo_system_prompt"`
}

// Home represents a home with specific attributes.
type Home struct {
	ID              uint      `gorm:"primaryKey"`
	Lat             float64   `gorm:"not null"`
	Lng             float64   `gorm:"not null"`
	PointType       string    `gorm:"default:null"`
	Title           string    `gorm:"default:null"`
	Url             string    `gorm:"default:null"`
	CleanAddress    string    `gorm:"default:null"`
	CleanSuburb     string    `gorm:"default:null"`
	ImageUrl        string    `gorm:"default:null"`
	Notes           string    `gorm:"default:null" form:"notes"`
	RemoveRequestAt time.Time `gorm:"default:null"`
	Postcode        string
	State           string
	Country         string
	Road            string
	HouseNumber     string
	DisplayName     string
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
	types       []PointTypes
	icons       []PointIcons
	factors     []Factor
	actionModes []ActionMode
	theme       Theme
}

type AddressInitInfo struct {
	SearchTerm  string
	Lat         string
	Lng         string
	Suburb      string
	Road        string
	Postcode    string
	State       string
	Country     string
	HouseNumber string
	DisplayName string
}

type FractalAISearchInitInfo struct {
	SearchTerm  string
	DisplayName string
	BoundingBox []float64
	PlaceId     string
	Country     string
	AddressType string
	Southwest   string
	Southeast   string
	Northeast   string
	Northwest   string
}

type FractalSearch struct {
	ID          uint   `gorm:"primaryKey"`
	ThemeID     uint   `json:"theme_id"`
	DisplayName string `json:"display_name"`
	Country     string `json:"country"`
	PlaceId     string `json:"place_id"`
	Query       string `json:"query"`
	Status      string `json:"status"`
}

type FractalSearchFull struct {
	FractalSearch
	Points   []Point
	Messages []Message
}

type Point struct {
	ID                         uint    `gorm:"primaryKey"`
	Title                      string  `gorm:"not null"`
	Description                string  `gorm:"default:null"`
	Lat                        float64 `gorm:"default:null"`
	Lng                        float64 `gorm:"default:null"`
	ThemeID                    uint    `gorm:"default:null"`
	FractalSearchID            uint    `gorm:"default:null"`
	FractalSearchResultGroupID uint    `gorm:"default:null"`
	PointType                  string  `gorm:"default:null"`
	Url                        string  `gorm:"default:null"`
	CleanAddress               string  `gorm:"default:null"`
	WarningMessage             string  `gorm:"default:null"`
}

type FractalSearchResultGroup struct {
	ID              uint   `gorm:"primaryKey"`
	FractalSearchID uint   `json:"fractal_search_id"`
	DisplayName     string `json:"display_name"`
	PointTypeName   string `json:"point_type_name"`
}

type Message struct {
	FractalSearchID uint   `json:"fractal_search_id"`
	Role            string `json:"role"`
	Content         string `json:"content"`
}

type ActionMode struct {
	ID        uint
	Key       string
	Name      string
	Details   templ.Component
	FullPanel bool
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

type ChatType struct {
	ID                        uint   `gorm:"primaryKey"`
	Name                      string `json:"name"`
	Prompt                    string `json:"prompt"`
	ThemeID                   uint   `json:"theme_id"`
	AddressType               string `json:"address_type"`
	StartSystemPromptOverride string `json:"start_system_prompt_override"`
}

type Chat struct {
	ID            uint         `gorm:"primaryKey"`
	ThemeID       uint         `json:"theme_id"`
	HomeID        uint         `json:"home_id"`
	Rating        int          `json:"rating"`
	ChatType      uint         `json:"chat_type"`
	ChatTypeTitle string       `json:"chat_type_title"`
	Prompt        string       `json:"prompt"`
	Results       []ChatResult `gorm:"foreignKey:ChatID"`
}

type ChatResult struct {
	ID     uint   `gorm:"primaryKey"`
	ChatID uint   `json:"chat_id"` // Foreign key to the Chat
	Result string `json:"result"`  // Actual result string
	Role   string `json:"role"`
}

type ChatMeta struct {
	SelectedChatID uint
	ChatTypeID     uint
	ThemeID        uint
	ThemeName      string
	HomeID         uint
}

func GetPointMeta(db *gorm.DB, themeIDOverride uint) PointMeta {

	activeTheme := GetActiveTheme(db, themeIDOverride)

	allFactors := GetFactors(db)

	return PointMeta{
		types: []PointTypes{
			{ID: 1, Name: "Home"},
			{ID: 2, Name: "RedFlag"},
			{ID: 3, Name: "Office"},
			{ID: 4, Name: "LocationOfInterest"},
			{ID: 5, Name: "FractalAISearch"},
		},
		icons: []PointIcons{
			{ID: 1, Name: "Home"},
			{ID: 2, Name: "Shape"},
		},
		factors: allFactors,
		actionModes: []ActionMode{
			{ID: 1, Key: "navigate", Name: "Navigate", Details: navigateDescription()},
			{ID: 2, Key: "queries", Name: "Queries", Details: loadFractalSearches(0)},
			{ID: 3, Key: "point", Name: "Add Points", Details: addPointsDescription()},
			{ID: 4, Key: "existing-points", Name: "Existing Points", Details: pointListLoad(), FullPanel: true},
			{ID: 5, Key: "image", Name: "Add Images", Details: resizeModeWords()},
			//{ID: 3, Key: "add-image", Name: "Add Image", Details: addImage()},
			{ID: 6, Key: "area", Name: "Areas", Details: addAreasDescription()},
			{ID: 7, Key: "manage", Name: "Manage", Details: manageDescription()},
			{ID: 8, Key: "factor", Name: "Factors", Details: factorListLoad()},
		},
		theme: activeTheme,
	}
}
