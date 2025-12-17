package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq" // Driver PostgreSQL standar
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DatabaseInstances menampung koneksi kedua DB agar mudah di-inject
type DatabaseInstances struct {
	Postgres *sql.DB
	Mongo    *mongo.Database
}

var DB *DatabaseInstances

func InitDB() *DatabaseInstances {
	pgDB := connectPostgres()
	mongoDB := connectMongo()

	DB = &DatabaseInstances{
		Postgres: pgDB,
		Mongo:    mongoDB,
	}

	return DB
}

// --- KONEKSI POSTGRESQL (Relational Data - Native SQL) ---
func connectPostgres() *sql.DB {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	// Buka koneksi menggunakan driver 'postgres'
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("❌ Gagal membuka driver PostgreSQL:", err)
	}

	// Cek koneksi (Ping)
	if err := db.Ping(); err != nil {
		log.Fatal("❌ Gagal koneksi ke PostgreSQL:", err)
	}

	log.Println("✅ Terhubung ke PostgreSQL (Native SQL)")

	// Migration dihapus sesuai permintaan.
	// Pastikan tabel sudah dibuat secara manual atau lewat script lain sebelum menjalankan aplikasi.

	return db
}

// --- KONEKSI MONGODB (Dynamic Data) ---
func connectMongo() *mongo.Database {
	uri := os.Getenv("MONGO_URI")
	dbName := os.Getenv("MONGO_DB_NAME")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal("❌ Gagal membuat client MongoDB:", err)
	}

	// Cek koneksi dengan Ping
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("❌ Gagal ping ke MongoDB:", err)
	}

	log.Println("✅ Terhubung ke MongoDB")

	return client.Database(dbName)
}