package impl

const (
	Int ParamType = iota
	String
	IOReader
	File
	Other
)

type (
	ParamType int
)