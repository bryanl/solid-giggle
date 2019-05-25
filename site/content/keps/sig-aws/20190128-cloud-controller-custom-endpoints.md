---
approvers:
- '@justinsb'
authors:
- '@micahhausler'
creation-date: "2019-01-28"
editor: '@micahhausler'
last-updated: "2019-01-28"
owning-sig: sig-aws
participating-sigs:
- sig-cloud-provider
reviewers:
- '@justinsb'
- '@mcrute'
status: provisional
title: Custom endpoints support for AWS Cloud Provider
---
# Custom endpoint support for AWS Cloud Provider

## Table of Contents

* [Summary](#summary)
* [Motivation](#motivation)
    * [Goals](#goals)
    * [Non-Goals](#non-goals)
* [Proposal](#proposal)
    * [Risks and Mitigations](#risks-and-mitigations)
* [Graduation Criteria](#graduation-criteria)
* [Implementation History](#implementation-history)

## Summary

AWS service APIs typically operate at fixed domain name endpoints, but in
certain cases may function at a different endpoint than the AWS SDKs are aware
of. The AWS Cloud Provider should support these custom endpoints.

## Motivation

Being able to support custom endpoints enables Kubernetes users to use alternate
implementations of AWS APIs such as [Eucalyptus][] and alernate AWS endpoints
for AWS Service APIs to support [AWS PrivateLink][]. AWS PrivateLink allows AWS users to
ensure their AWS API calls do not transit the public internet.

[Eucalyptus]: https://www.eucalyptus.cloud/
[AWS PrivateLink]: https://docs.aws.amazon.com/vpc/latest/userguide/vpce-interface.html

### Goals

- Allow Kubernetes Cloud Controller to use custom endpoints for AWS services
- Extend existing CloudConfig INI file to specify endpoints
- Allow Kubelet to use custom endpoints for ECR credential retrieval

### Non-Goals

- Multi-region AWS cloud provider support

## Proposal


## Graduation Criteria

Support for custom endpoints in both the kubelet and cloud controller.

## Implementation History

- Initial CloudController implementation [#72245][]

[#72245]: https://github.com/kubernetes/kubernetes/pull/72245/files