package authorization

import m "github.com/infinit-lab/gravity/model"

type Authorization struct {
	m.Id
	ResourceId   int                   `json:"resourceId" db:"resourceId"`
	UserId       int                   `json:"userId" db:"userId"`
	IsHeritable  bool                  `json:"isHeritable" db:"isHeritable"`
	IsUpdatable  bool                  `json:"isUpdatable" db:"isUpdatable"`
	IsDeletable  bool                  `json:"isDeletable" db:"isDeletable"`
	OperationMap map[string]*Operation `json:"operationMap"`
}

type Operation struct {
	m.Id
	AuthorizationId int    `json:"authorizationId" db:"authorizationId"`
	Operation       string `json:"operation" db:"operation"`
}
