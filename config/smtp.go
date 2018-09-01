package config

type SmtpConfiguration struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}
