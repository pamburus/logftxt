package logftxt

import "os"

// Environment is an environment variable lookup function.
type Environment func(string) (string, bool)

func (f Environment) toFSEnvOptions(oo *fsEnvOptions) {
	oo.env = f
}

func (f Environment) toAppenderOptions(oo *appenderOptions) {
	oo.env = f
}

// ---

type fsEnvOption interface {
	toFSEnvOptions(*fsEnvOptions)
}

// ---

func defaultEnvOptions() envOptions {
	return envOptions{os.LookupEnv}
}

type envOptions struct {
	env Environment
}

// ---

func defaultFSEnvOptions() fsEnvOptions {
	return fsEnvOptions{
		fsOptions{DefaultFS()},
		defaultEnvOptions(),
	}
}

type fsEnvOptions struct {
	fsOptions
	envOptions
}

func (o fsEnvOptions) With(other []fsEnvOption) fsEnvOptions {
	for _, oo := range other {
		oo.toFSEnvOptions(&o)
	}

	return o
}
