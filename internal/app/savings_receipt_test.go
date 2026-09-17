package app

import "testing"

func subscriptionDecision() routeDecision {
	return routeDecision{ToolMode: "codex", ToolLabel: "Codex CLI handoff"}
}

func TestRouteClassForDecision(t *testing.T) {
	if got := routeClassForDecision(subscriptionDecision(), providerConfig{}); got != savingsRouteSubscription {
		t.Fatalf("cli handoff should be subscription, got %q", got)
	}
	if got := routeClassForDecision(routeDecision{}, providerConfig{ID: "local-preview"}); got != savingsRouteLocal {
		t.Fatalf("local-preview should be local, got %q", got)
	}
	if got := routeClassForDecision(routeDecision{}, providerConfig{ID: "openai"}); got != savingsRouteMetered {
		t.Fatalf("remote provider should be metered, got %q", got)
	}
}

func TestComputeTaskSavingsSubscriptionEarnsImputed(t *testing.T) {
	entry := computeTaskSavings(subscriptionDecision(), providerConfig{}, 4000, 4000, providerGenerationRequest{Title: "fix parser"})
	if entry == nil {
		t.Fatal("subscription route must earn an imputed saving")
	}
	if !entry.Imputed {
		t.Fatal("MVP entries are always imputed")
	}
	if entry.USDSaved <= 0 {
		t.Fatalf("expected positive saving, got %v", entry.USDSaved)
	}
	if entry.RouteClass != savingsRouteSubscription || entry.Title != "fix parser" {
		t.Fatalf("bad entry: %+v", entry)
	}
}

func TestComputeTaskSavingsMeteredEarnsNothing(t *testing.T) {
	entry := computeTaskSavings(routeDecision{ToolLabel: "OpenAI API"}, providerConfig{ID: "openai"}, 4000, 4000, providerGenerationRequest{})
	if entry != nil {
		t.Fatalf("metered route the user paid for must earn no saving, got %+v", entry)
	}
}

func TestRecordSavingsStandaloneWritesNothing(t *testing.T) {
	withSavingsHome(t)
	decision := recordSavingsOnDecision(subscriptionDecision(), providerConfig{}, 4000, 4000, providerGenerationRequest{Standalone: true, Source: "what is 2+2?"}, throttleSurvivalReport{})
	if decision.SavingsEntry != nil {
		t.Fatal("simple question must stay clean — no entry")
	}
	if loadSavingsLedger() != nil {
		t.Fatal("standalone question must not write to the ledger")
	}
}

func TestRecordSavingsWritesExactlyOneEntry(t *testing.T) {
	withSavingsHome(t)
	decision := recordSavingsOnDecision(subscriptionDecision(), providerConfig{}, 4000, 4000, providerGenerationRequest{Title: "work task"}, throttleSurvivalReport{})
	if decision.SavingsEntry == nil {
		t.Fatal("work task on subscription route must attach an entry")
	}
	ledger := loadSavingsLedger()
	if ledger == nil || ledger.Totals.Tasks != 1 {
		t.Fatalf("expected exactly one ledger entry, got %+v", ledger)
	}
}
