package v2

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/cometbft/cometbft/votepool"

	"github.com/cometbft/cometbft/libs/bytes"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	cmtpubsub "github.com/cometbft/cometbft/libs/pubsub"
	"github.com/cometbft/cometbft/libs/service"
	cmtsync "github.com/cometbft/cometbft/libs/sync"
	rpcclient "github.com/cometbft/cometbft/rpc/client"
	ctypes "github.com/cometbft/cometbft/rpc/core/types"
	jsonrpcclient "github.com/cometbft/cometbft/rpc/jsonrpc/client"
	rpctypes "github.com/cometbft/cometbft/rpc/jsonrpc/types"
	"github.com/cometbft/cometbft/types"
)

var errNotRunning = errors.New("client is not running. Use .Start() method to start")

// WSEvents is a wrapper around WSClient, which implements EventsClient.
type WSEvents struct {
	service.BaseService
	remote   string
	endpoint string
	ws       *jsonrpcclient.WSClient

	mtx                cmtsync.RWMutex
	subscriptions      map[string]chan ctypes.ResultEvent // query -> chan
	rpcResponseChanMap sync.Map
}

func (w *WSEvents) GetClient() *jsonrpcclient.WSClient {
	w.mtx.Lock()
	defer w.mtx.Unlock()
	return w.ws
}

func newWSEvents(remote, endpoint string) (*WSEvents, error) {
	w := &WSEvents{
		endpoint:      endpoint,
		remote:        remote,
		subscriptions: make(map[string]chan ctypes.ResultEvent),
	}
	w.BaseService = *service.NewBaseService(nil, "WSEvents", w)

	var err error
	w.ws, err = jsonrpcclient.NewWS(w.remote, w.endpoint, jsonrpcclient.OnReconnect(func() {
		// resubscribe immediately
		w.redoSubscriptionsAfter(0 * time.Second)
	}))
	if err != nil {
		return nil, err
	}
	w.ws.SetLogger(w.Logger)
	return w, nil
}

// OnStart implements service.Service by starting WSClient and event loop.
func (w *WSEvents) OnStart() error {
	if err := w.GetClient().Start(); err != nil {
		return err
	}

	go w.eventListener()

	return nil
}

// OnStop implements service.Service by stopping WSClient.
func (w *WSEvents) OnStop() {
	if err := w.ws.Stop(); err != nil {
		w.Logger.Error("Can't stop ws client", "err", err)
	}
}

// Subscribe implements EventsClient by using WSClient to subscribe given
// subscriber to query. By default, returns a channel with cap=1. Error is
// returned if it fails to subscribe.
//
// Channel is never closed to prevent clients from seeing an erroneous event.
//
// It returns an error if WSEvents is not running.
func (w *WSEvents) Subscribe(ctx context.Context, subscriber, query string,
	outCapacity ...int) (out <-chan ctypes.ResultEvent, err error) {

	if !w.IsRunning() {
		return nil, errNotRunning
	}

	if err := w.GetClient().Subscribe(ctx, query); err != nil {
		return nil, err
	}

	outCap := 1
	if len(outCapacity) > 0 {
		outCap = outCapacity[0]
	}

	outc := make(chan ctypes.ResultEvent, outCap)
	w.mtx.Lock()
	// subscriber param is ignored because CometBFT will override it with
	// remote IP anyway.
	w.subscriptions[query] = outc
	w.mtx.Unlock()

	return outc, nil
}

// Unsubscribe implements EventsClient by using WSClient to unsubscribe given
// subscriber from query.
//
// It returns an error if WSEvents is not running.
func (w *WSEvents) Unsubscribe(ctx context.Context, subscriber, query string) error {
	if !w.IsRunning() {
		return errNotRunning
	}

	if err := w.GetClient().Unsubscribe(ctx, query); err != nil {
		return err
	}

	w.mtx.Lock()
	_, ok := w.subscriptions[query]
	if ok {
		delete(w.subscriptions, query)
	}
	w.mtx.Unlock()

	return nil
}

