// +build !windows

package channel

import (
	"os"
	"os/user"
	"strconv"
	"syscall"
	"time"
)

func getOwner(fi os.FileInfo) Person {
	uid := fi.Sys().(*syscall.Stat_t).Uid
	uidStr := strconv.FormatUint(uint64(uid), 10)
	user, err := user.LookupId(uidStr)
	if err != nil {
		return Person{}
	}

	return Person{
		Name: firstDefined(user.Name, user.Username),
	}
}

func getCreateTime(fi os.FileInfo) time.Time {
	stat := fi.Sys().(*syscall.Stat_t)
	ctimRaw, hasField := maybeGetField(stat, "Ctim")
	if !hasField {
		ctimRaw, _ = maybeGetField(stat, "Birthtimespec")
	}
	ctim := ctimRaw.(syscall.Timespec)
	return time.Unix(int64(ctim.Sec), int64(ctim.Nsec))
}
