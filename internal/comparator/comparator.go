package comparator

import (
	"fmt"
	"time"

	"github.com/AshitomW/Drifty/internal/models"
	"github.com/google/uuid"
)

type SeverityRules struct {
	CriticalPackages []string
	CriticalServices []string
	CriticalFiles    []string
	CriticalEnvVars  []string
}

type Comparator struct {
	severityRules SeverityRules
}

func New(rules SeverityRules) *Comparator {
	return &Comparator{
		severityRules: rules,
	}
}

// Compare will generate a drift report between two snapshots

func (c *Comparator) Compare(source, target *models.EnvironmentSnapshot) *models.DriftReport {
	report := &models.DriftReport{
		ID:             uuid.New().String(),
		Timestamp:      time.Now().UTC(),
		SourceEnv:      source.Name,
		TargetEnv:      target.Name,
		SourceSnapshot: source.ID,
		TargetSnapshot: target.ID,
		Drifts:         make([]models.DriftItem, 0),
		Summary: models.DriftSummary{
			ByCategory: make(map[string]int),
			ByType:     make(map[string]int),
		},
	}

	// compare the files
	c.compareFiles(source.Files, target.Files, report)

	// compare environment variables
	c.compareEnvVars(source.EnvVars, target.EnvVars, report)

	// compare packages
	c.comparePackages(source.Packages, target.Packages, report)

	// compare Services

	c.compareServices(source.Services, target.Services, report)

	// compare Network Config
	c.compareNetworkConfig(source.NetworkConfig, target.NetworkConfig, report)

	// compare Docker Config
	c.compareDockerConfig(source.DockerConfig, target.DockerConfig, report)

	// compare System Resources
	c.compareSystemResources(source.SystemResources, target.SystemResources, report)

	// compare Scheduled Tasks
	c.compareScheduledTasks(source.ScheduledTasks, target.ScheduledTasks, report)

	// compare Certificates
	c.compareCertificates(source.Certificates, target.Certificates, report)

	// compare User/Group Config
	c.compareUserGroupConfig(source.UserGroupConfig, target.UserGroupConfig, report)

	// Update summary
	c.updateSummary(report)

	return report
}

func (c *Comparator) compareFiles(source, target map[string]models.FileInfo, report *models.DriftReport) {

	// Find modified and removed files

	for path, srcFile := range source {
		if tgtFile, exists := target[path]; exists {
			if diff := c.diffFile(srcFile, tgtFile); diff != nil {
				drift := models.DriftItem{
					Type:      "modified",
					Category:  "file",
					Name:      path,
					SourceVal: srcFile,
					TargetVal: tgtFile,
					Severity:  c.getFileSeverity(path),
					Message:   fmt.Sprintf("File modified: %v", diff),
				}
				report.Drifts = append(report.Drifts, drift)
			}
		} else {
			drift := models.DriftItem{
				Type:      "removed",
				Category:  "file",
				Name:      path,
				SourceVal: srcFile,
				Severity:  c.getFileSeverity(path),
				Message:   "File exists in source but not in target",
			}
			report.Drifts = append(report.Drifts, drift)
		}
	}

	// Find added files

	for path, tgtFile := range target {
		if _, exists := source[path]; !exists {
			drift := models.DriftItem{
				Type:      "added",
				Category:  "file",
				Name:      path,
				TargetVal: tgtFile,
				Severity:  c.getFileSeverity(path),
				Message:   "File exists in target but not in source",
			}
			report.Drifts = append(report.Drifts, drift)
		}
	}
}

func matchPattern(s, pattern string) bool {
	// simple glob matching
	if pattern == "*" {
		return true
	}
	return s == pattern
}

func (c *Comparator) getFileSeverity(path string) string {
	for _, p := range c.severityRules.CriticalFiles {
		if matchPattern(path, p) {
			return "critical"
		}
	}

	return "info"
}

func (c *Comparator) getEnvVarSeverity(name string) string {
	for _, p := range c.severityRules.CriticalEnvVars {
		if matchPattern(name, p) {
			return "critical"
		}
	}
	return "warning"
}

func (c *Comparator) getPackageSeverity(name string) string {
	for _, p := range c.severityRules.CriticalPackages {
		if matchPattern(name, p) {
			return "critical"
		}

	}

	return "warning"
}

func (c *Comparator) getServiceSeverity(name string) string {
	for _, p := range c.severityRules.CriticalServices {
		if matchPattern(name, p) {
			return "critical"
		}
	}

	return "warning"
}

