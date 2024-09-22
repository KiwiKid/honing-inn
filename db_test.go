package main

import (
	"testing"

	"gorm.io/gorm"
)

func TestDeleteChatType(t *testing.T) {
	t.Parallel()

	// Initialize DB using DBInit
	config := EnvConfig{DBUrl: ":memory:"}
	db, err := DBInit(config)
	if err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}
	t.Cleanup(func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	})

	// Define test cases
	tests := []struct {
		name    string
		id      uint
		setup   func(db *gorm.DB)
		wantErr bool
	}{
		{
			name: "Delete existing ChatType",
			id:   1,
			setup: func(db *gorm.DB) {
				db.Create(&ChatType{ID: 1})
			},
			wantErr: false,
		},
		{
			name:    "Delete non-existent ChatType",
			id:      2,
			setup:   func(db *gorm.DB) {},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(db)

			_, err := DeleteChatType(db, tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteChatType() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetChats(t *testing.T) {
	t.Parallel()

	config := EnvConfig{DBUrl: ":memory:"}
	db, err := DBInit(config)
	if err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}
	t.Cleanup(func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	})

	tests := []struct {
		name       string
		themeID    uint
		homeID     uint
		chatTypeID uint
		setup      func(db *gorm.DB)
		wantCount  int
		wantErr    bool
	}{
		{
			name:       "Get chats by theme, home and chat type",
			themeID:    1,
			homeID:     2,
			chatTypeID: 3,
			setup: func(db *gorm.DB) {
				db.Create(&Chat{ThemeID: 1, HomeID: 2, ChatType: 3})
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:       "No chats found",
			themeID:    1,
			homeID:     99,
			chatTypeID: 1,
			setup:      func(db *gorm.DB) {},
			wantCount:  0,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(db)

			got, err := GetChats(db, tt.themeID, tt.homeID, tt.chatTypeID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetChats() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got) != tt.wantCount {
				t.Errorf("GetChats() got = %d chats, want %d", len(got), tt.wantCount)
			}
		})
	}
}

func TestGetChat(t *testing.T) {
	t.Parallel()

	config := EnvConfig{DBUrl: ":memory:"}
	db, err := DBInit(config)
	if err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}
	t.Cleanup(func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	})

	tests := []struct {
		name    string
		id      uint
		setup   func(db *gorm.DB)
		wantErr bool
	}{
		{
			name: "Get existing chat",
			id:   1,
			setup: func(db *gorm.DB) {
				db.Create(&Chat{ID: 1})
			},
			wantErr: false,
		},
		{
			name:    "Chat does not exist",
			id:      2,
			setup:   func(db *gorm.DB) {},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(db)

			_, err := GetChat(db, tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetChat() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
