package config

//Config for Config Unmarshalling
type Config struct {
	Kibana              Kibana  `validate:"required"`
	Logging             Logging `validate:"required"`
	AutoIndexPattern    AutoIndexPattern
	RefreshIndexPattern RefreshIndexPattern
}

//Kibana for Config Unmarshalling
type Kibana struct {
	Host     string `validate:"required,uri"`
	User     string `validate:"required_with=password"`
	Password string `validate:"required_with=User"`
}

//GeneralPattern for Config Unmarshalling
type GeneralPattern struct {
	Pattern       string `validate:"required"`
	TimeFieldName string
}

//AutoIndexPattern for Config Unmarshalling
type AutoIndexPattern struct {
	Enabled         bool
	GeneralPatterns []GeneralPattern
	Schedule        string `validate:"required"`
}

//RefreshIndexPattern for Config Unmarshalling
type RefreshIndexPattern struct {
	Enabled  bool
	Patterns []string
	Schedule string `validate:"required"`
}

//Logging for Config Unmarshalling
type Logging struct {
	Level  string `validate:"required,oneof=debug info warn fatal panic"`
	Format string `validate:"required,oneof=console json logfmt"`
	Debug  bool
	Color  bool
}

//Default Return App Default Config
func Default() *Config {
	return &Config{
		Kibana: Kibana{
			Host:     "localhost:5601",
			User:     "elastic",
			Password: "changeme",
		},
		Logging: Logging{
			Level:  "info",
			Format: "json",
			Debug:  false,
			Color:  false,
		},
		AutoIndexPattern: AutoIndexPattern{
			Enabled:         false,
			GeneralPatterns: nil,
			Schedule:        "*/5 * * * *",
		},
		RefreshIndexPattern: RefreshIndexPattern{
			Enabled:  false,
			Patterns: nil,
			Schedule: "*/5 * * * *",
		},
	}
}
