package service

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/maahdima/mwp/api/dataservice/model"
)

func TestBillingExportIncludesCompleteExactRialHistory(t *testing.T) {
	billing, _, admin, member, id, _ := billingFixture(t, model.BillingPrepaid)
	const amount int64 = 9007199254740993
	if err := billing.Adjust(admin, id, "large", amount, "=SUM(1,2)"); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 105; i++ {
		if err := billing.Adjust(admin, id, fmt.Sprintf("entry-%d", i), 1, "small credit"); err != nil {
			t.Fatal(err)
		}
	}
	book, err := billing.Book(member, id)
	if err != nil {
		t.Fatal(err)
	}
	if len(book.Ledger) != 100 {
		t.Fatal("dashboard recent history bound changed")
	}
	path, err := billing.ExportHistory(context.Background(), member, id, "ledger")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(path)
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatal("export is not private")
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	rows, err := csv.NewReader(file).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 107 {
		t.Fatal("older entries omitted", len(rows))
	}
	found := false
	for _, row := range rows[1:] {
		if row[3] == "9007199254740993" {
			found = true
			if row[6] != "'=SUM(1,2)" {
				t.Fatal("spreadsheet formula injection not neutralized")
			}
		}
	}
	if !found {
		t.Fatal("integer rial precision lost")
	}
	other := member
	wrong := id + 1
	other.ResellerID = &wrong
	if _, err = billing.ExportHistory(context.Background(), other, id, "ledger"); !errors.Is(err, ErrRoleForbidden) {
		t.Fatal("cross-tenant export accepted", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err = billing.ExportHistory(ctx, member, id, "ledger"); err == nil {
		t.Fatal("cancelled export succeeded")
	}
	if csvText("  @SUM(1)") != "'  @SUM(1)" || strings.HasPrefix(csvText("ordinary reason"), "'") {
		t.Fatal("CSV text handling")
	}
}