func (c *Comparator) diffFile(src, tgt models.FileInfo) map[string]interface{} {
	diff := make(map[string]interface{})
	if src.Hash != tgt.Hash && src.Hash != "" && tgt.Hash != "" {
		diff["hash"] = map[string]string{"source": src.Hash, "target": tgt.Hash}
	}

	if src.Mode != tgt.Mode {
		diff["mode"] = map[string]string{"source": src.Mode, "target": tgt.Mode}
	}

	if src.Owner != tgt.Owner {
		diff["owner"] = map[string]string{"source": src.Owner, "target": tgt.Owner}
	}

	if src.Group != tgt.Group {
		diff["group"] = map[string]string{"source": src.Group, "target": tgt.Group}
	}

	if len(diff) == 0 {
		return nil
	}

	return diff
}

func (c *Comparator) compareEnvVars(source, target map[string]models.EnvVar, report *models.DriftReport) {

	for name, srcVar := range source {
		if tgtVar, exists := target[name]; exists {
			if srcVar.Value != tgtVar.Value {
				drift := models.DriftItem{
					Type:      "modified",
					Category:  "envvar",
					Name:      name,
					SourceVal: srcVar.Value,
					TargetVal: tgtVar.Value,
					Severity:  c.getEnvVarSeverity(name),
					Message:   "Environment variable value changed.",
				}
				report.Drifts = append(report.Drifts, drift)
			}
		} else {
			drift := models.DriftItem{
				Type:      "removed",
				Category:  "envvar",
				Name:      name,
				SourceVal: srcVar.Value,
				Severity:  c.getEnvVarSeverity((name)),
				Message:   "Environment variable missing in target",
			}
			report.Drifts = append(report.Drifts, drift)
		}
	}

	for name, tgtVar := range target {
		if _, exists := source[name]; !exists {
			drift := models.DriftItem{
				Type:      "added",
				Category:  "envvar",
				Name:      name,
				TargetVal: tgtVar.Value,
				Severity:  c.getEnvVarSeverity(name),
				Message:   "Environment variable added in target",
			}
			report.Drifts = append(report.Drifts, drift)
		}
	}
}

func (c *Comparator) comparePackages(source, target map[string]models.PackageInfo, report *models.DriftReport) {

	for name, srcPkg := range source {
		if tgtPkg, exists := target[name]; exists {
			if srcPkg.Version != tgtPkg.Version {
				drift := models.DriftItem{
					Type:      "modified",
					Category:  "package",
					Name:      name,
					SourceVal: srcPkg.Version,
					TargetVal: tgtPkg.Version,
					Severity:  c.getPackageSeverity(name),
					Message:   fmt.Sprintf("Package version changed: %s -> %s", srcPkg.Version, tgtPkg.Version),
				}
				report.Drifts = append(report.Drifts, drift)
			}
		} else {
			drift := models.DriftItem{
				Type:      "removed",
				Category:  "package",
				Name:      name,
				SourceVal: srcPkg.Version,
				Severity:  c.getPackageSeverity(name),
				Message:   "Package missing in target",
			}
			report.Drifts = append(report.Drifts, drift)
		}
	}

	for name, tgtPkg := range target {
		if _, exists := source[name]; !exists {
			drift := models.DriftItem{
				Type:      "added",
				Category:  "package",
				Name:      name,
				TargetVal: tgtPkg.Version,
				Severity:  c.getPackageSeverity(name),
				Message:   "Package added in target",
			}
			report.Drifts = append(report.Drifts, drift)
		}
	}
}

func (c *Comparator) compareServices(source, target map[string]models.ServiceInfo, report *models.DriftReport) {

	for name, srcSvc := range source {
		if tgtSvc, exists := target[name]; exists {
			changes := make(map[string]interface{})

			if srcSvc.Status != tgtSvc.Status {
				changes["status"] = map[string]string{"source": srcSvc.Status, "target": tgtSvc.Status}
			}

			if srcSvc.Enabled != tgtSvc.Enabled {
				changes["enabled"] = map[string]bool{"source": srcSvc.Enabled, "target": tgtSvc.Enabled}
			}

			if len(changes) > 0 {
				drift := models.DriftItem{
					Type:      "modified",
					Category:  "services",
					Name:      name,
					SourceVal: srcSvc,
					TargetVal: tgtSvc,
					Severity:  c.getServiceSeverity(name),
					Message:   fmt.Sprintf("Service state changed: %v", changes),
				}
				report.Drifts = append(report.Drifts, drift)
			}
		} else {
			drift := models.DriftItem{
				Type:      "removed",
				Category:  "service",
				Name:      name,
				SourceVal: srcSvc,
				Severity:  c.getServiceSeverity(name),
				Message:   "Service missing in target",
			}
			report.Drifts = append(report.Drifts, drift)
		}
	}

	for name, tgtSvc := range target {
		if _, exists := source[name]; !exists {
			drift := models.DriftItem{
				Type:      "added",
				Category:  "service",
				Name:      name,
				TargetVal: tgtSvc,
				Severity:  c.getServiceSeverity(name),
				Message:   "service added in target",
			}
			report.Drifts = append(report.Drifts, drift)
		}
	}

}

