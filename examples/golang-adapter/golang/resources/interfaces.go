package resources

type IDataResource interface {
	Read()
	State() map[string]string
}

type IResource interface {
	Create()
	Read()
	Update()
	Delete()
	State() map[string]string
}
