package client

//go:generate moq -out mock_navigation_clienter.go -pkg client . Clienter

import (
	"context"
	"time"

	topicModel "github.com/ONSdigital/dp-topic-api/models"
	topicCli "github.com/ONSdigital/dp-topic-api/sdk"
)

// Clienter is an interface with methods required for navigation cache
type Clienter interface {
	AddNavigationCache(ctx context.Context, updateInterval time.Duration) error
	GetNavigationData(ctx context.Context, lang string) (*topicModel.Navigation, error)
	Close()
	StartBackgroundUpdate(ctx context.Context, errorChannel chan error)
}

// Clients contains all the required Clients for navigation cache
type Clients struct {
	Topic topicCli.Clienter
}
