package main

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"os"

	qtree "github.com/Johniel/go-quadtree/src/tree"
	_ "github.com/mattn/go-sqlite3"
)

type repository struct {
	db    *sql.DB     // 点とハッシュ値を管理するDB
	tree  *qtree.Tree // ハッシュ計算用の木
	depth int32       // 木の深さの上限
}

func (r *repository) init(minPoint, maxPoint *qtree.Point, depth int32) error {
	os.Remove("./demo.db")
	db, err := sql.Open("sqlite3", "./demo.db")
	if err != nil {
		return err
	}

	createTable := `
CREATE TABLE Points (
  id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
  x REAL NOT NULL,
  y REAL NOT NULL,
  path TEXT NOT NULL
);`
	_, err = db.Exec(createTable)
	if err != nil {
		return err
	}

	createIndex := `CREATE INDEX indexHash ON Points(path);`
	_, err = db.Exec(createIndex)
	if err != nil {
		return err
	}

	r.db = db
	r.tree = qtree.NewTree(minPoint, maxPoint)
	r.depth = depth
	return nil
}

func (r *repository) finalize() error {
	return r.db.Close()
}

func (r *repository) insert(p *qtree.Point) error {
	_, h := r.tree.Hash(p, 10)
	// 座標と共に経路もINSERTする
	_, err := r.db.Exec("INSERT INTO Points (x, y, path) VALUES(?,?,?)", p.X, p.Y, h)
	return err
}

func (r *repository) search(p *qtree.Point, depth int32) ([]*qtree.Point, error) {
	_, path := r.tree.Hash(p, depth)
	// 内包する深さdepthの領域の子孫に位置する点をSELECTする
	rows, err := r.db.Query("SELECT x, y FROM Points WHERE ? <= path AND path <= ?", path, path+"~")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ps := []*qtree.Point{}
	for rows.Next() {
		var x, y float64
		err := rows.Scan(&x, &y)
		if err != nil {
			return nil, err
		}
		q := &qtree.Point{
			X: x,
			Y: y,
		}
		ps = append(ps, q)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return ps, nil
}

func (r *repository) circleSearch(center *qtree.Point, radius float64) ([]*qtree.Point, error) {
	root, _ := r.tree.Hash(center, 0)

	depth := int32(0)
	// 求めたい距離超えない最小の幅になるように深さを決める
	for ; radius < (root.Max.X-root.Min.X)/math.Pow(2.0, float64(depth)); depth++ {
	}
	depth--
	// 中心点はこの頂点の持つ領域にある。
	centerNode, _ := r.tree.Hash(center, depth)
	candidates, err := r.search(center, depth)
	if err != nil {
		return nil, err
	}
	// 近傍にはみ出る場合があるので周辺も調べる.
	// 本当は3つで済むが簡単のために対象は全ての近傍とする.
	for _, adj := range centerNode.Adjacent() {
		matched, err := r.search(adj.Mid(), depth)
		if err != nil {
			return nil, err
		}
		candidates = append(candidates, matched...)
	}
	matched := []*qtree.Point{}
	for _, c := range candidates {
		// 三平方の定理から本当に円の中にあるか調べる
		if (c.X-center.X)*(c.X-center.X)+(c.Y-center.Y)*(c.Y-center.Y) <= radius*radius {
			matched = append(matched, c)
		}
	}
	return matched, nil
}

func main() {
	dataset, err := os.Create("./dataset.tsv")
	if err != nil {
		log.Fatal(err)
		return
	}
	defer dataset.Close()

	liner, err := os.Create("./liner.tsv")
	if err != nil {
		log.Fatal(err)
		return
	}
	defer liner.Close()

	circle, err := os.Create("./circle.tsv")
	if err != nil {
		log.Fatal(err)
		return
	}
	defer circle.Close()

	// 扱いたい領域の端2点
	minPoint := &qtree.Point{
		X: 0.0,
		Y: 0.0,
	}
	maxPoint := &qtree.Point{
		X: 32.0,
		Y: 32.0,
	}

	demo := &repository{}
	err = demo.init(minPoint, maxPoint, 10)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer demo.finalize()

	// とりあえず等間隔で作ってINSERT
	for i := minPoint.X; i < maxPoint.X; i += 0.7 {
		for j := minPoint.Y; j < maxPoint.Y; j += 0.7 {
			p := &qtree.Point{
				X: i,
				Y: j,
			}
			err := demo.insert(p)
			if err != nil {
				log.Fatal(err)
				return
			}
			// プロットしたいから出力
			dataset.Write(([]byte)(fmt.Sprintf("%f\t%f\n", i, j)))
		}
	}

	// pを内包する深さ3の領域に含まれる点をSELECTする
	p := &qtree.Point{
		X: 8.1,
		Y: 8.2,
	}
	ps, err := demo.search(p, 3)
	if err != nil {
		log.Fatal(err)
		return
	}
	for _, p := range ps {
		fmt.Printf("matched: (%f,%f)\n", p.X, p.Y)
	}
	fmt.Println("")

	// pを内包する深さ5の領域と8近傍の子孫に含まれる点をSELECTする
	node, _ := demo.tree.Hash(p, 5)
	for _, a := range node.Adjacent() {
		ps, err := demo.search(a.Mid(), node.Depth)
		if err != nil {
			log.Fatal(err)
			return
		}
		for _, p := range ps {
			fmt.Printf("matched: (%f,%f)\n", p.X, p.Y)
		}
	}
	fmt.Println("")

	// 2点間近辺
	begin := &qtree.Point{
		X: 4.0,
		Y: 2.0,
	}
	end := &qtree.Point{
		X: 20.0,
		Y: 30.0,
	}
	curr, _ := demo.tree.Hash(begin, 5)
	for curr.Min.X <= end.X && curr.Min.Y <= end.Y {
		ps, err := demo.search(curr.Mid(), 5)
		if err != nil {
			log.Fatal(err)
			return
		}
		for _, q := range ps {
			// プロットしたいから出力
			liner.Write(([]byte)(fmt.Sprintf("%f\t%f\n", q.X, q.Y)))
		}
		curr = curr.Adjacent()[7]
	}

	// 円形(簡単のために9領域を取得してフィルタする)
	center := &qtree.Point{
		X: 20.0,
		Y: 20.0,
	}
	radius := 5.0
	ps, _ = demo.circleSearch(center, radius)
	for _, p := range ps {
		// プロットしたいから出力
		circle.Write(([]byte)(fmt.Sprintf("%f\t%f\n", p.X, p.Y)))
	}
}
