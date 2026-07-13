# Research: EC2 SSM Shell Integration & TUI Picker Design

This document details the command query formats, interactive execution strategies, and Bubble Tea model structures required to implement the TUI Resource Picker and EC2 SSH workflow.

---

## 1. AWS CLI EC2 Instance Discovery

To query all EC2 instances from the AWS CLI in a single structured JSON block, we run:

```bash
aws ec2 describe-instances \
  --query "Reservations[*].Instances[*].{Name:Tags[?Key=='Name'].Value|[0],InstanceId:InstanceId,State:State.Name,PrivateIpAddress:PrivateIpAddress}" \
  --output json
```

### Example Output:
```json
[
  [
    {
      "Name": "bastion-prod",
      "InstanceId": "i-0123456789abcdef0",
      "State": "running",
      "PrivateIpAddress": "10.0.1.25"
    },
    {
      "Name": "payment-worker",
      "InstanceId": "i-09f87c4f1c901844a",
      "State": "running",
      "PrivateIpAddress": "10.0.1.86"
    }
  ]
]
```

### Go Parser Mapping:
Because of the nested structure of Reservations and Instances, `json.Unmarshal` must decode into a two-dimensional slice `[][]EC2Instance` and then flatten it:

```go
type EC2Instance struct {
	Name             string `json:"Name"`
	InstanceID       string `json:"InstanceId"`
	State            string `json:"State"`
	PrivateIPAddress string `json:"PrivateIpAddress"`
}
```

---

## 2. Interactive SSM Shell Session Handoff

SSM Session Manager is fully interactive. When launching the connection, we cannot redirect stdout to buffers. We must bind the parent Go process's standard streams directly:

```go
func ConnectSSM(instanceID string, profile, region string) error {
	args := []string{"ssm", "start-session", "--target", instanceID}
	if profile != "" {
		args = append(args, "--profile", profile)
	}
	if region != "" {
		args = append(args, "--region", region)
	}

	cmd := exec.Command("aws", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
```

---

## 3. Reusable TUI Picker Bubble Tea Model

We will build the picker using Charm's `bubbletea` and `textinput` components. 

### Model Structure:
```go
type PickerModel struct {
	Title      string
	Items      []Row
	Filtered   []int
	Cursor     int
	SearchInput textinput.Model
	Selected   int // Index of selected item, -1 if cancelled
}
```

### Custom Rendering:
To render column-aligned rows cleanly:
1. Scan all items to calculate the maximum length of each column.
2. Format each field with spacing to match the column width.
3. Apply Lip Gloss styles (like bolding the selected item, coloring state names).
