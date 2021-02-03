package authorization

import (
	"github.com/infinit-lab/gravity/database"
	m "github.com/infinit-lab/gravity/model"
	"github.com/infinit-lab/gravity/printer"
)

type Resource struct {
	m.PrimaryKey
	ResourceCode string `json:"resourceCode" db:"resourceCode" db_index:"index" db_omit:"update" db_type:"VARCHAR(64)" db_default:"''"`
	ResourceType string `json:"resourceType" db:"resourceType" db_index:"index" db_omit:"update" db_type:"VARCHAR(64)" db_default:"''"`
	ParentCode   string `json:"parentCode" db:"parentCode" db_index:"index" db_omit:"update" db_type:"VARCHAR(64)" db_default:"''"`
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

func (r *resourceModel) getResource(resourceCode string, resourceType string) (*Resource, error) {
	value, err := r.model.Get("WHERE `resourceCode` = ? AND `resourceType` = ?", resourceCode, resourceType)
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	return value.(*Resource), nil
}

func (r *resourceModel) getResourceListByResourceType(resourceType string) ([]*Resource, error) {
	values, err := r.model.GetList("WHERE `resourceType` = ?", resourceType)
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	var resourceList []*Resource
	for _, value := range values {
		resource := value.(*Resource)
		resourceList = append(resourceList, resource)
	}
	return resourceList, nil
}
