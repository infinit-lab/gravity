package authorization

import (
	"errors"
	m "github.com/infinit-lab/gravity/model"
	"github.com/infinit-lab/gravity/printer"
)

type model struct {
	resourceModel               m.Model
	resourceParentModel         m.Model
	resourceRelativeModel       m.Model
	userModel                   m.Model
	userRelativeModel           m.Model
	authorizationModel          m.Model
	authorizationOperationModel m.Model
}

func (m *model) beforeInsertResourceLayer(id int, resource *Resource) {
	parentList, err := m.resourceParentModel.GetList()
	if err != nil {
		printer.Error(err)
		return
	}
	relativeList, err := m.resourceRelativeModel.GetList()
	if err != nil {
		printer.Error(err)
		return
	}
	for _, value := range parentList {
		parent, ok := value.(*ResourceParent)
		if !ok {
			continue
		}
		if parent.ResourceId == id {
			resource.ParentList = append(resource.ParentList, parent)
		}
	}
	for _, value := range relativeList {
		relative, ok := value.(*ResourceRelative)
		if !ok {
			continue
		}
		if relative.ResourceId == id {
			resource.RelativeList = append(resource.RelativeList, relative)
		}
	}
}

func (m *model) beforeInsertAuthorizationLayer(id int, authorization *Authorization) {
	operationList, err := m.authorizationOperationModel.GetList()
	if err != nil {
		printer.Error(err)
		return
	}
	if authorization.OperationMap == nil {
		authorization.OperationMap = make(map[string]*Operation)
	}
	for _, value := range operationList {
		operation, ok := value.(*Operation)
		if !ok {
			continue
		}
		if operation.AuthorizationId == id {
			authorization.OperationMap[operation.Operation] = operation
		}
	}
}

func (m *model) beforeInsertUserLayer(id int, user *User) {
	authorizationList, err := m.authorizationModel.GetList()
	if err != nil {
		printer.Error(err)
		return
	}
	relativeList, err := m.userRelativeModel.GetList()
	if err != nil {
		printer.Error(err)
		return
	}
	for _, value := range relativeList {
		relative, ok := value.(*UserRelative)
		if !ok {
			continue
		}
		user.RelativeList = append(user.RelativeList, relative)
	}
	if user.AuthorizationMap == nil {
		user.AuthorizationMap = make(map[int]*Authorization)
	}
	for _, value := range authorizationList {
		authorization, ok := value.(*Authorization)
		if !ok {
			continue
		}
		user.AuthorizationMap[authorization.ResourceId] = authorization
	}
}

func (m *model) GetResource(resourceId int) (*Resource, error) {
	value, err := m.resourceModel.Get(resourceId)
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	r, ok := value.(*Resource)
	if !ok {
		return nil, errors.New("Invalid value. ")
	}
	return r, nil
}

func (m *model) GetUserInfo(userId int) (*UserInfo, error) {
	value, err := m.userModel.Get(userId)
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	user, ok := value.(*User)
	if !ok {
		return nil, errors.New("Invalid value. ")
	}
	userInfo := new(UserInfo)
	userInfo.UserType = user.UserType
	for _, relative := range user.RelativeList {
		userInfo.RelativeList = append(userInfo.RelativeList, relative.RelativeId)
	}
	return userInfo, nil
}

func (m *model) getRelativeResourceIdList(resourceId int) []int {
	r, err := m.GetResource(resourceId)
	if err != nil {
		return nil
	}
	var resourceIdList []int
	resourceIdList = append(resourceIdList, resourceId)
	for _, relative := range r.RelativeList {
		list := m.getRelativeResourceIdList(relative.RelativeId)
		if list != nil {
			resourceIdList = append(resourceIdList, list...)
		}
	}
	return resourceIdList
}

