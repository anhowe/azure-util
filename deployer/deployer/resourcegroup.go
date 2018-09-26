package deployer

import (
	"fmt"
	"time"
)

func GetResourceGroupPrefix(uniquePrefix string, t time.Time) string {
	return fmt.Sprintf("%s-%02d%02d-%02d%02d", uniquePrefix, t.Month(), t.Day(), t.Hour(), t.Minute())
}

func GetResourceGroupName(rgPrefix string, rgId int, vmCount int) string {
	return fmt.Sprintf("%s-%03d-%03d", rgPrefix, rgId, vmCount)
}