package tenantcoremodel

import (
	"github.com/samber/lo"
)

type Tags []TagItem

type TagItem struct {
	Key    string
	Values []string
}

func (t Tags) GetValues(key string) []string {
	for _, v := range t {
		if v.Key == key {
			return v.Values
		}
	}
	return []string{}
}

func (t Tags) Set(key, value string) Tags {
	for index, v := range t {
		if v.Key == key {
			for _, tv := range v.Values {
				if tv == value {
					return t
				}
			}
			v.Values = append(v.Values, value)
			t[index] = v
			return t
		}
	}

	t = append(t, TagItem{Key: key, Values: []string{value}})
	return t
}

func (t Tags) Remove(key, value string) Tags {
	for index, v := range t {
		if v.Key == key {
			for _, tv := range v.Values {
				if tv == value {
					return t
				}
			}
			v.Values = append(v.Values, value)
			t[index] = v
			return t
		}
	}

	return t
}

func (t Tags) RemoveKey(key string) Tags {
	newT := lo.Filter(t, func(obj TagItem, idx int) bool {
		return obj.Key != key
	})
	t = newT
	return t
}

func (t Tags) Check(key, value string) bool {
	for _, i := range t {
		if i.Key == key {
			for _, v := range i.Values {
				if v == value {
					return true
				}
			}
			return false
		}
	}
	return false
}
