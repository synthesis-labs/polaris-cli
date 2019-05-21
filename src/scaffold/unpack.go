package scaffold

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/synthesis-labs/polaris-cli/src/options"

	"github.com/synthesis-labs/polaris-cli/src/config"
	yaml "gopkg.in/yaml.v2"
)

// unpackScaffold low level unpacking of a template from a repo to a local path
//
func unpackScaffold(polarisType string, scaffold *config.PolarisScaffold, scaffoldValues interface{}, localPath string, overwrite bool) error {
	// Clean paths
	//
	localPath = path.Clean(localPath)

	err := filepath.Walk(fmt.Sprintf("%s/", scaffold.LocalPath), func(sourcePath string, info os.FileInfo, err error) error {

		if err != nil {
			fmt.Println(err)
			return err
		}

		// Source files to ignore
		//
		if info.Name() == "polaris-project.yaml" || info.Name() == "polaris-component.yaml" {
			return nil
		}

		// filename -> file from the scaffold
		// targetPath -> file to be written (in the target)

		targetPath := fmt.Sprintf("%s%s", localPath, strings.Replace(sourcePath, fmt.Sprintf("%s", scaffold.LocalPath), "", 1))

		if options.IsVerbose() {
			fmt.Println("scaffold.LocalPath", scaffold.LocalPath)
			fmt.Println("--------------------")
			fmt.Println("sourcePath", sourcePath)
			fmt.Println("targetPath", targetPath)
			fmt.Println("--------------------")
		}

		// targetPath could be a templated name, so we must render it
		//
		targetPathTemplate, err := template.
			New("PolarisFilenameTemplate").
			Funcs(template.FuncMap{}).
			Delims("[[", "]]").
			Parse(string(targetPath))
		if err != nil {
			fmt.Println("Error during template parsing", targetPath)
			return err
		}
		var targetPathBuff bytes.Buffer

		err = targetPathTemplate.Execute(&targetPathBuff, scaffoldValues)
		if err != nil {
			return fmt.Errorf("Error during filename template generation: %s", err)
		}

		// Set the name to whatever the template rendered
		//
		targetPath = targetPathBuff.String()

		if info.IsDir() {
			err := os.MkdirAll(targetPath, os.ModePerm)
			if err != nil {
				return err
			}
			if options.IsVerbose() {
				fmt.Println("Created directory", targetPath)
			}
		} else {

			sourceContents, err := ioutil.ReadFile(sourcePath)
			if err != nil {
				return err
			}

			var buff bytes.Buffer
			if !shouldExcludeFile(sourcePath) {
				tmpl, err := template.
					New(fmt.Sprintf("PolarisScaffoldTemplate:%s", sourcePath)).
					Funcs(template.FuncMap{}).
					Delims("[[", "]]").
					Parse(string(sourceContents))
				if err != nil {
					fmt.Println("Error during template parsing", localPath)
					return err
				}

				err = tmpl.Execute(&buff, scaffoldValues)
				if err != nil {
					return fmt.Errorf("Error during template generation: %s", err)
				}
			} else {
				buff.Write(sourceContents)
			}

			if _, err := os.Stat(targetPath); !os.IsNotExist(err) && !overwrite {
				return fmt.Errorf("%s already exists", localPath)
			}
			err = ioutil.WriteFile(targetPath, buff.Bytes(), 0644)
			if err != nil {
				return err
			}
			if options.IsVerbose() {
				fmt.Println("Wrote file", sourcePath, targetPath)
			}
		}

		return nil
	})

	// Any errors from templating or walking
	//
	if err != nil {
		return err
	}

	// Write the values to the base/polaris.yaml if the polaris-type is specified
	//
	if polarisType != "" {
		projectMarshalled, err := yaml.Marshal(scaffoldValues)
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(fmt.Sprintf("%s/polaris-%s.yaml", localPath, polarisType), projectMarshalled, 0644)
	}

	return err
}

// GetLocalProject scans the local directory for a polaris-%s.yaml (project or whatever) and returns it
//
func GetLocalProject(polarisType string) (*config.PolarisProject, error) {

	projectData, err := ioutil.ReadFile(fmt.Sprintf("./polaris-%s.yaml", polarisType))
	if err != nil {
		return nil, err
	}

	project := config.PolarisProject{}
	err = yaml.Unmarshal(projectData, &project)
	if err != nil {
		return nil, err
	}

	return &project, nil
}

// UnpackProject unpacks an Application scaffold into the local path
//
func UnpackProject(scaffold *config.PolarisScaffold, parameters map[string]string, localPath string, overwrite bool) error {
	// Clean paths
	//
	localName := path.Clean(localPath)

	// Setup the project object for use by the template later
	//
	project := config.PolarisProject{
		Project:    localName,
		Parameters: map[string]string{},
		Scaffold:   scaffold.Name,
	}

	// Populate all the scaffold default parameter values first
	//
	for _, parameter := range scaffold.Spec.Parameters {
		project.Parameters[parameter.Name] = parameter.Default
	}

	// Overwrite them with the ones provided
	//
	for paramKey, paramValue := range parameters {
		project.Parameters[paramKey] = paramValue
	}

	// Print them
	//
	if options.IsVerbose() {
		for paramKey, paramValue := range project.Parameters {
			fmt.Println("Project parameter", paramKey, paramValue)
		}
	}

	err := unpackScaffold("project", scaffold, &project, localPath, overwrite)
	return err

}

// UnpackComponent unpacks a Component scaffold into the local path
//
func UnpackComponent(componentScaffold *config.PolarisScaffold, project *config.PolarisProject, parameters map[string]string, componentName string, localPath string, overwrite bool) error {
	// Clean paths
	//
	localName := path.Clean(localPath)

	// Setup the project object for use by the template later
	//
	component := config.PolarisComponent{
		Project:           project.Project,
		Component:         localName,
		Parameters:        map[string]string{},
		ProjectParameters: project.Parameters,
		ProjectScaffold:   project.Scaffold,
		ComponentScaffold: componentName,
	}

	// Populate all the scaffold default parameter values first
	//
	for _, parameter := range componentScaffold.Spec.Parameters {
		if options.IsVerbose() {
			fmt.Println("Got component spec param", parameter.Name)
		}
		component.Parameters[parameter.Name] = parameter.Default
	}

	// Overwrite them with the ones provided on the command line
	//
	for paramKey, paramValue := range parameters {
		if options.IsVerbose() {
			fmt.Println("Got component param", paramKey, "->", paramValue)
		}
		component.Parameters[paramKey] = paramValue
	}

	// Print them
	//
	if options.IsVerbose() {
		for paramKey, paramValue := range component.Parameters {
			fmt.Println("Component parameter", paramKey, paramValue)
		}
	}

	err := unpackScaffold("", componentScaffold, &component, ".", overwrite)
	return err
}

func shouldExcludeFile(filePath string) bool {
	// Files to be excluded from templatizing
	//
	var excludedFiles = []string{".jar"}

	for _, exclude := range excludedFiles {
		if strings.Contains(filePath, exclude) {
			return true
		}
	}
	return false
}
