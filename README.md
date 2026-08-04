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

### Monitor ECS Service Tasks
To interactively select an ECS cluster and service, and print its active and stopped tasks:
```bash
cx ecs
```
To monitor task states in real-time (auto-refreshing every 5 seconds):
```bash
cx ecs --watch
# Or using the short flag:
cx ecs -w
```

### ECS Caching (1-Step Selection)
To bypass the slow network queries and cluster picker, you can enable persistent local caching.

1. **Enable caching** for the active workspace:
   ```bash
   cx ecs --cache true
   ```
   *Note: If no default cluster is set, you will be prompted to select one from the interactive TUI. This configuration is saved to your `config.yaml` (`ecs.cache: true`, `ecs.default_cluster: "your-cluster"`).*

2. **Run queries instantly**:
   ```bash
   cx ecs
   ```
   *Bypasses the cluster picker and loads the service selection list instantly from `~/.config/cx/ecs_cache.json`.*

3. **Force refresh the cache**:
   If clusters or services change in AWS, force a fresh fetch to rebuild the local cache:
   ```bash
   cx ecs --refresh
   # Or using the short flag:
   cx ecs -r
   ```

4. **Disable caching**:
   ```bash
   cx ecs --cache false
   ```

To manage configuration or caching for other workspaces without switching to them, pass the `--ws` override flag:
```bash
cx ecs --cache true --ws staging
```

### Tail ECS CloudWatch Logs
To tail the CloudWatch logs of a service in real-time, run:
```bash
cx ecs logs [service]
# Or using aliases:
cx ecs cloudwatch [service]
cx ecs cw [service]
```
If the service name is omitted, you will be prompted with the cluster and service selection TUI picker.

#### Log Group Resolution Caching
The first time you select a service, `cx` queries AWS to fetch its active task definition and extract its Log Group and Log Stream prefix. This resolved configuration is cached in `~/.config/cx/ecs_cache.json`. Subsequent logs runs will load the log group from the cache instantly (under 200ms) without making slow AWS API calls.

*To force discovery again (e.g. if the task definition log group changed), run with the refresh flag:*
```bash
cx ecs logs [service] -r
```

#### Custom History & Filtering
*   **Show recent history (default 10m)**:
    ```bash
    cx ecs logs my-service --since 30m
    ```
*   **Filter/Search logs for a pattern**:
    ```bash
    cx ecs logs my-service --filter "ERROR"
    ```