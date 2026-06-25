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

type result[T any] struct {
	value T
	err   error
}

func (r *Repository) loadSupportingResourcesData(ctx context.Context, cli *client.Client) ([]image.Summary, []network.Inspect, []*volume.Volume, system.Info, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	imagesCh := make(chan result[[]image.Summary], 1)
	networksCh := make(chan result[[]network.Inspect], 1)
	volumesCh := make(chan result[[]*volume.Volume], 1)
	infoCh := make(chan result[system.Info], 1)

	go func() {
		items, err := cli.ImageList(ctx, image.ListOptions{})
		imagesCh <- result[[]image.Summary]{value: items, err: err}
	}()
	go func() {
		items, err := cli.NetworkList(ctx, network.ListOptions{})
		networksCh <- result[[]network.Inspect]{value: items, err: err}
	}()
	go func() {
		du, err := cli.DiskUsage(ctx, types.DiskUsageOptions{Types: []types.DiskUsageObject{types.VolumeObject}})
		if err != nil {
			volumesCh <- result[[]*volume.Volume]{value: nil, err: err}
			return
		}
		volumesCh <- result[[]*volume.Volume]{value: du.Volumes, err: nil}
	}()
	go func() {
		item, err := cli.Info(ctx)
		infoCh <- result[system.Info]{value: item, err: err}
	}()

	imagesRes := <-imagesCh
	if imagesRes.err != nil {
		cancel()
		return nil, nil, nil, system.Info{}, fmt.Errorf("repository.list images: %w", imagesRes.err)
	}

	networksRes := <-networksCh
	if networksRes.err != nil {
		cancel()
		return nil, nil, nil, system.Info{}, fmt.Errorf("repository.list networks: %w", networksRes.err)
	}

	volumesRes := <-volumesCh
	if volumesRes.err != nil {
		cancel()
		return nil, nil, nil, system.Info{}, fmt.Errorf("repository.list volumes: %w", volumesRes.err)
	}

	infoRes := <-infoCh
	info := infoRes.value
	if infoRes.err != nil {
		info = system.Info{}
	}

	return imagesRes.value, networksRes.value, volumesRes.value, info, nil
}
