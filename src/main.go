package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	git "gopkg.in/src-d/go-git.v4"

	"github.com/synthesis-labs/polaris-cli/src/config"
	"github.com/synthesis-labs/polaris-cli/src/repo"
	"github.com/urfave/cli"
)

func main() {
	// Get current users home folder
	//
	polarisHome, polarisConfig := config.GetConfig()
	needSync, err := repo.NeedSynchronizeRepositories(polarisHome, polarisConfig)
	if err != nil {
		log.Fatal(err)
	}
	if needSync {
		err := repo.SynchronizeRepositories(polarisHome, polarisConfig, false)
		if err != nil {
			log.Fatal(err)
		}
	}

	app := cli.NewApp()
	app.Commands = []cli.Command{
		{
			Name:  "scaffold",
			Usage: "Scaffolding commands",
			Subcommands: []cli.Command{
				{
					Name:  "repo",
					Usage: "Scaffolding repository management",
					Subcommands: []cli.Command{
						{
							Name:  "list",
							Usage: "List repositories configured",
							Action: func(c *cli.Context) error {
								for repoName, repoConfig := range polarisConfig.Repositories {

									// Open the repository
									//
									repository, err := git.PlainOpen(fmt.Sprintf("%s/repos/%s", polarisHome, repoName))
									if err != nil {
										log.Fatal(err)
									}

									headRef, err := repository.Head()
									if err != nil {
										log.Fatal(err)
									}

									// Print whatever we need
									//
									fmt.Println(repoName, "->", repoConfig.URI, repoConfig.Ref, headRef.Hash().String()[:7])
								}
								return nil
							},
						},
						{
							Name:      "add",
							Usage:     "Add a repository",
							ArgsUsage: "<NAME> <URL> <Ref>",
							Action: func(c *cli.Context) error {
								if c.NArg() != 3 {
									cli.ShowCommandHelp(c, "add")
									return errors.New("Invalid number of arguments")
								}
								name := c.Args().Get(0)
								url := c.Args().Get(1)
								ref := c.Args().Get(2)

								polarisConfig.Repositories[name] = config.PolarisRepository{
									URI: url,
									Ref: fmt.Sprintf("refs/heads/%s", ref),
								}
								config.SaveConfig(polarisHome, polarisConfig)

								err := repo.SynchronizeRepositories(polarisHome, polarisConfig, false, name)
								if err != nil {
									log.Fatal(err)
								}
								return nil
							},
						},
						{
							Name:      "remove",
							Usage:     "Remove a repository",
							ArgsUsage: "<NAME>",
							Action: func(c *cli.Context) error {
								if c.NArg() != 1 {
									cli.ShowCommandHelp(c, "remove")
									return errors.New("Invalid number of arguments")
								}
								name := c.Args().Get(0)

								delete(polarisConfig.Repositories, name)
								config.SaveConfig(polarisHome, polarisConfig)

								// Must delete repository too
								//
								reposHome := fmt.Sprintf("%s/repos/", polarisHome)
								os.RemoveAll(fmt.Sprintf("%s/%s", reposHome, name))

								return nil
							},
						},
						{
							Name:  "update",
							Usage: "Update the scaffolds in all repositories",
							Flags: []cli.Flag{
								cli.BoolFlag{Name: "force", Usage: "Force a full refresh"},
							},
							Action: func(c *cli.Context) error {

								err := repo.SynchronizeRepositories(polarisHome, polarisConfig, c.Bool("force"))

								return err
							},
						},
					},
				},
				{
					Name:  "list",
					Usage: "List available scaffolds in all repositories",
					Action: func(c *cli.Context) error {

						scaffolds, err := repo.ListScaffolds(polarisHome, polarisConfig)
						if err != nil {
							log.Fatal(err)
						}

						for name, detail := range scaffolds {
							fmt.Println(name, "->", detail.Spec.Description)
						}

						return nil
					},
				},
				{
					Name:      "describe",
					ArgsUsage: "<NAME>",
					Usage:     "Describe a scaffold",
					Action: func(c *cli.Context) error {
						if c.NArg() != 1 {
							cli.ShowCommandHelp(c, "describe")
							return errors.New("Invalid number of arguments")
						}
						scaffoldName := c.Args().Get(0)

						scaffold, err := repo.GetScaffold(polarisHome, polarisConfig, scaffoldName)
						if err != nil {
							log.Fatal(err)
						}

						fmt.Println("Name:", scaffoldName)
						fmt.Println("Description:", scaffold.Spec.Description)
						fmt.Println("Help:", scaffold.Spec.Help)
						fmt.Println("Parameters:")
						for _, param := range scaffold.Spec.Parameters {
							fmt.Println(" -", param.Name, "default", param.Default)
						}

						return nil
					},
				},
				{
					Name:      "unpack",
					ArgsUsage: "<scaffold name> <local name> [--parameters name=value,name2=value2]",
					Usage:     "Unpack and deploy a scaffold locally",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "parameters", Usage: "Provide a template parameters"},
					},
					Action: func(c *cli.Context) error {
						if c.NArg() != 2 {
							cli.ShowCommandHelp(c, "unpack")
							return errors.New("Invalid number of arguments")
						}

						scaffoldName := c.Args().Get(0)
						localName := c.Args().Get(1)

						scaffold, err := repo.GetScaffold(polarisHome, polarisConfig, scaffoldName)
						if err != nil {
							log.Fatal(err)
						}

						// Check whether the local path already exists
						//
						localName = path.Clean(localName)
						if _, err := os.Stat(localName); !os.IsNotExist(err) {
							return fmt.Errorf("%s already exists", localName)
						}

						// Setup the project object for use by the template later
						//
						project := config.PolarisScaffoldProject{
							Name:       localName,
							Parameters: map[string]string{},
							Scaffold:   scaffold,
						}

						// Populate all the scaffold default parameter values first
						//
						for _, parameter := range project.Scaffold.Spec.Parameters {
							project.Parameters[parameter.Name] = parameter.Default
						}

						// Overwrite them with the ones provided on the command line
						//
						var parameters = c.String("parameters")
						for _, parameter := range strings.Split(parameters, ",") {
							split := strings.Split(parameter, "=")
							if len(split) == 2 {
								if _, contains := project.Parameters[split[0]]; !contains {
									return fmt.Errorf("Parameter %s provided by not in scaffold spec", split[0])
								}
								project.Parameters[split[0]] = split[1]
							}
						}

						for key, value := range project.Parameters {
							fmt.Println("Got param", key, "->", value)
						}

						err = filepath.Walk(scaffold.LocalPath, func(filename string, info os.FileInfo, err error) error {

							// filename -> file from the scaffold
							// localPath -> file to be written (in the target)

							localPath := fmt.Sprintf("%s%s", localName, strings.Replace(filename, fmt.Sprintf("%s", scaffold.LocalPath), "", 1))

							// localPath could be a templated name, so we must render it
							//
							localPathTmpl, err := template.
								New("PolarisFilenameTemplate").
								Funcs(template.FuncMap{}).
								Delims("[[", "]]").
								Parse(string(localPath))
							if err != nil {
								fmt.Println("Error during template parsing", localPath)
								return err
							}
							var localPathBuff bytes.Buffer

							err = localPathTmpl.Execute(&localPathBuff, project)
							if err != nil {
								return fmt.Errorf("Error during filename template generation: %s", err)
							}

							// Set the name to whatever the template rendered
							//
							localPath = localPathBuff.String()

							if info.IsDir() {
								os.Mkdir(localPath, os.ModePerm)
								fmt.Println("Created directory", localPath)
							} else {

								fileContents, err := ioutil.ReadFile(filename)
								if err != nil {
									return err
								}

								tmpl, err := template.
									New("PolarisScaffoldTemplate").
									Funcs(template.FuncMap{}).
									Delims("[[", "]]").
									Parse(string(fileContents))
								if err != nil {
									fmt.Println("Error during template parsing", localPath)
									return err
								}
								var buff bytes.Buffer

								err = tmpl.Execute(&buff, project)
								if err != nil {
									return fmt.Errorf("Error during template generation: %s", err)
								}

								ioutil.WriteFile(localPath, buff.Bytes(), 0644)
								fmt.Println("Wrote file", filename, localPath)
							}

							return nil
						})

						return err
					},
				},
			},
		},
	}

	err = app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
