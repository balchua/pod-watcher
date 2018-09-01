package main

import (
	"log"
	"os"

	"pod-watcher/config"
	"pod-watcher/controller"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/urfave/cli"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var clientset *kubernetes.Clientset

var namespace string

var pathToConfig string

var selectors string

var action string

func init() {
	// Log as JSON instead of the default ASCII formatter.
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true})

	// Output to stdout instead of the default stderr, could also be a file.
	logrus.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	logrus.SetLevel(logrus.DebugLevel)

}

func main() {
	app := cli.NewApp()
	app.Description = "Boink an application to bounce your apps in Kubernetes."
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "config, c",
			Usage:       "Kube config path for  outside of cluster access",
			Destination: &pathToConfig,
		},

		cli.StringFlag{
			Name:        "namespace, n",
			Value:       "default",
			Usage:       "the namespace  where the application will poll the service.",
			Destination: &namespace,
		},
		cli.StringFlag{
			Name:        "label, l",
			Value:       "default",
			Usage:       "The deployment selector based on the labels .",
			Destination: &selectors,
		},
	}

	app.Action = func(c *cli.Context) error {
		var err error
		var listOptions metav1.ListOptions

		clientset, err = getClient()
		if err != nil {
			logrus.Error(err)
			return err
		}

		if selectors != "" {
			listOptions = metav1.ListOptions{
				LabelSelector: selectors,
				Limit:         100,
			}

		} else {
			listOptions = metav1.ListOptions{}
		}

		controller.Start(clientset, namespace, listOptions, getConfig())

		if err != nil {
			panic(err)
		}
		return nil
	}
	app.Run(os.Args)
}

func getClient() (*kubernetes.Clientset, error) {
	var config *rest.Config
	var err error
	if pathToConfig == "" {
		logrus.Info("Using in cluster config")
		config, err = rest.InClusterConfig()
		// in cluster access
	} else {
		logrus.Info("Using out of cluster config")
		config, err = clientcmd.BuildConfigFromFlags("", pathToConfig)
	}
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

func getConfig() config.Configuration {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	var configuration config.Configuration

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}
	err := viper.Unmarshal(&configuration)
	if err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}
	logrus.Infof("SMTP host is %s", configuration.SMTP.Host)
	logrus.Infof("Mail configuration From: %s", configuration.Mail.From)
	return configuration
}
