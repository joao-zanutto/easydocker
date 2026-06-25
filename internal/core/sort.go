package core

import (
	"sort"
	"strings"
	"time"
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

type byCreatedAt interface {
	CreatedAtTime() time.Time
	CreatedAtName() string
}

func (r NetworkRow) CreatedAtTime() time.Time { return r.CreatedAt }
func (r NetworkRow) CreatedAtName() string    { return r.Name }
func (r VolumeRow) CreatedAtTime() time.Time  { return r.CreatedAt }
func (r VolumeRow) CreatedAtName() string     { return r.Name }

func SortByCreatedAt[T byCreatedAt](rows []T) {
	sort.Slice(rows, func(i, j int) bool {
		lt := rows[i].CreatedAtTime()
		rt := rows[j].CreatedAtTime()
		if !lt.Equal(rt) {
			return lt.After(rt)
		}
		return strings.ToLower(rows[i].CreatedAtName()) < strings.ToLower(rows[j].CreatedAtName())
	})
}

func containerStateRank(container ContainerRow) int {
	switch {
	case container.State == StateRunning && container.Healthy:
		return 0
	case container.State == StateRunning:
		return 1
	case container.State == StateCreated:
		return 2
	case container.State == StateRestarting || container.State == StatePaused:
		return 3
	case container.State == StateExited:
		return 4
	case container.State == StateDead:
		return 5
	default:
		return 6
	}
}
