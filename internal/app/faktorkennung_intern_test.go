package app

// **Die Bestandskennung wird beim Gebrauch gelesen, nicht beim Verdrahten** —
// Nachlauf zum Zweitblick über die Phasen 4.0 bis 4.13 (2026-09-07), Befund 2.
//
// # Der Fehler, den dieser Test festhält
//
// `App.starten` ruft `verdrahten(db)` **vor** `aufDenHeutigenStand`. Ein
// Bestand, der zum ersten Mal geöffnet wird, hat die Tafel `bestandskennung`
// (Wanderung 0039) zu diesem Zeitpunkt nicht.
//
// Vorher wurde die Kennung dort gelesen und der Fehler **verworfen**. Sie
// blieb damit für die ganze erste Sitzung leer, und der Schlüssel im
// Schlüsselbund hieß `.1` statt `<kennung>.1`. Wer den zweiten Faktor beim
// ersten Start einrichtete, fand ihn beim zweiten nicht wieder.
//
// **Ein interner Test**, weil `faktorgeheimnisse` und `bestandskennung`
// paketintern sind.

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

// merkbund merkt sich, unter welchem Konto geschrieben und gelöscht wurde.
type merkbund struct{ konten []string }

func (m *merkbund) Schreibe(_ context.Context, _, konto, _ string) error {
	m.konten = append(m.konten, konto)

	return nil
}

func (m *merkbund) Lies(context.Context, string, string) (string, error) {
	return "", nil
}

func (m *merkbund) Loesche(_ context.Context, _, konto string) error {
	m.konten = append(m.konten, konto)

	return nil
}

// **Das Konto trägt die Kennung, auch wenn die Naht vor der Migration
// entstand.**
func TestDasFaktorkontoTraegtDieKennungAuchVorDerMigration(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "bestand.db")

	// **Die Naht wird an der **rohen** Datenbank gebaut** — genau die Lage,
	// die `verdrahten` herstellt: geöffnet, noch nicht migriert.
	roh, fehler := sql.Open("sqlite3", pfad)
	if fehler != nil {
		t.Fatalf("öffnen: %v", fehler)
	}

	t.Cleanup(func() { _ = roh.Close() })

	bund := &merkbund{}
	geheimnisse := faktorgeheimnisse{bund: bund, kennung: bestandskennung(roh)}

	// **Jetzt erst wird migriert** — dieselbe Reihenfolge wie im Start.
	anwendung := NeueApp(pfad)
	if starterfehler := anwendung.Starten(t.Context()); starterfehler != nil {
		t.Fatalf("starten: %v", starterfehler)
	}

	t.Cleanup(func() {
		if fehler := anwendung.Beenden(); fehler != nil {
			t.Errorf("beenden: %v", fehler)
		}
	})

	if setzfehler := geheimnisse.Setze(t.Context(), 1, "geheim"); setzfehler != nil {
		t.Fatalf("setzen: %v", setzfehler)
	}

	if len(bund.konten) != 1 {
		t.Fatalf("%d Konten statt einem: %v", len(bund.konten), bund.konten)
	}

	// **Der Kontoname beginnt nicht mit einem Punkt.**
	//
	// Das ist der ganze Befund in einer Zusicherung: `.1` heißt „die Kennung
	// war leer", `<kennung>.1` heißt „sie wurde gelesen, als sie dastand".
	if strings.HasPrefix(bund.konten[0], ".") {
		t.Errorf("das Konto heißt %q — die Kennung war leer, und das "+
			"Geheimnis wäre beim nächsten Start unauffindbar", bund.konten[0])
	}

	if !strings.HasSuffix(bund.konten[0], ".1") {
		t.Errorf("das Konto heißt %q und endet nicht auf die Nutzerkennung",
			bund.konten[0])
	}
}

