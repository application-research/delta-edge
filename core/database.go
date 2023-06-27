package core

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"github.com/application-research/edge-ur/config"
	"github.com/filecoin-project/go-address"
	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p/core/peer"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"time"
)

func OpenDatabase(cfg config.EdgeConfig) (*gorm.DB, error) {

	// use postgres
	var DB *gorm.DB
	var err error

	if cfg.Node.DbDsn[:8] == "postgres" {
		DB, err = gorm.Open(postgres.Open(cfg.Node.DbDsn), &gorm.Config{})
	} else {
		DB, err = gorm.Open(sqlite.Open(cfg.Node.DbDsn), &gorm.Config{})
	}

	sqldb, err := DB.DB()
	if err != nil {
		return nil, err
	}
	sqldb.SetMaxIdleConns(250)
	sqldb.SetMaxOpenConns(250)
	sqldb.SetConnMaxIdleTime(time.Hour)
	sqldb.SetConnMaxLifetime(time.Hour)

	// generate new models.
	ConfigureModels(DB) // create models.

	if err != nil {
		return nil, err
	}
	return DB, nil
}

func ConfigureModels(db *gorm.DB) {
	db.AutoMigrate(&Content{}, &ContentDeal{}, &LogEvent{}, &Bucket{}, &Policy{}, &ContentSignatureMeta{})
}

type LogEvent struct {
	ID             int64     `gorm:"primaryKey"` // auto increment
	SourceHost     string    `json:"source_host"`
	SourceIP       string    `json:"source_ip"`
	LogEventType   string    `json:"log_event_type"` // content, deal, piece_commitment, upload, miner, info
	LogEventObject []byte    `json:"event_object"`
	LogEventId     int64     `json:"log_event_id"` // object id
	LogEvent       string    `json:"log_event"`    // description
	DeltaUuid      string    `json:"delta_uuid"`
	CreatedAt      time.Time `json:"created_at"` // auto set
	UpdatedAt      time.Time `json:"updated_at"`
}

type Policy struct {
	gorm.Model
	ID         int64     `gorm:"primaryKey"`
	Name       string    `json:"name"`
	BucketSize int64     `json:"bucket_size"`
	SplitSize  int64     `json:"split_size"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Bucket struct {
	gorm.Model
	ID               int64     `gorm:"primaryKey"`
	Uuid             string    `gorm:"index" json:"uuid"`
	Name             string    `json:"name"`
	Size             int64     `json:"size"`
	RequestingApiKey string    `json:"requesting_api_key,omitempty"`
	Miner            string    `json:"miner"`
	PieceCid         string    `json:"piece_cid"`
	PieceSize        int64     `json:"piece_size"`
	DirCid           string    `json:"dir_cid"`
	Cid              string    `json:"cid"`
	Status           string    `json:"status"`
	PolicyId         int64     `json:"policy_id"`
	LastMessage      string    `json:"last_message"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type ContentSignatureMeta struct {
	ID                  int64     `gorm:"primaryKey"`
	ContentId           int64     `json:"content_id"`
	Signature           string    `json:"signature"`
	CurrentTimestamp    time.Time `json:"current_timestamp"`
	ExpirationTimestamp time.Time `json:"expiration_timestamp"`
	SignedUrl           string    `json:"signed_url"`
	Message             string    `json:"message"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// main content record
type Content struct {
	gorm.Model
	ID               int64     `gorm:"primaryKey"`
	Name             string    `json:"name"`
	Size             int64     `json:"size"`
	Cid              string    `json:"cid"`
	RequestingApiKey string    `json:"requesting_api_key,omitempty"`
	BucketUuid       string    `json:"bucket_uuid"`
	Status           string    `json:"status"`
	PieceCid         string    `json:"piece_cid"`
	PieceSize        int64     `json:"piece_size"`
	LastMessage      string    `json:"last_message"`
	Miner            string    `json:"miner"`
	MakeDeal         bool      `json:"make_deal"`
	TagName          string    `json:"tag_name"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type ContentTag struct {
	ID        int64     `gorm:"primaryKey"`
	ContentId int64     `json:"content_id"`
	TagId     int64     `json:"tag_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Collection struct {
	ID               int64     `gorm:"primaryKey"`
	UUID             string    `gorm:"index" json:"uuid"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	RequestingApiKey string    `json:"requesting_api_key"`
	Cid              string    `json:"cid"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type CollectionRef struct {
	ID         uint      `gorm:"primaryKey"`
	Collection int64     `gorm:"index:,option:CONCURRENTLY;not null"`
	Content    uint64    `gorm:"index:,option:CONCURRENTLY;not null"`
	Path       *string   `gorm:"null"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type ContentReplication struct {
	ID                    int64     `gorm:"primaryKey"`
	PrimaryContentID      int64     `json:"primary_content_id"`
	PrimaryDeltaContentID int64     `json:"primary_delta_content_id"`
	ReplicaContentID      int64     `json:"replica_content_id"`
	ReplicaDeltaContentID int64     `json:"replica_delta_content_id"`
	CreatedAt             time.Time `json:"CreatedAt"`
	UpdatedAt             time.Time `json:"UpdatedAt"`
}

type ContentCarSplit struct {
	ID        int64 `gorm:"primaryKey"`
	ContentID int64 `json:"content_id"`
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

type DbAddrInfo struct {
	AddrInfo peer.AddrInfo
}

func (dba *DbAddrInfo) Scan(v interface{}) error {
	b, ok := v.([]byte)
	if !ok {
		return fmt.Errorf("DbAddrInfo must be bytes")
	}

	if len(b) == 0 {
		return nil
	}

	var addrInfo peer.AddrInfo
	if err := json.Unmarshal(b, &addrInfo); err != nil {
		return err
	}

	dba.AddrInfo = addrInfo
	return nil
}

func (dba DbAddrInfo) Value() (driver.Value, error) {
	return dba.AddrInfo.MarshalJSON()
}

type DbAddr struct {
	Addr address.Address
}

func (dba *DbAddr) Scan(v interface{}) error {
	s, ok := v.(string)
	if !ok {
		return fmt.Errorf("DbAddrs must be strings")
	}

	addr, err := address.NewFromString(s)
	if err != nil {
		return err
	}

	dba.Addr = addr
	return nil
}

func (dba DbAddr) Value() (driver.Value, error) {
	return dba.Addr.String(), nil
}

type DbCID struct {
	CID cid.Cid
}

func (dbc *DbCID) Scan(v interface{}) error {
	b, ok := v.([]byte)
	if !ok {
		return fmt.Errorf("dbcids must get bytes!")
	}

	if len(b) == 0 {
		return nil
	}

	c, err := cid.Cast(b)
	if err != nil {
		return err
	}

	dbc.CID = c
	return nil
}

func (dbc DbCID) Value() (driver.Value, error) {
	return dbc.CID.Bytes(), nil
}

func (dbc DbCID) MarshalJSON() ([]byte, error) {
	return json.Marshal(dbc.CID.String())
}

func (dbc *DbCID) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	c, err := cid.Decode(s)
	if err != nil {
		return err
	}

	dbc.CID = c
	return nil
}
