package network

const (
	ENDBYTES = "\000\005\007\001\001\007\005\000"
	DMAXSIZE = (2 << 20) // (2^20)*2 = 2MiB
	BUFFSIZE = (4 << 10) // (2^10)*4 = 4KiB
)

type Package struct {
	Option string
	Data   string
}
