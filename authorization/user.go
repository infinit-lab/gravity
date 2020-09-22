package authorization

import (
	"errors"
	"github.com/infinit-lab/gravity/database"
	m "github.com/infinit-lab/gravity/model"
	"github.com/infinit-lab/gravity/printer"
)

type User struct {
	m.PrimaryKey
	UserCode string `json:"userCode" db:"userCode" db_omit:"update" db_default:"0"`
	UserType string `json:"userType" db:"userType" db_omit:"update" db_default:"0"`
}

type UserWithAuthorization struct {
	User
	AuthorizationMap map[int]*Authorization `json:"authorizationMap"`
}

type userModel struct {
	model              m.Model
	authorizationModel *authorizationModel
}

const (
	TopicUser string = "auth_user"
	userTable string = "t_auth_user"
)

func newUserModel(db database.Database) (*userModel, error) {
	u := new(userModel)
	var err error
	if u.model, err = m.New(db, &UserWithAuthorization{}, TopicUser, true, userTable); err != nil {
		printer.Error(err)
		return nil, err
	}
	if u.authorizationModel, err = newAuthorizationModel(db); err != nil {
		printer.Error(err)
		return nil, err
	}
	u.model.SetBeforeInsertLayer(func(id int, resource interface{}) {
		user, ok := resource.(*UserWithAuthorization)
		if !ok {
			return
		}
		list, err := u.authorizationModel.getAuthorizationListByUserId(id)
		if err != nil {
			printer.Error(err)
			return
		}
		user.AuthorizationMap = make(map[int]*Authorization)
		for _, a := range list {
			user.AuthorizationMap[a.ResourceId] = a
		}
	})
	return u, nil
}

func (u *userModel) GetResource(resourceCode, resourceType string) (*Resource, error) {
	return u.authorizationModel.resourceModel.getResource(resourceCode, resourceType)
}

func (u *userModel) GetUserWithAuthorization(userCode, userType string) (*UserWithAuthorization, error) {
	values, err := u.model.GetList()
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	for _, value := range values {
		user, ok := value.(*UserWithAuthorization)
		if !ok {
			continue
		}
		if user.UserCode == userCode && user.UserType == userType {
			return user, nil
		}
	}
	return nil, errors.New("Not Found. ")
}

func (u *userModel) GetUser(userCode, userType string) (*User, error) {
	user, err := u.GetUserWithAuthorization(userCode, userType)
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	return &user.User, nil
}

func (u *userModel) GetAuthorization(userCode, userType, resourceCode, resourceType string) (*Authorization, error) {
	user, err := u.GetUserWithAuthorization(userCode, userType)
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	resource, err := u.GetResource(resourceCode, resourceType)
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	a, ok := user.AuthorizationMap[resource.GetId()]
	if !ok {
		return nil, errors.New("Unauthorized. ")
	}
	return a, nil
}

func (u *userModel) GetAuthorizationList(userCode, userType string) ([]*Authorization, error) {
	user, err := u.GetUserWithAuthorization(userCode, userType)
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	var authList []*Authorization
	for _, a := range user.AuthorizationMap {
		authList = append(authList, a)
	}
	return authList, nil
}

func (u *userModel) GetAuthorizationListByResourceType(userCode, userType, resourceType string) ([]*Authorization, error) {
	user, err := u.GetUserWithAuthorization(userCode, userType)
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	var authList []*Authorization
	for _, a := range user.AuthorizationMap {
		if a.Resource.ResourceType == resourceType {
			authList = append(authList, a)
		}
	}
	return authList, nil
}

func (u *userModel) GetUserListByResource(resourceCode, resourceType string) ([]*User, error) {
	userIdList, err := u.authorizationModel.getUserListByResourceCode(resourceCode, resourceType)
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	var userList []*User
	for _, userId := range userIdList {
		value, err := u.model.Get(userId)
		if err != nil {
			printer.Error(err)
			continue
		}
		user, ok := value.(*UserWithAuthorization)
		if !ok {
			printer.Error("Not UserWithAuthorization. ")
			continue
		}
		userList = append(userList, &user.User)
	}
	return userList, nil
}

func (u *userModel) CreateUser(userCode, userType string) error {
	_, err := u.GetUser(userCode, userType)
	if err == nil {
		return errors.New("Exists. ")
	}
	var user User
	user.UserCode = userCode
	user.UserType = userType
	_, err = u.model.Create(&user, nil)
	if err != nil {
		printer.Error(err)
		return err
	}
	return nil
}

