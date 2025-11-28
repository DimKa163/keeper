package shared

const (
	_        = iota
	KB int64 = 1 << (10 * iota) // 1 KiB = 1024
	MB
	GB
	TB
	PB
)
