package main

// importteams -dsn-base 'root:root@(127.0.0.1:3306)' < portal/data/teams.tsv

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

var (
	dsnBase    = flag.String("dsn-base", "", "`dsn` base address (w/o database name) for isu6fportal")
	dbNameDay0 = flag.String("db-day0", "isu6fportal_day0", "`database` name for day 0")
	dbNameDay1 = flag.String("db-day1", "isu6fportal_day1", "`database` name for day 1")
)

const (
	operatorTeamID   = 9999
	operatorPassword = "FCedmJpLcF7DByju"
)

func main() {
	flag.Parse()

	db0, err := sql.Open("mysql", *dsnBase+"/"+*dbNameDay0)
	if err != nil {
		log.Fatal(err)
	}
	db1, err := sql.Open("mysql", *dsnBase+"/"+*dbNameDay1)
	if err != nil {
		log.Fatal(err)
	}

	for _, db := range []*sql.DB{db1} {
		_, err = db.Exec("SET SESSION sql_mode='TRADITIONAL,NO_AUTO_VALUE_ON_ZERO,ONLY_FULL_GROUP_BY'")
		if err != nil {
			log.Fatal(err)
		}
	}

	var count1 int

	s := bufio.NewScanner(os.Stdin)
	s.Scan() // drop first line
	for s.Scan() {
		parts := strings.Split(s.Text(), "\t")
		var (
			teamID   int64
			name     string = parts[3]
			password string = parts[6]
			err      error
		)

		teamID, err = strconv.ParseInt(parts[1], 10, 0)
		if err != nil {
			log.Fatal(err)
		}

		var db *sql.DB
		switch parts[5] {
		case "10月22日(土)":
			db = db1
			count1++
		default:
			log.Fatalf("unknown day: %q", parts[5])
		}

		var category string
		switch parts[2] {
		case "一般":
			category = "general"
		case "学生":
			category = "students"
		default:
			log.Fatalf("unknown category: %q", parts[2])
		}

		_, err = db.Exec("REPLACE INTO teams (id, name, password, category, azure_resource_group) VALUES (?, ?, ?, ?, ?)", teamID, name, password, category, parts[7])
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("inserted id=%#v name=%#v password=%#v category=%#v azure_resource_group=%#v", teamID, name, password, category, parts[7])
	}

	// day0 はダミーデータで埋める
	for n := 1; n <= 26; n++ {
		var category string
		if n%2 == 0 {
			category = "general"
		} else {
			category = "students"
		}
		_, err := db0.Exec("REPLACE INTO teams (id, name, password, category, azure_resource_group) VALUES (?, ?, ?, ?, ?)", 1000+n, fmt.Sprintf("ダミーチーム%d", n), fmt.Sprintf("dummy-pass-%d", n), category, fmt.Sprintf("dummy-isucon6f-%04d", 1000+n))
		if err != nil {
			log.Fatal(err)
		}
	}

	// 運営アカウントいれる
	for _, db := range []*sql.DB{db0, db1} {
		_, err := db.Exec("REPLACE INTO teams (id, name, password, category, azure_resource_group) VALUES (?, ?, ?, ?, ?)", operatorTeamID, "運営", operatorPassword, "general", "")
		if err != nil {
			log.Fatal(err)
		}
	}

	// check data
	for _, p := range []struct {
		day   int
		db    *sql.DB
		count int
	}{{0, db0, 26}, {1, db1, count1}} {
		var c int
		err := p.db.QueryRow("SELECT COUNT(*) FROM teams").Scan(&c)
		if err != nil {
			log.Fatal(err)
		}

		c-- // 運営アカウントの分

		if c != p.count {
			log.Fatalf("team count for day %d is incorrect!!", p.day)
		} else {
			log.Printf("#teams for day %d: %d", p.day, p.count)
		}
	}
}
