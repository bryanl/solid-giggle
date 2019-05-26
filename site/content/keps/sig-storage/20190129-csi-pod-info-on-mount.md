---
approvers:
- '@saad-ali'
authors:
- '@jsafrane'
creation-date: "2019-01-29"
date: "2019-01-29T00:00:00Z"
draft: false
editor: TBD
last-updated: "2019-01-29"
owning-sig: sig-storage
participating-sigs:
- sig-storage
reviewers:
- '@msau42'
- '@saad-ali'
status: implementable
tags:
- sig-storage
title: CSI Pod Info on Mount
---
# CSI Pod Info on Mount

## Table of Contents

* [Summary](#summary)
* [Test Plan](#test-plan)
* [Graduation Criteria](#graduation-criteria)
  * [Alpha tp Beta](#alpha-to-beta)
  * [Beta to GA](#beta-to-ga)
* [Implementation History](#implementation-history)

[Tools for generating]: https://github.com/ekalinin/github-markdown-toc

## Summary

This document presents a design to allow Kubernetes to pass metadata such as Pod name and namespace to the CSI `NodePublishVolume` call if a CSI driver requires it.

The detailed design was originally implemented as a [design proposal](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/storage/container-storage-interface-pod-information.md).

This KEP contains details that are missing from the design proposal.

## Test Plan

* Unit tests in kubelet volume manager.
* E2E tests:
    * `CSI workload information [Feature:CSIDriverRegistry]` via CSI mock driver 

## Graduation Criteria

### Alpha to Beta

* Basic unit and e2e tests as outlined in the test plan.

### Beta to GA

* At least one CSI driver implemented using this feature in production.

## Implementation History

* K8s 1.12: Alpha implementation