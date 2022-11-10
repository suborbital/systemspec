# Suborbital System Spec

This repository contains the specification and canonical Go implementions of the core Suborbital system. Suborbital uses a control plane (SE2) / data plane (E2Core) architecture which allows the system as a whole to be distributed in various ways. There are five main components:

## System
The main interface that allows a data plane instance to learn its state from a control plane server.

## Tenant
Also known as the `tenant.json` file or 'tenant config', this defines the declarative schema for describing a Suborbital tenant

## Bundle
The format used to package (and unpackage) tenants into a deploy-able archive.

## FQMN
The 'fully-qualified module name' spec is a globally adressable name and URI scheme that makes executing versioned modules easy.

## Capabilities
Configuration and implementation of the capabilities available to Suborbital modules.