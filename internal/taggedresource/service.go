package taggedresource

import (
	"errors"
	"github.com/holmes89/tags/internal"
	"github.com/holmes89/tags/internal/database"
	"github.com/sirupsen/logrus"
)

type SearchOptions struct {
	Type string
	ID string
	Tag string
}

type Service interface {
	FindTaggedResources(options *SearchOptions) ([]internal.Resource, error)
	AddTagToResource(resource internal.Resource, tag string) error
	RemoveTagFromResource(resource internal.Resource, tag string) error
}

type service struct {
	repo Repository
}

func NewTaggedResourceService(repo database.Repository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) FindTaggedResources(options *SearchOptions) ([]internal.Resource, error) {
	return nil, nil
}

func (s *service) AddTagToResource(resource internal.Resource, tag string) error {
	t, err := s.repo.FindTagByName(tag)
	if err != nil {
		logrus.WithError(err).Error("unable to find tag")
		return errors.New("unable to find tag")
	}
	if t.Color == "" {
		t.Color = internal.GetRandomColor()
		t.Name = tag
		if err := s.repo.CreateTag(t); err != nil {
			logrus.WithError(err).Error("unable to create tag")
			return errors.New("unable to create tag")
		}
	}
	if err := s.repo.CreateResource(resource, t); err != nil {
		logrus.WithError(err).Error("unable to create tag")
		return errors.New("unable to create tag")
	}
	return nil
}

func (s *service) RemoveTagFromResource(resource internal.Resource, tag string) error {
	return nil
}