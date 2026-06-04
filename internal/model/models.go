package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        string `gorm:"primaryKey;size:36"`
	Name      string `gorm:"size:100;not null"`
	Email     string `gorm:"size:100;uniqueIndex;not null"`
	Password  string `gorm:"size:255;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Transaction struct {
	ID         string `gorm:"primaryKey;size:36"`
	UserID     string `gorm:"index;size:36;not null"`
	Type       string `gorm:"size:10;not null;index"` // income | expense
	Amount     int64  `gorm:"not null"`
	CategoryID string `gorm:"size:36;index"`
	OccurredAt string `gorm:"size:10;not null;index"` // YYYY-MM-DD
	Note       string `gorm:"size:200"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type Category struct {
	ID        string `gorm:"primaryKey;size:36"`
	UserID    string `gorm:"index;size:36"`
	Name      string `gorm:"size:50;not null"`
	Type      string `gorm:"size:10;not null;index"` // income | expense
	Icon      string `gorm:"size:10"`
	Color     string `gorm:"size:10"`
	IsDefault bool   `gorm:"default:false"`
}

type SavingGoal struct {
	ID           string `gorm:"primaryKey;size:36"`
	UserID       string `gorm:"index;size:36;not null"`
	Name         string `gorm:"size:100;not null"`
	TargetAmount int64  `gorm:"not null"`
	CurrentSaved int64  `gorm:"default:0"`
	TargetDate   string `gorm:"size:10"`
	Status       string `gorm:"size:10;default:active;index"` // active | archived
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type GoalContribution struct {
	ID        string `gorm:"primaryKey;size:36"`
	GoalID    string `gorm:"index;size:36;not null"`
	Amount    int64  `gorm:"not null"`
	Note      string `gorm:"size:200"`
	Date      string `gorm:"size:10;not null;index"`
	CreatedAt time.Time
}

type Budget struct {
	ID         string `gorm:"primaryKey;size:36"`
	UserID     string `gorm:"index;size:36;not null"`
	Name       string `gorm:"size:100"`
	Amount     int64  `gorm:"not null"`
	Period     string `gorm:"size:10;not null;index"` // daily | weekly | monthly | yearly
	CategoryID string `gorm:"size:36;index;not null"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (t *Transaction) BeforeCreate(tx *gorm.DB) error {
	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()
	return nil
}

func (g *SavingGoal) BeforeCreate(tx *gorm.DB) error {
	g.CreatedAt = time.Now()
	g.UpdatedAt = time.Now()
	return nil
}

func (b *Budget) BeforeCreate(tx *gorm.DB) error {
	b.CreatedAt = time.Now()
	b.UpdatedAt = time.Now()
	return nil
}
