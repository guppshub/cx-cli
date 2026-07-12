# Data Model: Picker Item & EC2 Instance

This document describes the structs and interfaces used to represent pickable rows and EC2 resource models.

---

## 1. Picker Interface & Row Model

The `picker` package operates on a generic `Row` interface or structure to decouple the UI from specific cloud resources.

### Picker Row Struct:
```go
package picker

// Row represents a single formatted entry in the picker menu.
type Row struct {
	ID     string   // Unique ID returned upon selection (e.g. InstanceId)
	Fields []string // Ordered list of fields to display as columns (e.g. Name, ID, State, IP)
}
```

---

## 2. EC2 Instance Domain Struct

When fetching instances from AWS, we deserialize the raw output into a concrete domain struct:

```go
package aws

type EC2Instance struct {
	Name             string `json:"Name"`
	InstanceID       string `json:"InstanceId"`
	State            string `json:"State"`
	PrivateIPAddress string `json:"PrivateIpAddress"`
}
```

### Mapping EC2 to Picker Row:
To feed EC2 instances into the picker, the list of `EC2Instance` is transformed into `picker.Row` objects:

```go
row := picker.Row{
	ID: instance.InstanceID,
	Fields: []string{
		instance.Name,
		instance.InstanceID,
		instance.State,
		instance.PrivateIPAddress,
	},
}
```
