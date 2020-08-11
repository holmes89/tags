package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cayleygraph/cayley"
	"github.com/cayleygraph/quad"
	"github.com/holmes89/tags/internal"
	"github.com/sirupsen/logrus"
	"strings"
)

type repository struct {
	conn *cayley.Handle
	kv map[string][]byte
}


type Repository interface {
	internal.ResourceFactory
	internal.ResourceRepository
	internal.TagRepository
	internal.TagFactory
	internal.ResourceTagger
}

func NewInMemoryRepository() Repository {
	return NewRepository("")
}

func NewRepository(path string) Repository {
	repo := &repository{
		kv : make(map[string][]byte),
	}
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

func (r *repository) writeResource(resource internal.Resource) {
	id := resourceKey(resource.ID)
	e, _ := json.Marshal(resource)
	r.kv[id] = e
}

func (r *repository) writeTag(tag internal.Tag) {
	id := tagKey(tag.Name)
	e, _ := json.Marshal(tag)
	r.kv[id] = e
}

func (r *repository) Delete(resource internal.Resource, tag string) error {
	id := resourceKey(resource.ID)
	resource, err := r.FindByID(resource.ID)
	if err != nil {
		logrus.WithError(err).Error("unable to find resource")
	}
	tagID := tagKey(tag)
	if err := r.conn.RemoveQuad(quad.Make(id, "tag", tagID, nil)); err != nil{
		logrus.WithError(err).Error("unable to delete tagged resource")
		return errors.New("unable to add resource tag")
	}

	// TODO transaction
	if err := r.conn.RemoveQuad(quad.Make(tagID, "resource", id, nil)); err != nil{
		logrus.WithError(err).Error("unable to delete tagged resource")
		return errors.New("unable to add resource tag")
	}

	var tags []internal.Tag
	for _, t := range resource.Tags {
		if t.Name != tag {
			tags = append(tags, t)
		}
	}
	resource.Tags = tags
	r.writeResource(resource)
	return nil
}

func (r *repository) Add(resource internal.Resource, tag string) (internal.Resource, error) {
	id := resourceKey(resource.ID)
	tagID := tagKey(tag)

	t, err := r.FindTagByName(tag)
	if err != nil && err != internal.ErrNotFound {
		return resource, errors.New("unable to find tag")
	}

	if err == internal.ErrNotFound {
		t = internal.Tag{
			Name:  tag,
			Color: internal.GetRandomColor(),
		}
		t, err = r.CreateTag(t)
		if err != nil {
			logrus.WithError(err).Error("unable to create tag")
			return resource, errors.New("unable to create tag")
		}
	}

	if err := r.conn.AddQuad(quad.Make(id, "tag", tagID, nil)); err != nil{
		logrus.WithError(err).Error("unable to delete tagged resource")
		return resource, errors.New("unable to add resource tag")
	}

	// TODO transaction
	if err := r.conn.AddQuad(quad.Make(tagID, "resource", id, nil)); err != nil{
		logrus.WithError(err).Error("unable to delete tagged resource")
		return resource, errors.New("unable to add resource tag")
	}
	resource.Tags = append(resource.Tags, t)
	r.writeResource(resource)

	return resource, nil
}

func (r *repository) FindByID(id string) (internal.Resource, error) {
	var resource internal.Resource
	resourceID := resourceKey(id)
	e, ok := r.kv[resourceID]
	if !ok {
		return resource, internal.ErrNotFound
	}
	if err := json.Unmarshal(e, &resource); err != nil {
		return resource, errors.New("unable to unmarshall resource")
	}

	return resource, nil
}

func (r *repository) FindTagByName(name string) (internal.Tag, error) {
	var tag internal.Tag
	tagID := tagKey(name)
	e, ok := r.kv[tagID]
	if !ok {
		return tag, internal.ErrNotFound
	}
	if err := json.Unmarshal(e, &tag); err != nil {
		return tag, errors.New("unable to unmarshall tag")
	}

	return tag, nil
}

func (r *repository) CreateResource(resource internal.Resource) (internal.Resource, error) {
	id := resourceKey(resource.ID)
	t := cayley.NewTransaction()

	t.AddQuad(quad.Make(id, "name", resource.Name, nil))
	t.AddQuad(quad.Make(id, "type", resource.Type, nil))
	t.AddQuad(quad.Make(resource.Type, "resource", id, nil))
	var tags []internal.Tag
	for _, tag := range resource.Tags {
		tagID := tagKey(tag.Name)
		te, err := r.FindTagByName(tag.Name)
		if err != nil && err != internal.ErrNotFound {
			return resource, errors.New("unable to find tag")
		}

		if err == internal.ErrNotFound {
			te = internal.Tag{
				Name:  tag.Name,
				Color: internal.GetRandomColor(),
			}
			te, err = r.CreateTag(te)
			if err != nil {
				logrus.WithError(err).Error("unable to create tag")
				return resource, errors.New("unable to create tag")
			}
		}
		tags = append(tags, te)
		t.AddQuad(quad.Make(id, "tag", tagID, nil))
		t.AddQuad(quad.Make(tagID, "resource", id, nil))
	}
	resource.Tags = tags
	if err := r.conn.ApplyTransaction(t); err != nil{
		logrus.WithError(err).Error("unable to write resource")
		return resource, errors.New("unable to add resource")
	}

	r.writeResource(resource)
	return resource, nil
}

func (r *repository) CreateTag(tag internal.Tag) (internal.Tag, error) {
	if tag.Color == "" {
		tag.Color = internal.GetRandomColor()
	}
	if err := r.conn.AddQuad(quad.Make(tagKey(tag.Name), "color", tag.Color, nil)); err != nil{
		logrus.WithError(err).Error("unable to write resource")
		return tag, errors.New("unable to add resource")
	}
	r.writeTag(tag)
	return tag, nil
}


func (r *repository) FindAll(params internal.ResourceParams) ([]internal.Resource, error){
	var resources []internal.Resource
	if params.Type != "" {
		logrus.WithField("type", params.Type).Info("searching by type")
		p := cayley.StartPath(r.conn, quad.String(params.Type)).Out(quad.String("resource"))
		var ids []string
		err := p.Iterate(nil).EachValue(nil, func(value quad.Value){
			id := value.String()
			logrus.WithField("id", id).Info("found entity")
			ids = append(ids, id)
		})
		if err != nil {
			logrus.WithError(err).Error("unable to query results")
			return nil, errors.New("unable to query results")
		}
		for _, id := range ids {
			resource, err := r.FindByID(id)
			if err != nil {
				logrus.WithField("id", id).Error("does not exist")
				return nil, errors.New("does not exist")
			}
			resources = append(resources, resource)
		}
		return resources, nil
	}

	for k, v := range r.kv {
		s := strings.Split(k, ":")
		if s[0] != "resource" {
			continue
		}
		var resource internal.Resource
		if err := json.Unmarshal(v, &resource); err != nil {
			return resources, err
		}
		resources = append(resources, resource)
	}
	return resources, nil
}

func (r *repository) FindAllTags(_ internal.TagParams) ([]internal.Tag, error){
	var tags []internal.Tag
	for k, v := range r.kv {
		s := strings.Split(k, ":")
		if s[0] != "tag" {
			continue
		}
		var tag internal.Tag
		if err := json.Unmarshal(v, &tag); err != nil {
			return tags, err
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

func resourceKey(id string) string {
	return fmt.Sprintf("resource:%s", id)
}

func tagKey(name string) string {
	return fmt.Sprintf("tag:%s", name)
}