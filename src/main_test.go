package main

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	qtree "github.com/Johniel/go-quadtree/src/tree"
	_ "github.com/mattn/go-sqlite3"
)

func createDB() *sql.DB {
	const filename = "./bench.db"
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		db, err := sql.Open("sqlite3", filename)
		if err != nil {
			panic(err)
		}

		createTable := `
CREATE TABLE Points (
  id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
  x REAL NOT NULL,
  y REAL NOT NULL,
  hash TEXT NOT NULL
);`
		_, err = db.Exec(createTable)
		if err != nil {
			panic(err)
		}

		createIndex := `
CREATE INDEX indexHash ON Points(hash);`
		_, err = db.Exec(createIndex)
		if err != nil {
			panic(err)
		}

		createIndex = `
CREATE INDEX indexX ON Points(x);`
		_, err = db.Exec(createIndex)
		if err != nil {
			panic(err)
		}

		createIndex = `
CREATE INDEX indexY ON Points(y);`
		_, err = db.Exec(createIndex)
		if err != nil {
			panic(err)
		}
		insertTestData(db)
		return db
	} else {
		db, err := sql.Open("sqlite3", filename)
		if err != nil {
			panic(err)
		}
		return db
	}
}

func insertTestData(db *sql.DB) *qtree.Tree {
	minPoint := &qtree.Point{
		X: -10.0,
		Y: -10.0,
	}
	maxPoint := &qtree.Point{
		X: +10.0,
		Y: +10.0,
	}
	tree := qtree.NewTree(minPoint, maxPoint)

	for _, prime := range []float64{0.3, 0.5, 0.7, 0.11, 0.13, 0.17, 0.19} {
		fmt.Printf("%v\n", prime)
		for i := -10.0; i+prime < 10.0; i += prime {
			for j := -10.0; j+prime < 10.0; j += prime {
				p := &qtree.Point{
					X: i + prime,
					Y: j + prime,
				}
				_, h := tree.Hash(p, 10)
				_, err := db.Exec("INSERT INTO Points (x, y, hash) VALUES(?,?,?)", i, j, h)
				if err != nil {
					panic(err)
				}
			}
		}
	}
	return tree
}

func BenchmarkNaive(b *testing.B) {
	db := createDB()
	defer db.Close()

	b.ResetTimer()
	rows, err := db.Query("SELECT x, y FROM Points WHERE 0.0 <= x AND x <= 0.625 AND 0.0 <= y AND y <= 0.625")
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		var x, y float64
		err = rows.Scan(&x, &y)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkIntersect(b *testing.B) {
	db := createDB()
	defer db.Close()

	b.ResetTimer()
	rows, err := db.Query("SELECT x, y FROM Points WHERE 0.0 <= x AND x <= 0.625 INTERSECT SELECT x, y FROM Points WHERE 0.0 <= y AND y <= 0.625")
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		var x, y float64
		err = rows.Scan(&x, &y)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkTree(b *testing.B) {
	db := createDB()
	defer db.Close()

	minPoint := &qtree.Point{
		X: -10.0,
		Y: -10.0,
	}
	maxPoint := &qtree.Point{
		X: +10.0,
		Y: +10.0,
	}
	tree := qtree.NewTree(minPoint, maxPoint)
	p := &qtree.Point{
		X: 0.3,
		Y: 0.4,
	}
	_, h := tree.Hash(p, 5) // &{0 0} &{0.625 0.625} 30000
	b.ResetTimer()
	rows, err := db.Query("SELECT x, y FROM Points WHERE ? <= hash AND hash <= ?", h, h+"~")
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		var x, y float64
		err = rows.Scan(&x, &y)
		if err != nil {
			panic(err)
		}
	}
}
