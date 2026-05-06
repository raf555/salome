package prommetric

// Label is a convenient wrapper of anything.
// Zero Label can provide label names with Labels().
type Label interface {
	// Labels returns label names of this Label.
	// It should be a static slice with fixed length.
	Labels() []string

	// Values returns values of this Label.
	// The length is the same as returned by Labels.
	Values() []string
}

type NoLabel struct{}

// Labels implements [Label].
func (n NoLabel) Labels() []string {
	return nil
}

// Values implements [Label].
func (n NoLabel) Values() []string {
	return nil
}

var _ Label = NoLabel{}