func (u *userModel) CreateResource(userCode, userType, resourceCode, resourceType, parentCode, parentType string,
	isPublic bool) error {

	_, err := u.GetResource(resourceCode, resourceType)
	if err == nil {
		return errors.New("Exists. ")
	}
	user, err := u.GetUser(userCode, userType)
	if err != nil {
		printer.Error(err)
		return err
	}
	var parent *Resource
	var resource Resource
	if len(parentCode) != 0 && len(parentType) != 0 {
		parent, err = u.authorizationModel.resourceModel.getResource(parentCode, parentType)
		if err != nil {
			printer.Error(err)
			return err
		}
		resource.ParentId = parent.GetId()
	}
	resource.ResourceCode = resourceCode
	resource.ResourceType = resourceType
	resourceId, err := u.authorizationModel.resourceModel.model.Create(&resource, nil)
	if err != nil {
		printer.Error(err)
		return err
	}

	var authorization Authorization
	authorization.UserId = user.GetId()
	authorization.ResourceId = resourceId
	authorization.IsOwner = true
	authorization.IsHeritable = true
	authorization.IsUpdatable = true
	authorization.IsDeletable = true
	_, err = u.authorizationModel.model.Create(&authorization, nil)
	if err != nil {
		printer.Error(err)
		_ = u.DeleteResource(resourceCode, resourceType)
		return err
	}
	_ = u.model.SyncSingle(user.GetId())
	if isPublic && parent != nil {
		userIdList, err := u.authorizationModel.getUserListByResourceCode(parentCode, parentType)
		if err != nil {
			printer.Error(err)
		}
		for _, userId := range userIdList {
			value, err := u.model.Get(userId)
			if err != nil {
				printer.Error(err)
				continue
			}
			user, ok := value.(*UserWithAuthorization)
			if !ok {
				printer.Error("Not UserWithAuthorization. ")
				continue
			}
			a, ok := user.AuthorizationMap[parent.GetId()]
			if !ok {
				printer.Error("Authorization is not found. ")
				continue
			}
			if a.IsHeritable {
				var authorization Authorization
				authorization.UserId = userId
				authorization.ResourceId = resourceId
				authorization.IsOwner = a.IsOwner
				authorization.IsHeritable = a.IsHeritable
				authorization.IsUpdatable = a.IsUpdatable
				authorization.IsDeletable = a.IsDeletable
				authorization.Operations = a.Operations
				_, err = u.authorizationModel.model.Create(&authorization, nil)
				if err != nil {
					printer.Error(err)
					_ = u.DeleteResource(resourceCode, resourceType)
					return err
				}
			}
			_ = u.model.SyncSingle(userId)
		}
	}
	return nil
}

func (u *userModel) CreateAuthorization(userCode, userType, resourceCode, resourceType, operations string,
	isHeritable, isUpdatable, isDeletable bool) error {

	_, err := u.GetAuthorization(userCode, userType, resourceCode, resourceType)
	if err == nil {
		return errors.New("Exists. ")
	}
	user, err := u.GetUser(userCode, userType)
	if err != nil {
		printer.Error(err)
		return err
	}
	resource, err := u.GetResource(resourceCode, resourceType)
	if err != nil {
		printer.Error(err)
		return err
	}
	var authorization Authorization
	authorization.UserId = user.GetId()
	authorization.ResourceId = resource.GetId()
	authorization.IsOwner = false
	authorization.IsHeritable = isHeritable
	authorization.IsUpdatable = isUpdatable
	authorization.IsDeletable = isDeletable
	authorization.Operations = operations
	_, err = u.authorizationModel.model.Create(&authorization, nil)
	if err != nil {
		printer.Error(err)
		return err
	}
	_ = u.model.SyncSingle(user.GetId())
	return nil
}

func (u *userModel) DeleteUser(userCode, userType string) error {
	user, err := u.GetUserWithAuthorization(userCode, userType)
	if err != nil {
		printer.Error(err)
		return err
	}
	for _, a := range user.AuthorizationMap {
		err := u.authorizationModel.model.Delete(a.GetId(), nil)
		if err != nil {
			printer.Error(err)
			continue
		}
	}
	err = u.model.Delete(user.GetId(), nil)
	if err != nil {
		printer.Error(err)
		return err
	}
	return nil
}

func (u *userModel) DeleteResource(resourceCode, resourceType string) error {
	resource, err := u.GetResource(resourceCode, resourceType)
	if err != nil {
		printer.Error(err)
		return err
	}
	userIdList, err := u.authorizationModel.getUserListByResourceCode(resourceCode, resourceType)
	if err != nil {
		printer.Error(err)
		return err
	}
	for _, userId := range userIdList {
		value, err := u.model.Get(userId)
		if err != nil {
			printer.Error(err)
			continue
		}
		user, ok := value.(*UserWithAuthorization)
		if !ok {
			printer.Error("Not UserWithAuthorization. ")
			continue
		}
		authorization, ok := user.AuthorizationMap[resource.GetId()]
		if !ok {
			printer.Error("Authorization is not found. ")
			continue
		}
		err = u.authorizationModel.model.Delete(authorization.GetId(), nil)
		if err != nil {
			printer.Error(err)
			continue
		}
		_ = u.model.SyncSingle(userId)
	}
	err = u.authorizationModel.resourceModel.model.Delete(resource.GetId(), nil)
	if err != nil {
		printer.Error(err)
		return err
	}
	return nil
}

func (u *userModel) DeleteAuthorization(userCode, userType, resourceCode, resourceType string) error {
	user, err := u.GetUser(userCode, userType)
	if err != nil {
		printer.Error(err)
		return err
	}
	authorization, err := u.GetAuthorization(userCode, userType, resourceCode, resourceType)
	if err != nil {
		printer.Error(err)
		return err
	}
	err = u.authorizationModel.model.Delete(authorization.GetId(), nil)
	if err != nil {
		printer.Error(err)
		return err
	}
	_ = u.model.SyncSingle(user.GetId())
	return nil
}
