# funny-device-plugin-mocker

It's a tool for test purpose, to mock the device plugin for k8s.

## Quick Start

```bash
# download project and build
$ cd cmd/funny-device-plugin-mocker
$ go build .

# create some empty dir to simulate the device plugin
$ mkdir -p /mnt/fake
$ cd /mnt/fake
$ mkdir fdpm_dir1
$ mkdir fdpm_dir2
$ mkdir fdpm_dir3

# run the mocker
$ ./funny-device-plugin-mocker run --resource-name nvidia.com/gpu

# check the node by describe it
$ kubectl describe node <node-name>
# you will see the device appear in the node status
# output
...
Capacity:
  nvidia.com/gpu: 3
```

