# PRD: cx - Workflow-Oriented AWS Operations CLI

## Status

Draft (v0.1)

---

# Vision

cx is an open-source, cross-platform CLI that simplifies repetitive cloud operational workflows.

Instead of exposing low-level cloud APIs, cx exposes high-level workflows and provides an interactive, discoverable experience for engineers.

The first version focuses on AWS, while the architecture should allow additional cloud providers in the future.

---

# Problem Statement

Cloud engineers perform dozens of repetitive operational tasks every day.

Typical examples include:

- Connecting to RDS
- Connecting to Redis
- Opening an OpenSearch tunnel
- SSH'ing into EC2 instances
- Checking ECS task status
- Tailing CloudWatch logs

Each workflow requires remembering:

- AWS profiles
- regions
- cluster names
- instance IDs
- endpoints
- local ports
- long AWS CLI commands

These operations consume cognitive effort without providing business value.

The AWS CLI is API-oriented.

cx aims to be workflow-oriented.

---

# Goals

Primary Goals

- Reduce operational cognitive load
- Minimize memorization
- Provide discoverable workflows
- Make common operational tasks one command away

Secondary Goals

- Open source
- Cross platform
- Extensible architecture
- Configuration driven

---

# Non Goals (v0.1)

- Azure support
- GCP support
- Kubernetes support
- AI integrations
- Deployment automation
- Terraform integration
- Infrastructure provisioning

---

# Design Principles

## Workflow over APIs

Users should think about what they want to do, not which AWS command performs it.

Bad

aws ssm start-session ...

Good

cx db

---

## Discoverability over Memorization

Whenever possible, cx should discover resources instead of requiring users to remember names.

Examples

One configured database

cx db

↓

Automatically connects.

Multiple configured databases

cx db

↓

Interactive picker.

---

## Human First

Interactive by default.

Automation friendly.

---

## Context Aware

Users select a context once.

Example

cx use staging

Subsequent commands inherit the current context.

Every command can override using

--context

---

## Configuration Driven

No company-specific logic should exist inside the application.

All infrastructure should come from configuration.

---

## Provider Extensible

The initial release focuses on AWS.

The architecture must allow future providers.

---

# Supported Resources (v0.1)

## SQL Databases (Amazon RDS)

Capabilities

- Open database tunnel
- Reuse existing tunnel
- Optional persistent tunnel
- Automatic reconnect
- MFA authentication
- STS authentication
- Local port forwarding

---

## Redis (Amazon ElastiCache)

Capabilities

- Open Redis tunnel
- Reuse connection
- Optional persistence
- Automatic reconnect

---

## OpenSearch

Capabilities

- Tunnel
- Reuse connection
- Optional persistence

---

## EC2

Capabilities

- Interactive SSM session

Default startup command

sudo su - ubuntu

Support configurable startup commands.

---

## ECS

Capabilities

- View clusters
- View services
- View running tasks
- Task status
- Desired vs Running
- Health

Future

Shell into ECS tasks

---

## CloudWatch

Capabilities

- Tail logs
- Follow logs
- Search logs
- Filter by time
- Interactive service selection

---

# Context Model

A Context represents an operational environment.

Examples

- development
- staging
- production

Each context owns its own resources.

---

# Resource Discovery

Users should rarely need to remember resource names.

Resolution strategy

If resource specified

↓

Use it

Else

↓

Number of configured resources

0

↓

Helpful error

1

↓

Automatically select

>1

↓

Interactive picker

---

# Connection Management

Connection management is a first-class feature.

Users should not manually manage tunnels.

Connections should support

- reuse
- status
- automatic reconnect
- optional persistence
- cleanup

The implementation mechanism is abstracted.

Initially this may use tmux.

Future versions may use platform-native background services.

---

# Commands

## Context

cx use

Switch active context.

Examples

cx use staging

cx use production

---

cx current

Display current context.

---

## Database

Canonical command

cx db

Behavior

One database

↓

Automatically connect.

Multiple databases

↓

Interactive picker.

Explicit

cx db mercury

---

## Cache

Canonical command

cx cache

Behavior identical to database.

---

## Search

Canonical command

cx search

Behavior identical to database.

---

## Compute

Canonical command

cx compute

Behavior

Interactive EC2 selection.

Starts SSM session.

Supports configurable startup command.

---

## Services

Canonical command

cx service

Displays ECS status.

Interactive selection.

---

## Logs

Canonical command

cx logs

Displays CloudWatch logs.

Interactive service selection.

---

## Resource Discovery

cx ls

Displays available resources inside the current context.

Example

Context: staging

Databases

● mercury
○ auth

Caches

● redis

Search

○ opensearch

Compute

○ bastion

Services

○ authorization-api

Legend

● Connected

○ Available

---

## Status

cx status

Displays

- Current context
- AWS profile
- Region
- STS status
- Active connections

---

# Canonical Commands vs Provider Aliases

cx defines provider-independent canonical commands.

Examples

compute

service

db

cache

search

Providers may expose native aliases.

AWS examples

ec2

ecs

rds

Future providers may expose their own aliases.

Azure

vm

aks

GCP

gce

gke

Aliases should behave identically to their canonical commands.

Documentation should primarily reference canonical commands.

---

# Configuration

Configuration should follow XDG.

Configuration

~/.config/cx/config.yaml

Runtime state

~/.local/state/cx/

Runtime state includes

- current context
- active connections

---

# Success Criteria

A backend engineer should no longer need to remember

- instance IDs
- RDS endpoints
- Redis endpoints
- OpenSearch endpoints
- cluster names
- local ports
- SSM commands

Instead, they should rely on

- context
- interactive discovery
- reusable workflows

---

# Roadmap

### Phase 1 — Foundation ✅
- [x] 001 Configuration Foundation
- [x] 002 Workspace Management

### Phase 2 — Cloud Model
- [ ] 003 Provider Framework
- [ ] 004 Resource Catalog

### Phase 3 — User Experience
- [ ] 005 cx init
- [ ] 006 doctor
- [ ] 007 completion
- [ ] 008 config edit

### Phase 4 — Workflows ⭐
- [ ] 009 db
- [ ] 010 cache
- [ ] 011 search
- [ ] 012 compute
- [ ] 013 logs
- [ ] 014 service

---

# Future Roadmap

Future versions may include:
- Azure
- GCP
- Kubernetes
- Docker
- GitHub
- AI-powered operational workflows
- Plugin system
- Connection dashboard
- Interactive TUI