VERSION 0.8

IMPORT ../../earthly/go AS common-go

copy-mod:  ## Import current package's mod
    DO common-go+COPY_MOD

copy-mod-dependencies:  ## Import deps packages' mod
    FROM +copy-mod
    DO common-go+COPY_MOD_DEPENDENCIES
    SAVE IMAGE --cache-hint

copy-code:  ## Import current package's code
    FROM +copy-mod
    DO common-go+COPY_CODE

copy-code-dependencies:  ## Import deps packages' code
    FROM +copy-mod-dependencies
    COPY --dir +copy-code/code/* /code/
    DO common-go+COPY_CODE_DEPENDENCIES

lint:  ## Lints the go code with various linters
    FROM +copy-code-dependencies
    DO common-go+LINT

test:  ## Run tests!
    FROM +copy-code-dependencies
    DO common-go+RUN_WITH_DOCKER --command=TEST

lintest: ## One step for lint and test
    BUILD +lint
    BUILD +test

autogen:
    FROM common-go+base-ci
    WORKDIR /code/
    COPY --dir sqlc .
    RUN sqlc generate -f sqlc/sqlc.yml
    COPY go.mod ./
    RUN go mod tidy
    RUN mockery --dir gen --all --output gen/mocks
    SAVE ARTIFACT /code/gen/database AS LOCAL gen/database
    SAVE ARTIFACT /code/gen/mocks AS LOCAL gen/mocks

build-binary:  ## Compile the service and save the compiled artifact for next steps (and locally)
    FROM +copy-code-dependencies
    DO common-go+BUILD_BINARY

build-image: ## Builds docker image and if added --push publish to AWS
    FROM alpine
    # Copy the server binary
    COPY --if-exists build-config/* /run/config/
    COPY +build-binary/app +copy-mod/this_package.txt ./
    DO common-go+BUILD_IMAGE
