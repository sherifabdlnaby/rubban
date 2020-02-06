package config

import (
	"fmt"
	"github.com/go-playground/validator"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"os"
	"strings"
)

func Load(configName string) (*Config, error) {
	v := viper.New()

	// load .env variables
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("error loading .env envvars: %w", err)
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
		return nil, fmt.Errorf("error reading config: %w", err)
	}

	// Unmarshalling
	cfg := Config{}
	err = v.Unmarshal(&cfg)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %w", err)
	}

	// Validate
	validate := validator.New()
	err = validate.Struct(cfg)
	if err != nil {
		return nil, fmt.Errorf("errors validating struct config: %s", err.Error())
	}

	errList := customValidate(&cfg)
	if len(errList) != 0 {
		return nil, fmt.Errorf("errors validating struct config: %v", errList)
	}

	return &cfg, nil
}
