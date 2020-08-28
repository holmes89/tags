package database

import (
	"errors"
	"github.com/holmes89/tags/internal"
	"github.com/sirupsen/logrus"
)

type Repository interface {
	internal.ResourceFactory
	internal.ResourceRepository
	internal.TagRepository
	internal.TagFactory
	internal.ResourceTagger
}

type repository struct {
	kvstore KVStore
	gdb     GraphDB
}

func NewRepository(kv KVStore, g GraphDB) Repository {
	r := &repository{
		kvstore: kv,
		gdb:     g,
	}
	r.initializeGraphDB()
	return r
}

func (r *repository) CreateResource(resource internal.Resource) (internal.Resource, error) {
	re, err := r.FindResourceByID(resource.ID)
	if err == internal.ErrNotFound {
		err = nil
		if resource.ID == "" || resource.Name == "" || resource.Type == "" {
			return resource, internal.ErrInvalid
		}

		var tags []internal.Tag
		for _, t := range resource.Tags {
			tag, err := r.CreateTag(t)
			if err != nil {
				logrus.WithError(err).Error("unable to create tag")
				return resource, errors.New("failed to save resource")
			}
			tags = append(tags, tag)
		}

		re = internal.Resource{
			ID:   resource.ID,
			Name: resource.Name,
			Type: resource.Type,
			Tags: tags,
		}

		if err := r.kvstore.PutResource(re.ID, re); err != nil {
			logrus.WithError(err).Error("unable to store to kv")
			return re, errors.New("unable to save resource")
		}

		if err := r.gdb.CreateResource(re); err != nil {
			logrus.WithError(err).Error("unable to store to gbd")
			return re, errors.New("unable to save resource")
		}
		return re, nil
	}
	if err != nil {
		logrus.WithError(err).Error("unable to create resource")
		return resource, errors.New("failed to save resource")
	}
	return re, internal.ErrConflict
}

func (r repository) FindResourceByID(id string) (internal.Resource, error) {
	return r.kvstore.GetResource(id)
}

func (r *repository) FindAllResources(params *internal.ResourceParams) ([]internal.Resource, error) {
	if params == nil {
		return r.kvstore.GetAllResources()
	}
	ids, err := r.gdb.FindAllResources(*params)
	if err != nil {
		logrus.WithError(err).Error("unable to find ids")
		return nil, errors.New("unable to find ids")
	}
	var resources []internal.Resource
	for _, id := range ids {
		resource, err := r.kvstore.GetResource(id)
		if err != nil {
			logrus.WithError(err).Error("unable to retrieve id")
			return resources, nil
		}
		resources = append(resources, resource)
	}
	return resources, nil
}

func (r *repository) FindTagByName(name string) (internal.Tag, error) {
	return r.kvstore.GetTag(name)
}

func (r *repository) FindAllTags(params *internal.TagParams) ([]internal.Tag, error) {
	if params == nil {
		return r.kvstore.GetAllTags()
	}
	ids, err := r.gdb.FindAllTags(*params)
	if err != nil {
		logrus.WithError(err).Error("unable to find ids")
		return nil, errors.New("unable to find ids")
	}
	var tags []internal.Tag
	for _, id := range ids {
		tag, err := r.kvstore.GetTag(id)
		if err != nil {
			logrus.WithError(err).Error("unable to retrieve id")
			return tags, nil
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

func (r *repository) CreateTag(tag internal.Tag) (internal.Tag, error) {
	t, err := r.FindTagByName(tag.Name)
	if err == internal.ErrNotFound {
		t = internal.Tag{
			Name:  tag.Name,
			Color: internal.GetRandomColor(),
		}
		if err := r.kvstore.PutTag(t.Name, t); err != nil {
			logrus.WithError(err).Error("unable to save tag kv")
			return tag, errors.New("not able to save tag")
		}
		if err := r.gdb.CreateTag(t); err != nil {
			logrus.WithError(err).Error("unable to save tag graph")
			return tag, errors.New("not able to save tag")
		}
		return t, nil
	}
	if err != nil {
		logrus.WithError(err).Error("unable to find tag")
		return tag, errors.New("not able to find tag")
	}
	return t, nil
}

func (r *repository) AddTagToResource(resource internal.Resource, tag string) (internal.Resource, error) {
	t, err := r.CreateTag(internal.Tag{Name: tag})
	if err != nil {
		logrus.WithError(err).Error("unable to find tag")
		return resource, errors.New("unable to find tag")
	}

	resource, err = r.FindResourceByID(resource.ID)
	if err != nil {
		logrus.WithError(err).Error("unable to find resource")
		return resource, errors.New("unable to find resource")
	}

	tags := []internal.Tag{t}
	for _, tg := range resource.Tags {
		if tg.Name != t.Name {
			tags = append(tags, tg)
		}
	}

	resource.Tags = tags
	if err := r.kvstore.PutResource(resource.ID, resource); err != nil {
		logrus.WithError(err).Error("unable to save resource")
		return resource, errors.New("unable to save resource")
	}

	if err := r.gdb.AddResourceTag(resource, tag); err != nil {
		logrus.WithError(err).Error("unable to save resource graph")
		return resource, errors.New("unable to save resource")
	}

	return resource, nil
}

func (r *repository) DeleteTagFromResource(resource internal.Resource, tag string) error {

	resource, err := r.FindResourceByID(resource.ID)
	if err != nil {
		logrus.WithError(err).Error("unable to find resource")
		return errors.New("unable to find resource")
	}

	var tags []internal.Tag
	for _, tg := range resource.Tags {
		if tg.Name != tag {
			tags = append(tags, tg)
		}
	}

	resource.Tags = tags
	if err := r.kvstore.PutResource(resource.ID, resource); err != nil {
		logrus.WithError(err).Error("unable to save resource kv")
		return errors.New("unable to save resource")
	}

	if err := r.gdb.DeleteResourceTag(resource, tag); err != nil {
		logrus.WithError(err).Error("unable to save resource graph")
		return errors.New("unable to save resource")
	}

	return nil
}

func (r *repository) initializeGraphDB() {
	logrus.Info("initializing graph database")
	tags, err := r.kvstore.GetAllTags()
	if err != nil {
		logrus.WithError(err).Fatal("unable to load tags in graph db")
	}
	for _, tag := range tags {
		if err != r.gdb.CreateTag(tag) {
			logrus.WithError(err).Fatal("unable to load tags in graph db")
		}
	}
	resources, err := r.kvstore.GetAllResources()
	if err != nil {
		logrus.WithError(err).Fatal("unable to load resources in graph db")
	}
	for _, resource := range resources {
		if err != r.gdb.CreateResource(resource) {
			logrus.WithError(err).Fatal("unable to load resources in graph db")
		}
	}
	logrus.Info("initializing graph database complete")
}
