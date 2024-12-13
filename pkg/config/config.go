package config

import (
	"errors"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/devthefuture-org/blastra/pkg/cache"
	log "github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

const (
	DefaultHTTPPort        = 8080
	DefaultHTTPSPort       = 8443
	DefaultCacheTTL        = 5 * time.Minute
	DefaultCacheSize       = 1000
	DefaultRateLimit       = 100 // requests per second
	DefaultBurst           = 200
	DefaultStaticDir       = "./dist/client"
	DefaultSSRScript       = "node node_modules/@blastra/core/output.js"
	DefaultMaxAgeStatic    = 86400
	DefaultMaxAgeSSR       = 60
	DefaultShutdownTimeout = 15 * time.Second
	DefaultBlastraCWD      = "."
	DefaultListStatic      = true
	DefaultSSRCacheEnabled = true
	DefaultTrustProxy      = false
	DefaultWorkerCommand   = "node"
	DefaultWorkerArgs      = "node_modules/.bin/blastra start"
)

type Configuration struct {
	// HTTP Server settings
	HTTPPort    int
	HTTPSPort   int
	EnableHTTPS bool
	TLSCertPath string
	TLSKeyPath  string
	GzipEnabled bool
	RateLimit   rate.Limit
	Burst       int
	TrustProxy  bool // Whether to trust proxy headers for client IP

	// Directory and script settings
	BlastraCWD        string            // Working directory for Blastra
	StaticDir         string            // Directory for static files
	SSRScript         []string          // SSR rendering script
	ListStaticContent bool              // Whether to list static content
	ExcludePatterns   []string          // Patterns to exclude from preloading
	CacheControl      map[string]string // Custom cache control headers

	// Cache settings
	SSRCacheEnabled   bool // Whether to enable SSR in-memory caching
	CacheTTL          time.Duration
	CacheSize         int
	NotFoundCacheTTL  time.Duration // Optional, defaults to CacheTTL/2 if not set
	NotFoundCacheSize int           // Optional, defaults to CacheSize/4 if not set
	ExternalCacheType cache.ExternalCacheType

	// Redis cache settings
	RedisURL      string
	RedisPassword string
	RedisDB       int

	// Filesystem cache settings
	CacheDir string

	// Cache durations
	MaxAgeStatic int
	MaxAgeSSR    int

	// Server settings
	ShutdownTimeout time.Duration
	CPUCount        int
	WorkerCount     int

	// Static file settings
	PreloadStaticFileList *bool // Whether to preload static file list for routing (default: true)
	PreloadStaticContent  *bool // Whether to preload static files into memory (default: true)
	StaticMaxAge          int   // Cache duration for static files in seconds

	// Worker settings
	WorkerCommand string   // Command to run worker process
	WorkerArgs    []string // Arguments for worker command
	WorkerURLs    []string // External worker URLs (if set, no local workers will be created)
}

// Helper function to get PreloadStaticFileList with default value
func (c *Configuration) ShouldPreloadFileList() bool {
	if c.PreloadStaticFileList == nil {
		return true // default enabled
	}
	return *c.PreloadStaticFileList
}

// Helper function to get PreloadStaticContent with default value
func (c *Configuration) ShouldPreloadContent() bool {
	if c.PreloadStaticContent == nil {
		return true // default enabled
	}
	return *c.PreloadStaticContent
}

// GetNotFoundCacheConfig returns the configuration for NotFoundCache,
// deriving defaults from main cache settings if not explicitly configured
func (c *Configuration) GetNotFoundCacheConfig() (time.Duration, int) {
	if !c.SSRCacheEnabled {
		return 0, 0
	}

	ttl := c.NotFoundCacheTTL
	if ttl == 0 {
		ttl = c.CacheTTL / 2
	}

	size := c.NotFoundCacheSize
	if size == 0 {
		size = c.CacheSize / 4
		if size < 250 { // Ensure minimum size
			size = 250
		}
	}

	return ttl, size
}

// GetSSRCacheConfig returns the configuration for SSR cache,
// returning zero values if caching is disabled
func (c *Configuration) GetSSRCacheConfig() (time.Duration, int) {
	if !c.SSRCacheEnabled {
		return 0, 0
	}
	return c.CacheTTL, c.CacheSize
}

// GetExternalCacheConfig returns the configuration for external cache
func (c *Configuration) GetExternalCacheConfig() cache.ExternalCacheConfig {
	return cache.ExternalCacheConfig{
		CacheConfig: cache.CacheConfig{
			TTL:     c.CacheTTL,
			MaxSize: c.CacheSize,
		},
		Type:          c.ExternalCacheType,
		RedisURL:      c.RedisURL,
		RedisPassword: c.RedisPassword,
		RedisDB:       c.RedisDB,
		CacheDir:      c.CacheDir,
	}
}

func LoadConfiguration() (*Configuration, error) {
	config := &Configuration{}

	getEnvInt := func(key string, defaultVal int) (int, error) {
		valStr := os.Getenv("BLASTRA_" + key)
		if valStr == "" {
			log.Debugf("Environment variable BLASTRA_%s not set, using default: %d", key, defaultVal)
			return defaultVal, nil
		}
		val, err := strconv.Atoi(valStr)
		if err != nil {
			log.Errorf("Failed to parse BLASTRA_%s: %v", key, err)
			return 0, err
		}
		log.Debugf("Loaded BLASTRA_%s: %d", key, val)
		return val, nil
	}

	getEnvBool := func(key string, defaultVal bool) bool {
		valStr := os.Getenv("BLASTRA_" + key)
		if valStr == "" {
			log.Debugf("Environment variable BLASTRA_%s not set, using default: %v", key, defaultVal)
			return defaultVal
		}
		val, err := strconv.ParseBool(valStr)
		if err != nil {
			log.Debugf("Failed to parse bool BLASTRA_%s: %v, using default: %v", key, err, defaultVal)
			return defaultVal
		}
		log.Debugf("Loaded BLASTRA_%s: %v", key, val)
		return val
	}

	getEnvDuration := func(key string, defaultVal time.Duration) (time.Duration, error) {
		valStr := os.Getenv("BLASTRA_" + key)
		if valStr == "" {
			log.Debugf("Environment variable BLASTRA_%s not set, using default: %v", key, defaultVal)
			return defaultVal, nil
		}
		val, err := time.ParseDuration(valStr)
		if err != nil {
			log.Errorf("Failed to parse BLASTRA_%s: %v", key, err)
			return 0, err
		}
		log.Debugf("Loaded BLASTRA_%s: %v", key, val)
		return val, nil
	}

	var err error

	// Load all configuration values
	config.HTTPPort, err = getEnvInt("HTTP_PORT", DefaultHTTPPort)
	if err != nil {
		return nil, errors.New("invalid BLASTRA_HTTP_PORT")
	}

	config.HTTPSPort, err = getEnvInt("HTTPS_PORT", DefaultHTTPSPort)
	if err != nil {
		return nil, errors.New("invalid BLASTRA_HTTPS_PORT")
	}

	config.EnableHTTPS = getEnvBool("ENABLE_HTTPS", false)
	config.TLSCertPath = os.Getenv("BLASTRA_TLS_CERT_PATH")
	config.TLSKeyPath = os.Getenv("BLASTRA_TLS_KEY_PATH")

	// Load proxy trust setting
	config.TrustProxy = getEnvBool("TRUST_PROXY", DefaultTrustProxy)

	// Load Blastra working directory
	config.BlastraCWD = os.Getenv("BLASTRA_CWD")
	if config.BlastraCWD == "" {
		config.BlastraCWD = DefaultBlastraCWD
		log.Debugf("No BLASTRA_CWD set, using default %s", DefaultBlastraCWD)
	}

	// Load static content settings
	config.StaticDir = os.Getenv("BLASTRA_STATIC_DIR")
	if config.StaticDir == "" {
		config.StaticDir = DefaultStaticDir
		log.Debugf("No BLASTRA_STATIC_DIR set, using default %s", DefaultStaticDir)
	}

	config.ListStaticContent = getEnvBool("LIST_STATIC_CONTENT", DefaultListStatic)

	// Load SSR script
	config.SSRScript = strings.Fields(os.Getenv("BLASTRA_SSR_SCRIPT"))
	if len(config.SSRScript) == 0 {
		config.SSRScript = strings.Fields(DefaultSSRScript)
		log.Debugf("No BLASTRA_SSR_SCRIPT set, using default %s", DefaultSSRScript)
	}

	// Load worker settings
	config.WorkerCommand = os.Getenv("BLASTRA_WORKER_COMMAND")
	if config.WorkerCommand == "" {
		config.WorkerCommand = DefaultWorkerCommand
		log.Debugf("No BLASTRA_WORKER_COMMAND set, using default %s", DefaultWorkerCommand)
	}

	workerArgs := os.Getenv("BLASTRA_WORKER_ARGS")
	if workerArgs == "" {
		config.WorkerArgs = strings.Fields(DefaultWorkerArgs)
		log.Debugf("No BLASTRA_WORKER_ARGS set, using default %s", DefaultWorkerArgs)
	} else {
		config.WorkerArgs = strings.Fields(workerArgs)
	}

	workerURLs := os.Getenv("BLASTRA_WORKER_URLS")
	if workerURLs != "" {
		config.WorkerURLs = strings.Split(workerURLs, ",")
		for i, url := range config.WorkerURLs {
			config.WorkerURLs[i] = strings.TrimSpace(url)
		}
		log.Debugf("Using external worker URLs: %v", config.WorkerURLs)
	}

	// Load cache settings
	config.SSRCacheEnabled = getEnvBool("SSR_CACHE_ENABLED", DefaultSSRCacheEnabled)

	config.CacheTTL, err = getEnvDuration("CACHE_TTL", DefaultCacheTTL)
	if err != nil {
		return nil, errors.New("invalid BLASTRA_CACHE_TTL")
	}

	config.CacheSize, err = getEnvInt("CACHE_SIZE", DefaultCacheSize)
	if err != nil {
		return nil, errors.New("invalid BLASTRA_CACHE_SIZE")
	}

	// Load external cache configuration
	externalCacheType := os.Getenv("BLASTRA_EXTERNAL_CACHE_TYPE")
	if externalCacheType == "" {
		config.ExternalCacheType = cache.ExternalCacheNone
	} else {
		config.ExternalCacheType = cache.ExternalCacheType(externalCacheType)
	}

	config.RedisURL = os.Getenv("BLASTRA_REDIS_URL")
	config.RedisPassword = os.Getenv("BLASTRA_REDIS_PASSWORD")
	config.RedisDB, _ = getEnvInt("REDIS_DB", 0)
	config.CacheDir = os.Getenv("BLASTRA_CACHE_DIR")

	// Load NotFoundCache settings
	config.NotFoundCacheTTL, err = getEnvDuration("NOTFOUND_CACHE_TTL", 0)
	if err != nil {
		return nil, errors.New("invalid BLASTRA_NOTFOUND_CACHE_TTL")
	}

	config.NotFoundCacheSize, err = getEnvInt("NOTFOUND_CACHE_SIZE", 0)
	if err != nil {
		return nil, errors.New("invalid BLASTRA_NOTFOUND_CACHE_SIZE")
	}

	// Load rate limiting settings
	rateLimitStr := os.Getenv("BLASTRA_RATE_LIMIT")
	if rateLimitStr == "" {
		config.RateLimit = rate.Limit(DefaultRateLimit)
	} else {
		rateLimit, err := strconv.Atoi(rateLimitStr)
		if err != nil {
			return nil, errors.New("invalid BLASTRA_RATE_LIMIT")
		}
		config.RateLimit = rate.Limit(rateLimit)
	}

	config.Burst, err = getEnvInt("BURST", DefaultBurst)
	if err != nil {
		return nil, errors.New("invalid BLASTRA_BURST")
	}

	// Load cache durations
	config.MaxAgeStatic, err = getEnvInt("MAX_AGE_STATIC", DefaultMaxAgeStatic)
	if err != nil {
		return nil, errors.New("invalid BLASTRA_MAX_AGE_STATIC")
	}

	config.MaxAgeSSR, err = getEnvInt("MAX_AGE_SSR", DefaultMaxAgeSSR)
	if err != nil {
		return nil, errors.New("invalid BLASTRA_MAX_AGE_SSR")
	}

	// Load shutdown timeout
	shutdownTimeout, err := getEnvDuration("SHUTDOWN_TIMEOUT", DefaultShutdownTimeout)
	if err != nil {
		return nil, errors.New("invalid BLASTRA_SHUTDOWN_TIMEOUT")
	}
	config.ShutdownTimeout = shutdownTimeout

	// Load compression settings
	config.GzipEnabled = getEnvBool("GZIP_ENABLED", false)

	// Load CPU and worker settings
	cpuLimitStr := os.Getenv("BLASTRA_CPU_LIMIT")
	if cpuLimitStr == "" {
		config.CPUCount = runtime.NumCPU()
	} else {
		cpuLimit, err := strconv.Atoi(cpuLimitStr)
		if err != nil || cpuLimit < 1 {
			config.CPUCount = runtime.NumCPU()
		} else {
			config.CPUCount = cpuLimit
		}
	}

	workerCountStr := os.Getenv("BLASTRA_SSR_WORKERS")
	if workerCountStr == "" {
		config.WorkerCount = config.CPUCount
	} else {
		wCount, err := strconv.Atoi(workerCountStr)
		if err != nil || wCount < 0 {
			config.WorkerCount = config.CPUCount
		} else {
			config.WorkerCount = wCount
		}
	}

	// Load preload settings
	if preloadList := os.Getenv("BLASTRA_PRELOAD_STATIC_FILE_LIST"); preloadList != "" {
		val := getEnvBool("PRELOAD_STATIC_FILE_LIST", true)
		config.PreloadStaticFileList = &val
	}

	if preloadContent := os.Getenv("BLASTRA_PRELOAD_STATIC_CONTENT"); preloadContent != "" {
		val := getEnvBool("PRELOAD_STATIC_CONTENT", true)
		config.PreloadStaticContent = &val
	}

	// Validate HTTPS settings
	if config.EnableHTTPS && (config.TLSCertPath == "" || config.TLSKeyPath == "") {
		return nil, errors.New("BLASTRA_TLS_CERT_PATH and BLASTRA_TLS_KEY_PATH must be set when BLASTRA_ENABLE_HTTPS is true")
	}

	return config, nil
}
