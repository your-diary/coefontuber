package play

import "errors"
import "log"
import "os/exec"

func Play(filenameChannel <-chan string, batonIn <-chan struct{}, batonOut chan<- struct{}, additionalArgs []string) {

	defer close(batonOut)

	var wavFile = <-filenameChannel
	if wavFile == "" {
		return
	}
	<-batonIn

	var args = []string{wavFile}
	args = append(args, additionalArgs...)

	var err = exec.Command("play", args...).Run()

	var e *exec.ExitError
	if errors.As(err, &e) {
		log.Printf("Failed to play the file [ %v ]: (%v) %v\n", wavFile, e.ProcessState.ExitCode(), string(e.Stderr))
		return
	} else if err != nil {
		log.Printf("Failed to play the file [ %v ]: %v\n", wavFile, err)
		return
	}

}
