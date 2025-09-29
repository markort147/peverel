#!/bin/bash
go build -o ./build/peverel ./cmd/peverel/
go build -o ./build/peverel-notifier ./cmd/notifier/
sudo install -o root -g root -m 0755 ./build/peverel /usr/local/bin/peverel
sudo install -o root -g root -m 0755 ./build/peverel-notifier /usr/local/bin/peverel-notifier