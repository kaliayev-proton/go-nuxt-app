prune:
	docker-compose down
	docker container prune -f
	docker volume prune -f

up:
	docker-compose up