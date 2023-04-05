package core

import (
	"github.com/spf13/viper"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"time"
)

func OpenDatabase() (*gorm.DB, error) {

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
	db.AutoMigrate(&Content{}, &ContentDeal{})
}

//	 main content record
type Content struct {
	ID               int64     `gorm:"primaryKey"`
	Name             string    `json:"name"`
	Size             int64     `json:"size"`
	Cid              string    `json:"cid"`
	RequestingApiKey string    `json:"requesting_api_key,omitempty"`
	DeltaContentId   int64     `json:"delta_content_id"`
	Status           string    `json:"status"`
	LastMessage      string    `json:"last_message"`
	Miner            string    `json:"miner"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type ContentDeal struct {
	ID                  int64     `gorm:"primaryKey"`
	CreatedAt           time.Time `json:"CreatedAt"`
	UpdatedAt           time.Time `json:"UpdatedAt"`
	DeletedAt           time.Time `json:"DeletedAt"`
	ContentId           int64     `json:"delta_content_id"`
	UserID              int       `json:"user_id"`
	PropCid             string    `json:"propCid"`
	DealUUID            string    `json:"dealUuid"`
	Miner               string    `json:"miner"`
	Status              string    `json:"status"`
	Failed              bool      `json:"failed"`
	Verified            bool      `json:"verified"`
	Slashed             bool      `json:"slashed"`
	FailedAt            time.Time `json:"failedAt"`
	DtChan              string    `json:"dtChan"`
	TransferStarted     time.Time `json:"transferStarted"`
	TransferFinished    time.Time `json:"transferFinished"`
	OnChainAt           time.Time `json:"onChainAt"`
	SealedAt            time.Time `json:"sealedAt"`
	DealProtocolVersion string    `json:"deal_protocol_version"`
	MinerVersion        string    `json:"miner_version"`
}
