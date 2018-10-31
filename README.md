# Node-Recovery

[![Build Status](https://travis-ci.com/kubevirt/node-recovery.svg?branch=master)](https://travis-ci.com/kubevirt/node-recovery)

**Node-Recovery** monitors all nodes under the cluster and call to [cluster API][cluster-api] to remediate non-ready nodes.

**Note:** Node-Recovery is a heavy work in progress.

# Introduction

Often you fall in the situatuion that your node start to be `non-Ready`, it can be resources issues, but it also can be network or kubelet issue and in this case reboot of the bare metal host can help to solve the problem. **Node-Recovery** controller will recognize different node conditions and apply remediation logic, that will re-create `machine` object that will trigger `machine` actuator.

# To start using Node-Recovery

### Prerequisite

- you must install `machine` and `cluster` controllers on your cluster, for bare metal nodes you can use [cluster-api-external-provider][cluster-api-external-provider], for other possibilities check [cluster API repository](https://github.com/kubernetes-sigs/cluster-api#provider-implementations)

Apply `noderecovery.yaml` manifest from the desired release.

# To start developing KubeVirt

To set up a development environment please read our
[Getting Started Guide](/docs/getting-started.md).

You can learn more about how KubeVirt is designed (and why it is that way),
and learn more about the major components by taking a look at
[our developer documentation](docs/):

 * [Architecture](docs/architecture.md) - High-level view on the architecture

# Community

If you got enough of code and want to speak to people, then you got a couple
of options:

* Follow us on [Twitter](https://twitter.com/kubevirt)
* Chat with us on IRC via [#kubevirt @ irc.freenode.net](https://kiwiirc.com/client/irc.freenode.net/kubevirt)
* Discuss with us on the [kubevirt-dev Google Group](https://groups.google.com/forum/#!forum/kubevirt-dev)
* Stay informed about designs and upcoming events by watching our [community content](https://github.com/kubevirt/community/)
* Take a glance at [future planning](https://trello.com/b/50CuosoD/kubevirt)

## Related resources

 * [Kubernetes][k8s]
 * [node-detector][node-detector]
 * [cluster-API][cluster-api]
 * [cluster-api-external-provider][cluster-api-external-provider]

## Submitting patches

When sending patches to the project, the submitter is required to certify that
they have the legal right to submit the code. This is achieved by adding a line

    Signed-off-by: Real Name <email@address.com>

to the bottom of every commit message. Existence of such a line certifies
that the submitter has complied with the Developer's Certificate of Origin 1.1,
(as defined in the file docs/developer-certificate-of-origin).

This line can be automatically added to a commit in the correct format, by
using the '-s' option to 'git commit'.
## License

Node-Recovery is distributed under the
[Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0.txt).

    Copyright 2016

    Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at

        http://www.apache.org/licenses/LICENSE-2.0

    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.

[//]: # (Reference links)
   [k8s]: https://kubernetes.io
   [node-detector]: https://github.com/kubernetes/node-problem-detector
   [cluster-api]: https://github.com/kubernetes-sigs/cluster-api
   [cluster-api-external-provider]: https://github.com/kubevirt/cluster-api-provider-external
