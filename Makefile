.PHONY: help test test-cli test-docs readiness preview-docs build-docs

help:
	@printf '%s\n' \
	'Jini developer targets:' \
	'  make test       Run the full Python regression suite.' \
	'  make test-cli   Run the main CLI conformance suite.' \
	'  make test-docs  Run the documentation and gate suites.' \
	'  make readiness  Run publish-readiness in JSON mode.' \
	'  make preview-docs  Serve the public docs locally with Jekyll or Docker.' \
	'  make build-docs  Build the public docs locally with Jekyll or Docker.'

test:
	python3 -m unittest discover -s tests -v

test-cli:
	python3 -m unittest tests/test_jini_cli.py -v

test-docs:
	python3 -m unittest discover -s tests -p 'test_*docs.py' -v

readiness:
	python3 tools/jini.py publish-readiness --format json

preview-docs:
	bash tools/preview_docs.sh serve

build-docs:
	bash tools/preview_docs.sh build
