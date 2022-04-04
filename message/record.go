package message

type Record struct {
	Module    string `json:"module"`
	Level     string `json:"level"`
	Datetime  string `json:"datetime"`
	Timestamp int64  `json:"timestamp"`
	FileName  string `json:"file"`
	//CallerLine string		 `json:"line"`
	FuncName string      `json:"func"`
	Message  interface{} `json:"msg"`
	ErrMsg   string      `json:"err_msg"`
}

type ReportMsg struct {
	PublicFields Sys         `json:"public_fields"`
	ExtFields    interface{} `json:"ext_fields"`
	Type         string      `json:"type"`
	SubType      string      `json:"sub_type"`
}

type Sys struct {
	Ip          string `json:"ip"`
	TMs         uint64 `json:"t_ms"`
	ServiceName string `json:"service_name"`
	ReqId       string `json:"req_id"`
	CorpId      uint32 `json:"corp_id"`
	HostName    string `json:"host_name"`
	CallerName  string `json:"caller_name"`
	CallerLine  int    `json:"caller_line"`
}
