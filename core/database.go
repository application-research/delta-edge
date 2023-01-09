package core

import (
	"github.com/spf13/viper"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"time"
)

func OpenDatabase() (*gorm.DB, error) {

	viper.SetConfigFile(".env")
	err := viper.ReadInConfig()

	dbName, okHost := viper.Get("DB_NAME").(string)
	if !okHost {
		panic("DB_NAME not set")
	}
	DB, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})

	// generate new models.
	ConfigureModels(DB) // create models.

	if err != nil {
		return nil, err
	}
	return DB, nil
}

func ConfigureModels(db *gorm.DB) {
	db.AutoMigrate(&Content{}, &Bucket{})
}

type Content struct {
	ID               uint   `gorm:"primaryKey"`
	Name             string `json:"name"`
	Size             int64  `json:"size"`
	Cid              string `json:"cid"`
	BucketUuid       string `json:"bucket_uuid"`
	RequestingApiKey string `json:"requesting_api_key"`
	EstuaryContentId string `json:"estuary_content_id"`
	Created_at       time.Time
	Updated_at       time.Time
}

type Bucket struct {
	ID         uint   `gorm:"primaryKey"`
	Name       string `json:"name"`
	UUID       string `json:"uuid"`
	Status     string `json:"status"` // open, in-progress, completed (closed).
	Created_at time.Time
	Updated_at time.Time
}
