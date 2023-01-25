package cache

import (
	"context"
	"testing"
	"time"

	topicModel "github.com/ONSdigital/dp-topic-api/models"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNewNavigationCache(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	Convey("Given a valid cache update interval which is greater than 0", t, func() {
		updateCacheInterval := 1 * time.Millisecond

		Convey("When NewNavigationCache is called", func() {
			testCache, err := NewNavigationCache(ctx, &updateCacheInterval)

			Convey("Then a navigation cache object should be successfully returned", func() {
				So(testCache, ShouldNotBeEmpty)

				Convey("And no error should be returned", func() {
					So(err, ShouldBeNil)
				})
			})
		})
	})

	Convey("Given no cache update interval (nil)", t, func() {
		Convey("When NewNavigationCache is called", func() {
			testCache, err := NewNavigationCache(ctx, nil)

			Convey("Then a cache object should be successfully returned", func() {
				So(testCache, ShouldNotBeEmpty)

				Convey("And no error should be returned", func() {
					So(err, ShouldBeNil)
				})
			})
		})
	})

	Convey("Given an invalid cache update interval which is less than or equal to 0", t, func() {
		updateCacheInterval := 0 * time.Second

		Convey("When NewNavigationCache is called", func() {
			testCache, err := NewNavigationCache(ctx, &updateCacheInterval)

			Convey("Then an error should be returned", func() {
				So(err, ShouldNotBeNil)

				Convey("And a nil cache object should be returned", func() {
					So(testCache, ShouldBeNil)
				})
			})
		})
	})
}

func TestNavigationAddUpdateFunc(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	Convey("Given an update function to update a topic", t, func() {
		mockNavigationCache, err := NewNavigationCache(ctx, nil)
		So(err, ShouldBeNil)

		navigationUpdateFunc := func() *topicModel.Navigation {
			return &topicModel.Navigation{
				Description: "Test",
			}
		}

		Convey("When AddUpdateFunc is called", func() {
			mockNavigationCache.AddUpdateFunc("test", navigationUpdateFunc)

			Convey("Then the update function is added to the cache", func() {
				So(mockNavigationCache.UpdateFuncs["test"], ShouldNotBeEmpty)
			})
		})
	})
}
