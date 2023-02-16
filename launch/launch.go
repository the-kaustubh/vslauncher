package launch

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/tidwall/jsonc"
)

type LaunchConfiguration struct {
	Name    string            `json:"name"`
	Type    string            `json:""`
	Request string            `json:"request"`
	Mode    string            `json:"mode"`
	Program string            `json:"program"`
	Env     map[string]string `json:"env"`
}

func LaunchConfigurationFromString(confStr string) (LaunchConfiguration, error) {
	var conf LaunchConfiguration
	err := json.Unmarshal(jsonc.ToJSON([]byte(confStr)), &conf)
	if err != nil {
		log.Println("error while Unmarshal - ", err)
	}

	return conf, err
}

func Start(launchConf LaunchConfiguration) (*os.Process, error) {

	// resolve path
	pathToMainFile := strings.ReplaceAll(launchConf.Program, "${workspaceFolder}/", "")
	command := exec.Command("go", "run", filepath.Base(pathToMainFile))
	command.Dir = filepath.Dir(pathToMainFile)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	var envs []string
	for key, value := range launchConf.Env {
		envs = append(envs, key+"="+value)
	}
	goenvs, err := getGoEnv()
	if err != nil {
		log.Fatal("error while getting go env", err)
	}

	log.Println("PORT env var", os.Getenv("PORT"))
	allEnvs := append(command.Env, goenvs...)
	allEnvs = append(allEnvs, envs...)
	allEnvs = append(allEnvs, os.Environ()...)

	log.Println(allEnvs)
	command.Env = allEnvs

	err = command.Start()
	if err != nil {
		log.Println("error while running program: ", err)
		return command.Process, err
	}
	return command.Process, nil
}

func getGoEnv() ([]string, error) {
	var buf bytes.Buffer
	cmd := exec.Command("go", "env")
	cmd.Stdout = &buf

	err := cmd.Run()
	noDblQuotes := strings.ReplaceAll(buf.String(), "\"", "")
	return strings.Split(noDblQuotes, "\n"), err
}
