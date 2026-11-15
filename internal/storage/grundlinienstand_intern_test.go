package storage

// **Es gibt nur die Grundlinie** — entschieden am 2026-10-12, Vermerk 0189
// (J02, #2266).
//
// Die 97 Wanderungen bis 0097 sind fort; ihr Umfang steht mit Name und
// Prüfsumme in `grundlinie/umfang.txt`. 0.7.0 hat genau diese 97 ausgeliefert
// — ein Bestand von dort steht auf der Grundlinie und öffnet sich ohne
// Änderung. Ein älterer wird **nicht angefasst** und mit Rat abgewiesen: einmal
// mit 0.7.0 öffnen, dort steigt er auf.

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io/fs"
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// grundlinienabdruck hält die Grundlinie fest. **Wer sie ändert, schreibt
// statt dessen eine Wanderung** — dieselbe Regel wie für eine angewendete
// Wanderung: ein Bestand, der über die alte entstand, sähe sonst anders aus
// als einer über die neue, und beide trügen dieselbe Buchführung.
const grundlinienabdruck = "5d9aa55b59431c39be5b3770a77fe73fcf5729d19075ae316db00c847298ff92"

func TestDieGrundlinieIstUnveraendert(t *testing.T) {
	namen, _ := fs.Glob(grundlinie, "grundlinie/*") //nolint:errcheck // ein fester Ausdruck
	slices.Sort(namen)

	summe := sha256.New()
	for _, name := range namen {
		inhalt, _ := grundlinie.ReadFile(name) //nolint:errcheck // eingebettet
		summe.Write([]byte(name + "\n"))
		summe.Write(inhalt)
	}

	if jetzt := hex.EncodeToString(summe.Sum(nil)); jetzt != grundlinienabdruck {
		t.Errorf("die Grundlinie hat sich geändert (%s) — eine Änderung am Schema ist eine neue Wanderung", jetzt)
	}
}

func TestDieGrundlinieNenntIhrenUmfang(t *testing.T) {
	namen, summen, fehler := grundlinienumfang(grundlinie)
	if fehler != nil {
		t.Fatal(fehler)
	}

	if len(namen) != 97 || namen[len(namen)-1] != grundlinieBis || len(summen) != 97 {
		t.Errorf("der Umfang nennt %d Wanderungen bis %s", len(namen), namen[len(namen)-1])
	}

	if fehler := luecken(namen); fehler != nil {
		t.Errorf("der Umfang hat eine Lücke: %v", fehler)
	}
}

func TestEinBestandAufDerGrundlinieBleibtUnveraendert(t *testing.T) {
	db := leererBestand(t)
	if fehler := Migriere(t.Context(), db, nil); fehler != nil {
		t.Fatal(fehler)
	}

	vorher := bestandsabdruck(t, db)

	if fehler := Migriere(t.Context(), db, nil); fehler != nil {
		t.Fatalf("ein Bestand auf der Grundlinie wurde abgewiesen: %v", fehler)
	}

	if nachher := bestandsabdruck(t, db); nachher != vorher {
		t.Errorf("das zweite Öffnen änderte den Bestand:\n%s", abdruckunterschied(vorher, nachher))
	}
}

// **Ein Bestand vor der Grundlinie wird abgewiesen und nicht angefasst** —
// etwa einer aus 0.6.0, dem 0097 fehlt.
func TestEinBestandVorDerGrundlinieWirdMitRatAbgewiesen(t *testing.T) {
	db := leererBestand(t)
	if fehler := Migriere(t.Context(), db, nil); fehler != nil {
		t.Fatal(fehler)
	}

	if _, fehler := db.ExecContext(t.Context(), "DELETE FROM migrationen WHERE name = ?", grundlinieBis); fehler != nil {
		t.Fatal(fehler)
	}

	vorher := bestandsabdruck(t, db)
	gebucht, _ := AngewendeteMigrationen(t.Context(), db) //nolint:errcheck // gezählt

	fehler := Migriere(t.Context(), db, nil)

	var aelter *domain.AeltereFassung
	if !errors.As(fehler, &aelter) || !errors.Is(fehler, ErrMigrationen) {
		t.Fatalf("der Fehler ist %v, erwartet eine ältere Fassung", fehler)
	}

	if !slices.Equal(aelter.Fehlend, []string{grundlinieBis}) || aelter.Grundlinie != grundlinieBis {
		t.Errorf("genannt sind %v bis %s", aelter.Fehlend, aelter.Grundlinie)
	}

	if nachher := bestandsabdruck(t, db); nachher != vorher {
		t.Errorf("der abgewiesene Bestand wurde verändert:\n%s", abdruckunterschied(vorher, nachher))
	}

	if verbucht, _ := AngewendeteMigrationen(t.Context(), db); len(verbucht) != len(gebucht) { //nolint:errcheck // gezählt
		t.Errorf("die Buchführung nennt %d Wanderungen, vorher %d", len(verbucht), len(gebucht))
	}
}

