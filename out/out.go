package out

import (
	"fmt"
	"time"
)

func init() {
	LogString("out/init")
}

// LogString will output the string to the logfile, including a timestamp.
// Returns an integer so that it can be used for variable initialization.
func LogString(s string) int {
	fmt.Printf("%s %0.3f\n", s, float64(time.Now().UnixNano())/(1000.0*1000.0*1000.0))
	return 1
}
