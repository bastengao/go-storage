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
