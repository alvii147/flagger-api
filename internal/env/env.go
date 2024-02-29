package env

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
)

// Config represents config variables for the server.
// Each field can be overridden using environment variables defined in field tags.
type Config struct {
	Hostname                string `env:"FLAGGERAPI_HOSTNAME"`
	Port                    int    `env:"FLAGGERAPI_PORT"`
	SecretKey               string `env:"FLAGGERAPI_SECRET_KEY"`
	HashingCost             int    `env:"FLAGGERAPI_HASHING_COST"`
	FrontendBaseURL         string `env:"FLAGGERAPI_FRONTEND_BASE_URL"`
	FrontendActivationRoute string `env:"FLAGGERAPI_FRONTEND_ACTIVATION_ROUTE"`
	AuthAccessLifetime      int64  `env:"FLAGGERAPI_AUTH_ACCESS_LIFETIME"`
	AuthRefreshLifetime     int64  `env:"FLAGGERAPI_AUTH_REFRESH_LIFETIME"`
	ActivationLifetime      int64  `env:"FLAGGERAPI_ACTIVATION_LIFETIME"`
	PostgresHostname        string `env:"FLAGGERAPI_POSTGRES_HOSTNAME"`
	PostgresPort            int    `env:"FLAGGERAPI_POSTGRES_PORT"`
	PostgresUsername        string `env:"FLAGGERAPI_POSTGRES_USERNAME"`
	PostgresPassword        string `env:"FLAGGERAPI_POSTGRES_PASSWORD"`
	PostgresDatabaseName    string `env:"FLAGGERAPI_POSTGRES_DATABASE_NAME"`
	SMTPHostname            string `env:"FLAGGERAPI_SMTP_HOSTNAME"`
	SMTPPort                int    `env:"FLAGGERAPI_SMTP_PORT"`
	SMTPUsername            string `env:"FLAGGERAPI_SMTP_USERNAME"`
	SMTPPassword            string `env:"FLAGGERAPI_SMTP_PASSWORD"`
	MailClientType          string `env:"FLAGGERAPI_MAIL_CLIENT_TYPE"`
}

// NewConfig reads environment variables and returns a new config
// and returns error when an environment variable is not found.
func NewConfig() (*Config, error) {
	config := &Config{}

	fields := reflect.TypeOf(config)
	values := reflect.ValueOf(config)

	for i := 0; i < fields.Elem().NumField(); i++ {
		field := fields.Elem().Field(i)
		value := values.Elem().Field(i)

		envKey, ok := field.Tag.Lookup("env")
		if !ok {
			continue
		}

		envValue, ok := os.LookupEnv(envKey)
		if !ok {
			return nil, fmt.Errorf("NewConfig failed, missing environment variable %s", envKey)
		}

		switch field.Type.Kind() {
		case reflect.String:
			value.SetString(envValue)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			envInt, err := strconv.ParseInt(envValue, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("NewConfig failed to strconv.ParseInt %s: %w", envValue, err)
			}

			value.SetInt(envInt)
		case reflect.Float32, reflect.Float64:
			envFloat, err := strconv.ParseFloat(envValue, 64)
			if err != nil {
				return nil, fmt.Errorf("NewConfig failed to strconv.ParseFloat %s: %w", envValue, err)
			}

			value.SetFloat(envFloat)
		case reflect.Bool:
			envBool, err := strconv.ParseBool(envValue)
			if err != nil {
				return nil, fmt.Errorf("NewConfig failed to strconv.ParseBool %s: %w", envValue, err)
			}

			value.SetBool(envBool)
		default:
			continue
		}
	}

	return config, nil
}
