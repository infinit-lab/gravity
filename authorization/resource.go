package authorization

import m "github.com/infinit-lab/gravity/model"

type Resource struct {
	m.Id
	ResourceType string              `json:"resourceType" db:"resourceType"`
	ParentList   []*ResourceParent   `json:"parentList"`
	RelativeList []*ResourceRelative `json:"relativeList"`
}

type ResourceRelative struct {
	m.Id
	ResourceId int `json:"resourceId" db:"resourceId"`
	RelativeId int `json:"relativeId" db:"relativeId"`
}

type ResourceParent struct {
	m.Id
	ResourceId int `json:"resourceId" db:"resourceId"`
	ParentId   int `json:"parentId" db:"parentId"`
}
