luminos: *.go host/*.go page/*.go go.*
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod vendor -a -tags 'netgo osusergo static_build' -o luminos

image:
	docker image build -t luminos:latest .

devrun:
	docker container run --name luminos-example -p 9000:9000 -v $(PWD)/_example:/var/www luminos:latest

clean:
	docker container kill luminos-example || :
	docker container rm luminos-example || :

allclean: clean
	rm luminos-linux || :
	docker image rm luminos:0.9.3
