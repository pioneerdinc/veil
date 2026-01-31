package generator

// Options holds configuration for secret generation
type Options struct {
	Type      string // password, apikey, jwt
	Length    int
	Format    string // uuid, hex, base64 (for apikey)
	Prefix    string // For API keys
	Bits      int    // For JWT
	NoSymbols bool   // For passwords
	ToEnv     string // Path to .env file
	Force     bool   // Overwrite existing in .env
}
