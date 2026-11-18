// Package verify asserts the platform is in a healthy state for the SRE drills.
package verify

import (
	"fmt"
	"strings"

	"github.com/fabianoflorentino/srectl/internal/platform"
)

// Assess returns nil when the cluster is up, Argo CD is ready and every app is
// Synced and Healthy; otherwise it returns an error listing the problems.
func Assess(report platform.Report) error {
	var problems []string
	if !strings.HasPrefix(report.Cluster, "up") {
		problems = append(problems, "cluster is down")
	}
	if report.Argocd != "Ready" {
		problems = append(problems, "argocd is not ready")
	}
	for _, app := range report.Apps {
		if app.Sync != "Synced" || app.Health != "Healthy" {
			problems = append(problems, fmt.Sprintf("%s: %s · %s", app.Name, app.Sync, app.Health))
		}
	}
	if len(problems) > 0 {
		return fmt.Errorf("platform not healthy:\n  - %s", strings.Join(problems, "\n  - "))
	}
	return nil
}
