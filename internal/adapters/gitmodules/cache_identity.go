package gitmodules

type cacheIdentity struct {
	Device            uint64
	Inode             uint64
	Links             uint64
	ChangeSeconds     int64
	ChangeNanoseconds int64
}
