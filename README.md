باشه، این یک README خیلی مینیمال فقط برای run کردن پروژه:

````md
# Run Project

## 1. Start PostgreSQL (Docker)

```powershell
docker run -d `
  --name postgres `
  -e POSTGRES_USER=postgres `
  -e POSTGRES_PASSWORD=postgres `
  -e POSTGRES_DB=mydb `
  -p 5432:5432 `
  postgres:17
````

---

## 2. Run Migration

```powershell
Get-Content .\internals\repository\postgres\migration\migration.sql -Raw |
docker exec -i postgres psql -U postgres -d mydb
```

---

## 3. Set Environment Variables

```powershell
$env:DB_HOST="localhost"
$env:DB_PORT="5432"
$env:DB_USER="postgres"
$env:DB_PASSWORD="postgres"
$env:DB_NAME="mydb"
$env:DB_SSLMODE="disable"
```

---

## 4. Run API

```powershell
go run .\cmd\api
```

---

## 5. Run Sync Job

```powershell
go run .\cmd\sync
```

---

## 6. Test API

```powershell
curl.exe http://localhost:8080/users/<id>
```

```
```
