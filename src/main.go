package main

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

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
							fmt.Println(" -", param.Name, "type", param.Type)
						}

						return nil
					},
				},
				{
					Name:      "unpack",
					ArgsUsage: "<scaffold name> <local name> [parameters]",
					Usage:     "Unpack and deploy a scaffold locally",
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

						unpackApp := cli.NewApp()

						unpackApp.Flags = []cli.Flag{}
						for _, param := range scaffold.Spec.Parameters {
							flag := cli.StringFlag{
								Name: param.Name,
							}
							unpackApp.Flags = append(unpackApp.Flags, flag)
						}

						unpackApp.Action = func(c2 *cli.Context) error {
							fmt.Println("Do the stuff here with some args hopefully")

							for _, param := range scaffold.Spec.Parameters {
								fmt.Println("Got param", param.Name, c2.String(param.Name))
							}

							project := config.PolarisScaffoldProject{
								Name:       localName,
								Parameters: map[string]string{},
							}

							err = filepath.Walk(scaffold.LocalPath, func(filename string, info os.FileInfo, err error) error {
								localPath := fmt.Sprintf("%s%s", localName, strings.Replace(filename, fmt.Sprintf("%s", scaffold.LocalPath), "", 1))

								if info.IsDir() {
									os.Mkdir(localPath, os.ModePerm)
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
										return err
									}
									var buff bytes.Buffer

									err = tmpl.Execute(&buff, project)

									ioutil.WriteFile(localPath, buff.Bytes(), 0644)

									fmt.Println(filename, localPath)
								}

								return nil
							})

							return nil
						}

						err = unpackApp.Run(c.Args())

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
