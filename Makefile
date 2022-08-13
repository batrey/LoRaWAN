migrateup:
	 $ migrate -path ./db/migrations -database "postgres://root:Hire_me@localhost:5432/app?sslmode=disable" -verbose up
migratedown:
	 $ migrate -path ./db/migrations -database "postgres://root:Hire_me@localhost:5432/app?sslmode=disable" -verbose down
testsuit:
	$ ENV	:= $(PWD)/.env
	$ incluude $(ENV)