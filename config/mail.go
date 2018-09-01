package config

type MailConfiguration struct {
	From string `yaml:"from"`
	To   string `yaml:"to"`
}
