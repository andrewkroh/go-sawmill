.PHONY: demo
demo: exe
	./pipeline

.PHONY: exe
exe: shared-lib
	gcc -o pipeline main.c ./pipeline.so

.PHONY: shared-lib
shared-lib:
	go build -buildmode=c-shared -o pipeline.so cshared/main.go

