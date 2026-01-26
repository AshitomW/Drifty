package collector

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
)



func getOSVersion() string{
	switch runtime.GOOS{
	case "linux":
		return getLinuxVersion()
	case "darwin":
		return getMacOSVersion()
	case "windows":
		return getWindowsVersion()
	default:
		return "unknown"
	}
}


func getLinuxVersion() string{
	// Trying /etc/os-release first

	data , err := os.ReadFile("/etc/os-release")
	if err == nil{
		lines := strings.Split(string(data), "\n")
		for _, line := range lines{
			if strings.HasPrefix(line,"PRETTY_NAME="){
				return strings.Trim(strings.TrimPrefix(line,"PRETTY_NAME="),"\"")
			}
		}
	}

	// Falling back for lsb release

	cmd := exec.Command("lsb_release","-d","-s")
	output, err := cmd.Output()
	if err != nil {
		return strings.TrimSpace(string(output))
	}


	return "Linux"
}



func getMacOSVersion() string{
	cmd := exec.Command("sw_vers","-productVersion")
	output, err := cmd.Output()
	if err != nil{
		return "macOS"
	}

	return "macOS" + strings.TrimSpace(string(output))
}




func getWindowsVersion() string {
	cmd := exec.Command("cmd","/c","ver")
	output , err := cmd.Output()
	if err != nil{
		return "Windows"
	}

	return strings.TrimSpace(string(output))
}



func getKernelVersion() string {
	switch runtime.GOOS{
	case "linux","darwin":
		cmd := exec.Command("uname","-r")
		output, err := cmd.Output()
		if err != nil{
			return "unknown"
		}
		return strings.TrimSpace(string(output))
	
	default:
		return "Unknown"
	}
}

