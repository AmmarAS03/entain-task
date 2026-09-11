#!/usr/bin/env bash

tools=(
	golang.org/x/tools/cmd/godoc@latest
	golang.org/x/tools/cmd/goimports@latest
	github.com/mikefarah/yq/v4@latest
	google.golang.org/protobuf/cmd/protoc-gen-go
)

# Code Quality
checks=(
	golang.org/x/lint/golint@latest
	github.com/kisielk/errcheck@latest
    honnef.co/go/tools/cmd/staticcheck@latest
	github.com/mgechev/revive@latest
)


bold="\033[1m"
highlight="\033[32m"
error="\033[31m"
reset="\033[0m"

update() {
	for t in "${@}"; do
		printf "${bold}Updating %s${reset}\n" "${t}"
		go install "${t}" &
	done
	wait
}

if [[ -z "$(command -v go)" ]]; then
	printf "${bold}${error}'go' binary not found.${reset}  Check installation and/or \$PATH${reset}\n"
	exit 1
fi

printf "${bold}${highlight}Building with %s${reset}\n" "$(go version)"

update "${tools[@]}"
update "${checks[@]}"
