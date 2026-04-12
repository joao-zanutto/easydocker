package tables

import (
	"fmt"
	"strings"

	"easydocker/internal/core"
	"easydocker/internal/tui/browse"
	"easydocker/internal/tui/util"

	"github.com/charmbracelet/x/ansi"
)

// ContainerColumns resolves container columns for a given table width.
func ContainerColumns(tableWidth int) []ColumnDef {
	return ResolveColumns(tableWidth, ContainerSchema)
}

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

// BuildContainerSpec builds a complete containers table spec.
func BuildContainerSpec(width, cursor int, items []core.ContainerRow, includeScopeHint bool, loadingIndicator string) Spec[core.ContainerRow] {
	tableWidth := ContentWidth(width)
	columns := ContainerColumns(tableWidth)
	stateWidth := ContainerStateColumnWidth(columns)
	emptyMessage := "No containers found."
	if includeScopeHint {
		emptyMessage += " Press a to switch between running and all containers."
	}

	return Spec[core.ContainerRow]{
		EmptyMessage: emptyMessage,
		Cursor:       cursor,
		Items:        items,
		Columns:      columns,
		RowBuilder: func(container core.ContainerRow) []string {
			return ContainerTableRow(container, stateWidth, loadingIndicator)
		},
	}
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

// ContainerTableRow builds a single row for the containers table.
func ContainerTableRow(container core.ContainerRow, stateWidth int, loadingIndicator string) []string {
	state := browse.ContainerStateText(container)
	if ansi.StringWidth(state) <= stateWidth {
		state = colorStateLabel(state, container.State)
	}

	return []string{
		container.Name,
		state,
		browse.ContainerCPUValue(container, loadingIndicator),
		browse.ContainerMemoryTableValue(container, loadingIndicator),
		container.Image,
		container.Status,
	}
}

// ImageTableRow builds a single row for the images table.
func ImageTableRow(image core.ImageRow) []string {
	return []string{image.Tags, image.Size, image.Created, image.ID}
}

// NetworkTableRow builds a single row for the networks table.
func NetworkTableRow(network core.NetworkRow) []string {
	return []string{
		network.Name,
		network.Driver,
		network.Scope,
		fmt.Sprintf("%d", network.Endpoints),
		fmt.Sprintf("internal:%s attach:%s", network.Internal, network.Attachable),
	}
}

// VolumeTableRow builds a single row for the volumes table.
func VolumeTableRow(volume core.VolumeRow) []string {
	return []string{
		volume.Name,
		volume.Driver,
		volume.Scope,
		volume.Size,
		util.RefCountText(volume.RefCount),
	}
}

func colorStateLabel(label, state string) string {
	code := "37"
	switch strings.ToLower(state) {
	case "running":
		code = "32"
	case "paused", "restarting", "created":
		code = "33"
	case "exited", "stopped":
		code = "31"
	case "dead":
		code = "91"
	default:
		code = "36"
	}
	// Reset only the foreground so selected-row background remains intact.
	return "\x1b[" + code + "m" + label + "\x1b[39m"
}
