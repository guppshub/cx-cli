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
    bastion_instance_id: i-0d1d909c5fea48c31

    # Catalog of secure resources accessible through the bastion
    resources:
      # Database resources (connect via: cx db <name>)
      databases:
        - name: mercury
          engine: postgres # Supported engines: postgres, mysql
          endpoint: dev-db-cluster.ckkwsfwzdc3c.us-east-1.rds.amazonaws.com
          port: 5432
          local_port: 5432 # Preferred local port to bind to
          # Optional: bastion_instance_id override

      # Redis cache resources (connect via: cx redis <name>)
      redis:
        - name: sequr-cache
          host: dev-cache.dfklpu.0001.use1.cache.amazonaws.com
          port: 6379
          local_port: 6379
          # Optional: bastion_instance_id override
`
