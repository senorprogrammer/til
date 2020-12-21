package src

import (
	"fmt"
	"time"
)

func Footer() string {
	return fmt.Sprintf(
		"<sup><sub>generated %s by <a href='https://github.com/senorprogrammer/til'>til</a></sub></sup>\n",
		time.Now().Format("2 Jan 2006 15:04:05"),
	)
}
