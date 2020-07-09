/*
@Time : 2019/3/30 14:09
@Author : yanKoo
@File : GRPCService
@Software: GoLand
@Description: grcp client连接池
*/
package grpc_pool

import (
	"context"
	"errors"
	"google.golang.org/grpc"
	"sync"
	"time"
)


var (
	// ErrClosed is the error when the client pool is closed
	ErrClosed = errors.New("grpc pool: client pool is closed")
	// ErrTimeout is the error when the client pool timed out
	ErrTimeout = errors.New("grpc pool: client pool timed out")
	// ErrAlreadyClosed is the error when the client conn was already closed
	ErrAlreadyClosed = errors.New("grpc pool: the connection was already closed")
	// ErrFullPool is the error when the pool is already full
	ErrFullPool = errors.New("grpc pool: closing a ClientConn into a full pool")
)

// Factory is a function type creating a grpc client
type Factory func() (*grpc.ClientConn, error)

// Pool is the grpc client pool
type Pool struct {
	clients         chan ClientConn
	factory         Factory
	idleTimeout     time.Duration
	maxLifeDuration time.Duration
	mu              sync.RWMutex
}

// ClientConn is the wrapper for a grpc client conn
type ClientConn struct {
	*grpc.ClientConn
	pool          *Pool
	timeUsed      time.Time
	timeInitiated time.Time
	unhealthy     bool
}

// New creates a new clients pool with the given initial amd maximum capacity,
// and the timeout for the idle clients. Returns an error if the initial
// clients could not be created
func New(factory Factory, init, capacity int, idleTimeout time.Duration,
	maxLifeDuration ...time.Duration) (*Pool, error) {

	if capacity <= 0 {
		capacity = 1
	}
	if init < 0 {
		init = 0
	}
	if init > capacity {
		init = capacity
	}
	p := &Pool{
		clients:     make(chan ClientConn, capacity),
		factory:     factory,
		idleTimeout: idleTimeout,
	}
	if len(maxLifeDuration) > 0 {
		p.maxLifeDuration = maxLifeDuration[0]
	}
	for i := 0; i < init; i++ {
		c, err := factory()
		if err != nil {
			return nil, err
		}

		p.clients <- ClientConn{
			ClientConn:    c,
			pool:          p,
			timeUsed:      time.Now(),
			timeInitiated: time.Now(),
		}
	}
	// Fill the rest of the pool with empty clients
	for i := 0; i < capacity-init; i++ {
		p.clients <- ClientConn{
			pool: p,
		}
	}
	return p, nil
}

func (p *Pool) getClients() chan ClientConn {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.clients
}

// Close empties the pool calling Close on all its clients.
// You can call Close while there are outstanding clients.
// It waits for all clients to be returned (Close).
// The pool channel is then closed, and Get will not be allowed anymore
func (p *Pool) Close() {
	p.mu.Lock()
	clients := p.clients
	p.clients = nil
	p.mu.Unlock()

	if clients == nil {
		return
	}

	close(clients)
	for i := 0; i < p.Capacity(); i++ {
		client := <-clients
		if client.ClientConn == nil {
			continue
		}
		client.ClientConn.Close()
	}
}

// IsClosed returns true if the client pool is closed.
func (p *Pool) IsClosed() bool {
	return p == nil || p.getClients() == nil
}

