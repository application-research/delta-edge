package core

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"sync"
	"syscall"
	"time"

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
	ipldbasicnode "github.com/ipld/go-ipld-prime/node/basic"
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
	Config *Configuration
}

type LocalWallet struct {
	keys     map[address.Address]*key.Key
	keystore types.KeyStore

	lk sync.Mutex
}

type Configuration struct {
	APINodeAddress string
}

type GatewayHandler struct {
	bs       blockstore.Blockstore
	dserv    mdagipld.DAGService
	resolver resolver.Resolver
	node     *whypfs.Node
}

var defaultTestBootstrapPeers []multiaddr.Multiaddr

// Creating a list of multiaddresses that are used to bootstrap the network.
func BootstrapEstuaryPeers() []peer.AddrInfo {

	for _, s := range []string{
		"/ip4/145.40.90.135/tcp/6746/p2p/12D3KooWNTiHg8eQsTRx8XV7TiJbq3379EgwG6Mo3V3MdwAfThsx",
		"/ip4/139.178.68.217/tcp/6744/p2p/12D3KooWCVXs8P7iq6ao4XhfAmKWrEeuKFWCJgqe9jGDMTqHYBjw",
		"/ip4/147.75.49.71/tcp/6745/p2p/12D3KooWGBWx9gyUFTVQcKMTenQMSyE2ad9m7c9fpjS4NMjoDien",
		"/ip4/147.75.86.255/tcp/6745/p2p/12D3KooWFrnuj5o3tx4fGD2ZVJRyDqTdzGnU3XYXmBbWbc8Hs8Nd",
		"/ip4/3.134.223.177/tcp/6745/p2p/12D3KooWN8vAoGd6eurUSidcpLYguQiGZwt4eVgDvbgaS7kiGTup",
		"/ip4/35.74.45.12/udp/6746/quic/p2p/12D3KooWLV128pddyvoG6NBvoZw7sSrgpMTPtjnpu3mSmENqhtL7",
		"/dnsaddr/bootstrap.libp2p.io/p2p/QmNnooDu7bfjPFoTZYxMNLWUQJyrVwtbZg5gBMjTezGAJN",
		"/dnsaddr/bootstrap.libp2p.io/p2p/QmQCU2EcMqAqQPR2i9bChDtGNJchTbq5TbXJJ16u19uLTa",
		"/dnsaddr/bootstrap.libp2p.io/p2p/QmbLHAnMoJPWSCR5Zhtx6BHJX9KiKNN6tpvbUcqanj75Nb",
		"/dnsaddr/bootstrap.libp2p.io/p2p/QmcZf59bWwK5XFi76CZX8cbJ4BhTzzA3gU1ZjYZcYW3dwt",
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

// NewEdgeNode Add a config to enable gateway or not.
// Add a config to enable content, bucket, commp, replication verifier processor
func NewEdgeNode(ctx context.Context, repo string) (*LightNode, error) {

	db, err := OpenDatabase()
	// node
	publicIp, err := GetPublicIP()
	newConfig := &whypfs.Config{
		ListenAddrs: []string{
			"/ip4/0.0.0.0/tcp/6745",
			"/ip4/" + publicIp + "/tcp/6745",
		},
		AnnounceAddrs: []string{
			"/ip4/0.0.0.0/tcp/6745",
			"/ip4/" + publicIp + "/tcp/6745",
		},
	}
	params := whypfs.NewNodeParams{
		Ctx:       ctx,
		Datastore: whypfs.NewInMemoryDatastore(),
		Repo:      repo,
	}

	params.Config = params.ConfigurationBuilder(newConfig)
	whypfsPeer, err := whypfs.NewNode(params)
	if err != nil {
		panic(err)
	}

	whypfsPeer.BootstrapPeers(BootstrapEstuaryPeers())

	// gateway
	gw, err := NewGatewayHandler(whypfsPeer)

	// create the global light node.
	return &LightNode{
		Node: whypfsPeer,
		Gw:   gw,
		DB:   db,
	}, nil
}

func NewGatewayHandler(node *whypfs.Node) (*GatewayHandler, error) {

	bsvc := blockservice.New(node.Blockstore, nil)
	ipldFetcher := bsfetcher.NewFetcherConfig(bsvc)

	ipldFetcher.PrototypeChooser = dagpb.AddSupportToChooser(func(lnk ipld.Link, lnkCtx ipld.LinkContext) (ipld.NodePrototype, error) {
		if tlnkNd, ok := lnkCtx.LinkNode.(schema.TypedLinkNode); ok {
			return tlnkNd.LinkTargetNodePrototype(), nil
		}
		return ipldbasicnode.Prototype.Any, nil
	})

	resolver := resolver.NewBasicResolver(ipldFetcher.WithReifier(unixfsnode.Reify))
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
