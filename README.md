# Velocity Application Spec

This repository contains the specification and canonical Go implementions of the core Velocity system. Velocity uses a fairly straightforward control plane / data plane architecture which allows the system as a whole to be distributed in various ways. There are five main components:

## AppSource
The main interface that allows a Velocity data plane instance to learn its state from a Velocity control plane server.

## Directive
Also known as the `velocity.yaml` file, this defines the declarative YAML schema for describing a Velocity application

## Bundle
The format used to package (and unpackage) Velocity applications into a deploy-able archive.

## FQFN
The fully-qualified function name spec is a globally adressable name and URI scheme that makes calling versioned functions easy to reason about.

## Capabilities
Configuration and implementation of the capabilities available to Velocity functions.