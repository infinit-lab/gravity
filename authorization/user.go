package authorization

import m "github.com/infinit-lab/gravity/model"

type User struct {
	m.Id
	UserType         string                 `json:"userType" db:"userType"`
	RelativeList     []*UserRelative        `json:"relativeList" db:"relativeList"`
	AuthorizationMap map[int]*Authorization `json:"authorizationMap"`
}

type UserInfo struct {
	m.Id
	UserType string `json:"userType" db:"userType"`
	RelativeList []int `json:"relativeList" db:"relativeList"`
}

type UserRelative struct {
	m.Id
	UserId     int `json:"userId" db:"userId"`
	RelativeId int `json:"relativeId" db:"relativeId"`
}
