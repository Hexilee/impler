package impl

const (
	TypeInt ParamType = iota
	TypeString
	IOReader
	TypeFile
	Other
)

type (
	ParamType int
)