# Note that for autograder reasons,
# the following variable must
# be spelled exactly PORT!
PORT ?= 9020
# Default min port=9000 if
# not provided
# Compile command:
# make compile
compile:
# Nothing to compile
# Run command:
# make primary PORT=9020
primary:
	python3 -u server.py primary $(PORT)
# Run command:
# make secondary PORT=9020
secondary:
	python3 -u server.py secondary $(PORT)
# Run command:
# make client1 PORT=9020
client1:
	python3 -u client.py 1 $(PORT)
# Run command:
# make client2 PORT=9020
client2:
	python3 -u client.py 2 $(PORT)
# Run command:
# make client3 PORT=9020
client3:
	python3 -u client.py 3 $(PORT)
