package logftxt

// Environment is an environment variable lookup function.
type Environment func(string) (string, bool)

func (f Environment) toAppenderOptions(oo *appenderOptions) {
	oo.env = f
}

func (f Environment) toEncoderOptions(oo *encoderOptions) {
	oo.env = f
}

func (f Environment) toDomain(d *domain) {
	d.env = f
}
