package runpwsh 

import (
        "os/exec"
        "bytes"
        "runtime"
        "strings"
        "errors"
)

func runCommand(args ...string) (string, error) {
        cmd := exec.Command(args[0], args[1:]...)

        var out bytes.Buffer
        var err bytes.Buffer

        cmd.Stdout = &out 
        cmd.Stderr = &err
        cmd.Run()

        // convert err to an error type if there is an error returned
        var e error
        if err.String() != "" {
                e = errors.New(err.String())
        }

        return strings.TrimRight(out.String(), "\r\n"), e
}

func RunPowershellCommand(command string) (string, error) {
        var pscommand string

        if runtime.GOOS == "windows" {
                pscommand = "powershell.exe"
        } else {
                pscommand = "pwsh"
        }

        out, err := runCommand(pscommand, "-command", command) 

        return out, err
}