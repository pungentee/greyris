GOCOM 				:= go
GO_FLAGS 			:= -buildmode=exe
OUTPUT_DIRECTORY 	:= out
SOURCE_DIRECTORY 	:= src
SOURCES     		:= $(wildcard $(SOURCE_DIRECTORY)/*.go)
OUTPUT 				:= greyris

all : build
	@echo "Done!"

build: dirs
	@echo "Building project..."
	go build -o $(OUTPUT_DIRECTORY)/$(OUTPUT) $(GO_FLAGS) $(SOURCES)

dirs:
	@mkdir -p $(OUTPUT_DIRECTORY)

clean:
	rm -rf out/* db/
