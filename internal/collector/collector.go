package collector

import (
	"runtime"

	"github.com/AshitomW/Drifty/internal/models"
)



type Collector struct{
	config models.CollectorConfig
	workers int 
}




// Creating a new Coollcter instance
func New(config models.CollectorConfig) *Collector{
	workers := runtime.NumCPU()
	if workers < 2 {
		workers = 2
	}

	return &Collector{
		config: config,
		workers: workers,
	}

}



// Collecting OS Information

func (c *Collector) collectOSInfo() models.OSInfo{
	return models.OSInfo{
		Name: runtime.GOOS,
		Arch: runtime.GOARCH,
		Version: getOSVersion(),
		Kernel: getKernelVersion(),
	}
}





