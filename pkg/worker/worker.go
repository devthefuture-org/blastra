package worker

import (
	"bufio"
	"context"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
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

// simple ring buffer to keep last N lines (stderr tail)
type ringBuffer struct {
	lines []string
	next  int
	full  bool
	mu    sync.Mutex
}

func newRingBuffer(n int) *ringBuffer {
	if n <= 0 {
		n = 1
	}
	return &ringBuffer{lines: make([]string, n)}
}

func (rb *ringBuffer) add(s string) {
	rb.mu.Lock()
	rb.lines[rb.next] = s
	rb.next = (rb.next + 1) % len(rb.lines)
	if !rb.full && rb.next == 0 {
		rb.full = true
	}
	rb.mu.Unlock()
}

func (rb *ringBuffer) snapshot() []string {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	if !rb.full {
		out := make([]string, rb.next)
		copy(out, rb.lines[:rb.next])
		return out
	}
	out := make([]string, len(rb.lines))
	copy(out, rb.lines[rb.next:])
	copy(out[len(rb.lines)-rb.next:], rb.lines[:rb.next])
	return out
}

func getEnvBool(key string, def bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	switch strings.ToLower(val) {
	case "1", "t", "true", "y", "yes", "on":
		return true
	case "0", "f", "false", "n", "no", "off":
		return false
	default:
		return def
	}
}

func getEnvInt(key string, def int) int {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	if n, err := strconv.Atoi(val); err == nil {
		return n
	}
	return def
}

func getEnvDuration(key string, def time.Duration) time.Duration {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	if d, err := time.ParseDuration(val); err == nil {
		return d
	}
	return def
}

func exitDetails(err error) (code int, signal string) {
	if ee, ok := err.(*exec.ExitError); ok {
		if ws, ok := ee.Sys().(syscall.WaitStatus); ok {
			code = ws.ExitStatus()
			if ws.Signaled() {
				signal = ws.Signal().String()
			}
			return
		}
	}
	return
}

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

	// Worker diagnostics configuration (env-driven to avoid API changes)
	streamStdio := getEnvBool("BLASTRA_WORKER_STDIO_STREAM", false)
	stderrTailLines := getEnvInt("BLASTRA_WORKER_STDERR_TAIL_LINES", 200)
	readyPattern := os.Getenv("BLASTRA_WORKER_READY_PATTERN")
	if readyPattern == "" {
		readyPattern = "BLASTRA_READY"
	}
	readyTimeout := getEnvDuration("BLASTRA_WORKER_READY_TIMEOUT", 10*time.Second)
	nodeOptionsExtra := os.Getenv("BLASTRA_WORKER_NODE_OPTIONS")
	debugEnv := os.Getenv("BLASTRA_WORKER_DEBUG")
	forceColor := getEnvBool("BLASTRA_WORKER_FORCE_COLOR", true)

	for i := 0; i < workerCount; i++ {
		port := getNextAvailablePort(5174)
		log.Debugf("Starting worker on port %d", port)

		cmd := exec.CommandContext(ctx, wp.command, wp.args...)
		cmd.Dir = cwd

		// Build environment
		env := append(os.Environ(), "PORT="+strconv.Itoa(port))
		if forceColor {
			env = append(env, "FORCE_COLOR=1")
		}
		baseNodeOpts := "--enable-source-maps --trace-uncaught"
		combinedNodeOpts := strings.TrimSpace(strings.Join([]string{os.Getenv("NODE_OPTIONS"), baseNodeOpts, nodeOptionsExtra}, " "))
		if combinedNodeOpts != "" {
			env = append(env, "NODE_OPTIONS="+combinedNodeOpts)
		}
		if debugEnv != "" {
			env = append(env, "DEBUG="+debugEnv)
		}
		cmd.Env = env

		// Always capture stdio to avoid deadlocks and keep diagnostics
		stdoutPipe, err := cmd.StdoutPipe()
		if err != nil {
			cancel()
			releaseAllPorts()
			return nil, err
		}
		stderrPipe, err := cmd.StderrPipe()
		if err != nil {
			cancel()
			releaseAllPorts()
			return nil, err
		}

		startedAt := time.Now()
		if err := cmd.Start(); err != nil {
			cancel()
			releaseAllPorts()
			log.Errorf("Failed to start worker on port %d: %v", port, err)
			return nil, err
		}
		pid := cmd.Process.Pid

		log.WithFields(log.Fields{
			"port": port,
			"pid":  pid,
			"cmd":  wp.command,
			"args": strings.Join(wp.args, " "),
			"cwd":  cwd,
		}).Info("Worker started")

		// stderr tail buffer
		stderrTail := newRingBuffer(stderrTailLines)

		// Ready detection
		readyCh := make(chan struct{}, 1)
		var readyOnce sync.Once

		// Stream stdout
		go func(p, procPid int, r io.Reader) {
			scanner := bufio.NewScanner(r)
			for scanner.Scan() {
				line := scanner.Text()
				// readiness token
				if readyPattern != "" && strings.Contains(line, readyPattern) {
					readyOnce.Do(func() { readyCh <- struct{}{} })
				}
				// optional streaming
				if streamStdio || log.IsLevelEnabled(log.DebugLevel) {
					log.WithFields(log.Fields{
						"port":   p,
						"pid":    procPid,
						"stream": "stdout",
					}).Debug(line)
				}
			}
		}(port, pid, stdoutPipe)

		// Stream stderr (always keep tail)
		go func(p, procPid int, r io.Reader) {
			scanner := bufio.NewScanner(r)
			for scanner.Scan() {
				line := scanner.Text()
				stderrTail.add(line)
				if streamStdio || log.IsLevelEnabled(log.DebugLevel) {
					log.WithFields(log.Fields{
						"port":   p,
						"pid":    procPid,
						"stream": "stderr",
					}).Debug(line)
				}
			}
		}(port, pid, stderrPipe)

		// Optionally wait for readiness (non-fatal on timeout to keep compatibility)
		if readyPattern != "" && readyTimeout > 0 {
			select {
			case <-readyCh:
				dur := time.Since(startedAt)
				log.WithFields(log.Fields{
					"port":        port,
					"pid":         pid,
					"ready_ms":    dur.Milliseconds(),
					"ready_token": readyPattern,
				}).Info("Worker is ready")
			case <-time.After(readyTimeout):
				log.WithFields(log.Fields{
					"port":        port,
					"pid":         pid,
					"timeout":     readyTimeout.String(),
					"ready_token": readyPattern,
					"stderr_tail": strings.Join(stderrTail.snapshot(), "\n"),
				}).Warn("Worker did not report readiness before timeout")
			}
		}

		worker := &Worker{
			port:     port,
			cmd:      cmd,
			endpoint: "http://localhost:" + strconv.Itoa(port),
		}

		wp.workers = append(wp.workers, worker)
		wp.wg.Add(1)

		// Monitor worker process
		go func(w *Worker, started time.Time, tail *ringBuffer, procPid int) {
			defer wp.wg.Done()
			err := w.cmd.Wait()
			dur := time.Since(started)

			if ctx.Err() == nil { // Only log if not cancelled
				if err != nil {
					code, sig := exitDetails(err)
					log.WithFields(log.Fields{
						"port":        w.port,
						"pid":         procPid,
						"exit_error":  err.Error(),
						"exit_code":   code,
						"signal":      sig,
						"duration_ms": dur.Milliseconds(),
						"stderr_tail": strings.Join(tail.snapshot(), "\n"),
					}).Error("Worker exited with error")
				} else {
					log.WithFields(log.Fields{
						"port":        w.port,
						"pid":         procPid,
						"duration_ms": dur.Milliseconds(),
					}).Info("Worker exited")
				}
			}
			releasePort(w.port) // Release port when worker exits
		}(worker, startedAt, stderrTail, pid)

		// Short sleep to stagger starts
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
