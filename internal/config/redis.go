package config

type Redis struct {
	RedisHost string `envconfig:"X_REDIS_HOST" default:"localhost"`
	RedisPort string `envconfig:"X_REDIS_PORT" default:"6379"`
	RedisPass string `envconfig:"X_REDIS_PASS"                     required:"true"`
}
