package authorization

import (
	"errors"
	"github.com/infinit-lab/gravity/database"
	m "github.com/infinit-lab/gravity/model"
	"github.com/infinit-lab/gravity/printer"
	"sort"
)

type User struct {
	m.PrimaryKey
	UserCode string `json:"userCode" db:"userCode" db_type:"VARCHAR(64)" db_index:"index" db_omit:"update" db_default:"''"`
	UserType string `json:"userType" db:"userType" db_type:"VARCHAR(64)" db_index:"index" db_omit:"update" db_default:"''"`
	ParentCode string    `json:"parentCode" db:"parentCode" db_type:"VARCHAR(64)" db_omit:"update" db_default:"''"`
}

type UserWithAuthorization struct {
	User
	AuthorizationMap map[string]*Authorization `json:"authorizationMap"`
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
	u.model.SetBeforeGetLayer(func(resource interface{}) {
		user, ok := resource.(*UserWithAuthorization)
		if !ok {
			return
		}
		list, err := u.authorizationModel.getAuthorizationListByUserCode(user.GetCode())
		if err != nil {
			printer.Error(err)
			return
		}
		user.AuthorizationMap = make(map[string]*Authorization)
		for _, a := range list {
			user.AuthorizationMap[a.ResourceCode] = a
		}
	})
	return u, nil
}

func (u *userModel) GetResource(resourceCode, resourceType string) (*Resource, error) {
	return u.authorizationModel.resourceModel.getResource(resourceCode, resourceType)
}

func (u *userModel) GetUserWithAuthorization(userCode, userType string) (*UserWithAuthorization, error) {
	value, err := u.model.Get("WHERE `userCode` = ? AND `userType` = ?", userCode, userType)
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	return value.(*UserWithAuthorization), nil
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
	a, ok := user.AuthorizationMap[resource.GetCode()]
	if !ok {
		return nil, errors.New("未授权")
	}
	return a, nil
}

func (u *userModel) GetAuthorizationList(userCode, userType string) ([]*Authorization, error) {
	user, err := u.GetUserWithAuthorization(userCode, userType)
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	var keyList []string
	for key, _ := range user.AuthorizationMap {
		keyList = append(keyList, key)
	}
	sort.Strings(keyList)
	var authList []*Authorization
	for _, key := range keyList {
		authList = append(authList, user.AuthorizationMap[key])
	}
	return authList, nil
}

func (u *userModel) GetAuthorizationListByResourceType(userCode, userType, resourceType string) ([]*Authorization, error) {
	user, err := u.GetUserWithAuthorization(userCode, userType)
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	var keyList []string
	for key, a := range user.AuthorizationMap {
		if a.Resource.ResourceType == resourceType {
			keyList = append(keyList, key)
		}
	}
	sort.Strings(keyList)
	var authList []*Authorization
	for _, key := range keyList {
		authList = append(authList, user.AuthorizationMap[key])
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
		value, err := u.model.GetByCode(userId)
		if err != nil {
			printer.Error(err)
			continue
		}
		user := value.(*UserWithAuthorization)
		userList = append(userList, &user.User)
	}
	return userList, nil
}

func (u *userModel) CreateUser(userCode, userType, parentCode, parentType string) error {
	_, err := u.GetUser(userCode, userType)
	if err == nil {
		return errors.New("用户已存在")
	}
	var user User
	user.UserCode = userCode
	user.UserType = userType
	if len(parentCode) != 0 && len(parentType) != 0 {
		parent, err := u.GetUser(parentCode, parentType)
		if err != nil {
			return err
		}
		user.ParentCode = parent.GetCode()
	}
	_, err = u.model.Create(&user, nil)
	if err != nil {
		printer.Error(err)
		return err
	}
	return nil
}

