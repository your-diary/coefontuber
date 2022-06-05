package play

import "fmt"
import "errors"
import "log"
import "os/exec"
import "runtime"

func Play(filenameChannel <-chan string, batonIn <-chan struct{}, batonOut chan<- struct{}) {
	defer close(batonOut)
	var wavFile = <-filenameChannel
	if wavFile == "" {
		return
	}
	<-batonIn
	var err = func() error {
		if runtime.GOOS == "darwin" {
			return exec.Command("afplay", wavFile).Run()
		} else if runtime.GOOS == "linux" {
			return exec.Command("play", "--no-show-progress", wavFile).Run()
		} else {
			return fmt.Errorf("unsupported platform")
		}
	}()
	var e *exec.ExitError
	if errors.As(err, &e) {
		log.Printf("Failed to play the file [ %v ]: (%v) %v\n", wavFile, e.ProcessState.ExitCode(), string(e.Stderr))
		return
	} else if err != nil {
		log.Printf("Failed to play the file [ %v ]: %v\n", wavFile, err)
		return
	}
	// log.Printf("Play: [ %v ]\n", wavFile)
}
