.PHONY: help test test-cli test-docs readiness preview-docs build-docs gates-commit gates-push gates-release

help:
	@printf '%s\n' \
	'Jini developer targets:' \
	'  make test       Run the full Python regression suite.' \
	'  make test-cli   Run the main CLI conformance suite.' \
	'  make test-docs  Run the documentation and gate suites.' \
	'  make gates-commit  Run the required per-commit gate set.' \
	'  make gates-push    Run the required pre-push gate set.' \
	'  make gates-release Run the required pre-release gate set.' \
	'  make readiness  Run publish-readiness in JSON mode.' \
	'  make preview-docs  Serve the public docs locally with Jekyll or Docker.' \
	'  make build-docs  Build the public docs locally with Jekyll or Docker.'

test:
	python3 -m unittest discover -s tests -v

test-cli:
	python3 -m unittest tests/test_jini_cli.py -v

test-docs:
	python3 -m unittest discover -s tests -p 'test_*docs.py' -v

gates-commit:
	bash tools/run_required_gates.sh commit

gates-push:
	bash tools/run_required_gates.sh push

gates-release:
	bash tools/run_required_gates.sh release

readiness:
	python3 tools/jini.py publish-readiness --format json

preview-docs:
	bash tools/preview_docs.sh serve

build-docs:
	bash tools/preview_docs.sh build