func (c *Comparator) updateSummary(report *models.DriftReport) {

	report.HasDrift = len(report.Drifts) > 0
	report.Summary.TotalDrifts = len(report.Drifts)

	for _, drift := range report.Drifts {
		switch drift.Severity {
		case "critical":
			report.Summary.CriticalCount++
		case "warning":
			report.Summary.WarningCount++
		case "info":
			report.Summary.InfoCount++
		}

		report.Summary.ByCategory[drift.Category]++
		report.Summary.ByType[drift.Type]++
	}
}

func (c *Comparator) compareNetworkConfig(source, target models.NetworkConfig, report *models.DriftReport) {
	for name, srcIface := range source.Interfaces {
		if tgtIface, exists := target.Interfaces[name]; exists {
			if srcIface.MACAddress != tgtIface.MACAddress {
				drift := models.DriftItem{
					Type:      "modified",
					Category:  "network",
					Name:      name + " (interface)",
					SourceVal: srcIface.MACAddress,
					TargetVal: tgtIface.MACAddress,
					Severity:  "warning",
					Message:   "Interface MAC address changed",
				}
				report.Drifts = append(report.Drifts, drift)
			}
		} else {
			drift := models.DriftItem{
				Type:      "removed",
				Category:  "network",
				Name:      name + " (interface)",
				SourceVal: srcIface,
				Severity:  "warning",
				Message:   "Interface removed",
			}
			report.Drifts = append(report.Drifts, drift)
		}
	}

	for name, tgtIface := range target.Interfaces {
		if _, exists := source.Interfaces[name]; !exists {
			drift := models.DriftItem{
				Type:      "added",
				Category:  "network",
				Name:      name + " (interface)",
				TargetVal: tgtIface,
				Severity:  "warning",
				Message:   "Interface added",
			}
			report.Drifts = append(report.Drifts, drift)
		}
	}
}

func (c *Comparator) compareDockerConfig(source, target models.DockerConfig, report *models.DriftReport) {
	for id, srcCont := range source.Containers {
		if tgtCont, exists := target.Containers[id]; exists {
			if srcCont.Status != tgtCont.Status || srcCont.State != tgtCont.State {
				drift := models.DriftItem{
					Type:      "modified",
					Category:  "docker",
					Name:      srcCont.Name,
					SourceVal: srcCont.Status + " " + srcCont.State,
					TargetVal: tgtCont.Status + " " + tgtCont.State,
					Severity:  "warning",
					Message:   "Container status/state changed",
				}
				report.Drifts = append(report.Drifts, drift)
			}
		} else {
			drift := models.DriftItem{
				Type:      "removed",
				Category:  "docker",
				Name:      srcCont.Name,
				SourceVal: srcCont,
				Severity:  "info",
				Message:   "Container removed",
			}
			report.Drifts = append(report.Drifts, drift)
		}
	}

	for id, tgtCont := range target.Containers {
		if _, exists := source.Containers[id]; !exists {
			drift := models.DriftItem{
				Type:      "added",
				Category:  "docker",
				Name:      tgtCont.Name,
				TargetVal: tgtCont,
				Severity:  "info",
				Message:   "Container added",
			}
			report.Drifts = append(report.Drifts, drift)
		}
	}
}

func (c *Comparator) compareSystemResources(source, target models.SystemResources, report *models.DriftReport) {
	if source.CPU.Cores != target.CPU.Cores {
		drift := models.DriftItem{
			Type:      "modified",
			Category:  "resources",
			Name:      "CPU cores",
			SourceVal: source.CPU.Cores,
			TargetVal: target.CPU.Cores,
			Severity:  "critical",
			Message:   "CPU core count changed",
		}
		report.Drifts = append(report.Drifts, drift)
	}

	if source.Memory.Total != target.Memory.Total {
		drift := models.DriftItem{
			Type:      "modified",
			Category:  "resources",
			Name:      "Memory total",
			SourceVal: source.Memory.Total,
			TargetVal: target.Memory.Total,
			Severity:  "critical",
			Message:   "Total memory changed",
		}
		report.Drifts = append(report.Drifts, drift)
	}
}

