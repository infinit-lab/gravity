package authorization

import (
	"github.com/infinit-lab/gravity/database"
)

type Model interface {
	GetResource(resourceCode, resourceType string) (*Resource, error)
	GetUser(userCode, userType string) (*User, error)
	GetAuthorization(userCode, userType, resourceCode, resourceType string) (*Authorization, error)
	GetAuthorizationList(userCode, userType string) ([]*Authorization, error)
	GetAuthorizationListByResourceType(userCode, userType, resourceType string) ([]*Authorization, error)
	GetUserListByResource(resourceCode, resourceType string) ([]*User, error)

	CreateUser(userCode, userType string) (userId int, err error)
	CreateResource(userCode, userType, resourceCode, resourceType, parentCode, parentType string, isPublic bool) (resourceId int, err error)
	CreateAuthorization(userCode, userType, resourceCode, resourceType, operations string,
		isHeritable, isUpdatable, isDeletable bool) error

	DeleteUser(userCode, userType string) error
	DeleteResource(resourceCode, resourceType string) error
	DeleteAuthorization(userCode, userType, resourceCode, resourceType string) error
}

func New(db database.Database) (Model, error) {
	return newUserModel(db)
}
