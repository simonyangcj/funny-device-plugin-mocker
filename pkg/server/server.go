package server

import (
	"context"
	"fmt"
	fDevice "funny-device-plugin-mocker/pkg/device"
	"net"
	"os"
	"path"
	"time"

	"funny-device-plugin-mocker/pkg/log"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	pluginApi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

const (
	serverSock = pluginApi.DevicePluginPath + "funny-device-plugin-mocker.sock"
)

type DevicePlugin struct {
	socket string

	stop   chan interface{}
	health chan *pluginApi.Device

	server       *grpc.Server
	rootPath     string
	prefix       string
	resourceName string
}

func NewDevicePlugin(rootPath, prefix, resourceName string) (*DevicePlugin, error) {

	return &DevicePlugin{
		socket:       serverSock,
		stop:         make(chan interface{}),
		health:       make(chan *pluginApi.Device),
		rootPath:     rootPath,
		prefix:       prefix,
		resourceName: resourceName,
	}, nil
}

// no option is required
func (m *DevicePlugin) GetDevicePluginOptions(context.Context, *pluginApi.Empty) (*pluginApi.DevicePluginOptions, error) {
	return &pluginApi.DevicePluginOptions{}, nil
}

func dial(unixSocketPath string, timeout time.Duration) (*grpc.ClientConn, error) {

	c, err := grpc.Dial(unixSocketPath, grpc.WithInsecure(), grpc.WithBlock(),
		grpc.WithTimeout(timeout),
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("unix", addr, timeout)
		}),
	)

	if err != nil {
		return nil, err
	}

	return c, nil
}

func (m *DevicePlugin) cleanup() error {
	if err := os.Remove(m.socket); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (m *DevicePlugin) Start() error {
	err := m.cleanup()
	if err != nil {
		return err
	}

	sock, err := net.Listen("unix", m.socket)
	if err != nil {
		return err
	}

	m.server = grpc.NewServer([]grpc.ServerOption{}...)
	pluginApi.RegisterDevicePluginServer(m.server, m)

	go m.server.Serve(sock)

	// Wait for server to start by launching a blocking connection
	conn, err := dial(m.socket, 5*time.Second)
	if err != nil {
		return err
	}
	conn.Close()

	// go m.healthcheck()

	return nil
}

// Stop stops the gRPC server
func (m *DevicePlugin) Stop() error {
	if m.server == nil {
		return nil
	}
	m.server.Stop()
	m.server = nil
	close(m.stop)

	return m.cleanup()
}

// Register registers the device plugin for the given resourceName with Kubelet.
func (m *DevicePlugin) Register(kubeletEndpoint, resourceName string) error {
	conn, err := dial(kubeletEndpoint, 5*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pluginApi.NewRegistrationClient(conn)
	reqt := &pluginApi.RegisterRequest{
		Version:      pluginApi.Version,
		Endpoint:     path.Base(m.socket),
		ResourceName: resourceName,
	}

	_, err = client.Register(context.Background(), reqt)
	if err != nil {
		return err
	}

	return nil
}

// ListAndWatch lists devices and update that list according to the health status
func (m *DevicePlugin) ListAndWatch(e *pluginApi.Empty, s pluginApi.DevicePlugin_ListAndWatchServer) error {
	devs, err := m.LoadDevices()
	if err != nil {
		log.Logger.Error("Failed to get devices", zap.String("error", err.Error()))
		return err
	}
	log.Logger.Info(fmt.Sprintf("Exposing devices: %v", devs))
	s.Send(&pluginApi.ListAndWatchResponse{Devices: devs})

	for {
		select {
		case <-m.stop:
			return nil
		case d := <-m.health:
			// FIXME: there is no way to recover from the Unhealthy state.
			d.Health = pluginApi.Unhealthy
			s.Send(&pluginApi.ListAndWatchResponse{Devices: devs})
		}
	}
}

// Allocate which return list of devices.
func (m *DevicePlugin) Allocate(ctx context.Context, reqs *pluginApi.AllocateRequest) (*pluginApi.AllocateResponse, error) {
	log.Logger.Info("Allocate request", zap.Any("request", reqs))
	devices, err := fDevice.GetDirectoriesToMap(m.rootPath, m.prefix)
	if err != nil {
		return nil, err
	}
	responses := pluginApi.AllocateResponse{}
	for _, req := range reqs.ContainerRequests {
		ds := make([]*pluginApi.DeviceSpec, len(req.DevicesIDs))
		response := pluginApi.ContainerAllocateResponse{Devices: ds}

		for i, dId := range req.DevicesIDs {
			ds[i] = &pluginApi.DeviceSpec{
				HostPath:      devices[dId].Path,
				ContainerPath: "/tmp", // for fake one just use tmp is good enough, because no app in that container will actually do something about it
				Permissions:   "rwm",
			}
		}
		responses.ContainerResponses = append(responses.ContainerResponses, &response)
	}

	log.Logger.Info("Allocate response", zap.Any("response", responses))
	return &responses, nil
}

// dir can be create or delete by user all the time
// so we always need to load devices
// performance is not good but it's ok for test
func (m *DevicePlugin) LoadDevices() ([]*pluginApi.Device, error) {
	devices, err := fDevice.GetDirectories(m.rootPath, m.prefix)
	if err != nil {
		return nil, err
	}
	var devs []*pluginApi.Device
	for _, dev := range devices {
		devs = append(devs, &pluginApi.Device{
			ID:     dev.Name,
			Health: pluginApi.Healthy,
		})
	}
	return devs, nil
}

func (m *DevicePlugin) LoadDevicesToMap() (map[string]*pluginApi.Device, error) {
	devices, err := fDevice.GetDirectories(m.rootPath, m.prefix)
	if err != nil {
		return nil, err
	}
	var devs map[string]*pluginApi.Device
	for _, dev := range devices {
		devs[dev.Name] = &pluginApi.Device{
			ID:     dev.Name,
			Health: pluginApi.Healthy,
		}
	}
	return devs, nil
}

func (m *DevicePlugin) GetPreferredAllocation(context context.Context, req *pluginApi.PreferredAllocationRequest) (*pluginApi.PreferredAllocationResponse, error) {
	return nil, nil
}

func (m *DevicePlugin) PreStartContainer(context context.Context, req *pluginApi.PreStartContainerRequest) (*pluginApi.PreStartContainerResponse, error) {
	return nil, nil
}

// Serve starts the gRPC server and register the device plugin to Kubelet
func (m *DevicePlugin) Serve() error {
	err := m.Start()
	if err != nil {
		log.Logger.Error("Could not start device plugin", zap.String("error", err.Error()))
		return err
	}
	log.Logger.Info("Started device plugin", zap.String("socket", m.socket))

	err = m.Register(pluginApi.KubeletSocket, m.resourceName)
	if err != nil {
		log.Logger.Error("Could not register device plugin", zap.String("error", err.Error()))
		m.Stop()
		return err
	}
	log.Logger.Info("Registered device plugin with Kubelet")
	return nil
}
