#!/bin/bash

# THIS PROGRAM HELPS TO SETUP A PROJECT STRUCTURE 
# AND SOME OF THE BOILERPLATE FILES/FOLDERS NEEDED IN A MICROSERVICE

# RUN IT (from project root) WITH 
# ./setup -s (if its a server) "project-folder_name" "module_name"

# Program Flags
s_flag=false

while getopts 's' flag; do
    case "${flag}" in
        s) s_flag=true ;;
        *) error "Unexpected option ${flag}" ;;
    esac
done

# Shift processed flags
shift $((OPTIND-1))

# Helpers
setup_folders(){
    mkdir bin cmd cmd/api internal internal/data
    touch cmd/api/main.go
    touch Dockerfile
}

setup_server_files(){
    touch cmd/api/{server,routes,middleware,healthcheck}.go

    for tpl_file in ../scripts/setup_project/files/*.tpl; do
        go_file="cmd/api/$(basename "${tpl_file}" .tpl)"
        cp "${tpl_file}" "${go_file}"
    done
}

setup_server_app(){
    echo -e "\nInstalling Postgres driver"
    go get github.com/lib/pq@v1.10.9

    # Create folder to hold db migrations
    mkdir migrations

    echo -e "\nInstalling HTTP Router"
    go get github.com/julienschmidt/httprouter@v1.3.0

    echo -e "\nInstalling Rate Limiter package"
    go get golang.org/x/time@v0.5.0
    go get golang.org/x/crypto@v0.22.0

    echo -e "\nInstalling FLHO utils"
    go get github.com/windevkay/flhoutils@latest

    setup_server_files

    # Tidy up stuff just in case unneeded imports sneak in
    go mod tidy
}

# Create project directory and setup GO module
mkdir "$1"
cd "$1" || exit 1 # exit if CD fails

go mod init "github.com/windevkay/flho/${2}"

# Create common folders and files
if [ "$?" -eq 0 ]; then
    setup_folders
    if [ "$s_flag" = true ]; then
        # Make things ready for program to be a server
        echo "Flag -s was provided"
        setup_server_app
    fi

    echo -e "\nAll done!"
else
    echo -e "There was a problem creating the GO module ... exiting"
    exit 1
fi