func (c *Comparator) compareScheduledTasks(source, target models.ScheduledTasks, report *models.DriftReport) {
	for name, srcTask := range source.CronJobs {
		if tgtTask, exists := target.CronJobs[name]; exists {
			if srcTask.Schedule != tgtTask.Schedule || srcTask.Command != tgtTask.Command {
				drift := models.DriftItem{
					Type:      "modified",
					Category:  "scheduled_task",
					Name:      name + " (cron)",
					SourceVal: srcTask.Command,
					TargetVal: tgtTask.Command,
					Severity:  "warning",
					Message:   "Cron job changed",
				}
				report.Drifts = append(report.Drifts, drift)
			}
		} else {
			drift := models.DriftItem{
				Type:      "removed",
				Category:  "scheduled_task",
				Name:      name + " (cron)",
				SourceVal: srcTask,
				Severity:  "warning",
				Message:   "Cron job removed",
			}
			report.Drifts = append(report.Drifts, drift)
		}
	}

	for name, tgtTask := range target.CronJobs {
		if _, exists := source.CronJobs[name]; !exists {
			drift := models.DriftItem{
				Type:      "added",
				Category:  "scheduled_task",
				Name:      name + " (cron)",
				TargetVal: tgtTask,
				Severity:  "warning",
				Message:   "Cron job added",
			}
			report.Drifts = append(report.Drifts, drift)
		}
	}
}

func (c *Comparator) compareCertificates(source, target map[string]models.Certificate, report *models.DriftReport) {
	for path, srcCert := range source {
		if tgtCert, exists := target[path]; exists {
			if srcCert.Fingerprint != tgtCert.Fingerprint {
				drift := models.DriftItem{
					Type:      "modified",
					Category:  "certificate",
					Name:      path,
					SourceVal: srcCert.Fingerprint,
					TargetVal: tgtCert.Fingerprint,
					Severity:  "warning",
					Message:   "Certificate changed",
				}
				report.Drifts = append(report.Drifts, drift)
			}
			if !srcCert.IsExpired && tgtCert.IsExpired {
				drift := models.DriftItem{
					Type:      "modified",
					Category:  "certificate",
					Name:      path,
					SourceVal: "valid",
					TargetVal: "expired",
					Severity:  "critical",
					Message:   "Certificate expired",
				}
				report.Drifts = append(report.Drifts, drift)
			}
		} else {
			drift := models.DriftItem{
				Type:      "removed",
				Category:  "certificate",
				Name:      path,
				SourceVal: srcCert,
				Severity:  "warning",
				Message:   "Certificate removed",
			}
			report.Drifts = append(report.Drifts, drift)
		}
	}

	for path, tgtCert := range target {
		if _, exists := source[path]; !exists {
			drift := models.DriftItem{
				Type:      "added",
				Category:  "certificate",
				Name:      path,
				TargetVal: tgtCert,
				Severity:  "info",
				Message:   "Certificate added",
			}
			report.Drifts = append(report.Drifts, drift)
		}
	}
}

func (c *Comparator) compareUserGroupConfig(source, target models.UserGroupConfig, report *models.DriftReport) {
	for name, srcUser := range source.Users {
		if tgtUser, exists := target.Users[name]; exists {
			if srcUser.UID != tgtUser.UID {
				drift := models.DriftItem{
					Type:      "modified",
					Category:  "user",
					Name:      name,
					SourceVal: srcUser.UID,
					TargetVal: tgtUser.UID,
					Severity:  "warning",
					Message:   "User UID changed",
				}
				report.Drifts = append(report.Drifts, drift)
			}
		} else {
			drift := models.DriftItem{
				Type:      "removed",
				Category:  "user",
				Name:      name,
				SourceVal: srcUser,
				Severity:  "warning",
				Message:   "User removed",
			}
			report.Drifts = append(report.Drifts, drift)
		}
	}

	for name, tgtUser := range target.Users {
		if _, exists := source.Users[name]; !exists {
			drift := models.DriftItem{
				Type:      "added",
				Category:  "user",
				Name:      name,
				TargetVal: tgtUser,
				Severity:  "warning",
				Message:   "User added",
			}
			report.Drifts = append(report.Drifts, drift)
		}
	}
}
