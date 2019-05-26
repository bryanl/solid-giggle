BASE=harbor.bryanl.dev/playground/kepsview
VERSION=$(shell git rev-parse --short HEAD)

build:
	docker build -t $(BASE):$(VERSION) .

publish:
	docker push $(BASE):$(VERSION)

current-version:
	@echo $(VERSION)

set-image:
	kubectl kustomize set image kepviewer=harbor.bryanl.dev/playground/kepviewer:$(VERSION)