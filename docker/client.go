package docker

import (
	"bufio"
	"errors"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ansriaz/redzilla/model"
	"github.com/ansriaz/redzilla/storage"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/sirupsen/logrus"

	"golang.org/x/net/context"
)

//ContainerEvent store a container event
type ContainerEvent struct {
	ID      string
	Name    string
	Action  string
	Message events.Message
}

var eventsChannel = make(chan ContainerEvent)
var dockerClient *client.Client

//GetEventsChannel return the main channel reporting docker events
func GetEventsChannel() <-chan ContainerEvent {
	return eventsChannel
}

// ListenEvents watches docker events an handle state modifications
func ListenEvents(cfg *model.Config) <-chan ContainerEvent {

	cli, err := getClient()
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	// ctx1 := context.Background()
	// ctx, cancel := context.WithCancel(ctx1)

	f := filters.NewArgs()
	f.Add("label", "redzilla=1")
	// <-chan events.Message, <-chan error
	msgChan, errChan := cli.Events(ctx, types.EventsOptions{
		Filters: f,
	})

	go func() {
		for {
			select {
			case event := <-msgChan:
				if &event != nil {

					logrus.Infof("Event recieved: %s %s ", event.Action, event.Type)
					if event.Actor.Attributes != nil {

						// logrus.Infof("%s: %s | %s | %s | %s | %s", event.Actor.Attributes["name"], event.Action, event.From, event.ID, event.Status, event.Type)

						name := event.Actor.Attributes["name"]
						switch event.Action {
						case "start":
							logrus.Debugf("Container started %s", name)
							break
						case "die":
							logrus.Debugf("Container exited %s", name)
							break
						}

						ev := ContainerEvent{
							Action:  event.Action,
							ID:      event.ID,
							Name:    name,
							Message: event,
						}
						eventsChannel <- ev

					}
				}
			case err := <-errChan:
				if err != nil {
					logrus.Errorf("Error event recieved: %s", err.Error())
				}
			}
		}
	}()

	return eventsChannel
}

//return a docker client
func getClient() (*client.Client, error) {

	if dockerClient == nil {
		cli, err := client.NewEnvClient()
		if err != nil {
			return nil, err
		}
		dockerClient = cli
	}

	return dockerClient, nil
}

func extractEnv(cfg *model.Config) []string {

	env := make([]string, 0)

	vars := os.Environ()

	envPrefix := strings.ToLower(cfg.EnvPrefix)
	pl := len(envPrefix)

	if pl > 0 {
		for _, e := range vars {

			if pl > 0 {
				if pl > len(e) {
					continue
				}
				if strings.ToLower(e[0:pl]) != envPrefix {
					continue
				}
			}

			//removed PREFIX_
			envVar := e[pl+1:]
			env = append(env, envVar)
		}

	}

	return env
}

