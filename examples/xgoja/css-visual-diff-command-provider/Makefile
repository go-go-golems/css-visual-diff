.DEFAULT_GOAL := smoke

WORKSPACE_ROOT := $(abspath ../../../..)
XGOJA_ROOT := $(WORKSPACE_ROOT)/go-go-goja
BIN := $(CURDIR)/dist/css-visual-diff-command-provider
XGOJA := cd $(XGOJA_ROOT) && GOWORK=off go run ./cmd/xgoja

.PHONY: smoke doctor list build run clean

smoke: doctor list build run

doctor:
	$(XGOJA) doctor -f $(CURDIR)/xgoja.yaml

list:
	$(XGOJA) list-modules -f $(CURDIR)/xgoja.yaml

build:
	$(XGOJA) build -f $(CURDIR)/xgoja.yaml --output $(BIN) --xgoja-replace $(XGOJA_ROOT) --keep-work

run:
	rm -rf $(CURDIR)/dist/artifacts
	$(BIN) css smoke compare-widget \
		--outDir $(CURDIR)/dist/artifacts \
		--question 'What changed in the generated smoke pricing card?' \
		--output json
	test -f $(CURDIR)/dist/artifacts/agent-brief.md
	test -f $(CURDIR)/dist/artifacts/evidence.json
	grep -q 'pricing card' $(CURDIR)/dist/artifacts/agent-brief.md
	grep -q 'pricing-card' $(CURDIR)/dist/artifacts/evidence.json
	grep -q 'font-size' $(CURDIR)/dist/artifacts/evidence.json

clean:
	rm -rf $(CURDIR)/dist
