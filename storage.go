package storage

type Storage interface {
	// TODO: Service
	Service() Service
	Variant(key string, options VariantOptions) Variant
}

type storage struct {
	service        Service
	variantFactory VariantFactory
}

// New creates a new storage. If variantFactory is nil, NewVariantFactory(NewTransformer()) will be used.
func New(service Service, variantFactory VariantFactory) Storage {
	if variantFactory == nil {
		variantFactory = NewVariantFactory(NewTransformer())
	}

	return &storage{
		service:        service,
		variantFactory: variantFactory,
	}
}

func (s *storage) Service() Service {
	return s.service
}

func (s *storage) Variant(key string, options VariantOptions) Variant {
	return s.variantFactory.NewVariant(s.service, key, options)
}
