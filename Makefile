

.PHONY: tidy
tidy:
	@go mod tidy -v
	go fmt ./... 


.PHONY: audit
audit:
	@echo "running checks"
	go mod verify
	go vet ./...
	go list -m all
	go run honnef.co/go/tools/cmd/staticcheck@latest -checks=all,-ST1000,-U1000 ./...
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

reuse:
	pipx run reuse lint

run:
	go run ./cmd/mfdcli/



.PHONY: no-dirty
no-dirty:
	git diff --exit-code
