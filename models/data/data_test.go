package data

import "testing"

func TestData_FromBytes(t *testing.T) {
	testBytes := []byte{1, 3, 0, 21, 0, 0, 0, 0, 0, 0, 124, 0, 0, 0, 0, 116, 0, 0, 0, 0, 123, 0, 0, 7}

	data := Data{}

	err := data.FromBytes(testBytes)
	if err != nil {
		t.Error(err)
	}

	t.Log(data)
}
