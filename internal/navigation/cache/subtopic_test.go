package cache

import (
	"testing"
	"time"

	"github.com/ONSdigital/dp-topic-api/models"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	t = time.Date(2022, time.November, 23, 9, 30, 0, 0, time.Local)

	subtopic1 = &models.Topic{
		ID:          "1234",
		ReleaseDate: &t,
		Title:       "Age",
	}

	subtopic2 = &models.Topic{
		ID:          "5678",
		ReleaseDate: &t,
		Title:       "Ethnicity",
	}

	subtopic3 = &models.Topic{
		ID:          "9012",
		ReleaseDate: &t,
		Title:       "Demography",
	}
)

func TestAppendSubtopicID(t *testing.T) {
	t.Parallel()

	Convey("Given an empty SubtopicsIDs object", t, func() {
		subtopicIDsStore := NewSubTopicsMap()

		Convey("When AppendSubtopicItems is called", func() {
			subtopicIDsStore.AppendSubtopicItems(subtopic1)

			Convey("Then the new subtopic id should be added to the map", func() {
				So(subtopicIDsStore.Get("1234"), ShouldBeTrue)
			})
		})
	})

	Convey("Given a nil map in the SubtopicsIDs object", t, func() {
		subtopicIDsStore := NewSubTopicsMap()

		Convey("When AppendSubtopicItems is called", func() {
			subtopicIDsStore.AppendSubtopicItems(subtopic1)

			Convey("Then the new subtopic id should be added to the map", func() {
				So(subtopicIDsStore.Get("1234"), ShouldBeTrue)
			})
		})
	})

	Convey("Given an existing SubtopicsIDs object with data", t, func() {
		subtopicIDsStore := NewSubTopicsMap()
		subtopicIDsStore.subtopicsMap.Store("1234", subtopic1)

		Convey("When AppendSubtopicID is called", func() {
			subtopicIDsStore.AppendSubtopicItems(subtopic2)

			Convey("Then the new subtopic id should be added to the map", func() {
				So(subtopicIDsStore.Get("5678"), ShouldBeTrue)
			})

			Convey("And previous subtopic id should still be in the map", func() {
				So(subtopicIDsStore.Get("1234"), ShouldBeTrue)
			})
		})
	})

	Convey("Given AppendSubtopicID is called synchronously", t, func() {
		subtopicIDsStore := NewSubTopicsMap()

		Convey("When AppendSubtopicID is called", func() {
			go subtopicIDsStore.AppendSubtopicItems(subtopic2)
			go subtopicIDsStore.AppendSubtopicItems(subtopic3)

			Convey("Then the new subtopic ids should be added", func() {
				// Wait for the goroutines to finish
				time.Sleep(300 * time.Millisecond)

				So(subtopicIDsStore.Get("5678"), ShouldBeTrue)
				So(subtopicIDsStore.Get("9012"), ShouldBeTrue)
			})
		})
	})
}

func TestGetSubtopicsIDsQuery(t *testing.T) {
	t.Parallel()

	Convey("Given an empty list of subtopics", t, func() {
		subtopicIDsStore := NewSubTopicsMap()

		Convey("When GetSubtopicsIDsQuery is called", func() {
			subtopicsIDQuery := subtopicIDsStore.GetSubtopicsIDsQuery()

			Convey("Then subtopic ids query should be empty", func() {
				So(subtopicsIDQuery, ShouldBeEmpty)
			})
		})
	})

	Convey("Given a list of subtopics", t, func() {
		subtopicIDsStore := SubtopicsIDs{}
		subtopicIDsStore.subtopicsMap.Store("1234", subtopic1)
		subtopicIDsStore.subtopicsMap.Store("5678", subtopic2)

		Convey("When GetSubtopicsIDsQuery is called", func() {
			subtopicsIDQuery := subtopicIDsStore.GetSubtopicsIDsQuery()

			Convey("Then subtopic ids query should be returned successfully", func() {
				So(subtopicsIDQuery, ShouldContainSubstring, "1234")
				So(subtopicsIDQuery, ShouldContainSubstring, "5678")
			})
		})
	})
}
