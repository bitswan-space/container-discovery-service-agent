package docker

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"bitswan.space/container-discovery-service-agent/internal/config"
	"bitswan.space/container-discovery-service-agent/internal/logger"
	"bitswan.space/container-discovery-service-agent/internal/mqtt"
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
	Topology     map[string]TopologyItem `json:"topology"`
	DisplayStyle string                  `json:"display-style"`
}

var cli *client.Client // Docker client
var cfg *config.Configuration

// Init initializes the Docker client
func Init() error {
	var err error
	cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	cfg = config.GetConfig()

	return nil
}

// Close closes the Docker client
func Close() {
	if cli != nil {
		cli.Close()
	}
}

func SendTopology() {
	var topology Topology = Topology{DisplayStyle: "list"}
	containers, err := cli.ContainerList(context.Background(), container.ListOptions{})
	if err != nil {
		logger.Error.Printf("Failed to list containers: %v", err)
	}

	info, err := cli.Info(context.Background())
	if err != nil {
		logger.Error.Printf("Failed to get Docker info: %v", err)
	}

	var topologyItems = make(map[string]TopologyItem)
	for _, container := range containers {
		// filter pumps
		if _, exists := container.Labels["space.bitswan.pipeline.protocol-version"]; exists {
			if _, exists := container.Labels["gitops.deployment_id"]; !exists {
				continue
			}

			deploymentId := container.Labels["gitops.deployment_id"]

			TopologyProperties := TopologyProperties{
				ContainerID:  container.ID,
				EndpointName: info.Name,
				CreatedAt:    time.Unix(container.Created, 0),
				Name:         strings.Replace(container.Names[0], "/", "", -1),
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

	b, err := json.MarshalIndent(topology, "", "  ")
	if err != nil {
		logger.Error.Println(err)
		return
	}

	mqtt.Publish(cfg.TopologyTopic, string(b))

}
