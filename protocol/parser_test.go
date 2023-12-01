package protocol

import (
	"fmt"
	"testing"
)

func TestParseSingleLine(t *testing.T) {
	msg1 := []byte("+OK\r\n")
	msg2 := []byte("-Error message\r\n")
	msg3 := []byte(":1000\r\n")
	msg4 := []byte(":-20\r\n")

	resp1, err := parseSingleLine(msg1)
	if string(resp1.GetBytesData()) != "OK" || err != nil {
		t.Error(fmt.Sprintf("Protocal error: %s", string(msg1)))
	}
	resp2, err := parseSingleLine(msg2)
	if string(resp2.GetBytesData()) != "Error message" || err != nil {
		t.Error(fmt.Sprintf("Protocal error: %s", string(msg2)))
	}
	resp3, err := parseSingleLine(msg3)
	if string(resp3.GetBytesData()) != "1000" || err != nil {
		t.Error(fmt.Sprintf("Protocal error: %s", string(msg3)))
	}
	resp4, err := parseSingleLine(msg4)
	if string(resp4.GetBytesData()) != "-20" || err != nil {
		t.Error(fmt.Sprintf("Protocal error: %s", string(msg4)))
	}
}
