package config

// DefaultConfigTemplate is the starting YAML template written by cx init.
const DefaultConfigTemplate = `# cx-cli Configuration File
# For documentation and help, visit https://github.com/guppshub/cx-cli

# Schema version of this configuration. Currently must be "1".
version: "1"

# The name of the active workspace context.
# Switch workspaces using: cx use <workspace>
current: dev

# Configured workspace environments (e.g., dev, staging, production)
workspaces:
  dev:
    # Target cloud provider (currently only "aws" is supported)
    provider: aws

    # AWS credential profile name from ~/.aws/credentials
    profile: default

    # AWS target region containing the resources
    region: us-east-1

    # Optional: Default Bastion host instance ID used for tunneling if not specified per-resource
    bastion_instance_id: i-0123456789abcdef0

    # Catalog of secure resources accessible through the bastion
    resources:
      # Database resources (connect via: cx db <name>)
      databases:
        - name: sample-postgres
          engine: postgres # Supported engines: postgres, mysql
          endpoint: postgres.example.com
          port: 5432
          local_port: 5432 # Preferred local port to bind to
          # Optional: bastion_instance_id override

      # Redis cache resources (connect via: cx redis <name>)
      redis:
        - name: sample-redis
          host: redis.example.com
          port: 6379
          local_port: 6379
          # Optional: bastion_instance_id override
`
