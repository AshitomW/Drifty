package collector

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"sync"
	"syscall"

	"github.com/AshitomW/Drifty/internal/models"
)



func getPathDepth(path, basePath string) int{
	relPath, _ := filepath.Rel(basePath,path)
	return len(filepath.SplitList(relPath))
}


func (c *Collector) calculateFileHash(path string) (string,error){
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()


	var h interface{
		io.Writer
		Sum([]byte) []byte
	}


	switch c.config.Files.HashAlgo{
	case "md5":
		h = md5.New()
	default:
		h = sha256.New()
	}


	if _,err := io.Copy(h.(io.Writer), f); err != nil{
		return "",err 
	}

	return fmt.Sprintf("%x",h.Sum(nil)),nil
}




func (c* Collector) processFile(path string, info os.FileInfo) models.FileInfo{
	fileInfo := models.FileInfo{
		Path: path,
		Size: info.Size(),
		Mode: info.Mode().String(),
		ModTime: info.ModTime(),
		IsDirectory: info.IsDir(),
		Exists: true,
	}


	// Get Owner / group (Unix Specific)



	if stat, ok := info.Sys().(*syscall.Stat_t);ok{
		if u, err := user.LookupId((strconv.Itoa((int(stat.Uid))))); err == nil{
			fileInfo.Owner = u.Username
		}

		if g, err := user.LookupGroupId(strconv.Itoa(int(stat.Gid))); err == nil{
			fileInfo.Group = g.Name
		}
	}


	// Calculating the has for files (not directories)

	// we will be skipping file sizes greater than 100 MB
	if !info.IsDir() && info.Size() < 100 * 1024 * 1024 {
		hash, err := c.calculateFileHash(path)
		if err == nil{
			fileInfo.Hash = hash 
		}
	}

	return fileInfo
}




func (c *Collector) collectFiles(ctx context.Context) (map[string]models.FileInfo, error){

	files := make(map[string]models.FileInfo)

	var mu sync.Mutex



	// Compiling exclude patterns


	var excludePatterns []*regexp.Regexp
	for _ , pattern := range c.config.Files.ExcludePaths {
		re, err := regexp.Compile(pattern)
		if err != nil{
			continue
		}

		excludePatterns = append(excludePatterns, re)
	}

	// Worker pool for the file processing
	type fileJob struct{
		path string
		info os.FileInfo
	}



	jobs := make(chan fileJob, 1000)
	results := make(chan models.FileInfo, 1000)


	var wg sync.WaitGroup



	// Execute the worksers / starting

	for i := 0; i < c.workers; i++{
		wg.Add(1)
		go func(){
			defer wg.Done()
			for job:= range jobs{
				select{
				case <- ctx.Done():
					return 
				default:
					fileInfo := c.processFile(job.path, job.info)
					results <- fileInfo
				}
			}
		}()
	}



	// Collect results


	go func(){
		wg.Wait()
		close(results)
	}()

	// walk torugh directores annd send jobs
	go func(){
		for _ , basePath := range c.config.Files.Paths{
			filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
				if err != nil{
					return nil // skip the files we are not allowed to have access to
				}

				// select context cancellation


				select {
				case <- ctx.Done():
					return ctx.Err()
				default:
				}


				// Checking for exclusions


				for _ , re := range excludePatterns{
					if re.MatchString(path){
						if info.IsDir(){
							return filepath.SkipDir
						}
						return nil
					}
				}


				if c.config.Files.MaxDepth > 0{
					depth := getPathDepth(path,basePath)
					if depth > c.config.Files.MaxDepth{
						if info.IsDir(){
							return filepath.SkipDir
						}
						return nil
					}
				}

				jobs <- fileJob{path: path,info: info}
				return nil
			})
		}
		close(jobs)
	}()

		// Collect results


		for fileInfo := range results{
			mu.Lock()
			files[fileInfo.Path] = fileInfo
			mu.Unlock()
		}


		return files,nil

}


