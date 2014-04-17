package main

import (
	"fmt"
	"os/exec"
)

func launchBrowser(port int) error {
	// Use xdg-open to open the default browser
	url := fmt.Sprintf("http://localhost:%d", port)

	cmd := exec.Command("open", url)

	if err := cmd.Start(); err != nil {
		return err
	}

	go func() {
		cmd.Wait()
	}()

	return nil
}
