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
	clientOnce sync.Once
	client     *client.Client
	clientErr  error
	now        func() time.Time
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

func (r *Repository) InspectContainer(ctx context.Context, containerID string) ([]string, error) {
	return r.inspectResource(ctx, "container", func(ctx context.Context, id string, cli *client.Client) (any, error) {
		return cli.ContainerInspect(ctx, id)
	}, containerID)
}

func (r *Repository) InspectImage(ctx context.Context, imageRef string) ([]string, error) {
	return r.inspectResource(ctx, "image", func(ctx context.Context, id string, cli *client.Client) (any, error) {
		data, _, err := cli.ImageInspectWithRaw(ctx, id)
		return data, err
	}, imageRef)
}

func (r *Repository) InspectNetwork(ctx context.Context, networkID string) ([]string, error) {
	return r.inspectResource(ctx, "network", func(ctx context.Context, id string, cli *client.Client) (any, error) {
		return cli.NetworkInspect(ctx, id, network.InspectOptions{})
	}, networkID)
}

func (r *Repository) InspectVolume(ctx context.Context, volumeName string) ([]string, error) {
	return r.inspectResource(ctx, "volume", func(ctx context.Context, id string, cli *client.Client) (any, error) {
		return cli.VolumeInspect(ctx, id)
	}, volumeName)
}

func (r *Repository) inspectResource(ctx context.Context, resourceType string, inspectFn func(context.Context, string, *client.Client) (any, error), id string) ([]string, error) {
	return withClientResult(r, func(cli *client.Client) ([]string, error) {
		result, err := inspectFn(ctx, id, cli)
		if err != nil {
			return nil, wrapDockerError("inspect "+resourceType, err)
		}
		return toInspectResult(result)
	})
}

func toInspectResult(data any) ([]string, error) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal json: %w", err)
	}
	lines := strings.Split(string(jsonData), "\n")
	return lines, nil
}

func (r *Repository) dockerClient() (*client.Client, error) {
	r.clientOnce.Do(func() {
		r.client, r.clientErr = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	})
	return r.client, r.clientErr
}

func withClientResult[T any](r *Repository, fn func(*client.Client) (T, error)) (T, error) {
	var zero T
	cli, err := r.dockerClient()
	if err != nil {
		return zero, err
	}
	return fn(cli)
}

func wrapDockerError(prefix string, err error) error {
	return fmt.Errorf("%s: %w", prefix, err)
}
