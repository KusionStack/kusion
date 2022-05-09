package diff

// Option is a function used as arguments to New in order to configure the resulting diff Options.
type Option func(*Options)

// Options holds diffing settings
type Options struct {
	// If set to true then differences caused by aggregated roles in RBAC resources are ignored.
	ignoreAggregatedRoles bool
	normalizer            Normalizer
}

func applyOptions(opts []Option) Options {
	o := Options{
		ignoreAggregatedRoles: false,
		normalizer:            GetNoopNormalizer(),
	}
	for _, opt := range opts {
		opt(&o)
	}
	return o
}

func WithNormalizer(normalizer Normalizer) Option {
	return func(o *Options) {
		o.normalizer = normalizer
	}
}
