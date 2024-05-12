package config

type ReaderConfig struct {
	IPAddress string // IP address of the UDP source
	Port      int    // Port number of the UDP source
	ServiceID int
	ID        string
	Name      string
}

type WriterConfig struct {
	IPAddress string
	Port      int
	Name      string
	// ...
}

type Config struct {
	InputStreams []ReaderConfig
	OutputStream WriterConfig
	// ...
}

func (c *Config) Read(fname string) {
	// if err := c.read(...); err != nil {
	// 	// handle err
	// }

	// if err := c.parse(...); err != nil {
	// 	// handle err
	// }
}
