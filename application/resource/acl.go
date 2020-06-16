package resource

import (
	"wsf/acl"
	"wsf/config"
)

// TYPEAcl id of resource
const TYPEAcl = "acl"

func init() {
	Register(TYPEAcl, NewACLResource)
}

// NewACLResource creates a new resource of type ACL
func NewACLResource(cfg config.Config) (Interface, error) {
	typ := cfg.GetString("type")
	a, err := acl.NewACL(typ, cfg)
	if err != nil {
		return nil, err
	}

	acl.SetInstance(a)
	return a, nil
}
