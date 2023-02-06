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
	db.AutoMigrate(&Content{}, &Bucket{}, &ContentStatus{}, &ContentDeal{}, &ContentSplitRequest{})
}

type ContentSplitRequest struct {
	ID        int64  `gorm:"primaryKey"`
	ContentId int64  `json:"content_id"`
	Cid       string `json:"cid"`
	Size      int64  `json:"size"`
	Name      string `json:"name"`
	Origins   string `json:"origins,omitempty"`
	ChunkSize int64  `json:"chunk_size"`
}

//	 main content record
type Content struct {
	ID               int64     `gorm:"primaryKey"`
	Name             string    `json:"name"`
	Size             int64     `json:"size"`
	Cid              string    `json:"cid"`
	BucketUuid       string    `json:"bucket_uuid,omitempty"`
	RequestingApiKey string    `json:"requesting_api_key,omitempty"`
	EstuaryContentId int64     `json:"estuary_content_id"`
	SplitRequestId   int64     `json:"split_request_id,omitempty"`
	Status           string    `json:"status"`
	Origins          string    `json:"origins,omitempty"`
	Created_at       time.Time `json:"created_at"`
	Updated_at       time.Time `json:"updated_at"`
}

type ContentStatus struct {
	ID            int64     `gorm:"primaryKey"`
	ContentId     int64     `json:"estuary_content_id"`
	CreatedAt     time.Time `json:"createdAtOnEstuary"`
	UpdatedAt     time.Time `json:"updatedAtOEstuary"`
	Cid           string    `json:"cid"`
	Name          string    `json:"name"`
	UserID        int       `json:"userId"`
	Description   string    `json:"description"`
	Size          int       `json:"size"`
	Type          int       `json:"type"`
	Active        bool      `json:"active"`
	Offloaded     bool      `json:"offloaded"`
	Replication   int       `json:"replication"`
	AggregatedIn  int       `json:"aggregatedIn"`
	Aggregate     bool      `json:"aggregate"`
	Pinning       bool      `json:"pinning"`
	PinMeta       string    `json:"pinMeta"`
	Replace       bool      `json:"replace"`
	Origins       string    `json:"origins"`
	Failed        bool      `json:"failed"`
	Location      string    `json:"location"`
	DagSplit      bool      `json:"dagSplit"`
	SplitFrom     int       `json:"splitFrom"`
	PinningStatus string    `json:"pinningStatus"`
	DealStatus    string    `json:"dealStatus"`
	Created_at    time.Time `json:"created_at"`
	Updated_at    time.Time `json:"updated_at"`
}

type ContentDeal struct {
	ID        int64     `gorm:"primaryKey"`
	CreatedAt time.Time `json:"CreatedAt"`
	UpdatedAt time.Time `json:"UpdatedAt"`
	DeletedAt time.Time `json:"DeletedAt"`
	ContentId int64     `json:"estuary_content_id"`
	UserID    int       `json:"user_id"`
	PropCid   string    `json:"propCid"`
	DealUUID  string    `json:"dealUuid"`
	Miner     string    `json:"miner"`
	//DealID              int         `json:"dealId"`
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

// buckets are aggregations of contents. It can either generate a car or just aggregate.
type Bucket struct {
	ID               int64     `gorm:"primaryKey"`
	Name             string    `json:"name"`
	UUID             string    `json:"uuid"`
	Status           string    `json:"status"`
	Cid              string    `json:"cid"`
	RequestingApiKey string    `json:"requesting_api_key,omitempty"`
	EstuaryContentId int64     `json:"estuary_content_id"`
	Created_at       time.Time `json:"created_at"`
	Updated_at       time.Time `json:"updated_at"`
}
