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
	if age, _ := JSON.Get("json").Get("name").Int(); jsonStruct.Age != age {
		t.Errorf("Json value error: excepted %v, got %v", jsonStruct.Age, age)
	}
}
