package store

import (
	"time"

	"github.com/fox-one/fox-data/models"
	"github.com/fox-one/gin-contrib/session"
	jsoniter "github.com/json-iterator/go"
)

type foiSQLStore struct {
	table string
}

// NewSQLFoiStore  mysql version  FoiStore
func NewSQLFoiStore(s *session.Session) models.FoiStore {
	return &foiSQLStore{table: "foi_points"}
}

func (store *foiSQLStore) Save(s *session.Session, point models.Point) error {
	point.Timestamp = time.Unix(point.Timestamp/1000, 0).Truncate(time.Hour).Unix() * 1000
	p := models.FoiPoint{Point: point}

	return s.MysqlWrite().Table(store.table).Where("timestamp = ?", p.Timestamp).Assign(p).FirstOrCreate(&p).Error
}

func (store *foiSQLStore) FetchLatest(s *session.Session, limit int) ([]models.Point, error) {
	points := []models.Point{}
	err := s.MysqlRead().Table(store.table).Order("timestamp DESC").Limit(limit).Find(&points).Error

	// reverse points
	for i, j := 0, len(points)-1; i < j; i, j = i+1, j-1 {
		points[i], points[j] = points[j], points[i]
	}

	return points, err
}

func (store *foiSQLStore) FetchRange(s *session.Session, mod int64, from, to time.Time) ([]models.Point, error) {
	points := []models.Point{}

	f, t := from.Unix()*1000, to.Unix()*1000
	err := s.MysqlRead().Table(store.table).
		Where("timestamp >= ? AND timestamp <= ? AND timestamp % ? = 0", f, t, mod).
		Order("timestamp").
		Find(&points).Error

	return points, err
}

type otcRedisStore struct {
	key string
}

func (store *otcRedisStore) Save(s *session.Session, points ...models.Point) error {
	fields := make(map[string]interface{})
	for _, point := range points {
		if bytes, err := jsoniter.Marshal(point); err == nil {
			fields[point.Symbol] = bytes
		}
	}

	return s.Redis().HMSet(store.key, fields).Err()
}

func (store *otcRedisStore) Dump(s *session.Session) ([]models.Point, error) {
	byteMap, err := s.Redis().HGetAll(store.key).Result()
	if err != nil {
		return nil, err
	}

	points := make([]models.Point, 0, len(byteMap))
	for _, data := range byteMap {
		p := models.Point{}
		if jsoniter.Unmarshal([]byte(data), &p) == nil {
			points = append(points, p)
		}
	}

	return points, err
}

func (store *otcRedisStore) DumpSymbol(s *session.Session, symbol string) (*models.Point, error) {
	data, err := s.Redis().HGet(store.key, symbol).Result()
	if err != nil {
		return nil, err
	}

	p := &models.Point{}
	err = jsoniter.Unmarshal([]byte(data), p)
	return p, err
}

// NewRedisOtcStore  redis version  OtcStore
func NewRedisOtcStore() models.OtcStore {
	return &otcRedisStore{key: "foxone_otc_latest_snapshots"}
}
