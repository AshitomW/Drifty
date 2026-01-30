package collector

import (
	"bufio"
	"context"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/AshitomW/Drifty/internal/models"
)

func (c *Collector) collectUserGroupConfig(ctx context.Context) (models.UserGroupConfig, error) {
	config := models.UserGroupConfig{
		Users:  make(map[string]models.UserInfo),
		Groups: make(map[string]models.GroupInfo),
	}

	if c.config.UsersGroups.Users {
		users, err := c.collectUsers(ctx)
		if err == nil {
			config.Users = users
		}
	}

	if c.config.UsersGroups.Groups {
		groups, err := c.collectGroups(ctx)
		if err == nil {
			config.Groups = groups
		}
	}

	if c.config.UsersGroups.SudoRules {
		rules, err := c.collectSudoRules(ctx)
		if err == nil {
			config.SudoRules = rules
		}
	}

	return config, nil
}

func (c *Collector) collectUsers(ctx context.Context) (map[string]models.UserInfo, error) {
	users := make(map[string]models.UserInfo)

	passwdPath := "/etc/passwd"
	file, err := os.Open(passwdPath)
	if err != nil {
		return users, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, ":")
		if len(fields) < 7 {
			continue
		}

		uid, _ := strconv.Atoi(fields[2])
		gid, _ := strconv.Atoi(fields[3])

		users[fields[0]] = models.UserInfo{
			Name:    fields[0],
			UID:     uid,
			GID:     gid,
			HomeDir: fields[5],
			Shell:   fields[6],
			Comment: fields[4],
		}
	}

	return users, nil
}

func (c *Collector) collectGroups(ctx context.Context) (map[string]models.GroupInfo, error) {
	groups := make(map[string]models.GroupInfo)

	groupPath := "/etc/group"
	file, err := os.Open(groupPath)
	if err != nil {
		return groups, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, ":")
		if len(fields) < 4 {
			continue
		}

		gid, _ := strconv.Atoi(fields[2])
		members := []string{}
		if fields[3] != "" {
			members = strings.Split(fields[3], ",")
		}

		groups[fields[0]] = models.GroupInfo{
			Name:    fields[0],
			GID:     gid,
			Members: members,
		}
	}

	return groups, nil
}

func (c *Collector) collectSudoRules(ctx context.Context) ([]models.SudoRule, error) {
	var rules []models.SudoRule

	sudoersPath := "/etc/sudoers"
	if _, err := os.Stat(sudoersPath); os.IsNotExist(err) {
		return rules, nil
	}

	file, err := os.Open(sudoersPath)
	if err != nil {
		return rules, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		user := fields[0]
		host := "ALL"
		runas := ""
		commands := ""

		for i := 1; i < len(fields); i++ {
			if strings.Contains(fields[i], "=") {
				parts := strings.Split(fields[i], "=")
				if len(parts) == 2 {
					if parts[0] == "host" {
						host = parts[1]
					} else if parts[0] == "runas" {
						runas = parts[1]
					}
				}
			} else {
				if commands == "" {
					commands = fields[i]
				} else {
					commands += " " + fields[i]
				}
			}
		}

		rules = append(rules, models.SudoRule{
			User:     user,
			Host:     host,
			RunAs:    runas,
			Commands: commands,
		})
	}

	if runtime.GOOS == "darwin" {
		sudoDPath := "/etc/sudoers.d"
		if info, err := os.Stat(sudoDPath); err == nil && info.IsDir() {
			entries, _ := os.ReadDir(sudoDPath)
			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}

				filePath := sudoDPath + "/" + entry.Name()
				fileRules, err := c.parseSudoersFile(ctx, filePath)
				if err == nil {
					rules = append(rules, fileRules...)
				}
			}
		}
	}

	return rules, nil
}

func (c *Collector) parseSudoersFile(ctx context.Context, path string) ([]models.SudoRule, error) {
	var rules []models.SudoRule

	file, err := os.Open(path)
	if err != nil {
		return rules, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		user := fields[0]
		host := "ALL"
		runas := ""
		commands := strings.Join(fields[1:], " ")

		rules = append(rules, models.SudoRule{
			User:     user,
			Host:     host,
			RunAs:    runas,
			Commands: commands,
		})
	}

	return rules, nil
}
