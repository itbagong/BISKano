package rbaclogic

import (
	"github.com/ariefdarmawan/datahub"
)

type RbacService interface {
	HasRole(userid, roleid string) bool
	AddRoles(uid string, roleids ...string)
	RemoveRole(uid string, roleids ...string)
	UserRoles(userid string) []string
	RoleUsers(userid string) []string
}

type proxy struct {
	h *datahub.Hub
}

func NewRbac(h *datahub.Hub) *proxy {
	p := new(proxy)
	p.h = h
	return p
}

func (p *proxy) DataHub() *datahub.Hub {
	return p.h
}

func (m *proxy) HasRole(userid, roleid string) bool {
	return false
}

func (m *proxy) AddRoles(uid string, roleids ...string) {
}

func (m *proxy) RemoveRole(uid string, roleids ...string) {
}

func (m *proxy) UserRoles(userid string) []string {
	return []string{}
}

func (m *proxy) RoleUsers(userid string) []string {
	return []string{}
}
