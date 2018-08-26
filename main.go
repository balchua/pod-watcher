package main

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"pod-watcher/controller"
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
		clientset, err = getClient()
		if err != nil {
			logrus.Error(err)
			return err
		}

		controller.Start(clientset, "test")

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
