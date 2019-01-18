package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/prisoner/genx"
	"github.com/urfave/cli"
)

type sflags []*[2]string

func flattenFlags(in []string) (out sflags) {
	for _, f := range in {
		for _, p := range strings.Split(f, ",") {
			kv := strings.Split(p, "=")
			switch len(kv) {
			case 2:
				out = append(out, &[2]string{strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1])})
			case 1:
				out = append(out, &[2]string{strings.TrimSpace(kv[0])})
			}
		}
	}
	return
}

func main() {
	log.SetFlags(log.Lshortfile)
	cli.VersionFlag = &cli.BoolFlag{
		Name:  "version",
		Usage: "print the version",
	}

	app := &cli.App{
		Name:    "genx",
		Usage:   "Generics For Go, Yet Again.",
		Version: "v0.5",
		Authors: []cli.Author{{
			Name:  "Ahmed <OneOfOne> W.",
			Email: "oneofone+genx <a.t> gmail <dot> com",
		},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "seed",
				Usage: "alias for -pkg github.com/OneOfOne/genx/seeds/`seed-name`",
			},
			&cli.StringFlag{
				Name:  "in,f",
				Usage: "`file` to process, use `-` to process stdin.",
			},

			&cli.StringFlag{
				Name:  "package,pkg",
				Usage: "`package` to process.",
			},

			&cli.StringFlag{
				Name:  "name,n",
				Usage: "package `name` to use for output, uses the input package's name by default.",
			},

			&cli.StringSliceFlag{
				Name:  "type,t",
				Usage: "generic `type` names to remove or rename (ex: -t 'KV=string,KV=interface{}' -t RemoveThisType).",
			},

			&cli.StringSliceFlag{
				Name:  "selector,s",
				Usage: "`selector`s to remove or rename (ex: -s 'cm.HashFn=hashers.Fnv32' -s 'x.Call=Something').",
			},

			&cli.StringSliceFlag{
				Name:  "field,fld",
				Usage: "struct `field`s to remove or rename (ex: -fld HashFn -fld privateFunc=PublicFunc).",
			},

			&cli.StringSliceFlag{
				Name:  "func,fn",
				Usage: "`func`tions to remove or rename (ex: -fn NotNeededFunc -fn Something=SomethingElse).",
			},

			&cli.StringFlag{
				Name:  "out,o",
				Value: "/dev/stdout",
				Usage: "output dir if parsing a package or output filename if you want the output to be merged.",
			},

			&cli.StringSliceFlag{
				Name:  "tags",
				Usage: "go extra build tags, used for parsing and automatically passed to any go subcommands.",
			},

			&cli.StringSliceFlag{
				Name:  "goFlags",
				Usage: "extra flags to pass to go subcommands `flags` (ex: --goFlags '-race')",
			},

			&cli.BoolFlag{
				Name:  "get",
				Usage: "go get the package if it doesn't exist",
			},

			&cli.BoolFlag{
				Name:  "verbose,v",
				Usage: "verbose output",
			},
		},
		Action: runGen,
	}

	// TODO: support other actions
	app.Run(os.Args)
}

