package repo

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/synthesis-labs/polaris-cli/src/config"
	yaml "gopkg.in/yaml.v2"
)

func searchRepoForBase(polarisHome string, polarisConfig *config.PolarisConfig, baseName string, matchingNames ...string) (map[string]*config.PolarisScaffold, error) {
	reposHome := filepath.Clean(filepath.Join(polarisHome, "repos"))
	result := map[string]*config.PolarisScaffold{}
	err := filepath.Walk(reposHome, func(filename string, info os.FileInfo, err error) error {
		filebase := filepath.Base(filename)
		if filebase == baseName {

			filenameRelativeToRepo, err := filepath.Rel(reposHome, filename)
			if err != nil {
				return err
			}

			scaffoldName := filepath.ToSlash(filepath.Dir(filenameRelativeToRepo))
			scaffoldData, err := ioutil.ReadFile(filename)
			if err != nil {
				return err
			}

			scaffold := config.PolarisScaffold{}
			err = yaml.Unmarshal(scaffoldData, &scaffold.Spec)
			if err != nil {
				return err
			}

			scaffold.Name = scaffoldName
			scaffold.LocalPath = filepath.Dir(filename)
			found := len(matchingNames) == 0
			for _, matching := range matchingNames {
				if matching == scaffoldName {
					found = true
				}
			}
			if found {
				result[scaffoldName] = &scaffold
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

// ListProjects returns the list of available projects in all repositories
//
func ListProjects(polarisHome string, polarisConfig *config.PolarisConfig, matchingNames ...string) (map[string]*config.PolarisScaffold, error) {
	return searchRepoForBase(polarisHome, polarisConfig, "polaris-project.yaml", matchingNames...)
}

// GetProject returns a particular project
//
func GetProject(polarisHome string, polarisConfig *config.PolarisConfig, projectName string) (*config.PolarisScaffold, error) {
	projects, err := ListProjects(polarisHome, polarisConfig, projectName)

	if err != nil {
		return nil, err
	}

	if len(projects) != 1 {
		return nil, fmt.Errorf("Unable to find project with name %s", projectName)
	}

	return projects[projectName], nil
}

// ListComponents returns the list of available components in all repositories
//
func ListComponents(polarisHome string, polarisConfig *config.PolarisConfig, matchingNames ...string) (map[string]*config.PolarisScaffold, error) {
	return searchRepoForBase(polarisHome, polarisConfig, "polaris-component.yaml", matchingNames...)
}

// GetComponent returns a particular component
//
func GetComponent(polarisHome string, polarisConfig *config.PolarisConfig, componentName string) (*config.PolarisScaffold, error) {
	components, err := ListComponents(polarisHome, polarisConfig, componentName)

	if err != nil {
		return nil, err
	}

	if len(components) != 1 {
		return nil, fmt.Errorf("Unable to find component with name %s", componentName)
	}

	return components[componentName], nil
}
