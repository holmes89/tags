package taggedresource

import "github.com/holmes89/tags/internal"

type Repository interface {
	FindTagByName(name string) (internal.Tag, error)
	CreateTag(tag internal.Tag) error

}
