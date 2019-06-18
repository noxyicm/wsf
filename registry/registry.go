package registry

var container Container

func init() {
	container = Container{}
}

// Container represents maped values register
type Container map[string]interface{}

// Get returns registered value
func Get(key string) interface{} {
	if v, ok := container[key]; ok {
		return v
	}

	return nil
}

// GetBool returns registered value as bool
func GetBool(key string) bool {
	if v, ok := container[key]; ok {
		if v, ok := v.(bool); ok {
			return v
		}
	}

	return false
}

// GetInt returns registered value as int
func GetInt(key string) int {
	if v, ok := container[key]; ok {
		if v, ok := v.(int); ok {
			return v
		}
	}

	return 0
}

// GetString returns registered value as string
func GetString(key string) string {
	if v, ok := container[key]; ok {
		if v, ok := v.(string); ok {
			return v
		}
	}

	return ""
}

// Set sets new or resets old value
func Set(key string, value interface{}) {
	container[key] = value
}

// Has return true if key registered in container
func Has(key string) bool {
	if _, ok := container[key]; ok {
		return true
	}

	return false
}
