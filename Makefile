NAME=blog
BUILD=go build -mod vendor -o build/blog .
# The -w and -s flags reduce binary sizes by excluding unnecessary symbols and debug info

.PHONY: build

build:
	$(BUILD)

linux:
	GOARCH=amd64 GOOS=linux $(BUILD)

macos:
	GOARCH=amd64 GOOS=darwin $(BUILD)

win64:
	GOARCH=amd64 GOOS=windows $(BUILD)