// **Und ein Bestand ohne die Tafel meldet es, statt still leer zu bleiben.**
//
// **Die andere Hälfte der Zusage** (§10): oben steht, dass die Kennung
// ankommt, wenn sie dasteht; hier, dass ein Fehlschlag nicht als leere
// Zeichenkette durchgeht.
func TestEineFehlendeKennungMeldetSichStattLeerZuSein(t *testing.T) {
	roh, fehler := sql.Open("sqlite3", filepath.Join(t.TempDir(), "leer.db"))
	if fehler != nil {
		t.Fatalf("öffnen: %v", fehler)
	}

	t.Cleanup(func() { _ = roh.Close() })

	bund := &merkbund{}
	geheimnisse := faktorgeheimnisse{bund: bund, kennung: bestandskennung(roh)}

	if setzfehler := geheimnisse.Setze(t.Context(), 1, "geheim"); setzfehler == nil {
		t.Fatal("das Setzen glückte ohne Kennung — unter welchem Namen?")
	}

	if len(bund.konten) != 0 {
		t.Errorf("es wurde trotzdem geschrieben: %v", bund.konten)
	}
}

// **Auch Lesen und Löschen scheitern ohne Kennung** — und nicht nur das
// Schreiben.
//
// **Drei Wege, dieselbe Zusage.** Der Test darüber prüft `Setze`; ohne diesen
// hier bliebe offen, ob `Hole` und `Nimm` denselben Weg gehen — und ein
// Löschen, das ohne Kennung **glückt**, löschte irgendetwas oder nichts.
func TestAuchHolenUndNehmenScheiternOhneKennung(t *testing.T) {
	roh, fehler := sql.Open("sqlite3", filepath.Join(t.TempDir(), "leer.db"))
	if fehler != nil {
		t.Fatalf("öffnen: %v", fehler)
	}

	t.Cleanup(func() {
		if schliessfehler := roh.Close(); schliessfehler != nil {
			t.Errorf("schließen: %v", schliessfehler)
		}
	})

	bund := &merkbund{}
	geheimnisse := faktorgeheimnisse{bund: bund, kennung: bestandskennung(roh)}

	if _, holfehler := geheimnisse.Hole(t.Context(), 1); holfehler == nil {
		t.Error("das Lesen glückte ohne Kennung — unter welchem Namen?")
	}

	if nimmfehler := geheimnisse.Nimm(t.Context(), 1); nimmfehler == nil {
		t.Error("das Löschen glückte ohne Kennung — was wurde gelöscht?")
	}

	if len(bund.konten) != 0 {
		t.Errorf("der Bund wurde trotzdem angefasst: %v", bund.konten)
	}
}

// **Eine Bildablage ohne Ort nimmt nichts weg und sagt warum.**
//
// # Der Unterschied, um den es geht
//
// Vorher gab `beiwerkNaehte` bei einem unbestimmbaren Bilderordner eine leere
// Ablage zurück, und `Bildablage.wegnehmen` antwortet darauf mit `nil` —
// **stiller Erfolg**. Ein Nutzer wurde gelöscht, sein Foto blieb liegen, und
// niemand erfuhr es.
//
// **Beide Wege werden geprüft**, Foto und Abzeichen: sie gehen durch dieselbe
// Hilfsfunktion, und ein Test über einen von beiden sagte nichts über den
// anderen.
func TestEineOrtloseAblageNimmtNichtsWegUndSagtWarum(t *testing.T) {
	t.Parallel()

	ablage := ortlos{fehler: errors.New("kein Zuhause")}

	bild := ablage.NutzerbildWegnehmen(3)
	if bild == nil {
		t.Fatal("das Bild wurde still übergangen")
	}

	if !strings.Contains(bild.Error(), "Bild") ||
		!strings.Contains(bild.Error(), "kein Zuhause") {
		t.Errorf("die Meldung lautet %q und nennt weder das Bild noch den Grund",
			bild.Error())
	}

	// **Und die übrigen drei** — Abzeichen, und seit G4 des Zweitblicks über
	// Phase 6.3 (#1666) Tisch- und Fensterbild. Jede nennt, was liegen bleibt.
	for was, fehler := range map[string]error{
		"Abzeichen":   ablage.ProfilbadgeWegnehmen(9),
		"Tischbild":   ablage.TischbildWegnehmen(3),
		"Fensterbild": ablage.FensterbildWegnehmen(3),
	} {
		if fehler == nil || !strings.Contains(fehler.Error(), was) {
			t.Errorf("%s: die Meldung lautet %v und nennt es nicht", was, fehler)
		}
	}
}
