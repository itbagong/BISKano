package kaos

import (
	"strings"

	"github.com/sebarcode/codekit"
)

func MergeMs(ms ...codekit.M) codekit.M {
	m := codekit.M{}
	for _, me := range ms {
		for k, v := range me {
			m.Set(k, v)
		}
	}
	return m
}

func EndPointName(name string) string {
	if NamingType == NamingAsIs {
		name = strings.Replace(name, "_", "/", -1)
		return name
	}

	names := []string{}
	buffer := ""
	prevIsLower := false
	currentIsUpper := false
	multiUpper := false
	for cIndex, c := range name {
		if cIndex == 0 {
			buffer += string(c)
			currentIsUpper = c >= 65 && c <= 90
		} else {
			prevIsLower = !currentIsUpper
			currentIsUpper = c >= 65 && c <= 90

			if prevIsLower && currentIsUpper {
				names = append(names, buffer)
				buffer = ""
				multiUpper = false
			} else if currentIsUpper && !prevIsLower {
				multiUpper = true
			} else if !currentIsUpper && !prevIsLower && multiUpper {
				names = append(names, buffer)
				buffer = ""
				multiUpper = false
			}
			buffer += string(c)
		}
	}
	names = append(names, buffer)

	for idx, nm := range names {
		nm := strings.ToLower(nm)
		names[idx] = nm
	}

	name = strings.Join(names, NamingJoiner)
	name = strings.Replace(name, "_", "/", -1)
	name = strings.Replace(name, "/"+NamingJoiner, "/", -1)

	return name
}
