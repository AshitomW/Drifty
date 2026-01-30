package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/AshitomW/Drifty/internal/models"
)

func formatTimestamp(timestamp float64) string {
	return time.Unix(int64(timestamp), 0).Format(time.RFC3339)
}

func (c *Collector) collectDockerConfig(ctx context.Context) (models.DockerConfig, error) {
	config := models.DockerConfig{
		Containers: make(map[string]models.Container),
		Images:     make(map[string]models.Image),
		Volumes:    make(map[string]models.Volume),
		Networks:   make(map[string]models.Network),
	}

	socketPath := c.config.Docker.SocketPath
	if socketPath == "" {
		socketPath = "/var/run/docker.sock"
	}

	if _, err := os.Stat(socketPath); os.IsNotExist(err) {
		return config, nil
	}

	client := &http.Client{}
	baseURL := "http://unix" + socketPath

	if c.config.Docker.Containers {
		containers, err := c.collectDockerContainers(ctx, client, baseURL)
		if err == nil {
			config.Containers = containers
		}
	}

	if c.config.Docker.Images {
		images, err := c.collectDockerImages(ctx, client, baseURL)
		if err == nil {
			config.Images = images
		}
	}

	if c.config.Docker.Volumes {
		volumes, err := c.collectDockerVolumes(ctx, client, baseURL)
		if err == nil {
			config.Volumes = volumes
		}
	}

	if c.config.Docker.Networks {
		networks, err := c.collectDockerNetworks(ctx, client, baseURL)
		if err == nil {
			config.Networks = networks
		}
	}

	return config, nil
}

func (c *Collector) collectDockerContainers(ctx context.Context, client *http.Client, baseURL string) (map[string]models.Container, error) {
	containers := make(map[string]models.Container)

	resp, err := client.Get(baseURL + "/containers/json?all=true")
	if err != nil {
		return containers, err
	}
	defer resp.Body.Close()

	var containerData []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&containerData); err != nil {
		return containers, err
	}

	for _, data := range containerData {
		id := ""
		name := ""
		image := ""
		state := ""
		status := ""
		created := ""

		if v, ok := data["Id"].(string); ok {
			id = v
		}
		if v, ok := data["Names"].([]interface{}); ok && len(v) > 0 {
			if n, ok := v[0].(string); ok {
				name = strings.TrimPrefix(n, "/")
			}
		}
		if v, ok := data["Image"].(string); ok {
			image = v
		}
		if v, ok := data["State"].(string); ok {
			state = v
		}
		if v, ok := data["Status"].(string); ok {
			status = v
		}
		if v, ok := data["Created"].(float64); ok {
			created = formatTimestamp(v)
		}

		ports := []string{}
		if p, ok := data["Ports"].([]interface{}); ok {
			for _, port := range p {
				if m, ok := port.(map[string]interface{}); ok {
					if ip, ok := m["IP"].(string); ok {
						if pubPort, ok := m["PublicPort"].(float64); ok {
							ports = append(ports, fmt.Sprintf("%s:%d", ip, int(pubPort)))
						}
					}
				}
			}
		}

		labels := make(map[string]string)
		if l, ok := data["Labels"].(map[string]interface{}); ok {
			for k, v := range l {
				if s, ok := v.(string); ok {
					labels[k] = s
				}
			}
		}

		containers[id] = models.Container{
			ID:      id,
			Name:    name,
			Image:   image,
			Status:  status,
			State:   state,
			Created: created,
			Ports:   ports,
			Labels:  labels,
		}
	}

	return containers, nil
}

func (c *Collector) collectDockerImages(ctx context.Context, client *http.Client, baseURL string) (map[string]models.Image, error) {
	images := make(map[string]models.Image)

	resp, err := client.Get(baseURL + "/images/json")
	if err != nil {
		return images, err
	}
	defer resp.Body.Close()

	var imageData []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&imageData); err != nil {
		return images, err
	}

	for _, data := range imageData {
		id := ""
		name := ""
		tag := ""
		size := int64(0)
		created := ""

		if v, ok := data["Id"].(string); ok {
			id = v
		}
		if v, ok := data["RepoTags"].([]interface{}); ok && len(v) > 0 {
			if rt, ok := v[0].(string); ok {
				parts := strings.Split(rt, ":")
				if len(parts) >= 2 {
					name = parts[0]
					tag = parts[1]
				}
			}
		}
		if v, ok := data["Size"].(float64); ok {
			size = int64(v)
		}
		if v, ok := data["Created"].(float64); ok {
			created = formatTimestamp(v)
		}

		labels := make(map[string]string)
		if l, ok := data["Labels"].(map[string]interface{}); ok {
			for k, v := range l {
				if s, ok := v.(string); ok {
					labels[k] = s
				}
			}
		}

		images[id] = models.Image{
			ID:      id,
			Name:    name,
			Tag:     tag,
			Size:    size,
			Created: created,
			Labels:  labels,
		}
	}

	return images, nil
}

func (c *Collector) collectDockerVolumes(ctx context.Context, client *http.Client, baseURL string) (map[string]models.Volume, error) {
	volumes := make(map[string]models.Volume)

	resp, err := client.Get(baseURL + "/volumes")
	if err != nil {
		return volumes, err
	}
	defer resp.Body.Close()

	var volumeData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&volumeData); err != nil {
		return volumes, err
	}

	if vols, ok := volumeData["Volumes"].([]interface{}); ok {
		for _, vol := range vols {
			if v, ok := vol.(map[string]interface{}); ok {
				name := ""
				driver := ""
				mountpoint := ""

				if n, ok := v["Name"].(string); ok {
					name = n
				}
				if d, ok := v["Driver"].(string); ok {
					driver = d
				}
				if m, ok := v["Mountpoint"].(string); ok {
					mountpoint = m
				}

				volumes[name] = models.Volume{
					Name:       name,
					Driver:     driver,
					Mountpoint: mountpoint,
				}
			}
		}
	}

	return volumes, nil
}

func (c *Collector) collectDockerNetworks(ctx context.Context, client *http.Client, baseURL string) (map[string]models.Network, error) {
	networks := make(map[string]models.Network)

	resp, err := client.Get(baseURL + "/networks")
	if err != nil {
		return networks, err
	}
	defer resp.Body.Close()

	var networkData []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&networkData); err != nil {
		return networks, err
	}

	for _, data := range networkData {
		id := ""
		name := ""
		driver := ""
		scope := ""
		subnet := ""

		if v, ok := data["Id"].(string); ok {
			id = v
		}
		if v, ok := data["Name"].(string); ok {
			name = v
		}
		if v, ok := data["Driver"].(string); ok {
			driver = v
		}
		if v, ok := data["Scope"].(string); ok {
			scope = v
		}
		if ipam, ok := data["IPAM"].(map[string]interface{}); ok {
			if config, ok := ipam["Config"].([]interface{}); ok && len(config) > 0 {
				if cfg, ok := config[0].(map[string]interface{}); ok {
					if s, ok := cfg["Subnet"].(string); ok {
						subnet = s
					}
				}
			}
		}

		labels := make(map[string]string)
		if l, ok := data["Labels"].(map[string]interface{}); ok {
			for k, v := range l {
				if s, ok := v.(string); ok {
					labels[k] = s
				}
			}
		}

		networks[id] = models.Network{
			ID:     id,
			Name:   name,
			Driver: driver,
			Scope:  scope,
			Subnet: subnet,
			Labels: labels,
		}
	}

	return networks, nil
}
