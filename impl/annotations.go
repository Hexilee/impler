package impl

const (
	// <Annotation Val> annotations. etc. @Get /item/{id}
	GetAnn     = "@Get" // path
	HeadAnn    = "@Head"
	PostAnn    = "@Post"
	PutAnn     = "@Put"
	PatchAnn   = "@Patch"
	DeleteAnn  = "@Delete"
	ConnectAnn = "@Connect"
	OptionsAnn = "@Options"
	TraceAnn   = "@Trace"
	BodyAnn    = "@Body" // json | xml | form | multipart
)

const (
// <Annotation(Key)> annotations. etc. @FilePath(path)
)

const (
	// <Annotation(Key) Val> annotations. etc. @Header(Content-Type) multipart/form
	Param    = "@Param"
	OnlyBody = "@OnlyBody" // json | xml
	Header   = "@Header"   // param type: string
	Cookie   = "@Cookie"   // param type: string
	FilePath = "@FilePath" // param type: string
)

type (
	Annotation string
)
