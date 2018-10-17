package impl

const (
	// <Annotation Val> first annotations. etc. @Get /item/{id} | @Get
	GetAnn        = "@Get" // path
	HeadAnn       = "@Head"
	PostAnn       = "@Post"
	PutAnn        = "@Put"
	PatchAnn      = "@Patch"
	DeleteAnn     = "@Delete"
	ConnectAnn    = "@Connect"
	OptionsAnn    = "@Options"
	TraceAnn      = "@Trace"
	BodyAnn       = "@Body"       // json | xml | form | multipart; default json
	SingleBodyAnn = "@SingleBody" // json | xml | form | multipart; default json; if singleBody, the type of single body var must be IOReader or Other
	ResultAnn     = "@Result"     // json | xml ; default json
	BaseAnn       = "@Base"
)

const (
	// <Annotation(Key) Val> second annotations. etc. @Header(Content-Type) multipart/form | @Header(Content-Type)
	ParamAnn  = "@Param"
	HeaderAnn = "@Header" // param type: string
	CookieAnn = "@Cookie" // param type: string
	FileAnn   = "@File"   // param type: string
)