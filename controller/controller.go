package controller

import (
	"fmt"
	"os"
	"os/signal"
	"pod-watcher/config"
	"pod-watcher/handler"
	"syscall"
	"time"

	"github.com/Sirupsen/logrus"

	api_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const maxRetries = 5

// Controller object
type Controller struct {
	clientset kubernetes.Interface
	queue     workqueue.RateLimitingInterface
	informer  cache.SharedIndexInformer
	handler   handler.MailHandler
	config    config.Configuration
}

/*
Start function starts setting up the informer.
This method also find pods with label appType=installer
*/
func Start(kubeClient *kubernetes.Clientset, namespace string, listOptions meta_v1.ListOptions, config config.Configuration) {

	informer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options meta_v1.ListOptions) (runtime.Object, error) {
				return kubeClient.CoreV1().Pods(namespace).List(listOptions)
			},
			WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
				return kubeClient.CoreV1().Pods(namespace).Watch(listOptions)
			},
		},
		&api_v1.Pod{},
		0, //Skip resync
		cache.Indexers{},
	)

	c := newResourceController(kubeClient, informer, config)
	stopCh := make(chan struct{})
	defer close(stopCh)

	go c.Run(stopCh)

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGTERM)
	signal.Notify(sigterm, syscall.SIGINT)
	<-sigterm
}

// Only act on when Pod is updated.
func newResourceController(client kubernetes.Interface, informer cache.SharedIndexInformer, config config.Configuration) *Controller {
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{

		UpdateFunc: func(old, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(old)
			if err == nil {
				logrus.Info("Adding to Queue")
				queue.Add(key)
			}
		},
	})

	handler := handler.NewMailHandler(config.SMTP.Host, config.SMTP.Port, "", "")

	return &Controller{
		clientset: client,
		informer:  informer,
		queue:     queue,
		handler:   handler,
		config:    config,
	}
}

// Run starts the podwatcher controller
func (c *Controller) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	logrus.Info("Starting pod-watcher controller")

	go c.informer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, c.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	logrus.Info("pod-watcher controller synced and ready")

	wait.Until(c.runWorker, time.Second, stopCh)
}

// HasSynced is required for the cache.Controller interface.
func (c *Controller) HasSynced() bool {
	return c.informer.HasSynced()
}

// LastSyncResourceVersion is required for the cache.Controller interface.
func (c *Controller) LastSyncResourceVersion() string {
	return c.informer.LastSyncResourceVersion()
}

func (c *Controller) runWorker() {
	for c.processNextItem() {
		// continue looping
	}
}

func (c *Controller) processNextItem() bool {
	key, quit := c.queue.Get()

	if quit {
		return false
	}

	defer c.queue.Done(key)
	err := c.processItem(key.(string))

	if err == nil {
		// No error, reset the ratelimit counters
		c.queue.Forget(key)
	} else if c.queue.NumRequeues(key) < maxRetries {
		c.queue.AddRateLimited(key)
	} else {
		// err != nil and too many retries
		c.queue.Forget(key)
		utilruntime.HandleError(err)
	}

	return true
}

func (c *Controller) processItem(key string) error {
	obj, _, err := c.informer.GetIndexer().GetByKey(key)
	if err != nil {
		return fmt.Errorf("Error fetching object with key %s from store: %v", key, err)
	}

	if obj != nil {
		pod := obj.(*api_v1.Pod)
		logrus.Infof("Processing the pod %s with status %s", pod.ObjectMeta.Name, pod.Status.Phase)
		if pod.Status.Phase == api_v1.PodFailed ||
			pod.Status.Phase == api_v1.PodSucceeded {

			subject := fmt.Sprintf("Pod (%s) on namespace (%s) status is %s", pod.ObjectMeta.Name,
				pod.ObjectMeta.Namespace, pod.Status.Phase)
				
			body := fmt.Sprintf("Pod (%s) status is %s due to %s", pod.ObjectMeta.Name, pod.Status.Phase, pod.Status.Reason)
			message := handler.NewMail(c.config.Mail.From, c.config.Mail.To, subject, body)
			c.handler.Notify(message)
		}
	}

	return nil
}
