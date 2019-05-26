---
approvers:
- TBD
authors:
- '@egernst'
creation-date: "2019-02-26"
date: "2019-02-26T00:00:00Z"
draft: false
editor: TBD
last-updated: "2019-04-12"
owning-sig: sig-node
participating-sigs:
- sig-scheduling
- sig-autoscaling
- sig-windows
reviewers:
- '@tallclair'
- '@derekwaynecarr'
- '@dchen1107'
status: implementable
tags:
- sig-node
title: Pod Overhead
---
# Pod Overhead

This includes the Summary and Motivation sections.

## Table of Contents

- [Release Signoff Checklist](#release-signoff-checklist)
- [Summary](#summary)
- [Motivation](#motivation)
  - [Goals](#goals)
  - [Non-Goals](#non-goals)
- [Proposal](#proposal)
  - [API Design](#api-design)
    - [Pod overhead](#pod-overhead)
    - [Container Runtime Interface (CRI)](#container-runtime-interface-cri)
  - [ResourceQuota changes](#resourcequota-changes)
  - [RuntimeClass changes](#runtimeclass-changes)
  - [RuntimeClass admission controller](#runtimeclass-admission-controller)
  - [Implementation Details](#implementation-details)
  - [Risks and Mitigations](#risks-and-mitigations)
- [Design Details](#design-details)
  - [Graduation Criteria](#graduation-criteria)
  - [Upgrade / Downgrade Strategy](#upgrade--downgrade-strategy)
  - [Version Skew Strategy](#version-skew-strategy)
- [Implementation History](#implementation-history)
- [Drawbacks](#drawbacks)
- [Alternatives](#alternatives)
  - [Introduce pod level resource requirements](#introduce-pod-level-resource-requirements)
  - [Leaving the PodSpec unchanged](#leaving-the-podspec-unchanged)

## Release Signoff Checklist

- [ ] kubernetes/enhancements issue in release milestone, which links to KEP (this should be a link to the KEP location in kubernetes/enhancements, not the initial KEP PR)
- [ ] KEP approvers have set the KEP status to `implementable`
- [ ] Design details are appropriately documented
- [ ] Test plan is in place, giving consideration to SIG Architecture and SIG Testing input
- [ ] Graduation criteria is in place
- [ ] "Implementation History" section is up-to-date for milestone
- [ ] User-facing documentation has been created in [kubernetes/website], for publication to [kubernetes.io]
- [ ] Supporting documentation e.g., additional design documents, links to mailing list discussions/SIG meetings, relevant PRs/issues, release notes

**Note:** Any PRs to move a KEP to `implementable` or significant changes once it is marked `implementable` should be approved by each of the KEP approvers. If any of those
approvers is no longer appropriate than changes to that list should be approved by the remaining approvers and/or the owning SIG (or SIG-arch for cross cutting KEPs).

**Note:** This checklist is iterative and should be reviewed and updated every time this enhancement is being considered for a milestone.

[kubernetes.io]: https://kubernetes.io/
[kubernetes/enhancements]: https://github.com/kubernetes/enhancements/issues
[kubernetes/kubernetes]: https://github.com/kubernetes/kubernetes
[kubernetes/website]: https://github.com/kubernetes/website

## Summary

Sandbox runtimes introduce a non-negligible overhead at the pod level which must be accounted for
effective scheduling, resource quota management, and constraining.

## Motivation

Pods have some resource overhead. In our traditional linux container (Docker) approach,
the accounted overhead is limited to the infra (pause) container, but also invokes some
overhead accounted to various system components including: Kubelet (control loops), Docker,
kernel (various resources), fluentd (logs). The current approach is to reserve a chunk
of resources for the system components (system-reserved, kube-reserved, fluentd resource
request padding), and ignore the (relatively small) overhead from the pause container, but
this approach is heuristic at best and doesn't scale well.

With sandbox pods, the pod overhead potentially becomes much larger, maybe O(100mb). For
example, Kata containers must run a guest kernel, kata agent, init system, etc. Since this
overhead is too big to ignore, we need a way to account for it, starting from quota enforcement
and scheduling.

### Goals

* Provide a mechanism for accounting pod overheads which are specific to a given runtime solution

### Non-Goals

* making runtimeClass selections
* auto-detecting overhead
* per-container overhead
* creation of pod-level resource requirements

## Proposal

Augment the RuntimeClass definition and the `PodSpec` to introduce
the field `Overhead *ResourceRequirements`. This field represents the overhead associated
with running a pod for a given runtimeClass.  A mutating admission controller is
introduced which will update the `Overhead` field in the workload's `PodSpec` to match
what is provided for the selected RuntimeClass, if one is specified.

Kubelet's creation of the pod cgroup will be calculated as the sum of container
`ResourceRequirements.Limits` fields, plus the Overhead associated with the pod.

The scheduler, resource quota handling, and Kubelet's pod cgroup creation and eviction handling
will take Overhead into account, as well as the sum of the pod's container requests.

Horizontal and Veritical autoscaling are calculated based on container level statistics,
so should not be impacted by pod Overhead.

### API Design

#### Pod overhead
Introduce a Pod.Spec.Overhead field on the pod to specify the pods overhead.

```
Pod {
  Spec PodSpec {
    // Overhead is the resource overhead incurred from the runtime.
    // +optional
    Overhead *ResourceRequirements
  }
}
```

All PodSpec and RuntimeClass fields are immutable, including the `Overhead` field. For scheduling,
the pod `Overhead` resource requests are added to the container resource requests.

We don't currently enforce resource limits on the pod cgroup, but this becomes feasible once
pod overhead is accountable. If the pod specifies a resource limit, and all containers in the
pod specify a limit, then the sum of those limits becomes a pod-level limit, enforced through the
pod cgroup.

Users are not expected to manually set `Overhead`; any prior values being set will result in the workload
being rejected. If runtimeClass is configured and selected in the PodSpec, `Overhead` will be set to the value
defined in the corresponding runtimeClass. This is described in detail in
[RuntimeClass admission controller](#runtimeclass-admission-controller).

Being able to specify resource requirements for a workload at the pod level instead of container
level has been discussed, but isn't proposed in this KEP.

In the event that pod-level requirements are introduced, pod overhead should be kept separate. This
simplifies several scenarios:
 - overhead, once added to the spec, stays with the workload, even if runtimeClass is redefined
 or removed.
 - the pod spec can be referenced directly from scheduler, resourceQuota controller and kubelet,
 instead of referencing a runtimeClass object which could have possibly been removed.

#### Container Runtime Interface (CRI)

The pod cgroup is managed by the Kubelet, so passing the pod-level resource to the CRI implementation
is not strictly necessary. However, some runtimes may wish to take advantage of this information, for
instance for sizing the Kata Container VM.

LinuxContainerResources is added to the LinuxPodSandboxConfig for both overhead and container
totals, as optional fields:

```
type LinuxPodSandboxConfig struct {
	Overhead *LinuxContainerResources
	ContainerResources *LinuxContainerResources
}
```

WindowsContainerResources is added to a newly created WindowsPodSandboxConfig for both overhead and container
totals, as optional fields:

```
type WindowsPodSandboxConfig struct {
	Overhead *WindowsContainerResources
	ContainerResources *WindowsContainerResources
}
```

ContainerResources field in the LinuxPodSandboxConfig and WindowsPodSandboxConfig matches the pod-level limits
(i.e. total of container limits). Overhead is tracked separately since the sandbox overhead won't necessarily
guide sandbox sizing, but instead used for better management of the resulting sandbox on the host. 

### ResourceQuota changes

Pod overhead will be counted against an entity's ResourceQuota. The controller will be updated to
add the pod `Overhead`to the container resource request summation.

### RuntimeClass changes

Expand the runtimeClass type to include sandbox overhead, `Overhead *Overhead`.

Where Overhead is defined as follows:

```
type Overhead struct {
  PodFixed *ResourceRequirements
}
```

In the future, the `Overhead` definition could be extended to include fields that describe a percentage
based overhead (scale the overhead based on the size of the pod), or container-level overheads. These are
left out of the scope of this proposal.

### RuntimeClass admission controller

The pod resource overhead must be defined prior to scheduling, and we shouldn't make the user
do it. To that end, we propose a mutating admission controller: RuntimeClass. This admission controller
is also proposed for the [native RuntimeClass scheduling KEP](https://github.com/kubernetes/enhancements/blob/master/keps/sig-node/runtime-class-scheduling.md).

In the scope of this KEP, The RuntimeClass controller will have a single job: set the pod overhead field in the
workload's PodSpec according to the runtimeClass specified.

It is expected that only the RuntimeClass controller will set Pod.Spec.Overhead. If a value is provided, the pod will
be rejected.

Going forward, I foresee additional controller scope around runtimeClass:
 - validating the runtimeClass selection: This would require applying some kind of pod-characteristic labels
 (runtimeClass selectors?) which would then be consumed by an admission controller and checked against known
 capabilities on a per runtimeClass basis. This is is beyond the scope of this proposal.
 - Automatic runtimeClass selection: A controller could exist which would attempt to automatically select the
 most appropriate runtimeClass for the given pod. This, again, is beyond the scope of this proposal.

### Implementation Details

With this proposal, the following changes are required:
 - Add the new API to the pod spec and RuntimeClass
 - Update the RuntimeClass controller to merge the overhead into the pod spec
 - Update the ResourceQuota controller to account for overhead
 - Update the scheduler to account for overhead
 - Update the kubelet (admission, eviction, cgroup limits) to handle overhead

### Risks and Mitigations

This proposal introduces changes across several Kubernetes components and a change in behavior *if* Overhead fields
are utilized. To help mitigate this risk, I propose that this be treated as a new feature with an independent feature gate.

## Design Details

### Graduation Criteria

This KEP will be treated as a new feature, and will be introduced with a new feature gate, PodOverhead.

Plan to introduce this utilizing maturity levels: alpha, beta and stable.

Graduation criteria between these levels to be determined.

### Upgrade / Downgrade Strategy

If applicable, how will the component be upgraded and downgraded? Make sure this is in the test plan.

Consider the following in developing an upgrade/downgrade strategy for this enhancement:
- What changes (in invocations, configurations, API use, etc.) is an existing cluster required to make on upgrade in order to keep previous behavior?
- What changes (in invocations, configurations, API use, etc.) is an existing cluster required to make on upgrade in order to make use of the enhancement?

### Version Skew Strategy

Set the overhead to the max of the two version until the rollout is complete.  This may be more problematic
if a new version increases (rather than decreases) the required resources.

## Implementation History

- 2019-04-04: Initial KEP published.

## Drawbacks

This KEP introduces further complexity, and adds a field the PodSpec which users aren't expected to modify.

## Alternatives

In order to achieve proper handling of sandbox runtimes, the scheduler/resourceQuota handling needs to take
into account the overheads associated with running a particular runtimeClass.

### Introduce pod level resource requirements

Rather than just introduce overhead, add support for general pod-level resource requirements. Pod level
resource requirements are useful for shared resources (hugepages, memory when doing emptyDir volumes).

Even if this were to be introduced, there is a benefit in keeping the overhead separate.
 - post-pod creation handling of pod events: if runtimeClass definition is removed after a pod is created,
  it will be very complicated to calculate which part of the pod resource requirements were associated with
  the workloads versus the sandbox overhead.
 - a kubernetes service provider can subsidize the charge-back model potentially and eat the cost of the
 runtime choice, but charge the user for the cpu/memory consumed independent of runtime choice.


### Leaving the PodSpec unchanged

Instead of tracking the overhead associated with running a workload with a given runtimeClass in the PodSpec,
the Kubelet (for pod cgroup creation), the scheduler (for honoring reqests overhead for the pod) and the resource
quota handling (for optionally taking requests overhead of a workload into account) will need to be augmented
to add a sandbox overhead when applicable.

Pros:
 * no changes to the pod spec
 * user does not have the option of setting the overhead
 * no need for a mutating admission controller

Cons:
 * handling of the pod overhead is spread out across a few components
 * Not user perceptible from a workload perspective.
 * very complicated if the runtimeClass policy changes after workloads are running