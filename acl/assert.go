package acl

// Assert defines an assertion
type Assert interface {
	Assert(roleID string, resourceID string, privilege string) bool
}
