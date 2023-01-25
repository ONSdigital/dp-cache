package cache

import (
	"fmt"
	"strings"
	"sync"

	"github.com/ONSdigital/dp-topic-api/models"
)

// SubtopicsIDs contains a list of subtopics in map form with addition to mutex locking
// The subtopicsMap is used to keep a record of subtopics to be later used to generate the subtopics id `query` for a topic
// and to check if the subtopic id given by a user exists
type SubtopicsIDs struct {
	subtopicsMap sync.Map
}

// NewSubTopicsMap creates a new subtopics id map to store subtopic ids with addition to mutex locking
func NewSubTopicsMap() *SubtopicsIDs {
	return &SubtopicsIDs{}
}

// Get returns a bool value for the given key (id) to inform that the subtopic id exists
func (t *SubtopicsIDs) Get(key string) bool {
	_, ok := t.subtopicsMap.Load(key)
	return ok
}

// GetSubtopicItems returns a list of subtopics for given topic
func (t *SubtopicsIDs) GetSubtopicItems() map[string]*models.Topic {
	subtopics := make(map[string]*models.Topic)
	t.subtopicsMap.Range(func(key, value any) bool {
		subtopics[fmt.Sprint(key)] = value.(*models.Topic)
		return true
	})
	return subtopics
}

// GetSubtopicsIDsQuery gets the subtopics ID query for a topic
func (t *SubtopicsIDs) GetSubtopicsIDsQuery() string {
	lenSyncMap := func(m *sync.Map) int {
		var i int
		m.Range(func(k, v interface{}) bool {
			i++
			return true
		})
		return i
	}

	ids := make([]string, 0, lenSyncMap(&t.subtopicsMap))
	t.subtopicsMap.Range(func(key, value any) bool {
		ids = append(ids, fmt.Sprint(key))
		return true
	})

	return strings.Join(ids, ",")
}

// AppendSubtopicID appends the subtopic id to the map stored in SubtopicsIDs with consideration to mutex locking
func (t *SubtopicsIDs) AppendSubtopicItems(topic *models.Topic) {
	t.subtopicsMap.Store(topic.ID, topic)
}
