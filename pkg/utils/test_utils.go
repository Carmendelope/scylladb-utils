/*
 * Copyright (C)  2019 Nalej - All Rights Reserved
 */

package utils

import (
	"os"
)

func RunIntegrationTests() bool {
	var runIntegration = os.Getenv("RUN_INTEGRATION_TEST")
	return runIntegration == "true"
}

