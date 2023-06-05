package config

import (
	"github.com/caarlos0/env/v6"
	logging "github.com/ipfs/go-log/v2"
	"github.com/joho/godotenv"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

var (
	log                       = logging.Logger("config")
	defaultTestBootstrapPeers []multiaddr.Multiaddr
)

type DeltaConfig struct {
	Node struct {
		Name        string `env:"NODE_NAME" envDefault:"edge-ur"`
		Description string `env:"NODE_DESCRIPTION"`
		Type        string `env:"NODE_TYPE"`
		DbDsn       string `env:"DB_DSN" envDefault:"edge-ur.db"`
		Repo        string `env:"REPO" envDefault:"./whypfs"`
		DsRepo      string `env:"DS_REPO" envDefault:"./whypfs"`
		Port        int    `env:"PORT" envDefault:"1414"`
	}

	Common struct {
		AggregateSize      int64 `env:"AGGREGATE_SIZE" envDefault:"1048576000"`
		AggregatePerApiKey bool  `env:"AGGREGATE_PER_API_KEY" envDefault:"false"`
		MaxSizeToSplit     int64 `env:"MAX_SIZE_TO_SPLIT" envDefault:"32000000000"`
		DealCheck          int   `env:"DEAL_CHECK" envDefault:"600"`
		ReplicationFactor  int   `env:"REPLICATION_FACTOR" envDefault:"0"`
		// Capacity Limit per Key: default 0 - unlimited
		CapacityLimitPerKeyInBytes int64 `env:"CAPACITY_LIMIT_PER_KEY_IN_BYTES" envDefault:"0"`
	}

	ExternalApi struct {
		DeltaNodeApiUrl string `env:"DELTA_NODE_API" envDefault:"http://localhost:1414"`
		AuthSvcUrl      string `env:"AUTH_SVC_API" envDefault:"https://auth.estuary.tech"`
	}
}

func InitConfig() DeltaConfig {
	godotenv.Load() // load from environment OR .env file if it exists
	var cfg DeltaConfig

	if err := env.Parse(&cfg); err != nil {
		log.Fatal("error parsing config: %+v\n", err)
	}

	log.Debug("config parsed successfully")

	return cfg
}

// BootstrapEstuaryPeers Creating a list of multiaddresses that are used to bootstrap the network.
func BootstrapEstuaryPeers() []peer.AddrInfo {

	for _, s := range []string{
		"/dnsaddr/bootstrap.libp2p.io/p2p/QmNnooDu7bfjPFoTZYxMNLWUQJyrVwtbZg5gBMjTezGAJN",
		"/dnsaddr/bootstrap.libp2p.io/p2p/QmQCU2EcMqAqQPR2i9bChDtGNJchTbq5TbXJJ16u19uLTa",
		"/dnsaddr/bootstrap.libp2p.io/p2p/QmbLHAnMoJPWSCR5Zhtx6BHJX9KiKNN6tpvbUcqanj75Nb",
		"/dnsaddr/bootstrap.libp2p.io/p2p/QmcZf59bWwK5XFi76CZX8cbJ4BhTzzA3gU1ZjYZcYW3dwt",
		"/ip4/104.131.131.82/tcp/4001/p2p/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",      // mars.i.ipfs.io
		"/ip4/104.131.131.82/udp/4001/quic/p2p/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ", // mars.i.ipfs.io
	} {
		ma, err := multiaddr.NewMultiaddr(s)
		if err != nil {
			panic(err)
		}
		defaultTestBootstrapPeers = append(defaultTestBootstrapPeers, ma)
	}

	peers, _ := peer.AddrInfosFromP2pAddrs(defaultTestBootstrapPeers...)
	return peers
}
