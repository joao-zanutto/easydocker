package docker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"easydocker/internal/core"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type Repository struct {
	clientOnce sync.Once
	client     *client.Client
	clientErr  error
	now        func() time.Time
}

func NewRepository() *Repository {
	return &Repository{now: time.Now}
}

func (r *Repository) LoadContainerRows(ctx context.Context) ([]core.ContainerRow, error) {
	cli, err := r.dockerClient()
	if err != nil {
		return nil, err
	}

	containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("list containers: %w", err)
	}

	rows := make([]core.ContainerRow, 0, len(containers))
	for _, item := range containers {
		rows = append(rows, mapContainerRow(item))
	}
	core.SortContainers(rows)

	return rows, nil
}

func (r *Repository) LoadSupportingResources(ctx context.Context) (core.Snapshot, error) {
	cli, err := r.dockerClient()
	if err != nil {
		return core.Snapshot{}, err
	}

	images, networks, volumes, info, err := r.loadSupportingResourcesData(ctx, cli)
	if err != nil {
		return core.Snapshot{}, err
	}

	snapshot := core.Snapshot{
		Images:    make([]core.ImageRow, 0, len(images)),
		Networks:  make([]core.NetworkRow, 0, len(networks)),
		Volumes:   make([]core.VolumeRow, 0, len(volumes.Volumes)),
		Timestamp: r.now(),
	}

	for _, item := range images {
		snapshot.Images = append(snapshot.Images, mapImageRow(item))
	}
	core.SortImages(snapshot.Images)

	for _, item := range networks {
		snapshot.Networks = append(snapshot.Networks, mapNetworkRow(item))
	}
	core.SortNetworks(snapshot.Networks)

	for _, item := range volumes.Volumes {
		snapshot.Volumes = append(snapshot.Volumes, mapVolumeRow(item))
	}
	core.SortVolumes(snapshot.Volumes)

	snapshot.TotalCPU = 0
	snapshot.TotalMem = 0
	snapshot.TotalLimit = uint64(info.MemTotal)

	return snapshot, nil
}

func (r *Repository) dockerClient() (*client.Client, error) {
	r.clientOnce.Do(func() {
		r.client, r.clientErr = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	})
	return r.client, r.clientErr
}
