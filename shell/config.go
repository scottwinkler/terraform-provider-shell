package shell

//Config is the config for the client.
type Config struct {
}

//Client is the client itself. Since we already have access to the shell no real provisioning needs to be done
type Client struct {
}

// Client configures and returns a fully initialized ShellClient
func (c *Config) Client() (interface{}, error) {
	var client Client
	return &client, nil
}
