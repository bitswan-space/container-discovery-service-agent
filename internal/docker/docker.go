package docker

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)


type TopologyItem struct {
	Wires      []string           `json:"wires"`
	Properties TopologyProperties `json:"properties"`
	Metrics    []int              `json:"metrics"`
}


type TopologyProperties struct {
	ContainerID  string    `json:"container-id"`
	EndpointName string    `json:"endpoint-name"`
	DeploymentId string    `json:"deployment-id"`
	CreatedAt    time.Time `json:"created-at"`
	Name         string    `json:"name"`
	State        string    `json:"state"`
	Status       string    `json:"status"`
}

type Topology struct {
	Topology        map[string]TopologyItem `json:"topology"`
	DisplayStyle    string                  `json:"display-style"`
}


var cli *client.Client // Docker client

// Init initializes the Docker client
func Init() error {
	var err error
	cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	return nil
}

// Close closes the Docker client
func Close() {
	if cli != nil {
		cli.Close()
	}
}

func GetTopology() (*Topology, error) {
	var topology Topology = Topology{DisplayStyle: "list"}
	containers, err := cli.ContainerList(context.Background(), container.ListOptions{})
	if err != nil {
		return nil, err
	}

	info, err := cli.Info(context.Background())
	if err != nil {
		fmt.Println(err)
	}

	var topologyItems = make(map[string]TopologyItem)
	for _, container := range containers {
		// filter pumps
		if _, exists := container.Labels["space.bitswan.pipeline.protocol-version"]; exists {
			deploymentId := GetDeploymentId(container.ID)
			TopologyProperties := TopologyProperties{
				ContainerID:  container.ID,
				EndpointName: info.Name,
				CreatedAt:	time.Unix(container.Created, 0),
				Name:         container.Names[0],
				State:        container.State,
				Status:       container.Status,
				DeploymentId: deploymentId,
		}

		topologyItems[deploymentId] = TopologyItem{
			Wires:      []string{},
			Properties: TopologyProperties,
			Metrics:    []int{},
		}
		}
	}

	topology.Topology = topologyItems
	return &topology, nil

}

func GetDeploymentId(containerId string) string {
	inspect, err := cli.ContainerInspect(context.Background(), containerId)
	if err != nil {
		return ""
	}

	for _, envVar := range inspect.Config.Env {
		parts := strings.SplitN(envVar, "=", 2)
		if len(parts) == 2 && parts[0] == "DEPLOYMENT_ID" {
			return parts[1] // Return a pointer to the string
		}
	}
	return "" // DEPLOYMENT_ID not found, return nil
}