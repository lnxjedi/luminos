image: luminos-linux
	docker image build -t luminos:0.9.3 .

luminos-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags 'netgo osusergo static_build' -o luminos-linux

devrun:
	docker container run --name luminos-example -p 9000:9000 -v $(PWD)/_example:/var/www luminos:0.9.3

clean:
	docker container kill luminos-example || :
	docker container rm luminos-example || :

allclean: clean
	rm luminos-linux || :
	docker image rm luminos:0.9.3
