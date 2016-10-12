package main

// importteams -u isucon -p isucon -h 127.0.0.1 < data/members.tsv

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

const (
	operatorTeamID   = 9999
	operatorPassword = "Btw5R5fskVvXOzT"
)

func main() {
	dbName := flag.String("db-day0", "isu6fportal_day0", "`database` name for day 0")
	user := flag.String("u", "root", "user name")
	pass := flag.String("p", "", "password")
	host := flag.String("h", "", "host")
	port := flag.String("port", "3306", "port number")

	flag.Parse()

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		*user,
		*pass,
		*host,
		*port,
		*dbName,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec("SET SESSION sql_mode='TRADITIONAL,NO_AUTO_VALUE_ON_ZERO,ONLY_FULL_GROUP_BY'")
	if err != nil {
		log.Fatal(err)
	}

	s := bufio.NewScanner(os.Stdin)
	s.Scan() // drop first line
	for s.Scan() {
		parts := strings.Split(s.Text(), "\t")
		var (
			teamID   int64
			name     string = parts[2]
			password string = parts[4]
			err      error
		)

		teamID, err = strconv.ParseInt(parts[0], 10, 0)
		if err != nil {
			log.Fatal(err)
		}

		var category string
		switch parts[1] {
		case "一般":
			category = "general"
		case "学生":
			category = "students"
		default:
			log.Fatalf("unknown category: %q", parts[1])
		}

		_, err = db.Exec("REPLACE INTO teams (id, name, password, category, azure_resource_group) VALUES (?, ?, ?, ?, ?)", teamID, name, password, category, parts[5])
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("inserted id=%#v name=%#v password=%#v category=%#v azure_resource_group=%#v", teamID, name, password, category, parts[5])
	}

	// 運営アカウントいれる
	_, err = db.Exec("REPLACE INTO teams (id, name, password, category, azure_resource_group) VALUES (?, ?, ?, ?, ?)", operatorTeamID, "運営", operatorPassword, "general", "")
	if err != nil {
		log.Fatal(err)
	}
}
