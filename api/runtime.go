package api

import (
	"github.com/ansriaz/redzilla/client"
	"github.com/ansriaz/redzilla/model"
)

//Reset reset container runtime information
func (i *Instance) Reset() error {

	i.instance.Url = ""
	i.instance.Status = model.InstanceUnavailable

	return nil
}

//GetIP return the container IP
// TODO return the url for the instance
func (i *Instance) GetUrl() (string, error) {
	return client.GetClient().GetInstanceUrl(i.instance.Name)
	// ip := ""
	//
	// if len(i.instance.IP) > 0 {
	// 	return i.instance.IP, nil
	// }

	//TODO define network resource fo container
	// net, err := docker.GetNetwork(i.cfg.Network)
	// if err != nil {
	// 	return ip, err
	// }
	//
	// for _, container := range net.Containers {
	// 	if container.Name == i.instance.Name {
	// 		ip = container.IPv4Address[:strings.Index(container.IPv4Address, "/")]
	// 		logrus.Debugf("Container IP %s", ip)
	// 		break
	// 	}
	// }
	//
	// if ip == "" {
	// 	return ip, fmt.Errorf("IP not found for container `%s`", i.instance.Name)
	// }
	//
	// i.instance.IP = ip
	//
	// return ip, nil
}
