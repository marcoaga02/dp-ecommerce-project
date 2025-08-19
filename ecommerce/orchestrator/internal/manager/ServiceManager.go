package manager

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/marcoaga02/dp-ecommerce-project/ecommerce/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

// ServiceStatus holds the connection and readiness state for one service.
type ServiceStatus struct {
	Conn  *grpc.ClientConn
	Ready bool
}

// ServiceManager monitors multiple gRPC services (health checks) and
// keeps ready ClientConn references available for callers.
type ServiceManager struct {
	services     map[string]*ServiceStatus // name -> status
	addresses    map[string]string         // name -> target address
	logger       logger.Logger
	upInterval   time.Duration
	downInterval time.Duration

	mu     sync.RWMutex
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewServiceManager creates a ServiceManager.
//
// Parameters:
//   - addresses: map serviceName -> grpc target (e.g. "auth:9000")
//   - logger: a logger that implements logger.Logger
//   - upInterval: how often to poll a service that is already READY
//   - downInterval: how often to retry a service that is NOT ready
//
// Returns:
//   - *ServiceManager: the pointer to the new ServiceManager
func NewServiceManager(addresses map[string]string, lg logger.Logger, upInterval, downInterval time.Duration) *ServiceManager {
	services := make(map[string]*ServiceStatus, len(addresses))
	for name := range addresses {
		services[name] = &ServiceStatus{}
	}
	if upInterval <= 0 {
		upInterval = 15 * time.Second
	}
	if downInterval <= 0 {
		downInterval = 2 * time.Second
	}
	return &ServiceManager{
		services:     services,
		addresses:    addresses,
		logger:       lg,
		upInterval:   upInterval,
		downInterval: downInterval,
	}
}

// StartMonitoring starts one monitor goroutine per service.
// It returns immediately; stops when Stop() is called.
func (sm *ServiceManager) StartMonitoring() {
	sm.mu.Lock()
	// if already started, do nothing
	if sm.cancel != nil {
		sm.mu.Unlock()
		sm.logger.Warn("ServiceManager already started")
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	sm.cancel = cancel
	sm.mu.Unlock()

	for name, addr := range sm.addresses {
		sm.wg.Add(1)
		go sm.monitorService(ctx, name, addr)
	}
	sm.logger.Info("ServiceManager started monitoring %d services", len(sm.addresses))
}

// Stop stops all monitoring goroutines and closes all connections.
func (sm *ServiceManager) Stop() {
	sm.mu.Lock()
	if sm.cancel != nil {
		sm.cancel() // all the channels Done associated to the context are closed
		sm.cancel = nil
	}
	sm.mu.Unlock()

	sm.wg.Wait() // all the go routine whose Done channel has been closed, return and this function wait all of them

	// Close all the opened connections
	sm.mu.Lock()
	for name, s := range sm.services {
		if s.Conn != nil {
			_ = s.Conn.Close()
			s.Conn = nil
			s.Ready = false
			sm.logger.Info("Closed connection for service %s", name)
		}
	}
	sm.mu.Unlock()
	sm.logger.Info("ServiceManager stopped")
}

// monitorService runs a loop trying to keep the named service healthy.
// It adjusts sleep interval depending on service state (ready vs not ready).
//
// Parameters:
//   - ctx: context.Context used to signal cancellation and stop the monitoring loop.
//   - name: the logical name of the service (must exist in sm.services).
//   - addr: the target network address of the service in host:port format.
func (sm *ServiceManager) monitorService(ctx context.Context, name, addr string) {
	defer sm.wg.Done()
	sleepInterval := sm.downInterval

	for {
		select {
		case <-ctx.Done():
			sm.logger.Info("monitorService(%s) exiting", name)
			return
		default: // channel Done still alive => cancel() has not been called on the context
		}

		// Try to create a client connection.
		conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			sm.logger.Warn("Attempt to connect to the '%s' service failed: %v", name, err)
			sm.setServiceStatus(name, nil, false)

			sleepInterval = sm.downInterval
			select { // select the first channel that returns
			case <-time.After(sleepInterval):
				continue
			case <-ctx.Done():
				return
			}
		}

		// Perform health check with a 2-second timeout.
		// If the check finishes before the timeout, cancel() is called to release resources.
		// Otherwise, the context times out and the check returns an error.
		healthClient := healthpb.NewHealthClient(conn)
		healthCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		resp, err := healthClient.Check(healthCtx, &healthpb.HealthCheckRequest{})
		cancel()

		if err != nil {
			msg_warn := fmt.Sprintf("Helth check failed for service '%s': %v", name, err)
			if !sm.handleServiceDown(ctx, name, conn, msg_warn) {
				return
			}
			continue
		}

		if resp.GetStatus() != healthpb.HealthCheckResponse_SERVING {
			msg_warn := fmt.Sprintf("Health check not SERVING for service '%s': %v", name, resp.GetStatus())
			if !sm.handleServiceDown(ctx, name, conn, msg_warn) {
				return
			}
			continue
		}

		// Service is healthy: set status and keep the connection.
		sm.logger.Info("Service '%s' is SERVING at address '%s'", name, addr)
		sm.setServiceStatus(name, conn, true)

		// while service is up, wait upInterval, then re-check.
		sleepInterval = sm.upInterval
		select {
		case <-time.After(sleepInterval):
			continue
		case <-ctx.Done():
			return
		}
	}
}

// handleServiceDown handles the scenario where a service is down or unresponsive during health checks.
//
// Parameters:
//   - ctx: context for cancellation and timeout control of the monitoring goroutine.
//   - name: the service name.
//   - conn: the active gRPC connection to the service, if any; will be closed on error.
//   - message: the warning message to log about the issue.
//
// Returns:
//   - bool: true if monitoring should continue after waiting, false if the context was cancelled and monitoring should stop.
func (sm *ServiceManager) handleServiceDown(ctx context.Context, name string, conn *grpc.ClientConn, message string) bool {
	sm.logger.Warn(message)
	if conn != nil {
		_ = conn.Close()
	}
	sm.setServiceStatus(name, nil, false)
	select {
	case <-time.After(sm.downInterval):
		return true // continue monitoring loop
	case <-ctx.Done():
		return false // exit monitoring loop
	}
}

// setServiceStatus safely updates the internal map and closes any previous connection
// when replacing it with a different one.
//
// Parameters:
//   - name: the service name
//   - conn: the active gRPC connection to the service, if any
//   - ready: the state of the service connection
func (sm *ServiceManager) setServiceStatus(name string, conn *grpc.ClientConn, ready bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	current, ok := sm.services[name]
	if !ok { // unknown service
		sm.logger.Error("Attempted to set status for unknown service '%s'", name)
		return
	}

	// If there is an existing connection different from the new one, close it.
	if current.Conn != nil && current.Conn != conn {
		_ = current.Conn.Close()
		// sm.logger.Info("Closed previous connection for service '%s'", name) // too much verbosity
	}

	current.Conn = conn
	current.Ready = ready

	// sm.logger.Info("Set status for service '%s': ready=%v", name, ready) // too much verbosity
}

// GetConn returns a ready *grpc.ClientConn for the given service name, or error if not ready.
//
// Parameters:
//   - name: the service name
//
// Returns:
//   - *grpc.ClientConn: the connection to the service, if ready
//   - error: if the connection is not ready
func (sm *ServiceManager) GetConn(name string) (*grpc.ClientConn, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	svc, ok := sm.services[name]
	if !ok {
		return nil, fmt.Errorf("Service '%s' not found", name)
	}
	if !svc.Ready || svc.Conn == nil {
		return nil, fmt.Errorf("Service '%s' not ready", name)
	}
	return svc.Conn, nil
}

// GetConnWithTimeout waits until the service is ready or the timeout expires.
//
// Parameters:
//   - name: the service name
//   - timeout: the maximum duration to wait for the service to become ready
//
// Returns:
//   - *grpc.ClientConn: the active connection to the service if ready within timeout
//   - error: if the service is not ready before timeout expires
func (sm *ServiceManager) GetConnWithTimeout(name string, timeout time.Duration) (*grpc.ClientConn, error) {
	deadline := time.Now().Add(timeout)
	for {
		conn, err := sm.GetConn(name)
		if err == nil {
			return conn, nil
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("Timeout waiting for service '%s': last error: %w", name, err)
		}
		time.Sleep(100 * time.Millisecond)
	}
}
