package worker

import (
	"bufio"
	"context"
	"io"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"
)

// IWorkerPool defines the interface for worker pools
type IWorkerPool interface {
	GetWorkerEndpoint() string
	Shutdown()
}

type Worker struct {
	port     int
	cmd      *exec.Cmd
	endpoint string // Used for both local and external workers
}

type WorkerPool struct {
	workers  []*Worker
	counter  uint64
	enabled  bool
	cwd      string
	cancelFn context.CancelFunc
	wg       sync.WaitGroup
	command  string
	args     []string
}

// Keep track of used ports and last used port to ensure uniqueness across restarts
var (
	usedPorts    = make(map[int]bool)
	lastUsedPort = 5173 // Start with base port - 1
	usedPortsMux sync.Mutex
)

// Default values for backward compatibility
var (
	defaultCommand = "node"
	defaultArgs    = []string{"node_modules/.bin/blastra", "start"}
)

func getNextAvailablePort(basePort int) int {
	usedPortsMux.Lock()
	defer usedPortsMux.Unlock()

	// Always start from the last used port + 1 to ensure different ports after restart
	port := lastUsedPort + 1
	for usedPorts[port] {
		port++
	}
	usedPorts[port] = true
	lastUsedPort = port
	return port
}

func releasePort(port int) {
	usedPortsMux.Lock()
	delete(usedPorts, port)
	usedPortsMux.Unlock()
}

func releaseAllPorts() {
	usedPortsMux.Lock()
	for port := range usedPorts {
		delete(usedPorts, port)
	}
	// Don't reset lastUsedPort to ensure new ports after restart
	usedPortsMux.Unlock()
}

// StartWorkerPool initializes a worker pool with default command and args
func StartWorkerPool(workerCount int, cwd string) (IWorkerPool, error) {
	return StartWorkerPoolWithConfig(workerCount, cwd, defaultCommand, defaultArgs, nil)
}

// StartWorkerPoolWithCommand initializes a worker pool with custom command and args (used for testing)
func StartWorkerPoolWithCommand(workerCount int, cwd string, command string, args []string) (IWorkerPool, error) {
	return StartWorkerPoolWithConfig(workerCount, cwd, command, args, nil)
}

// StartWorkerPoolWithConfig is the internal implementation that handles all configuration options
func StartWorkerPoolWithConfig(workerCount int, cwd string, command string, args []string, externalURLs []string) (IWorkerPool, error) {
	// If external URLs are provided, create a worker pool with those URLs
	if len(externalURLs) > 0 {
		log.Debug("Using external worker URLs")
		wp := &WorkerPool{
			enabled: true,
		}

		for _, url := range externalURLs {
			worker := &Worker{
				endpoint: url,
			}
			wp.workers = append(wp.workers, worker)
		}

		return wp, nil
	}

	// Otherwise, proceed with local worker creation
	if workerCount <= 0 {
		log.Debug("Worker pool disabled (workerCount <= 0)")
		return &WorkerPool{enabled: false}, nil
	}

	// Release all ports before starting new workers to ensure clean state
	releaseAllPorts()

	log.Debugf("Starting worker pool with %d workers", workerCount)
	ctx, cancel := context.WithCancel(context.Background())
	wp := &WorkerPool{
		enabled:  true,
		cwd:      cwd,
		cancelFn: cancel,
		command:  command,
		args:     args,
	}

	for i := 0; i < workerCount; i++ {
		port := getNextAvailablePort(5174)
		log.Debugf("Starting worker on port %d", port)
		cmd := exec.CommandContext(ctx, wp.command, wp.args...)
		cmd.Dir = cwd
		// Set PORT environment variable
		cmd.Env = append(os.Environ(), "PORT="+strconv.Itoa(port))

		if log.IsLevelEnabled(log.DebugLevel) {
			stdoutPipe, err := cmd.StdoutPipe()
			if err != nil {
				cancel()
				releaseAllPorts() // Release ports on error
				return nil, err
			}
			stderrPipe, err := cmd.StderrPipe()
			if err != nil {
				cancel()
				releaseAllPorts() // Release ports on error
				return nil, err
			}

			// Log worker's stdout
			go func(p int, r io.Reader) {
				scanner := bufio.NewScanner(r)
				for scanner.Scan() {
					log.Debugf("Worker %d stdout: %s", p, scanner.Text())
				}
			}(port, stdoutPipe)

			// Log worker's stderr
			go func(p int, r io.Reader) {
				scanner := bufio.NewScanner(r)
				for scanner.Scan() {
					log.Debugf("Worker %d stderr: %s", p, scanner.Text())
				}
			}(port, stderrPipe)
		}

		if err := cmd.Start(); err != nil {
			cancel()
			releaseAllPorts() // Release ports on error
			log.Errorf("Failed to start worker on port %d: %v", port, err)
			return nil, err
		}

		worker := &Worker{
			port:     port,
			cmd:      cmd,
			endpoint: "http://localhost:" + strconv.Itoa(port),
		}

		wp.workers = append(wp.workers, worker)
		wp.wg.Add(1)

		// Monitor worker process
		go func(w *Worker) {
			defer wp.wg.Done()
			if err := w.cmd.Wait(); err != nil {
				if ctx.Err() == nil { // Only log if not cancelled
					log.Errorf("Worker on port %d exited with error: %v", w.port, err)
				}
			}
			releasePort(w.port) // Release port when worker exits
		}(worker)

		// Wait a short time for the server to start
		time.Sleep(100 * time.Millisecond)
	}

	log.Debug("All workers started successfully")
	return wp, nil
}

func (wp *WorkerPool) Shutdown() {
	if !wp.enabled {
		return
	}

	// If we have no cancelFn, we're using external workers
	if wp.cancelFn == nil {
		log.Debug("External worker pool shutdown - no local workers to stop")
		return
	}

	log.Debug("Shutting down worker pool")
	wp.cancelFn() // Signal all workers to stop

	// Create a channel to signal timeout
	done := make(chan struct{})
	go func() {
		wp.wg.Wait()
		close(done)
	}()

	// Wait for workers to shut down with timeout
	select {
	case <-done:
		log.Debug("All workers shut down successfully")
	case <-time.After(2 * time.Second):
		log.Warn("Worker shutdown timed out, forcefully terminating")
		for _, worker := range wp.workers {
			if worker.cmd != nil && worker.cmd.Process != nil {
				// Send SIGTERM first
				worker.cmd.Process.Signal(os.Interrupt)
				// Give it a short time to clean up
				time.Sleep(100 * time.Millisecond)
				// If still running, force kill
				worker.cmd.Process.Kill()
			}
			if worker.port > 0 {
				releasePort(worker.port)
			}
		}
		// Wait for cleanup after kill
		wp.wg.Wait()
	}

	// Clear worker list
	wp.workers = nil
}

func (wp *WorkerPool) GetWorkerEndpoint() string {
	if !wp.enabled || len(wp.workers) == 0 {
		log.Debug("No available workers in pool")
		return ""
	}
	idx := atomic.AddUint64(&wp.counter, 1)
	worker := wp.workers[int(idx)%len(wp.workers)]
	log.Debugf("Dispatching request to worker endpoint: %s", worker.endpoint)
	return worker.endpoint
}
