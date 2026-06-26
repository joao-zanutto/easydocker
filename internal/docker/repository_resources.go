package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/system"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
)

func (r *Repository) loadSupportingResourcesData(ctx context.Context, cli *client.Client) ([]image.Summary, []network.Inspect, []*volume.Volume, system.Info, error) {
	images, err := cli.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return nil, nil, nil, system.Info{}, fmt.Errorf("repository.list images: %w", err)
	}

	networks, err := cli.NetworkList(ctx, network.ListOptions{})
	if err != nil {
		return nil, nil, nil, system.Info{}, fmt.Errorf("repository.list networks: %w", err)
	}

	du, err := cli.DiskUsage(ctx, types.DiskUsageOptions{Types: []types.DiskUsageObject{types.VolumeObject}})
	if err != nil {
		return nil, nil, nil, system.Info{}, fmt.Errorf("repository.list volumes: %w", err)
	}

	info, err := cli.Info(ctx)
	if err != nil {
		info = system.Info{}
	}

	return images, networks, du.Volumes, info, nil
}
