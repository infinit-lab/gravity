package authorization

import (
	"encoding/json"
	"github.com/infinit-lab/gravity/database"
	"github.com/infinit-lab/gravity/printer"
	"testing"
)

var mdl Model
var userCode string
var userType string
var childUserCode string
var childUserType string
var resourceCode string
var resourceType string
var childResourceCode string
var childResourceType string

func TestNew(t *testing.T) {
	db, err := database.NewDatabase("sqlite3", "authorization.db")
	if err != nil {
		t.Fatal(err)
	}
	mdl, err = New(db)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUserModel_CreateUser(t *testing.T) {
	userCode = "123"
	userType = "role"
	err := mdl.CreateUser(userCode, userType)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUserModel_GetUser(t *testing.T) {
	user, err := mdl.GetUser(userCode, userType)
	if err != nil {
		t.Fatal(err)
	}
	printer.Tracef("%v", *user)
	if user.UserCode != userCode || user.UserType != userType {
		t.Fatal("Wrong params. ")
	}
}

func TestUserModel_CreateUser2(t *testing.T) {
	childUserCode = "111"
	childUserType = "role"
	err := mdl.CreateUser(childUserCode, childUserType)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUserModel_GetUser2(t *testing.T) {
	user, err := mdl.GetUser(childUserCode, childUserType)
	if err != nil {
		t.Fatal(err)
	}
	printer.Tracef("%v", *user)
	if user.UserCode != childUserCode || user.UserType != childUserType {
		t.Fatal("Wrong params. ")
	}
}

func TestUserModel_CreateResource(t *testing.T) {
	resourceCode = "456"
	resourceType = "device"
	err := mdl.CreateResource(userCode, userType, resourceCode, resourceType, "", "", false)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUserModel_GetResource(t *testing.T) {
	resource, err := mdl.GetResource(resourceCode, resourceType)
	if err != nil {
		t.Fatal(err)
	}
	printer.Tracef("%v", *resource)
	if resource.ResourceCode != resourceCode || resource.ResourceType != resourceType {
		t.Fatal("Wrong params. ")
	}
}

func TestUserModel_CreateResource2(t *testing.T) {
	childResourceCode = "789"
	childResourceType = "channel"
	err := mdl.CreateResource(childUserCode, childUserType, childResourceCode, childResourceType, resourceCode, resourceType, true)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUserModel_GetResource2(t *testing.T) {
	resource, err := mdl.GetResource(childResourceCode, childResourceType)
	if err != nil {
		t.Fatal(err)
	}
	printer.Tracef("%v", *resource)
	if resource.ResourceCode != childResourceCode || resource.ResourceType != childResourceType {
		t.Fatal("Wrong params. ")
	}
}

func TestUserModel_GetAuthorization(t *testing.T) {
	authorization, err := mdl.GetAuthorization(userCode, userType, resourceCode, resourceType)
	if err != nil {
		t.Fatal(err)
	}
	printer.Tracef("%v", *authorization)
	if !authorization.IsOwner || !authorization.IsHeritable || !authorization.IsUpdatable || !authorization.IsDeletable {
		t.Fatal("Wrong params. ")
	}
	authorization, err = mdl.GetAuthorization(userCode, userType, childResourceCode, childResourceType)
	if err != nil {
		t.Fatal(err)
	}
	printer.Tracef("%v", *authorization)
	if !authorization.IsOwner || !authorization.IsHeritable || !authorization.IsUpdatable || !authorization.IsDeletable {
		t.Fatal("Wrong params. ")
	}
}

func TestUserModel_GetAuthorization3(t *testing.T) {
	authorization, err := mdl.GetAuthorization(childUserCode, childUserType, childResourceCode, childResourceType)
	if err != nil {
		t.Fatal(err)
	}
	printer.Tracef("%v", *authorization)
	if !authorization.IsOwner || !authorization.IsHeritable || !authorization.IsUpdatable || !authorization.IsDeletable {
		t.Fatal("Wrong params. ")
	}
	_, err = mdl.GetAuthorization(childUserCode, childUserType, resourceCode, resourceType)
	if err == nil {
		t.Fatal("Should not authorized. ")
	}
}

func TestUserModel_GetUserListByResource(t *testing.T) {
	users, err := mdl.GetUserListByResource(childResourceCode, childResourceType)
	if err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(users)
	printer.Trace(string(data))
}

func TestUserModel_GetAuthorizationList(t *testing.T) {
	authorizations, err := mdl.GetAuthorizationList(userCode, userType)
	if err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(authorizations)
	printer.Trace(string(data))
}

func TestUserModel_GetAuthorizationListByResourceType(t *testing.T) {
	authorizations, err := mdl.GetAuthorizationListByResourceType(userCode, userType, resourceType)
	if err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(authorizations)
	printer.Trace(string(data))
}

func TestUserModel_DeleteUser2(t *testing.T) {
	err := mdl.DeleteUser(childUserCode, childUserType)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUserModel_DeleteResource2(t *testing.T) {
	err := mdl.DeleteResource(childResourceCode, childResourceType)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUserModel_DeleteResource(t *testing.T) {
	err := mdl.DeleteResource(resourceCode, resourceType)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUserModel_GetAuthorization2(t *testing.T) {
	_, err := mdl.GetAuthorization(userCode, userType, resourceCode, resourceType)
	if err == nil {
		t.Fatal("Should not be authorized. ")
	}
}

func TestUserModel_DeleteUser(t *testing.T) {
	err := mdl.DeleteUser(userCode, userType)
	if err != nil {
		t.Fatal(err)
	}
}
