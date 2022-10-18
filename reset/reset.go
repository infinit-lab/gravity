package reset

import (
	"errors"
	"fmt"
	"github.com/infinit-lab/gravity/printer"
	"os"
	"sync"
)

type Reseter interface {
	Name() string
	Depends() []string
	Check(data interface{}) (checkData interface{}, err error)
	Reset(data interface{}, checkData interface{}) (resetData interface{}, isExit bool, err error)
	AfterReset(data interface{}, checkData interface{}, resetData interface{}, isExit bool) error
}

type DefaultReseter struct {
}

func (d *DefaultReseter) Depends() []string {
	return []string{}
}

func (d *DefaultReseter) Check(data interface{}) (interface{}, error) {
	return nil, nil
}

func (d *DefaultReseter) AfterReset(data interface{}, checkData interface{}, resetData interface{}, isExit bool) error {
	return nil
}

type reseter struct {
	r         Reseter
	checkData interface{}
	resetData interface{}
}

type manager struct {
	reseters      []*reseter
	resetersMutex sync.Mutex
	isExit        bool
}

var m *manager

func init() {
	m = new(manager)
}

func RegisterReseter(r Reseter) error {
	if len(r.Name()) == 0 {
		return errors.New("名称不能为空")
	}
	_, err := m.getReseterByName(r.Name())
	if err == nil {
		return errors.New("该名称已经注册")
	}
	depends := r.Depends()
	for _, d := range depends {
		t, err := m.getReseterByName(d)
		if err != nil {
			err := fmt.Sprintf("%s 不存在", d)
			return errors.New(err)
		}
		for _, td := range t.r.Depends() {
			if td == r.Name() {
				err := fmt.Sprintf("%s %s 循环依赖", r.Name(), d)
				return errors.New(err)
			}
		}
	}
	m.resetersMutex.Lock()
	m.resetersMutex.Unlock()
	rst := new(reseter)
	rst.r = r
	m.reseters = append(m.reseters, rst)
	return nil
}

func Reset(data interface{}) error {
	if err := m.check(data); err != nil {
		printer.Error(err)
		return err
	}
	if err := m.reset(data); err != nil {
		printer.Error(err)
		return err
	}
	if err := m.after(data); err != nil {
		printer.Error(err)
		return err
	}
	if m.isExit {
		printer.Warning("Exit.")
		os.Exit(0)
	}
	return nil
}

func (m *manager) getReseterByName(name string) (*reseter, error) {
	m.resetersMutex.Lock()
	defer m.resetersMutex.Unlock()

	for _, r := range m.reseters {
		if r.r.Name() == name {
			return r, nil
		}
	}
	return nil, errors.New("未找到")
}

func (m *manager) nameMap() map[string]map[string]int {
	m.resetersMutex.Lock()
	defer m.resetersMutex.Unlock()

	nameMap := make(map[string]map[string]int)
	for _, r := range m.reseters {
		depends := make(map[string]int)
		for _, d := range r.r.Depends() {
			depends[d] = 0
		}
		nameMap[r.r.Name()] = depends
	}
	return nameMap
}

func (m *manager) loop(data interface{}, f func(data interface{}, r *reseter) error) error {
	nameMap := m.nameMap()
	for {
		if len(nameMap) == 0 {
			break
		}
		var nameList []string
		for name, depends := range nameMap {
			if len(depends) != 0 {
				continue
			}
			r, err := m.getReseterByName(name)
			if err != nil {
				printer.Error(err)
				return err
			}
			err = f(data, r)
			if err != nil {
				printer.Error(err)
				return err
			}
			nameList = append(nameList, name)
		}
		for _, name := range nameList {
			delete(nameMap, name)
		}
		for _, name := range nameList {
			for _, depends := range nameMap {
				delete(depends, name)
			}
		}
	}
	return nil
}

func (m *manager) check(data interface{}) error {
	return m.loop(data, func(data interface{}, r *reseter) error {
		var err error
		r.checkData, err = r.r.Check(data)
		if err != nil {
			printer.Error(err)
			return err
		}
		return nil
	})
}

func (m *manager) reset(data interface{}) error {
	m.isExit = false
	return m.loop(data, func(data interface{}, r *reseter) error {
		var err error
		isExit := false
		r.resetData, isExit, err = r.r.Reset(data, r.checkData)
		if err != nil {
			printer.Error(err)
			return err
		}
		if isExit {
			m.isExit = true
		}
		return nil
	})
}

func (m *manager) after(data interface{}) error {
	return m.loop(data, func(data interface{}, r *reseter) error {
		err := r.r.AfterReset(data, r.checkData, r.resetData, m.isExit)
		if err != nil {
			printer.Error(err)
			return err
		}
		return nil
	})
}
