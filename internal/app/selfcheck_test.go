package app

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestRunFunctionalSelfCheck_AllPass(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runFunctionalSelfCheck(&stdout, &stderr)
	out := stdout.String()
	if code != 0 {
		t.Fatalf("self-check should pass, code=%d\n%s", code, out)
	}
	total := len(selfCheckScenarios())
	if !strings.Contains(out, fmt.Sprintf("%d/%d scenarios passed", total, total)) {
		t.Fatalf("expected all %d scenarios to pass, got:\n%s", total, out)
	}
	if strings.Contains(out, "FAIL") {
		t.Fatalf("no scenario should report FAIL:\n%s", out)
	}
}

func TestCheckSelfCheckResult_Detects(t *testing.T) {
	sc := selfCheckScenario{Name: "x", WantExit: 0, Contains: []string{"hello"}}
	if ok, _ := checkSelfCheckResult(sc, 0, "well hello there", ""); !ok {
		t.Fatal("matching output should pass")
	}
	if ok, detail := checkSelfCheckResult(sc, 1, "hello", ""); ok || !strings.Contains(detail, "exit=1") {
		t.Fatalf("wrong exit should fail with detail, got ok=%v detail=%q", ok, detail)
	}
	if ok, detail := checkSelfCheckResult(sc, 0, "goodbye", ""); ok || !strings.Contains(detail, "missing") {
		t.Fatalf("missing substring should fail, got ok=%v detail=%q", ok, detail)
	}
}

func TestCheckFunctionalCommand_RoutesThroughCheck(t *testing.T) {
	// `jini check functional` dispatches to the self-check and exits 0.
	var stdout, stderr bytes.Buffer
	if code := RunInteractive([]string{"check", "functional"}, strings.NewReader(""), &stdout, &stderr); code != 0 {
		t.Fatalf("`check functional` should exit 0, got %d\n%s%s", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "scenarios passed") {
		t.Fatalf("expected a self-check summary, got:\n%s", stdout.String())
	}
}
