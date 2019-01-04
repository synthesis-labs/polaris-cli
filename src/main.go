package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	git "gopkg.in/src-d/go-git.v4"

	"github.com/synthesis-labs/polaris-cli/src/cluster"
	"github.com/synthesis-labs/polaris-cli/src/config"
	"github.com/synthesis-labs/polaris-cli/src/repo"
	"github.com/synthesis-labs/polaris-cli/src/scaffold"
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
			Name:      "init",
			ArgsUsage: "",
			Usage:     "Install the polaris operator to the cluster",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "namespace", Usage: "Namespace to use"},
			},
			Action: func(c *cli.Context) error {

				// Connect
				//
				client, apiextensionsClient, ns, err := cluster.ConnectToCluster()

				// Check if namespace is overridden on cmdline (otherwise use the default configured one)
				//
				if c.String("namespace") != "" {
					ns = c.String("namespace")
				}

				if err != nil {
					return err
				}

				// Find the polaris-operator in the list of pods
				//
				err = cluster.GetPolarisOperator(client, apiextensionsClient, ns)
				if err != nil {
					return err
				}

				fmt.Println("Succesfully installed polaris-operator to current cluster")
				return nil
			},
		},
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
			Name:  "project",
			Usage: "Project management",
			Subcommands: []cli.Command{
				{
					Name:  "list",
					Usage: "List projects available to scaffold",
					Action: func(c *cli.Context) error {

						scaffolds, err := repo.ListProjects(polarisHome, polarisConfig)
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
					Usage:     "Describe a project",
					Action: func(c *cli.Context) error {
						if c.NArg() != 1 {
							cli.ShowCommandHelp(c, "describe")
							return errors.New("Invalid number of arguments")
						}
						scaffoldName := c.Args().Get(0)

						scaffold, err := repo.GetProject(polarisHome, polarisConfig, scaffoldName)
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
					Name:      "new",
					ArgsUsage: "<local name>",
					Usage:     "Unpack a project locally",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "from", Usage: "From which project upstream (default: core/stable/starter/project)"},
						cli.BoolFlag{Name: "overwrite", Usage: "Allow overwriting of target files"},
						cli.StringFlag{Name: "parameters", Usage: "Provide template parameters"},
					},
					Action: func(c *cli.Context) error {
						if c.NArg() != 1 {
							cli.ShowCommandHelp(c, "new")
							return errors.New("Invalid number of arguments")
						}

						localName := c.Args().Get(0)

						var parametersOption = c.String("parameters")
						var parameters = map[string]string{}
						for _, parameter := range strings.Split(parametersOption, ",") {
							split := strings.Split(parameter, "=")
							if len(split) == 2 {
								parameters[split[0]] = split[1]
							}
						}

						var fromOption = c.String("from")
						if fromOption == "" {
							fromOption = "core/stable/starter/project"
						}

						applicationScaffold, err := repo.GetProject(polarisHome, polarisConfig, fromOption)
						if err != nil {
							return err
						}
						err = scaffold.UnpackProject(applicationScaffold, parameters, localName, c.Bool("overwrite"))
						if err != nil {
							return err
						}

						return nil
					},
				},
				{
					Name:      "status",
					ArgsUsage: "<local name>|.",
					Usage:     "Show status of project",
					Flags:     []cli.Flag{},
					Action: func(c *cli.Context) error {
						return nil
					},
				},
			},
		},
		{
			Name:  "component",
			Usage: "Component management",
			Subcommands: []cli.Command{
				{
					Name:  "list",
					Usage: "List components available to scaffold into your project",
					Action: func(c *cli.Context) error {

						components, err := repo.ListComponents(polarisHome, polarisConfig)
						if err != nil {
							log.Fatal(err)
						}

						for name, detail := range components {
							fmt.Println(name, "->", detail.Spec.Description)
						}

						return nil

					},
				},
				{
					Name:      "describe",
					ArgsUsage: "<NAME>",
					Usage:     "Describe a component",
					Action: func(c *cli.Context) error {
						if c.NArg() != 1 {
							cli.ShowCommandHelp(c, "describe")
							return errors.New("Invalid number of arguments")
						}

						return nil
					},
				},
				{
					Name:      "new",
					ArgsUsage: "<local name>",
					Usage:     "Unpack a component locally",
					Flags: []cli.Flag{
						cli.StringFlag{Name: "from", Usage: "From which component upstream (default: core/stable/starter/kotlin/microservice)"},
						cli.BoolFlag{Name: "overwrite", Usage: "Allow overwriting of target files"},
						cli.StringFlag{Name: "parameters", Usage: "Provide template parameters"},
					},
					Action: func(c *cli.Context) error {
						if c.NArg() != 1 {
							cli.ShowCommandHelp(c, "new")
							return errors.New("Invalid number of arguments")
						}

						localName := c.Args().Get(0)

						var parametersOption = c.String("parameters")
						var parameters = map[string]string{}
						for _, parameter := range strings.Split(parametersOption, ",") {
							split := strings.Split(parameter, "=")
							if len(split) == 2 {
								parameters[split[0]] = split[1]
							}
						}

						var fromOption = c.String("from")
						if fromOption == "" {
							fromOption = "core/stable/starter/kotlin/microservice"
						}

						// Read the project from the local directory?
						//
						project, err := scaffold.GetLocalProject("project")
						if err != nil {
							return err
						}

						// Read the scaffold from the repo (if we need it?)
						//
						/*
							_, err = repo.GetProject(polarisHome, polarisConfig, project.Scaffold)
							if err != nil {
								return err
							}
						*/

						// Read the component from the repo
						//
						componentScaffold, err := repo.GetComponent(polarisHome, polarisConfig, fromOption)
						if err != nil {
							return err
						}

						err = scaffold.UnpackComponent(componentScaffold, project, parameters, fromOption, localName, c.Bool("overwrite"))
						if err != nil {
							return err
						}

						return nil
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
