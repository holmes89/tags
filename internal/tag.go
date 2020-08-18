package internal

type Tag struct {
	Name string `json:"name"`
	Color Color `json:"color"`
}

type TagFactory interface {
	CreateTag(tag Tag) (Tag, error)
}

type TagRepository interface {
	FindTagByName(name string) (Tag, error)
	FindAllTags(params *TagParams) ([]Tag, error)
}

type TagParams struct {
	Type string `schema:"type"`
}