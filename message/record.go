package message

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type Record struct {
	Module     string      `json:"module,omitempty"`
	Level      string      `json:"level,omitempty"`
	Datetime   string      `json:"datetime,omitempty"`
	Timestamp  int64       `json:"timestamp,omitempty"`
	FileName   string      `json:"file,omitempty"`
	CallerName string      `json:"caller_name,omitempty"`
	CallerLine int         `json:"caller_line,omitempty"`
	Pid        int         `json:"pid,omitempty"`
	RoutineId  int64       `json:"routine_id,omitempty"`
	Ip         string      `json:"ip,omitempty"`
	HostName   string      `json:"host,omitempty"`
	TradeId    string      `json:"trade_id,omitempty"`
	Message    interface{} `json:"msg,omitempty"`
	ErrMsg     string      `json:"err_msg,omitempty"`
}

func (r *Record) Bytes(disableHTMLEscape, prettyPrint bool) ([]byte, error) {
	b := &bytes.Buffer{}
	encoder := json.NewEncoder(b)
	encoder.SetEscapeHTML(!disableHTMLEscape)
	if prettyPrint {
		encoder.SetIndent("", "  ")
	}
	if err := encoder.Encode(r); err != nil {
		return nil, fmt.Errorf("failed to encoding record to JSON, %w", err)
	}
	return b.Bytes(), nil
}
