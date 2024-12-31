run-taskmanager:
	go run github.com/luk4z7/taskmanager

run-notificationhub:
	go run github.com/luk4z7/notificationhub

run: build
	docker-compose up --pull .

build:
	docker build -t taskmanager --file Dockerfile . --build-arg APP_NAME_ARG=github.com/luk4z7/taskmanager
	docker build -t notificationhub --file Dockerfile . --build-arg APP_NAME_ARG=github.com/luk4z7/notificationhub

k8s-deploy:
	helm upgrade db ./deployment/helm/db -f ./deployment/helm/db/values.yaml --namespace=taskhub --install --create-namespace --wait --atomic
	helm upgrade broker ./deployment/helm/broker -f ./deployment/helm/broker/values.yaml --namespace=taskhub --install --create-namespace --wait --atomic
	helm upgrade notificationhub ./deployment/helm/notificationhub -f ./deployment/helm/notificationhub/values.yaml --namespace=taskhub --install --create-namespace --wait --atomic
	helm upgrade taskmanager ./deployment/helm/taskmanager -f ./deployment/helm/taskmanager/values.yaml --namespace=taskhub --install --create-namespace --wait --atomic

test:
	go test -v -race github.com/luk4z7/taskmanager/...
	go test -v -race github.com/luk4z7/notificationhub/...