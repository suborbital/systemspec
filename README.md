# DeltaV System Spec

This repository contains the specification and canonical Go implementions of the core DeltaV system. DeltaV uses a control plane / data plane architecture which allows the system as a whole to be distributed in various ways. There are five main components:

## AppSource
The main interface that allows a DeltaV data plane instance to learn its state from a DeltaV control plane server.

## Tenant
Also known as the `tenant.yaml` file or 'tenant config', this defines the declarative schema for describing a DeltaV tenant

## Bundle
The format used to package (and unpackage) DeltaV tenants into a deploy-able archive.

## FQMN
The 'fully-qualified module name' spec is a globally adressable name and URI scheme that makes calling versioned modules easy to reason about.

## Capabilities
Configuration and implementation of the capabilities available to DeltaV modules.