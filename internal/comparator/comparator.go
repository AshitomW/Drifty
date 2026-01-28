package comparator

import (
	"fmt"
	"time"

	"github.com/AshitomW/Drifty/internal/models"
	"github.com/google/uuid"
)

type SeverityRules struct{
	CriticalPackages []string 
	CriticalServices []string 
	CriticalFiles []string
	CriticalEnvVars []string 
}

type Comparator struct {
	severityRules SeverityRules
}


func New(rules SeverityRules) *Comparator{
	return &Comparator{
		severityRules: rules,
	}
}


// Compare will generate a drift report between two snapshots


func (c *Comparator) Compare(source, target *models.EnvironmentSnapshot) *models.DriftReport{
	report := &models.DriftReport{
		ID: uuid.New().String(),
		Timestamp: time.Now().UTC(),
		SourceEnv: source.Name,
		TargetEnv: target.Name,
		SourceSnapshot: source.ID,
		TargetSnapshot: target.ID,
		Drifts:  make([]models.DriftItem,0),
		Summary: models.DriftSummary{
			ByCategory: make(map[string]int),
			ByType: make(map[string]int),

		},
	}


	// compare the files
	c.compareFiles(source.Files, target.Files,report)

	// compare environment variables
	c.compareEnvVars(source.EnvVars,target.EnvVars,report)

	// compare packages
	c.comparePackages(source.Packages,target.Packages,report)

	// compare Services

	c.compareServices(source.Services,target.Services,report)


	// Update summary
	c.updateSummary(report)

	return report;
}







func (c *Comparator) compareFiles(source, target map[string]models.FileInfo, report *models.DriftReport){

	// Find modified and removed files

	for path, srcFile := range source{
		if tgtFile, exists := target[path]; exists{
			if diff := c.diffFile(srcFile,tgtFile); diff != nil{
				drift := models.DriftItem{
					Type: "modified",
					Category: "file",
					Name: path,
					SourceVal: srcFile,
					TargetVal: tgtFile,
					Severity: c.getFileSeverity(path),
					Message: fmt.Sprintf("File modified: %v",diff),
				}
				report.Drifts = append(report.Drifts, drift)
			}
		}else {
			drift := models.DriftItem{
				Type: "removed",
				Category: "file",
				Name: path,
				SourceVal: srcFile,
				Severity: c.getFileSeverity(path),
				Message: "File exists in source but not in target",
			}
			report.Drifts = append(report.Drifts,drift)
		}
	}


	// Find added files


	for path, tgtFile := range target{
		if _ , exists := source[path]; !exists{
			drift := models.DriftItem{
				Type:"added",
				Category: "file",
				Name: path,
				TargetVal: tgtFile,
				Severity: c.getFileSeverity(path),
				Message: "File exists in target but not in source",
			}
			report.Drifts = append(report.Drifts, drift)
		}
	}
} 


func matchPattern(s, pattern string) bool{
	// simple glob matching
	if pattern == "*"{
		return true 
	}
	return s == pattern
}


func (c *Comparator) getFileSeverity(path string) string {
	for _, p := range c.severityRules.CriticalFiles{
		if matchPattern(path,p){
			return "critical";
		}
	}

	return "info"
}


func (c *Comparator) getEnvVarSeverity(name string) string{
	for _, p := range c.severityRules.CriticalEnvVars{
		if matchPattern(name,p){
			return "critical"
		}	
	}
	return "warning"
}

func (c *Comparator) getPackageSeverity(name string) string{
	for _, p := range c.severityRules.CriticalPackages {
		if matchPattern(name,p){
			return "critical";
		}

	}

	return "warning";
}


func (c *Comparator) getServiceSeverity(name string) string{
	for _, p := range c.severityRules.CriticalServices{
		if matchPattern(name,p){
			return "critical";
		}
	}

	return "warning";
}








func (c *Comparator) diffFile(src, tgt models.FileInfo)map[string]interface{}{
	diff := make(map[string]interface{})
	if src.Hash != tgt.Hash && src.Hash != "" && tgt.Hash != ""{
		diff["hash"] = map[string]string{"source":src.Hash,"target":tgt.Hash}
	}

	if src.Mode != tgt.Mode{
		diff["mode"] = map[string]string{"source":src.Mode,"target":tgt.Mode}
	}


	if src.Owner != tgt.Owner {
		diff["owner"] = map[string]string{"source":src.Owner, "target":tgt.Owner}
	}

	if src.Group != tgt.Group{
		diff["group"] = map[string]string{"source":src.Group,"target":tgt.Group}
	}


	if len(diff) == 0 {
		return nil 
	}

	return diff
}




