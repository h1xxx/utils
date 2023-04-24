package main

import (
	"syscall"
)

func getDiskInfo(st *sttsT, vars *varsT) {
	var fsInfo syscall.Statfs_t
	syscall.Statfs("/", &fsInfo)
	st.rootDiskFree = float64(fsInfo.Bavail * uint64(fsInfo.Bsize))
	st.rootDiskFree /= 1024 * 1024
}
