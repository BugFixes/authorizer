package validator

import (
  "database/sql"
  "fmt"
  "os"

  _ "github.com/lib/pq"
)

type AgentData struct {
	ID        string
	Key       string
	Secret    string
	CompanyID string
	Name      string
}

func LookupAgentId(key, secret string) (string, error) {
  agentId := ""

  db, err := sql.Open(
    "postgres",
    fmt.Sprintf(
      "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
      os.Getenv("DB_HOSTNAME"),
      os.Getenv("DB_PORT"),
      os.Getenv("DB_USERNAME"),
      os.Getenv("DB_PASSWORD"),
      os.Getenv("DB_DATABASE")))
  if err != nil {
    return agentId, fmt.Errorf("LoopkupAgentId db.open: %w", err)
  }
  defer func() {
    err := db.Close()
    if err != nil {
      fmt.Printf("LoopkupAgentId db.close: %v", err)
    }
  }()
  row := db.QueryRow("SELECT id FROM agent WHERE key=$1 AND secret=$2", key, secret)
  err = row.Scan(&agentId)
  if err != nil {
    switch err {
    case sql.ErrNoRows:
      return agentId, fmt.Errorf("LookupAgentId no rows")
    default:
      return agentId, fmt.Errorf("LookupAgentId db.query: %w", err)
    }
  }

  return agentId, nil
}

func AgentId(agentId string) (bool, error) {
  agentFound := false

  db, err := sql.Open(
    "postgres",
    fmt.Sprintf(
      "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
      os.Getenv("DB_HOSTNAME"),
      os.Getenv("DB_PORT"),
      os.Getenv("DB_USERNAME"),
      os.Getenv("DB_PASSWORD"),
      os.Getenv("DB_DATABASE")))
  if err != nil {
    return agentFound, fmt.Errorf("AgentId db.open: %w", err)
  }
  defer func() {
    err := db.Close()
    if err != nil {
      fmt.Printf("AgentId db.close: %v", err)
    }
  }()
  row := db.QueryRow("SELECT true FROM agent WHERE id=$1", agentId)
  err = row.Scan(&agentFound)
  if err != nil {
    switch err {
    case sql.ErrNoRows:
      return agentFound, fmt.Errorf("AgentId no rows")
    default:
      return agentFound, fmt.Errorf("AgentId db.query: %w", err)
    }
  }

  return agentFound, nil
}
