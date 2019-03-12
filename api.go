package requests4go

// Get takes url and RequestArguments struct as parameters, it returns a Response struct.
// If you don't want use "RequestArguments", you can just pass nil,
// like: resp, err := Get("example.com", nil)
func Get(url string, args *RequestArguments) (*Response, error) {
	return sendRequest("GET", url, args)
}
