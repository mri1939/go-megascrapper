all:
	mkdir -p bin
	go build -mod vendor -o bin/solution main.go

docker:
	docker build -t solution .

clean:
	rm -rf bin/

