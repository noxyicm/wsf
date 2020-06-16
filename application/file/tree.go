package file

import (
	"wsf/utils"
)

// Tree represents a file tree
type Tree map[string]interface{}

// Push pushes new file upload into it's proper place
func (f Tree) Push(k string, v *File) {
	keys := utils.FetchIndexes(k)
	if len(keys) <= utils.MaxTreeLevel {
		f.Mount(keys, v)
	}
}

// Get retreives a file from tree
func (f Tree) Get(k string) *File {
	keys := utils.FetchIndexes(k)
	if len(keys) <= utils.MaxTreeLevel {
		if v, ok := f.Unmount(keys).(*File); ok {
			return v
		}
	}

	return nil
}

// Mount mounts data tree recursively
func (f Tree) Mount(i []string, v *File) {
	if len(i) == 1 {
		f[i[0]] = v
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

// Unmount retrives data from tree recursively
func (f Tree) Unmount(i []string) interface{} {
	if len(i) == 1 {
		return f[i[0]]
	}

	if len(i) == 2 && i[1] == "" {
		return f[i[0]]
	}

	if p, ok := f[i[0]]; ok {
		return p.(Tree).Unmount(i[1:])
	}

	return nil
}
