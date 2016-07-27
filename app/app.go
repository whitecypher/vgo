package app

import (
	"path/filepath"
	"strings"
	"fmt"
	"os"
	"os/exec"

	"github.com/codegangsta/cli"

)

var (
	Version = "v?.?.?"
)

type App struct {
	cli *cli.App
	root string
	gosrc string
}

func New(root, gopath string) *App {
	gosrc := filepath.Join(gopath, "src")
	if !strings.HasPrefix(root, gosrc) {
		fmt.Println("Your project isn't in the gopath. We haven't tested this with VGO yet so we recommend you move you project into your gopath.")
		os.Exit(1)
	}

	name, err := filepath.Rel(gosrc, root)
	if err != nil {
		name = filepath.Base(root)
	}
	_ = name
	//r := NewRepo(name, nil, resolveManifestFilePath(cwd))

	vgo := cli.NewApp()
	vgo.Name = "vgo"
	vgo.Usage = "Installs the dependencies listed in the manifest at the designated reference point.\n   If no manifest exists, use `vgo discover` to resolve dependencies and create one."
	vgo.Version = Version
	vgo.EnableBashCompletion = true
	vgo.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "dry",
			Usage: "Prevent updates to manifest for trial runs",
		},
	}
	vgo.Before = func(c *cli.Context) (err error) {
		//r.LoadManifest()
		return
	}
	vgo.After = func(c *cli.Context) (err error) {
		//if !c.Bool("dry") {
		//	err = r.SaveManifest()
		//	if err != nil {
		//		fmt.Println(err.Error())
		//	}
		//} else {
		//	// r.Print("  ", os.Stdout)
		//	data, err := yaml.Marshal(r)
		//	if err != nil {
		//		fmt.Println(err.Error())
		//	}
		//	fmt.Println(string(data))
		//}
		return
	}
	 //vgo.Authors = []cli.Author{
	 //	{
	 //		Name:  "Merten van Gerven",
	 //		Email: "merten.vg@gmail.com",
	 //	},
	 //}
	vgo.Commands = []cli.Command{
		{
			Name:        "discover",
			Aliases:     []string{"q"},
			Usage:       "Discover dependencies",
			Description: `Scan project for packages, install them if not already vendored and store results into vgo.yaml`,
			Action: func(c *cli.Context) {
				//if len(r.Main) > 0 {
				//	for _, m := range r.Main {
				//		NewPkg(path.Join(name, m), cwd, nil)
				//	}
				//} else {
				//	NewPkg(name, cwd, nil)
				//}
			},
		},
		{
			Name:        "get",
			Aliases:     []string{"i"},
			Usage:       "Get a dependency",
			Description: `Get a dependency compatible with the optionally specified version, branch, tag, or commit`,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "u",
					Usage: "Update the package to the latest compatible reference",
				},
			},
			Action: func(c *cli.Context) {
			},
		},
		{
			Name:        "remove",
			Aliases:     []string{"d"},
			Usage:       "Remove a dependency",
			Description: `Remove one or more dependencies matching the given paths`,
			Action: func(c *cli.Context) {
			},
		},
		{
			Name:        "main",
			Aliases:     []string{"m"},
			Usage:       "Add an entrypoint (main package)",
			Description: `Add a main (entrypoint) package to the project manifest`,
			Action: func(c *cli.Context) {
				//paths := c.Args()
				//for _, path := range paths {
				//	r.AddMain(path)
				//}
			},
		},
		{
			Name:        "unmain",
			Aliases:     []string{"r"},
			Usage:       "Remove an entrypoint",
			Description: `Remove a main (entrypoint) package from the project manifest`,
			Action: func(c *cli.Context) {
				//paths := c.Args()
				//for _, path := range paths {
				//	r.RemoveMain(path)
				//}
			},
		},
		//{
		//	Name:        "version",
		//	Aliases:     []string{"v"},
		//	Usage:       "Show the version",
		//	Description: `Show the version of your currently installed vgo tool`,
		//	Action: func(c *cli.Context) {
		//		fmt.Println(Version)
		//		//paths := c.Args()
		//		//for _, path := range paths {
		//		//	r.RemoveMain(path)
		//		//}
		//	},
		//},
	}
	vgo.Action = func(c *cli.Context) {
		//if r.hasManifest {
		//	r.InstallDeps()
		//} else {
		//	Log("No manifest found. Running discover task.")
		//	if len(r.Main) > 0 {
		//		for _, m := range r.Main {
		//			NewPkg(path.Join(name, m), cwd, nil)
		//		}
		//	} else {
		//		NewPkg(name, cwd, nil)
		//	}
		//}
		// pass command through to go
		args := c.Args()
		if len(args) > 0 {
			if args[0] == "--dry" {
				args = args[1:]
			}
			cmd := exec.Command("go", args...)
			cmd.Stdout = os.Stdout
			cmd.Stdin = os.Stdin
			cmd.Stderr = os.Stderr
			cmd.Run()
		}
	}

	return &App{
		cli: vgo,
		root: root,
		gosrc: gosrc,
	}
}

func (a *App) Discover() {

}

func (a *App) Get() {

}

func (a *App) Remove() {

}

func (a *App) RegisterMain() {

}

func (a *App) DeregisterMain() {

}

func (a *App) Run(args []string) error {
	return a.cli.Run(args)
}
