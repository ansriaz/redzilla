package client

import (
	"github.com/ansriaz/redzilla/model"
)

type Client interface {
	Init(*model.Config) error
	// Connect() error
	// Log(FUNCTION)
	DeployInstance(name string) error
	StopInstance(name string) error
	DeleteInstance(name string) error
	GetInstanceStatus(name string) (model.InstanceStatus, error)
	UpdateInstanceInformation(instance *model.Instance) error
	GetInstanceUrl(name string) (string, error)
}

var client Client

func SetClient(c Client) {
	client = c
}

func GetClient() (c Client) {
	return client
}
