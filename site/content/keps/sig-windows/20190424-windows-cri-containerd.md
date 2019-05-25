---
approvers:
- '@michmike'
authors:
- '@patricklang'
creation-date: "2019-04-24"
editor: TBD
last-updated: "2019-04-24"
owning-sig: sig-windows
participating-sigs:
- sig-windows
reviewers:
- '@yujuhong'
- '@derekwaynecarr'
- '@tallclair'
status: implementable
title: Supporting CRI-ContainerD on Windows
---
# Supporting CRI-ContainerD on Windows

## Table of Contents

<!-- TOC -->

- [Table of Contents](#table-of-contents)
- [Release Signoff Checklist](#release-signoff-checklist)
- [Summary](#summary)
- [Motivation](#motivation)
    - [Goals](#goals)
    - [Non-Goals](#non-goals)
- [Proposal](#proposal)
    - [User Stories](#user-stories)
        - [Improving Kubernetes integration for Windows Server containers](#improving-kubernetes-integration-for-windows-server-containers)
        - [Improved isolation and compatibility between Windows pods using Hyper-V](#improved-isolation-and-compatibility-between-windows-pods-using-hyper-v)
        - [Improve Control over Memory & CPU Resources with Hyper-V](#improve-control-over-memory--cpu-resources-with-hyper-v)
        - [Improved Storage Control with Hyper-V](#improved-storage-control-with-hyper-v)
        - [Enable runtime resizing of container resources](#enable-runtime-resizing-of-container-resources)
    - [Implementation Details/Notes/Constraints](#implementation-detailsnotesconstraints)
        - [Proposal: Use Runtimeclass Scheduler to simplify deployments based on OS version requirements](#proposal-use-runtimeclass-scheduler-to-simplify-deployments-based-on-os-version-requirements)
        - [Proposal: Standardize hypervisor annotations](#proposal-standardize-hypervisor-annotations)
    - [Dependencies](#dependencies)
    - [Risks and Mitigations](#risks-and-mitigations)
        - [CRI-ContainerD availability](#cri-containerd-availability)
- [Design Details](#design-details)
    - [Test Plan](#test-plan)
    - [Graduation Criteria](#graduation-criteria)
    - [Alpha release (proposed 1.15)](#alpha-release-proposed-115)
    - [Alpha -> Beta Graduation](#alpha---beta-graduation)
    - [Upgrade / Downgrade Strategy](#upgrade--downgrade-strategy)
    - [Version Skew Strategy](#version-skew-strategy)
- [Implementation History](#implementation-history)
- [Alternatives](#alternatives)
    - [CRI-O](#cri-o)
- [Infrastructure Needed](#infrastructure-needed)

<!-- /TOC -->

## Release Signoff Checklist

**ACTION REQUIRED:** In order to merge code into a release, there must be an issue in [kubernetes/enhancements] referencing this KEP and targeting a release milestone **before [Enhancement Freeze](https://github.com/kubernetes/sig-release/tree/master/releases)
of the targeted release**.

For enhancements that make changes to code or processes/procedures in core Kubernetes i.e., [kubernetes/kubernetes], we require the following Release Signoff checklist to be completed.

Check these off as they are completed for the Release Team to track. These checklist items _must_ be updated for the enhancement to be released.

- [ ] kubernetes/enhancements issue in release milestone, which links to KEP (this should be a link to the KEP location in kubernetes/enhancements, not the initial KEP PR)
- [ ] KEP approvers have set the KEP status to `implementable`
- [ ] Design details are appropriately documented
- [ ] Test plan is in place, giving consideration to SIG Architecture and SIG Testing input
- [ ] Graduation criteria is in place
- [ ] "Implementation History" section is up-to-date for milestone
- [ ] User-facing documentation has been created in [kubernetes/website], for publication to [kubernetes.io]
- [ ] Supporting documentation e.g., additional design documents, links to mailing list discussions/SIG meetings, relevant PRs/issues, release notes

**Note:** Any PRs to move a KEP to `implementable` or significant changes once it is marked `implementable` should be approved by each of the KEP approvers. If any of those approvers is no longer appropriate than changes to that list should be approved by the remaining approvers and/or the owning SIG (or SIG-arch for cross cutting KEPs).

**Note:** This checklist is iterative and should be reviewed and updated every time this enhancement is being considered for a milestone.

[kubernetes.io]: https://kubernetes.io/
[kubernetes/enhancements]: https://github.com/kubernetes/enhancements/issues
[kubernetes/kubernetes]: https://github.com/kubernetes/kubernetes
[kubernetes/website]: https://github.com/kubernetes/website

## Summary

The ContainerD maintainers have been working on CRI support which is stable on Linux, but is not yet available for Windows as of ContainerD 1.2. Currently it’s planned for ContainerD 1.3, and the developers in the Windows container platform team have most of the key work merged into master already. Supporting CRI-ContainerD on Windows means users will be able to take advantage of the latest container platform improvements that shipped in Windows Server 2019 / 1809 and beyond.


## Motivation

Windows Server 2019 includes an updated host container service (HCS v2) that offers more control over how containers are managed. This can remove some limitations and improve some Kubernetes API compatibility. However, the current Docker EE 18.09 release has not been updated to work with the Windows HCSv2, only ContainerD has been migrated. Moving to CRI-ContainerD allows the Windows OS team and Kubernetes developers to focus on an interface designed to work with Kubernetes to improve compatibility and accelerate development.

Additionally, users could choose to run with only CRI-ContainerD instead of Docker EE if they wanted to reduce the install footprint or produce their own self-supported CRI-ContainerD builds.

### Goals

- Improve the matrix of Kubernetes features that can be supported on Windows
- Provide a path forward to implement Kubernetes-specific features that are not available in the Docker API today
- Align with `dockershim` deprecation timelines once they are defined

### Non-Goals

- Running Linux containers on Windows nodes
- Deprecating `dockershim`. This is out of scope for this KEP. The effort to migrate that code out of tree is in [KEP PR 866](https://github.com/kubernetes/enhancements/pull/866) and deprecation discussions will happen later.

## Proposal

### User Stories

#### Improving Kubernetes integration for Windows Server containers

Moving to the new Windows HCSv2 platform and ContainerD would allow Kubernetes to add support for:

- Mounting single files, not just folders, into containers
- Termination messages (depends on single file mounts)
- /etc/hosts (c:\windows\system32\drivers\etc\hosts) file mapping

#### Improved isolation and compatibility between Windows pods using Hyper-V 

Hyper-V enables each pod to run within it’s own hypervisor partition, with a separate kernel. This means that we can build forward-compatibility for containers across Windows OS versions - for example a container built using Windows Server 1809, could be run on a node running Windows Server 19H1. This pod would use the Windows Server 1809 kernel to preserve full compatibility, and other pods could run using either a shared kernel with the node, or their own isolated Windows Server 19H1 kernels. Containers requiring 1809 and 19H1 (or later) cannot be mixed in the same pod, they must be deployed in separate pods so the matching kernel may be used. Running Windows Server version 19H1 containers on a Windows Server 2019/1809 host will not work.

In addition, some customers may desire hypervisor-based isolation as an additional line of defense against a container break-out attack.

Adding Hyper-V support would use [RuntimeClass](https://kubernetes.io/docs/concepts/containers/runtime-class/#runtime-class). 
3 typical RuntimeClass names would be configured in CRI-ContainerD to support common deployments:
- runhcs-wcow-process [default] - process isolation is used, container & node OS version must match
- runhcs-wcow-hypervisor - Hyper-V isolation is used, Pod will be compatible with containers built with Windows Server 2019 / 1809. Physical memory overcommit is allowed with overages filled from pagefile.
- runhcs-wcow-hypervisor-19H1 - Hyper-V isolation is used, Pod will be compatible with containers built with Windows Server 19H1. Physical memory overcommit is allowed with overages filled from pagefile.

Using Hyper-V isolation does require some extra memory for the isolated kernel & system processes. This could be accounted for by implementing the [PodOverhead](https://kubernetes.io/docs/concepts/containers/runtime-class/#runtime-class) proposal for those runtime classes. We would include a recommended PodOverhead in the default CRDs, likely between 100-200M.


#### Improve Control over Memory & CPU Resources with Hyper-V

The Windows kernel itself cannot provide reserved memory for pods, containers or processes. They are always fulfilled using virtual allocations which could be paged out later. However, using a Hyper-V partition improves control over memory and CPU cores. Hyper-V can either allocate memory on-demand (while still enforcing a hard limit), or it can be reserved as a physical allocation up front. Physical allocations may be able to enable large page allocations within that range (to be confirmed) and improve cache coherency. CPU core counts may also be limited so a pod only has certain cores available, rather than shares spread across all cores, and applications can tune thread counts to the actually available cores.

Operators could deploy additional RuntimeClasses with more granular control for performance critical workloads:
- 2019-Hyper-V-Reserve: Hyper-V isolation is used, Pod will be compatible with containers built with Windows Server 2019 / 1809. Memory reserve == limit, and is guaranteed to not page out.
  - 2019-Hyper-V-Reserve-<N>Core: Same as above, except all but <N> CPU cores are masked out.
- 19H1-Hyper-V-Reserve: Hyper-V isolation is used, Pod will be compatible with containers built with Windows Server 19H1. Memory reserve == limit, and is guaranteed to not page out.
  - 19H1-Hyper-V-Reserve-<N>Core: Same as above, except all but <N> CPU cores are masked out.


#### Improved Storage Control with Hyper-V


Hyper-V also brings the capability to attach storage to pods using block-based protocols (SCSI) instead of file-based protocols (host file mapping / NFS / SMB). These capabilities could be enabled in HCSv2 with CRI-ContainerD, so this could be an area of future work. Some examples could include:

Attaching a "physical disk" (such as a local SSD, iSCSI target, Azure Disk or Google Persistent Disk) directly to a pod. The kubelet would need to identify the disk beforehand, then attach it as the pod is created with CRI. It could then be formatted and used within the pod without being mounted or accessible on the host.

Creating [Persistent Local Volumes](https://kubernetes.io/docs/concepts/storage/volumes/#local) using a local virtual disk attached directly to a pod. This would create local, non-resilient storage that could be formatted from the pod without being mounted on the host. This could be used to build out more resource controls such as fixed disk sizes and QoS based on IOPs or throughput and take advantage of high speed local storage such as temporary SSDs offered by cloud providers.


#### Enable runtime resizing of container resources

With virtual-based allocations and Hyper-V, it should be possible to increase the limit for a running pod. This won’t give it a guaranteed allocation, but will allow it to grow without terminating and scheduling a new pod. This could be a path to vertical pod autoscaling.


### Implementation Details/Notes/Constraints

#### Proposal: Use Runtimeclass Scheduler to simplify deployments based on OS version requirements

As of version 1.14, RuntimeClass is not considered by the Kubernetes scheduler. There’s no guarantee that a node can start a pod, and it could fail until it’s scheduled on an appropriate node. Additional node labels and nodeSelectors are required to avoid this problem. [RuntimeClass Scheduling](https://github.com/kubernetes/enhancements/blob/master/keps/sig-node/runtime-class-scheduling.md) proposes being able to add nodeSelectors automatically when using a RuntimeClass, simplifying the deployment.

Windows forward compatibility will bring a new challenge as well because there are two ways a container could be run:
- Constrained to the OS version it was designed for, using process-based isolation
- Running on a newer OS version using Hyper-V.
This second case could be enabled with a RuntimeClass. If a separate RuntimeClass was used based on OS version, this means the scheduler could find a node with matching class.

#### Proposal: Standardize hypervisor annotations

There are large number of [Windows annotations](https://github.com/Microsoft/hcsshim/blob/master/internal/oci/uvm.go#L15) defined that can control how Hyper-V will configure its hypervisor partition for the pod. Today, these could be set in the runtimeclasses defined in the CRI-ContainerD configuration file on the node, but it would be easier to maintain them if key settings around resources (cpu+memory+storage) could be aligned across multiple hypervisors and exposed in CRI.

Doing this would make pod definitions more portable between different isolation types. It would also avoid the need for a "t-shirt size" list of RuntimeClass instances to choose from:
- 1809-Hyper-V-Reserve-2Core-PhysicalMemory
- 19H1-Hyper-V-Reserve-1Core-VirtualMemory
- 19H1-Hyper-V-Reserve-4Core-PhysicalMemory
...

### Dependencies

##### Windows Server 2019

This work would be carried out and tested using the already-released Windows Server 2019. That will enable customers a migration path from Docker 18.09 to CRI-ContainerD if they want to get this new functionality. Windows Server 19H1 and later will also be supported once they’re tested.

##### CRI-ContainerD

It was announced that the upcoming 1.3 release would include Windows support, but that release and timeline are still in planning as of early April 2019.

The code needed to run ContainerD is merged, and [experimental support in moby](https://github.com/moby/moby/pull/38541) has merged. The CRI plugin changes for Windows are still in a development branch [jterry75/cri](https://github.com/jterry75/cri/tree/windows_port/cmd/containerd) and don’t have an upstream PR open yet. 

Code: mostly done
CI+CD: lacking

##### CNI: Flannel 
Flannel isn’t expected to require any changes since the Windows-specific metaplugins ship outside of the main repo. However, there is still not a stable release supporting Windows so it needs to be built from source. Additionally, the Windows-specific metaplugins to support ContainerD are being developed in a new repo [Microsoft/windows-container-networking](https://github.com/Microsoft/windows-container-networking). It’s still TBD whether this code will be merged into [containernetworking/plugins](https://github.com/containernetworking/plugins/), or maintained in a separate repo.
- Sdnbridge - this works with host-gw mode, replaces win-bridge
- Sdnoverlay - this works with vxlan overlay mode, replaces win-overlay

Code: in progress
CI+CD: lacking

##### CNI: Kubenet

The same sdnbridge plugin should work with kubenet as well. If someone would like to use kubenet instead of flannel, that should be feasible.

##### CNI: GCE

GCE uses the win-bridge meta-plugin today for managing Windows network interfaces. This would also need to migrate to sdnbridge.

##### Storage: in-tree AzureFile, AzureDisk, Google PD

These are expected to work and the same tests will be run for both dockershim and CRI-ContainerD.

##### Storage: FlexVolume for iSCSI & SMB
These out-of-tree plugins are expected to work, and are not tested in prow jobs today. If they graduate to stable we’ll add them to testgrid.

### Risks and Mitigations

#### CRI-ContainerD availability

As mentioned earlier, builds are not yet available. We will publish the setup steps required to build & test in the kubernetes-sigs/windows-testing repo during the course of alpha so testing can commence.

## Design Details

### Test Plan

The existing test cases running on Testgrid that cover Windows Server 2019 with Docker will be reused with CRI-ContainerD. Testgrid will be updated so that both ContainerD and dockershim results are visible. 

Test cases that depend on ContainerD and won't pass with Dockershim will be marked with `[feature:windows-containerd]` until `dockershim` is deprecated.

### Graduation Criteria

### Alpha release (proposed 1.15)

- Windows Server 2019 containers can run with process level isolation
- TestGrid has results for Kubernetes master branch. CRI-ContainerD and CNI built from source and may include non-upstream PRs.
- Support RuntimeClass to enable Hyper-V isolation for Windows Server 2019 on 2019


### Alpha -> Beta Graduation

- Feature parity with dockershim, including:
  - Group Managed Service Account support
  - Named pipe & Unix domain socket mounts
- Support RuntimeClass to enable Hyper-V isolation and run Windows Server 2019 containers on 19H1
- Publically available builds (beta or better) of CRI-ContainerD, at least one CNI
- TestGrid results for above builds with Kubernetes master branch


##### Beta -> GA Graduation

- Stable release of CRI-ContainerD on Windows, at least one CNI
- Master & release branches on TestGrid

### Upgrade / Downgrade Strategy

Because no Kubernetes API changes are expected, there is no planned upgrade/downgrade testing at the cluster level.

Node upgrade/downgrade is currently out of scope of the Kubernetes project, but we'll aim to include CRI-ContainerD in other efforts such as `kubeadm` bootstrapping for nodes.

As discussed in SIG-Node, there's also no testing on switching CRI on an existing node. These are expected to be installed and configured as a prerequisite before joining a node to the cluster.

### Version Skew Strategy

There's no version skew considerations needed for the same reasons described in upgrade/downgrade strategy.

## Implementation History

- 2014-04-24 - KEP started, based on the [earlier doc shared SIG-Windows and SIG-Node](https://docs.google.com/document/d/1NigFz1nxI9XOi6sGblp_1m-rG9Ne6ELUrNO0V_TJqhI/edit)

## Alternatives

### CRI-O

[CRI-O](https://cri-o.io/) is another runtime that aims to closely support all the fields available in the CRI spec. Currently there aren't any maintainers porting it to Windows so it's not a viable alternative.

## Infrastructure Needed

No new infrastructure is currently needed from the Kubernetes community. The existing test jobs using prow & testgrid will be copied and modified to test CRI-ContainerD in addition to dockershim.