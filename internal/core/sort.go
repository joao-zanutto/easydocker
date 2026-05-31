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
		leftRepository, leftTag := parseImageSortKey(left.Tags)
		rightRepository, rightTag := parseImageSortKey(right.Tags)
		if leftRepository != rightRepository {
			return leftRepository < rightRepository
		}
		if leftTag != rightTag {
			return leftTag < rightTag
		}
		return strings.ToLower(left.Tags) < strings.ToLower(right.Tags)
	})
}

func parseImageSortKey(tags string) (string, string) {
	primaryTag := strings.TrimSpace(tags)
	if comma := strings.Index(primaryTag, ","); comma >= 0 {
		primaryTag = strings.TrimSpace(primaryTag[:comma])
	}
	separator := strings.LastIndex(primaryTag, ":")
	if separator <= 0 {
		return strings.ToLower(primaryTag), ""
	}
	repository := strings.ToLower(primaryTag[:separator])
	tag := strings.ToLower(primaryTag[separator+1:])
	return repository, tag
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
		if !left.CreatedAt.Equal(right.CreatedAt) {
			return left.CreatedAt.After(right.CreatedAt)
		}
		return strings.ToLower(left.Name) < strings.ToLower(right.Name)
	})
}

func containerStateRank(container ContainerRow) int {
	switch container.State {
	case StateRunning:
		if container.Healthy {
			return 0
		}
		return 1
	case StateCreated:
		return 2
	case StateRestarting, StatePaused:
		return 3
	case StateExited:
		return 4
	case StateDead:
		return 5
	default:
		return 6
	}
}
