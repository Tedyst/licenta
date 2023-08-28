package config

import (
	"github.com/spf13/viper"
	"github.com/tedyst/licenta/models"
	"github.com/tedyst/licenta/telemetry"
)

type Config struct {
	Telemetry telemetry.TelemetryConfig
	Models    models.ModelsConfig
	Debug     bool
}

func SetConfigDefaults(prefix string) {
	if prefix != "" {
		prefix = prefix + "."
	}
	viper.SetDefault(prefix+"debug", false)
	telemetry.SetConfigDefaults(prefix + "telemetry")
	models.SetConfigDefaults(prefix + "models")
}
