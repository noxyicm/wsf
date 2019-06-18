package helper

// Interface maybe?
type Interface interface {
	SetView(vi ViewInterface) error
}

// ViewInterface is a halper view interface interpretation
type ViewInterface interface {
	RegisterHelper(name string, hlp Interface) error
	Helper(name string) Interface
}
