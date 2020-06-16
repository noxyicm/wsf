package session

import (
	"wsf/config"
	"wsf/errors"
	"wsf/session/validator"
)

// ManagerConfig defines set of session manager variables
type ManagerConfig struct {
	Type                    string `json:"type"`
	Priority                int
	SessionName             string              `json:"sessionName"`
	SessionIDLength         int                 `json:"sessionIDLength"`
	SessionIDPrefix         string              `json:"sessionIDPrefix"`
	SessionNameInHTTPHeader string              `json:"SessionNameInHTTPHeader"`
	SessionLifeTime         int64               `json:"sessionLifeTime"`
	HTTPOnly                bool                `json:"HTTPOnly"`
	Secure                  bool                `json:"secure"`
	EnableSetCookie         bool                `json:"enableSetCookie,omitempty"`
	EnableSidInHTTPHeader   bool                `json:"EnableSidInHTTPHeader"`
	EnableSidInURLQuery     bool                `json:"EnableSidInURLQuery"`
	Strict                  bool                `json:"strict"`
	Storage                 config.Config       `json:"storage"`
	Session                 config.Config       `json:"session"`
	Valds                   []*validator.Config `json:"validators"`
}

// Populate populates Config values using given Config source
func (c *ManagerConfig) Populate(cfg config.Config) error {
	if scfg := cfg.Get("handler"); scfg != nil {
		c.Session = scfg
	}

	if c.Session == nil {
		c.Session = config.NewBridge()
	}

	if scfg := cfg.Get("storage"); scfg != nil {
		c.Storage = scfg
	}

	if c.Storage == nil {
		c.Storage = config.NewBridge()
	}

	if vscfg := cfg.Get("validators"); vscfg != nil {
		for _, k := range vscfg.GetKeys() {
			validatorCfg := &validator.Config{}
			validatorCfg.Defaults()
			validatorCfg.Populate(vscfg.Get(k))

			c.Valds = append(c.Valds, validatorCfg)
		}
	}

	if err := cfg.Unmarshal(c); err != nil {
		return err
	}

	return c.Valid()
}

// Defaults sets configuration default values
func (c *ManagerConfig) Defaults() error {
	c.Type = "default"
	c.Priority = 15
	c.SessionName = "WSFSESS"
	c.EnableSetCookie = false
	c.EnableSidInHTTPHeader = false
	c.EnableSidInURLQuery = false
	c.SessionIDLength = 16
	c.SessionLifeTime = 900
	c.Valds = make([]*validator.Config, 0)

	if c.Session == nil {
		c.Session = config.NewBridge()
	}

	if c.Storage == nil {
		c.Storage = config.NewBridge()
	}

	return nil
}

// Valid validates the configuration
func (c *ManagerConfig) Valid() error {
	if c.Session == nil {
		return errors.New("Invalid session configuration")
	}

	if c.Storage == nil {
		return errors.New("Invalid storage configuration")
	}

	return nil
}
