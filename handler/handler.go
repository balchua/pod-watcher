package handler

import (
	"github.com/Sirupsen/logrus"
	"k8s.io/api/core/v1"
)

// MailHandler object
type MailHandler struct {
	Subject string
	From    string
	To      string
	Body    string
}

// Notify do the actual sending of emails.
func (h *MailHandler) Notify(pod *v1.Pod) {
	//TODO send the mail from here.
	logrus.Infof("Pod (%s) installer failed.", pod.ObjectMeta.Name)

}
