package requests4go

import (
	"os"
	"testing"
)

var testMap = map[string]string{
	"name": "name",
	"age":  "age",
}

func TestBaseGet(t *testing.T) {
	resp, err := Get("http://httpbin.org/get", nil)

	if err != nil {
		t.Errorf("Get error: excepted no error, got %v", err)
	}

	if !resp.Ok() {
		t.Errorf("Request got wrong response, excepted \"200 OK\", got \"%v\"", resp.Status)
	}
}

func TestCookieGet(t *testing.T) {
	args := DefaultRequestArguments
	args.Cookies = testMap

	resp, err := Get("http://httpbin.org/cookies", args)

	if err != nil {
		t.Fatalf("Request error: %v", err)
	}

	JSON, err := resp.Json()
	if err != nil {
		t.Fatalf("Get json error: %v", err)
	}

	if name, _ := JSON.Get("cookies").Get("name").String(); name != testMap["name"] {
		t.Errorf("Cookies set error: excepted %v, got %v", testMap["name"], name)
	}

	if age, _ := JSON.Get("cookies").Get("age").String(); age != testMap["age"] {
		t.Errorf("Cookies set error: excepted %v, got %v", testMap["age"], age)
	}
}

func TestBaseJsonPost(t *testing.T) {
	jsonStruct := struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}{
		"name",
		21,
	}
	args := DefaultRequestArguments
	args.Json = jsonStruct

	resp, err := Post("http://httpbin.org/post", args)
	if err != nil {
		t.Errorf("Reqeust erro: got %v", err)
	}

	JSON, err := resp.Json()
	if err != nil {
		t.Errorf("%v \n", err)
	}
	if name, _ := JSON.Get("json").Get("name").String(); jsonStruct.Name != name {
		t.Errorf("Json value error: excepted %v, got %v", jsonStruct.Name, name)
	}
	if age, _ := JSON.Get("json").Get("age").Int(); jsonStruct.Age != age {
		t.Errorf("Json value error: excepted %v, got %v", jsonStruct.Age, age)
	}
}

func TestBaseDataPost(t *testing.T) {
	args := DefaultRequestArguments
	args.Data = map[string]string{
		"name": "name",
		"age":  "21",
	}

	resp, err := Post("http://httpbin.org/post", args)
	if err != nil {
		t.Errorf("Request error: got %v", err)
	}

	JSON, err := resp.Json()
	if err != nil {
		t.Errorf("Get json error: %v", err)
	}

	if name, _ := JSON.Get("form").Get("name").String(); name != args.Data["name"] {
		t.Errorf("Json value error: excepted %v, got %v", args.Data["name"], name)
	}

	if age, _ := JSON.Get("form").Get("age").String(); age != args.Data["age"] {
		t.Errorf("Json value error: excepted %v, got %v", args.Data["age"], age)
	}
}

func TestBaseFilePost(t *testing.T) {
	args := DefaultRequestArguments
	f, err := os.Open("testdata/file4upload")

	if err != nil {
		t.Fatalf("Open file error: %v", err)
	}

	f2, err := os.Open("testdata/file2")

	if err != nil {
		t.Fatalf("Open file error: %v", err)
	}

	files := []FileField{
		{
			FileName:    "file4upload",
			FieldName:   "file",
			FileContent: f,
		},
		{
			FileName:    "file2",
			FieldName:   "file2",
			FileContent: f2,
		},
	}

	args.Files = files
	resp, err := Post("http://www.httpbin.org/post", args)
	if err != nil {
		t.Errorf("Reqeust error: %v", err)
	}
	fString1 := "Hey, I am test file."
	fString2 := "Hey, I am test file too."
	JSON, err := resp.Json()
	if err != nil {
		t.Fatalf("Get response JSON error: %v", err)
	}

	if file1, _ := JSON.Get("files").Get("file").String(); file1 != fString1 {
		t.Errorf("Post file error: excepted \"%v\", got \"%v\"", fString1, file1)
	}

	if file2, _ := JSON.Get("files").Get("file2").String(); file2 != fString2 {
		t.Errorf("Post file error: excepted \"%v\", got \"%v\"", fString2, file2)
	}
}

func TestBaseBodyPost(t *testing.T) {
	data := map[string]string{
		"name": "name",
		"age":  "21",
	}

	body, _ := prepareDataBody(data)

	args := DefaultRequestArguments
	args.Body = body

	resp, err := Post("http://httpbin.org/post", args)
	if err != nil {
		t.Fatalf("Request error: got %v", err)
	}

	JSON, err := resp.Json()
	if err != nil {
		t.Errorf("Get json error: %v", err)
	}

	excepted := "age=21&name=name"
	if got, _ := JSON.Get("data").String(); got != excepted {
		t.Errorf("Post Body error: excepted: %v, got %v", excepted, got)
	}
}
