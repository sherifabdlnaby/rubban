package config

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/joho/godotenv"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

func Load(configName string) (Config, error) {
	v := viper.New()
	var err error
	// load .env variables
	if _, err = os.Stat("./.env"); err == nil || !os.IsNotExist(err) {
		err := godotenv.Load()
		if err != nil {
			return Config{}, fmt.Errorf("error loading .env envvars: %w", err)
		}
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
		if _, err = os.Stat("."); err == nil || !os.IsNotExist(err) {
			v.AddConfigPath(".")
		}
	}

	err = v.ReadInConfig()
	if err != nil {
		return Config{}, fmt.Errorf("error reading config: %w", err)
	}

	// Unmarshalling
	cfg := Config{}
	err = v.Unmarshal(&cfg, viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToIPHookFunc(),
		StringJsonArrayOrSlicesToConfig(),
	)))

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

func StringJsonArrayOrSlicesToConfig() func(f reflect.Kind, t reflect.Kind, data interface{}) (interface{}, error) {
	return func(
		f reflect.Kind,
		t reflect.Kind,
		data interface{}) (interface{}, error) {
		if f != reflect.String || (t != reflect.Map && t != reflect.Slice) {
			return data, nil
		}

		raw := data.(string)
		if raw == "" {
			return []string{}, nil
		}

		var ret interface{}
		if t == reflect.Map {
			jsonMap := make(map[interface{}]interface{}, 0)
			err := json.Unmarshal([]byte(raw), &jsonMap)
			if err != nil {
				return raw, fmt.Errorf("couldn't map string-ifed Json to Map: %s", err.Error())
			}
			ret = jsonMap
		} else if t == reflect.Slice {
			jsonArray := make([]interface{}, 0)
			err := json.Unmarshal([]byte(raw), &jsonArray)
			if err != nil {
				// Try comma separated format too
				val, err := mapstructure.StringToSliceHookFunc(",").(func(f reflect.Kind, t reflect.Kind, data interface{}) (interface{}, error))(f, t, data)
				if err != nil {
					return val, err
				}
				ret = val
			}
			ret = jsonArray
		}

		return ret, nil
	}
}
