#!/bin/bash

GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GODEP=govendor

CRAWLAPP=crawlapp

export NOW=$(shell date --rfc-3339=ns)
export PKGS=$(shell go list ./... | grep -v vendor/)

all: test build-app run-app

test:
	@echo "${NOW} TESTING..."
	@$(GOTEST) -v -cover -race ${PKGS}

build-app:
	@echo "${NOW} BUILDING..."
	@$(GOBUILD) -race -o $(CRAWLAPP) ./web

run-app:
	@echo "${NOW} RUNNING..."
	@./$(CRAWLAPP)
