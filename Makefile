all: samples/*
	for file in $^ ; do \
		cat $${file} |  go run challenge.go ; \
	done