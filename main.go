package main

import (
	"context"
	"log"
	"os"
	"reflect"
	"time"

	"github.com/layerhub-io/api/cloud/github"
	"github.com/layerhub-io/api/cloud/google"
	"github.com/layerhub-io/api/db/mongodb"
	"github.com/layerhub-io/api/db/mysql"
	"github.com/layerhub-io/api/db/redis"
	"github.com/layerhub-io/api/feeds/pexels"
	"github.com/layerhub-io/api/feeds/pixabay"
	"github.com/layerhub-io/api/http"
	"github.com/layerhub-io/api/layerhub"
	"github.com/layerhub-io/api/payments/paypal"
	"github.com/layerhub-io/api/upload/s3"
	"github.com/segmentio/analytics-go"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Config struct {
	Port               string `mapstructure:"PORT"`
	RedisAddr          string `mapstructure:"REDIS_ADDR"`
	RedisUsername      string `mapstructure:"REDIS_USERNAME"`
	RedisPassword      string `mapstructure:"REDIS_PASSWORD"`
	MongoURL           string `mapstructure:"MONGO_URL"`
	MongoDBName        string `mapstructure:"MONGO_DB_NAME"`
	MySQLDSN           string `mapstructure:"MYSQL_DSN"`
	AWSAccessKeyID     string `mapstructure:"AWS_ACCESS_KEY_ID"`
	AWSSecretAccessKey string `mapstructure:"AWS_SECRET_ACCESS_KEY"`
	AWSRegion          string `mapstructure:"AWS_REGION"`
	AWSBucket          string `mapstructure:"AWS_BUCKET"`
	PixabayKey         string `mapstructure:"PIXABAY_API_KEY"`
	PexelsKey          string `mapstructure:"PEXELS_API_KEY"`
	PaypalClientID     string `mapstructure:"PAYPAL_CLIENT_ID"`
	PaypalSecret       string `mapstructure:"PAYPAL_SECRET"`
	CDNBase            string `mapstructure:"CDN_BASE"`
	RendererSocket     string `mapstructure:"RENDERER_SOCKET"`
	GithubClientID     string `mapstructure:"GITHUB_CLIENT_ID"`
	GithubClientSecret string `mapstructure:"GITHUB_CLIENT_SECRET"`
	GithubRedirectURI  string `mapstructure:"GITHUB_REDIRECT_URI"`
	GoogleClientID     string `mapstructure:"GOOGLE_CLIENT_ID"`
	GoogleClientSecret string `mapstructure:"GOOGLE_CLIENT_SECRET"`
	GoogleRedirectURI  string `mapstructure:"GOOGLE_REDIRECT_URI"`
	SegmentWriteKey    string `mapstructure:"SEGMENT_WRITE_KEY"`
}

func loadConfig(path string) (Config, error) {
	var config Config

	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")
	viper.AutomaticEnv()
	viper.ReadInConfig()

	if err := unmarshalConfig(&config); err != nil {
		return config, err
	}

	// Load credentials for aws-sdk
	os.Setenv("AWS_ACCESS_KEY_ID", config.AWSAccessKeyID)
	os.Setenv("AWS_SECRET_ACCESS_KEY", config.AWSSecretAccessKey)

	return config, nil
}

func unmarshalConfig(cfg *Config) error {
	r := reflect.TypeOf(cfg).Elem()
	for i := 0; i < r.NumField(); i++ {
		env := r.Field(i).Tag.Get("mapstructure")
		if err := viper.BindEnv(env); err != nil {
			return err
		}
	}
	return viper.Unmarshal(cfg)
}

func main() {
	config, err := loadConfig(".")
	if err != nil {
		log.Panic(err)
	}

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	logger.Sugar().Infof("%+v", config)

	uploader, err := s3.New(config.AWSRegion, config.AWSBucket, config.CDNBase)
	if err != nil {
		log.Panic(err)
	}

	layerhub.RunRenderer(logger.Sugar())

	pixabayFeed := pixabay.NewImageFeed(config.PixabayKey)
	pexelsFeed := pexels.NewImageFeed(config.PixabayKey)
	paymentProvider := paypal.NewPaymentProvider(config.PaypalClientID, config.PaypalSecret)
	renderer := layerhub.NewRenderer(config.RendererSocket, logger.Sugar(), uploader)
	githubClient := github.NewClient(github.Config{
		ClientID:    config.GithubClientID,
		Secret:      config.GithubClientSecret,
		RedirectURI: config.GithubRedirectURI,
	})
	googleClient := google.NewClient(google.Config{
		ClientID:    config.GoogleClientID,
		Secret:      config.GoogleClientSecret,
		RedirectURI: config.GoogleRedirectURI,
	})

	mysqlDB, err := mysql.New(&mysql.Config{
		DSN:             config.MySQLDSN,
		ConnMaxIdleTime: 15 * time.Minute,
		MaxOpenConns:    10,
		MaxIdleConns:    5,
	})
	if err != nil {
		log.Panic(err)
	}

	mongoDB, err := mongodb.New(&mongodb.Config{
		URI: config.MongoURL,
		DB:  config.MongoDBName,
	})
	if err != nil {
		log.Panic(err)
	}
	defer mongoDB.Close(context.TODO())

	redisClient, err := redis.New(&redis.Config{
		Addr:         config.RedisAddr,
		Username:     config.RedisUsername,
		Password:     config.RedisPassword,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  5 * time.Minute,
	})
	if err != nil {
		log.Panic(err)
	}
	defer redisClient.Close(context.TODO())

	segClient := analytics.New(config.SegmentWriteKey)
	defer segClient.Close()

	server := http.NewServer(http.Config{
		Core: layerhub.New(layerhub.CoreConfig{
			Logger:          logger,
			DB:              mysqlDB,
			JSONDB:          mongoDB,
			Uploader:        uploader,
			Pixabay:         pixabayFeed,
			Pexels:          pexelsFeed,
			PaymentProvider: paymentProvider,
			Renderer:        renderer,
			GithubClient:    githubClient,
			GoogleClient:    googleClient,
		}),
		SessionDB:     redisClient,
		ReadTimeout:   15 * time.Second,
		WriteTimeout:  10 * time.Second,
		IdleTimeout:   15 * time.Minute,
		SegmentClient: segClient,
	})
	if err := server.ListenAndServe(":" + config.Port); err != nil {
		log.Panic(err)
	}
}
