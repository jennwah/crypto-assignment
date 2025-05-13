package config

type Postgres struct {
	Port       int64  `envconfig:"X_POSTGRESQL_PORT"      default:"5432"`
	Host       string `envconfig:"X_POSTGRESQL_HOST"                     required:"true"`
	User       string `envconfig:"X_POSTGRESQL_USER"                     required:"true"`
	Password   string `envconfig:"X_POSTGRESQL_PASSWORD"                 required:"true"`
	DBName     string `envconfig:"X_POSTGRESQL_DB_NAME"                  required:"true"`
	SchemaName string `envconfig:"X_POSTGRESQL_DB_SCHEMA"                required:"true"`
}
