package config

type Config struct {
	SIPServer          string
	SIPUsername        string
	SIPPassword        string
	SIPDomain          string
	SIPTransport       string // "UDP" or "TCP" (default "UDP")
	SIPMaxICECandidates int   // max ICE candidates per media section (0 = unlimited)
	ListenAddr         string
	LogLevel           string
	APIBasePath        string
}
