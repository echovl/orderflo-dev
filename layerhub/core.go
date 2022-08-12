package layerhub

import (
	"github.com/echovl/orderflo-dev/cloud/github"
	"github.com/echovl/orderflo-dev/cloud/google"
	"github.com/echovl/orderflo-dev/feeds"
	"github.com/echovl/orderflo-dev/payments"
	"github.com/echovl/orderflo-dev/upload"
	"go.uber.org/zap"
)

// Core is responsible for managing storage, parsing and manipulation of templates.
// It is the primary interface for API handlers.
type Core struct {
	db              DB
	jsonDB          JSONDB
	uploader        upload.SignedUploader
	pixabay         feeds.MediaFeed
	pexels          feeds.MediaFeed
	paymentProvider payments.Provider
	github          *github.Client
	google          *google.Client

	renderer Renderer

	Logger *zap.SugaredLogger
}

type CoreConfig struct {
	Logger          *zap.Logger
	DB              DB
	JSONDB          JSONDB
	Uploader        upload.SignedUploader
	Pixabay         feeds.MediaFeed
	Pexels          feeds.MediaFeed
	PaymentProvider payments.Provider
	Renderer        Renderer
	GithubClient    *github.Client
	GoogleClient    *google.Client
}

func New(cfg CoreConfig) *Core {
	return &Core{
		Logger:          cfg.Logger.Sugar(),
		db:              cfg.DB,
		jsonDB:          cfg.JSONDB,
		uploader:        cfg.Uploader,
		pixabay:         cfg.Pixabay,
		pexels:          cfg.Pexels,
		paymentProvider: cfg.PaymentProvider,
		github:          cfg.GithubClient,
		google:          cfg.GoogleClient,
		renderer:        cfg.Renderer,
	}
}
