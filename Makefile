GOCMD := CGO_ENABLED=0 go
BINARY := walgot
BINDIR := ./bin
VERSION := 0.0.1-alpha

GOLDFLAGS := -s -w -X main.Version=$(VERSION)

BUILD_TIME := ${shell date "+%Y-%m-%dT%H:%M"}

.PHONY: build
build:
	${GOCMD} build -ldflags "$(GOLDFLAGS)" -o ${BINDIR}/${BINARY}

.PHONY: clean
clean:
	rm -f ${BINDIR}/${BINARY}

fmt:
	go fmt ./...

.PHONY: release
release:
	echo "Tagging version ${VERSION}"
	git tag -a v${VERSION} -m "New released tag: v${VERSION}"
	GOOS=linux GOARCH=amd64 ${GOCMD} build -ldflags "$(GOLDFLAGS)" -o ${BINDIR}/${BINARY}_${VERSION}_linux_amd64
	GOOS=linux GOARCH=arm ${GOCMD} build -ldflags "$(GOLDFLAGS)" -o ${BINDIR}/${BINARY}_${VERSION}_linux_arm
	GOOS=linux GOARCH=arm64 ${GOCMD} build -ldflags "$(GOLDFLAGS)" -o ${BINDIR}/${BINARY}_${VERSION}_linux_arm64
	GOOS=linux GOARCH=386 ${GOCMD} build -ldflags "$(GOLDFLAGS)" -o ${BINDIR}/${BINARY}_${VERSION}_linux_386

.PHONY: dependencies
dependencies:
	${GOCMD} get github.com/Strubbl/wallabago/v7 v7.0.4
	${GOCMD} get github.com/charmbracelet/bubbles v0.14.0
	${GOCMD} get github.com/charmbracelet/bubbletea v0.22.1
	${GOCMD} get github.com/charmbracelet/lipgloss v0.6.0
	${GOCMD} get github.com/mitchellh/go-homedir v1.1.0
	${GOCMD} get github.com/muesli/reflow v0.3.0
	${GOCMD} get github.com/k3a/html2text v1.0.8
