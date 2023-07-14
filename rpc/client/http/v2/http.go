package v2

import (
	"net/http"
	"time"

	"github.com/cometbft/cometbft/libs/log"
	jsonrpcclient "github.com/cometbft/cometbft/rpc/jsonrpc/client"
)

/*
HTTP is a Client implementation that communicates with a CometBFT node over
JSON RPC and WebSockets.

This is the main implementation you probably want to use in production code.
There are other implementations when calling the CometBFT node in-process
(Local), or when you want to mock out the server for test code (mock).

You can subscribe for any event published by CometBFT using Subscribe method.
Note delivery is best-effort. If you don't read events fast enough or network is
slow, CometBFT might cancel the subscription. The client will attempt to
resubscribe (you don't need to do anything). It will keep trying every second
indefinitely until successful.

Request batching is available for JSON RPC requests over HTTP, which conforms to
the JSON RPC specification (https://www.jsonrpc.org/specification#batch). See
the example for more details.

Example:

	c, err := New("http://192.168.1.10:26657", "/websocket")
	if err != nil {
		// handle error
	}

	// call Start/Stop if you're subscribing to events
	err = c.Start()
	if err != nil {
		// handle error
	}
	defer c.Stop()

	res, err := c.Status()
	if err != nil {
		// handle error
	}

	// handle result
*/
type HTTP struct {
	remote string
	*WSEvents
}

//-----------------------------------------------------------------------------
// HTTP

// New takes a remote endpoint in the form <protocol>://<host>:<port> and
// the websocket path (which always seems to be "/websocket")
// An error is returned on invalid remote. The function panics when remote is nil.
func New(remote, wsEndpoint string) (*HTTP, error) {
	httpClient, err := jsonrpcclient.DefaultHTTPClient(remote)
	if err != nil {
		return nil, err
	}
	return NewWithClient(remote, wsEndpoint, httpClient)
}

// Create timeout enabled http client
func NewWithTimeout(remote, wsEndpoint string, timeout uint) (*HTTP, error) {
	httpClient, err := jsonrpcclient.DefaultHTTPClient(remote)
	if err != nil {
		return nil, err
	}
	httpClient.Timeout = time.Duration(timeout) * time.Second
	return NewWithClient(remote, wsEndpoint, httpClient)
}

// NewWithClient allows for setting a custom http client (See New).
// An error is returned on invalid remote. The function panics when remote is nil.
func NewWithClient(remote, wsEndpoint string, client *http.Client) (*HTTP, error) {
	if client == nil {
		panic("nil http.Client provided")
	}
	wsEvents, err := newWSEvents(remote, wsEndpoint)
	if err != nil {
		return nil, err
	}
	httpClient := &HTTP{
		remote:   remote,
		WSEvents: wsEvents,
	}

	return httpClient, nil
}

// SetLogger sets a logger.
func (c *HTTP) SetLogger(l log.Logger) {
	c.WSEvents.SetLogger(l)
}

// Remote returns the remote network address in a string form.
func (c *HTTP) Remote() string {
	return c.remote
}
