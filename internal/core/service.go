package core

import (
	"context"
	"time"
)

type Repository interface {
	LoadContainerRows(ctx context.Context) ([]ContainerRow, error)
	LoadSupportingResources(ctx context.Context) (Snapshot, error)
	LoadContainerMetrics(ctx context.Context, rows []ContainerRow) (map[string]ContainerMetrics, float64, uint64, error)
	LoadContainerLiveData(ctx context.Context, containerID string, previousCPU, previousMem []float64, tail int) (ContainerLiveData, error)
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

func (s *Service) liveDataTimeoutForTail(tail int) time.Duration {
	if tail == 0 || tail > s.config.LiveDataLargeTailThreshold {
		return s.config.LiveDataLargeTailTimeout
	}
	if tail > s.config.LiveDataMediumTailThreshold {
		return s.config.LiveDataMediumTailTimeout
	}
	return s.config.RequestTimeout
}

func (s *Service) LoadSnapshot() (Snapshot, error) {
	containers, err := s.LoadContainerRows()
	if err != nil {
		return Snapshot{}, err
	}

	resources, err := s.LoadSupportingResources()
	if err != nil {
		return Snapshot{}, err
	}

	metricsByID, totalCPU, totalMem, err := s.LoadContainerMetrics(containers)
	if err != nil {
		return Snapshot{}, err
	}

	resources.Containers = ApplyMetricsToContainers(containers, metricsByID)
	resources.TotalCPU = totalCPU
	resources.TotalMem = totalMem
	resources.Timestamp = time.Now()

	return resources, nil
}