//StartContainer start a container
func StartContainer(name string, cfg *model.Config) error {

	logrus.Debugf("Starting docker container %s", name)

	cli, err := getClient()
	if err != nil {
		return err
	}

	// containerID := "red3"
	// options := types.ContainerStartOptions{}
	ctx := context.Background()

	logrus.Debugf("Pulling image %s if not available", cfg.ImageName)
	_, err = cli.ImagePull(ctx, cfg.ImageName, types.ImagePullOptions{})
	if err != nil {
		return err
	}

	logrus.Debugf("Pulled image %s", cfg.ImageName)

	info, err := GetContainer(name)
	if err != nil {
		return err
	}

	exists := info.ContainerJSONBase != nil
	logrus.Debugf("Container %s exists: %t", name, exists)

	var containerID string

	if !exists {

		labels := map[string]string{
			"redzilla":          "1",
			"redzilla_instance": "redzilla_" + name,
		}
		exposedPorts := nat.PortSet{
			"1880/tcp": {},
		}

		instanceConfigPath := storage.GetConfigPath(cfg)
		instanceDataPath := storage.GetInstancesDataPath(name, cfg)
		binds := []string{
			instanceDataPath + ":/data",
			instanceConfigPath + ":/config",
		}

		envVars := extractEnv(cfg)

		logrus.Debugf("Creating new container %s ", name)
		logrus.Debugf("Bind paths: %v", binds)
		logrus.Debugf("Env: %v", envVars)

		resp, err1 := cli.ContainerCreate(ctx,
			&container.Config{
				User:         strconv.Itoa(os.Getuid()), // avoid permission issues
				Image:        cfg.ImageName,
				AttachStdin:  false,
				AttachStdout: true,
				AttachStderr: true,
				Tty:          true,
				ExposedPorts: exposedPorts,
				Labels:       labels,
				Env:          envVars,
			},
			&container.HostConfig{
				Binds:       binds,
				NetworkMode: container.NetworkMode(cfg.Network),
				PortBindings: nat.PortMap{
					"1880": []nat.PortBinding{
						nat.PortBinding{
							HostIP:   "",
							HostPort: "1880",
						},
					}},
				AutoRemove: true,
				// Links           []string          // List of links (in the name:alias form)
				// PublishAllPorts bool              // Should docker publish all exposed port for the container
				// Mounts []mount.Mount `json:",omitempty"`
			},
			nil, // &network.NetworkingConfig{},
			name,
		)
		if err1 != nil {
			return err1
		}

		containerID = resp.ID
		logrus.Debugf("Created new container %s", name)
	} else {
		containerID = info.ContainerJSONBase.ID
		logrus.Debugf("Reusing container %s", name)
	}

	logrus.Debugf("Container %s with ID %s", name, containerID)

	if err = cli.ContainerStart(ctx, containerID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	logrus.Debugf("Started container %s", name)

	return nil
}

// ContainerWatchLogs pipe logs from the container instance
func ContainerWatchLogs(ctx context.Context, name string, writer io.Writer) error {

	cli, err := getClient()
	if err != nil {
		return err
	}

	info, err := GetContainer(name)
	if err != nil {
		return err
	}
	if info.ContainerJSONBase == nil {
		return errors.New("Container not found " + name)
	}

	containerID := info.ContainerJSONBase.ID

	out, err := cli.ContainerLogs(ctx, containerID, types.ContainerLogsOptions{
		ShowStderr: true,
		ShowStdout: true,
		Follow:     true,
	})

	if err != nil {
		logrus.Warnf("Failed to open logs %s: %s", name, err.Error())
		return err
	}

	// if logrus.GetLevel() == logrus.DebugLevel {
	go func() {
		logrus.Debug("Printing instances log")
		buf := bufio.NewScanner(out)
		for buf.Scan() {
			logrus.Debugf("%s", buf.Text())
		}
	}()
	// }

	go func() {
		// pipe stream, will stop when container stops
		if _, err := io.Copy(writer, out); err != nil {
			logrus.Warnf("Error copying log stream %s", name)
		}
	}()

	return nil
}

//StopContainer stop a container
func StopContainer(name string) error {

	logrus.Debugf("Stopping container %s", name)

	cli, err := getClient()
	if err != nil {
		return err
	}

	ctx := context.Background()

	info, err := GetContainer(name)
	if err != nil {
		return err
	}

	if info.ContainerJSONBase == nil {
		logrus.Warnf("Cannot stop %s, does not exists", name)
		return nil
	}

	containerID := info.ContainerJSONBase.ID
	timeout := time.Second * 5

	err = cli.ContainerStop(ctx, containerID, &timeout)
	if err != nil {
		return err
	}

	logrus.Debugf("Stopped container %s", name)
	return nil
}

// GetContainer return container info by name
func GetContainer(name string) (*types.ContainerJSON, error) {

	ctx := context.Background()
	emptyJSON := &types.ContainerJSON{}

	if len(name) == 0 {
		return emptyJSON, errors.New("GetContainer(): name is empty")
	}

	cli, err := getClient()
	if err != nil {
		return emptyJSON, err
	}

	json, err := cli.ContainerInspect(ctx, name)
	if err != nil {
		if client.IsErrNotFound(err) {
			return emptyJSON, nil
		}
		return emptyJSON, err
	}

	return &json, nil
}

//GetNetwork inspect a network by networkID
func GetNetwork(networkID string) (*types.NetworkResource, error) {

	n := types.NetworkResource{}

	cli, err := getClient()
	if err != nil {
		return &n, err
	}

	ctx := context.Background()
	n, err = cli.NetworkInspect(ctx, networkID, types.NetworkInspectOptions{})
	if err != nil {
		return &n, err
	}

	return &n, nil
}
