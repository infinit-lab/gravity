package model2

import (
	"github.com/infinit-lab/gravity/printer"
	"gorm.io/gorm"
	"time"
)

type IData interface {
	GetID() uint
	SetID(id uint)
	GetCode() string
	SetCode(code string)
	Topic() string
}

type Data struct {
	ID        uint           `gorm:"primarykey"`
	CreatedAt time.Time      `json:"createAt"`
	UpdatedAt time.Time      `json:"updateAt"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	Code      string         `json:"code" gorm:"column:code;type:VARCHAR(256);unique_index;not null;"`
}

func (r *Data) GetCode() string {
	return r.Code
}

func (r *Data) SetCode(code string) {
	r.Code = code
}

func (r *Data) GetID() uint {
	return r.ID
}

func (r *Data) SetID(id uint) {
	r.ID = id
}

func (r *Data) Topic() string {
	return ""
}

type Model interface {
	Begin()
	Commit()
	Rollback()
	Create(data IData, context interface{}) (code string, err error)
	Update(data IData, context interface{}) error
	Delete(code string, context interface{}) error
	Get(code string) (IData, error)
	GetList(query string, conditions ...interface{}) (interface{}, error)
}

func New(db *gorm.DB, data IData) (Model, error) {
	m := new(model)
	err := m.init(db, data)
	if err != nil {
		printer.Error(err)
		return nil, err
	}
	return m, nil
}
