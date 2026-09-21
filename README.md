# Order Service

REST API backend toko online sederhana: kelola produk, pelanggan, dan pesanan. Dibangun dengan Go (Gin + GORM) dan PostgreSQL.

> Author: **Syauqi Nur Hibatullah**

## 1. Instalasi Lokal

Prasyarat: **Go 1.27+** dan **PostgreSQL 12+**.

```bash
git clone <url-repo> order-service
cd order-service

go mod download                 # install dependency

# buat database kosong
# CREATE DATABASE "order-service";

cp .env.example .env            # isi DB_PASSWORD dengan password postgres kamu

go run src/main.go
```

Server jalan di `http://localhost:5000`.

## 2. Instalasi Docker

Prasyarat: **Docker** (Docker Desktop / docker + compose).

```bash
cp .env.example .env
docker compose up --build
```

Buka `http://localhost:5001` (host port `5000` & `5432` sudah dipakai aplikasi lain, jadi di-mapping ke `5001`/`5433`).

Stop: `docker compose down` · Reset data: `docker compose down -v`