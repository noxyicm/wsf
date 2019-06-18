package environment

const (
	// ID of service
	ID = "env"

	// ENV defines an environment prefix
	ENV = "DC"
)

// Service provides ability to map _ENV values from config file
type Service struct {
	values map[string]string
}

// Init environment service
func (s *Service) Init(cfg *Config) (bool, error) {
	if s.values == nil {
		s.values = make(map[string]string)
		s.values[ENV] = "true"
	}

	for k, v := range cfg.Values {
		s.values[ENV+"_"+k] = v
	}

	return true, nil
}

// GetEnv must return list of environment variables
func (s *Service) GetEnv() (map[string]string, error) {
	return s.values, nil
}

// SetEnv sets or creates environment variable
func (s *Service) SetEnv(key, value string) {
	s.values[ENV+"_"+key] = value
}

// Copy all environment variables to setter
func (s *Service) Copy(setter Setter) error {
	values, err := s.GetEnv()
	if err != nil {
		return err
	}

	for k, v := range values {
		setter.SetEnv(k, v)
	}

	return nil
}

// NewService creates new env service instance
func NewService(defaults map[string]string) *Service {
	s := &Service{values: defaults}
	return s
}
