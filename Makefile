.PHONY: help test readiness preview-docs build-docs gates-commit gates-push gates-release

help:
	@printf '%s\n' \
	'Jini developer targets:' \
	'  make test       Run the full Go regression suite.' \
	'  make gates-commit  Run the required per-commit gate set.' \
	'  make gates-push    Run the required pre-push gate set.' \
	'  make gates-release Run the required pre-release gate set.' \
	'  make readiness  Run publish-readiness in JSON mode.' \
	'  make preview-docs  Serve the public docs locally with Jekyll or Docker.' \
	'  make build-docs  Build the public docs locally with Jekyll or Docker.'

test:
	go test ./...

gates-commit:
	bash tools/run_required_gates.sh commit

gates-push:
	bash tools/run_required_gates.sh push

gates-release:
	bash tools/run_required_gates.sh release

readiness:
	go run ./cmd/jini publish-readiness --format json

preview-docs:
	bash tools/preview_docs.sh serve

build-docs:
	bash tools/preview_docs.sh build
