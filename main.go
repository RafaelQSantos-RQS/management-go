package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

// DockerClient defines the interface for Docker SDK methods used by this tool.
type DockerClient interface {
	ContainerList(ctx context.Context, options container.ListOptions) ([]types.Container, error)
	ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error
	VolumesPrune(ctx context.Context, pruneFilters filters.Args) (types.VolumesPruneReport, error)
	NetworksPrune(ctx context.Context, pruneFilters filters.Args) (types.NetworksPruneReport, error)
	ImageList(ctx context.Context, options image.ListOptions) ([]image.Summary, error)
	ImageRemove(ctx context.Context, imageID string, options image.RemoveOptions) ([]image.DeleteResponse, error)
	Close() error
}

func getShortID(id string) string {
	if len(id) > 12 {
		return id[:12]
	}
	return id
}

func logInfo(msg string) {
	fmt.Printf("[%s] INFO: %s\n", time.Now().Format("2006-01-02 15:04:05"), msg)
}

func logWarn(msg string) {
	fmt.Printf("[%s] WARN: %s\n", time.Now().Format("2006-01-02 15:04:05"), msg)
}

func logError(msg string) {
	fmt.Printf("[%s] ERROR: %s\n", time.Now().Format("2006-01-02 15:04:05"), msg)
}

func logSuccess(msg string) {
	fmt.Printf("[%s] SUCCESS: %s\n", time.Now().Format("2006-01-02 15:04:05"), msg)
}

func main() {
	logInfo("Docker cleanup script started")

	// Create a new Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		logError(fmt.Sprintf("Failed to create Docker client: %v", err))
		log.Fatal(err)
	}
	defer cli.Close()

	ctx := context.Background()

	intervalStr := os.Getenv("CLEANUP_INTERVAL")
	if intervalStr == "" {
		executeCleanup(ctx, cli)
		logSuccess("Docker cleanup completed")
		return
	}

	interval, err := time.ParseDuration(intervalStr)
	if err != nil {
		logError(fmt.Sprintf("Invalid CLEANUP_INTERVAL '%s': %v. Running once instead.", intervalStr, err))
		executeCleanup(ctx, cli)
		logSuccess("Docker cleanup completed")
		return
	}

	logInfo(fmt.Sprintf("Running in daemon mode with interval: %s", interval))
	for {
		executeCleanup(ctx, cli)
		logInfo(fmt.Sprintf("Sleeping for %s...", interval))
		time.Sleep(interval)
	}
}

func executeCleanup(ctx context.Context, cli DockerClient) {
	// Execute all cleanup operations
	cleanupStoppedContainers(ctx, cli)
	cleanupUnusedVolumes(ctx, cli)
	cleanupUnusedNetworks(ctx, cli)
	cleanupUnusedImages(ctx, cli)
}

func cleanupStoppedContainers(ctx context.Context, cli DockerClient) {
	logInfo("Starting stopped containers cleanup")

	containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		logError(fmt.Sprintf("Failed to list containers: %v", err))
		return
	}

	removed := 0
	for _, cont := range containers {
		if cont.State != "running" {
			containerName := cont.Names[0][1:]
			logInfo(fmt.Sprintf("Removing stopped container: %s (ID: %s, State: %s)", containerName, getShortID(cont.ID), cont.State))

			err := cli.ContainerRemove(ctx, cont.ID, container.RemoveOptions{})
			if err != nil {
				logError(fmt.Sprintf("Failed to remove container %s: %v", getShortID(cont.ID), err))
			} else {
				removed++
			}
		}
	}

	if removed == 0 {
		logInfo("No stopped containers found")
	} else {
		logSuccess(fmt.Sprintf("Removed %d stopped containers", removed))
	}
}

func cleanupUnusedVolumes(ctx context.Context, cli DockerClient) {
	logInfo("Starting unused volumes cleanup")

	report, err := cli.VolumesPrune(ctx, filters.Args{})
	if err != nil {
		logError(fmt.Sprintf("Failed to prune volumes: %v", err))
		return
	}

	if len(report.VolumesDeleted) == 0 {
		logInfo("No unused volumes found")
	} else {
		spaceMB := float64(report.SpaceReclaimed) / (1024 * 1024)
		logSuccess(fmt.Sprintf("Removed %d volumes, reclaimed %.2f MB", len(report.VolumesDeleted), spaceMB))
	}
}

func cleanupUnusedNetworks(ctx context.Context, cli DockerClient) {
	logInfo("Starting unused networks cleanup")

	report, err := cli.NetworksPrune(ctx, filters.Args{})
	if err != nil {
		logError(fmt.Sprintf("Failed to prune networks: %v", err))
		return
	}

	if len(report.NetworksDeleted) == 0 {
		logInfo("No unused networks found")
	} else {
		logSuccess(fmt.Sprintf("Removed %d networks", len(report.NetworksDeleted)))
	}
}

func cleanupUnusedImages(ctx context.Context, cli DockerClient) {
	logInfo("Starting unused images cleanup")

	// Get all containers (including stopped ones)
	containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		logError(fmt.Sprintf("Failed to list containers: %v", err))
		return
	}

	// Build a set of images in use
	imagesInUse := make(map[string]bool)
	for _, cont := range containers {
		imagesInUse[cont.ImageID] = true
		imagesInUse[cont.Image] = true
	}

	// Get all images
	images, err := cli.ImageList(ctx, image.ListOptions{All: true})
	if err != nil {
		logError(fmt.Sprintf("Failed to list images: %v", err))
		return
	}

	removed := 0
	var totalSize int64

	// Remove images not in use
	for _, img := range images {
		if !imagesInUse[img.ID] {
			// Also check by RepoTags
			inUse := false
			for _, tag := range img.RepoTags {
				if imagesInUse[tag] {
					inUse = true
					break
				}
			}

			if !inUse {
				imageName := "<none>"
				if len(img.RepoTags) > 0 && img.RepoTags[0] != "<none>:<none>" {
					imageName = img.RepoTags[0]
				}

				sizeMB := float64(img.Size) / (1024 * 1024)
				logInfo(fmt.Sprintf("Removing unused image: %s (ID: %s, Size: %.2f MB)", imageName, getShortID(img.ID), sizeMB))

				_, err := cli.ImageRemove(ctx, img.ID, image.RemoveOptions{Force: false, PruneChildren: true})
				if err != nil {
					logError(fmt.Sprintf("Failed to remove image %s: %v", getShortID(img.ID), err))
				} else {
					removed++
					totalSize += img.Size
				}
			}
		}
	}

	if removed == 0 {
		logInfo("No unused images found")
	} else {
		spaceMB := float64(totalSize) / (1024 * 1024)
		logSuccess(fmt.Sprintf("Removed %d images, reclaimed %.2f MB", removed, spaceMB))
	}
}
