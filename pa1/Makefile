# Makefile for PA1
compile:

primary:
	go run server.go -ports "9000,9001" -leader "9000"

secondary:
	go run server.go -ports "9001,9000" -leader "9000"

client1:
	go run client.go -ports "9000,9001"

client2:
	go run client.go -ports "9000,9001"

.PHONY: compile primary secondary client1 client2
