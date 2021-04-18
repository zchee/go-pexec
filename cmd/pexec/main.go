// SPDX-FileCopyrightText: Copyright 2021 The go-pexec Authors
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"

	json "github.com/goccy/go-json"
	yaml "github.com/goccy/go-yaml"
	"github.com/mattn/go-shellwords"
	exec "golang.org/x/sys/execabs"

	pexec "github.com/zchee/go-pexec"
)

var (
	flagDir               = flag.String("dir", "", "The directory to run the commands in")
	flagFastFail          = flag.Bool("fast-fail", false, "Fail on the first command failure")
	flagMaxConcurrentCmds = flag.Int("max-concurrent-cmds", runtime.NumCPU(), "Maximum number of processes to run concurrently, or unlimited if 0")
	flagNoLog             = flag.Bool("no-log", false, "Do not output logs")

	errUsage               = fmt.Errorf("usage: %s configFile", os.Args[0])
	errConfigNil           = errors.New("config is nil")
	errConfigCommandsEmpty = errors.New("config commands is empty")
)

type config struct {
	Dir      string   `json:"dir,omitempty" yaml:"dir,omitempty"`
	Commands []string `json:"commands,omitempty" yaml:"commands,omitempty"`
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("")
	flag.Parse()

	ctx := context.Background()
	if err := do(ctx); err != nil {
		log.Fatal(err)
	}
}

func do(ctx context.Context) error {
	if len(flag.Args()) != 1 {
		log.Fatal(errUsage.Error())
	}

	config, err := readConfig(flag.Args()[0])
	if err != nil {
		return err
	}

	if !*flagNoLog {
		data, err := json.Marshal(config)
		if err != nil {
			return err
		}
		log.Print(string(data))
	}

	cmds, err := getCmds(ctx, config, *flagDir)
	if err != nil {
		return err
	}

	runnerOptions := []pexec.RunnerOption{pexec.WithMaxConcurrentCmds(*flagMaxConcurrentCmds)}
	if *flagNoLog {
		runnerOptions = append(runnerOptions, pexec.WithEventHandler(func(*pexec.Event) {}))
	}

	if *flagFastFail {
		runnerOptions = append(runnerOptions, pexec.WithFastFail())
	}

	return pexec.NewRunner(runnerOptions...).Run(pexec.ExecCmds(ctx, cmds))
}

func readConfig(configFilePath string) (*config, error) {
	data, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}

	config := &config{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, err
	}

	if config.Dir == "" {
		config.Dir = filepath.Dir(configFilePath)
	} else if !filepath.IsAbs(config.Dir) {
		config.Dir = filepath.Join(filepath.Dir(configFilePath), config.Dir)
	}
	if err := validateConfig(config); err != nil {
		return nil, err
	}

	return config, nil
}

func validateConfig(config *config) error {
	if config == nil {
		return errConfigNil
	}

	if len(config.Commands) == 0 {
		return errConfigCommandsEmpty
	}

	return nil
}

func getCmds(ctx context.Context, config *config, dirPath string) ([]*exec.Cmd, error) {
	var cmds []*exec.Cmd
	for _, line := range config.Commands {
		if line == "" {
			continue
		}

		args, err := shellwords.Parse(line)
		if err != nil {
			return nil, err
		}

		// could happen if args = "$FOO" and FOO is not set
		if len(args) == 0 {
			continue
		}
		cmd := exec.CommandContext(ctx, args[0], args[1:]...)

		if dirPath != "" {
			cmd.Dir = dirPath
		} else {
			cmd.Dir = config.Dir
		}

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmds = append(cmds, cmd)
	}

	return cmds, nil
}
