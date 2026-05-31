package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"easydocker/internal/core"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

type Repository struct {
	clientMu  sync.Mutex
	client    *client.Client
	clientErr error
	now       func() time.Time
}

func NewRepository() *Repository {
	return &Repository{now: time.Now}
}

func (r *Repository) LoadContainerRows(ctx context.Context) ([]core.ContainerRow, error) {
	return withClientResult(r, func(cli *client.Client) ([]core.ContainerRow, error) {
		containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
		if err != nil {
			return nil, wrapDockerError("list containers", err)
		}

		rows := make([]core.ContainerRow, 0, len(containers))
		for _, item := range containers {
			rows = append(rows, mapContainerRow(item))
		}
		core.SortContainers(rows)
		return rows, nil
	})
}

func (r *Repository) LoadSupportingResources(ctx context.Context) (core.Snapshot, error) {
	return withClientResult(r, func(cli *client.Client) (core.Snapshot, error) {
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
		if info.MemTotal > 0 {
			snapshot.TotalLimit = uint64(info.MemTotal)
		}

		return snapshot, nil
	})
}

func (r *Repository) InspectResource(ctx context.Context, resourceType core.ResourceType, id string) ([]string, error) {
	return withClientResult(r, func(cli *client.Client) ([]string, error) {
		var result any
		var err error
		switch resourceType {
		case core.ResourceContainer:
			result, _, err = cli.ContainerInspectWithRaw(ctx, id, false)
		case core.ResourceImage:
			result, _, err = cli.ImageInspectWithRaw(ctx, id)
		case core.ResourceNetwork:
			result, err = cli.NetworkInspect(ctx, id, network.InspectOptions{})
		case core.ResourceVolume:
			result, err = cli.VolumeInspect(ctx, id)
		default:
			return nil, nil
		}
		if err != nil {
			return nil, wrapDockerError("inspect resource", err)
		}
		return toInspectResult(result)
	})
}

func toInspectResult(data any) ([]string, error) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("repository.marshal json: %w", err)
	}
	lines := strings.Split(string(jsonData), "\n")
	return lines, nil
}

func (r *Repository) dockerClient() (*client.Client, error) {
	r.clientMu.Lock()
	defer r.clientMu.Unlock()
	if r.client != nil {
		return r.client, nil
	}
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		r.clientErr = err
		return nil, err
	}
	r.client = cli
	return cli, nil
}

func withClientResult[T any](r *Repository, fn func(*client.Client) (T, error)) (T, error) {
	var zero T
	cli, err := r.dockerClient()
	if err != nil {
		return zero, err
	}
	return fn(cli)
}

func wrapDockerError(operation string, err error) error {
	return fmt.Errorf("repository.%s: %w", operation, err)
}
