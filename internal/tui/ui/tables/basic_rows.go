package tables

import (
	"fmt"
	"strings"

	"easydocker/internal/core"
)

// ImageColumns resolves image columns for a given table width.
func ImageColumns(tableWidth int) []ColumnDef {
	return ResolveColumns(tableWidth, ImageSchema)
}

// NetworkColumns resolves network columns for a given table width.
func NetworkColumns(tableWidth int) []ColumnDef {
	return ResolveColumns(tableWidth, NetworkSchema)
}

// VolumeColumns resolves volume columns for a given table width.
func VolumeColumns(tableWidth int) []ColumnDef {
	return ResolveColumns(tableWidth, VolumeSchema)
}

// BuildImageSpec builds a complete images table spec.
func BuildImageSpec(width, cursor int, items []core.ImageRow) Spec[core.ImageRow] {
	return SimpleSpec(width, "No images found.", cursor, items, ImageColumns, ImageTableRow)
}

// BuildNetworkSpec builds a complete networks table spec.
func BuildNetworkSpec(width, cursor int, items []core.NetworkRow) Spec[core.NetworkRow] {
	return SimpleSpec(width, "No networks found.", cursor, items, NetworkColumns, NetworkTableRow)
}

// BuildVolumeSpec builds a complete volumes table spec.
func BuildVolumeSpec(width, cursor int, items []core.VolumeRow) Spec[core.VolumeRow] {
	return SimpleSpec(width, "No volumes found.", cursor, items, VolumeColumns, VolumeTableRow)
}

// ImageTableRow builds a single row for the images table.
func ImageTableRow(image core.ImageRow) []string {
	repository, tags := splitImageTags(image.Tags)
	return []string{repository, tags, image.Size, image.Created, image.ID}
}

func splitImageTags(formattedTags string) (string, string) {
	if formattedTags == "" {
		return "-", "-"
	}

	repositories := make([]string, 0, 4)
	tags := make([]string, 0, 4)
	for _, reference := range strings.Split(formattedTags, ", ") {
		repository, tag := splitImageTagReference(reference)
		repositories = append(repositories, repository)
		tags = append(tags, tag)
	}

	return strings.Join(repositories, ", "), strings.Join(tags, ", ")
}

func splitImageTagReference(reference string) (string, string) {
	if reference == "<none>:<none>" {
		return "<none>", "<none>"
	}

	lastColon := strings.LastIndex(reference, ":")
	if lastColon <= 0 || lastColon == len(reference)-1 {
		return reference, "-"
	}

	return reference[:lastColon], reference[lastColon+1:]
}

// NetworkTableRow builds a single row for the networks table.
func NetworkTableRow(network core.NetworkRow) []string {
	return []string{
		network.Name,
		network.Driver,
		fmt.Sprintf("%d", network.Endpoints),
		network.Created,
	}
}

// VolumeTableRow builds a single row for the volumes table.
func VolumeTableRow(volume core.VolumeRow) []string {
	return []string{
		volume.Name,
		volume.Mountpoint,
		volume.Size,
		volume.Created,
	}
}
