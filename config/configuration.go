package config

//Configuration object for the pod-watcher application
type Configuration struct {
	Mail MailConfiguration `yaml:"mail"`
	SMTP SmtpConfiguration `yaml:"smtp"`
}
