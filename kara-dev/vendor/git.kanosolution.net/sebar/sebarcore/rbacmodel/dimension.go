package rbacmodel

import (
	"fmt"
	"sort"
	"strings"
)

type DimensionItem struct {
	Kind  string
	Value string
}

type DimensionItems []DimensionItem

type Dimension struct {
	Items DimensionItems
	Hash  string
}

func (dm DimensionItems) String() string {
	if len(dm) == 0 {
		return ""
	}

	sort.Sort(dm)
	ms := make([]string, len(dm))
	for idx, m := range dm {
		ms[idx] = fmt.Sprintf("%s:%s", m.Kind, m.Value)
	}

	return strings.Join(ms, "_")
}

func (dm DimensionItems) Len() int {
	return len(dm)
}

func (dm DimensionItems) Less(i, j int) bool {
	return dm[i].Kind < dm[j].Kind
}

func (dm DimensionItems) Swap(i, j int) {
	dm[i], dm[j] = dm[j], dm[i]
}
