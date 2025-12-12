package deployer

import (
	"fmt"

	"git.kanosolution.net/kano/kaos"
	"github.com/ariefdarmawan/byter"
)

var (
	engines = map[string]func() (Deployer, error){}
)

type Deployer interface {
	This() Deployer
	Name() string
	SetByter(byter.Byter) Deployer
	Byter() byter.Byter
	SetThis(Deployer) Deployer
	PreDeploy(interface{}) error
	Deploy(*kaos.Service, interface{}) error
	DeployRoute(*kaos.Service, *kaos.ServiceRoute, interface{}) error
}

func NewDeployer(name string) (Deployer, error) {
	fn, found := engines[name]
	if !found {
		return new(BaseDeployer), fmt.Errorf("deployer %s is not exist", name)
	}
	d, e := fn()
	if e != nil {
		return new(BaseDeployer), fmt.Errorf("fail create deployer %s: %s", name, e.Error())
	}
	d.SetThis(d)
	return d, nil
}

func RegisterDeployer(name string, fn func() (Deployer, error)) {
	engines[name] = fn
}

type BaseDeployer struct {
	_this  Deployer
	_byter byter.Byter
}

func (d *BaseDeployer) Name() string {
	return "base-deployer"
}

func (d *BaseDeployer) PreDeploy(interface{}) error {
	return nil
}

func (d *BaseDeployer) This() Deployer {
	if d._this == nil {
		return d
	}
	return d._this
}

func (d *BaseDeployer) SetByter(b byter.Byter) Deployer {
	d._byter = b
	return d
}

func (d *BaseDeployer) Byter() byter.Byter {
	if d._byter == nil {
		d._byter = byter.NewByter("")
	}
	return d._byter
}

func (d *BaseDeployer) SetThis(d0 Deployer) Deployer {
	d._this = d0
	return d
}

func (d *BaseDeployer) Deploy(s *kaos.Service, obj interface{}) error {
	if e := d.This().PreDeploy(obj); e != nil {
		return fmt.Errorf("fail deploying service %s: %s", s.BasePoint(), e.Error())
	}

	routes, e := s.PrepareRoutes(d._this.Name())
	if e != nil {
		return fmt.Errorf("fail deploying service %s: %s", s.BasePoint(), e.Error())
	}
	for _, route := range routes {
		if e := d.This().DeployRoute(s, route, obj); e != nil {
			return fmt.Errorf("fail deploying %s: %s", route.Path, e.Error())
		}
	}
	return nil
}

func (d *BaseDeployer) DeployRoute(*kaos.Service, *kaos.ServiceRoute, interface{}) error {
	return fmt.Errorf("not implemented")
}
