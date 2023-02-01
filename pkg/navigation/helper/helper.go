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
	Clienter  client.Clienter
	Config    Config
	CacheList cache.List
	Clients   *client.Clients
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

func Init(ctx context.Context, cfg Config) (svc *Helper, err error) {

	svc = &Helper{}
	svc.CacheList = cache.List{}
	svc.Clients = &client.Clients{
		Topic: topicCli.New(cfg.APIRouterURL),
	}
	svc.Config = cfg

	if svc.Config.IsPublishingMode {
		svc.Clienter = client.NewPublishingClient(ctx, svc.Clients, cfg.Languages)
	} else {
		svc.Clienter, err = client.NewWebClient(ctx, svc.Clients, cfg.CacheUpdateInterval, cfg.Languages)
		if err != nil {
			log.Fatal(ctx, "failed to create homepage web client", err)
			return
		}
	}
	if err = svc.Clienter.AddNavigationCache(ctx, cfg.CacheUpdateInterval); err != nil {
		log.Fatal(ctx, "failed to add navigation cache to homepage client", err)
		return
	}

	if svc.Config.EnableCensusTopicSubsection {
		// Initialise caching census topics
		cache.CensusTopicID = svc.Config.CensusTopicID
		svc.CacheList.CensusTopic, err = cache.NewTopicCache(ctx, &svc.Config.CacheUpdateInterval)
		if err != nil {
			log.Error(ctx, "failed to create topics cache", err)
			return
		}

		if svc.Config.IsPublishingMode {
			if err = svc.CacheList.CensusTopic.AddUpdateFunc(ctx, cache.CensusTopicID, cachePrivate.UpdateCensusTopic(ctx, svc.Config.CensusTopicID, svc.Config.ServiceAuthToken, svc.Clients.Topic)); err != nil {
				log.Error(ctx, "failed to create topics cache", err)
				return
			}
		} else {
			if err = svc.CacheList.CensusTopic.AddUpdateFunc(ctx, cache.CensusTopicID, cachePublic.UpdateCensusTopic(ctx, svc.Config.CensusTopicID, svc.Clients.Topic)); err != nil {
				log.Error(ctx, "failed to create topics cache", err)
				return
			}
		}
	}
	return
}

func (svc *Helper) RunUpdates(ctx context.Context, svcErrors chan error) {
	// Start background polling of topics API for navbar data (changes)
	go svc.Clienter.StartBackgroundUpdate(ctx, svcErrors)

	if svc.Config.EnableCensusTopicSubsection {
		go svc.CacheList.CensusTopic.StartUpdates(ctx, svcErrors)
	}
}

func (svc *Helper) GetMappedNavigationContent(ctx context.Context, lang string) (content []model.NavigationItem, err error) {
	navigationContent, err := svc.Clienter.GetNavigationData(ctx, lang)
	if err != nil {
		return
	}
	content = mapper.MapNavigationContent(*navigationContent)
	return
}
