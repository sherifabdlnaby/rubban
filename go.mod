module github.com/sherifabdlnaby/bosun

go 1.13

replace github.com/census-instrumentation/opencensus-proto v0.1.0-0.20181214143942-ba49f56771b8 => github.com/census-instrumentation/opencensus-proto v0.0.3-0.20181214143942-ba49f56771b8

require (
	github.com/Masterminds/semver/v3 v3.0.3
	github.com/dustin/go-humanize v1.0.0
	github.com/go-playground/locales v0.13.0
	github.com/go-playground/universal-translator v0.17.0
	github.com/go-playground/validator v9.31.0+incompatible
	github.com/joho/godotenv v1.3.0
	github.com/leodido/go-urn v1.2.0 // indirect
	github.com/mitchellh/mapstructure v1.1.2
	github.com/robfig/cron/v3 v3.0.1
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.3.2
	github.com/sykesm/zap-logfmt v0.0.3
	go.uber.org/zap v1.13.0
	golang.org/x/sys v0.0.0-20190422165155-953cdadca894 // indirect
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gopkg.in/go-playground/validator.v9 v9.31.0
)
