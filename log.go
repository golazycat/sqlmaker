package sqlmaker

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

var (
	logger *log.Logger
	debug  = false
)

func init() {
	logger = log.New(os.Stdout, colorString("[sqlmaker-DEBUG] "),
		log.Ldate|log.Ltime)
}

func DebugMode() {
	debug = true
}

func wLog(s string, v ...interface{}) {
	if debug {
		logger.Printf(s, v...)
	}
}

// 返回有颜色的字体
func colorString(s string) string {
	return fmt.Sprintf("\033[%d;1m%s\033[0m", 32, s)
}

func printValues(vs []interface{}) string {
	s := make([]string, 0)
	for _, v := range vs {
		var valStr string
		switch v.(type) {
		case int:
			valStr = fmt.Sprintf("%d", v)
		case time.Time:
			valStr = v.(time.Time).Format(datetimeFormat)
		default:
			valStr = fmt.Sprintf("%s", v)

		}
		s = append(s, valStr)
	}
	return strings.Join(s, ",")
}