func (u *userModel) CreateResource(userCode, userType, resourceCode, resourceType, parentCode, parentType string) error {
	_, err := u.GetResource(resourceCode, resourceType)
	if err == nil {
		return errors.New("Exists. ")
	}
	var user *User
	if len(userCode) != 0 || len(userType) != 0 {
		user, err = u.GetUser(userCode, userType)
		if err != nil {
			printer.Error(err)
			return err
		}
	}
	var parent *Resource
	var resource Resource
	if len(parentCode) != 0 && len(parentType) != 0 {
		parent, err = u.authorizationModel.resourceModel.getResource(parentCode, parentType)
		if err != nil {
			printer.Error(err)
			return err
		}
		resource.ParentCode = parent.GetCode()
	}
	resource.ResourceCode = resourceCode
	resource.ResourceType = resourceType
	resourceId, err := u.authorizationModel.resourceModel.model.Create(&resource, nil)
	if err != nil {
		printer.Error(err)
		return err
	}

	if user != nil {
		var authorization Authorization
		authorization.UserCode = user.GetCode()
		authorization.ResourceCode = resourceId
		authorization.IsOwner = true
		authorization.IsHeritable = true
		authorization.IsUpdatable = true
		authorization.IsDeletable = true
		err = u.createAuthorization(authorization)
		if err != nil {
			printer.Error(err)
			_ = u.DeleteResource(resourceCode, resourceType)
			return err
		}
	}

	if parent != nil {
		userIdList, err := u.authorizationModel.getUserListByResourceCode(parentCode, parentType)
		if err != nil {
			printer.Error(err)
		}
		for _, userId := range userIdList {
			value, err := u.model.GetByCode(userId)
			if err != nil {
				printer.Error(err)
				continue
			}
			user := value.(*UserWithAuthorization)
			a, ok := user.AuthorizationMap[parent.GetCode()]
			if !ok {
				continue
			}
			if a.IsHeritable {
				var authorization Authorization
				authorization.UserCode = userId
				authorization.ResourceCode = resourceId
				authorization.IsOwner = a.IsOwner
				authorization.IsHeritable = a.IsHeritable
				authorization.IsUpdatable = a.IsUpdatable
				authorization.IsDeletable = a.IsDeletable
				authorization.Operations = a.Operations
				err = u.createAuthorization(authorization)
				if err != nil {
					printer.Error(err)
					_ = u.DeleteResource(resourceCode, resourceType)
					return err
				}
			}
		}
	}
	return nil
}

func (u *userModel) createAuthorization(authorization Authorization) error {
	value, err := u.model.GetByCode(authorization.UserCode)
	if err != nil {
		printer.Error(err)
		return err
	}
	user := value.(*UserWithAuthorization)
	_, ok := user.AuthorizationMap[authorization.ResourceCode]
	if ok {
		return nil
	}
	authorization.Code = ""
	_, err = u.authorizationModel.model.Create(&authorization, nil)
	if err != nil {
		printer.Error(err)
		return err
	}
	if user.ParentCode != "" {
		authorization.UserCode = user.ParentCode
		err := u.createAuthorization(authorization)
		if err != nil {
			printer.Error(err)
			return err
		}
	}
	return nil
}

func (u *userModel) CreateAuthorization(userCode, userType, resourceCode, resourceType, operations string,
	isHeritable, isUpdatable, isDeletable bool) error {

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
	authorization.UserCode = user.GetCode()
	authorization.ResourceCode = resource.GetCode()
	authorization.IsOwner = false
	authorization.IsHeritable = isHeritable
	authorization.IsUpdatable = isUpdatable
	authorization.IsDeletable = isDeletable
	authorization.Operations = operations
	err = u.createAuthorization(authorization)
	if err != nil {
		printer.Error(err)
		return err
	}
	return nil
}

func (u *userModel) DeleteUser(userCode, userType string) error {
	user, err := u.GetUserWithAuthorization(userCode, userType)
	if err != nil {
		printer.Error(err)
		return err
	}
	err = u.authorizationModel.model.Delete(nil, "WHERE `userCode` = ?", user.GetCode())
	if err != nil {
		printer.Error(err)
		return err
	}
	err = u.model.DeleteByCode(user.GetCode(), nil)
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
	err = u.authorizationModel.model.Delete(nil, "WHERE `resourceCode` = ?", resource.GetCode())
	if err != nil {
		printer.Error(err)
		return err
	}
	err = u.authorizationModel.resourceModel.model.DeleteByCode(resource.GetCode(), nil)
	if err != nil {
		printer.Error(err)
		return err
	}
	return nil
}

func (u *userModel) DeleteAuthorization(userCode, userType, resourceCode, resourceType string) error {
	authorization, err := u.GetAuthorization(userCode, userType, resourceCode, resourceType)
	if err != nil {
		printer.Error(err)
		return err
	}
	err = u.authorizationModel.model.DeleteByCode(authorization.GetCode(), nil)
	if err != nil {
		printer.Error(err)
		return err
	}
	return nil
}
