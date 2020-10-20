#!/bin/bash

set -e

go build -o go-compose
mv go-compose /usr/local/bin # may needs root privileges.
