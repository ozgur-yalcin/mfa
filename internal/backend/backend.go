package backend

type Backend interface {
	Engine() string
	Params() string
}
