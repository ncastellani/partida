package bootstrap

import "os"

// - region translations

// application region to AWS region map
var regionToAWSRegion map[string]string = map[string]string{
	"saopaulo": "sa-east-1",
	"montreal": "ca-central-1",
	"virginia": "us-east-1",
}

// Fly.io region into application region name
var flyioRegionToAPPRegion map[string]string = map[string]string{
	"gru": "saopaulo",
	"yul": "montreal",
	"iad": "virginia",
}

// - Fly.io utils

// determine if the current running instance is a Fly.io instance
func IsFlyioInstance() (isFly bool, region string) {
	if os.Getenv("FLY_MACHINE_ID") != "" {
		return true, flyioRegionToAPPRegion[os.Getenv("FLY_REGION")]
	}

	return
}