func (c *Comparator) compareEnvVars(source, target map[string]models.EnvVar, report *models.DriftReport){

	for name, srcVar := range source{
		if tgtVar, exists := target[name]; exists {
			if srcVar.Value != tgtVar.Value{
				drift := models.DriftItem{
					Type: "modified",
					Category: "envvar",
					Name: name,
					SourceVal: srcVar.Value,
					TargetVal: tgtVar.Value,
					Severity: c.getEnvVarSeverity(name),
					Message: "Environment variable value changed",
				}
				report.Drifts = append(report.Drifts, drift)
			}
		}else {
			drift := models.DriftItem{
				Type: "removed",
				Category: "envvar",
				Name: name,
				SourceVal: srcVar.Value,
				Severity: c.getEnvVarSeverity((name)),
				Message:"Environment variable missing in target",
			}
			report.Drifts = append(report.Drifts, drift)
		}
	}


	for name, tgtVar := range target{
		if _, exists := source[name];!exists {
			drift := models.DriftItem{
				Type:"added",
				Category: "envvar",
				Name: name,
				TargetVal: tgtVar.Value,
				Severity: c.getEnvVarSeverity(name),
				Message: "Environment variable added in target",
			}
			report.Drifts = append(report.Drifts, drift)
		}
	}
} 


func (c *Comparator) comparePackages(source, target map[string]models.PackageInfo, report *models.DriftReport){


	for name, srcPkg := range source {
		if tgtPkg, exists := target[name];exists{
			if srcPkg.Version != tgtPkg.Version {
				drift := models.DriftItem{
					Type:"modified",
					Category: "package",
					Name: name,
					SourceVal: srcPkg.Version,
					TargetVal: tgtPkg.Version,
					Severity: c.getPackageSeverity(name),
					Message: fmt.Sprintf("Package version changed: %s -> %s",srcPkg.Version,tgtPkg.Version),
				}
				report.Drifts = append(report.Drifts, drift)
			}
		}else {
			drift := models.DriftItem{
				Type: "removed",
				Category: "package",
				Name: name,
				SourceVal: srcPkg.Version,
				Severity: c.getPackageSeverity(name),
				Message: "Package missing in target",
			}
			report.Drifts = append(report.Drifts, drift)
		}
	}


	for name, tgtPkg := range target {
		if _, exists := source[name]; !exists{
			drift := models.DriftItem{
				Type:"added",
				Category: "package",
				Name: name,
				TargetVal: tgtPkg.Version,
				Severity: c.getPackageSeverity(name),
				Message: "Package added in target",
			}
			report.Drifts = append(report.Drifts, drift)
		}
	}
}




func (c *Comparator) compareServices(source, target map[string]models.ServiceInfo,report *models.DriftReport){


		for name, srcSvc := range source {
			if tgtSvc, exists := target[name]; exists{
				changes := make(map[string]interface{})
				

				if srcSvc.Status != tgtSvc.Status{
					changes["status"] = map[string]string{"source":srcSvc.Status,"target":tgtSvc.Status}
				}

				if srcSvc.Enabled != tgtSvc.Enabled{
					changes["enabled"] = map[string]bool{"source":srcSvc.Enabled,"target":tgtSvc.Enabled}
				}


				if len(changes) > 0 {
					drift := models.DriftItem{
						Type: "modified",
						Category: "services",
						Name: name,
						SourceVal: srcSvc,
						TargetVal: tgtSvc,
						Severity: c.getServiceSeverity(name),
						Message: fmt.Sprintf("Service state changed: %v",changes),
					}
					report.Drifts = append(report.Drifts,drift)
				}
			}else {
				drift := models.DriftItem{
					Type:"removed",
					Category: "service",
					Name: name,
					SourceVal: srcSvc,
					Severity: c.getServiceSeverity(name),
					Message: "Service missing in target",
				}
				report.Drifts = append(report.Drifts, drift)
			}
		}


		for name, tgtSvc := range target{
			if _, exists := source[name]; !exists{
			drift := models.DriftItem{
				Type:"added",
				Category: "service",
				Name: name,
				TargetVal: tgtSvc,
				Severity: c.getServiceSeverity(name),
				Message: "service added in target",
			}
			report.Drifts = append(report.Drifts, drift)
		}
	}

}


func (c *Comparator) updateSummary(report *models.DriftReport){

	report.HasDrift = len(report.Drifts) > 0 
	report.Summary.TotalDrifts = len(report.Drifts)


	for _, drift := range report.Drifts {
		switch drift.Severity{
		case "critical":
			report.Summary.CriticalCount++;
		case "warning":
			report.Summary.WarningCount++;
		case "info":
			report.Summary.InfoCount++
		}


		report.Summary.ByCategory[drift.Category]++
		report.Summary.ByType[drift.Type]++
	}
}