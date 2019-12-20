package models

import (
	"errors"

	"github.com/fox-one/gin-contrib/session"
)

const (
	// SnapshotIntervalDaily daily
	SnapshotIntervalDaily = iota + 1
	// SnapshotIntervalMonthly monthly
	SnapshotIntervalMonthly
)

var (
	// ErrInvalidInterval invalid interval
	ErrInvalidInterval = errors.New("invalid interval")
)

// Asset asset
type Asset struct {
	AssetID  string `gorm:"SIZE:36;PRIMARY_KEY;" json:"asset_id"`
	ChainID  string `gorm:"SIZE:36;NOT NULL;"    json:"chain_id"`
	AssetKey string `gorm:"SIZE:255;"            json:"asset_key,omitempty"`
	Name     string `gorm:"SIZE:255;NOT NULL;"   json:"name,omitempty"`
	Symbol   string `gorm:"SIZE:255;NOT NULL;"   json:"symbol,omitempty"`
	IconURL  string `gorm:"SIZE:512;"            json:"icon_url,omitempty"`
	CmcSlug  string `gorm:"SIZE:50;NOT NULL:"    json:"cmc_slug,omitempty"`
}

// TableName table name
func (Asset) TableName() string {
	return "assets"
}

// AllAssets all assets
func AllAssets(s *session.Session) ([]*Asset, error) {
	const limit = 500
	assets := make([]*Asset, 0)

	offset := 0
	query := s.MysqlRead().Limit(limit)
	for {
		var arr []*Asset
		if err := query.Offset(offset).Find(&arr).Error; err != nil {
			return nil, err
		}
		assets = append(assets, arr...)

		if len(arr) < limit {
			return assets, nil
		}

		offset += len(arr)
	}
}
