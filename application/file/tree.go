package file

import (
	"wsf/utils"
)

// Tree represents a file tree
type Tree map[string]interface{}

// Push pushes new file upload into it's proper place
func (f Tree) push(k string, v []*File) {
	keys := utils.FetchIndexes(k)
	if len(keys) <= utils.MaxTreeLevel {
		f.Mount(keys, v)
	}
}

// Mount mounts data tree recursively
func (f Tree) Mount(i []string, v []*File) {
	if len(i) == 1 {
		f[i[0]] = v[0]
		return
	}

	if len(i) == 2 && i[1] == "" {
		f[i[0]] = v
		return
	}

	if p, ok := f[i[0]]; ok {
		p.(Tree).Mount(i[1:], v)
		return
	}

	f[i[0]] = make(Tree)
	f[i[0]].(Tree).Mount(i[1:], v)
}
