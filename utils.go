package requests4go

const (
	defaultUserAgent   = "Request4go"
	defaultContentType = "application/x-www-form-urlencoded; charset=utf-8"
	defaultJsonType    = "application/json; charset=utf-8"
)

var defaultHeaders = map[string]string{
	"Connection":      "keep-alive",
	"Accept-Encoding": "gzip, deflate",
	"Accept":          "*/*",
	"User-Agent":      defaultUserAgent,
}
