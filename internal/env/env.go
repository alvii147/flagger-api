package env

import (
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

var config *Config

// GetConfig returns currently set Config.
func GetConfig() *Config {
	return config
}

// SetConfig sets current Config.
func SetConfig(c *Config) {
	config = c
}

// defaultConfig returns config struct with default values if the field is not provided.
func defaultConfig() *Config {
	return &Config{
		Hostname:                "localhost",
		Port:                    8080,
		SecretKey:               "DEADBEEF",
		HashingCost:             14,
		FrontendBaseURL:         "http://localhost:3000",
		FrontendActivationRoute: "/signup/activate/%s",
		AuthAccessLifetime:      30,
		AuthRefreshLifetime:     30 * 24 * 60,
		ActivationLifetime:      30 * 24 * 60,
		PostgresHostname:        "localhost",
		PostgresPort:            5432,
		PostgresUsername:        "postgres",
		PostgresPassword:        "postgres",
		PostgresDatabaseName:    "flaggerdb",
		SMTPHostname:            "smtp.gmail.com",
		SMTPPort:                587,
		SMTPUsername:            "",
		SMTPPassword:            "",
		MailClientType:          "console",
	}
}

// NewConfig reads environment variables and returns a new config,
// overridden by environment variables where possible.
func NewConfig() *Config {
	config := defaultConfig()

	fields := reflect.TypeOf(config)
	values := reflect.ValueOf(config)

	for i := 0; i < fields.Elem().NumField(); i++ {
		field := fields.Elem().Field(i)
		value := values.Elem().Field(i)

		envKey, ok := field.Tag.Lookup("env")
		if !ok {
			continue
		}

		overrideValue, ok := os.LookupEnv(envKey)
		if !ok {
			continue
		}

		switch field.Type.Kind() {
		case reflect.String:
			value.SetString(overrideValue)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			overrideInt, err := strconv.ParseInt(overrideValue, 10, 64)
			if err != nil {
				continue
			}

			value.SetInt(overrideInt)
		case reflect.Float32, reflect.Float64:
			overrideFloat, err := strconv.ParseFloat(overrideValue, 64)
			if err != nil {
				continue
			}

			value.SetFloat(overrideFloat)
		case reflect.Bool:
			overrideBool, err := strconv.ParseBool(overrideValue)
			if err != nil {
				continue
			}

			value.SetBool(overrideBool)
		default:
			continue
		}
	}

	return config
}
