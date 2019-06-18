package storage

// Interface is a session storage interface
type Interface interface {
	RememberMe(seconds int)
	ForgetMe()
	RememberUntil(seconds int)
}
