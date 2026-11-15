package storage_test

// **Was ein Bestand über seine Größe hergibt** — Werkstück O7 von Phase 5.2
// (#1431).

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// exemplarLegen trägt ein Exemplar unmittelbar ein — **ohne den Speicher**,
// damit dieser Lauf über die Zählung spricht und nicht über die Ablage.
func exemplarLegen(
	t *testing.T, db *sql.DB, profil int64, printing, zustand string,
	finish domain.Finish, menge int,
) {
	t.Helper()

	if _, fehler := db.ExecContext(context.Background(),
		`INSERT INTO sammlung (profil_id, printing, menge, zustand, sprache, finish, name, wunsch)
		 VALUES (?, ?, ?, ?, 'en', ?, 'Bolt', 0)`,
		profil, printing, menge, zustand, string(finish)); fehler != nil {
		t.Fatalf("ein Exemplar legen: %v", fehler)
	}
}

// **Stücke und nicht Zeilen.** Eine Zeile mit vier Stück ist vier Karten — und
// wer nach einer Rückkehr wissen will, was fehlt, meint die Karten.
func TestDerBestandsstandZaehltStueckeUndNichtZeilen(t *testing.T) {
	db := frischeDatenbank(t)

	exemplarLegen(t, db, 1, "scryfall:abc", "nearMint", domain.FinishNonfoil, 4)
	exemplarLegen(t, db, 1, "scryfall:abc", "nearMint", domain.FinishFoil, 1)

	stand := storage.Bestandsstand(probekontext(t), db)

	if stand.Exemplare != 5 {
		t.Errorf("es sind %d Stück, erwartet waren 5", stand.Exemplare)
	}

	if !stand.SammlungGelesen {
		t.Fatal("die Sammlung gilt als ungelesen")
	}

	if len(stand.Sammlung) != 2 {
		t.Fatalf("es stehen %d Mengen da: %+v", len(stand.Sammlung), stand.Sammlung)
	}

	// **Foil und nicht-Foil bleiben getrennt** — sonst gäbe die Sperre später
	// eine Karte frei, die es so nicht gibt.
	for _, menge := range stand.Sammlung {
		if menge.Schluessel != "scryfall:abc" {
			t.Errorf("der Schlüssel ist %q", menge.Schluessel)
		}

		if menge.Ausfuehrung() == domain.FinishFoil && menge.Menge != 1 ||
			menge.Ausfuehrung() == domain.FinishNonfoil && menge.Menge != 4 {
			t.Errorf("die Menge passt nicht: %+v", menge)
		}
	}
}

// **Ein Wunsch ist kein Besitz** (0057) — was man nicht hat, kann man nicht
// abgegeben haben.
func TestEinWunschZaehltNichtZumBesitz(t *testing.T) {
	db := frischeDatenbank(t)

	if _, fehler := db.ExecContext(probekontext(t),
		`INSERT INTO sammlung (profil_id, printing, menge, zustand, sprache, finish, wunsch)
		 VALUES (1, 'scryfall:abc', 3, 'nearMint', 'en', 'nonfoil', 1)`); fehler != nil {
		t.Fatalf("einen Wunsch legen: %v", fehler)
	}

	stand := storage.Bestandsstand(probekontext(t), db)

	if stand.Exemplare != 0 {
		t.Errorf("der Wunsch zählt als %d Stück Besitz", stand.Exemplare)
	}

	if len(stand.Sammlung) != 0 {
		t.Errorf("der Wunsch steht in der Sammlung: %+v", stand.Sammlung)
	}
}

// **Ein leerer Bestand ist gelesen und leer** — und das ist etwas anderes als
// ungelesen.
func TestEinLeererBestandGiltAlsGelesen(t *testing.T) {
	stand := storage.Bestandsstand(probekontext(t), frischeDatenbank(t))

	if !stand.SammlungGelesen {
		t.Error("ein leerer Bestand gilt als ungelesen")
	}

	if stand.Exemplare != 0 || len(stand.Sammlung) != 0 {
		t.Errorf("der leere Bestand ist nicht leer: %+v", stand)
	}
}

