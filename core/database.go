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

func OpenDatabase(cfg config.DeltaConfig) (*gorm.DB, error) {

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
	db.AutoMigrate(&Content{}, &ContentDeal{}, &Collection{}, &CollectionRef{}, &LogEvent{}, &Bucket{}, &Bundle{})
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

type Bundle struct {
	ID                int64     `gorm:"primaryKey"`
	Uuid              string    `gorm:"index" json:"uuid"`
	Name              string    `json:"name"`
	Size              int64     `json:"size"`
	DeltaContentId    int64     `json:"delta_content_id"`
	DeltaNodeUrl      string    `json:"delta_node_url"`
	RequestingApiKey  string    `json:"requesting_api_key,omitempty"`
	Miner             string    `json:"miner"`
	FileCid           string    `json:"file_cid"`
	AggregatePieceCid string    `json:"aggregate_piece_cid"`
	InclusionProof    string    `json:"inclusion_proof"`
	Status            string    `json:"status"` // open, processing, filled, uploaded-to-delta
	LastMessage       string    `json:"last_message"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
type Bucket struct {
	ID               int64     `gorm:"primaryKey" json:"id,omitempty"`
	Uuid             string    `gorm:"index" json:"uuid" json:"uuid,omitempty"`
	BundleUuid       string    `gorm:"index" json:"bundle_uuid" json:"bundle_uuid,omitempty"`
	Name             string    `json:"name" json:"name,omitempty"`
	Size             int64     `json:"size" json:"size,omitempty"`
	RequestingApiKey string    `json:"requesting_api_key,omitempty" json:"requesting_api_key,omitempty"`
	DeltaContentId   int64     `json:"delta_content_id" json:"delta_content_id,omitempty"`
	DeltaNodeUrl     string    `json:"delta_node_url" json:"delta_node_url,omitempty"`
	Miner            string    `json:"miner" json:"miner,omitempty"`
	PieceCid         string    `json:"piece_cid" json:"piece_cid,omitempty"`
	PieceSize        int64     `json:"piece_size" json:"piece_size,omitempty"`
	InclusionProof   []byte    `json:"inclusion_proof" json:"inclusion_proof,omitempty"`
	CommPa           string    `json:"comm_pa,omitempty"`
	SizePa           int64     `json:"size_pa,omitempty"`
	Cid              string    `json:"cid" json:"cid,omitempty"`
	Status           string    `json:"status" json:"status,omitempty"` // open, processing, filled, bundled
	LastMessage      string    `json:"last_message" json:"last_message,omitempty"`
	CreatedAt        time.Time `json:"created_at" json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" json:"updated_at"`
}

//	 main content record
type Content struct {
	ID               int64     `gorm:"primaryKey"`
	Name             string    `json:"name"`
	Size             int64     `json:"size"`
	Cid              string    `json:"cid"`
	RequestingApiKey string    `json:"requesting_api_key,omitempty"`
	DeltaContentId   int64     `json:"delta_content_id"`
	DeltaNodeUrl     string    `json:"delta_node_url"`
	BucketUuid       string    `json:"bucket_uuid"`
	Status           string    `json:"status"`
	PieceCid         string    `json:"piece_cid"`
	PieceSize        int64     `json:"piece_size"`
	InclusionProof   []byte    `json:"inclusion_proof" json:"inclusion_proof,omitempty"`
	LastMessage      string    `json:"last_message"`
	Miner            string    `json:"miner"`
	MakeDeal         bool      `json:"make_deal"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
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
