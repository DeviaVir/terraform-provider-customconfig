package customconfig

type Config struct {
	TimeoutSeconds int
}

func (c *Config) loadAndValidate() error {
	return nil
}
