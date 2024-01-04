package fly

import (
	"github.com/superfly/lambdo/internal/config"
)

func OrgName() string {
	if config.GetConfig().Environment == "production" || config.GetConfig().Environment == "prod-debug" {
		return "personal"
	}

	return "personal"
}
