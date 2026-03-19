package database

import (
    "database/sql"
    "errors"
    "os"
    _ "github.com/mattn/go-sqlite3"
)

var (
    connections = map[string]*sql.DB{}
    mu          sync.Mutex
)

var ErrDomainExists = errors.New("domain already exists")

func CreateNewTarget(name string) error {
    fullFileName := "./database/" + name + "_db.sql"

    if _, err := os.Stat(fullFileName); err == nil {
        return nil, ErrDomainExists
    }

    db, err := sql.Open("sqlite3", fullFileName)
    if err != nil {
        return err
    }

    mu.Lock()
    connections[name] = db
    mu.Unlock()

    _, err = db.Exec(`CREATE TABLE IF NOT EXISTS domains (
        id           INTEGER PRIMARY KEY AUTOINCREMENT,
        domain_name  TEXT UNIQUE,
        status_code  TEXT,
        open_ports   TEXT,
        title        TEXT,
        tech_stack   TEXT,
        content_type TEXT,
        server       TEXT,
        ips          TEXT,
        cname        TEXT,
        badges       TEXT
    );`)
    if err != nil {
        return err
    }

    _, err = db.Exec(`CREATE TABLE IF NOT EXISTS juicy_hits (
        id          INTEGER PRIMARY KEY AUTOINCREMENT,
        url         TEXT UNIQUE,
        status_code TEXT,
        size        TEXT,
        severity    TEXT
    );`)
    if err != nil {
        return err
    }

    return nil
}