// Get will return the next available client. If capacity
// has not been reached, it will create a new one using the factory. Otherwise,
// it will wait till the next client becomes available or a timeout.
// A timeout of 0 is an indefinite wait
func (p *Pool) Get(ctx context.Context) (*ClientConn, error) {
	clients := p.getClients()
	if clients == nil {
		return nil, ErrClosed
	}

	wrapper := ClientConn{
		pool: p,
	}
	select {
	case wrapper = <-clients:
		// All good
	case <-ctx.Done():
		return nil, ErrTimeout
	}

	// If the wrapper was idle too long, close the connection and create a new
	// one. It's safe to assume that there isn't any newer client as the client
	// we fetched is the first in the channel
	idleTimeout := p.idleTimeout
	if wrapper.ClientConn != nil && idleTimeout > 0 &&
		wrapper.timeUsed.Add(idleTimeout).Before(time.Now()) {

		wrapper.ClientConn.Close()
		wrapper.ClientConn = nil
	}

	var err error
	if wrapper.ClientConn == nil {
		wrapper.ClientConn, err = p.factory()
		if err != nil {
			// If there was an error, we want to put back a placeholder
			// client in the channel
			clients <- ClientConn{
				pool: p,
			}
		}
		// This is a new connection, reset its initiated time
		wrapper.timeInitiated = time.Now()
	}

	return &wrapper, err
}

// Unhealthy marks the client conn as unhealthy, so that the connection
// gets reset when closed
func (c *ClientConn) Unhealthy() {
	c.unhealthy = true
}

// Close returns a ClientConn to the pool. It is safe to call multiple time,
// but will return an error after first time
func (c *ClientConn) Close() error {
	if c == nil {
		return nil
	}
	if c.ClientConn == nil {
		return ErrAlreadyClosed
	}
	if c.pool.IsClosed() {
		return ErrClosed
	}
	// If the wrapper connection has become too old, we want to recycle it. To
	// clarify the logic: if the sum of the initialization time and the max
	// duration is before Now(), it means the initialization is so old adding
	// the maximum duration couldn't put in the future. This sum therefore
	// corresponds to the cut-off point: if it's in the future we still have
	// time, if it's in the past it's too old
	maxDuration := c.pool.maxLifeDuration
	if maxDuration > 0 && c.timeInitiated.Add(maxDuration).Before(time.Now()) {
		c.Unhealthy()
	}

	// We're cloning the wrapper so we can set ClientConn to nil in the one
	// used by the user
	wrapper := ClientConn{
		pool:       c.pool,
		ClientConn: c.ClientConn,
		timeUsed:   time.Now(),
	}
	if c.unhealthy {
		wrapper.ClientConn.Close()
		wrapper.ClientConn = nil
	} else {
		wrapper.timeInitiated = c.timeInitiated
	}
	select {
	case c.pool.clients <- wrapper:
		// All good
	default:
		return ErrFullPool
	}

	c.ClientConn = nil // Mark as closed
	return nil
}

// Capacity returns the capacity
func (p *Pool) Capacity() int {
	if p.IsClosed() {
		return 0
	}
	return cap(p.clients)
}

// Available returns the number of currently unused clients
func (p *Pool) Available() int {
	if p.IsClosed() {
		return 0
	}
	return len(p.clients)
}


