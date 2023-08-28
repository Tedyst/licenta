package models

import "github.com/spf13/viper"

type ModelsConfig struct {
	User UserConfig `mapstructure:"user"`
}

type UserConfig struct {
	PasswordPepper string `mapstructure:"password_pepper"`
}

func SetConfigDefaults(prefix string) {
	viper.SetDefault(prefix+".user.password_pepper", "pepper")
}
