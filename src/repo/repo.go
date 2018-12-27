package repo

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/synthesis-labs/polaris-cli/src/config"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	yaml "gopkg.in/yaml.v2"
)

// SynchronizeRepositories synchronizes repositories to local /.polaris folder
//
func SynchronizeRepositories(polarisHome string, polarisConfig *config.PolarisConfig, force bool, onlyThese ...string) error {
	reposHome := fmt.Sprintf("%s/repos", polarisHome)

	// Remove and recreate the repos folder if force is set
	//
	if force {
		fmt.Println("Removing existing repositories")
		os.RemoveAll(reposHome)
		os.MkdirAll(reposHome, os.ModePerm)
	}

	for repoName, repoConfig := range polarisConfig.Repositories {
		// If we are being told to filter
		//
		if len(onlyThese) > 0 {
			found := false
			for _, v := range onlyThese {
				if repoName == v {
					found = true
				}
			}
			if !found {
				continue
			}
		}

		fmt.Println("Syncing repository", repoName, repoConfig.URI, repoConfig.Ref)

		_, err := git.PlainClone(fmt.Sprintf("%s/repos/%s", polarisHome, repoName), false, &git.CloneOptions{
			URL:           repoConfig.URI,
			SingleBranch:  true,
			ReferenceName: plumbing.ReferenceName(repoConfig.Ref),
			Progress:      os.Stdout,
		})
		if err == git.ErrRepositoryAlreadyExists {
			fmt.Println(" .. already exists")
			repository, err := git.PlainOpen(fmt.Sprintf("%s/repos/%s", polarisHome, repoName))
			if err != nil {
				return err
			}
			fmt.Println("Pulling repository", repoName, repoConfig.URI, repoConfig.Ref)

			worktree, err := repository.Worktree()
			if err != nil {
				return err
			}
			err = worktree.Pull(&git.PullOptions{RemoteName: "origin", ReferenceName: plumbing.ReferenceName(repoConfig.Ref)})
			if err == git.NoErrAlreadyUpToDate {
				fmt.Println(" .. already up to date")
			} else if err != nil {
				return err
			}

		} else if err != nil {
			return err
		} else {
			fmt.Println("Cloned repository", repoName, repoConfig.URI, repoConfig.Ref)
		}
	}

	// Update the "lastsync" file to remember this sync
	//
	file, err := os.OpenFile(fmt.Sprintf("%s/.lastsync", reposHome), os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	_, err = file.Write([]byte{0x00})
	if err != nil {
		return err
	}
	err = file.Close()
	if err != nil {
		return err
	}

	return nil
}

// NeedSynchronizeRepositories checks whether we need to resync based on when last we synced
//
func NeedSynchronizeRepositories(polarisHome string, polarisConfig *config.PolarisConfig) (bool, error) {
	reposHome := fmt.Sprintf("%s/repos", polarisHome)
	fileinfo, err := os.Stat(fmt.Sprintf("%s/.lastsync", reposHome))
	if os.IsNotExist(err) {
		return true, nil
	} else if err != nil {
		return false, err
	}

	sinceLastSync := time.Since(fileinfo.ModTime())
	return sinceLastSync > 24*time.Hour, nil
}

// ListScaffolds returns the list of available scaffolds in all repositories
//
func ListScaffolds(polarisHome string, polarisConfig *config.PolarisConfig, matchingNames ...string) (map[string]*config.PolarisScaffold, error) {

	reposHome := path.Clean(fmt.Sprintf("%s/repos", polarisHome))

	result := map[string]*config.PolarisScaffold{}

	err := filepath.Walk(reposHome, func(filename string, info os.FileInfo, err error) error {
		filebase := path.Base(filename)
		if filebase == "scaffold.yaml" {
			scaffoldName := strings.Replace(strings.Replace(filename, fmt.Sprintf("%s/", reposHome), "", 1), "/scaffold.yaml", "", 1)

			scaffoldData, err := ioutil.ReadFile(filename)
			if err != nil {
				return err
			}

			scaffold := config.PolarisScaffold{}
			err = yaml.Unmarshal(scaffoldData, &scaffold.Spec)
			if err != nil {
				return err
			}

			scaffold.LocalPath = path.Dir(filename)
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

// GetScaffold returns a particular scaffold
//
func GetScaffold(polarisHome string, polarisConfig *config.PolarisConfig, scaffoldName string) (*config.PolarisScaffold, error) {
	scaffolds, err := ListScaffolds(polarisHome, polarisConfig, scaffoldName)

	if err != nil {
		return nil, err
	}

	if len(scaffolds) != 1 {
		return nil, fmt.Errorf("Unable to find scaffold with name %s", scaffoldName)
	}

	return scaffolds[scaffoldName], nil
}
