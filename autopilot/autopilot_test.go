package autopilot

import (
	"context"
	"testing"
)

type fakeStrategy struct{ action Action }

func (f fakeStrategy) OnThrottle(context.Context, ThrottleEvent) (Decision, error) {
	return Decision{Action: f.action}, nil
}

func TestRegisterRoundTrip(t *testing.T) {
	t.Cleanup(func() { Register(nil) })
	if _, ok := Registered(); ok {
		t.Fatal("nothing should be registered by default")
	}
	Register(fakeStrategy{action: Hold})
	s, ok := Registered()
	if !ok || s == nil {
		t.Fatal("strategy not registered")
	}
	Register(nil)
	if _, ok := Registered(); ok {
		t.Fatal("Register(nil) must clear")
	}
}
