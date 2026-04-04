package core

import (
	"sort"
	"strings"
)

func SortContainers(rows []ContainerRow) {
	sort.Slice(rows, func(i, j int) bool {
		left := rows[i]
		right := rows[j]
		leftRank := containerStateRank(left)
		rightRank := containerStateRank(right)
		if leftRank != rightRank {
			return leftRank < rightRank
		}
		if left.CreatedUnix != right.CreatedUnix {
			return left.CreatedUnix > right.CreatedUnix
		}
		return strings.ToLower(left.Name) < strings.ToLower(right.Name)
	})
}

func SortImages(rows []ImageRow) {
	sort.Slice(rows, func(i, j int) bool {
		left := rows[i]
		right := rows[j]
		if left.CreatedUnix != right.CreatedUnix {
			return left.CreatedUnix > right.CreatedUnix
		}
		return strings.ToLower(left.Tags) < strings.ToLower(right.Tags)
	})
}

func SortNetworks(rows []NetworkRow) {
	sort.Slice(rows, func(i, j int) bool {
		left := rows[i]
		right := rows[j]
		if !left.CreatedAt.Equal(right.CreatedAt) {
			return left.CreatedAt.After(right.CreatedAt)
		}
		return strings.ToLower(left.Name) < strings.ToLower(right.Name)
	})
}

func SortVolumes(rows []VolumeRow) {
	sort.Slice(rows, func(i, j int) bool {
		left := rows[i]
		right := rows[j]
		if left.CreatedAt != right.CreatedAt {
			return left.CreatedAt > right.CreatedAt
		}
		return strings.ToLower(left.Name) < strings.ToLower(right.Name)
	})
}

func containerStateRank(container ContainerRow) int {
	state := strings.ToLower(container.State)
	switch {
	case state == "running" && container.Healthy:
		return 0
	case state == "running":
		return 1
	case state == "created":
		return 2
	case state == "restarting" || state == "paused":
		return 3
	case state == "exited" || state == "stopped":
		return 4
	case state == "dead":
		return 5
	default:
		return 6
	}
}
