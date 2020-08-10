package internal

type Tag struct {
	Name string `json:"name"`
	Color Color `json:"color"`
}

type TagFactory interface {
	CreateTag(tag Tag) error
}

type TagRepository interface {
	FindTagByName(name string) (Tag, error)
}