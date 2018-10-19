image:
	docker image build -t luminos:0.9.1 .

devrun:
	docker container run --name luminos-example -p 9000:9000 -v $(PWD)/_example:/var/www luminos:0.9.1

clean:
	docker container kill luminos-example || :
	docker container rm luminos-example || :

allclean: clean
	docker image rm luminos:0.9.1
