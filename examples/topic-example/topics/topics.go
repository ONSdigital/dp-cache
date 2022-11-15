package topics

import (
	"context"
	"errors"
	"fmt"
	"time"

	dpcache "github.com/ONSdigital/dp-cache"
)

var MOCK_API_CLIENT_ID = 1

type SubTopicItem struct {
	ID   int
	Name string
}

type TopicItem struct {
	ID        int
	Name      string
	SubTopics []SubTopicItem
}

// ---------------------------------------
// Step #1
// ---------------------------------------

type TopicCache struct {
	*dpcache.Cache
}

// NewTopicCache create a topic cache object to be used in the service which will update at every updateInterval
// If updateInterval is nil, this means that the cache will only be updated once at the start of the service
func NewTopicCache(ctx context.Context, updateInterval *time.Duration) (*TopicCache, error) {
	config := dpcache.Config{
		UpdateInterval: updateInterval,
	}

	c, err := dpcache.NewCache(ctx, config)
	if err != nil {
		return nil, err
	}

	return &TopicCache{c}, nil
}

func (dc *TopicCache) GetData(ctx context.Context, key string) (*TopicItem, error) {
	topicCacheInterface, ok := dc.Get(key)
	if !ok {
		err := fmt.Errorf("cached topic data with key %s not found", key)
		return &TopicItem{}, err
	}

	topicCacheData, ok := topicCacheInterface.(*TopicItem)
	if !ok {
		err := errors.New("topicCacheInterface is not type *Topic")
		return &TopicItem{}, err
	}

	if topicCacheData == nil {
		err := errors.New("topicCacheData is nil")
		return &TopicItem{}, err
	}

	return topicCacheData, nil
}

// AddUpdateFunc adds an update function to the topic cache for a topic with the `key` passed to the function
// This update function will then be triggered once or at every fixed interval as per the prior setup of the TopicCache
func (dc *TopicCache) AddUpdateFunc(key string, updateFunc func() *TopicItem) {
	dc.UpdateFuncs[key] = func() (interface{}, error) {
		// error handling is done within the updateFunc
		return updateFunc(), nil
	}
}

var count = 0

func UpdateTopic() func() *TopicItem {
	return func() *TopicItem {
		// add logic to get topic data
		var subTopicItem SubTopicItem
		if count == 0 {
			subTopicItem = SubTopicItem{
				ID:   8341,
				Name: "age",
			}
			count = 1
		} else if count == 1 {
			subTopicItem = SubTopicItem{
				ID:   2223,
				Name: "Migration",
			}
			count = 2
		} else if count == 2 {
			subTopicItem = SubTopicItem{
				ID:   7845,
				Name: "Sex",
			}
			count = 0
		}

		mockClient := func() *TopicItem {
			MOCK_API_CLIENT_ID++
			return &TopicItem{
				ID:   MOCK_API_CLIENT_ID,
				Name: "Census",
				SubTopics: []SubTopicItem{
					subTopicItem,
				},
			}
		}

		// then return topic
		return mockClient()
	}
}
