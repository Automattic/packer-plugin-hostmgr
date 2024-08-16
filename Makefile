GOLANG_VERSION := '1.20'
RUBY_VERSION := '3.3.4'
GOLANGCI_LINT_VERSION := '1.60.1'

fmt:
	docker run -v $(shell pwd):/usr/src/plugin -v $(shell pwd)/.build:/go -w /usr/src/plugin golang:${GOLANG_VERSION} go fmt

lint:
	docker run --rm -v $(shell pwd):/app -w /app golangci/golangci-lint:v${GOLANGCI_LINT_VERSION} /bin/bash -c "go mod vendor && golangci-lint run -v --modules-download-mode vendor --timeout=10m"

lint-ruby:
	docker run --rm -v $(shell pwd):/app -w /app ruby:${RUBY_VERSION} /bin/bash -c "bundle install && bundle exec rubocop"

codegen:
	docker run -v $(shell pwd):/usr/src/plugin -v $(shell pwd)/.build:/go -w /usr/src/plugin golang:${GOLANG_VERSION} go install github.com/hashicorp/packer-plugin-sdk/cmd/packer-sdc@v0.4.0
	docker run -v $(shell pwd):/usr/src/plugin -v $(shell pwd)/.build:/go -w /usr/src/plugin golang:${GOLANG_VERSION} go generate ./provisioner
	docker run -v $(shell pwd):/usr/src/plugin -v $(shell pwd)/.build:/go -w /usr/src/plugin golang:${GOLANG_VERSION} go generate ./uploader
	docker run -v $(shell pwd):/usr/src/plugin -v $(shell pwd)/.build:/go -w /usr/src/plugin golang:${GOLANG_VERSION} go generate ./builder

build:
	go build
	chmod +x packer-plugin-hostmgr

validate-version:
	@[ "${VERSION}" ] || (echo "Error: VERSION is not set. Use 'make package VERSION=<version>'"; exit 1)

package: validate-version
	docker run -v $(shell pwd):/app -w /app -e GOOS=darwin -e GOARCH=amd64 golang:${GOLANG_VERSION} go build -o packer-plugin-hostmgr_$(VERSION)_x5.0_darwin_amd64 -buildvcs=false
	docker run -v $(shell pwd):/app -w /app -e GOOS=darwin -e GOARCH=arm64 golang:${GOLANG_VERSION} go build -o packer-plugin-hostmgr_$(VERSION)_x5.0_darwin_arm64 -buildvcs=false

	zip "packer-plugin-hostmgr_$(VERSION)_x5.0_darwin_amd64.zip" "packer-plugin-hostmgr_$(VERSION)_x5.0_darwin_amd64"
	zip "packer-plugin-hostmgr_$(VERSION)_x5.0_darwin_arm64.zip" "packer-plugin-hostmgr_$(VERSION)_x5.0_darwin_arm64"

publish: validate-version
	docker run -v $(shell pwd):/app -w /app -e BUILDKITE_TAG=$(VERSION) -e GITHUB_TOKEN ruby:${RUBY_VERSION} /bin/bash -c "bundle install && bundle exec ruby upload-artifacts.rb"

debug: build
	@PACKER_LOG=1 packer build -debug -var "vm_name=test" example

run: build
	packer build -var "vm_name=test" example
