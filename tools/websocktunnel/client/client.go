package client

import (
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/taskcluster/taskcluster/v29/tools/websocktunnel/util"
	"github.com/taskcluster/taskcluster/v29/tools/websocktunnel/wsmux"
)

type clientState int

const (
	stateRunning = iota
	stateBroken
	stateClosed
)

// Config contains the configuration for a Client.  This is generated by
// a Configurer.
type Config struct {
	// The client ID to register
	ID string

	// The address of the websocktunnel server (https:// or wss://)
	TunnelAddr string

	// The JWT to authenticate to the websocktunnel server.  This should be
	// a "fresh" token for each call to the Configurer.
	Token string

	// Configuration for retrying connections to the server
	Retry RetryConfig

	// A Logger for logging status updates; default is no logging
	Logger util.Logger
}

// Configurer is a function which can generate a Config object to be used by
// the client.  This is called whenever a reconnection is made, and should
// return a Config with a "fresh" token at that time.
type Configurer func() (Config, error)

// Client is used to connect to a websocktunnel instance and serve content over
// the tunnel.  Client implements net.Listener.
type Client struct {
	m          sync.Mutex
	id         string
	tunnelAddr string
	token      string
	url        atomic.Value
	retry      RetryConfig
	logger     util.Logger
	configurer Configurer
	session    *wsmux.Session
	state      clientState
	closed     chan struct{}
	acceptErr  net.Error
}

// New creates a new Client instance.
func New(configurer Configurer) (*Client, error) {
	config, err := configurer()
	if err != nil {
		return nil, err
	}

	cl := &Client{configurer: configurer}
	cl.setConfig(config)
	cl.closed = make(chan struct{}, 1)
	conn, url, err := cl.connectWithRetry()
	if err != nil {
		return nil, err
	}
	cl.url.Store(url)
	cl.session = wsmux.Client(conn, wsmux.Config{})
	return cl, nil
}

// URL returns the url at which the websocktunnel server serves the client's
// endpoints.  Users should use this value to create URLs (by appending) for
// viewers to access the client.
func (c *Client) URL() string {
	return c.url.Load().(string)
}

// Accept is used to accept multiplexed streams from the tunnel as a net.Conn
// implementer.
//
// This is a net.Listener interface method.
func (c *Client) Accept() (net.Conn, error) {
	select {
	case <-c.closed:
		return nil, ErrClientClosed
	default:
	}

	c.m.Lock()
	defer c.m.Unlock()
	if c.state == stateBroken || c.state == stateClosed {
		return nil, c.acceptErr
	}

	stream, err := c.session.Accept()
	if err != nil {
		c.state = stateBroken
		c.acceptErr = ErrClientReconnecting
		go c.reconnect()
		return nil, c.acceptErr
	}
	return stream, nil
}

// Addr returns the net.Addr of the underlying wsmux session
//
// This is a net.Listener method.  Its return value in this case is
// not especially useful.
func (c *Client) Addr() net.Addr {
	return c.session.Addr()
}

// Close connection to the tunnel.
//
// This is a net.Listener method.
func (c *Client) Close() error {
	select {
	case <-c.closed:
		return nil
	default:
		close(c.closed)
		go func() {
			c.m.Lock()
			defer c.m.Unlock()
			c.acceptErr = ErrClientClosed
			_ = c.session.Close()
		}()
	}
	return nil
}

func (c *Client) setConfig(config Config) {
	c.id = config.ID
	c.tunnelAddr = util.MakeWsURL(config.TunnelAddr)
	c.token = config.Token

	c.retry = config.Retry.withDefaultValues()
	c.logger = config.Logger
	if c.logger == nil {
		c.logger = &util.NilLogger{}
	}
}

// connectWithRetry returns a websocket connection to the tunnel
func (c *Client) connectWithRetry() (*websocket.Conn, string, error) {
	// if token is expired or not usable, get a new token from the authorizer
	if !util.IsTokenUsable(c.token) {
		config, err := c.configurer()
		if err != nil {
			return nil, "", err
		}
		c.setConfig(config)
	}

	header := make(http.Header)
	header.Set("Authorization", "Bearer "+c.token)
	header.Set("x-websocktunnel-id", c.id)

	currentDelay := c.retry.InitialDelay
	maxTimer := time.After(c.retry.MaxElapsedTime)
	backoff := time.After(currentDelay)

	for {
		c.logger.Printf("trying to connect to %s", c.tunnelAddr)
		conn, res, err := websocket.DefaultDialer.Dial(c.tunnelAddr, header)
		if err == nil {
			c.logger.Printf("connected to %s ", c.tunnelAddr)
			url := res.Header.Get("x-websocktunnel-client-url")
			return conn, url, err
		}

		if !shouldRetry(res) {
			c.logger.Printf("connection failed with error:%v, response:%v", err, res)
			if isAuthError(res) {
				return nil, "", ErrAuthFailed
			}
			return nil, "", ErrRetryFailed
		}
		c.logger.Printf("connection to %s failed -- retrying.", c.tunnelAddr)

		// wait for the next time to try connecting
		select {
		case <-maxTimer:
			return nil, "", ErrRetryTimedOut
		case <-backoff:
			c.logger.Printf("trying to connect to %s", c.tunnelAddr)
			conn, res, err := websocket.DefaultDialer.Dial(c.tunnelAddr, header)
			if err == nil {
				url := res.Header.Get("x-websocktunnel-client-url")
				return conn, url, nil
			}
			if !shouldRetry(res) {
				c.logger.Printf("connection to %s failed. could not connect", c.tunnelAddr)
				return nil, "", ErrRetryFailed
			}

			currentDelay = c.retry.nextDelay(currentDelay)
			backoff = time.After(currentDelay)
		}
	}
}

// reconnect is used to repair broken connections
func (c *Client) reconnect() {
	c.m.Lock()
	defer c.m.Unlock()
	conn, url, err := c.connectWithRetry()
	if err != nil {
		// set error and return
		c.logger.Printf("unable to reconnect to %s", c.tunnelAddr)
		c.acceptErr = ErrRetryFailed
		return
	}

	if c.session != nil {
		_ = c.session.Close()
		c.session = nil
	}

	sessionConfig := wsmux.Config{
		// Log:              c.logger,
		StreamBufferSize: 4 * 1024,
	}
	c.session = wsmux.Client(conn, sessionConfig)
	c.url.Store(url)
	c.state = stateRunning
	c.logger.Printf("state: running")
	c.acceptErr = nil

}

// simple utility to check if client should retry connection
func shouldRetry(r *http.Response) bool {
	// retry on connection failures (e.g., server down)
	if r == nil {
		return true
	}
	// retry on anything but 4xx or 2xx responses
	if r.StatusCode/100 != 4 && r.StatusCode/100 != 2 {
		return true
	}
	return false
}

func isAuthError(r *http.Response) bool {
	if r == nil {
		return false
	}
	return r.StatusCode == 401
}
