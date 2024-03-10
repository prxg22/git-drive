dev-server: 
	gow run ./cmd/main/main.go -pk ./.ssh/id_rsa -o prxg22 -r drive   
dev-client:
	cd app && npm run dev
test-server:
	go test -v ./...