package config

func (c *Config) SetEnvBindings() map[string]string {
	return map[string]string{
		// App
		"app.name":                   "APP_NAME",
		"app.version":                "APP_VERSION",
		"app.environment":            "APP_ENVIRONMENT",
		"app.features.recovery":      "APP_FEATURES_RECOVERY",
		"app.features.tracing":       "APP_FEATURES_TRACING",
		"app.features.metrics":       "APP_FEATURES_METRICS",
		"app.features.profiling":     "APP_FEATURES_PROFILING",
		"app.logging.level":          "APP_LOGGING_LEVEL",
		"app.logging.format":         "APP_LOGGING_FORMAT",
		"app.logging.output":         "APP_LOGGING_OUTPUT",
		"app.security_headers":       "APP_SECURITY_HEADERS",
		"app.cors.allow_origins":     "APP_CORS_ALLOW_ORIGINS",
		"app.cors.allow_methods":     "APP_CORS_ALLOW_METHODS",
		"app.cors.allow_headers":     "APP_CORS_ALLOW_HEADERS",
		"app.cors.expose_headers":    "APP_CORS_EXPOSE_HEADERS",
		"app.cors.allow_credentials": "APP_CORS_ALLOW_CREDENTIALS",
		"app.cors.max_age":           "APP_CORS_MAX_AGE",
		"app.rate_limit.enabled":     "APP_RATE_LIMIT_ENABLED",
		"app.rate_limit.max":         "APP_RATE_LIMIT_MAX",
		"app.module.base_path":       "APP_MODULE_BASE_PATH",
		"app.module.disabled":        "APP_MODULE_DISABLED",

		// Server
		"server.host":          "SERVER_HOST",
		"server.port":          "SERVER_PORT",
		"server.path":          "SERVER_PATH",
		"server.read_timeout":  "SERVER_READ_TIMEOUT",
		"server.write_timeout": "SERVER_WRITE_TIMEOUT",

		// Auth
		"auth.control":        "AUTH_CONTROL",
		"auth.store":          "AUTH_STORE",
		"auth.type":           "AUTH_TYPE",
		"auth.secret_key":     "AUTH_SECRET_KEY",
		"auth.expires_in":     "AUTH_EXPIRES_IN",
		"auth.api_key_header": "AUTH_API_KEY_HEADER",
		"auth.api_key_name":   "AUTH_API_KEY_NAME",

		// Database
		"database.driver":            "DATABASE_DRIVER",
		"database.host":              "DATABASE_HOST",
		"database.port":              "DATABASE_PORT",
		"database.user":              "DATABASE_USER",
		"database.password":          "DATABASE_PASSWORD",
		"database.name":              "DATABASE_NAME",
		"database.ssl_mode":          "DATABASE_SSL_MODE",
		"database.max_open_conns":    "DATABASE_MAX_OPEN_CONNS",
		"database.max_idle_conns":    "DATABASE_MAX_IDLE_CONNS",
		"database.conn_max_lifetime": "DATABASE_CONN_MAX_LIFETIME",

		// Redis
		"redis.host":     "REDIS_HOST",
		"redis.port":     "REDIS_PORT",
		"redis.password": "REDIS_PASSWORD",
		"redis.db":       "REDIS_DB",

		// Kafka
		"kafka.brokers":      "KAFKA_BROKERS",
		"kafka.group_id":     "KAFKA_GROUP_ID",
		"kafka.topics":       "KAFKA_TOPICS",
		"kafka.offset_reset": "KAFKA_AUTO_OFFSET_RESET",

		// PubSub
		"pubsub.project_id":   "PROJECT_ID",
		"pubsub.topic":        "PUBSUB_TOPIC",
		"pubsub.subscription": "PUBSUB_SUBSCRIPTION",
		"pubsub.credentials":  "PUBSUB_CREDENTIALS",

		// PubSub Consumer
		"pubsub.consumer.maxmessages":             "PUBSUB_CONSUMER_MAXMESSAGES",
		"pubsub.consumer.sleeptime":               "PUBSUB_CONSUMER_SLEEPTIME",
		"pubsub.consumer.acktimeout":              "PUBSUB_CONSUMER_ACKTIMEOUT",
		"pubsub.consumer.retrycount":              "PUBSUB_CONSUMER_RETRYCOUNT",
		"pubsub.consumer.retrydelay":              "PUBSUB_CONSUMER_RETRYDELAY",
		"pubsub.consumer.flowcontrol.enabled":     "PUBSUB_CONSUMER_FLOWCONTROL_ENABLED",
		"pubsub.consumer.flowcontrol.maxmessages": "PUBSUB_CONSUMER_FLOWCONTROL_MAXMESSAGES",
		"pubsub.consumer.flowcontrol.maxbytes":    "PUBSUB_CONSUMER_FLOWCONTROL_MAXBYTES",

		// PubSub Producer
		"pubsub.producer.enableordering": "PUBSUB_PRODUCER_ENABLEORDERING",
		"pubsub.producer.batchsize":      "PUBSUB_PRODUCER_BATCHSIZE",
		"pubsub.producer.attributes":     "PUBSUB_PRODUCER_ATTRIBUTES",
	}
}
