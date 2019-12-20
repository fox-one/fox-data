package models

import (
	"time"

	"github.com/fox-one/gin-contrib/session"
)

// Point definitation
type Point struct {
	Timestamp int64  `gorm:"INDEX"               json:"timestamp"`
	Sell      string `gorm:"TYPE:VARCHAR(10)"    json:"sell"`
	Buy       string `gorm:"TYPE:VARCHAR(10)"    json:"buy"`
	USD       string `gorm:"TYPE:VARCHAR(10)"    json:"usd,omitempty"`
	Symbol    string `gorm:"TYPE:VARCHAR(10)"    json:"symbol"`
}

// FoiPoint foi point definitation
type FoiPoint struct {
	ID uint `gorm:"PRIMARY_KEY"`

	Point
}

// TableName define gorm table name
func (FoiPoint) TableName() string {
	return "foi_points"
}

// FoiStore the interface to read&write foi
type FoiStore interface {
	Save(s *session.Session, point Point) error
	FetchRange(s *session.Session, mod int64, from, to time.Time) ([]Point, error)
	FetchLatest(s *session.Session, limit int) ([]Point, error)
}

// OtcStore the interface to read&write otc
type OtcStore interface {
	Save(s *session.Session, points ...Point) error
	Dump(s *session.Session) ([]Point, error)
	DumpSymbol(s *session.Session, symbol string) (*Point, error)
}
