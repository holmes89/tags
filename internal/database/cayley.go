package database

import (
	"errors"
	"fmt"
	"github.com/cayleygraph/cayley"
	"github.com/cayleygraph/quad"
	"github.com/holmes89/tags/internal"
	"github.com/sirupsen/logrus"
	"reflect"
	"strings"
)

type graphdb struct {
	conn *cayley.Handle
}

type GraphDB interface {
	DeleteResourceTag(resource internal.Resource, tag string) error
	AddResourceTag(resource internal.Resource, tag string) error
	CreateResource(resource internal.Resource) error
	CreateTag(tag internal.Tag) error
	FindAllResources(params internal.ResourceParams) ([]string, error)
	FindAllTags(params internal.TagParams) ([]string, error)
}


func NewGraphDatabase() GraphDB {
	conn, err := cayley.NewMemoryGraph()
	if err != nil {
		logrus.WithError(err).Fatal("unable to create graph database")
	}
	logrus.Info("graph database established")

	return &graphdb{conn: conn}
}


func (r *graphdb) DeleteResourceTag(resource internal.Resource, tag string) error {
	id := resourceKey(resource.ID)
	tagID := tagKey(tag)

	trans := cayley.NewTransaction()
	trans.RemoveQuad(quad.Make(id, "tag", tagID, nil))
	trans.RemoveQuad(quad.Make(tagID, "resource", id, nil))

	if err := r.conn.ApplyTransaction(trans); err != nil{
		logrus.WithError(err).Error("unable to delete tagged resource")
		return errors.New("unable to add resource tag")
	}
	return nil
}

func (r *graphdb) AddResourceTag(resource internal.Resource, tag string) error {
	id := resourceKey(resource.ID)
	tagID := tagKey(tag)

	trans := cayley.NewTransaction()

	trans.AddQuad(quad.Make(id, "tag", tagID, nil))
	trans.AddQuad(quad.Make(tagID, "resource", id, nil))

	if err := r.conn.ApplyTransaction(trans); err != nil{
		logrus.WithError(err).Error("unable to delete tagged resource")
		return errors.New("unable to add resource tag")
	}
	return nil
}


func (r *graphdb) CreateResource(resource internal.Resource) error {
	id := resourceKey(resource.ID)
	t := cayley.NewTransaction()

	logrus.WithField("id", id).Info("adding resource")
	t.AddQuad(quad.Make(id, "name", "name:"+resource.Name, nil))
	t.AddQuad(quad.Make(id, "type", "type:"+resource.Type, nil))
	t.AddQuad(quad.Make("type:"+resource.Type, "resource", id, nil))
	for _, tag := range resource.Tags {
		tagID := tagKey(tag.Name)
		t.AddQuad(quad.Make(id, "tag", tagID, nil))
		t.AddQuad(quad.Make(tagID, "resource", id, nil))
	}
	if err := r.conn.ApplyTransaction(t); err != nil{
		logrus.WithError(err).Error("unable to write resource")
		return errors.New("unable to add resource")
	}

	return  nil
}

func (r *graphdb) CreateTag(tag internal.Tag) error {
	id := tagKey(tag.Name)
	logrus.WithField("id", id).Info("adding tag")
	if err := r.conn.AddQuad(quad.Make(id, "color", "color:"+tag.Color, nil)); err != nil{
		logrus.WithError(err).Error("unable to write tag graph")
		return errors.New("unable to add tag")
	}
	return nil
}


func (r *graphdb) FindAllResources(params internal.ResourceParams) ([]string, error){
	return r.findAll("resource", reflect.ValueOf(params))
}



func (r *graphdb) FindAllTags(params internal.TagParams) ([]string, error){
	return r.findAll("tag", reflect.ValueOf(params))
}

func (r *graphdb) findAll(t string, v reflect.Value) ([]string, error){
	var ids []string
	numOfFields := v.NumField()
	for i := 0; i < numOfFields; i++ {
		var tids []string
		path := strings.ToLower(v.Type().Field(i).Name)
		value := v.Field(i).String()
		if value == "" {
			continue
		}
		logrus.WithFields(logrus.Fields{
			"path": path,
			"value": value,
			"type": t,
		}).Info("searching")
		query := fmt.Sprintf("%s:%s", path, value)
		p := cayley.StartPath(r.conn, quad.String(query)).Out(quad.String(t))
		err := p.Iterate(nil).EachValue(nil, func(value quad.Value){
			id := strings.Split((quad.NativeOf(value)).(string), ":")[1]
			tids = append(tids, id)
		})
		if err != nil {
			logrus.WithError(err).Error("unable to find path")
			return nil, errors.New("unable to find results in path")
		}
		logrus.WithField("count", len(tids)).Info("results")
		if ids == nil {
			ids = tids
		} else {
			ids = intersection(ids, tids)
		}
	}
	return ids, nil
}


func resourceKey(id string) string {
	return fmt.Sprintf("resource:%s", id)
}

func tagKey(name string) string {
	return fmt.Sprintf("tag:%s", name)
}

func intersection(a []string, b []string) []string {
	set := make([]string, 0)
	hash := make(map[string]bool)

	for _, v := range a{
		hash[v] = true
	}

	for _, v := range b {
		if _, found := hash[v]; found {
			set = append(set, v)
		}
	}

	return set
}