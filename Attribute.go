package main

// Attribute represent a k8s attribute
type Attribute struct {
	Name      string
	Value     interface{}
	Flagged   bool
	Default   bool
	Set       bool
	Reason    string
	Container string
}

// NewAttribute creates new attribute and set defaults
func NewAttribute(name string, value interface{}, container string) *Attribute {
	a := Attribute{
		Name:      name,
		Default:   true,
		Flagged:   false,
		Set:       false,
		Container: container,
	}

	if container == "" {
		a.Container = "[Global Attribute - ALL]"
	}

	switch value.(type) {
	case (*int64):
		if value != (*int64)(nil) {
			a.Value = *(value.(*int64))
			a.Set = true
			a.Default = false
		}
	case (*bool):
		if value != (*bool)(nil) {
			a.Value = *(value.(*bool))
			a.Set = true
			a.Default = false
		}
	case (bool):
		a.Value = value
		a.Set = true
		a.Default = false
	case (string):
		if value != nil {
			if value != "" {
				a.Value = value
				a.Set = true
				a.Default = false
			}
		}
	default:
		if value != nil {
			a.Value = value
			a.Set = true
			a.Default = false
		}
	}

	if a.Set != true {
		a.Flagged = true
		a.Value = value
	}

	return &a
}
