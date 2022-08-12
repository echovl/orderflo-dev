package assign

import (
	"reflect"
	"testing"
)

func String(s string) *string {
	return &s
}

func Int(i int) *int {
	return &i
}

func Float(f float64) *float64 {
	return &f
}

func TestAssignStructs(t *testing.T) {
	type A struct {
		S      string
		I      int
		F      float64
		Ptr    *string
		Struct struct{ S string }
		Arr    []int
		Map    map[string]string
		It     interface{}
	}

	type B struct {
		S      *string
		I      *int
		F      *float64
		Ptr    *string
		Struct *struct{ S string }
		Arr    []int
		Map    map[string]string
		It     interface{}
	}

	testcases := []struct {
		name   string
		target *A
		source *B
		want   *A
	}{
		{
			"int",
			&A{"a", 10, 3.2, String("b"), struct{ S string }{"c"}, []int{1, 2, 3}, nil, 10},
			&B{nil, Int(6), nil, nil, nil, nil, nil, nil},
			&A{"a", 6, 3.2, String("b"), struct{ S string }{"c"}, []int{1, 2, 3}, nil, 10},
		},
		{
			"string",
			&A{"a", 10, 3.2, String("b"), struct{ S string }{"c"}, []int{1, 2, 3}, nil, 10},
			&B{String("b"), nil, nil, nil, nil, nil, nil, nil},
			&A{"b", 10, 3.2, String("b"), struct{ S string }{"c"}, []int{1, 2, 3}, nil, 10},
		},
		{
			"stringPtr",
			&A{"a", 10, 3.2, String("b"), struct{ S string }{"c"}, []int{1, 2, 3}, nil, 10},
			&B{nil, nil, nil, String("c"), nil, nil, nil, nil},
			&A{"a", 10, 3.2, String("c"), struct{ S string }{"c"}, []int{1, 2, 3}, nil, 10},
		},
		{
			"structPtr",
			&A{"a", 10, 3.2, String("b"), struct{ S string }{"c"}, []int{1, 2, 3}, nil, 10},
			&B{nil, nil, nil, nil, &struct{ S string }{"z"}, nil, nil, nil},
			&A{"a", 10, 3.2, String("b"), struct{ S string }{"z"}, []int{1, 2, 3}, nil, 10},
		},
		{
			"slice",
			&A{"a", 10, 3.2, String("b"), struct{ S string }{"c"}, []int{1, 2, 3}, nil, 10},
			&B{nil, nil, nil, nil, nil, []int{9, 9}, nil, nil},
			&A{"a", 10, 3.2, String("b"), struct{ S string }{"c"}, []int{9, 9}, nil, 10},
		},
		{
			"map",
			&A{"a", 10, 3.2, String("b"), struct{ S string }{"c"}, []int{1, 2, 3}, map[string]string{"a": "b"}, 10},
			&B{nil, nil, nil, nil, nil, nil, map[string]string{"c": "d"}, nil},
			&A{"a", 10, 3.2, String("b"), struct{ S string }{"c"}, []int{1, 2, 3}, map[string]string{"c": "d"}, 10},
		},
		{
			"interface",
			&A{"a", 10, 3.2, String("b"), struct{ S string }{"c"}, []int{1, 2, 3}, nil, 10},
			&B{nil, nil, nil, nil, nil, nil, nil, "changed"},
			&A{"a", 10, 3.2, String("b"), struct{ S string }{"c"}, []int{1, 2, 3}, nil, "changed"},
		},
		{
			"multiple",
			&A{"a", 10, 3.2, String("b"), struct{ S string }{"c"}, []int{1, 2, 3}, nil, 10},
			&B{String("b"), Int(99), Float(3.33), String("zz"), &struct{ S string }{"z"}, []int{1, 2}, nil, 99},
			&A{"b", 99, 3.33, String("zz"), struct{ S string }{"z"}, []int{1, 2}, nil, 99},
		},
		{
			"nochange",
			&A{"a", 10, 3.2, String("b"), struct{ S string }{"c"}, []int{1, 2, 3}, nil, 10},
			&B{nil, nil, nil, nil, nil, nil, nil, nil},
			&A{"a", 10, 3.2, String("b"), struct{ S string }{"c"}, []int{1, 2, 3}, nil, 10},
		},
	}

	for _, tc := range testcases {
		err := Structs(tc.target, tc.source)
		if err != nil {
			t.Error(err)
		}

		if ok := reflect.DeepEqual(*tc.target, *tc.want); !ok {
			t.Errorf("%s:\ngot: %v\nwant: %v", tc.name, tc.target, tc.want)
		}
	}
}
