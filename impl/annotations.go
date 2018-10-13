package impl

import (
	"bufio"
	"bytes"
	"regexp"
)

const (
	// <Annotation Val> first annotations. etc. @Get /item/{id}
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
	SingleBodyAnn = "@SingleBody" // json | xml | form | multipart; default json; only
	ResultAnn     = "@Result"     // json | xml ; default json
	BaseAnn       = "@Base"
)

const (
// <Annotation(Key)> second annotations. etc. @FilePath(path)
)

const (
	// <Annotation(Key) Val> third annotations. etc. @Header(Content-Type) multipart/form
	ParamAnn  = "@Param"
	HeaderAnn = "@Header" // param type: string
	CookieAnn = "@Cookie" // param type: string
	FileAnn   = "@File"   // param type: string
)

type (
	Processor struct {
		*bufio.Scanner
	}
)

const (
	FirstAnnRegex  = `(@[a-zA-Z_][0-9a-zA-Z_]*)\s+(.+)`
	SecondAnnRegex = `(@[a-zA-Z_][0-9a-zA-Z_]*)\((.+?)\)`
	ThirdAnnRegex  = `(@[a-zA-Z_][0-9a-zA-Z_]*)\((.+?)\)\s+(.+)`
)

var (
	FirstAnnRe  = regexp.MustCompile(FirstAnnRegex)
	SecondAnnRe = regexp.MustCompile(SecondAnnRegex)
	ThirdAnnRe  = regexp.MustCompile(ThirdAnnRegex)
)

func NewProcessor(text string) *Processor {
	return &Processor{bufio.NewScanner(bytes.NewBufferString(text))}
}

func (process *Processor) Scan(fn func(ann, key, value string) (err error)) (err error) {
	for process.Scanner.Scan() {
		text := process.Text()
		if values := ThirdAnnRe.FindStringSubmatch(text); len(values) == 4 {
			err = fn(values[1], values[2], values[3])
		} else if values := FirstAnnRe.FindStringSubmatch(text); len(values) == 3 {
			err = fn(values[1], ZeroStr, values[2])
		} else if values := SecondAnnRe.FindStringSubmatch(text); len(values) == 3 {
			err = fn(values[1], values[2], ZeroStr)
		}

		if err != nil {
			break
		}
	}
	return
}
