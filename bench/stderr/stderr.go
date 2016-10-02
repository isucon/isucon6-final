package stderr

import (
	"log"
	"os"
)

var Log *log.Logger = log.New(os.Stderr, "", 0)
