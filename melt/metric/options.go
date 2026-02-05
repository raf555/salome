package metric

type recordOptions struct {
	label Labeler
}

type RecordOption func(*recordOptions)

func buildOptions(opts ...RecordOption) recordOptions {
	defaultOpts := recordOptions{
		label: NoLabel{},
	}

	for _, opt := range opts {
		opt(&defaultOpts)
	}

	return defaultOpts
}

func WithLabel(labeler Labeler) RecordOption {
	return func(ro *recordOptions) {
		if labeler == nil {
			return
		}

		ro.label = labeler
	}
}
