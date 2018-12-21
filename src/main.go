package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	git "gopkg.in/src-d/go-git.v4"

	"github.com/synthesis-labs/polaris-cli/src/config"
	"github.com/synthesis-labs/polaris-cli/src/repo"
	"github.com/urfave/cli"
)

func main() {
	// Get current users home folder
	//
	polarisHome, polarisConfig := config.GetConfig()

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
								if err != nil {
									log.Fatal(err)
								}

								return nil
							},
						},
					},
				},
				{
					Name:  "list",
					Usage: "List available scaffolds in all repositories",
					Action: func(c *cli.Context) error {
						fmt.Println("Listing scaffolds...")
						return nil
					},
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
