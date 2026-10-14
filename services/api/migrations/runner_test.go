package migrations

import (
	"strings"
	"testing"
)

func TestSplitStatementsIgnoresCommentSemicolons(t *testing.T) {
	statements := splitStatements("-- comment; do not execute\nCREATE TABLE demo (id integer);\n-- another comment;\nCREATE INDEX demo_idx ON demo(id);")
	if len(statements) != 2 {
		t.Fatalf("got %d statements: %#v", len(statements), statements)
	}
	if !strings.HasPrefix(statements[0], "CREATE TABLE ") || !strings.HasPrefix(statements[1], "CREATE INDEX ") {
		t.Fatalf("unexpected SQL statements: %#v", statements)
	}
}

func TestInitialMigrationStatements(t *testing.T) {
	data, err := scripts.ReadFile("0001_initial.sql")
	if err != nil {
		t.Fatal(err)
	}
	statements := splitStatements(string(data))
	if len(statements) < 10 {
		t.Fatalf("missing schema definitions, found %d statements", len(statements))
	}
	for _, statement := range statements {
		if !strings.HasPrefix(statement, "CREATE ") {
			t.Fatalf("non-DDL statement encountered: %.80s", statement)
		}
	}
}
