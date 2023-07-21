package Daemonize

type Severity int

const (
	// LOG_DEBUG Useful data for debugging
	LOG_DEBUG Severity = iota
	// LOG_INFO Non-important information, considered to be the default level
	LOG_INFO
	// LOG_WARNING Rare or unexpected conditions
	LOG_WARNING
	// LOG_ERR Errors
	LOG_ERR
	// LOG_EMERG Fatal errors
	LOG_EMERG
)