func runGen(c *cli.Context) error {
	rewriters := map[string]string{}

	for _, kv := range flattenFlags(c.StringSlice("type")) {
		key, val := kv[0], kv[1]
		if key == "" {
			continue
		}
		if val == "" {
			val = "-"
		}
		rewriters["type:"+key] = val
	}

	for _, kv := range flattenFlags(c.StringSlice("selector")) {
		key, val := kv[0], kv[1]
		if key == "" {
			continue
		}
		if val == "" {
			val = "-"
		}
		rewriters["selector:"+key] = val
	}

	for _, kv := range flattenFlags(c.StringSlice("field")) {
		key, val := kv[0], kv[1]

		if key == "" {
			continue
		}
		if val == "" {
			val = "-"
		}
		rewriters["field:"+key] = val
	}

	for _, kv := range flattenFlags(c.StringSlice("func")) {
		key, val := kv[0], kv[1]

		if key == "" {
			continue
		}
		if val == "" {
			val = "-"
		}
		rewriters["func:"+key] = val
	}

	g := genx.New(c.String("name"), rewriters)
	g.BuildTags = append(g.BuildTags, c.StringSlice("tags")...)

	if c.Bool("verbose") {
		log.Printf("rewriters: %+q", g.OrderedRewriters())
		log.Printf("build tags: %+q", g.BuildTags)
	}

	var (
		outPath = c.String("out")

		mergeFiles bool
		inPkg      string
	)

	switch outPath {
	case "", "-", "/dev/stdout":
		outPath = "/dev/stdout"
		mergeFiles = true
	}

	// auto merge files if the output is a file not a dir.
	mergeFiles = !mergeFiles && filepath.Ext(outPath) == ".go"

	if seed := c.String("seed"); seed != "" {
		inPkg = "github.com/OneOfOne/genx/seeds/" + seed
		mergeFiles = true
	} else {
		inPkg = c.String("package")
	}

	if inPkg != "" {
		out, err := goListThenGet(c, g.BuildTags, inPkg)
		if err != nil {
			return cli.NewExitError(err, 2)
		}

		inPkg = out
		pkg, err := g.ParsePkg(inPkg, false)

		if err != nil {
			return cli.NewExitError(fmt.Sprintf("error parsing package (%s): %v\n", inPkg, err), 1)
		}

		if mergeFiles {
			err = pkg.WriteAllMerged(outPath, false)
		} else {
			err = pkg.WritePkg(outPath)
		}

		if err != nil {
			return cli.NewExitError(err, 1)
		}
	} else if inFile := c.String("in"); inFile != "" {
		if inFile != "-" && inFile != "/dev/stdin" {
			out, err := goListThenGet(c, g.BuildTags, inFile)
			if err != nil {
				return cli.NewExitError(err, 2)
			}
			inFile = out
		}

		pf, err := g.Parse(inFile, nil)
		if err != nil {
			return cli.NewExitError(fmt.Sprintf("error parsing file (%s): %v\n%s", inFile, err, pf.Src), 1)
		}

		if err := pf.WriteFile(outPath); err != nil {
			return cli.NewExitError(err, 1)
		}
	} else {
		log.Println(c.FlagNames())
		// cli.ShowAppHelpAndExit(c, 1)
	}

	return nil
}

func execCmd(ctx *cli.Context, c string, args ...string) (string, error) {
	cmd := exec.Command(c, args...)
	if ctx.Bool("verbose") {
		log.Printf("executing: %s %s", c, strings.Join(args, " "))
	}
	out, err := cmd.Output()
	ret := strings.TrimSpace(string(out))
	return ret, err
}

func goListThenGet(ctx *cli.Context, tags []string, path string) (out string, err error) {
	if _, err = os.Stat(path); err == nil {
		return path, nil
	}

	isFile := filepath.Ext(path) == ".go"
	dir := path
	if isFile {
		dir = filepath.Dir(path)
	}

	args := []string{"-tags", strings.Join(tags, " ")}
	args = append(args, ctx.StringSlice("goFlags")...)

	args = append(args, dir)

	listArgs := append([]string{"list", "-f", "{{.Dir}}"}, args...)

	if out, err = execCmd(ctx, "go", listArgs...); err != nil && strings.Contains(out, "cannot find package") {
		// log.Printf("err %+v, out =%s\n", err, out)
		if !ctx.Bool("get") {
			out = fmt.Sprintf("`%s` not found and `--get` isn't specified.", path)
			return
		}
		if out, err = execCmd(ctx, "go", append([]string{"get", "-u", "-v"}, args...)...); err == nil && isFile {
			out, err = execCmd(ctx, "go", listArgs...)
		}
	}

	if err == nil && isFile {
		out = filepath.Join(out, filepath.Base(path))
	}
	return
}
