package storage

type VariantFactory interface {
	NewVariant(service Service, originPath string, options VariantOptions) Variant
}

type NewVariantFunc func(service Service, originKey string, options VariantOptions) Variant

func (f NewVariantFunc) NewVariant(service Service, originPath string, options VariantOptions) Variant {
	return f(service, originPath, options)
}

type variantFactory struct {
	transformer Transformer
}

func NewVariantFactory(transformer Transformer) VariantFactory {
	return variantFactory{
		transformer: transformer,
	}
}

func (f variantFactory) NewVariant(service Service, originPath string, options VariantOptions) Variant {
	return NewVariant(service, originPath, options, f.transformer)
}
