package channel

import (
	"os"
	"time"
)

func getOwner(fi os.FileInfo) Person {
	return Person{}
}

func getCreateTime(fi os.FileInfo) time.Time {
	return time.Time{}
}
