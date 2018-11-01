# Getting Started

A quick start guide to get Node-Recovery up and running inside our container based
development cluster.

## Building

The Node-Recovery build system runs completely inside docker. In order to build
Node-Recovery you need to have `docker` and `rsync` installed. You also need to have `docker`
running, and have the [permissions](https://docs.docker.com/install/linux/linux-postinstall/#manage-docker-as-a-non-root-user) to access it.

### Dockerized environment

Runs master and nodes containers, when each one of them run virtual machine via QEMU.
In additional it runs dnsmasq and docker registry containers.

### Compile and run it

To build all required artifacts and launch the
dockerizied environment, clone the KubeVirt repository, `cd` into it, and:

```bash
# Build and deploy KubeVirt on Kubernetes 1.10.4 in our vms inside containers
export CLUSTER_PROVIDER=k8s-1.10.4 # this is also the default if no CLUSTER_PROVIDER is set
make cluster-up
make cluster-sync
```

This will create a virtual machine called `node01` which acts as node and master. To create
more nodes which will register themselves on master, you can use the
`CLUSTER_NUM_NODES` environment variable. This would create a master and one
node:

```bash
export CLUSTER_NUM_NODES=2 # schedulable master + one additional node
make cluster-up
```

To destroy the created cluster, type

```
make cluster-down
```

**Note:** Whenever you type `make cluster-down && make cluster-up`, you will
have a completely fresh cluster to play with.

### Accessing the containerized nodes via ssh

The containerized nodes are named starting from `node01`, `node02`, and so
forth. `node01` is always the master of the cluster.

Every node can be accessed via its name:

```bash
cluster/cli.sh ssh node01
```

To execute a remote command, e.g `ls`, simply type

```bash
cluster/cli.sh ssh node01 -- ls -lh
```

### Automatic Code Generation

Some of the code in our source tree is auto-generated (see `git ls-files|grep '^pkg/.*generated.*go$'`).
On certain occasions (but not when building git-cloned code), you would need to regenerate it
with

```bash
make generate
```

Typical cases where code regeneration should be triggered are:

 * When changing APIs, REST paths or their comments (gets reflected in the api documentation, clients, generated cloners...)

 We have a check in our CI system, so that you don't miss when `make generate` needs to be called.

### Testing

After a successful build you can run the *unit tests*:

```bash
    make test
```

They do not need a running Node-Recovery environment to succeed.

## Use

Congratulations you are still with us and you have build Node-Recovery.

Now it's time to get hands on and give it a try.
