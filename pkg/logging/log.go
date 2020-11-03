package logging

import (
	"fmt"
	"log"
	"runtime"
)

// Info logging
func Info(args ...interface{}) {
	log.SetPrefix("[INFO] ")
	log.Println(args...)
}

// Danger for error logging
func Danger(args ...interface{}) {
	log.SetPrefix("[ERROR] ")
	_, fn, line, _ := runtime.Caller(1)
	funcLine := fmt.Sprintf("%s:%d: ", fn, line)
	args = append([]interface{}{funcLine}, args...)
	log.Println(args...)
}

// Warning logging
func Warning(args ...interface{}) {
	log.SetPrefix("[WARNING] ")
	_, fn, line, _ := runtime.Caller(1)
	funcLine := fmt.Sprintf("%s:%d: ", fn, line)
	args = append([]interface{}{funcLine}, args...)
	log.Println(args...)
}

// Debug logging
func Debug(args ...interface{}) {
	log.SetPrefix("[DEBUG] ")
	log.Println(args...)
}
