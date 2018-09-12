
all: compile

clean:
	rm -rf build

compile: clean
	sh -c 'go get -d -t && go build -v -o build/terraform-provider-shell'
