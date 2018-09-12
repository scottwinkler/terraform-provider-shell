package main

type Config struct {
}

type ShellClient struct {
}

// Client configures and returns a fully initialized ShellClient
func (c *Config) Client() (interface{}, error) {
	var client ShellClient
	return &client, nil
}
