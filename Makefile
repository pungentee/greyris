GOCOM 				:= go
GO_FLAGS 			:= -buildmode=exe
OUTPUT_DIRECTORY 	:= out/
SOURCE_DIRECTORY 	:= src/
SOURCES     		:= $(wildcard $(SOURCE_DIRECTORY)*.go)
OUTPUT 				:= greyris

.PHONY : all
all : build
	@echo "Done!"

build: dirs
	@echo "Building project..."
	go build -o $(OUTPUT_DIRECTORY)$(OUTPUT) $(GO_FLAGS) $(SOURCES)

dirs:
	@mkdir -p $(OUTPUT_DIRECTORY)

.PHONY : clean
clean :
	rm -rf out/*

.PHONY : run
run :
	@$(OUTPUT_DIRECTORY)$(OUTPUT)