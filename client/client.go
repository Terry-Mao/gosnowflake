package client

import (
	log "code.google.com/p/log4go"
	"encoding/json"
	"errors"
	sf "github.com/guanguan241/gosnowflake"
	"github.com/samuel/go-zookeeper/zk"
	"math/rand"
	"net/rpc"
	"path"
	"time"
)

const (
	zkTimeout           = 30 * time.Second
	rpcClientPingSleep  = 1 * time.Second // rpc client ping need sleep
	rpcClientRetrySleep = 1 * time.Second // rpc client retry connect need sleep

	RPCPing   = "SnowflakeRPC.Ping"
	RPCNextId = "SnowflakeRPC.NextId"
)

var (
	ErrNoChild      = errors.New("zk: children is nil")
	ErrNodeNotExist = errors.New("zk: node not exist")
	ErrNoRpcClient  = errors.New("rpc: no rpc client service")

	zkConn *zk.Conn // zk connect
)

// Client is gosnowfalke client.
type Client struct {
	reStop    bool
	zkServers []string
	zkPath    string
	clients   []*rpc.Client
}

// NewClient new a gosnowfalke client.
func NewClient(zkServers []string, zkPath string) *Client {
	return &Client{zkServers: zkServers, zkPath: zkPath}
}

// Dial connects to an RPC server at the specified network address from zk.
func (c *Client) Dial() (err error) {
	if err = c.initZK(); err != nil {
		log.Error("c.initZK error(%v)", err)
		return err
	}
	go c.watchNodes()
	return nil
}

// Id generate a snowflake id.
func (c *Client) Id(workerId int64) (int64, error) {
	id := int64(0)
	client := c.getRClient()
	if client == nil {
		return 0, ErrNoRpcClient
	}
	if err := client.Call(RPCNextId, workerId, &id); err != nil {
		log.Error("rpc.Call(%s, %d, &id) error(%v)", RPCNextId, workerId, err)
		return 0, err
	}
	return id, nil
}

// Close close all rpc client.
func (c *Client) Close() {
	c.reStop = true
	for _, client := range c.clients {
		if err := client.Close(); err != nil {
			log.Error("client.Close() error(%v)", err)
		}
	}
}

// initZK init zk connect.
func (c *Client) initZK() error {
	conn, session, err := zk.Connect(c.zkServers, zkTimeout)
	if err != nil {
		log.Error("zk.Connect(\"%v\", %d) error(%v)", c.zkServers, zkTimeout, err)
		return err
	}
	zkConn = conn
	go func() {
		for {
			event := <-session
			log.Info("zookeeper get a event: %s", event.State.String())
		}
	}()
	return nil
}

// getRClient get a rand rpc client.
func (c *Client) getRClient() *rpc.Client {
	if len(c.clients) == 0 {
		return nil
	}
	n := rand.Intn(len(c.clients))
	return c.clients[n]
}

// getChildrenWatch get zk path children and watch event.
func (c *Client) getChildrenWatch() (nodes []string, watch <-chan zk.Event, err error) {
	nodes, stat, watch, err := zkConn.ChildrenW(c.zkPath)
	if err != nil {
		return
	}
	if stat == nil {
		err = ErrNodeNotExist
		return
	}
	if len(nodes) == 0 {
		err = ErrNoChild
		return
	}
	return
}

// watchNodes watch node change.
func (c *Client) watchNodes() {
	for {
		nodes, watch, err := c.getChildrenWatch()
		if err != nil {
			log.Error("c.getChildrenWatch() error(%v)", err)
			continue
		}
		watchPath := path.Join(c.zkPath, nodes[0])
		nodes, _, err = zkConn.Children(watchPath)
		if err != nil {
			log.Error("zkConn.Children(%s) error(%v)", watchPath, err)
			continue
		}
		// first is leader
		// node := nodes[0]
		c.eventHandler(path.Join(watchPath, nodes[0]))
		event := <-watch
		log.Info("zk path: \"%s\" receive a event %v", c.zkPath, event)
	}
}

// eventHandler handle the node change.
func (c *Client) eventHandler(npath string) {
	bs, _, err := zkConn.Get(npath)
	if err != nil {
		log.Error("zkConn.Get(%s) error(%v)", npath, err)
		return
	}
	peer := &sf.Peer{}
	err = json.Unmarshal(bs, peer)
	if err != nil {
		log.Error("json.Unmarshal(%s, peer) error(%v)", string(bs), err)
		return
	}
	var clt *rpc.Client
	for _, addr := range peer.RPC {
		if clt, err = rpc.Dial("tcp", addr); err != nil {
			log.Error("rpc.Dial(tcp, %s) error(%v)", addr, err)
			return
		}
		c.clients = append(c.clients, clt)
		go c.retryClient(clt, addr)
	}
}

// retryClient re connect rpc when has error.
func (c *Client) retryClient(client *rpc.Client, addr string) {
	defer func() {
		if err := client.Close(); err != nil {
			log.Error("client.Close() error(%v)", err)
		}
	}()
	var (
		failed bool
		status int
		err    error
		tmp    *rpc.Client
	)
	for {
		if c.reStop {
			return
		}
		if !failed {
			if err = client.Call(RPCPing, 0, &status); err != nil {
				log.Error("client.Call(ping) error(%v)", err)
				failed = true
				continue
			} else {
				failed = false
				time.Sleep(rpcClientPingSleep)
				continue
			}
		}
		if tmp, err = rpc.Dial("tcp", addr); err != nil {
			log.Error("rpc.Dial(tcp, %s) error(%v)", addr, err)
			time.Sleep(rpcClientRetrySleep)
			continue
		}
		client = tmp
		failed = false
		log.Info("SnowflakeRPC reconnect %s ok", addr)
	}
}