//
//import (
//	"context"
//	"fmt"
//	"sync"
//	"time"
//
//	"google.golang.org/grpc"
//	"google.golang.org/grpc/connectivity"
//)
//
//const (
//	defaultTimeout    = 100 * time.Second
//	checkReadyTimeout = 5 * time.Second
//	heartbeatInterval = 20 * time.Second
//)
//
//var (
//	errNoReady = fmt.Errorf("no ready")
//)
//
//// DialFunc dial function
//type DialFunc func(addr string) (*grpc.ClientConn, error)
//
//// ReadyCheckFunc check conn is ready function
//type ReadyCheckFunc func(ctx context.Context, conn *grpc.ClientConn) connectivity.State
//
//// ConnectionTracker keep connections and maintain their status
//type ConnectionTracker struct {
//	sync.RWMutex
//	dial              DialFunc
//	readyCheck        ReadyCheckFunc
//	connections       map[string]*trackedConn
//	alives            map[string]*trackedConn
//	timeout           time.Duration
//	checkReadyTimeout time.Duration
//	heartbeatInterval time.Duration
//
//	ctx    context.Context
//	cannel context.CancelFunc
//}
//
//// TrackerOption initialization options
//type TrackerOption func(*ConnectionTracker)
//
//// SetTimeout custom timeout
//func SetTimeout(timeout time.Duration) TrackerOption {
//	return func(o *ConnectionTracker) {
//		o.timeout = timeout
//	}
//}
//
//// SetCheckReadyTimeout custom checkReadyTimeout
//func SetCheckReadyTimeout(timeout time.Duration) TrackerOption {
//	return func(o *ConnectionTracker) {
//		o.checkReadyTimeout = timeout
//	}
//}
//
//// SetHeartbeatInterval custom heartbeatInterval
//func SetHeartbeatInterval(interval time.Duration) TrackerOption {
//	return func(o *ConnectionTracker) {
//		o.heartbeatInterval = interval
//	}
//}
//
//// CustomReadyCheck custom ready check function
//func CustomReadyCheck(f ReadyCheckFunc) TrackerOption {
//	return func(o *ConnectionTracker) {
//		o.readyCheck = f
//	}
//}
//
//// New initialization ConnectionTracker
//func New(dial DialFunc, opts ...TrackerOption) *ConnectionTracker {
//	ctx, cannel := context.WithCancel(context.Background())
//	ct := &ConnectionTracker{
//		dial:              dial,
//		readyCheck:        defaultReadyCheck,
//		connections:       make(map[string]*trackedConn),
//		alives:            make(map[string]*trackedConn),
//		timeout:           defaultTimeout,
//		checkReadyTimeout: checkReadyTimeout,
//		heartbeatInterval: heartbeatInterval,
//
//		ctx:    ctx,
//		cannel: cannel,
//	}
//
//	for _, opt := range opts {
//		opt(ct)
//	}
//
//	return ct
//}
//
//// GetConn create or get an existing connection
//func (ct *ConnectionTracker) GetConn(addr string) (*grpc.ClientConn, error) {
//	return ct.getConn(addr, false)
//}
//
//// Dial force to create new connection, this operation will close old connection!
//func (ct *ConnectionTracker) Dial(addr string) (*grpc.ClientConn, error) {
//	return ct.getConn(addr, true)
//}
//
//func (ct *ConnectionTracker) getConn(addr string, force bool) (*grpc.ClientConn, error) {
//	ct.Lock()
//	tc, ok := ct.connections[addr]
//	if !ok {
//		tc = &trackedConn{
//			addr:    addr,
//			tracker: ct,
//		}
//		ct.connections[addr] = tc
//	}
//	ct.Unlock()
//
//	err := tc.tryconn(ct.ctx, force)
//	if err != nil {
//		return nil, err
//	}
//	return tc.conn, nil
//}
//
//func (ct *ConnectionTracker) connReady(tc *trackedConn) {
//	ct.Lock()
//	defer ct.Unlock()
//	ct.alives[tc.addr] = tc
//}
//
//func (ct *ConnectionTracker) connUnReady(addr string) {
//	ct.Lock()
//	defer ct.Unlock()
//	delete(ct.alives, addr)
//}
//
//// Alives current live connections
//func (ct *ConnectionTracker) Alives() []string {
//	ct.RLock()
//	defer ct.RUnlock()
//	alives := []string{}
//	for addr := range ct.alives {
//		alives = append(alives, addr)
//	}
//	return alives
//}
//
//type trackedConn struct {
//	sync.RWMutex
//	addr    string
//	conn    *grpc.ClientConn
//	tracker *ConnectionTracker
//	state   connectivity.State
//	expires time.Time
//	retry   int
//	cannel  context.CancelFunc
//}
//
//func (tc *trackedConn) tryconn(ctx context.Context, force bool) error {
//	tc.Lock()
//	defer tc.Unlock()
//	if !force && tc.conn != nil { // another goroutine got the write lock first
//		if tc.state == connectivity.Ready {
//			return nil
//		}
//		if tc.state == connectivity.Idle {
//			return errNoReady
//		}
//	}
//
//	if tc.conn != nil { // close shutdown conn
//		tc.conn.Close()
//	}
//	conn, err := tc.tracker.dial(tc.addr)
//	if err != nil {
//		return err
//	}
//	tc.conn = conn
//
//	readyCtx, cancel := context.WithTimeout(ctx, tc.tracker.checkReadyTimeout)
//	defer cancel()
//
//	checkStatus := tc.tracker.readyCheck(readyCtx, tc.conn)
//	hbCtx, hbCancel := context.WithCancel(ctx)
//	tc.cannel = hbCancel
//	go tc.heartbeat(hbCtx)
//
//	if checkStatus != connectivity.Ready {
//		return errNoReady
//	}
//	tc.ready()
//	return nil
//}
//
//func (tc *trackedConn) getState() connectivity.State {
//	tc.RLock()
//	defer tc.RUnlock()
//	return tc.state
//}
//
//func (tc *trackedConn) healthCheck(ctx context.Context) {
//	tc.Lock()
//	defer tc.Unlock()
//	ctx, cancel := context.WithTimeout(ctx, tc.tracker.checkReadyTimeout)
//	defer cancel()
//
//	switch tc.tracker.readyCheck(ctx, tc.conn) {
//	case connectivity.Ready:
//		tc.ready()
//	case connectivity.Shutdown:
//		tc.shutdown()
//	case connectivity.Idle:
//		if tc.expired() {
//			tc.shutdown()
//		} else {
//			tc.idle()
//		}
//	}
//}
//
//func defaultReadyCheck(ctx context.Context, conn *grpc.ClientConn) connectivity.State {
//	for {
//		s := conn.GetState()
//		if s == connectivity.Ready || s == connectivity.Shutdown {
//			return s
//		}
//		if !conn.WaitForStateChange(ctx, s) {
//			return connectivity.Idle
//		}
//	}
//}
//
//func (tc *trackedConn) ready() {
//	tc.state = connectivity.Ready
//	tc.expires = time.Now().Add(tc.tracker.timeout)
//	tc.retry = 0
//	tc.tracker.connReady(tc)
//}
//
//func (tc *trackedConn) idle() {
//	tc.state = connectivity.Idle
//	tc.retry++
//	tc.tracker.connUnReady(tc.addr)
//}
//
//func (tc *trackedConn) shutdown() {
//	tc.state = connectivity.Shutdown
//	tc.conn.Close()
//	tc.cannel()
//	tc.tracker.connUnReady(tc.addr)
//}
//
//func (tc *trackedConn) expired() bool {
//	return tc.expires.Before(time.Now())
//}
//
//func (tc *trackedConn) heartbeat(ctx context.Context) {
//	ticker := time.NewTicker(tc.tracker.heartbeatInterval)
//	for tc.getState() != connectivity.Shutdown {
//		select {
//		case <-ctx.Done():
//			tc.shutdown()
//			break
//		case <-ticker.C:
//			tc.healthCheck(ctx)
//		}
//	}
//}
//
//var (
//	// defaultPool default pool
//	defaultPool *ConnectionTracker
//	once        sync.Once
//	dialF       = func(addr string) (*grpc.ClientConn, error) {
//		return grpc.Dial(
//			addr,
//			grpc.WithInsecure(),
//		)
//	}
//)
//
//func pool() *ConnectionTracker {
//	once.Do(func() {
//		defaultPool = New(dialF)
//	})
//	return defaultPool
//}
//
//// GetConn create or get an existing connection from default pool
//func GetConn(addr string) (*grpc.ClientConn, error) {
//	return pool().GetConn(addr)
//}
//
//// Dial force to create new connection from default pool, this operation will close old connection!
//func Dial(addr string) (*grpc.ClientConn, error) {
//	return pool().Dial(addr)
//}
//
//// Alives current live connections from default pool
//func Alives() []string {
//	return pool().Alives()
//}
