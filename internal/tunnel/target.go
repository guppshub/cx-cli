package tunnel

// Target defines the target endpoints and settings for a network tunnel.
type Target struct {
	BastionInstanceID  string
	RemoteHost         string
	RemotePort         int
	PreferredLocalPort int
}
