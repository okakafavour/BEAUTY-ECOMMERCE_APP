package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DB *mongo.Database

// ConnectDB connects to the main application database
func ConnectDB() {
	connect("MONGODB_URI", "beauty-ecommerce")
}

// ConnectTestDB connects to the test database
func ConnectTestDB() {
	connect("MONGODB_TEST_URI", "beauty-ecommerce_test")
}

// internal reusable connection logic
func connect(envVar, dbName string) {
	mongoURI := os.Getenv(envVar)
	if mongoURI == "" {
		log.Fatalf("❌ Environment variable %s not set", envVar)
	}

	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to MongoDB
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Wait until MongoDB is reachable
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal("❌ Could not ping MongoDB:", err)
	}

	// Assign the database to the global DB variable
	DB = client.Database(dbName)
	fmt.Println("✅ Connected to MongoDB:", dbName)
}

// GetCollection returns a Mongo collection from the connected DB
func GetCollection(name string) *mongo.Collection {
	if DB == nil {
		log.Fatal("❌ Database not connected. Call ConnectDB() or ConnectTestDB() first.")
	}
	return DB.Collection(name)
}