func (m *model) getAuthorizationList(userId int, resourceIdList []int) ([]*Authorization, error) {
	value, err := m.userModel.Get(userId)
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	u, ok := value.(*User)
	if !ok {
		return nil, errors.New("Invalid user. ")
	}
	var authorizationList []*Authorization
	for _, id := range resourceIdList {
		authorization, ok := u.AuthorizationMap[id]
		if !ok {
			continue
		}
		authorizationList = append(authorizationList, authorization)
	}

	for _, relative := range u.RelativeList {
		list, err := m.getAuthorizationList(relative.RelativeId, resourceIdList)
		if err != nil {
			continue
		}
		if list == nil {
			continue
		}
		authorizationList = append(authorizationList, list...)
	}
	return authorizationList, nil
}

func (m *model) GetAuthorization(userId int, resourceId int) (*Authorization, error) {
	resourceIdList := m.getRelativeResourceIdList(resourceId)
	if resourceIdList == nil {
		return nil, errors.New("Invalid resource. ")
	}
	authorizationList, _ := m.getAuthorizationList(userId, resourceIdList)
	if authorizationList == nil {
		return nil, errors.New("Unauthorized. ")
	}
	authorization := new(Authorization)
	authorization.UserId = userId
	authorization.ResourceId = resourceId
	authorization.OperationMap = make(map[string]*Operation)
	for _, a := range authorizationList {
		if a.IsHeritable == true {
			authorization.IsHeritable = true
		}
		if a.IsUpdatable == true {
			authorization.IsUpdatable = true
		}
		if a.IsDeletable == true {
			authorization.IsDeletable = true
		}
		for _, o := range a.OperationMap {
			_, ok := authorization.OperationMap[o.Operation]
			if !ok {
				authorization.OperationMap[o.Operation] = o
			}
		}
	}
	return authorization, nil
}

func (m *model) CreateUser(relativeList []int, userType string) (userId int, err error) {
	for _, relativeId := range relativeList {
		if _, err := m.userModel.Get(relativeId); err != nil {
			return 0, errors.New("Invalid relative. ")
		}
	}
	user := new(User)
	user.UserType = userType
	if userId, err = m.userModel.Create(user, nil); err != nil {
		printer.Error(err)
		return 0, err
	}

	relative := new(UserRelative)
	relative.UserId = userId
	for _, relativeId := range relativeList {
		relative.RelativeId = relativeId
		if _, err = m.userRelativeModel.Create(relative, nil); err != nil {
			printer.Error(err)
			_ = m.DeleteUser(userId)
			return 0, err
		}
	}
	err = m.userModel.SyncSingle(userId)
	if err != nil {
		printer.Error(err)
		_ = m.DeleteUser(userId)
		return 0, err
	}
	return userId, nil
}

func (m *model) CreateResource(resourceType string, parentList, relativeList []int, isPublic bool) (resourceId int, err error) {
	for _, parentId := range parentList {
		if _, err := m.GetResource(parentId); err != nil {
			return 0, errors.New("Invalid parent. ")
		}
	}
	for _, relativeId := range relativeList {
		if _, err := m.GetResource(relativeId); err != nil {
			return 0, errors.New("Invalid relative. ")
		}
	}
	resource := new(Resource)
	resource.ResourceType = resourceType
	if resourceId, err = m.resourceModel.Create(resource, nil); err != nil {
		printer.Error(err)
		return 0, err
	}

	parent := new(ResourceParent)
	parent.ResourceId = resourceId
	for _, parentId := range parentList {
		parent.ParentId = parentId
		if _, err = m.resourceParentModel.Create(parent, nil); err != nil {
			printer.Error(err)
			_ = m.DeleteResource(resourceId)
			return 0, err
		}
	}

	relative := new(ResourceRelative)
	relative.ResourceId = resourceId
	for _, relativeId := range relativeList {
		relative.RelativeId = relativeId
		if _, err = m.resourceRelativeModel.Create(relative, nil); err != nil {
			printer.Error(err)
			_ = m.DeleteResource(resourceId)
			return 0, err
		}
	}

	if isPublic {
		userList, err := m.userModel.GetList()
		if err == nil {
			for _, parentId := range parentList {
				for _, value := range userList {
					user := value.(*User)
					authorization, ok := user.AuthorizationMap[parentId]
					if !ok {
						continue
					}
					if !authorization.IsHeritable {
						continue
					}
					var operationList []string
					for key, _ := range authorization.OperationMap {
						operationList = append(operationList, key)
					}
					if err := m.CreateAuthorization(user.GetId(), resourceId, operationList, authorization.IsHeritable,
						authorization.IsUpdatable, authorization.IsDeletable); err != nil {
						printer.Error(err)
						_ = m.DeleteResource(resourceId)
						return 0, err
					}
				}
			}
		}
	}

	err = m.resourceModel.SyncSingle(resourceId)
	if err != nil {
		printer.Error(err)
		_ = m.DeleteResource(resourceId)
		return 0, err
	}
	return resourceId, nil
}

