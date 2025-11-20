package env

import (
	"log"
	"os"
	"strconv"
)

var CleanRowsCountToIncreaseSpeed = uint8(getEnvInt("CLEAN_ROWS_COUNT", 12))

func getEnvInt(key string, def int) int {
	valStr := os.Getenv(key)
	if valStr == "" {
		return def
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		log.Printf("Invalid %s value, using default %d", key, def)
		return def
	}
	return val
}
