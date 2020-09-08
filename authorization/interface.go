package authorization

import (
	"github.com/infinit-lab/gravity/database"
	m "github.com/infinit-lab/gravity/model"
	"github.com/infinit-lab/gravity/printer"
)

type Model interface {
	GetResource(resourceId int) (*Resource, error)
	GetUserInfo(userId int) (*UserInfo, error)
	GetAuthorization(userId int, resourceId int) (*Authorization, error)

	CreateUser(relativeList []int, userType string) (userId int, err error)
	CreateResource(resourceType string, parentList, relativeList []int, isPublic bool) (resourceId int, err error)
	CreateAuthorization(userId, resourceId int, operationList []string, isHeritable, isUpdatable, isDeletable bool) error

	DeleteUser(userId int) error
	DeleteResource(resourceId int) error
	DeleteAuthorization(userId, resourceId int) error
}

func New(db database.Database) (Model, error) {
	resourceModel, err := m.New(db, &Resource{}, "", true, "t_auth_resource")
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	resourceParentModel, err := m.New(db, &ResourceParent{}, "", true, "t_auth_resource_parent")
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	resourceRelativeModel, err := m.New(db, &ResourceRelative{}, "", true, "t_auth_resource_relative")
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	authorizationModel, err := m.New(db, &Authorization{}, "", true, "t_auth_authorization")
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	authorizationOperationModel, err := m.New(db, &Operation{}, "", true, "t_auth_authorization_operation")
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	userModel, err := m.New(db, &User{}, "", true, "t_auth_user")
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	userRelativeModel, err := m.New(db, &UserRelative{}, "", true, "t_auth_user_relative")
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	model := new(model)
	model.resourceModel = resourceModel
	model.resourceParentModel = resourceParentModel
	model.resourceRelativeModel = resourceRelativeModel
	model.authorizationModel = authorizationModel
	model.authorizationOperationModel = authorizationOperationModel
	model.userModel = userModel
	model.userRelativeModel = userRelativeModel
	model.resourceModel.SetBeforeInsertLayer(func(id int, value interface{}) {
		resource, ok := value.(*Resource)
		if !ok {
			return
		}
		model.beforeInsertResourceLayer(id, resource)
	})
	model.authorizationModel.SetBeforeInsertLayer(func(id int, value interface{}) {
		authorization, ok := value.(*Authorization)
		if !ok {
			return
		}
		model.beforeInsertAuthorizationLayer(id, authorization)
	})
	model.userModel.SetBeforeInsertLayer(func(id int, value interface{}) {
		user, ok := value.(*User)
		if !ok {
			return
		}
		model.beforeInsertUserLayer(id, user)
	})
	return model, nil
}
