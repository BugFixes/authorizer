package validator_test

import (
  "database/sql"
	"fmt"
  "os"
	"testing"

  "github.com/bugfixes/authorizer/service/validator"
  "github.com/joho/godotenv"
  "github.com/stretchr/testify/assert"
  _ "github.com/lib/pq"
)

func injectAgent(data validator.AgentData) error {
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
    return fmt.Errorf("injectAgent db.open: %w", err)
  }
  defer func() {
    err := db.Close()
    if err != nil {
      fmt.Printf("injectAgent db.close: %v", err)
    }
  }()
  _, err = db.Exec(
    "INSERT INTO agent (id, key, secret, company_id, name) VALUES ($1, $2, $3, $4, $5)",
    data.ID,
    data.Key,
    data.Secret,
    data.CompanyID,
    data.Name)
  if err != nil {
    return fmt.Errorf("injectAgent db.exec: %w", err)
  }

  return nil
}

func deleteAgent(id string) error {
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
    return fmt.Errorf("deleteAgent db.open: %w", err)
  }
  defer func() {
    err := db.Close()
    if err != nil {
      fmt.Printf("deleteAgent db.close: %v", err)
    }
  }()
  _, err = db.Exec("DELETE FROM agent WHERE id = $1", id)
  if err != nil {
    return fmt.Errorf("deleteAgent db.exec: %w", err)
  }

  return nil
}

func TestAgentId(t *testing.T) {
	if os.Getenv("GITHUB_ACTOR") == "" {
		err := godotenv.Load()
		if err != nil {
			t.Errorf("godotenv err: %w", err)
		}
	}

	tests := []struct {
		name    string
		request validator.AgentData
		expect  bool
		err     error
	}{
		{
			name: "agentid valid",
			request: validator.AgentData{
				ID:        "ad4b99e1-dec8-4682-862a-6b017e7c7c74",
				Key:       "94365b00-c6df-483f-804e-363312750500",
				Secret:    "f7356946-5814-4b5e-ad45-0348a89576ef",
				CompanyID: "b9e9153a-028c-4173-a7a8-e5063334416a",
				Name:      "bugfixes test frontend",
			},
			expect: true,
		},
		{
			name: "agentid invalid",
			request: validator.AgentData{
				ID:        "ad4b99e1-dec8-4682-862a-6b017e7c7c75",
				Key:       "94365b00-c6df-483f-804e-363312750500",
				Secret:    "f7356946-5814-4b5e-ad45-0348a89576ef",
				CompanyID: "b9e9153a-028c-4173-a7a8-e5063334416a",
				Name:      "bugfixes test frontend",
			},
			err: fmt.Errorf("AgentId no rows"),
		},
	}

	injErr := injectAgent(tests[0].request)
	if injErr != nil {
		t.Errorf("injection err: %w", injErr)
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, err := validator.AgentId(test.request.ID)
			passed := assert.IsType(t, test.err, err)
			if !passed {
				t.Errorf("validator err: %w", err)
			}
			passed = assert.IsType(t, test.expect, resp)
			if !passed {
				t.Errorf("validator type test failed: %+v", test.expect)
			}
			passed = assert.Equal(t, test.expect, resp)
			if !passed {
				t.Errorf("validator equal test failed: %+v, resp: %+v", test.expect, resp)
			}
		})
	}

	delErr := deleteAgent(tests[0].request.ID)
	if delErr != nil {
		t.Errorf("delete err: %w", delErr)
	}
}

func TestLookupAgentId(t *testing.T) {
	if os.Getenv("GITHUB_ACTOR") == "" {
		err := godotenv.Load()
		if err != nil {
			t.Errorf("godotenv err: %w", err)
		}
	}

	tests := []struct {
		name    string
		request validator.AgentData
		expect  string
		err     error
	}{
		{
			name: "agentid found",
			request: validator.AgentData{
				ID:        "ad4b99e1-dec8-4682-862a-6b017e7c7c72",
				Key:       "94365b00-c6df-483f-804e-363312750500",
				Secret:    "f7356946-5814-4b5e-ad45-0348a89576ef",
				CompanyID: "b9e9153a-028c-4173-a7a8-e5063334416a",
				Name:      "bugfixes test frontend",
			},
			expect: "ad4b99e1-dec8-4682-862a-6b017e7c7c72",
		},
	}

	injErr := injectAgent(tests[0].request)
	if injErr != nil {
		t.Errorf("injection err: %w", injErr)
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, err := validator.LookupAgentId(test.request.Key, test.request.Secret)
			passed := assert.IsType(t, test.err, err)
			if !passed {
				t.Errorf("validator err: %w", err)
			}
			passed = assert.Equal(t, test.expect, resp)
			if !passed {
				t.Errorf("validator equal: %v, resp: %v", test.expect, resp)
			}
		})
	}

	delErr := deleteAgent(tests[0].request.ID)
	if delErr != nil {
		t.Errorf("delete err: %w", delErr)
	}
}

