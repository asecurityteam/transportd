TAG := $(shell git rev-parse --short HEAD)
DIR := $(shell pwd -L)
SDCLI=asecurityteam/sdcli:v1.2.3

dep:
	docker run -ti \
        --mount src="$(DIR)",target="$(DIR)",type="bind" \
        -w "$(DIR)" \
        $(SDCLI) go dep

lint:
	docker run -ti \
        --mount src="$(DIR)",target="$(DIR)",type="bind" \
        -w "$(DIR)" \
        $(SDCLI) go lint

test:
	docker run -ti \
        --mount src="$(DIR)",target="$(DIR)",type="bind" \
        -w "$(DIR)" \
        $(SDCLI) go test

integration:
	docker run -ti \
        --mount src="$(DIR)",target="$(DIR)",type="bind" \
        -w "$(DIR)" \
        $(SDCLI) go integration

coverage:
	docker run -ti \
        --mount src="$(DIR)",target="$(DIR)",type="bind" \
        -w "$(DIR)" \
        $(SDCLI) go coverage

doc: ;

build-dev: ;

build: ;

run: ;

deploy-dev: ;

deploy: ;
