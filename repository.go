package light

type Repositories struct {
	Users UserRepository
}

type Tabler interface {
	TableName() string
	OnCreate() string
}
