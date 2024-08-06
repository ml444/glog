package message

type Record struct {
	Pid        int    `json:"pid,omitempty"`
	RoutineID  int64  `json:"routine_id,omitempty"`
	Service    string `json:"module,omitempty"`
	Level      string `json:"level,omitempty"`
	Datetime   string `json:"datetime,omitempty"`
	Timestamp  int64  `json:"timestamp,omitempty"`
	CallerLine int    `json:"caller_line,omitempty"`
	CallerPath string `json:"caller_path,omitempty"`
	CallerName string `json:"caller_name,omitempty"`
	IP         string `json:"ip,omitempty"`
	HostName   string `json:"host,omitempty"`
	TraceID    string `json:"trade_id,omitempty"`
	Message    string `json:"msg,omitempty"`
}