// UnsubscribeAll implements EventsClient by using WSClient to unsubscribe
// given subscriber from all the queries.
//
// It returns an error if WSEvents is not running.
func (w *WSEvents) UnsubscribeAll(ctx context.Context, subscriber string) error {
	if !w.IsRunning() {
		return errNotRunning
	}

	if err := w.GetClient().UnsubscribeAll(ctx); err != nil {
		return err
	}

	w.mtx.Lock()
	w.subscriptions = make(map[string]chan ctypes.ResultEvent)
	w.mtx.Unlock()

	return nil
}

// After being reconnected, it is necessary to redo subscription to server
// otherwise no data will be automatically received.
func (w *WSEvents) redoSubscriptionsAfter(d time.Duration) {
	time.Sleep(d)

	w.mtx.RLock()
	defer w.mtx.RUnlock()
	for q := range w.subscriptions {
		err := w.GetClient().Subscribe(context.Background(), q)
		if err != nil {
			w.Logger.Error("Failed to resubscribe", "err", err)
		}
	}
}

func isErrAlreadySubscribed(err error) bool {
	return strings.Contains(err.Error(), cmtpubsub.ErrAlreadySubscribed.Error())
}

func (w *WSEvents) eventListener() {
	for {
		select {
		case resp, ok := <-w.GetClient().ResponsesCh:
			if !ok {
				return
			}

			if resp.Error != nil {
				w.Logger.Error("WS error", "err", resp.Error.Error())
				// Error can be ErrAlreadySubscribed or max client (subscriptions per
				// client) reached or CometBFT exited.
				// We can ignore ErrAlreadySubscribed, but need to retry in other
				// cases.
				if !isErrAlreadySubscribed(resp.Error) {
					// Resubscribe after 1 second to give CometBFT time to restart (if
					// crashed).
					w.redoSubscriptionsAfter(1 * time.Second)
				}
				continue
			}

			// check event is a rpc response, otherwise it is a subscription event
			id, ok := resp.ID.(rpctypes.JSONRPCIntID)
			if !ok {
				w.Logger.Error("unexpected request id type")
				continue
			}
			// if received event id not found in rpcResponseChanMap, means it is a subscription event
			if out, ok := w.rpcResponseChanMap.Load(id); ok {
				outChan, ok := out.(chan rpctypes.RPCResponse)
				if !ok {
					w.Logger.Error("unexpected data type in responseChanMap")
					continue
				}
				select {
				case outChan <- resp:
				default:
					w.Logger.Error("wanted to publish response, but out channel is full", "result", resp.Result)
				}
				continue
			}

			// receive subscription event
			result := new(ctypes.ResultEvent)
			err := cmtjson.Unmarshal(resp.Result, result)
			if err != nil {
				w.Logger.Error("failed to unmarshal response", "err", err)
				continue
			}
			w.mtx.RLock()
			if out, ok := w.subscriptions[result.Query]; ok {
				if cap(out) == 0 {
					out <- *result
				} else {
					select {
					case out <- *result:
					default:
						w.Logger.Error("wanted to publish ResultEvent, but out channel is full", "result", result, "query", result.Query)
					}
				}
			}
			w.mtx.RUnlock()
		case <-w.Quit():
			return
		}
	}
}

func (w *WSEvents) SimpleCall(doRPC func(id rpctypes.JSONRPCIntID) error, proto interface{}) error {
	id := w.GetClient().NextRequestID()
	outChan := make(chan rpctypes.RPCResponse, 1)
	w.rpcResponseChanMap.Store(id, outChan)
	defer close(outChan)
	defer w.rpcResponseChanMap.Delete(id)
	if err := doRPC(id); err != nil {
		return err
	}
	// how long is the timeout
	waitCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	return w.WaitForResponse(waitCtx, outChan, proto)
}

