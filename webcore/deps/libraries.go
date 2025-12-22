package deps

import (
	"github.com/semanggilab/webcore-go/app/core"
	"github.com/semanggilab/webcore-go/lib/auth/apikey"
	"github.com/semanggilab/webcore-go/lib/authstore/yaml"
	"github.com/semanggilab/webcore-go/lib/mongo"
	"github.com/semanggilab/webcore-go/lib/pubsub"
)

var APP_LIBRARIES = map[string]core.LibraryLoader{
	// "db:postgres":     &postgres.PostgresLoader{},
	"db:mongodb": &mongo.MongoLoader{},
	// "redis":           &redis.RedisLoader{},
	"pubsub":          &pubsub.PubSubLoader{},
	"auth.store:yaml": &yaml.YamlLoader{},
	"authn:apikey":    &apikey.ApiKeyLoader{},

	// Add your library here
}
