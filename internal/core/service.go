package core

import (
	"context"
	"io"
	"strings"
	"sync"
	"time"
)

//go:generate mockgen -source=service.go -destination=mock_repository_test.go -package=core Repository,ContainerLister,MetricsProvider,LogsProvider,Inspector,ShellProvider

type ContainerLister interface {
	LoadContainerRows(ctx context.Context) ([]ContainerRow, error)
	LoadSupportingResources(ctx context.Context) (Snapshot, error)
}

type MetricsProvider interface {
	LoadContainerMetrics(ctx context.Context, rows []ContainerRow) (map[string]ContainerMetrics, float64, uint64, error)
}

type LogsProvider interface {
	LoadContainerLogs(ctx context.Context, containerID string, tail int) ([]string, error)
}

type Inspector interface {
	InspectResource(ctx context.Context, resourceType ResourceType, id string) ([]string, error)
}

type ShellProvider interface {
	ExecShell(ctx context.Context, containerID string, stdin io.Reader, stdout, stderr io.Writer) error
}

type Repository interface {
	ContainerLister
	MetricsProvider
	LogsProvider
	Inspector
	ShellProvider
}

type ServiceInterface interface {
	LoadContainerRows(ctx context.Context) ([]ContainerRow, error)
	LoadSupportingResources(ctx context.Context) (Snapshot, error)
	LoadContainerMetrics(ctx context.Context, rows []ContainerRow) (map[string]ContainerMetrics, float64, uint64, error)
	LoadContainerLogs(ctx context.Context, containerID string, tail int) ([]string, error)
	LoadSnapshot(ctx context.Context) (Snapshot, error)
	InspectResource(ctx context.Context, resourceType ResourceType, id string) ([]string, error)
	ExecShell(ctx context.Context, containerID string, stdin io.Reader, stdout, stderr io.Writer) error
}

type ServiceConfig struct {
	RequestTimeout              time.Duration
	LiveDataMediumTailThreshold int
	LiveDataMediumTailTimeout   time.Duration
	LiveDataLargeTailThreshold  int
	LiveDataLargeTailTimeout    time.Duration
	ConsecutiveBackoffCap       time.Duration
}

func DefaultServiceConfig() ServiceConfig {
	return ServiceConfig{
		RequestTimeout:              5 * time.Second,
		LiveDataMediumTailThreshold: 500,
		LiveDataMediumTailTimeout:   20 * time.Second,
		LiveDataLargeTailThreshold:  2000,
		LiveDataLargeTailTimeout:    60 * time.Second,
		ConsecutiveBackoffCap:       30 * time.Second,
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
	if c.ConsecutiveBackoffCap <= 0 {
		c.ConsecutiveBackoffCap = defaults.ConsecutiveBackoffCap
	}
	return c
}

type Service struct {
	repo                Repository
	config              ServiceConfig
	consecutiveFailures int
	mu                  sync.Mutex
}

func NewService(repo Repository) *Service {
	return NewServiceWithConfig(repo, DefaultServiceConfig())
}

func NewServiceWithConfig(repo Repository, config ServiceConfig) *Service {
	return &Service{repo: repo, config: config.normalized()}
}

func (s *Service) LoadContainerRows(ctx context.Context) ([]ContainerRow, error) {
	ctx, cancel := context.WithTimeout(ctx, s.config.RequestTimeout)
	defer cancel()
	return s.repo.LoadContainerRows(ctx)
}

func (s *Service) LoadSupportingResources(ctx context.Context) (Snapshot, error) {
	ctx, cancel := context.WithTimeout(ctx, s.config.RequestTimeout)
	defer cancel()
	return s.repo.LoadSupportingResources(ctx)
}

func (s *Service) LoadContainerMetrics(ctx context.Context, rows []ContainerRow) (map[string]ContainerMetrics, float64, uint64, error) {
	ctx, cancel := context.WithTimeout(ctx, s.config.RequestTimeout)
	defer cancel()
	return s.repo.LoadContainerMetrics(ctx, rows)
}

func (s *Service) LoadContainerLogs(ctx context.Context, containerID string, tail int) ([]string, error) {
	timeout := s.liveDataTimeoutForTail(tail)
	ctx, cancel := context.WithTimeout(ctx, timeout)
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

func (s *Service) ExecShell(ctx context.Context, containerID string, stdin io.Reader, stdout, stderr io.Writer) error {
	return s.repo.ExecShell(ctx, containerID, stdin, stdout, stderr)
}

func (s *Service) LoadSnapshot(ctx context.Context) (Snapshot, error) {
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
		containers, containersErr = s.LoadContainerRows(ctx)
	}()
	go func() {
		defer wg.Done()
		resources, resourcesErr = s.LoadSupportingResources(ctx)
	}()
	wg.Wait()

	if containersErr != nil {
		s.mu.Lock()
		s.consecutiveFailures++
		backoff := time.Duration(min(1<<s.consecutiveFailures, int(s.config.ConsecutiveBackoffCap.Seconds()))) * time.Second
		s.mu.Unlock()
		time.Sleep(backoff)
		return Snapshot{}, containersErr
	}

	s.mu.Lock()
	s.consecutiveFailures = 0
	s.mu.Unlock()

	resources.Containers = containers

	resources.Images = ApplyImageContainerCounts(resources.Images, containers)
	resources.Networks = ApplyNetworkEndpointCounts(resources.Networks, containers)

	if resourcesErr == nil {
		metricsByID, totalCPU, totalMem, metricsErr := s.LoadContainerMetrics(ctx, containers)
		if metricsErr == nil {
			resources.Containers = ApplyMetricsToContainers(containers, metricsByID)
			resources.TotalCPU = totalCPU
			resources.TotalMem = totalMem
		}
		resources.ComposeProjects = AggregateComposeProjects(resources.Containers)
	} else {
		resources.ComposeProjects = AggregateComposeProjects(containers)
	}
	return resources, nil
}

func (s *Service) InspectResource(ctx context.Context, rt ResourceType, id string) ([]string, error) {
	return s.repo.InspectResource(ctx, rt, id)
}

func ApplyImageContainerCounts(images []ImageRow, containers []ContainerRow) []ImageRow {
	counts := make(map[string]int)
	for _, c := range containers {
		imageID := strings.TrimPrefix(c.ImageID, "sha256:")
		if len(imageID) >= 12 {
			counts[imageID[:12]]++
		}
	}
	for i := range images {
		images[i].Containers = int64(counts[images[i].ID])
	}
	return images
}

func ApplyNetworkEndpointCounts(networks []NetworkRow, containers []ContainerRow) []NetworkRow {
	counts := make(map[string]int)
	for _, c := range containers {
		for _, net := range c.Networks {
			counts[net]++
		}
	}
	for i := range networks {
		networks[i].Endpoints = counts[networks[i].Name]
	}
	return networks
}
