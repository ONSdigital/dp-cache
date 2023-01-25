package public

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
	mockTopicCli "github.com/ONSdigital/dp-topic-api/sdk/mocks"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	testCensusTitle       = "Census"
	testCensusTopicID     = "1234"
	testCensusSubTopicID1 = "5678"
	testCensusSubTopicID2 = "1235"
)

var (
	releaseDate = time.Date(2022, time.November, 23, 9, 30, 0, 0, time.Local)

	testCensusRootTopic = models.Topic{
		ID:          testCensusTopicID,
		Title:       testCensusTitle,
		SubtopicIds: []string{testCensusSubTopicID1, testCensusSubTopicID2},
	}

	testCensusSubTopics = &models.PublicSubtopics{
		Count:       2,
		Offset:      0,
		Limit:       50,
		TotalCount:  2,
		PublicItems: &[]models.Topic{testCensusSubTopic1, testCensusSubTopic2},
	}

	testCensusSubTopic1 = models.Topic{
		ID:          testCensusSubTopicID1,
		ReleaseDate: &releaseDate,
		Title:       "Census Sub 1",
	}

	testCensusSubTopic2 = models.Topic{
		ID:          testCensusSubTopicID2,
		ReleaseDate: &releaseDate,
		Title:       "Census Sub 2",
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

	mockedTopicClient := &mockTopicCli.ClienterMock{
		GetTopicPublicFunc: func(ctx context.Context, reqHeaders sdk.Headers, id string) (*models.Topic, topicCliErr.Error) {
			return &testCensusRootTopic, nil
		},

		GetSubtopicsPublicFunc: func(ctx context.Context, reqHeaders sdk.Headers, id string) (*models.PublicSubtopics, topicCliErr.Error) {
			switch id {
			case testCensusTopicID:
				return testCensusSubTopics, nil
			default:
				return nil, topicCliErr.StatusError{
					Err: errors.New("unexpected error"),
				}
			}
		},
	}

	Convey("Given census root topic does exist and has subtopics", t, func() {
		Convey("When UpdateCensusTopic is called", func() {
			respCensusTopicCache := UpdateCensusTopic(ctx, testCensusTopicID, mockedTopicClient)()

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
		failedCensusTopicClient := &mockTopicCli.ClienterMock{
			GetTopicPublicFunc: func(ctx context.Context, reqHeaders sdk.Headers, id string) (*models.Topic, topicCliErr.Error) {
				return nil, topicCliErr.StatusError{
					Err: errors.New("unexpected error"),
				}
			},
		}

		Convey("When UpdateCensusTopic is called", func() {
			respCensusTopicCache := UpdateCensusTopic(ctx, testCensusTopicID, failedCensusTopicClient)()

			Convey("Then an empty census topic cache should be returned", func() {
				So(respCensusTopicCache, ShouldResemble, cache.GetEmptyCensusTopic())
			})
		})
	})

	Convey("Given census topics public items is nil", t, func() {
		censusTopicNilClient := &mockTopicCli.ClienterMock{
			GetTopicPublicFunc: func(ctx context.Context, reqHeaders sdk.Headers, id string) (*models.Topic, topicCliErr.Error) {
				return &models.Topic{ID: "1234", Title: "Census", SubtopicIds: []string{}}, nil
			},
			GetSubtopicsPublicFunc: func(ctx context.Context, reqHeaders sdk.Headers, id string) (*models.PublicSubtopics, topicCliErr.Error) {
				return nil, topicCliErr.StatusError{
					Err: errors.New("unexpected error"),
				}
			},
		}

		Convey("When UpdateCensusTopic is called", func() {
			respCensusTopicCache := UpdateCensusTopic(ctx, testCensusTopicID, censusTopicNilClient)()

			Convey("Then an empty census topic cache should be returned", func() {
				So(respCensusTopicCache, ShouldResemble, cache.GetEmptyCensusTopic())
			})
		})
	})
}
