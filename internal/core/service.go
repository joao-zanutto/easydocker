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

type Service struct {
	repo    Repository
	timeout time.Duration
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo, timeout: 5 * time.Second}
}

func (s *Service) LoadContainerRows() ([]ContainerRow, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	return s.repo.LoadContainerRows(ctx)
}

func (s *Service) LoadSupportingResources() (Snapshot, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	return s.repo.LoadSupportingResources(ctx)
}

func (s *Service) LoadContainerMetrics(rows []ContainerRow) (map[string]ContainerMetrics, float64, uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	return s.repo.LoadContainerMetrics(ctx, rows)
}

func (s *Service) LoadContainerLiveData(containerID string, previousCPU, previousMem []float64, tail int) (ContainerLiveData, error) {
	timeout := s.timeout
	if tail == 0 || tail > 2000 {
		timeout = 60 * time.Second
	} else if tail > 500 {
		timeout = 20 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return s.repo.LoadContainerLiveData(ctx, containerID, previousCPU, previousMem, tail)
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
