build:
	docker build --build-arg GH_TOKEN=$(token)  -t registry.digitalocean.com/athenabot/modules/walmart:latest .

tidy:
	go mod tidy -compat=1.17