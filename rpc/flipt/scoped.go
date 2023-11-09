package flipt

type Namespaced interface {
	// Namespace returns the namespace of the entity
	GetNamespaceKey() string
}

func (req *GetNamespaceRequest) GetNamespaceKey() string {
	return req.Key
}

func (req *DeleteNamespaceRequest) GetNamespaceKey() string {
	return req.Key
}

func (req *UpdateNamespaceRequest) GetNamespaceKey() string {
	return req.Key
}
