.PHONY: gen-growthbookapi

gen-growthbookapi:
	@bash scripts/gen_growthbookapi.sh

.PHONY: test-integration

test-integration:
	@bash scripts/test_integration.sh

.PHONY: test-e2e

test-e2e:
	@bash scripts/test_e2e_local.sh

.PHONY: fmt

# fmt runs gofmt on tracked Go files, excluding the generated GrowthBook API client.
fmt:
	@gofmt -w $$(git ls-files '*.go' ':!internal/growthbookapi/**')


