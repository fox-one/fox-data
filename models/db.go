package models

import (
	"github.com/fox-one/gin-contrib/session"
	"github.com/jinzhu/gorm"
)

func setupDB(db *gorm.DB) error {
	if err := db.AutoMigrate(&Asset{}, &FoiPoint{}).Error; err != nil {
		return err
	}
	if err := db.Table(AssetSnapshot{}.TableName(SnapshotIntervalDaily)).AutoMigrate(&AssetSnapshot{}).Error; err != nil {
		return err
	}
	if err := db.Table(AssetSnapshot{}.TableName(SnapshotIntervalMonthly)).AutoMigrate(&AssetSnapshot{}).Error; err != nil {
		return err
	}
	return nil
}

func init() {
	session.RegisterSetdb(setupDB)
}
