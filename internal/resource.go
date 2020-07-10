package internal

type Resource struct {
	ID string
	Name string
	Type string
}

type ResourceFactory interface {
	CreateResource(resource Resource) error
}