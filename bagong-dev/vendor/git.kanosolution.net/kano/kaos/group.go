package kaos

import (
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
)

type Group struct {
	mws            []MWFunc
	postMws        []MWFunc
	deployers      []string
	mods           []Mod
	onlyRoutes     []string
	disableRoutes  []string
	defaultHubName string
	defaultEvName  string
}

func (s *Service) Group() *Group {
	g := new(Group)
	return g
}

func (g *Group) RegisterMWs(mws ...MWFunc) *Group {
	g.mws = mws
	return g
}

func (g *Group) RegisterPostMWs(mws ...MWFunc) *Group {
	g.postMws = mws
	return g
}

func (g *Group) SetMod(mods ...Mod) *Group {
	g.mods = mods
	return g
}

func (g *Group) SetDeployer(deployerNames ...string) *Group {
	g.deployers = deployerNames
	return g
}

func (g *Group) AllowOnlyRoute(names ...string) *Group {
	g.onlyRoutes = names
	return g
}

func (g *Group) SetHubName(name string) *Group {
	g.defaultHubName = name
	return g
}

func (g *Group) SetEvName(name string) *Group {
	g.defaultEvName = name
	return g
}

func (g *Group) DisableRoute(names ...string) *Group {
	g.disableRoutes = names
	return g
}

func (g *Group) Apply(models ...*ServiceModel) []*ServiceModel {
	for _, model := range models {
		// deployers
		if len(g.deployers) > 0 {
			model.SetDeployer(g.deployers...)
		}

		// midware
		if len(g.mws) > 0 {
			items := []*MWItem{}
			for _, mw := range g.mws {
				items = append(items, &MWItem{"fn-" + codekit.RandomString(5), mw})
			}
			items = append(items, model.mws...)
			model.mws = items
		}

		// postMws
		if len(g.postMws) > 0 {
			items := []*MWItem{}
			for _, fn := range g.postMws {
				items = append(items, &MWItem{"fn-" + codekit.RandomString(5), fn})
			}
			items = append(items, model.postMws...)
			model.postMws = items
		}

		// mods
		if len(g.mods) > 0 {
			model.SetMod(g.mods...)
		}

		// routes
		if len(g.onlyRoutes) > 0 {
			if len(model.allowOnlyRoutes) == 0 {
				model.AllowOnlyRoute(g.onlyRoutes...)
			} else {
				model.AllowOnlyRoute(lo.Filter(g.onlyRoutes, func(name string, index int) bool {
					return lo.IndexOf(model.allowOnlyRoutes, name) >= 0
				})...)

				//-- if no overlap between model allowroutes and group allowroute, will cause only routes len = 0, which is allowing everything, so need to allow to a non-existen
				if len(model.allowOnlyRoutes) == 0 {
					model.AllowOnlyRoute(codekit.RandomString(32))
				}
			}
		}

		// disable routes
		if len(g.disableRoutes) > 0 {
			model.DisableRoute(append(model.disableRoutes, g.disableRoutes...)...)
		}

		// set hub name
		if g.defaultHubName != "" {
			model.SetHubName(g.defaultHubName)
		}

		// set event name
		if g.defaultEvName != "" {
			model.SetEventName(g.defaultEvName)
		}
	}

	return models
}
