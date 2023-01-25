package private

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/ONSdigital/dp-cache/internal/navigation/cache"
	"github.com/ONSdigital/dp-topic-api/models"
	"github.com/ONSdigital/dp-topic-api/sdk"
	topicCliErr "github.com/ONSdigital/dp-topic-api/sdk/errors"
	mockTopic "github.com/ONSdigital/dp-topic-api/sdk/mocks"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	testCensusTitle       = "Census"
	testCensusTopicID     = "1234"
	testCensusSubTopicID1 = "5678"
	testCensusSubTopicID2 = "9012"
)

var (
	releaseDate = time.Date(2022, time.November, 23, 9, 30, 0, 0, time.Local)

	testCensusRootTopicPrivate = models.TopicResponse{
		ID:      testCensusTopicID,
		Next:    &testCensusRootTopic,
		Current: &testCensusRootTopic,
	}

	// census sub topic level (when GetSubTopics is called with `testCensusTopicID` - testRootCensusTopic)
	testCensusSubTopicsPrivate = &models.PrivateSubtopics{
		Count:        2,
		Offset:       0,
		Limit:        50,
		TotalCount:   2,
		PrivateItems: &[]models.TopicResponse{testCensusSubtopicResponse1, testCensusSubtopicResponse2},
	}

	testCensusSubtopicResponse1 = models.TopicResponse{
		ID:      testCensusSubTopicID1,
		Next:    &testCensusSubTopic1,
		Current: &testCensusSubTopic1,
	}

	testCensusSubtopicResponse2 = models.TopicResponse{
		ID:      testCensusSubTopicID2,
		Next:    &testCensusSubTopic2,
		Current: &testCensusSubTopic2,
	}
)

var (
	testCensusRootTopic = models.Topic{
		ID:          cache.CensusTopicID,
		Title:       testCensusTitle,
		SubtopicIds: []string{"5678", "9012"},
	}

	testCensusSubTopic1 = models.Topic{
		ID:          testCensusSubTopicID1,
		Title:       "Age",
		ReleaseDate: &releaseDate,
	}

	testCensusSubTopic2 = models.Topic{
		ID:          testCensusSubTopicID2,
		Title:       "Ethnicity",
		ReleaseDate: &releaseDate,
	}

	expectedCensusTopicCache = &cache.Topic{
		ID:              testCensusTopicID,
		LocaliseKeyName: testCensusTitle,
		Query:           fmt.Sprintf("%s,%s,%s", testCensusTopicID, testCensusSubTopicID1, testCensusSubTopicID2),
	}
)

func TestUpdateCensusTopic(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	mockedTopicClient := &mockTopic.ClienterMock{
		GetTopicPrivateFunc: func(ctx context.Context, reqHeaders sdk.Headers, id string) (*models.TopicResponse, topicCliErr.Error) {
			return &testCensusRootTopicPrivate, nil
		},

		GetSubtopicsPrivateFunc: func(ctx context.Context, reqHeaders sdk.Headers, id string) (*models.PrivateSubtopics, topicCliErr.Error) {
			switch id {
			case testCensusTopicID:
				return testCensusSubTopicsPrivate, nil
			default:
				return nil, topicCliErr.StatusError{
					Err: errors.New("unexpected error"),
				}
			}
		},
	}

	Convey("Given census root topic does exist and has subtopics", t, func() {
		Convey("When UpdateCensusTopic is called", func() {
			respCensusTopicCache := UpdateCensusTopic(ctx, testCensusTopicID, "", mockedTopicClient)()

			Convey("Then the census topic cache is returned", func() {
				So(respCensusTopicCache, ShouldNotBeNil)

				So(respCensusTopicCache.ID, ShouldEqual, expectedCensusTopicCache.ID)
				So(respCensusTopicCache.LocaliseKeyName, ShouldEqual, expectedCensusTopicCache.LocaliseKeyName)

				So(respCensusTopicCache.Query, ShouldContainSubstring, expectedCensusTopicCache.ID)
				So(respCensusTopicCache.Query, ShouldContainSubstring, testCensusSubTopicID1)
				So(respCensusTopicCache.Query, ShouldContainSubstring, testCensusSubTopicID2)
			})
		})
	})

	Convey("Given an error in getting census topic from topic-api", t, func() {
		failedCensusTopicClient := &mockTopic.ClienterMock{
			GetTopicPrivateFunc: func(ctx context.Context, reqHeaders sdk.Headers, id string) (*models.TopicResponse, topicCliErr.Error) {
				return nil, topicCliErr.StatusError{
					Err: errors.New("unexpected error"),
				}
			},
		}

		Convey("When UpdateCensusTopic is called", func() {
			respCensusTopicCache := UpdateCensusTopic(ctx, testCensusTopicID, "", failedCensusTopicClient)()

			Convey("Then an empty census topic cache should be returned", func() {
				So(respCensusTopicCache, ShouldResemble, cache.GetEmptyCensusTopic())
			})
		})
	})

	Convey("Given census topics private items is nil", t, func() {
		censusTopicsNilClient := &mockTopic.ClienterMock{
			GetTopicPrivateFunc: func(ctx context.Context, reqHeaders sdk.Headers, id string) (*models.TopicResponse, topicCliErr.Error) {
				return &models.TopicResponse{
					ID:   "1234",
					Next: &models.Topic{ID: "1234", Title: "Census", SubtopicIds: []string{}},
				}, nil
			},

			GetSubtopicsPrivateFunc: func(ctx context.Context, reqHeaders sdk.Headers, id string) (*models.PrivateSubtopics, topicCliErr.Error) {
				return nil, topicCliErr.StatusError{
					Err: errors.New("unexpected error"),
				}
			},
		}

		Convey("When UpdateCensusTopic is called", func() {
			respCensusTopicCache := UpdateCensusTopic(ctx, testCensusTopicID, "", censusTopicsNilClient)()

			Convey("Then an empty census topic cache should be returned", func() {
				So(respCensusTopicCache, ShouldResemble, cache.GetEmptyCensusTopic())
			})
		})
	})
}
