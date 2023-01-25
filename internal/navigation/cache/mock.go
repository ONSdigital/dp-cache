package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/ONSdigital/dp-topic-api/models"
)

// GetMockCacheList returns a mocked list of cache which contains the census topic cache and navigation cache
func GetMockCacheList(ctx context.Context, lang string) (*List, error) {
	testCensusTopicCache, err := getMockCensusTopicCache(ctx)
	if err != nil {
		return nil, err
	}

	cacheList := List{
		CensusTopic: testCensusTopicCache,
	}

	return &cacheList, nil
}

// getMockCensusTopicCache returns a mocked Cenus topic which contains all the information for the mock census topic
func getMockCensusTopicCache(ctx context.Context) (*TopicCache, error) {
	testCensusTopicCache, err := NewTopicCache(ctx, nil)
	if err != nil {
		return nil, err
	}

	testCensusTopicCache.Set(CensusTopicID, GetMockCensusTopic())

	return testCensusTopicCache, nil
}

// GetMockCensusTopic returns a mocked Cenus topic which contains all the information for the mock census topic
func GetMockCensusTopic() *Topic {
	t := time.Date(2022, time.November, 23, 9, 30, 0, 0, time.Local)

	mockCensusTopic := &Topic{
		ID:              CensusTopicID,
		LocaliseKeyName: "Census",
		Query:           fmt.Sprintf("1234,5678,%s", CensusTopicID),
	}

	subtopic1 := &models.Topic{
		ID:          "1234",
		ReleaseDate: &t,
		Title:       "Age",
	}

	subtopic2 := &models.Topic{
		ID:          "5678",
		ReleaseDate: &t,
		Title:       "Ethnicity",
	}

	mockCensusTopic.List = NewSubTopicsMap()
	mockCensusTopic.List.AppendSubtopicItems(subtopic1)
	mockCensusTopic.List.AppendSubtopicItems(subtopic2)

	return mockCensusTopic
}
