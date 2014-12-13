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
	"strconv"
	"time"
)

const (
	zkTimeout           = 30 * time.Second // zk connect timeout
	zkNodeDelayChild    = 3 * time.Second  // zk node delay get children
	rpcClientPingSleep  = 1 * time.Second  // rpc client ping need sleep
	rpcClientRetrySleep = 1 * time.Second  // rpc client retry connect need sleep

	RPCPing   = "SnowflakeRPC.Ping"
	RPCNextId = "SnowflakeRPC.NextId"
)

var (
	ErrNoRpcClient = errors.New("rpc: no rpc client service")

	zkConn *zk.Conn // zk connect
)

// Client is gosnowfalke client.
type Client struct {
	zkServers []string
	zkPath    string
	clientMap map[string][]*rpc.Client // key is workerId
}

// NewClient new a gosnowfalke client.
func NewClient(zkServers []string, zkPath string) *Client {
	return &Client{
		zkServers: zkServers,
		zkPath:    zkPath,
		clientMap: make(map[string][]*rpc.Client),
	}
}

// Dial connects to an RPC server at the specified network address from zk.
func (c *Client) Dial() (err error) {
	if err = c.initZK(); err != nil {
		log.Error("c.initZK error(%v)", err)
		return err
	}
	c.watchWorkers()
	return nil
}

// Id generate a snowflake id.
func (c *Client) Id(workerId int64) (int64, error) {
	id := int64(0)
	client := c.getRClient(workerId)
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
	for _, clients := range c.clientMap {
		for n, client := range clients {
			if client == nil {
				log.Info("client is nil %d", n)
				continue
			}
			if err := client.Close(); err != nil {
				log.Error("client.Close() error(%v)", err)
			}
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
			log.Info("zk connect get a event: %s", event.Type.String())
		}
	}()
	return nil
}

// getRClient get a rand rpc client.
func (c *Client) getRClient(workerId int64) *rpc.Client {
	clients, ok := c.clientMap[strconv.FormatInt(workerId, 10)]
	if ok && len(clients) > 0 {
		n := rand.Intn(len(clients))
		return clients[n]
	}
	return nil
}

// watchWorkers watch workers.
func (c *Client) watchWorkers() {
	workers, _, err := zkConn.Children(c.zkPath)
	if err != nil {
		log.Error("zkConn.Children(%s) error(%v)", c.zkPath, err)
		return
	}
	for _, worker := range workers {
		go c.workerHandler(worker)
	}
}

// workerHandler handle the node change.
func (c *Client) workerHandler(worker string) {
	for {
		rpcs, _, watch, err := zkConn.ChildrenW(path.Join(c.zkPath, worker))
		if err != nil {
			log.Error("zkConn.ChildrenW(%s) error(%v)", path.Join(c.zkPath, worker), err)
			time.Sleep(zkNodeDelaySleep)
			continue
		}
		if len(rpcs) == 0 {
			log.Error("zkConn.ChildrenW(%s) error(%v)", path.Join(c.zkPath, worker), err)
			time.Sleep(zkNodeDelaySleep)
			continue
		}
		// rpcs[0]: first is leader
		bs, _, err := zkConn.Get(path.Join(c.zkPath, worker, rpcs[0]))
		if err != nil {
			log.Error("zkConn.Get(%s) error(%v)", path.Join(c.zkPath, worker, rpcs[0]), err)
			return
		}
		peer := &sf.Peer{}
		err = json.Unmarshal(bs, peer)
		if err != nil {
			log.Error("json.Unmarshal(%s, peer) error(%v)", string(bs), err)
			return
		}
		tmpClients := make([]*rpc.Client, len(peer.RPC))
		stop := make(chan bool)
		var clt *rpc.Client
		for _, addr := range peer.RPC {
			if clt, err = rpc.Dial("tcp", addr); err != nil {
				log.Error("rpc.Dial(tcp, %s) error(%v)", addr, err)
				return
			}
			tmpClients = append(tmpClients, clt)
			go c.pingAndRetry(stop, clt, addr)
		}
		c.clientMap[worker] = tmpClients
		log.Info("worker rpc client initd worker(%s)", worker)
		event := <-watch
		close(stop)
		log.Error("zk node(%s) changed %s", path.Join(c.zkPath, worker), event.Type.String())
	}
}

// pingAndRetry ping the rpc connect and re connect when has an error.
func (c *Client) pingAndRetry(stop <-chan bool, client *rpc.Client, addr string) {
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
		select {
		case <-stop:
			return
		default:
		}
		if !failed {
			if err = client.Call(RPCPing, 0, &status); err != nil {
				log.Error("client.Call(%s) error(%v)", RPCPing, err)
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
		log.Info("client reconnect %s ok", addr)
	}
}
