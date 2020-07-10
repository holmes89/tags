package database

import (
	"errors"
	"github.com/cayleygraph/cayley"
	"github.com/cayleygraph/quad"
	"github.com/holmes89/tags/internal"
	"github.com/sirupsen/logrus"
)

type repository struct {
	conn *cayley.Handle
}



type Repository interface {
	internal.TaggedResourceFactory
	internal.ResourceFactory
	internal.TagFactory
}

func NewRepository(path string) Repository {
	repo := &repository{}
	var err error
	if path == "" {
		logrus.Warn("using in memory graph")
		repo.conn, err = cayley.NewMemoryGraph()
		if err != nil {
			logrus.WithError(err).Fatal("unable to create connection")
		}
	} else {
		logrus.WithField("path", path).Info("connecting to database")
		repo.conn, err = cayley.NewGraph("tagdb", path, nil)
	}

	return repo
}

func (r repository) AddTagResource(tr internal.TaggedResource) error {
	if err := r.conn.AddQuad(quad.Make(tr.ResourceID, "has tag", tr.TagID, nil)); err != nil{
		logrus.WithError(err).Error("unable to write tag")
		return errors.New("unable to add resource tag")
	}
	return nil
}

func (r repository) RemoveTagResource(tr internal.TaggedResource) error {
	if err := r.conn.RemoveQuad(quad.Make(tr.ResourceID, "has tag", tr.TagID, nil)); err != nil{
		logrus.WithError(err).Error("unable to delete tagged resource")
		return errors.New("unable to add resource tag")
	}
	return nil
}

func (r repository) CreateResource(resource internal.Resource) error {
	if err := r.conn.AddQuad(quad.Make(resource.ID, "is named", resource.Name, nil)); err != nil{
		logrus.WithError(err).Error("unable to write resource")
		return errors.New("unable to add resource")
	}
	return nil
}

func (r repository) CreateTag(tag internal.Tag) error {
	if err := r.conn.AddQuad(quad.Make(tag.ID, "has color", tag.Color, nil)); err != nil{
		logrus.WithError(err).Error("unable to write resource")
		return errors.New("unable to add resource")
	}
	return nil
}
