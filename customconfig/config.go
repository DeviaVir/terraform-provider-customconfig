package customconfig

type Config struct {
	TimeoutMinutes int
}

func (c *Config) loadAndValidate() error {
	return nil
}
