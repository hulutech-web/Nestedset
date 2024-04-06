package contracts

type Nestedset interface {
	ToTree(model interface{}) (interface{}, error)
	CreateTree(modelPrt interface{}) error
	GetTree(model interface{}) (interface{}, error)
	AppendChild(model interface{}, pid int) error
	RemoveNode(model interface{}) error
	PrettyPrint(data interface{})
}
