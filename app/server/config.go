// Package server
package server

type Config struct {
	Addr            string `env:"ADDR" envDefault:":3300"`
	Database        string `env:"DATABASE,required"`
	FilePath        string `env:"FilePath" envDefault:""`
	Secret          string `env:"SECRET,required"`
	TokenExpiration uint   `env:"TOKEN_EXPIRATION" envDefault:"3600"`
	Memory          uint   `env:"ARGON_MEMORY" envDefault:"64"`
	Iterations      uint   `env:"ARGON_ITERATIONS" envDefault:"3"`
	Parallelism     uint   `env:"ARGON_PARALLELISM" envDefault:"2"`
	SaltLength      uint   `env:"ARGON_SALT_LENGTH" envDefault:"16"`
	KeyLength       uint   `env:"ARGON_KEY_LENGTH" envDefault:"32"`
}
