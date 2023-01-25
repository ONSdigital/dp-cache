package private

import (
	"context"
	"errors"
	"fmt"

	"github.com/ONSdigital/dp-cache/internal/navigation/cache"
	"github.com/ONSdigital/dp-topic-api/models"
	topicCli "github.com/ONSdigital/dp-topic-api/sdk"
	"github.com/ONSdigital/log.go/v2/log"
)

// UpdateCensusTopic is a function to update the census topic cache in publishing (private) mode.
// This function talks to the dp-topic-api via its private endpoints to retrieve the census topic and its subtopic ids
// The data returned by the dp-topic-api is of type *models.PrivateSubtopics which is then transformed in this function for the controller
// If an error has occurred, this is captured in log.Error and then an empty census topic is returned
func UpdateCensusTopic(ctx context.Context, censusTopicID, serviceAuthToken string, topicClient topicCli.Clienter) func() *cache.Topic {
	return func() *cache.Topic {
		// TODO can put two requests in Go routines

		// get census topic from dp-topic-api
		censusTopic, err := topicClient.GetTopicPrivate(ctx, topicCli.Headers{ServiceAuthToken: serviceAuthToken}, censusTopicID)
		if err != nil {
			logData := log.Data{
				"req_headers": topicCli.Headers{},
			}
			log.Error(ctx, "failed to get root topics from topic-api", err, logData)
			return cache.GetEmptyCensusTopic()
		}

		// get census subtopics from dp-topic-api
		censusSubtopics, err := topicClient.GetSubtopicsPrivate(ctx, topicCli.Headers{ServiceAuthToken: serviceAuthToken}, censusTopicID)
		if err != nil {
			logData := log.Data{
				"req_headers": topicCli.Headers{},
			}
			log.Error(ctx, "failed to get census subtopics from topic-api", err, logData)
			return cache.GetEmptyCensusTopic()
		}

		if censusSubtopics.PrivateItems == nil {
			err := errors.New("root topic public items is nil")
			log.Error(ctx, "failed to deference root topics items pointer", err)
			return cache.GetEmptyCensusTopic()
		}

		censusTopicCache := setTopicCachePublic(ctx, *censusTopic, *censusSubtopics.PrivateItems)

		if censusTopicCache == nil {
			err := errors.New("census root topic not found")
			log.Error(ctx, "failed to get census topic to cache", err)
			return cache.GetEmptyCensusTopic()
		}

		return censusTopicCache
	}
}

func setTopicCachePublic(ctx context.Context, censusTopic models.TopicResponse, subtopics []models.TopicResponse) *cache.Topic {
	censusTopicCache := &cache.Topic{
		ID:              censusTopic.ID,
		LocaliseKeyName: censusTopic.Next.Title,
	}

	subtopicCache := cache.NewSubTopicsMap()
	for _, subtopic := range subtopics {
		subtopicCache.AppendSubtopicItems(subtopic.Next)
	}

	censusTopicCache.List = subtopicCache
	censusTopicCache.Query = subtopicCache.GetSubtopicsIDsQuery()
	censusTopicCache.Query = fmt.Sprintf("%s,%s", censusTopic.ID, censusTopicCache.Query)

	return censusTopicCache
}
