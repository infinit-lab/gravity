package authorization

import (
	"errors"
	"github.com/infinit-lab/gravity/database"
	m "github.com/infinit-lab/gravity/model"
	"github.com/infinit-lab/gravity/printer"
)

type Resource struct {
	m.Id
	ResourceCode string `json:"resourceCode" db:"resourceCode" db_omit:"update" db_type:"VARCHAR(64)" db_default:"''"`
	ResourceType string `json:"resourceType" db:"resourceType" db_omit:"update" db_type:"VARCHAR(64)" db_default:"''"`
	ParentId int `json:"parentId" db:"parentId" db_omit:"update" db_default:"0"`
}

type resourceModel struct {
	model m.Model
}

const (
	TopicResource string = "auth_resource"
	resourceTable string = "t_auth_resource"
)

func newResourceModel(db database.Database) (*resourceModel, error) {
	r := new(resourceModel)
	var err error
	if r.model, err = m.New(db, &Resource{}, TopicResource, true, resourceTable); err != nil {
		printer.Error(err)
		return nil, err
	}
	return r, err
}

func (r *resourceModel)getResource(resourceCode string, resourceType string) (*Resource, error) {
	values, err := r.model.GetList()
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	for _, value := range values {
		resource, ok := value.(*Resource)
		if !ok {
			printer.Error("Not Resource. ")
			continue
		}
		if resource.ResourceCode == resourceCode && resource.ResourceType == resourceType {
			return resource, nil
		}
	}
	return nil, errors.New("Not Found. ")
}

func (r *resourceModel)getResourceListByResourceType(resourceType string) ([]*Resource, error) {
	var resourceList []*Resource
	values, err := r.model.GetList()
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	for _, value := range values {
		resource, ok := value.(*Resource)
		if !ok {
			printer.Error("Not Resource. ")
			continue
		}
		if resource.ResourceType == resourceType {
			resourceList = append(resourceList, resource)
		}
	}
	return resourceList, nil
}
