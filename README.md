# dp-cache

dp-cache is a library which provides functionality for a service to cache data by creating a cache object which contains the data (stored in memory). The update interval is used to define how often the cache should be updated and the use of Go channels for the application to update the cache object.

```go
// Cache contains all the information to start, update and close caching data
type Cache struct {
 data        sync.Map
 config      Config
 close       chan struct{}
 UpdateFuncs map[string]func() (interface{}, error)
}
```

## Setup app to cache data

Please note that the data is stored in a `sync.Map` which is like a Go `map[interface{}]interface{}` but is safe for concurrent use by multiple goroutines without additional locking or coordination [[1]][golang-sync-map].

### 1. Initialising the cache

- Create a `cache` package in the app
- Within the package, create a file which is appropriately named to the data which is being cached

#### a. Add a wrapper NEW cache function

The library gives us the option to cache any form of data. We want to create a wrapper function which allows us to extend the cache object for our needs.

##### Example

```go
import (
    dpcache "github.com/ONSdigital/dp-cache"
)
```

```go
// TopicCache is a wrapper to dpcache.Cache which has additional fields and methods specifically for caching topics
type TopicCache struct {
    *dpcache.Cache
}
```

```go
// NewTopicCache create a topic cache object to be used in the service which will update at every updateInterval
// If updateInterval is nil, this means that the cache will only be updated once at the start of the service
func NewTopicCache(ctx context.Context, updateInterval *time.Duration) (*TopicCache, error) {
    config := dpcache.Config{
        UpdateInterval: updateInterval,
    }

    cache, err := dpcache.NewCache(ctx, config)
    if err != nil {
        logData := log.Data{
            "update_interval": updateInterval,
        }
        log.Error(ctx, "failed to create cache from dpcache", err, logData)
        return nil, err
    }

    topicCache := &TopicCache{cache}

    return topicCache, nil
}
```

#### b. Add a GET function to retrieve data from cache

The existing `Get` function in the library is able to return any type of data (`interface{}`). We want to create our own GET function to ensure that the data returned is the data type which we want to cache.

##### Example

```go
func (dc *TopicCache) GetData(ctx context.Context, key string) (*Topic, error) {
    topicCacheInterface, ok := dc.Get(key)
    if !ok {
        err := fmt.Errorf("cached topic data with key %s not found", key)
        log.Error(ctx, "failed to get cached topic data", err)
        return getEmptyTopic(), err
    }

    topicCacheData, ok := topicCacheInterface.(*Topic)
    if !ok {
        err := errors.New("topicCacheInterface is not type *Topic")
        log.Error(ctx, "failed type assertion on topicCacheInterface", err)
        return getEmptyTopic(), err
    }

    if topicCacheData == nil {
        err := errors.New("topicCacheData is nil")
        log.Error(ctx, "cached topic data is nil", err)
        return getEmptyTopic(), err
    }

    return topicCacheData, nil
}
```

Please note that, in the example above, an empty object is returned in the event of an error. A nil value can be returned as well. Implement the function according to what is best for the app

#### c. Add a wrapper AddUpdateFunc

The existing `AddUpdateFunc` function in the library is able to handle any type of data (`interface{}`). We want to create our own `AddUpdateFunc` function to ensure that the updated data is the data type which we want to cache.

##### Example

```go
// AddUpdateFunc adds an update function to the topic cache for a topic with the `key` passed to the function
// This update function will then be triggered once or at every fixed interval as per the prior setup of the TopicCache
func (dc *TopicCache) AddUpdateFunc(key string, updateFunc func() *Topic) {
    dc.UpdateFuncs[key] = func() (interface{}, error) {
        // error handling is done within the updateFunc
        return updateFunc(), nil
    }
}
```

#### d. Add an update function to update the cache data

We need to implement an update function which will be called at every `updateInterval` to update the cache data. In most cases, the logic of this function is to retrieve the data again from the client at that present moment.

##### Example

```go
func UpdateTopic(ctx context.Context, topicClient topicCli.Clienter) func() *cache.Topic {
    return func() *cache.Topic {
        // add logic to get topic from dp-topic-api
        // then return topic
    }
}
```

### 2. Use cache in the service

- After completing [Step 1](#1-initialising-the-cache), you have all the necessary struct and functions to start the caching of data in the service

#### a. Create a cache object using func in [Step 1a](#a-add-a-wrapper-new-cache-function)

- Create the cache object where you initialise the service (usually `service.go` file)

##### Example

```go
svc.Cache.CensusTopic, err = cache.NewTopicCache(ctx, &cfg.CacheCensusTopicUpdateInterval)
    if err != nil {
        log.Error(ctx, "failed to create topics cache", err)
        return err
    }
```

- Please note that it is ideal for the update interval to be configurable for each environment (i.e. `cfg.CacheCensusTopicUpdateInterval`)
- If you pass `nil` for the `updateInterval`, this means that the service will only cache the data **once** at the start of the service

#### b. Add update function to cache

- Add your update function implemented in [Step 1d](#d-add-an-update-function-to-update-the-cache-data) so that the cache knows how to update the cache data

##### Example

```go
    svc.Cache.CensusTopic.AddUpdateFunc(cache.CensusTopicID, cachePublic.UpdateCensusTopic(ctx, clients.Topic))
```

- In the example above, `cache.CensusTopicID` is the `key` of the data stored in `sync.Map`. This key is required to retrieve the specific cache data from the `sync.Map` by using the GET function implemented in [Step 1b](#b-add-a-get-function-to-retrieve-data-from-cache).
- Also, in this example, `cachePublic.UpdateCensusTopic` is the update function which was implemented in [Step 1d](#d-add-an-update-function-to-update-the-cache-data).

#### c. Start updates for cache

- When you run the service, you want to start the cache so that it will update at every `updateInterval`

##### Example

```go
// Start caching
    go svc.Cache.CensusTopic.StartUpdates(ctx, svcErrors)
```

- This is usually done in `service.Run` function in the DP apps

#### d. Handle closing of cache

- Whenever an error occurs and the service tries to gracefully shutdown, it is important to close the cache gracefully

##### Example

```go
// stop caching
    svc.Cache.CensusTopic.Close()
```

## Examples

Examples of caching implementation in services can be found below:

- <https://github.com/ONSdigital/dp-frontend-search-controller/blob/develop/service/service.go#L78>
- <https://github.com/ONSdigital/dp-frontend-homepage-controller/blob/develop/service/service.go#L68>

[//]: # (Reference Links and Images)
   [golang-sync-map]: <https://pkg.go.dev/sync#Map>
