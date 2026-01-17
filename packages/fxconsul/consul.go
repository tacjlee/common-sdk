package fxconsul

import (
	"log"
	"os"
	"sync"
	"time"

	"github.com/hashicorp/consul/api"
)

// ConfigChangeCallback is called when configuration changes are detected
type ConfigChangeCallback func(changedKeys []string)

type cacheEntry struct {
	value     string
	expiresAt time.Time
}

// ConsulClient provides configuration management with Consul KV store
type ConsulClient struct {
	client    *api.Client
	cache     map[string]cacheEntry
	cacheMu   sync.RWMutex
	cacheTTL  time.Duration
	available bool
	basePath  string

	// Watch-related fields
	callbacks  []ConfigChangeCallback
	callbackMu sync.RWMutex
	stopChan   chan struct{}
	watching   bool
	watchMu    sync.Mutex
	lastIndex  uint64
}

var (
	consulInstance *ConsulClient
	consulOnce     sync.Once
)

// GetConsulClient returns a singleton ConsulClient instance
func GetConsulClient() *ConsulClient {
	consulOnce.Do(func() {
		consulInstance = newConsulClient()
	})
	return consulInstance
}

func newConsulClient() *ConsulClient {
	enabled := os.Getenv("CONSUL_ENABLED")
	if enabled == "false" {
		log.Println("Consul is disabled via CONSUL_ENABLED=false")
		return &ConsulClient{
			available: false,
			cache:     make(map[string]cacheEntry),
			cacheTTL:  60 * time.Second,
			basePath:  "config/dev/settings",
			callbacks: make([]ConfigChangeCallback, 0),
		}
	}

	host := os.Getenv("CONSUL_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("CONSUL_PORT")
	if port == "" {
		port = "8500"
	}

	ttlSeconds := os.Getenv("CONSUL_CACHE_TTL")
	cacheTTL := 60 * time.Second
	if ttlSeconds != "" {
		if parsed, err := time.ParseDuration(ttlSeconds + "s"); err == nil {
			cacheTTL = parsed
		}
	}

	basePath := os.Getenv("CONSUL_BASE_PATH")
	if basePath == "" {
		basePath = "config/dev/settings"
	}

	config := api.DefaultConfig()
	config.Address = host + ":" + port

	client, err := api.NewClient(config)
	if err != nil {
		log.Printf("Warning: Failed to create Consul client: %v", err)
		return &ConsulClient{
			available: false,
			cache:     make(map[string]cacheEntry),
			cacheTTL:  cacheTTL,
			basePath:  basePath,
			callbacks: make([]ConfigChangeCallback, 0),
		}
	}

	// Test connection
	_, err = client.Agent().Self()
	if err != nil {
		log.Printf("Warning: Consul is not reachable at %s:%s - falling back to environment variables: %v", host, port, err)
		return &ConsulClient{
			client:    client,
			available: false,
			cache:     make(map[string]cacheEntry),
			cacheTTL:  cacheTTL,
			basePath:  basePath,
			callbacks: make([]ConfigChangeCallback, 0),
		}
	}

	log.Printf("Consul connected successfully at %s:%s", host, port)
	return &ConsulClient{
		client:    client,
		available: true,
		cache:     make(map[string]cacheEntry),
		cacheTTL:  cacheTTL,
		basePath:  basePath,
		callbacks: make([]ConfigChangeCallback, 0),
	}
}

// IsAvailable returns whether Consul is reachable
func (c *ConsulClient) IsAvailable() bool {
	return c.available
}

// OnConfigChange registers a callback to be invoked when configuration changes
func (c *ConsulClient) OnConfigChange(callback ConfigChangeCallback) {
	c.callbackMu.Lock()
	defer c.callbackMu.Unlock()
	c.callbacks = append(c.callbacks, callback)
}

// WatchConfig starts watching for configuration changes in Consul
// This runs in a background goroutine and calls registered callbacks when changes are detected
func (c *ConsulClient) WatchConfig() {
	c.watchMu.Lock()
	if c.watching {
		c.watchMu.Unlock()
		log.Println("Config watch already running")
		return
	}
	c.watching = true
	c.stopChan = make(chan struct{})
	c.watchMu.Unlock()

	go c.watchLoop()
	log.Println("Started watching Consul configuration changes")
}

// StopWatch stops the configuration watch goroutine
func (c *ConsulClient) StopWatch() {
	c.watchMu.Lock()
	defer c.watchMu.Unlock()

	if !c.watching {
		return
	}

	close(c.stopChan)
	c.watching = false
	log.Println("Stopped watching Consul configuration changes")
}

func (c *ConsulClient) watchLoop() {
	retryInterval := 5 * time.Second

	for {
		select {
		case <-c.stopChan:
			return
		default:
		}

		if !c.available || c.client == nil {
			// Try to reconnect
			c.tryReconnect()
			select {
			case <-c.stopChan:
				return
			case <-time.After(retryInterval):
				continue
			}
		}

		// Use blocking query to watch for changes
		opts := &api.QueryOptions{
			WaitIndex: c.lastIndex,
			WaitTime:  30 * time.Second, // Long poll timeout
		}

		pairs, meta, err := c.client.KV().List(c.basePath, opts)
		if err != nil {
			log.Printf("Warning: Error watching Consul KV: %v", err)
			c.available = false
			select {
			case <-c.stopChan:
				return
			case <-time.After(retryInterval):
				continue
			}
		}

		// Check if index changed (meaning data changed)
		if meta.LastIndex != c.lastIndex {
			if c.lastIndex != 0 { // Skip first iteration
				changedKeys := make([]string, 0)
				for _, pair := range pairs {
					// Extract key name from full path
					if len(pair.Key) > len(c.basePath)+1 {
						key := pair.Key[len(c.basePath)+1:] // Remove basePath/ prefix
						changedKeys = append(changedKeys, key)
					}
				}

				log.Printf("Consul configuration changed, refreshing cache")
				c.RefreshCache()
				c.notifyCallbacks(changedKeys)
			}
			c.lastIndex = meta.LastIndex
		}
	}
}

func (c *ConsulClient) tryReconnect() {
	if c.client == nil {
		return
	}

	_, err := c.client.Agent().Self()
	if err == nil {
		c.available = true
		c.lastIndex = 0 // Reset index on reconnect
		log.Println("Consul connection restored")
	}
}

func (c *ConsulClient) notifyCallbacks(changedKeys []string) {
	c.callbackMu.RLock()
	callbacks := make([]ConfigChangeCallback, len(c.callbacks))
	copy(callbacks, c.callbacks)
	c.callbackMu.RUnlock()

	for _, callback := range callbacks {
		go callback(changedKeys)
	}
}

// GetSetting retrieves a configuration value with fallback:
// 1. Check cache (if not expired)
// 2. Try Consul KV
// 3. Fall back to environment variable
// 4. Return default value
func (c *ConsulClient) GetSetting(key string, defaultValue string) string {
	// Check cache first
	c.cacheMu.RLock()
	if entry, ok := c.cache[key]; ok && time.Now().Before(entry.expiresAt) {
		c.cacheMu.RUnlock()
		return entry.value
	}
	c.cacheMu.RUnlock()

	// Try Consul if available
	if c.available && c.client != nil {
		consulKey := c.basePath + "/" + key
		pair, _, err := c.client.KV().Get(consulKey, nil)
		if err != nil {
			log.Printf("Warning: Failed to get key %s from Consul: %v", consulKey, err)
			// Mark as unavailable for subsequent calls
			c.available = false
		} else if pair != nil && len(pair.Value) > 0 {
			value := string(pair.Value)
			// Update cache
			c.cacheMu.Lock()
			c.cache[key] = cacheEntry{
				value:     value,
				expiresAt: time.Now().Add(c.cacheTTL),
			}
			c.cacheMu.Unlock()
			return value
		}
	}

	// Fall back to environment variable
	if envValue := os.Getenv(key); envValue != "" {
		return envValue
	}

	return defaultValue
}

// GetSettingInt retrieves an integer configuration value
func (c *ConsulClient) GetSettingInt(key string, defaultValue int) int {
	strValue := c.GetSetting(key, "")
	if strValue == "" {
		return defaultValue
	}

	var intValue int
	_, err := parseIntValue(strValue, &intValue)
	if err != nil {
		return defaultValue
	}
	return intValue
}

// GetSettingBool retrieves a boolean configuration value
func (c *ConsulClient) GetSettingBool(key string, defaultValue bool) bool {
	strValue := c.GetSetting(key, "")
	if strValue == "" {
		return defaultValue
	}

	switch strValue {
	case "true", "True", "TRUE", "1", "yes", "Yes", "YES":
		return true
	case "false", "False", "FALSE", "0", "no", "No", "NO":
		return false
	default:
		return defaultValue
	}
}

func parseIntValue(s string, v *int) (int, error) {
	var result int
	negative := false
	for i, c := range s {
		if c == '-' && i == 0 {
			negative = true
			continue
		}
		if c < '0' || c > '9' {
			return 0, &parseError{s}
		}
		result = result*10 + int(c-'0')
	}
	if negative {
		result = -result
	}
	*v = result
	return result, nil
}

type parseError struct {
	value string
}

func (e *parseError) Error() string {
	return "cannot parse: " + e.value
}

// RefreshCache clears the cache to force fresh reads from Consul
func (c *ConsulClient) RefreshCache() {
	c.cacheMu.Lock()
	c.cache = make(map[string]cacheEntry)
	c.cacheMu.Unlock()

	// Retry connection if previously unavailable
	if !c.available && c.client != nil {
		_, err := c.client.Agent().Self()
		if err == nil {
			c.available = true
			log.Println("Consul connection restored")
		}
	}
}