// **Der Zeitpunkt kommt aus `bestand`** — er ist der Datenstand, auf den eine
// Rückkehr zurückginge, und die Abnahme verlangt ihn genau.
func TestDerBestandsstandNenntSeinenZeitpunkt(t *testing.T) {
	stand := storage.Bestandsstand(probekontext(t), frischeDatenbank(t))

	if stand.AngelegtAm.IsZero() {
		t.Error("der Bestand nennt seinen Zeitpunkt nicht")
	}
}

// **Ohne Sammlungstabelle sagt er es, statt null zu melden.** Eine sehr alte
// Datei kennt sie nicht — und „keine unklaren Mengen" wäre dort die
// gefährlichste Auskunft.
func TestOhneSammlungstabelleGiltDieSammlungAlsUngelesen(t *testing.T) {
	db := frischeDatenbank(t)

	if _, fehler := db.ExecContext(probekontext(t), `DROP TABLE sammlung`); fehler != nil {
		t.Fatalf("die Sammlung wegräumen: %v", fehler)
	}

	stand := storage.Bestandsstand(probekontext(t), db)

	if stand.SammlungGelesen {
		t.Error("die fehlende Sammlung gilt als gelesen")
	}

	if stand.Exemplare != -1 {
		t.Errorf("die fehlende Sammlung zählt %d statt −1", stand.Exemplare)
	}
}

// **Und ein Bestand ohne Zeitpunkt bleibt bei der Nullzeit statt zu raten.**
//
// **Verdorben wird die Tabelle und nicht der Wert**: der `CHECK` aus `0001`
// lässt einen unlesbaren `angelegt_am` gar nicht erst hinein — und ein Lauf,
// der es trotzdem versucht, prüft die Migration statt diesen Weg.
func TestEinBestandOhneZeitpunktBleibtBeiDerNullzeit(t *testing.T) {
	db := frischeDatenbank(t)

	if _, fehler := db.ExecContext(probekontext(t), `DROP TABLE bestand`); fehler != nil {
		t.Fatalf("die Bestandstabelle wegräumen: %v", fehler)
	}

	if stand := storage.Bestandsstand(probekontext(t), db); !stand.AngelegtAm.IsZero() {
		t.Errorf("ohne Tabelle wurde %v daraus", stand.AngelegtAm)
	}
}

// **Eine Datei, die es nicht gibt, meldet sich** — der Weg über den Pfad hat
// einen Fehler, weil dort das Öffnen scheitern kann.
func TestEineDateiDieEsNichtGibtMeldetSich(t *testing.T) {
	if _, fehler := storage.BestandsstandAus(probekontext(t),
		t.TempDir()+"/gibtesnicht.db"); fehler == nil {
		t.Error("eine fehlende Datei kam durch")
	}
}

// **Und eine, die es gibt, wird gelesen, ohne sie anzufassen.**
func TestEineFremdeDateiWirdGelesenOhneSieAnzufassen(t *testing.T) {
	ort := t.TempDir() + "/fremd.db"

	db, fehler := storage.Oeffnen(probekontext(t), ort)
	if fehler != nil {
		t.Fatalf("die fremde Datei anlegen: %v", fehler)
	}

	if migrationsfehler := storage.Migriere(probekontext(t), db, nil); migrationsfehler != nil {
		t.Fatalf("die fremde Datei migrieren: %v", migrationsfehler)
	}

	exemplarLegen(t, db, 1, "scryfall:abc", "nearMint", domain.FinishNonfoil, 2)

	if schliessfehler := db.Close(); schliessfehler != nil {
		t.Fatalf("die fremde Datei schließen: %v", schliessfehler)
	}

	stand, lesefehler := storage.BestandsstandAus(probekontext(t), ort)
	if lesefehler != nil {
		t.Fatalf("die fremde Datei lesen: %v", lesefehler)
	}

	if stand.Exemplare != 2 || len(stand.Sammlung) != 1 {
		t.Errorf("der fremde Stand ist %+v", stand)
	}
}
