package flipt

type Namespaced interface {
	// Namespace returns the namespace of the entity
	GetNamespaceKey() string
}

type BatchNamespaced interface {
	GetNamespaceKeys() []string
}
