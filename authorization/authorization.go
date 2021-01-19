package authorization

import (
	"github.com/infinit-lab/gravity/database"
	m "github.com/infinit-lab/gravity/model"
	"github.com/infinit-lab/gravity/printer"
)

type Authorization struct {
	m.PrimaryKey
	ResourceCode  string      `json:"resourceCode" db:"resourceCode" db_type:"VARCHAR(64)" db_index:"index" db_omit:"update" db_default:"''"`
	UserCode      string      `json:"userCode" db:"userCode" db_type:"VARCHAR(64)" db_index:"index" db_omit:"update" db_default:"''"`
	IsOwner     bool     `json:"isOwner" db:"isOwner" db_omit:"update" db_default:"0"`
	IsHeritable bool     `json:"isHeritable" db:"isHeritable" db_default:"0"`
	IsUpdatable bool     `json:"isUpdatable" db:"isUpdatable" db_default:"0"`
	IsDeletable bool     `json:"isDeletable" db:"isDeletable" db_default:"0"`
	Operations  string   `json:"operations" db:"operations" db_type:"VARCHAR(4096)" db_default:"''"`
	Resource    Resource `json:"resource" db_skip:"true"`
}

type authorizationModel struct {
	model         m.Model
	resourceModel *resourceModel
}

const (
	TopicAuthorization string = "auth_authorization"
	authorizationTable string = "t_auth_authorization"
)

func newAuthorizationModel(db database.Database) (*authorizationModel, error) {
	a := new(authorizationModel)
	var err error
	if a.model, err = m.New(db, &Authorization{}, TopicAuthorization, true, authorizationTable); err != nil {
		printer.Error(err)
		return nil, err
	}
	if a.resourceModel, err = newResourceModel(db); err != nil {
		printer.Error(err)
		return nil, err
	}
	a.model.SetBeforeGetLayer(func(resource interface{}) {
		auth, ok := resource.(*Authorization)
		if !ok {
			return
		}
		value, err := a.resourceModel.model.GetByCode(auth.ResourceCode)
		if err != nil {
			printer.Error(err)
			return
		}
		r := value.(*Resource)
		auth.Resource = *r
	})
	return a, nil
}

func (a *authorizationModel) getAuthorizationListByUserCode(userCode string) ([]*Authorization, error) {
	var authorizationList []*Authorization
	values, err := a.model.GetList("WHERE `userCode` = ?", userCode)
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	for _, value := range values {
		authorization := value.(*Authorization)
		authorizationList = append(authorizationList, authorization)
	}
	return authorizationList, nil
}

func (a *authorizationModel) getUserListByResourceCode(resourceCode string, resourceType string) ([]string, error) {
	resource, err := a.resourceModel.getResource(resourceCode, resourceType)
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	var userList []string
	values, err := a.model.GetList("WHERE `resourceCode` = ?", resource.GetCode())
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	for _, value := range values {
		authorization := value.(*Authorization)
		userList = append(userList, authorization.UserCode)
	}
	return userList, nil
}
