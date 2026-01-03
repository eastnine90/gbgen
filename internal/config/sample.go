package config

// Sample returns a sample config intended for writing to disk via `gbgen init`.
// It includes a placeholder API key value.
func Sample() Config {
	cfg := Defaults()
	cfg.GrowthBook.APIKey = "secret_***"
	return cfg
}
