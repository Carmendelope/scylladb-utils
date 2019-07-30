.DEFAULT_GOAL := all

build:
	@echo "This component does not generate binaries"

build-custom:
	@echo "This component does not generate binaries"

docker-build:
	@echo "This component has no docker images"

include scripts/Makefile.golang
include scripts/Makefile.common
