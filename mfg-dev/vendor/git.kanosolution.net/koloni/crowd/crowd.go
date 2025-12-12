/*Package crowd is a go function to handle data pipelining process*/
package crowd

import "reflect"

func makeChannel(t reflect.Type, chanDir reflect.ChanDir, buffer int) reflect.Value {
	ctype := reflect.ChanOf(chanDir, t)
	return reflect.MakeChan(ctype, buffer)
}
