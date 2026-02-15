stag:
	systemctl stop phisio-stag.service
	rm deployment/systemctl/phisio-stag
	go build -o deployment/systemctl/phisio-stag cmd/api/api.go
	systemctl restart phisio-stag.service
