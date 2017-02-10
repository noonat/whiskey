package prefork

// Logger is the interface required for logging performed by this package.
type Logger interface {
	Printf(fmt string, args ...interface{})
	Println(args ...interface{})
	SetPrefix(prefix string)
}
