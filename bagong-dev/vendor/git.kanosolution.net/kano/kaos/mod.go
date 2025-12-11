package kaos

type Mod interface {
	Name() string
	MakeGlobalRoute(svc *Service) ([]*ServiceRoute, error)
	MakeModelRoute(svc *Service, model *ServiceModel) ([]*ServiceRoute, error)
	//MakeEvent(svc *Service, model *EventModel, ev EventHub, disableRoutes ...string) error
}
