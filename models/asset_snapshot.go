package models

import (
	"encoding/json"
	"time"

	"github.com/fox-one/gin-contrib/session"
	"github.com/go-redis/cache"
	"github.com/shopspring/decimal"
)

// AssetSnapshot asset snapshot
type AssetSnapshot struct {
	ID        int64           `gorm:"PRIMARY_KEY;"              json:"-"`
	AssetID   string          `gorm:"SIZE:36;unique_index:asset_date;"  json:"asset_id"`
	Date      time.Time       `gorm:"SIZE:36;unique_index:asset_date;"  json:"date"`
	Price     decimal.Decimal `gorm:"TYPE:VARCHAR(50);"         json:"price"`
	MarketCap decimal.Decimal `gorm:"TYPE:VARCHAR(50);"         json:"market_cap"`
	Volume    decimal.Decimal `gorm:"TYPE:VARCHAR(50);"         json:"volume"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// TableName mysql table
func (AssetSnapshot) TableName(interval int) string {
	switch interval {
	case SnapshotIntervalDaily:
		return "daily_asset_snapshots"

	case SnapshotIntervalMonthly:
		return "monthly_asset_snapshots"
	}

	return ""
}

// AssetSnapshots asset snapshots
type AssetSnapshots []*AssetSnapshot

// QueryAssetSnapshots query asset snapshots
func QueryAssetSnapshots(s *session.Session, interval int, assetIDs []string, from, to time.Time) (map[string]AssetSnapshots, error) {
	const limit = 500

	var tableName string
	switch interval {
	case SnapshotIntervalDaily, SnapshotIntervalMonthly:
		tableName = AssetSnapshot{}.TableName(interval)

	default:
		return nil, ErrInvalidInterval
	}

	m := map[string]AssetSnapshots{}
	db := s.MysqlRead().Table(tableName)
	for _, assetID := range assetIDs {
		query := db.Where("asset_id = ? AND date <= ?", assetID, to).
			Order("date ASC").Limit(limit)

		cursor := from
		var snapshots AssetSnapshots
		for {
			var arr AssetSnapshots
			if err := query.Where("date > ?", cursor).Find(&arr).Error; err != nil {
				return nil, err
			}

			snapshots = append(snapshots, arr...)
			if len(arr) < limit {
				break
			}
			cursor = snapshots[len(snapshots)-1].Date
		}
		m[assetID] = snapshots
	}
	return m, nil
}

// QueryLastestAssetSnapshots get lastest price
func QueryLastestAssetSnapshots(s *session.Session) (AssetSnapshots, error) {
	const limit = 500
	offset := 0
	db := s.MysqlRead().Table(AssetSnapshot{}.TableName(SnapshotIntervalDaily))

	var assets []*AssetSnapshot
	for {
		var tmp []*AssetSnapshot
		if err := db.Where("date = ?", time.Now().Truncate(24*time.Hour)).Offset(offset).Limit(limit).Find(&tmp).Error; err != nil {
			return nil, err
		}
		assets = append(assets, tmp...)

		if len(tmp) < limit {
			break
		}
		offset += len(tmp)
	}
	return assets, nil
}

const (
	assetSnapshotsRedisKey = "fox-data:daily_asset_snapshots"
)

// GetLastestAssetSnapshots get lastest price
func GetLastestAssetSnapshots(s *session.Session) (AssetSnapshots, error) {
	var assets AssetSnapshots
	err := cacheCodec(s).Get(assetSnapshotsRedisKey, &assets)
	return assets, err
}

// CacheLastestAssetSnapshots cache lastest price
func CacheLastestAssetSnapshots(s *session.Session, assets AssetSnapshots) error {
	return cacheCodec(s).Set(&cache.Item{
		Key:        assetSnapshotsRedisKey,
		Object:     assets,
		Expiration: 24 * time.Hour,
	})
}

func cacheCodec(s *session.Session) *cache.Codec {
	return &cache.Codec{
		Redis: s.Redis(),

		Marshal: func(v interface{}) ([]byte, error) {
			return json.Marshal(v)
		},
		Unmarshal: func(b []byte, v interface{}) error {
			return json.Unmarshal(b, v)
		},
	}
}
