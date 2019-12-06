package main

// Record represents a K8s object and all it's attributes
type Record struct {
	Cluster    string
	Ns         string
	Kind       string
	KindName   string
	Attributes []Attribute
	Global     map[string]interface{}
}
