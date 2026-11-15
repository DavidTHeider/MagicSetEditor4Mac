// **Die Fehlerwege des Setlöschens** —
// #1218.
//
// Paketintern und über den Stolpertreiber, aus demselben Grund wie überall
// hier: mit einer funktionierenden Datenbank sind diese Wege nicht erreichbar,
// weil sie funktioniert.
//
// **Und sie gehören geprüft, weil ein halber Stand hier ohne Rückweg
// dastünde.** Ein Set hat keinen Verlauf: was hier zur Hälfte geschieht, nimmt
// kein ⌘Z zurück.

package storage

import (
	"database/sql"
	"testing"
)

func TestDasSetloeschenMeldetJedenFehlerweg(t *testing.T) {
	for name, treiber := range map[string]string{
		"die Transaktion beginnt nicht": "",
		"der Name ist nicht zu lesen":   "stolper-setname-lesen",
		"das Löschen scheitert":         "stolper-set-loeschen",
		"das Bestätigen scheitert":      "stolper-setloesch-commit",
	} {
		t.Run(name, func(t *testing.T) {
			db := stolperbestand(t, treiber)

			// **Der Name ist der, den der Stolpertreiber zurückgibt.** Ohne
			// ihn scheiterte jeder Fall schon an der Namensprüfung, und die
			// drei Wege dahinter blieben ungeprüft — der Test sähe grün aus
			// und beträte dreimal denselben Zweig.
			if fehler := NeuerSetSpeicher(db).SetLoeschen(probekontext(t), 1, "neu"); fehler == nil {
				t.Error("der Fehler kam nicht durch")
			}
		})
	}
}

// stolperbestand gibt eine Datenbank, die an der gewünschten Stelle stolpert —
// oder eine geschlossene, wenn kein Treiber genannt ist.
func stolperbestand(t *testing.T, treiber string) *sql.DB {
	t.Helper()

	if treiber == "" {
		return geschlosseneDatenbank(t)
	}

	db, err := sql.Open(treiber, "")
	if err != nil {
		t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })

	return db
}
