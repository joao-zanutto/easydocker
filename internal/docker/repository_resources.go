package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/system"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
)

type imagesResult struct {
	items []image.Summary
	err   error
}

type networksResult struct {
	items []network.Inspect
	err   error
}

type volumesResult struct {
	items volume.ListResponse
	err   error
}

type infoResult struct {
	item system.Info
	err  error
}

func (r *Repository) loadSupportingResourcesData(ctx context.Context, cli *client.Client) ([]image.Summary, []network.Inspect, volume.ListResponse, system.Info, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	imagesCh := make(chan imagesResult, 1)
	networksCh := make(chan networksResult, 1)
	volumesCh := make(chan volumesResult, 1)
	infoCh := make(chan infoResult, 1)

	go func() {
		items, err := cli.ImageList(ctx, image.ListOptions{})
		imagesCh <- imagesResult{items: items, err: err}
	}()
	go func() {
		items, err := cli.NetworkList(ctx, network.ListOptions{})
		networksCh <- networksResult{items: items, err: err}
	}()
	go func() {
		items, err := cli.VolumeList(ctx, volume.ListOptions{})
		volumesCh <- volumesResult{items: items, err: err}
	}()
	go func() {
		item, err := cli.Info(ctx)
		infoCh <- infoResult{item: item, err: err}
	}()

	imagesRes := <-imagesCh
	if imagesRes.err != nil {
		cancel()
		return nil, nil, volume.ListResponse{}, system.Info{}, fmt.Errorf("list images: %w", imagesRes.err)
	}

	networksRes := <-networksCh
	if networksRes.err != nil {
		cancel()
		return nil, nil, volume.ListResponse{}, system.Info{}, fmt.Errorf("list networks: %w", networksRes.err)
	}

	volumesRes := <-volumesCh
	if volumesRes.err != nil {
		cancel()
		return nil, nil, volume.ListResponse{}, system.Info{}, fmt.Errorf("list volumes: %w", volumesRes.err)
	}

	infoRes := <-infoCh
	info := infoRes.item
	if infoRes.err != nil {
		info = system.Info{}
	}

	return imagesRes.items, networksRes.items, volumesRes.items, info, nil
}
