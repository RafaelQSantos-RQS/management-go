package main

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
)

// MockDockerClient implements the DockerClient interface for testing.
type MockDockerClient struct {
	ContainerListFunc   func(ctx context.Context, options container.ListOptions) ([]types.Container, error)
	ContainerRemoveFunc func(ctx context.Context, containerID string, options container.RemoveOptions) error
	VolumesPruneFunc    func(ctx context.Context, pruneFilters filters.Args) (types.VolumesPruneReport, error)
	NetworksPruneFunc   func(ctx context.Context, pruneFilters filters.Args) (types.NetworksPruneReport, error)
	ImageListFunc       func(ctx context.Context, options image.ListOptions) ([]image.Summary, error)
	ImageRemoveFunc     func(ctx context.Context, imageID string, options image.RemoveOptions) ([]image.DeleteResponse, error)
	CloseFunc           func() error
}

func (m *MockDockerClient) ContainerList(ctx context.Context, options container.ListOptions) ([]types.Container, error) {
	return m.ContainerListFunc(ctx, options)
}

func (m *MockDockerClient) ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error {
	return m.ContainerRemoveFunc(ctx, containerID, options)
}

func (m *MockDockerClient) VolumesPrune(ctx context.Context, pruneFilters filters.Args) (types.VolumesPruneReport, error) {
	return m.VolumesPruneFunc(ctx, pruneFilters)
}

func (m *MockDockerClient) NetworksPrune(ctx context.Context, pruneFilters filters.Args) (types.NetworksPruneReport, error) {
	return m.NetworksPruneFunc(ctx, pruneFilters)
}

func (m *MockDockerClient) ImageList(ctx context.Context, options image.ListOptions) ([]image.Summary, error) {
	return m.ImageListFunc(ctx, options)
}

func (m *MockDockerClient) ImageRemove(ctx context.Context, imageID string, options image.RemoveOptions) ([]image.DeleteResponse, error) {
	return m.ImageRemoveFunc(ctx, imageID, options)
}

func (m *MockDockerClient) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

func TestCleanupStoppedContainers(t *testing.T) {
	removedCount := 0
	mock := &MockDockerClient{
		ContainerListFunc: func(ctx context.Context, options container.ListOptions) ([]types.Container, error) {
			return []types.Container{
				{ID: "run1", State: "running", Names: []string{"/running-container"}},
				{ID: "stop1", State: "exited", Names: []string{"/stopped-container-1"}},
				{ID: "stop2", State: "created", Names: []string{"/stopped-container-2"}},
			}, nil
		},
		ContainerRemoveFunc: func(ctx context.Context, containerID string, options container.RemoveOptions) error {
			removedCount++
			return nil
		},
	}

	cleanupStoppedContainers(context.Background(), mock)

	if removedCount != 2 {
		t.Errorf("Expected 2 containers to be removed, got %d", removedCount)
	}
}

func TestCleanupUnusedVolumes(t *testing.T) {
	pruned := false
	mock := &MockDockerClient{
		VolumesPruneFunc: func(ctx context.Context, pruneFilters filters.Args) (types.VolumesPruneReport, error) {
			pruned = true
			return types.VolumesPruneReport{
				VolumesDeleted: []string{"vol1", "vol2"},
				SpaceReclaimed: 1024 * 1024 * 10, // 10MB
			}, nil
		},
	}

	cleanupUnusedVolumes(context.Background(), mock)

	if !pruned {
		t.Error("VolumesPrune was not called")
	}
}

func TestCleanupUnusedNetworks(t *testing.T) {
	pruned := false
	mock := &MockDockerClient{
		NetworksPruneFunc: func(ctx context.Context, pruneFilters filters.Args) (types.NetworksPruneReport, error) {
			pruned = true
			return types.NetworksPruneReport{
				NetworksDeleted: []string{"net1"},
			}, nil
		},
	}

	cleanupUnusedNetworks(context.Background(), mock)

	if !pruned {
		t.Error("NetworksPrune was not called")
	}
}

func TestCleanupUnusedImages(t *testing.T) {
	removedCount := 0
	mock := &MockDockerClient{
		ContainerListFunc: func(ctx context.Context, options container.ListOptions) ([]types.Container, error) {
			// One container using 'img1'
			return []types.Container{
				{ImageID: "img1", Image: "image-in-use"},
			}, nil
		},
		ImageListFunc: func(ctx context.Context, options image.ListOptions) ([]image.Summary, error) {
			return []image.Summary{
				{ID: "img1", RepoTags: []string{"image-in-use:latest"}, Size: 100},
				{ID: "img2", RepoTags: []string{"unused-image:latest"}, Size: 200},
				{ID: "img3", RepoTags: []string{"<none>:<none>"}, Size: 50},
			}, nil
		},
		ImageRemoveFunc: func(ctx context.Context, imageID string, options image.RemoveOptions) ([]image.DeleteResponse, error) {
			removedCount++
			return nil, nil
		},
	}

	cleanupUnusedImages(context.Background(), mock)

	// img2 and img3 should be removed
	if removedCount != 2 {
		t.Errorf("Expected 2 images to be removed, got %d", removedCount)
	}
}
