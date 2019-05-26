---
approvers:
- '@msau42'
- '@saad-ali'
authors:
- '@verult'
creation-date: "2019-01-24"
date: "2019-01-24T00:00:00Z"
draft: false
editor: TBD
last-updated: "2019-01-24"
owning-sig: sig-storage
participating-sigs:
- sig-storage
reviewers:
- '@msau42'
- '@saad-ali'
status: implementable
tags:
- sig-storage
title: CSI Volume Topology
---
# Title

CSI Volume Topology

## Table of Contents

  * [Title](#title)
      * [Table of Contents](#table-of-contents)
      * [Summary](#summary)
      * [Test Plan](#test-plan)
      * [Graduation Criteria](#graduation-criteria)

## Summary

This KEP is written after the original design doc has been approved and implemented. Design for CSI Volume Topology Support in Kubernetes is incorporated as part of the [CSI Volume Plugins in Kubernetes Design Doc](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/storage/container-storage-interface.md).

The rest of the document includes required information missing from the original design document: test plan and graduation criteria.

## Test Plan
* Unit tests around topology logic in kubelet and CSI external-provisioner.
* New e2e tests around topology features will be added in CSI e2e test suites, which test various volume operation behaviors from the perspective of the end user. Tests include:
  * (Positive) Volume provisioning with immediate volume binding and AllowedTopologies set.
  * (Positive) Volume provisioning with delayed volume binding.
  * (Positive) Volume provisioning with delayed volume binding and AllowedTopologies set.
  * (Negative) Volume provisioning with immediate volume binding and pod zone missing from AllowedTopologies.
  * (Negative) Volume provisioning with delayed volume binding and pod zone missing from AllowedTopologies.
Initially topology tests are run against a single CSI driver. As the CSI test suites become modularized they will run against arbitrary CSI drivers.

## Graduation Criteria

### Alpha->Beta

* Feature complete, including:
  * Volume provisioning with required topology constraints
  * Volume provisioning with preferred topology
  * Cluster-wide topology aggregation
  * StatefulSet volume spreading
* Depends on: CSINodeInfo beta or above; Kubelet Plugins Watcher beta or above
* Unit and e2e tests implemented

### Beta->GA

* Depends on: CSINodeInfo GA; Kubelet Plugins Watcher GA
* Stress test: provisioning load tests; node scale tests; component crash tests
* Feature deployed in production and have gone through at least one K8s upgrade.