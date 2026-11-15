package storage_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Anlegen, ändern, löschen** —
// #1062.
func TestEinSchluesselwortLaesstSichPflegen(t *testing.T) {
	db := frischeDatenbank(t)
	speicher := storage.NeuerSetSpeicher(db)
	pflege := storage.NeueSchluesselwortablage(db)
	leser := storage.NeuerSetLeser(db)

	setID, err := speicher.SetAnlegen(probekontext(t), &domain.Set{
		Name: "Probe", Game: "probe",
		Keywords: []*domain.Keyword{{Word: "Flug", Match: "Flug"}},
	}, nil)
	if err != nil {
		t.Fatalf("SetAnlegen: %v", err)
	}

	// **Ein neues geht ans Ende** — die Reihenfolge entscheidet, welches
	// zuerst auf einen Text passt.
	neu, anlegefehler := pflege.SchluesselwortAnlegen(probekontext(t), setID, "custom")
	if anlegefehler != nil {
		t.Fatalf("Anlegen: %v", anlegefehler)
	}

	woerter := muessenWoerter(t, leser, setID)
	if len(woerter) != 2 || woerter[1].ID != neu || woerter[1].Mode != "custom" {
		t.Fatalf("nach dem Anlegen: %+v", woerter)
	}

	// **Ein Feld setzen.**
	if setzfehler := pflege.SchluesselwortSetzen(
		probekontext(t), setID, neu, "muster", "Beben 2"); setzfehler != nil {
		t.Fatalf("Setzen: %v", setzfehler)
	}

	woerter = muessenWoerter(t, leser, setID)
	if woerter[1].Match != "Beben 2" {
		t.Errorf("nach dem Setzen: %+v", woerter[1])
	}

	// **Löschen nimmt genau eines.**
	if fehler := pflege.SchluesselwortLoeschen(probekontext(t), setID, neu); fehler != nil {
		t.Fatalf("Löschen: %v", fehler)
	}

	woerter = muessenWoerter(t, leser, setID)
	if len(woerter) != 1 || woerter[0].Word != "Flug" {
		t.Errorf("nach dem Löschen: %+v", woerter)
	}
}

// **Ein Schlüsselwort aus einem anderen Set wird nicht angefasst.**
//
// Die Setkennung steht in **jeder** Bedingung, nicht nur zur Zierde: eine
// Kennung aus einem anderen Set ist ein Zugriff über die Setgrenze hinweg, und
// die Oberfläche kann sich vertun (§7, Regel 1).
func TestEinFremdesSchluesselwortWirdAbgewiesen(t *testing.T) {
	db := frischeDatenbank(t)
	speicher := storage.NeuerSetSpeicher(db)
	pflege := storage.NeueSchluesselwortablage(db)

	eines := muessenSet(t, speicher, &domain.Set{
		Name: "Eines", Game: "probe",
		Keywords: []*domain.Keyword{{Word: "Flug"}},
	})

	anderes := muessenSet(t, speicher, &domain.Set{Name: "Anderes", Game: "probe"})

	woerter := muessenWoerter(t, storage.NeuerSetLeser(db), eines)

	if fehler := pflege.SchluesselwortSetzen(
		probekontext(t), anderes, woerter[0].ID, "wort", "X"); !errors.Is(
		fehler, domain.ErrNotFound) {
		t.Errorf("das Setzen gab %v", fehler)
	}

	if fehler := pflege.SchluesselwortLoeschen(
		probekontext(t), anderes, woerter[0].ID); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("das Löschen gab %v", fehler)
	}
}

// **Ein Feldname, den es nicht gibt, wird abgewiesen.**
//
// Er kommt aus einer Bindung; wer ihn ungeprüft in ein `UPDATE` setzte, hätte
// eine Spaltenwahl aus dem Fenster.
func TestEinUnbekanntesSchluesselwortfeldWirdAbgewiesen(t *testing.T) {
	db := frischeDatenbank(t)

	setID := muessenSet(t, storage.NeuerSetSpeicher(db), &domain.Set{
		Name: "Probe", Game: "probe", Keywords: []*domain.Keyword{{Word: "Flug"}},
	})

	woerter := muessenWoerter(t, storage.NeuerSetLeser(db), setID)

	if fehler := storage.NeueSchluesselwortablage(db).SchluesselwortSetzen(
		probekontext(t), setID, woerter[0].ID, "gibtsnicht", "X"); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Errorf("gab %v", fehler)
	}
}

// **Und jeder Weg, der an einer fehlenden Tabelle scheitert.**
//
// Ein Fehlerzweig, den kein Test betritt, ist eine Behauptung darüber, was im
// Ernstfall geschieht (§3) — dieselbe Vorgehensweise wie bei den Setangaben.
func TestDieSchluesselwortwegeMeldenEineFehlendeTabelle(t *testing.T) {
	db := frischeDatenbank(t)

	setID := muessenSet(t, storage.NeuerSetSpeicher(db), &domain.Set{
		Name: "Probe", Game: "probe", Keywords: []*domain.Keyword{{Word: "Flug"}},
	})

	wortID := muessenWoerter(t, storage.NeuerSetLeser(db), setID)[0].ID

	ausfuehren(t, db, `DROP TABLE set_keywords`)

	if fehler := storage.NeueSchluesselwortablage(db).SchluesselwortSetzen(
		probekontext(t), setID, wortID, "wort", "X"); fehler == nil {
		t.Error("das Setzen kam ohne Tabelle durch")
	}

	if _, fehler := storage.NeuerSetLeser(db).Schluesselwoerter(
		probekontext(t), setID); fehler == nil {
		t.Error("das Lesen kam ohne Tabelle durch")
	}

	if _, fehler := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t), &domain.Set{
		Name: "Zweites", Game: "probe", Keywords: []*domain.Keyword{{Word: "Beben"}},
	}, nil); fehler == nil {
		t.Error("das Schreiben eines Sets kam ohne Tabelle durch")
	}
}

// muessenWoerter liest die Schlüsselwörter und bricht bei einem Fehler ab.
func muessenWoerter(t *testing.T, leser *storage.SetLeser, setID int64) []*domain.Keyword {
	t.Helper()

	woerter, err := leser.Schluesselwoerter(probekontext(t), setID)
	if err != nil {
		t.Fatalf("Schluesselwoerter: %v", err)
	}

	return woerter
}

// muessenSet legt ein Set an und bricht bei einem Fehler ab.
func muessenSet(t *testing.T, speicher *storage.SetSpeicher, set *domain.Set) int64 {
	t.Helper()

	kennung, err := speicher.SetAnlegen(probekontext(t), set, nil)
	if err != nil {
		t.Fatalf("SetAnlegen: %v", err)
	}

	return kennung
}
