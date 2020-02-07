package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func Load(configName string) (Config, error) {
	v := viper.New()

	// load .env variables
	err := godotenv.Load()
	if err != nil {
		return Config{}, fmt.Errorf("error loading .env envvars: %w", err)
	}

	// setup env configs
	envPrefix := strings.ToUpper(configName)
	v.SetEnvPrefix(envPrefix)

	// Set replacer
	// Viper add the `prefix` + '_' to the Key *before* passing it to Key Replacer,causing the replacer to replace the '_' with '__' when it shouldn't.
	// by adding the Prefix to the replacer twice, this will let the replacer escapes the prefix as it scans through the string.
	v.SetEnvKeyReplacer(strings.NewReplacer(envPrefix+"_", envPrefix+"_", "_", "__", ".", "_", "-", "_"))
	v.AutomaticEnv()

	// config name
	v.SetConfigName(strings.ToLower(configName))

	// declare config file path
	configLocationEnv := fmt.Sprintf("%s_CONFIG_DIR", strings.ToUpper(configName))
	if configDir, isSet := os.LookupEnv(configLocationEnv); isSet {
		v.AddConfigPath(configDir + "/")
	} else {
		v.AddConfigPath(".")
	}

	err = v.ReadInConfig()
	if err != nil {
		return Config{}, fmt.Errorf("error reading config: %w", err)
	}

	// Unmarshalling
	cfg := Config{}
	err = v.Unmarshal(&cfg)
	if err != nil {
		return Config{}, fmt.Errorf("error unmarshalling config: %w", err)
	}

	// Validate
	err = validate(cfg)
	if err != nil {
		return Config{}, err
	}

	return cfg, nil
}
