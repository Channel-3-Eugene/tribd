package main

type ReaderConfig struct {
	IPAddress string      // IP address of the UDP source
	Port      int         // Port number of the UDP source
	ServiceID int
	ID        string
	Name      string
}

type Config struct {
	ReaderConfig ReaderConfig
	// ...
}

func (c *Config) Read(fname string) {
	if err := c.read(...); err != nil {
		// handle err
	}
	
	if err := c.parse(...); err != nil {
		// handle err
	}
}