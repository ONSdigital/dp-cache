package cache

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ONSdigital/log.go/v2/log"
)

// Cacher defines the required methods to initialise a cache
type Cacher interface {
	Close()
	Get(key string) (interface{}, bool)
	Set(key string, data interface{})
	AddUpdateFunc(key string, updateFunc func() (interface{}, error))
	StartUpdates(ctx context.Context, channel chan error)
}

// Cache contains all the information to start, update and close caching data
type Cache struct {
	data        sync.Map
	config      Config
	close       chan struct{}
	UpdateFuncs map[string]func() (interface{}, error)
}

// Configs contains all the configurations for the cache
type Config struct {
	UpdateInterval *time.Duration
}

// NewCache create a cache object which will update at every updateInterval
// If updateInterval is nil, this means that the cache will only be updated once at the start of a service
func NewCache(ctx context.Context, config Config) (*Cache, error) {
	if config.UpdateInterval != nil {
		if *config.UpdateInterval <= 0 {
			err := errors.New("cache update interval duration is less than or equal to 0")
			log.Error(ctx, "invalid cache update interval given", err)
			return nil, err
		}
	}

	return &Cache{
		data:        sync.Map{},
		config:      config,
		close:       make(chan struct{}),
		UpdateFuncs: make(map[string]func() (interface{}, error)),
	}, nil
}

// Get retrieves the specific value for the specified key stored in `data` within the `Cache`
func (dc *Cache) Get(key string) (interface{}, bool) {
	return dc.data.Load(key)
}

// Set adds the specified value with the specified key in `data` within the `Cache`
func (dc *Cache) Set(key string, data interface{}) {
	dc.data.Store(key, data)
}

// Close closes the caching of data when called where the data will no longer be updated and the data itself is reset
func (dc *Cache) Close() {
	if dc.config.UpdateInterval != nil {
		dc.close <- struct{}{}
		for key := range dc.UpdateFuncs {
			dc.data.Store(key, "")
		}
		dc.UpdateFuncs = make(map[string]func() (interface{}, error))
	}
}

// AddUpdateFunc adds an update function to the cache for a specific data corresponding to the `key` passed to the function
// This update function will then be triggered once or at every fixed interval as per the prior setup of the TopicCache
func (dc *Cache) AddUpdateFunc(key string, updateFunc func() (interface{}, error)) {
	dc.UpdateFuncs[key] = updateFunc
}

// UpdateContent calls all the update functions with a key value stored in the Cache to update the relevant data with the same key values
func (dc *Cache) UpdateContent(_ context.Context) error {
	for key, updateFunc := range dc.UpdateFuncs {
		updatedContent, err := updateFunc()
		if err != nil {
			return fmt.Errorf("failed to update search cache for %s. error: %v", key, err)
		}
		dc.Set(key, updatedContent)
	}
	return nil
}

// StartUpdates informs the cache to start updating the cache data at every update interval which was configured when setting up the cache
func (dc *Cache) StartUpdates(ctx context.Context, errorChannel chan error) {
	if len(dc.UpdateFuncs) == 0 {
		return
	}

	if dc.config.UpdateInterval != nil {
		// Create a new goroutine to handle periodic updates with the specified interval
		go func() {
			ticker := time.NewTicker(*dc.config.UpdateInterval)
			defer ticker.Stop()

			// Wait for the first ticker and handle periodic updates
			for {
				select {
				case <-ticker.C:
					err := dc.UpdateContent(ctx)
					if err != nil {
						log.Error(ctx, err.Error(), err)
						errorChannel <- err
					}

				case <-dc.close:
					return
				case <-ctx.Done():
					return
				}
			}
		}()
	}
}

// StartAndManageUpdates performs an initial synchronous cache update once called
// and then hands over control to the cache to manage periodic updates asynchronously.
func (dc *Cache) StartAndManageUpdates(ctx context.Context, errorChannel chan error) {
	// Step 1: Perform the initial synchronous cache update
	err := dc.UpdateContent(ctx)
	if err != nil {
		errorChannel <- err
		dc.Close()
		return
	}

	// Step 2: Start periodic updates managed by the cache internally
	dc.StartUpdates(ctx, errorChannel)
}
