package tyk

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TestContainers struct {
	Containers []testcontainers.Container
	testcontainers.Network
}

func (t TestContainers) CleanUp(ctx context.Context) error {
	for _, c := range t.Containers {
		if err := c.Terminate(ctx); err != nil {
			return err
		}
	}
	if err := t.Network.Remove(ctx); err != nil {
		return err
	}
	return nil
}

type TykContainers struct {
	*TestContainers
	URI string
}

func NewTykContainers(ctx context.Context) (*TykContainers, error) {

	networkName := uuid.New().String()
	network, err := testcontainers.GenericNetwork(ctx, testcontainers.GenericNetworkRequest{
		NetworkRequest: testcontainers.NetworkRequest{
			Name:           networkName,
			CheckDuplicate: true,
		},
	})
	if err != nil {
		return nil, err
	}

	redis, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:      "redis:6.2.7-alpine",
			WaitingFor: wait.ForExec([]string{"redis-cli", "ping"}).WithStartupTimeout(10 * time.Second),
			Networks: []string{
				networkName,
			},
			NetworkAliases: map[string][]string{
				networkName: {"tyk-redis"},
			},
		},
		Started: true,
	})
	if err != nil {
		return nil, err
	}

	httpbin, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "kennethreitz/httpbin",
			ExposedPorts: []string{"80/tcp"},
			WaitingFor:   wait.ForHTTP("/").WithStartupTimeout(10 * time.Second),
			Networks: []string{
				networkName,
			},
			NetworkAliases: map[string][]string{
				networkName: {"httpbin"},
			},
		},
		Started: true,
	})
	if err != nil {
		return nil, err
	}

	tyk, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "docker.tyk.io/tyk-gateway/tyk-gateway:v5.1.0",
			ExposedPorts: []string{"8080/tcp"},
			WaitingFor: wait.ForAll(
				wait.ForHTTP("/hello").WithStartupTimeout(10*time.Second),
				wait.ForLog("Initialised API Definitions")),
			Networks: []string{
				networkName,
			},
			NetworkAliases: map[string][]string{
				networkName: {"tyk-gateway"},
			},
			Files: []testcontainers.ContainerFile{
				{
					HostFilePath:      "../../apps",
					ContainerFilePath: "/opt/tyk-gateway/apps",
					FileMode:          0o700,
				},
				{
					HostFilePath:      "../../middleware",
					ContainerFilePath: "/opt/tyk-gateway/middleware",
					FileMode:          0o700,
				},
				{
					HostFilePath:      "../../tyk.standalone.conf",
					ContainerFilePath: "/opt/tyk-gateway/tyk.conf",
					FileMode:          0o700,
				},
			},
			Env: map[string]string{
				"TYK_GW_SECRET": "foo",
			},
		},
		Started: true,
	})
	if err != nil {
		return nil, err
	}

	ip, err := tyk.Host(ctx)
	if err != nil {
		return nil, err
	}

	mappedPort, err := tyk.MappedPort(ctx, "8080")
	if err != nil {
		return nil, err
	}

	uri := fmt.Sprintf("http://%s:%s", ip, mappedPort.Port())

	return &TykContainers{
		TestContainers: &TestContainers{
			Containers: []testcontainers.Container{
				tyk,
				redis,
				httpbin,
			},
			Network: network,
		},
		URI: uri,
	}, nil
}
