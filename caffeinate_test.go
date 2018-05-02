package caffeinate

import (
	"testing"
	"time"

	"github.com/caseymrm/go-assertions"
)

func TestBasic(t *testing.T) {
	startAsserts := assertions.GetAssertions()
	if startAsserts["PreventSystemSleep"] != 0 {
		t.Errorf("Somebody's already preventing system sleep!")
		return
	}
	c := Caffeinate{
		System:  true,
		Timeout: 2,
	}
	c.Run()
	time.Sleep(time.Second)
	endAsserts := assertions.GetAssertions()
	if endAsserts["PreventSystemSleep"] != 1 {
		t.Errorf("Failed to prevent system sleep")
		for k, v := range startAsserts {
			if endAsserts[k] != v {
				t.Errorf("%s diff: %v -> %v", k, startAsserts[k], endAsserts[k])
			}
		}
		return
	}
	if err := c.Wait(); err != nil {
		t.Errorf("Waiting: %v", err)
	}
}

func TestPreempt(t *testing.T) {
	c := Caffeinate{
		System:  true,
		Timeout: 2,
	}
	c.Run()
	time.Sleep(time.Second)
	if !c.Running() {
		t.Errorf("Not running after one second")
	}
	c.Timeout = 3
	c.Run()
	time.Sleep(2 * time.Second)
	if !c.Running() {
		t.Errorf("Not running after two seconds")
	}
	time.Sleep(2 * time.Second)
	if c.Running() {
		t.Errorf("Still running after four seconds")
	}
	if err := c.Wait(); err != nil {
		t.Errorf("Waiting: %v", err)
	}
}
