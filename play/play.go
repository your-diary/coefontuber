package play

import "errors"
import "log"
import "os/exec"

func Play(wavFile string) {
	var err = exec.Command("afplay", wavFile).Run()
	var e *exec.ExitError
	if errors.As(err, &e) {
		log.Printf("Failed to play the file [ %v ]: (%v) %v\n", wavFile, e.ProcessState.ExitCode(), string(e.Stderr))
		return
	} else if err != nil {
		log.Printf("Failed to play the file [ %v ]: %v\n", wavFile, err)
		return
	}
}
