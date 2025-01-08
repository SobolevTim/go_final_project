docker build -t go-final:test . 
docker run -d --mount type=bind,src=./data,dst=/root/data --env-file .env -p 8080:8080 --name web-server-task go-final:test