func BenchmarkLookupAgentId(b *testing.B) {
	b.ReportAllocs()

  if os.Getenv("GITHUB_ACTOR") == "" {
    err := godotenv.Load()
    if err != nil {
      b.Errorf("godotenv err: %w", err)
    }
  }

  tests := []struct {
    name    string
    request validator.AgentData
    expect  string
    err     error
  }{
    {
      name: "agentid found",
      request: validator.AgentData{
        ID:        "ad4b99e1-dec8-4682-862a-6b017e7c7c72",
        Key:       "94365b00-c6df-483f-804e-363312750500",
        Secret:    "f7356946-5814-4b5e-ad45-0348a89576ef",
        CompanyID: "b9e9153a-028c-4173-a7a8-e5063334416a",
        Name:      "bugfixes test frontend",
      },
      expect: "ad4b99e1-dec8-4682-862a-6b017e7c7c72",
    },
  }

  injErr := injectAgent(tests[0].request)
  if injErr != nil {
    b.Errorf("injection err: %w", injErr)
  }

  b.ResetTimer()

  for _, test := range tests {
    b.Run(test.name, func(t *testing.B) {
      t.StartTimer()
      resp, err := validator.LookupAgentId(test.request.Key, test.request.Secret)
      passed := assert.IsType(t, test.err, err)
      if !passed {
        t.Errorf("validator err: %w", err)
      }
      passed = assert.Equal(t, test.expect, resp)
      if !passed {
        t.Errorf("validator equal: %v, resp: %v", test.expect, resp)
      }
      t.StopTimer()
    })
  }

  delErr := deleteAgent(tests[0].request.ID)
  if delErr != nil {
    b.Errorf("delete err: %w", delErr)
  }
}

func BenchmarkAgentId(b *testing.B) {
  b.ReportAllocs()

  if os.Getenv("GITHUB_ACTOR") == "" {
    err := godotenv.Load()
    if err != nil {
      b.Errorf("godotenv err: %w", err)
    }
  }

  tests := []struct {
    name    string
    request validator.AgentData
    expect  bool
    err     error
  }{
    {
      name: "agentid valid",
      request: validator.AgentData{
        ID:        "ad4b99e1-dec8-4682-862a-6b017e7c7c74",
        Key:       "94365b00-c6df-483f-804e-363312750500",
        Secret:    "f7356946-5814-4b5e-ad45-0348a89576ef",
        CompanyID: "b9e9153a-028c-4173-a7a8-e5063334416a",
        Name:      "bugfixes test frontend",
      },
      expect: true,
    },
    {
      name: "agentid invalid",
      request: validator.AgentData{
        ID:        "ad4b99e1-dec8-4682-862a-6b017e7c7c75",
        Key:       "94365b00-c6df-483f-804e-363312750500",
        Secret:    "f7356946-5814-4b5e-ad45-0348a89576ef",
        CompanyID: "b9e9153a-028c-4173-a7a8-e5063334416a",
        Name:      "bugfixes test frontend",
      },
      err: fmt.Errorf("AgentId no rows"),
    },
  }

  injErr := injectAgent(tests[0].request)
  if injErr != nil {
    b.Errorf("injection err: %w", injErr)
  }

  b.ResetTimer()

  for _, test := range tests {
    b.Run(test.name, func(t *testing.B) {
      t.StartTimer()

      resp, err := validator.AgentId(test.request.ID)
      passed := assert.IsType(t, test.err, err)
      if !passed {
        t.Errorf("validator err: %w", err)
      }
      passed = assert.IsType(t, test.expect, resp)
      if !passed {
        t.Errorf("validator type test failed: %+v", test.expect)
      }
      passed = assert.Equal(t, test.expect, resp)
      if !passed {
        t.Errorf("validator equal test failed: %+v, resp: %+v", test.expect, resp)
      }

      t.StopTimer()
    })
  }

  delErr := deleteAgent(tests[0].request.ID)
  if delErr != nil {
    b.Errorf("delete err: %w", delErr)
  }
}
