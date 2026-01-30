package collector

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/AshitomW/Drifty/internal/models"
	"github.com/google/uuid"
)

type Collector struct {
	config  models.CollectorConfig
	workers int
}

// Creating a new Coollcter instance
func New(config models.CollectorConfig) *Collector {
	workers := runtime.NumCPU()
	if workers < 2 {
		workers = 2
	}

	return &Collector{
		config:  config,
		workers: workers,
	}

}

// Collect will gather complete environment snapshot

func (c *Collector) Collect(ctx context.Context, name string) (*models.EnvironmentSnapshot, error) {

	hostname, _ := os.Hostname()

	snapshot := &models.EnvironmentSnapshot{
		ID:             uuid.New().String(),
		Name:           name,
		Hostname:       hostname,
		Timestamp:      time.Now().UTC(),
		OS:             c.collectOSInfo(),
		Files:          make(map[string]models.FileInfo),
		EnvVars:        make(map[string]models.EnvVar),
		ProcessEnvVars: make(map[int]models.ProcessEnvVar),
		Packages:       make(map[string]models.PackageInfo),
		Services:       make(map[string]models.ServiceInfo),
		Metadata:       make(map[string]string),
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	errChan := make(chan error, 12)

	// Collect Files Concurrently
	if c.config.Files.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			files, err := c.collectFiles(ctx)
			if err != nil {
				errChan <- fmt.Errorf("File Collection: %w", err)
				return
			}

			mu.Lock()
			snapshot.Files = files
			mu.Unlock()
		}()
	}

	// Collect Environment Variables

	if c.config.EnvVars.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			envVars, err := c.collectEnvVars(ctx)
			if err != nil {
				errChan <- fmt.Errorf("env var collection: %w", err)
				return
			}
			mu.Lock()
			snapshot.EnvVars = envVars
			mu.Unlock()
		}()
	}

	if c.config.ProcessEnvVars.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			processEnvVars, err := c.collectProcessEnvVars(ctx)
			if err != nil {
				errChan <- fmt.Errorf("process env var collection: %w", err)
				return
			}
			mu.Lock()
			snapshot.ProcessEnvVars = processEnvVars
			mu.Unlock()
		}()
	}

	// Collect Packages

	if c.config.Packages.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			packages, err := c.collectPackages(ctx)
			if err != nil {
				errChan <- fmt.Errorf("package collection: %w", err)
				return
			}
			mu.Lock()
			snapshot.Packages = packages
			mu.Unlock()
		}()
	}

	// Collect Servicess

	if c.config.Services.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			services, err := c.collectServices(ctx)
			if err != nil {
				errChan <- fmt.Errorf("service collection: %w", err)
				return
			}

			mu.Lock()
			snapshot.Services = services
			mu.Unlock()

		}()
	}

	if c.config.Network.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			network, err := c.collectNetworkConfig(ctx)
			if err != nil {
				errChan <- fmt.Errorf("network collection: %w", err)
				return
			}
			mu.Lock()
			snapshot.NetworkConfig = network
			mu.Unlock()
		}()
	}

	if c.config.Docker.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			docker, err := c.collectDockerConfig(ctx)
			if err != nil {
				errChan <- fmt.Errorf("docker collection: %w", err)
				return
			}
			mu.Lock()
			snapshot.DockerConfig = docker
			mu.Unlock()
		}()
	}

	if c.config.SystemResources.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resources, err := c.collectSystemResources(ctx)
			if err != nil {
				errChan <- fmt.Errorf("system resources collection: %w", err)
				return
			}
			mu.Lock()
			snapshot.SystemResources = resources
			mu.Unlock()
		}()
	}

	if c.config.ScheduledTasks.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tasks, err := c.collectScheduledTasks(ctx)
			if err != nil {
				errChan <- fmt.Errorf("scheduled tasks collection: %w", err)
				return
			}
			mu.Lock()
			snapshot.ScheduledTasks = tasks
			mu.Unlock()
		}()
	}

	if c.config.Certificates.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			certs, err := c.collectCertificates(ctx)
			if err != nil {
				errChan <- fmt.Errorf("certificates collection: %w", err)
				return
			}
			mu.Lock()
			snapshot.Certificates = certs
			mu.Unlock()
		}()
	}

	if c.config.UsersGroups.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			users, err := c.collectUserGroupConfig(ctx)
			if err != nil {
				errChan <- fmt.Errorf("user/group collection: %w", err)
				return
			}
			mu.Lock()
			snapshot.UserGroupConfig = users
			mu.Unlock()
		}()
	}

	wg.Wait()
	close(errChan)

	// Collect errors

	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errChan) > 0 {
		return snapshot, fmt.Errorf("collection errors: %v", errors)
	}

	return snapshot, nil
}

// Collecting OS Information

func (c *Collector) collectOSInfo() models.OSInfo {
	return models.OSInfo{
		Name:    runtime.GOOS,
		Arch:    runtime.GOARCH,
		Version: getOSVersion(),
		Kernel:  getKernelVersion(),
	}
}
