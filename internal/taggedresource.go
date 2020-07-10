package internal

type TaggedResource struct {
	TagID string
	ResourceID string
	Tag Tag
	Resource Resource
}

type TaggedResourceFactory interface {
	AddTagResource(tr TaggedResource) error
	RemoveTagResource(tr TaggedResource) error
}

type TaggedResourceRepository interface {
	FindTaggedResourcesByTagName(name string) ([]TaggedResource, error)
	FindTaggedResourcesByResourceID(id string) ([]TaggedResource, error)
}