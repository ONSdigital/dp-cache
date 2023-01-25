package helper

import (
	"context"
	"time"

	"github.com/ONSdigital/dp-cache/internal/navigation/cache"
	cachePrivate "github.com/ONSdigital/dp-cache/internal/navigation/cache/private"
	cachePublic "github.com/ONSdigital/dp-cache/internal/navigation/cache/public"
	mapper "github.com/ONSdigital/dp-cache/internal/navigation/mapper"
	client "github.com/ONSdigital/dp-cache/pkg/navigation/client"
	model "github.com/ONSdigital/dp-renderer/model"
	topicCli "github.com/ONSdigital/dp-topic-api/sdk"
	"github.com/ONSdigital/log.go/v2/log"
)

type Helper struct {
	Clienter client.Clienter
}

type Config struct {
	APIRouterURL                string
	CacheUpdateInterval         time.Duration
	EnableNewNavBar             bool
	EnableCensusTopicSubsection bool
	CensusTopicID               string
	IsPublishingMode            bool
	Languages                   []string
	ServiceAuthToken            string
}

func Init(ctx context.Context, cfg Config, svcErrors chan error) (helper *Helper, err error) {

	cacheList := cache.List{}
	clients := &client.Clients{
		Topic: topicCli.New(cfg.APIRouterURL),
	}
	helper = &Helper{}

	if cfg.IsPublishingMode {
		helper.Clienter = client.NewPublishingClient(ctx, clients, cfg.Languages)
	} else {
		helper.Clienter, err = client.NewWebClient(ctx, clients, cfg.CacheUpdateInterval, cfg.Languages)
		if err != nil {
			log.Fatal(ctx, "failed to create homepage web client", err)
			return
		}
	}
	if err = helper.Clienter.AddNavigationCache(ctx, cfg.CacheUpdateInterval); err != nil {
		log.Fatal(ctx, "failed to add navigation cache to homepage client", err)
		return
	}
	// Start background polling of topics API for navbar data (changes)
	go helper.Clienter.StartBackgroundUpdate(ctx, svcErrors)

	if cfg.EnableCensusTopicSubsection {
		// Initialise caching census topics
		cache.CensusTopicID = cfg.CensusTopicID
		cacheList.CensusTopic, err = cache.NewTopicCache(ctx, &cfg.CacheUpdateInterval)
		if err != nil {
			log.Error(ctx, "failed to create topics cache", err)
			return
		}

		if cfg.IsPublishingMode {
			if err = cacheList.CensusTopic.AddUpdateFunc(ctx, cache.CensusTopicID, cachePrivate.UpdateCensusTopic(ctx, cfg.CensusTopicID, cfg.ServiceAuthToken, clients.Topic)); err != nil {
				log.Error(ctx, "failed to create topics cache", err)
				return
			}
		} else {
			if err = cacheList.CensusTopic.AddUpdateFunc(ctx, cache.CensusTopicID, cachePublic.UpdateCensusTopic(ctx, cfg.CensusTopicID, clients.Topic)); err != nil {
				log.Error(ctx, "failed to create topics cache", err)
				return
			}
		}

		go cacheList.CensusTopic.StartUpdates(ctx, svcErrors)
	}
	return
}

func (svc *Helper) GetMappedNavigationContent(ctx context.Context, lang string) (content []model.NavigationItem, err error) {
	navigationContent, err := svc.Clienter.GetNavigationData(ctx, lang)
	if err != nil {
		return
	}
	content = mapper.MapNavigationContent(*navigationContent)
	return
}
