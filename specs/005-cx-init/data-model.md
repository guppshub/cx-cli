# Data Model: Configuration Schema

This document details the schema rules for `config.yaml` created by `cx init`.

---

## 1. Schema Validation Rules

Every configuration file must validate against the following properties:

| KeyPath | Type | Required | Rules / Description |
|---|---|---|---|
| `version` | string | Yes | Must be exactly `"1"`. Checked by the configuration loader. |
| `current` | string | Yes | Active workspace name. Must match one of the keys in `workspaces`. |
| `workspaces` | map | Yes | A dictionary of workspace contexts. |
| `workspaces.<name>.provider` | string | Yes | Currently must be `"aws"`. |
| `workspaces.<name>.profile` | string | Yes | AWS CLI credential profile name. |
| `workspaces.<name>.region` | string | Yes | AWS target region (e.g. `us-east-1`). |
| `workspaces.<name>.bastion_instance_id` | string | No | Default EC2 Instance ID used for SSH/SSM tunneling. |
| `workspaces.<name>.resources` | map | No | Namespace for secure catalog resources. |
| `workspaces.<name>.resources.databases` | array | No | List of target database configurations. |
| `workspaces.<name>.resources.redis` | array | No | List of target Redis cache configurations. |

---

## 2. Resource Fields

### Database Resource
* **`name`**: Unique string identifier for the resource.
* **`engine`**: Database driver type (currently `postgres` or `mysql`).
* **`endpoint`**: Remote hostname of the database server.
* **`port`**: Connection port on the remote host (e.g., 5432, 3306).
* **`local_port`**: Preferred local port on loopback interface (e.g., 5432).
* **`bastion_instance_id`**: (Optional) Override for the bastion host instance.

### Redis Resource
* **`name`**: Unique string identifier for the resource.
* **`host`**: Remote hostname of the Redis cache cluster.
* **`port`**: Connection port on the remote host (usually 6379).
* **`local_port`**: Preferred local port on loopback interface (usually 6379).
* **`bastion_instance_id`**: (Optional) Override for the bastion host instance.
