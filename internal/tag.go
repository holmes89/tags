package internal

type Tag struct {
	ID string
	Color Color
}

type TagFactory interface {
	CreateTag(tag Tag) error
}