func (w *WSEvents) WaitForResponse(ctx context.Context, outChan chan rpctypes.RPCResponse, result interface{}) error {
	select {
	case resp, ok := <-outChan:
		if !ok {
			return errors.New("response channel is closed")
		}
		if resp.Error != nil {
			return resp.Error
		}
		err := cmtjson.Unmarshal(resp.Result, result)
		if err != nil {
			return err
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (w *WSEvents) Status(ctx context.Context) (*ctypes.ResultStatus, error) {
	result := new(ctypes.ResultStatus)
	wsClient := w.GetClient()
	err := w.SimpleCall(func(id rpctypes.JSONRPCIntID) error {
		return wsClient.Status(ctx, id)
	}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *WSEvents) ABCIInfo(ctx context.Context) (*ctypes.ResultABCIInfo, error) {
	result := new(ctypes.ResultABCIInfo)
	wsClient := w.GetClient()
	err := w.SimpleCall(func(id rpctypes.JSONRPCIntID) error {
		return wsClient.ABCIInfo(ctx, id)
	}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *WSEvents) ABCIQuery(
	ctx context.Context,
	path string,
	data bytes.HexBytes) (*ctypes.ResultABCIQuery, error) {
	result := new(ctypes.ResultABCIQuery)
	wsClient := w.GetClient()
	err := w.SimpleCall(func(id rpctypes.JSONRPCIntID) error {
		return wsClient.ABCIQueryWithOptions(ctx, id, path, data, rpcclient.DefaultABCIQueryOptions)
	}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *WSEvents) ABCIQueryWithOptions(
	ctx context.Context,
	path string,
	data bytes.HexBytes,
	opts rpcclient.ABCIQueryOptions) (*ctypes.ResultABCIQuery, error) {
	result := new(ctypes.ResultABCIQuery)
	wsClient := w.GetClient()
	err := w.SimpleCall(func(id rpctypes.JSONRPCIntID) error {
		return wsClient.ABCIQueryWithOptions(ctx, id, path, data, opts)
	}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *WSEvents) BroadcastTxCommit(ctx context.Context, tx types.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	result := new(ctypes.ResultBroadcastTxCommit)
	wsClient := w.GetClient()
	err := w.SimpleCall(func(id rpctypes.JSONRPCIntID) error {
		return wsClient.BroadcastTxCommit(ctx, id, tx)
	}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *WSEvents) BroadcastTxAsync(ctx context.Context, tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	result := new(ctypes.ResultBroadcastTx)
	wsClient := w.GetClient()
	err := w.SimpleCall(func(id rpctypes.JSONRPCIntID) error {
		return wsClient.BroadcastTxAsync(ctx, id, tx)
	}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *WSEvents) BroadcastTxSync(ctx context.Context, tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	result := new(ctypes.ResultBroadcastTx)
	wsClient := w.GetClient()
	err := w.SimpleCall(func(id rpctypes.JSONRPCIntID) error {
		return wsClient.BroadcastTxSync(ctx, id, tx)
	}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *WSEvents) UnconfirmedTxs(
	ctx context.Context,
	limit *int,
) (*ctypes.ResultUnconfirmedTxs, error) {
	result := new(ctypes.ResultUnconfirmedTxs)
	wsClient := w.GetClient()
	err := w.SimpleCall(func(id rpctypes.JSONRPCIntID) error {
		return wsClient.UnconfirmedTxs(ctx, id, limit)
	}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
func (w *WSEvents) NumUnconfirmedTxs(ctx context.Context) (*ctypes.ResultUnconfirmedTxs, error) {
	result := new(ctypes.ResultUnconfirmedTxs)
	wsClient := w.GetClient()
	err := w.SimpleCall(func(id rpctypes.JSONRPCIntID) error {
		return wsClient.NumUnconfirmedTxs(ctx, id)
	}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *WSEvents) CheckTx(ctx context.Context, tx types.Tx) (*ctypes.ResultCheckTx, error) {
	result := new(ctypes.ResultCheckTx)
	wsClient := w.GetClient()
	err := w.SimpleCall(func(id rpctypes.JSONRPCIntID) error {
		return wsClient.CheckTx(ctx, id, tx)
	}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *WSEvents) NetInfo(ctx context.Context) (*ctypes.ResultNetInfo, error) {
	result := new(ctypes.ResultNetInfo)
	wsClient := w.GetClient()
	err := w.SimpleCall(func(id rpctypes.JSONRPCIntID) error {
		return wsClient.NetInfo(ctx, id)
	}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *WSEvents) DumpConsensusState(ctx context.Context) (*ctypes.ResultDumpConsensusState, error) {
	result := new(ctypes.ResultDumpConsensusState)
	wsClient := w.GetClient()
	err := w.SimpleCall(func(id rpctypes.JSONRPCIntID) error {
		return wsClient.DumpConsensusState(ctx, id)
	}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *WSEvents) ConsensusState(ctx context.Context) (*ctypes.ResultConsensusState, error) {
	result := new(ctypes.ResultConsensusState)
	wsClient := w.GetClient()
	err := w.SimpleCall(func(id rpctypes.JSONRPCIntID) error {
		return wsClient.ConsensusState(ctx, id)
	}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *WSEvents) ConsensusParams(
	ctx context.Context,
	height *int64,
) (*ctypes.ResultConsensusParams, error) {
	result := new(ctypes.ResultConsensusParams)
	wsClient := w.GetClient()
	err := w.SimpleCall(func(id rpctypes.JSONRPCIntID) error {
		return wsClient.ConsensusParams(ctx, id, height)
	}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *WSEvents) Health(ctx context.Context) (*ctypes.ResultHealth, error) {
	result := new(ctypes.ResultHealth)
	wsClient := w.GetClient()
	err := w.SimpleCall(func(id rpctypes.JSONRPCIntID) error {
		return wsClient.Health(ctx, id)
	}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *WSEvents) BlockchainInfo(
	ctx context.Context,
	minHeight,
	maxHeight int64,
) (*ctypes.ResultBlockchainInfo, error) {
	result := new(ctypes.ResultBlockchainInfo)
	wsClient := w.GetClient()
	err := w.SimpleCall(func(id rpctypes.JSONRPCIntID) error {
		return wsClient.BlockchainInfo(ctx, id, minHeight, maxHeight)
	}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *WSEvents) Genesis(ctx context.Context) (*ctypes.ResultGenesis, error) {
	result := new(ctypes.ResultGenesis)
	wsClient := w.GetClient()
	err := w.SimpleCall(func(id rpctypes.JSONRPCIntID) error {
		return wsClient.NetInfo(ctx, id)
	}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *WSEvents) GenesisChunked(ctx context.Context, id uint) (*ctypes.ResultGenesisChunk, error) {
	result := new(ctypes.ResultGenesisChunk)
	wsClient := w.GetClient()
	err := w.SimpleCall(func(id rpctypes.JSONRPCIntID) error {
		return wsClient.GenesisChunked(ctx, id)
	}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *WSEvents) Block(ctx context.Context, height *int64) (*ctypes.ResultBlock, error) {
	result := new(ctypes.ResultBlock)
	wsClient := w.GetClient()
	err := w.SimpleCall(func(id rpctypes.JSONRPCIntID) error {
		return wsClient.Block(ctx, id, height)
	}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *WSEvents) BlockByHash(ctx context.Context, hash []byte) (*ctypes.ResultBlock, error) {
	result := new(ctypes.ResultBlock)
	wsClient := w.GetClient()
	err := w.SimpleCall(func(id rpctypes.JSONRPCIntID) error {
		return wsClient.BlockByHash(ctx, hash, id)
	}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *WSEvents) BlockResults(
	ctx context.Context,
	height *int64,
) (*ctypes.ResultBlockResults, error) {
	result := new(ctypes.ResultBlockResults)
	wsClient := w.GetClient()
	err := w.SimpleCall(func(id rpctypes.JSONRPCIntID) error {
		return wsClient.BlockResults(ctx, height, id)
	}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *WSEvents) Header(ctx context.Context, height *int64) (*ctypes.ResultHeader, error) {
	result := new(ctypes.ResultHeader)
	wsClient := w.GetClient()
	err := w.SimpleCall(func(id rpctypes.JSONRPCIntID) error {
		return wsClient.Header(ctx, height, id)
	}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *WSEvents) HeaderByHash(ctx context.Context, hash bytes.HexBytes) (*ctypes.ResultHeader, error) {
	result := new(ctypes.ResultHeader)
	wsClient := w.GetClient()
	err := w.SimpleCall(func(id rpctypes.JSONRPCIntID) error {
		return wsClient.HeaderByHash(ctx, hash, id)
	}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *WSEvents) Commit(ctx context.Context, height *int64) (*ctypes.ResultCommit, error) {
	result := new(ctypes.ResultCommit)
	wsClient := w.GetClient()
	err := w.SimpleCall(func(id rpctypes.JSONRPCIntID) error {
		return wsClient.Commit(ctx, height, id)
	}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *WSEvents) Tx(ctx context.Context, hash []byte, prove bool) (*ctypes.ResultTx, error) {
	result := new(ctypes.ResultTx)
	wsClient := w.GetClient()
	err := w.SimpleCall(func(id rpctypes.JSONRPCIntID) error {
		return wsClient.Tx(ctx, hash, prove, id)
	}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *WSEvents) TxSearch(
	ctx context.Context,
	query string,
	prove bool,
	page,
	perPage *int,
	orderBy string,
) (*ctypes.ResultTxSearch, error) {

	result := new(ctypes.ResultTxSearch)
	wsClient := w.GetClient()
	err := w.SimpleCall(func(id rpctypes.JSONRPCIntID) error {
		return wsClient.TxSearch(ctx, query, prove, page, perPage, orderBy, id)
	}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *WSEvents) BlockSearch(
	ctx context.Context,
	query string,
	page, perPage *int,
	orderBy string,
) (*ctypes.ResultBlockSearch, error) {
	result := new(ctypes.ResultBlockSearch)
	wsClient := w.GetClient()
	err := w.SimpleCall(func(id rpctypes.JSONRPCIntID) error {
		return wsClient.BlockSearch(ctx, query, page, perPage, orderBy, id)
	}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *WSEvents) Validators(
	ctx context.Context,
	height *int64,
	page,
	perPage *int,
) (*ctypes.ResultValidators, error) {
	result := new(ctypes.ResultValidators)
	wsClient := w.GetClient()
	err := w.SimpleCall(func(id rpctypes.JSONRPCIntID) error {
		return wsClient.Validators(ctx, height, page, perPage, id)
	}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *WSEvents) BroadcastEvidence(
	ctx context.Context,
	ev types.Evidence,
) (*ctypes.ResultBroadcastEvidence, error) {
	result := new(ctypes.ResultBroadcastEvidence)
	wsClient := w.GetClient()
	err := w.SimpleCall(func(id rpctypes.JSONRPCIntID) error {
		return wsClient.BroadcastEvidence(ctx, ev, id)
	}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (w *WSEvents) BroadcastVote(
	ctx context.Context,
	vote votepool.Vote,
) (*ctypes.ResultBroadcastVote, error) {
	result := new(ctypes.ResultBroadcastVote)
	wsClient := w.GetClient()
	err := w.SimpleCall(func(id rpctypes.JSONRPCIntID) error {
		return wsClient.BroadcastVote(ctx, vote, id)
	}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
func (w *WSEvents) QueryVote(
	ctx context.Context,
	eventType int,
	eventHash []byte,
) (*ctypes.ResultQueryVote, error) {
	result := new(ctypes.ResultQueryVote)
	wsClient := w.GetClient()
	err := w.SimpleCall(func(id rpctypes.JSONRPCIntID) error {
		return wsClient.QueryVote(ctx, eventType, eventHash, id)
	}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
