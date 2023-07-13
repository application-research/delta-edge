package core

import (
	"context"
	"fmt"
	"strconv"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/application-research/edge-ur/config"
	"github.com/ipfs/go-datastore"

	"github.com/application-research/whypfs-core"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/chain/wallet/key"
	"github.com/ipfs/go-blockservice"
	bsfetcher "github.com/ipfs/go-fetcher/impl/blockservice"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	mdagipld "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag"
	"github.com/ipfs/go-path/resolver"
	"github.com/ipfs/go-unixfsnode"
	dagpb "github.com/ipld/go-codec-dagpb"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/node/basicnode"
	"github.com/ipld/go-ipld-prime/schema"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"gorm.io/gorm"
)

type LightNode struct {
	Node   *whypfs.Node
	Api    url.URL
	DB     *gorm.DB
	Gw     *GatewayHandler
	Config *config.DeltaConfig
}

type LocalWallet struct {
	keys     map[address.Address]*key.Key
	keystore types.KeyStore

	lk sync.Mutex
}

type GatewayHandler struct {
	bs       blockstore.Blockstore
	dserv    mdagipld.DAGService
	resolver resolver.Resolver
	node     *whypfs.Node
}

// NewEdgeNode Add a config to enable gateway or not.
// Add a config to enable content, bucket, commp, replication verifier processor
func NewEdgeNode(ctx context.Context, cfg config.DeltaConfig) (*LightNode, error) {

	db, err := OpenDatabase(cfg)
	// node
	// If we don't have an explicit Public IP set as an env var, fetch one
	var publicIp string

	if cfg.Node.PublicIp == "" {
		publicIp, err = GetPublicIP()
		if err != nil {
			fmt.Printf("Error getting public IP: %v", err)
		}
	} else {
		publicIp = cfg.Node.PublicIp
	}
	// Fetch the IpfsPort from DeltaConfig
	IpfsPort := strconv.Itoa(cfg.Node.IpfsPort)

	newConfig := &whypfs.Config{
		ListenAddrs: []string{
			"/ip4/0.0.0.0/tcp/" + IpfsPort,
			"/ip4/" + publicIp + "/tcp/" + IpfsPort,
		},
		AnnounceAddrs: []string{
			"/ip4/0.0.0.0/tcp/" + IpfsPort,
			"/ip4/" + publicIp + "/tcp/" + IpfsPort,
		},
		
	}

	params := whypfs.NewNodeParams{
		Ctx:       ctx,
		Datastore: datastore.NewMapDatastore(),
		Repo:      cfg.Node.Repo,
	}

	params.Config = params.ConfigurationBuilder(newConfig)
	whypfsPeer, err := whypfs.NewNode(params)
	if err != nil {
		panic(err)
	}

	whypfsPeer.BootstrapPeers(config.BootstrapEstuaryPeers())

	// gateway
	gw, err := NewGatewayHandler(whypfsPeer)

	// create the global light node.
	return &LightNode{
		Node:   whypfsPeer,
		Gw:     gw,
		DB:     db,
		Config: &cfg,
	}, nil
}

func NewGatewayHandler(node *whypfs.Node) (*GatewayHandler, error) {

	bsvc := blockservice.New(node.Blockstore, nil)
	fetcherFactory := bsfetcher.NewFetcherConfig(bsvc)
	fetcherFactory.NodeReifier = unixfsnode.Reify
	fetcherFactory.PrototypeChooser = dagpb.AddSupportToChooser(func(lnk ipld.Link, lnkCtx ipld.LinkContext) (ipld.NodePrototype, error) {
		if tlnkNd, ok := lnkCtx.LinkNode.(schema.TypedLinkNode); ok {
			return tlnkNd.LinkTargetNodePrototype(), nil
		}
		return basicnode.Prototype.Any, nil
	})
	resolver := resolver.NewBasicResolver(fetcherFactory)
	return &GatewayHandler{
		bs:       node.Blockstore,
		dserv:    merkledag.NewDAGService(bsvc),
		resolver: resolver,
		node:     node,
	}, nil
}

func (ln *LightNode) GetLocalhostOrigins() []string {
	var origins []string
	for _, origin := range ln.Node.Config.AnnounceAddrs {
		peerInfo, err := peer.AddrInfoFromString(origin)
		if err != nil {
			continue
		}
		origins = append(origins, peerInfo.ID.String())
	}

	publicIp, err := GetPublicIP()
	if err != nil {
		panic(err)
	}
	origins = append(origins, "/ip4/"+publicIp+"/tcp/6745/p2p/"+ln.Node.Host.ID().String())
	return origins
}

func GetPublicIP() (string, error) {
	resp, err := http.Get("https://ifconfig.me") // important to get the public ip if possible.
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (ln *LightNode) ConnectToDelegates(ctx context.Context, delegates []string) error {
	peers := make(map[peer.ID][]multiaddr.Multiaddr)
	for _, d := range delegates {
		ai, err := peer.AddrInfoFromString(d)
		if err != nil {
			return err
		}

		peers[ai.ID] = append(peers[ai.ID], ai.Addrs...)
	}

	for p, addrs := range peers {
		ln.Node.Host.Peerstore().AddAddrs(p, addrs, time.Hour)

		if ln.Node.Host.Network().Connectedness(p) != network.Connected {
			if err := ln.Node.Host.Connect(ctx, peer.AddrInfo{
				ID: p,
			}); err != nil {
				return err
			}

			ln.Node.Host.ConnManager().Protect(p, "pinning")
		}
	}

	return nil
}

func GetHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}

func ScanHostComputeResources(ln *LightNode, repo string) error {

	memStats := &runtime.MemStats{}
	runtime.ReadMemStats(memStats)
	totalMemory := memStats.Sys
	fmt.Printf("Total memory: %v bytes\n", totalMemory)

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("Total memory: %v bytes\n", m.Alloc)
	fmt.Printf("Total system memory: %v bytes\n", m.Sys)
	fmt.Printf("Total heap memory: %v bytes\n", m.HeapSys)
	fmt.Printf("Heap in use: %v bytes\n", m.HeapInuse)
	fmt.Printf("Stack in use: %v bytes\n", m.StackInuse)
	// get the 80% of the total disk usage
	var stat syscall.Statfs_t
	syscall.Statfs(repo, &stat) // blockstore size
	totalStorage := stat.Blocks * uint64(stat.Bsize)
	fmt.Println("Total storage: ", totalStorage)
	// set the number of CPUs
	numCPU := runtime.NumCPU()
	fmt.Printf("Total number of CPUs: %d\n", numCPU)
	fmt.Printf("Number of CPUs that this Delta will use: %d\n", numCPU/(1200/1000))
	runtime.GOMAXPROCS(numCPU / (1200 / 1000))

	return nil

}
