package goforarun

type LoggingConfig struct {
	// Level is the log level (debug, info, warn, error, fatal, panic)
	Level string `yaml:"level"`
	// Format is the log format (text, json)
	Format string `yaml:"format"`
	// Output is the log output (stdout, stderr, file)
	Output string `yaml:"output"`
	// FilteredChannels is the list of channels to filter [channel: level]
	FilteredChannels map[string]string `yaml:"filtered_channels"`
}
