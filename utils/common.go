package utils

type CmdType string

const (
	REQ_OPENCHANNEL CmdType = "REQ_OPENCHANNEL"
	REP_OPENCHANNEL         = "REP_OPENCHANNEL"
	POP                     = "POP"
	ADD                     = "ADD"
	SUB                     = "SUB"
	MUL                     = "MUL"
	DIV                     = "DIV"
)
