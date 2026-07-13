package todo

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	pkgtodo "github.com/SirSobhan0/gotodo/pkg/todo"
)

var ConfigDir string

func init() {
	homeDir, _ := os.UserHomeDir()
	ConfigDir = filepath.Join(homeDir, ".config", "gotodo")
	os.MkdirAll(ConfigDir, 0755)
}

func GetProjectFile(projectName string) string {
	return filepath.Join(ConfigDir, projectName+".json")
}

func ListProjects() ([]string, error) {
	files, err := os.ReadDir(ConfigDir)
	if err != nil {
		return nil, fmt.Errorf("read directory: %w", err)
	}
	var projects []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".json") {
			name := strings.TrimSuffix(f.Name(), ".json")
			projects = append(projects, name)
		}
	}
	return projects, nil
}

func SaveTasksToFile(projectName string, tasks []pkgtodo.Task) error {
	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal tasks: %w", err)
	}
	err = os.WriteFile(GetProjectFile(projectName), data, 0644)
	if err != nil {
		return fmt.Errorf("write tasks: %w", err)
	}
	return nil
}

func LoadTasksFromFile(projectName string) ([]pkgtodo.Task, error) {
	data, err := os.ReadFile(GetProjectFile(projectName))
	if err != nil {
		if os.IsNotExist(err) {
			return []pkgtodo.Task{}, nil
		}
		return nil, fmt.Errorf("read tasks file: %w", err)
	}
	var tasks []pkgtodo.Task
	err = json.Unmarshal(data, &tasks)
	if err != nil {
		return nil, fmt.Errorf("unmarshal tasks: %w", err)
	}
	return tasks, nil
}
