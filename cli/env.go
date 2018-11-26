
package cli

import (
	"fmt"
	"os"
)

func VerifyEnvVar(envvar string) bool {
	if _, available := os.LookupEnv(envvar); !available {
		fmt.Fprintf(os.Stderr, "ERROR: Missing Environment Variable %s\n", envvar)
		return false
	}
	return true
}

func GetEnv(envVarName string) string {
	s := os.Getenv(envVarName)
	
	if len(s) > 0 && s[0] == '"' {
		s = s[1:]
	}
	
	if len(s) > 0 && s[len(s)-1] == '"' {
		s = s[:len(s)-1]
	}

	return s
}