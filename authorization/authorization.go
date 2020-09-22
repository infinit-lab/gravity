package authorization

import (
	"github.com/infinit-lab/gravity/database"
	m "github.com/infinit-lab/gravity/model"
	"github.com/infinit-lab/gravity/printer"
)

type Authorization struct {
	m.Id
	ResourceId   int                   `json:"resourceId" db:"resourceId" db_omit:"update" db_default:"0"`
	UserId       int                   `json:"userId" db:"userId" db_omit:"update" db_default:"0"`
	IsOwner      bool                  `json:"isOwner" db:"isOwner" db_type:"TINYINT" db_default:"0"`
	IsHeritable  bool                  `json:"isHeritable" db:"isHeritable" db_type:"TINYINT" db_default:"0"`
	IsUpdatable  bool                  `json:"isUpdatable" db:"isUpdatable" db_type:"TINYINT" db_default:"0"`
	IsDeletable  bool                  `json:"isDeletable" db:"isDeletable" db_type:"TINYINT" db_default:"0"`
	Operations string `json:"operations" db:"operations" db_type:"VARCHAR(4096)" db_default:"''"`
	Resource    Resource `json:"resource"`
}

type authorizationModel struct {
	model m.Model
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
	a.model.SetBeforeInsertLayer(func(id int, resource interface{}) {
		auth, ok := resource.(*Authorization)
		if !ok {
			return
		}
		value, err := a.resourceModel.model.Get(auth.ResourceId)
		if err != nil {
			printer.Error(err)
			return
		}
		r, ok := value.(*Resource)
		if !ok {
			printer.Error("Not Resource. ")
			return
		}
		auth.Resource = *r
	})
	return a, nil
}

func (a *authorizationModel)getAuthorizationListByUserId(userId int) ([]*Authorization, error) {
	var authorizationList []*Authorization
	values, err := a.model.GetList()
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	for _, value := range values {
		authorization, ok := value.(*Authorization)
		if !ok {
			printer.Error("Not Authorization. ")
			continue
		}
		if authorization.UserId == userId {
			authorizationList = append(authorizationList, authorization)
		}
	}
	return authorizationList, nil
}

func (a *authorizationModel)getUserListByResourceCode(resourceCode string, resourceType string) ([]int, error) {
	resource, err := a.resourceModel.getResource(resourceCode, resourceType)
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	var userList []int
	values, err := a.model.GetList()
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	for _, value := range values {
		authorization, ok := value.(*Authorization)
		if !ok {
			printer.Error("Not Authorization. ")
			continue
		}
		if authorization.ResourceId == resource.GetId() {
			userList = append(userList, authorization.UserId)
		}
	}
	return userList, nil
}
