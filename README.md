weather loc service

go api to get weather and location data from open meteo and nominatim

docker:
docker compose up --build -d
docker cp setup.sql weather_loc_service-mssql-1:/tmp/setup.sql
docker exec weather_loc_service-mssql-1 /opt/mssql-tools18/bin/sqlcmd -S localhost -U sa -P "sqlPass!223!!" -C -i /tmp/setup.sql

without docker:
go run .

api info:
GET http://localhost:8080/api/v1/health

postman exaplles:
GET http://localhost:8080/api/v1/weather/Delhi
GET http://localhost:8080/api/v1/weather?lat=28.61&lon=77.23
GET http://localhost:8080/api/v1/location/search?q=Mumbai
GET http://localhost:8080/api/v1/insights?city=Bangalore
GET http://localhost:8080/api/v1/history?limit=10
