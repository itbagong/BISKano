package kaos

type RegisterOptions struct {
}

func MergeOpts(opts ...*RegisterOptions) *RegisterOptions {
	o := new(RegisterOptions)
	for _, m := range opts {
		o.merge(m)
	}
	return o
}

func (o *RegisterOptions) merge(m *RegisterOptions) {

}
