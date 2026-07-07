#!/bin/bash
cd backend
go test -v ./internal/domain/sync/service/ -run TestAmazonPoller
