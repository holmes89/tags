package internal

type Resource struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
	Tags []Tag `json:"tags"`
}

type ResourceFactory interface {
	CreateResource(resource Resource, tag Tag) error
}

type ResourceRepository interface {
	FindResourceByTypeAndID(rtype, id string) (Resource, error)
	FindResourcesByTag(tag string) ([]Resource, error)
	FindAll() ([]Resource, error)
}