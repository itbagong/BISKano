package kaos

type MWFunc func(*Context, interface{}) (bool, error)

type MWItem struct {
	Name string
	Fn   MWFunc
}

func (s *Service) RegisterMW(fn MWFunc, name string) *Service {
	s.mws = append(s.mws, &MWItem{name, fn})
	return s
}

func (s *Service) RegisterPostMW(fn MWFunc, name string) *Service {
	s.postMWS = append(s.postMWS, &MWItem{name, fn})
	return s
}

func (s *Service) Middlewares() []*MWItem {
	return s.mws
}

func (s *Service) PostMiddlewares() []*MWItem {
	return s.postMWS
}
