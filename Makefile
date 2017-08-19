ROOT_DIR := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
export GOROOT := $(HOME)/go
export GOPATH := $(ROOT_DIR)
export PATH := $(GOROOT)/bin:$(PATH)

all:
	echo $$GOROOT
	echo $$GOPATH
	go build 