func TestEineGeaenderteWanderungDerGrundlinieWirdAbgewiesen(t *testing.T) {
	db := leererBestand(t)
	if fehler := Migriere(t.Context(), db, nil); fehler != nil {
		t.Fatal(fehler)
	}

	if _, fehler := db.ExecContext(t.Context(), "UPDATE migrationen SET pruefsumme = 'anders' WHERE name = '0050_arbeitsverlauf.sql'"); fehler != nil {
		t.Fatal(fehler)
	}

	fehler := Migriere(t.Context(), db, nil)

	var aelter *domain.AeltereFassung
	if !errors.Is(fehler, ErrMigrationen) || errors.As(fehler, &aelter) {
		t.Errorf("der Fehler ist %v, erwartet eine geänderte Wanderung", fehler)
	}
}

// **Eine künftige Wanderung läuft auf der Grundlinie** — über einem neuen
// Bestand wie über einem aus 0.7.0.
func TestEineKuenftigeWanderungLaeuftAufDerGrundlinie(t *testing.T) {
	quelle, probe := kuenftigeWanderung(t, "probe", "CREATE TABLE probe (x);")
	zahl := len(grundlinienumfangOder(t)) + len(danach(t)) + 1

	for name, vorher := range map[string]bool{"neu": false, "aus 0.7.0": true} {
		t.Run(name, func(t *testing.T) {
			db := leererBestand(t)
			if vorher {
				// **Wirklich der Stand von 0.7.0**: die Grundlinie und nichts
				// danach — [Migriere] brächte seit M01 schon `0098` mit.
				if fehler := migriereAusMitVorbereitung(t.Context(), db, nurGrundlinie,
					PruefeBestand, nil, grundlinie); fehler != nil {
					t.Fatal(fehler)
				}

				fuelleBestand(t, db)
			}

			if fehler := MigriereMit(t.Context(), db, quelle, nil); fehler != nil {
				t.Fatalf("wandern: %v", fehler)
			}

			verbucht, _ := AngewendeteMigrationen(t.Context(), db) //nolint:errcheck // gezählt
			if len(verbucht) != zahl || verbucht[zahl-1] != probe {
				t.Errorf("verbucht sind %d, zuletzt %s", len(verbucht), strings.Join(verbucht[len(verbucht)-1:], ""))
			}

			// **Und eine Zeile von vorher übersteht den Weg** — die eigentliche
			// Zusage hinter „aufwärts" (#795).
			if vorher {
				if zahl := treffer(t, db, `"Feuerball"*`); zahl != 1 {
					t.Errorf("die Karte von vorher wird %d-mal gefunden", zahl)
				}
			}
		})
	}
}

// **Ein abgewiesener Bestand bleibt unberührt — auch seine Buchführung.**
//
// Ein Bestand aus der Zeit vor der Prüfsumme führt `migrationen` ohne die
// Spalte. Trüge `buchfuehrungAnlegen` sie nach, bevor die Abweisung greift,
// hätte die Anwendung einen Bestand geändert, den sie gar nicht öffnet — und
// 0.7.0, die ihn aufsteigen lassen soll, fände ihn anders vor als geschrieben.
func TestEinBestandOhnePruefsummenspalteBleibtUnberuehrt(t *testing.T) {
	db := leererBestand(t)

	for _, anweisung := range []string{
		`CREATE TABLE migrationen (name TEXT PRIMARY KEY, angewendet TEXT NOT NULL)`,
		`INSERT INTO migrationen VALUES ('0001_grundlage.sql', '2026-08-01T00:00:00.000Z')`,
		`CREATE TABLE bestand (id INTEGER PRIMARY KEY AUTOINCREMENT, angelegt_am TEXT)`,
	} {
		if _, fehler := db.ExecContext(t.Context(), anweisung); fehler != nil {
			t.Fatal(fehler)
		}
	}

	vorher := bestandsabdruck(t, db)
	buchVorher := abfrage(t, db, "SELECT sql FROM sqlite_master WHERE name = 'migrationen'")

	// **Und gesichert wird nichts** — es ändert sich ja nichts.
	var aelter *domain.AeltereFassung
	if fehler := Migriere(t.Context(), db, func(context.Context) error {
		t.Error("ein abgewiesener Bestand wurde gesichert")

		return nil
	}); !errors.As(fehler, &aelter) {
		t.Fatalf("der Fehler ist %v, erwartet eine ältere Fassung", fehler)
	}

	if len(aelter.Fehlend) != 96 {
		t.Errorf("genannt sind %d fehlende Wanderungen, erwartet 96", len(aelter.Fehlend))
	}

	if nachher := bestandsabdruck(t, db); nachher != vorher {
		t.Errorf("der abgewiesene Bestand wurde verändert:\n%s", abdruckunterschied(vorher, nachher))
	}

	if buchNachher := abfrage(t, db, "SELECT sql FROM sqlite_master WHERE name = 'migrationen'"); !slices.EqualFunc(buchVorher, buchNachher, slices.Equal) {
		t.Errorf("die Buchführung wurde umgebaut: %v statt %v", buchNachher, buchVorher)
	}
}
