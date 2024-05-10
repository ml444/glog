package message

type Record struct {
	CallerLine int    `json:"caller_line,omitempty"`
	Pid        int    `json:"pid,omitempty"`
	RoutineID  int64  `json:"routine_id,omitempty"`
	Module     string `json:"module,omitempty"`
	Level      string `json:"level,omitempty"`
	Datetime   string `json:"datetime,omitempty"`
	Timestamp  int64  `json:"timestamp,omitempty"`
	FileName   string `json:"file,omitempty"`
	CallerName string `json:"caller_name,omitempty"`
	IP         string `json:"ip,omitempty"`
	HostName   string `json:"host,omitempty"`
	TraceID    string `json:"trade_id,omitempty"`
	Message    string `json:"msg,omitempty"`
	ErrMsg     string `json:"err_msg,omitempty"`
}