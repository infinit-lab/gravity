package model2

import (
	"github.com/infinit-lab/gravity/printer"
	"gorm.io/gorm"
	"time"
)

type IResource interface {
	GetID() uint
	SetID(id uint)
	GetCode() string
	SetCode(code string)
	Topic() string
}

type Resource struct {
	ID        uint           `gorm:"primarykey"`
	CreatedAt time.Time      `json:"createAt"`
	UpdatedAt time.Time      `json:"updateAt"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	Code      string         `json:"code" gorm:"column:code;type:VARCHAR(256);unique_index;not null;"`
}

func (r *Resource) GetCode() string {
	return r.Code
}

func (r *Resource) SetCode(code string) {
	r.Code = code
}

func (r *Resource) GetID() uint {
	return r.ID
}

func (r *Resource) SetID(id uint) {
	r.ID = id
}

func (r *Resource) Topic() string {
	return ""
}

type Model interface {
	Begin()
	Commit()
	Rollback()
	Create(r IResource, context interface{}) (code string, err error)
	Update(r IResource, context interface{}) error
	Delete(code string, context interface{}) error
	Get(code string) (IResource, error)
	GetList(query string, conditions ...interface{}) (interface{}, error)
}

func New(db *gorm.DB, resource IResource) (Model, error) {
	m := new(model)
	err := m.init(db, resource)
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	return m, nil
}
