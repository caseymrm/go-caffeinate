package caffeinate

import (
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// Caffeinate represents an invokation of the osx 'caffeinate' command
type Caffeinate struct {
	//-d: Prevent the display from sleeping.
	Display bool
	//-i: Prevent the system from idle sleeping.
	IdleSystem bool
	//-m: Prevent the disk from idle sleeping.
	IdleDisk bool
	//-s: Prevent the system from sleeping. Valid only on AC power.
	System bool
	//-u: Declare that user is active. Turns display on if off. Default 5 second timeout if no timeout set.
	UserActive bool
	//-t: Specifies the timeout value in seconds
	Timeout int
	//-w: Specifies pid to wait for.
	PID int

	cmd         *exec.Cmd
	running     bool
	waitChannel chan bool
	waitError   error
}

// Run the caffeinate command with these settings
func (c *Caffeinate) Run() {
	if c.running {
		c.Stop()
	}
	args := make([]string, 0)
	if c.Display {
		args = append(args, "-d")
	}
	if c.IdleSystem {
		args = append(args, "-i")
	}
	if c.IdleDisk {
		args = append(args, "-m")
	}
	if c.System {
		args = append(args, "-s")
	}
	if c.UserActive {
		args = append(args, "-u")
	}
	if c.Timeout > 0 {
		args = append(args, "-t", strconv.Itoa(c.Timeout))
	}
	if c.PID > 0 {
		args = append(args, "-w", strconv.Itoa(c.PID))
	} else {
		args = append(args, "-w", strconv.Itoa(os.Getpid()))
	}
	log.Printf("Command: %s %s", "/usr/bin/caffeinate", strings.Join(args, " "))
	c.cmd = exec.Command("/usr/bin/caffeinate", args...)
	if err := c.cmd.Start(); err != nil {
		log.Fatal("error starting process: ", err)
	}
	go c.waitForProcess()
	c.running = true
}

// Stop this caffeinate process
func (c *Caffeinate) Stop() error {
	if err := c.cmd.Process.Kill(); err != nil {
		return err
	}
	return c.Wait()
}

// Wait blocks until the caffeinate command exits
func (c *Caffeinate) Wait() error {
	if !c.running {
		return nil
	}
	<-c.waitChannel
	return c.waitError
}

// Running returns whether caffeinate is currently running
func (c *Caffeinate) Running() bool {
	return c.running
}

func (c *Caffeinate) waitForProcess() {
	c.waitChannel = make(chan bool)
	c.waitError = c.cmd.Wait()
	c.running = false
	close(c.waitChannel)
	c.waitChannel = nil
}
