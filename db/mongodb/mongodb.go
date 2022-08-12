package mongodb

import (
	"context"

	"github.com/echovl/orderflo-dev/errors"
	"github.com/echovl/orderflo-dev/layerhub"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Config struct {
	URI string
	DB  string
}

type MongoDB struct {
	client *mongo.Client
	DB     string
}

// Verify that MongoStorage implements Storage
var _ layerhub.JSONDB = (*MongoDB)(nil)

func New(conf *Config) (layerhub.JSONDB, error) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(conf.URI))
	if err != nil {
		return nil, errors.E(errors.KindUnexpected, err)
	}

	return &MongoDB{client, conf.DB}, nil
}

func (s *MongoDB) PutTemplate(ctx context.Context, template *layerhub.Template) error {
	collection := s.client.Database(s.DB).Collection("templates")

	filter := bson.M{"_id": template.ID}
	query := bson.M{"$set": template}
	opts := options.Update().SetUpsert(true)
	_, err := collection.UpdateOne(ctx, filter, query, opts)
	if err != nil {
		return errors.E(errors.KindUnexpected, err)
	}

	return nil
}

func (s *MongoDB) DeleteTemplate(ctx context.Context, id string) error {
	collection := s.client.Database(s.DB).Collection("templates")

	filter := bson.M{"_id": id}
	_, err := collection.DeleteOne(ctx, filter)
	if err == mongo.ErrNoDocuments {
		return errors.E(errors.KindNotFound, err)
	}
	if err != nil {
		return errors.E(errors.KindUnexpected, err)
	}

	return nil
}

func (s *MongoDB) FindTemplates(ctx context.Context, filter *layerhub.Filter) ([]layerhub.Template, error) {
	collection := s.client.Database(s.DB).Collection("templates")

	templates := []layerhub.Template{}
	parsedFilter := parseFilters(filter)
	cursor, err := collection.Find(ctx, parsedFilter)
	if err != nil {
		return nil, errors.E(errors.KindUnexpected, err)
	}

	err = cursor.All(ctx, &templates)
	if err != nil {
		return nil, errors.E(errors.KindUnexpected, err)
	}

	return templates, nil
}

func (s *MongoDB) Close(ctx context.Context) error {
	return s.client.Disconnect(ctx)
}

func parseFilters(filter *layerhub.Filter) bson.M {
	parsedFilter := bson.M{}
	if filter != nil {
		if filter.ID != "" {
			parsedFilter["_id"] = filter.ID
		}
		if filter.UserID != "" {
			parsedFilter["user_id"] = filter.UserID
		}
	}
	return parsedFilter
}
