package core

import (
	"fmt"
	"strings"
	"time"
)

func AggregateComposeProjects(containers []ContainerRow) []ComposeProject {
	projects := make(map[string]*ComposeProject)
	order := make([]string, 0)
	memoryPercentSums := make(map[string]float64)
	memoryPercentCounts := make(map[string]int)

	for _, container := range containers {
		projectName := strings.TrimSpace(container.ComposeProject)
		if projectName == "" {
			continue
		}
		project, ok := projects[projectName]
		if !ok {
			project = &ComposeProject{Name: projectName, CreatedUnix: container.CreatedUnix}
			projects[projectName] = project
			order = append(order, projectName)
		}
		project.Containers = append(project.Containers, container)
		project.ContainerCount++
		if strings.EqualFold(container.State, "running") {
			project.RunningCount++
		}
		if container.Healthy {
			project.HealthyCount++
		}
		if container.CreatedUnix > project.CreatedUnix {
			project.CreatedUnix = container.CreatedUnix
		}
		if project.WorkingDir == "" {
			project.WorkingDir = strings.TrimSpace(container.ComposeWorkingDir)
		}
		if project.ConfigFiles == "" {
			project.ConfigFiles = strings.TrimSpace(container.ComposeConfigFiles)
		}
		if project.Network == "" {
			project.Network = deriveComposeNetwork(projectName)
		}
		if service := strings.TrimSpace(container.ComposeService); service != "" {
			if !containsString(project.Services, service) {
				project.Services = append(project.Services, service)
			}
		}
		project.CPUPercent += maxFloat(container.CPUPercent, 0)
		if container.MemoryUsageBytes > 0 {
			project.MemoryUsageBytes += container.MemoryUsageBytes
		}
		if container.MemoryLimitBytes > 0 {
			project.MemoryLimitBytes = container.MemoryLimitBytes
		}
		if hasComposeMemoryPercent(container) {
			memoryPercentSums[projectName] += container.MemoryPercent
			memoryPercentCounts[projectName]++
		}
	}

	for _, name := range order {
		project := projects[name]
		SortContainers(project.Containers)
		if project.CreatedUnix > 0 {
			project.Created = HumanAge(time.Unix(project.CreatedUnix, 0))
		} else {
			project.Created = "-"
		}
		if project.MemoryUsageBytes > 0 {
			project.MemoryUsage = HumanBytes(int64(project.MemoryUsageBytes))
		} else {
			project.MemoryUsage = "-"
		}
		if project.MemoryLimitBytes > 0 {
			project.MemoryLimit = HumanBytes(int64(project.MemoryLimitBytes))
		} else {
			project.MemoryLimit = "-"
		}
		if memoryPercentCounts[name] > 0 {
			project.MemoryPercent = memoryPercentSums[name]
		} else {
			project.MemoryPercent = 0
		}
	}

	projectsOut := make([]ComposeProject, 0, len(order))
	for _, name := range order {
		projectsOut = append(projectsOut, *projects[name])
	}
	return projectsOut
}

func HumanAge(then time.Time) string {
	delta := time.Since(then)
	switch {
	case delta < time.Minute:
		return "just now"
	case delta < time.Hour:
		return fmtDurationMinutes(delta)
	case delta < 24*time.Hour:
		return fmtDurationHours(delta)
	case delta < 30*24*time.Hour:
		return fmtDurationDays(delta)
	case delta < 365*24*time.Hour:
		return fmtDurationMonths(delta)
	default:
		return fmtDurationYears(delta)
	}
}

func deriveComposeNetwork(projectName string) string {
	if projectName == "" {
		return "-"
	}
	return projectName + "_default"
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func maxFloat(value, floor float64) float64 {
	if value > floor {
		return value
	}
	return floor
}

func hasComposeMemoryPercent(container ContainerRow) bool {
	if container.MemoryPercent <= 0 {
		return false
	}
	if container.MemoryUsageBytes > 0 {
		return true
	}
	memoryUsage := strings.TrimSpace(strings.ToLower(container.MemoryUsage))
	return memoryUsage != "" && memoryUsage != "-" && memoryUsage != "loading"
}

func fmtDurationMinutes(delta time.Duration) string {
	return fmt.Sprintf("%dm ago", int(delta.Minutes()))
}

func fmtDurationHours(delta time.Duration) string {
	return fmt.Sprintf("%dh ago", int(delta.Hours()))
}

func fmtDurationDays(delta time.Duration) string {
	return fmt.Sprintf("%dd ago", int(delta.Hours()/24))
}

func fmtDurationMonths(delta time.Duration) string {
	return fmt.Sprintf("%dmo ago", int(delta.Hours()/(24*30)))
}

func fmtDurationYears(delta time.Duration) string {
	return fmt.Sprintf("%dy ago", int(delta.Hours()/(24*365)))
}
