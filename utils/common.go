package utils

type CmdType string

const (
	REQ_OPENCHANNEL     CmdType = "REQ_OPENCHANNEL"
	REP_OPENCHANNEL             = "REP_OPENCHANNEL"
	MSG_ERROR                   = "MSG_ERROR"
	CONFIRM_OPENCHANNEL         = "CONFIRM_OPENCHANNEL"
	SUB                         = "SUB"
	MUL                         = "MUL"
	DIV                         = "DIV"
)

const PubKeySize = 66
