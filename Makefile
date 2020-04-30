
TAG?=latest
REGION=us-west-2

build:
	$(eval USERID=$(shell aws sts get-caller-identity  --output text --query "Account"))
	docker build . -t cohttp:$(TAG)
	docker tag cohttp:$(TAG) $(USERID).dkr.ecr.$(REGION).amazonaws.com/cohttp:$(TAG)
	docker push $(USERID).dkr.ecr.$(REGION).amazonaws.com/cohttp:$(TAG)

repo:
	aws ecr create-repository --repository-name cohttp

login:
	`aws ecr get-login --no-include-email`
