package internal

type Resource struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
	Tags []Tag `json:"tags"`
}

type ResourceFactory interface {
	CreateResource(resource Resource) (Resource, error)
}

type ResourceRepository interface {
	FindByID(id string) (Resource, error)
	FindAll(params map[string]string) ([]Resource, error)
}

type ResourceTagger interface {
	Add(resource Resource, tag string) (Resource, error)
	Delete(resource Resource, tag string) error
}