func (m *model) CreateAuthorization(userId, resourceId int, operationList []string, isHeritable, isUpdatable, isDeletable bool) error {
	value, err := m.userModel.Get(userId)
	if err != nil {
		printer.Error(err)
		return nil
	}
	user, ok := value.(*User)
	if !ok {
		return errors.New("Invalid user. ")
	}
	_, ok = user.AuthorizationMap[resourceId]
	if ok {
		return nil
	}
	if _, err := m.GetResource(resourceId); err != nil {
		printer.Error(err)
		return nil
	}
	authorization := new(Authorization)
	authorization.UserId = userId
	authorization.ResourceId = resourceId
	authorization.IsHeritable = isHeritable
	authorization.IsUpdatable = isUpdatable
	authorization.IsDeletable = isDeletable
	authorizationId, err := m.authorizationModel.Create(authorization, nil)
	if err != nil {
		printer.Error(err)
		return err
	}
	operation := new(Operation)
	operation.AuthorizationId = authorizationId
	for _, o := range operationList {
		operation.Operation = o
		_, err := m.authorizationOperationModel.Create(operation, nil)
		if err != nil {
			printer.Error(err)
			_ = m.DeleteAuthorization(userId, resourceId)
			return err
		}
	}
	if err := m.authorizationModel.SyncSingle(authorizationId); err != nil {
		printer.Error(err)
		_ = m.DeleteAuthorization(userId, resourceId)
		return err
	}
	if err := m.userModel.SyncSingle(userId); err != nil {
		printer.Error(err)
		_ = m.DeleteAuthorization(userId, resourceId)
		return err
	}

	return nil
}

func (m *model) DeleteUser(userId int) error {
	if err := m.userModel.Delete(userId, nil); err != nil {
		printer.Error(err)
		return err
	}
	_ = m.userRelativeModel.Sync()
	_ = m.userModel.Sync()
	return nil
}

func (m *model) DeleteResource(resourceId int) error {
	if err := m.resourceModel.Delete(resourceId, nil); err != nil {
		printer.Error(err)
		return err
	}
	_ = m.resourceParentModel.Sync()
	_ = m.resourceRelativeModel.Sync()
	_ = m.authorizationOperationModel.Sync()
	_ = m.authorizationModel.Sync()
	_ = m.userModel.Sync()
	return nil
}

func (m *model) DeleteAuthorization(userId, resourceId int) error {
	value, err := m.userModel.Get(userId)
	if err != nil {
		printer.Error(err)
		return err
	}
	user, ok := value.(*User)
	if !ok {
		return errors.New("Invalid user. ")
	}
	authorization, ok := user.AuthorizationMap[resourceId]
	if !ok {
		return errors.New("Unauthorized. ")
	}
	if err := m.authorizationModel.Delete(authorization.GetId(), nil); err != nil {
		printer.Error(err)
		return err
	}
	_ = m.userModel.SyncSingle(userId)
	return nil
}
