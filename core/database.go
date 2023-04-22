package core

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"github.com/application-research/edge-ur/config"
	"github.com/filecoin-project/go-address"
	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p/core/peer"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"time"
)

func OpenDatabase(cfg config.DeltaConfig) (*gorm.DB, error) {

	DB, err := gorm.Open(sqlite.Open(cfg.Node.DbName), &gorm.Config{})

	// generate new models.
	ConfigureModels(DB) // create models.

	if err != nil {
		return nil, err
	}
	return DB, nil
}

func ConfigureModels(db *gorm.DB) {
	db.AutoMigrate(&Content{}, &ContentDeal{}, &Collection{}, CollectionRef{})
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
	Status           string    `json:"status"`
	LastMessage      string    `json:"last_message"`
	Miner            string    `json:"miner"`
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
