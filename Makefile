GOCMD := CGO_ENABLED=0 go
BINARY := walgot
BINDIR := ./bin
VERSION := v0.2.0

GOLDFLAGS := -s -w -X main.Version=$(VERSION)

.PHONY: build
build:
	${GOCMD} build -ldflags "$(GOLDFLAGS)" -o ${BINDIR}/${BINARY}

.PHONY: buildAll
buildAll:
	GOOS=linux GOARCH=amd64 ${GOCMD} build -ldflags "$(GOLDFLAGS)" -o ${BINDIR}/${BINARY}_linux_amd64
	GOOS=linux GOARCH=arm ${GOCMD} build -ldflags "$(GOLDFLAGS)" -o ${BINDIR}/${BINARY}_linux_arm
	GOOS=linux GOARCH=arm64 ${GOCMD} build -ldflags "$(GOLDFLAGS)" -o ${BINDIR}/${BINARY}_linux_arm64
	GOOS=linux GOARCH=386 ${GOCMD} build -ldflags "$(GOLDFLAGS)" -o ${BINDIR}/${BINARY}_linux_386

.PHONY: clean
clean:
	rm -f ${BINDIR}/${BINARY}*

fmt:
	go fmt ./...

.PHONY: release
release:
	@echo "Tagging version $(VERSION)"
	@./create_release_tag.sh $(VERSION)
	GOOS=linux GOARCH=amd64 ${GOCMD} build -ldflags "$(GOLDFLAGS)" -o ${BINDIR}/${BINARY}_${VERSION}_linux_amd64
	GOOS=linux GOARCH=arm ${GOCMD} build -ldflags "$(GOLDFLAGS)" -o ${BINDIR}/${BINARY}_${VERSION}_linux_arm
	GOOS=linux GOARCH=arm64 ${GOCMD} build -ldflags "$(GOLDFLAGS)" -o ${BINDIR}/${BINARY}_${VERSION}_linux_arm64
	GOOS=linux GOARCH=386 ${GOCMD} build -ldflags "$(GOLDFLAGS)" -o ${BINDIR}/${BINARY}_${VERSION}_linux_386

.PHONY: dependencies
dependencies:
	${GOCMD} get github.com/Strubbl/wallabago/v7
	${GOCMD} get github.com/charmbracelet/bubbles
	${GOCMD} get github.com/charmbracelet/bubbletea
	${GOCMD} get github.com/charmbracelet/lipgloss
	${GOCMD} get github.com/k3a/html2text
	${GOCMD} get github.com/mitchellh/go-homedir
	${GOCMD} get github.com/muesli/reflow
