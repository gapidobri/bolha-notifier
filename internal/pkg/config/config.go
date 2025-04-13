package config

type Config struct {
	Urls          []string `mapstructure:"urls" validate:"required,dive,url"`
	WebhookUrl    string   `mapstructure:"webhook_url" validate:"required,url"`
	CheckInterval int      `mapstructure:"check_interval" validate:"omitempty,numeric"`
}
