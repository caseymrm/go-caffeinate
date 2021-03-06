package caffeinate

import (
	"log"
	"os"
	"os/exec"
	"strconv"
	"sync"
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

	mutex       sync.Mutex
	cmd         *exec.Cmd
	running     bool
	waitChannel chan bool
	waitError   error
}

// Start the caffeinate command with these settings
func (c *Caffeinate) Start() {
	c.mutex.Lock()
	if c.waitChannel == nil {
		c.waitChannel = make(chan bool)
	}
	for c.running {
		c.mutex.Unlock()
		c.Stop()
		c.mutex.Lock()
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
	c.cmd = exec.Command("/usr/bin/caffeinate", args...)
	if err := c.cmd.Start(); err != nil {
		log.Fatal("error starting process: ", err)
	}
	go c.waitForProcess()
	c.running = true
	c.mutex.Unlock()
}

// Stop this caffeinate process
func (c *Caffeinate) Stop() error {
	c.mutex.Lock()
	if !c.running {
		c.mutex.Unlock()
		return nil
	}
	if err := c.cmd.Process.Kill(); err != nil {
		c.mutex.Unlock()
		return err
	}
	c.mutex.Unlock()
	return c.Wait()
}

// Wait blocks until the caffeinate command exits
func (c *Caffeinate) Wait() error {
	c.mutex.Lock()
	if !c.running {
		c.mutex.Unlock()
		return nil
	}
	channel := c.waitChannel
	c.mutex.Unlock()
	<-channel
	return c.waitError
}

// Running returns whether caffeinate is currently running
func (c *Caffeinate) Running() bool {
	return c.running
}

// CaffeinatePID returns the pid of caffeinate, if it's running
func (c *Caffeinate) CaffeinatePID() int {
	if !c.running {
		return 0
	}
	return c.cmd.Process.Pid
}

func (c *Caffeinate) waitForProcess() {
	c.waitError = c.cmd.Wait()
	c.mutex.Lock()
	c.running = false
	oldWaitChannel := c.waitChannel
	c.waitChannel = make(chan bool)
	c.mutex.Unlock()
	close(oldWaitChannel)
}
