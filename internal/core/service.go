package core

import (
	"context"
	"io"
	"sync"
	"time"
)

//go:generate mockgen -source=service.go -destination=mock_repository_test.go -package=core Repository
type Repository interface {
	LoadContainerRows(ctx context.Context) ([]ContainerRow, error)
	LoadSupportingResources(ctx context.Context) (Snapshot, error)
	LoadContainerMetrics(ctx context.Context, rows []ContainerRow) (map[string]ContainerMetrics, float64, uint64, error)
	LoadContainerLiveData(ctx context.Context, containerID string, previousCPU, previousMem []float64, tail int) (ContainerLiveData, error)
	LoadContainerLogs(ctx context.Context, containerID string, tail int) ([]string, error)
	InspectContainer(ctx context.Context, containerID string) ([]string, error)
	InspectImage(ctx context.Context, imageRef string) ([]string, error)
	InspectNetwork(ctx context.Context, networkID string) ([]string, error)
	InspectVolume(ctx context.Context, volumeName string) ([]string, error)
	ExecShell(ctx context.Context, containerID string, stdin io.Reader, stdout, stderr io.Writer) error
}

type ServiceConfig struct {
	RequestTimeout              time.Duration
	LiveDataMediumTailThreshold int
	LiveDataMediumTailTimeout   time.Duration
	LiveDataLargeTailThreshold  int
	LiveDataLargeTailTimeout    time.Duration
}

func DefaultServiceConfig() ServiceConfig {
	return ServiceConfig{
		RequestTimeout:              5 * time.Second,
		LiveDataMediumTailThreshold: 500,
		LiveDataMediumTailTimeout:   20 * time.Second,
		LiveDataLargeTailThreshold:  2000,
		LiveDataLargeTailTimeout:    60 * time.Second,
	}
}

func (c ServiceConfig) normalized() ServiceConfig {
	defaults := DefaultServiceConfig()
	if c.RequestTimeout <= 0 {
		c.RequestTimeout = defaults.RequestTimeout
	}
	if c.LiveDataMediumTailThreshold <= 0 {
		c.LiveDataMediumTailThreshold = defaults.LiveDataMediumTailThreshold
	}
	if c.LiveDataMediumTailTimeout <= 0 {
		c.LiveDataMediumTailTimeout = defaults.LiveDataMediumTailTimeout
	}
	if c.LiveDataLargeTailThreshold <= 0 {
		c.LiveDataLargeTailThreshold = defaults.LiveDataLargeTailThreshold
	}
	if c.LiveDataLargeTailTimeout <= 0 {
		c.LiveDataLargeTailTimeout = defaults.LiveDataLargeTailTimeout
	}
	return c
}

type Service struct {
	repo   Repository
	config ServiceConfig
}

func NewService(repo Repository) *Service {
	return NewServiceWithConfig(repo, DefaultServiceConfig())
}

func NewServiceWithConfig(repo Repository, config ServiceConfig) *Service {
	return &Service{repo: repo, config: config.normalized()}
}

func (s *Service) LoadContainerRows() ([]ContainerRow, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.config.RequestTimeout)
	defer cancel()
	return s.repo.LoadContainerRows(ctx)
}

func (s *Service) LoadSupportingResources() (Snapshot, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.config.RequestTimeout)
	defer cancel()
	return s.repo.LoadSupportingResources(ctx)
}

func (s *Service) LoadContainerMetrics(rows []ContainerRow) (map[string]ContainerMetrics, float64, uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.config.RequestTimeout)
	defer cancel()
	return s.repo.LoadContainerMetrics(ctx, rows)
}

func (s *Service) LoadContainerLiveData(containerID string, previousCPU, previousMem []float64, tail int) (ContainerLiveData, error) {
	timeout := s.liveDataTimeoutForTail(tail)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return s.repo.LoadContainerLiveData(ctx, containerID, previousCPU, previousMem, tail)
}

func (s *Service) LoadContainerLogs(containerID string, tail int) ([]string, error) {
	timeout := s.liveDataTimeoutForTail(tail)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return s.repo.LoadContainerLogs(ctx, containerID, tail)
}

func (s *Service) liveDataTimeoutForTail(tail int) time.Duration {
	if tail == 0 || tail > s.config.LiveDataLargeTailThreshold {
		return s.config.LiveDataLargeTailTimeout
	}
	if tail > s.config.LiveDataMediumTailThreshold {
		return s.config.LiveDataMediumTailTimeout
	}
	return s.config.RequestTimeout
}

func (s *Service) ExecShell(containerID string, stdin io.Reader, stdout, stderr io.Writer) error {
	return s.repo.ExecShell(context.Background(), containerID, stdin, stdout, stderr)
}

func (s *Service) LoadSnapshot() (Snapshot, error) {
	var (
		containers    []ContainerRow
		resources     Snapshot
		containersErr error
		resourcesErr  error
	)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		containers, containersErr = s.LoadContainerRows()
	}()
	go func() {
		defer wg.Done()
		resources, resourcesErr = s.LoadSupportingResources()
	}()
	wg.Wait()

	if containersErr != nil {
		return Snapshot{}, containersErr
	}

	resources.Containers = containers

	if resourcesErr == nil {
		metricsByID, totalCPU, totalMem, metricsErr := s.LoadContainerMetrics(containers)
		if metricsErr == nil {
			resources.Containers = ApplyMetricsToContainers(containers, metricsByID)
			resources.TotalCPU = totalCPU
			resources.TotalMem = totalMem
		}
		resources.ComposeProjects = AggregateComposeProjects(resources.Containers)
	} else {
		resources.ComposeProjects = AggregateComposeProjects(containers)
	}
	resources.Timestamp = time.Now()

	return resources, nil
}

func (s *Service) InspectContainer(containerID string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.config.RequestTimeout)
	defer cancel()
	return s.repo.InspectContainer(ctx, containerID)
}

func (s *Service) InspectImage(imageRef string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.config.RequestTimeout)
	defer cancel()
	return s.repo.InspectImage(ctx, imageRef)
}

func (s *Service) InspectNetwork(networkID string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.config.RequestTimeout)
	defer cancel()
	return s.repo.InspectNetwork(ctx, networkID)
}

func (s *Service) InspectVolume(volumeName string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.config.RequestTimeout)
	defer cancel()
	return s.repo.InspectVolume(ctx, volumeName)
}

func (s *Service) InspectResource(rt ResourceType, id string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.config.RequestTimeout)
	defer cancel()

	switch rt {
	case ResourceContainer:
		return s.repo.InspectContainer(ctx, id)
	case ResourceImage:
		return s.repo.InspectImage(ctx, id)
	case ResourceNetwork:
		return s.repo.InspectNetwork(ctx, id)
	case ResourceVolume:
		return s.repo.InspectVolume(ctx, id)
	default:
		return nil, nil
	}
}
