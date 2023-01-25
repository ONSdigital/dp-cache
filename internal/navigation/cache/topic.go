package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	dpcache "github.com/ONSdigital/dp-cache"
	"github.com/ONSdigital/log.go/v2/log"
)

var (
	// CensusTopicID is the id of the Census topic stored in mongodb which is accessible by using dp-topic-api
	CensusTopicID string

	// List of possible errors
	ErrEmptyTopicCache     = errors.New("topicCacheData is nil")
	ErrMissingCacheID      = errors.New("missing id value, set to empty string")
	ErrTopicCacheWrongType = errors.New("topicCacheInterface is not type *Topic")
)

// TopicCache is a wrapper to dpcache.Cache which has additional fields and methods specifically for caching topics
type TopicCache struct {
	*dpcache.Cache
}

// Topic represents the data which is cached for a topic to be used by the dp-frontend-search-controller
type Topic struct {
	ID              string
	LocaliseKeyName string
	// Query is a comma separated string of topic id and its subtopic ids which will be used by the controller to create the query
	Query string
	// List is a map[string]bool which contains the topic id and its subtopic ids which will be used to check if the topic id given by the user exists
	List *SubtopicsIDs
}

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

// ErrTopicCacheKeyMissing to generate error message
func ErrTopicCacheKeyMissing(key string) error {
	return fmt.Errorf("cached topic data with key %s not found", key)
}

func (dc *TopicCache) GetData(ctx context.Context, key string) (*Topic, error) {
	topicCacheInterface, ok := dc.Get(key)
	if !ok {
		log.Error(ctx, "failed to get cached topic data", ErrTopicCacheKeyMissing(key))
		return getEmptyTopic(), ErrTopicCacheKeyMissing(key)
	}

	topicCacheData, ok := topicCacheInterface.(*Topic)
	if !ok {
		log.Error(ctx, "failed type assertion on topicCacheInterface", ErrTopicCacheWrongType)
		return getEmptyTopic(), ErrTopicCacheWrongType
	}

	if topicCacheData == nil {
		log.Error(ctx, "cached topic data is nil", ErrEmptyTopicCache)
		return getEmptyTopic(), ErrEmptyTopicCache
	}

	return topicCacheData, nil
}

// AddUpdateFunc adds an update function to the topic cache for a topic with the `title` passed to the function
// This update function will then be triggered once or at every fixed interval as per the prior setup of the TopicCache
func (dc *TopicCache) AddUpdateFunc(ctx context.Context, id string, updateFunc func() *Topic) error {
	if id == "" {
		log.Error(ctx, "failed to add update function to topic cache", ErrMissingCacheID)
		return ErrMissingCacheID
	}

	if dc.UpdateFuncs[id] != nil {
		log.Warn(ctx, "an update function already exists for id, continuing to overwrite update function", log.Data{"id": id})
	}

	dc.UpdateFuncs[id] = func() (interface{}, error) {
		// error handling is done within the updateFunc
		return updateFunc(), nil
	}

	return nil
}

func (dc *TopicCache) GetCensusData(ctx context.Context) (*Topic, error) {
	censusTopicCache, err := dc.GetData(ctx, CensusTopicID)
	if err != nil {
		logData := log.Data{
			"key": CensusTopicID,
		}
		log.Error(ctx, "failed to get cached census topic data", err, logData)
		return GetEmptyCensusTopic(), err
	}

	return censusTopicCache, nil
}

// GetEmptyCensusTopic returns an empty census topic cache in the event when updating the cache of the census topic fails
func GetEmptyCensusTopic() *Topic {
	return &Topic{
		ID:   CensusTopicID,
		List: NewSubTopicsMap(),
	}
}

// GetEmptyTopic returns an empty topic cache in the event when updating the cache of the topic fails
func getEmptyTopic() *Topic {
	return &Topic{
		List: NewSubTopicsMap(),
	}
}
