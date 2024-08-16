fmt:
	docker run -v $(shell pwd):/usr/src/plugin -v $(shell pwd)/.build:/go -w /usr/src/plugin golang:1.20 go fmt

lint:
	docker run --rm -v $(shell pwd):/app -w /app golangci/golangci-lint:v1.60.1 /bin/bash -c "go mod vendor && golangci-lint run -v --modules-download-mode vendor --timeout=10m"

codegen:
	docker run -v $(shell pwd):/usr/src/plugin -v $(shell pwd)/.build:/go -w /usr/src/plugin golang:1.20 go install github.com/hashicorp/packer-plugin-sdk/cmd/packer-sdc@v0.4.0
	docker run -v $(shell pwd):/usr/src/plugin -v $(shell pwd)/.build:/go -w /usr/src/plugin golang:1.20 go generate ./provisioner
	docker run -v $(shell pwd):/usr/src/plugin -v $(shell pwd)/.build:/go -w /usr/src/plugin golang:1.20 go generate ./uploader
	docker run -v $(shell pwd):/usr/src/plugin -v $(shell pwd)/.build:/go -w /usr/src/plugin golang:1.20 go generate ./builder

build:
	go build
	chmod +x packer-plugin-hostmgr

debug: build
	@PACKER_LOG=1 packer build -debug -var "vm_name=test" example

run: build
	packer build -var "vm_name=test" example
