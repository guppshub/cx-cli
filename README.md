# cx-cli

A workflow-oriented cloud operations CLI designed to simplify repetitive developer tasks.

## Roadmap

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

### Phase 5 — Release & Packaging
- [ ] Migrate repository to a dedicated GitHub Organization (e.g., `github.com/cx-cli`) to standardize import paths and decouple them from personal GitHub accounts.

## Usage

### Check CLI Version
To display the installed version and build metadata:
```bash
cx version
# Or use the version flags on the root command:
cx -v
cx --version
```

### Update CLI to Latest Release
To check if a new version is available:
```bash
cx update --check
```
To perform an interactive update:
```bash
cx update
```
To perform a non-interactive update (e.g., in CI or automation scripts):
```bash
cx update --yes
```

### Connect to EC2 via SSM
To list and connect to an EC2 instance in the active workspace:
```bash
cx ec2
```
To execute a specific command upon starting the interactive session (e.g., automatically switching to user `ubuntu` and changing directory to their home folder):
```bash
cx ec2 -c "sudo su - ubuntu"
# Or using the long flag:
cx ec2 --command "sudo su - ubuntu"
```

Alternatively, you can define a default startup command for the active workspace in your `~/.config/cx/config.yaml` file so you don't have to specify it on every command run:
```yaml
workspaces:
  dev:
    provider: aws
    profile: dev-profile
    region: us-east-1
    ec2_startup_command: "sudo su - ubuntu"
```