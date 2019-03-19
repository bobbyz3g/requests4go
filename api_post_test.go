package requests4go

import (
	"testing"
)

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
