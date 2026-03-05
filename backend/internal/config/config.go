package config

type Config struct {
	SIPServer    string
	SIPUsername  string
	SIPPassword  string
	SIPDomain    string
	SIPTransport string // "UDP" or "TCP" (default "UDP")
	ListenAddr   string
	LogLevel     string
	APIBasePath  string
}
