package logftxt

import "os"

// ---

// Domain provides external dependencies.
type Domain interface {
	Environment() Environment
	FS() FS

	private()
}

// DefaultDomain returns a default implementation of Domain that uses operating system bindings as dependencies.
func DefaultDomain() Domain {
	return defaultDomain()
}

// NewDomain constructs a new Domain based on the provided base but with some values overridden by the given options.
func NewDomain(base Domain, options ...domainOption) Domain {
	return domain{
		base.Environment(),
		base.FS(),
	}.with(options)
}

// ---

func defaultDomain() domain {
	return domain{
		Environment(os.LookupEnv),
		SystemFS(),
	}
}

// ---

type domain struct {
	env Environment
	fs  FS
}

func (d domain) Environment() Environment {
	return d.env
}

func (d domain) FS() FS {
	return d.fs
}

func (d domain) with(opts []domainOption) domain {
	for _, o := range opts {
		o.toDomain(&d)
	}

	return d
}

func (d domain) private() {}

// ---

type domainOption interface {
	toDomain(*domain)
}
