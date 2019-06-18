package environment

// Interface aggregates list of environment variables.
// This interface can be used in custom implementation to drive
// values from external sources.
type Interface interface {
	Setter
	Getter
	Copy(setter Setter) error
}

// Setter provides ability to set environment variable
type Setter interface {
	SetEnv(key, value string)
}

// Getter provides ability to get environment variables
type Getter interface {
	GetEnv() (map[string]string, error)
}
