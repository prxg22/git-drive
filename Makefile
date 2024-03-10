dev-server: 
	gow run ./cmd/main/main.go -key ./.ssh/id_rsa -owner prxg22 -repo drive -path ./example
dev-client:
	cd app && npm run dev
test-server:
	go test -v ./...