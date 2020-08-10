package database

import (
	"errors"
	"fmt"
	"github.com/cayleygraph/cayley"
	"github.com/cayleygraph/quad"
	"github.com/holmes89/tags/internal"
	"github.com/sirupsen/logrus"
)

type repository struct {
	conn *cayley.Handle
	kv map[string]string
}

type Repository interface {
	internal.ResourceFactory
	internal.ResourceRepository
	internal.TagRepository
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

func (r repository) RemoveTagFromResource(resource internal.Resource, tag string) error {
	id := resourceKey(resource.ID)
	tagID := tagKey(tag)
	if err := r.conn.RemoveQuad(quad.Make(id, "has tag", tagID, nil)); err != nil{
		logrus.WithError(err).Error("unable to delete tagged resource")
		return errors.New("unable to add resource tag")
	}
	return nil
}

func (r repository) CreateResource(resource internal.Resource, tag internal.Tag) error {
	id := resourceKeyFromResource(resource)
	tagID := tagKey(tag.Name)
	t := cayley.NewTransaction()

	t.AddQuad(quad.Make(id, "is named", resource.Name, nil))
	t.AddQuad(quad.Make(id, "has tag", tagID, nil))
	if err := r.conn.ApplyTransaction(t); err != nil{
		logrus.WithError(err).Error("unable to write resource")
		return errors.New("unable to add resource")
	}
	return nil
}

func (r repository) CreateTag(tag internal.Tag) error {
	if err := r.conn.AddQuad(quad.Make(tagKey(tag.Name), "has color", tag.Color, nil)); err != nil{
		logrus.WithError(err).Error("unable to write resource")
		return errors.New("unable to add resource")
	}
	return nil
}

func (r repository) FindResourceByTypeAndID(rtype, id string) (internal.Resource, error) {

}

func (r repository) FindResourcesByTag(tag string) ([]internal.Resource, error) {
	panic("implement me")
}

func (r repository) FindAll() ([]internal.Resource, error) {
	key := resourceKey(rtype, id)
	p := cayley.StartPath(r.conn, quad.String(key))
	resource := internal.Resource{
		ID:   id,
		Type: rtype,
	}
	err := p.Clone().Out(quad.String("has name")).Iterate(nil).EachValue(nil, func(value quad.Value) {
		resource.Name = value.String()
	})

	if err != nil {
		logrus.WithError(err).Error("unable to fetch results")
		return resource, errors.New("unable to fetch results")
	}
	return resource, nil
}

func (r repository) FindTagByName(name string) (internal.Tag, error) {
	panic("implement me")
}


func resourceKey(id string) string {
	return fmt.Sprintf("resource:%s", id)
}

func tagKey(name string) string {
	return fmt.Sprintf("tag:%s", name)
}