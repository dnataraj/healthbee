package postgres

import (
	"database/sql"
	"github.com/dnataraj/healthbee/pkg"
	"io/ioutil"
	"os"
	"testing"
)

func newTestDB(t *testing.T) (*sql.DB, func()) {
	db, err := pkg.OpenDB(os.Getenv("HB_TEST_DSN"))
	if err != nil {
		t.Fatal(err)
	}
	script, err := ioutil.ReadFile("./testdata/setup.sql")
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(string(script))
	if err != nil {
		t.Fatal(err)
	}
	script, err = ioutil.ReadFile("./testdata/testdata.sql")
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(string(script))
	if err != nil {
		t.Fatal(err)
	}

	return db, func() {
		defer db.Close()
		script, err := ioutil.ReadFile("./testdata/teardown.sql")
		if err != nil {
			t.Fatal(err)
		}
		_, err = db.Exec(string(script))
		if err != nil {
			t.Fatal(err)
		}
	}
}
