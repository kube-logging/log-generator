
BIN := ${PWD}/bin
export PATH := ${BIN}:${PATH}

CUSTOM_FORMATS := formats/custom

LICENSEI := ${BIN}/licensei
LICENSEI_VERSION := v0.8.0

HELM_DOCS := ${BIN}/helm-docs
HELM_DOCS_VERSION = 1.11.0

GOFLAGS =

${BIN}:
	mkdir -p ${BIN}

.PHONY: license-check
license-check: ${LICENSEI} ## Run license check
	${LICENSEI} check
	${LICENSEI} header

.PHONY: license-cache
license-cache: ${LICENSEI} go.work ## Generate license cache
	${LICENSEI} cache

.PHONY: check
check:
	go fmt ./...
	go vet ./...
	cd log && $(MAKE) check

go.work:
	go work init . log ${CUSTOM_FORMATS}

.PHONY: reinit
reinit:
	rm -f go.work
	@$(MAKE) go.work

.PHONY: build
build: go.work
	CGO_ENABLED=0 go build ${GOFLAGS} -a -o ${BIN}/loggen main.go

.PHONY: test
test: go.work
	go test ${GOFLAGS} ./...
	cd log && $(MAKE) test

.PHONY: docker-run
docker-run:
	docker build . -t log-generator:local && docker run -p 11000:11000 log-generator:local

.PHONY: helm-docs
helm-docs: ${HELM_DOCS}
	${HELM_DOCS} -s file -c charts/ -t ../charts-docs/templates/overrides.gotmpl -t README.md.gotmpl

.PHONY: check-diff
check-diff: helm-docs check
	git diff --exit-code ':(exclude)./ADOPTERS.md' ':(exclude)./docs/*'

${LICENSEI}: ${LICENSEI}_${LICENSEI_VERSION} | ${BIN}
	ln -sf $(notdir $<) $@

${LICENSEI}_${LICENSEI_VERSION}: IMPORT_PATH := github.com/goph/licensei/cmd/licensei
${LICENSEI}_${LICENSEI_VERSION}: VERSION := ${LICENSEI_VERSION}
${LICENSEI}_${LICENSEI_VERSION}: | ${BIN}
	${go_install_binary}

${HELM_DOCS}: ${HELM_DOCS}-${HELM_DOCS_VERSION}
	@ln -sf ${HELM_DOCS}-${HELM_DOCS_VERSION} ${HELM_DOCS}
${HELM_DOCS}-${HELM_DOCS_VERSION}:
	@mkdir -p bin
	curl -L https://github.com/norwoodj/helm-docs/releases/download/v${HELM_DOCS_VERSION}/helm-docs_${HELM_DOCS_VERSION}_$(shell uname)_x86_64.tar.gz | tar -zOxf - helm-docs > ${HELM_DOCS}-${HELM_DOCS_VERSION} && chmod +x ${HELM_DOCS}-${HELM_DOCS_VERSION}

define go_install_binary
find ${BIN} -name '$(notdir ${IMPORT_PATH})_*' -exec rm {} +
GOBIN=${BIN} go install ${IMPORT_PATH}@${VERSION}
mv ${BIN}/$(notdir ${IMPORT_PATH}) ${BIN}/$(notdir ${IMPORT_PATH})_${VERSION}
endef
