package service_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **`extra_data:` behält beim Schreiben seine zweite Ebene** —
// #2298.
//
// # Warum das ein eigener Lauf ist
//
// `extra_data:` ist der einzige Kartenwert, der **zwei** Ebenen trägt: den
// Stil und darunter seine Felder. Ein Set, das ohne Rohbaum geschrieben wird
// — frisch angelegt oder an einen neuen Ort gesichert —, setzt den Wert aus
// dem Bestand zusammen. Kam dort `magic-nouveau:` und `keyword_icons:` auf
// derselben Tiefe an, stand in der Datei ein Stil ohne Felder und ein Feld
// ohne Stil, und MSE2 las die Angabe des Stils als leer.
func TestZusatzdatenBehaltenBeimSchreibenIhreTiefe(t *testing.T) {
	bestand := setOhneQuelle()
	bestand.karten = []*domain.CardRow{{ID: 1, Number: 0, Values: []domain.Value{
		{Field: "name", Text: "Nebelhüterin"},
		{Field: "extra data", Text: "magic-nouveau:\n\tkeyword_icons: flying, haste"},
	}}}

	ort := filepath.Join(t.TempDir(), "sternengarten.mse-set")

	if _, err := service.NeuerSetschreiber(bestand, &ortsschreiber{}).
		Schreibe(t.Context(), 7, ort); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	gelesen, lesefehler := os.ReadFile(filepath.Join(ort, "set")) // #nosec G304 -- aus t.TempDir()
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	const erwartet = "\textra_data:\n" +
		"\t\tmagic-nouveau:\n" +
		"\t\t\tkeyword_icons: flying, haste\n"

	if !strings.Contains(string(gelesen), erwartet) {
		t.Errorf("die Zusatzdaten fehlen in ihrer Gestalt:\n%s\nerwartet darin\n%q", gelesen, erwartet)
	}
}

// **Und eingelesen kommt sie so an, wie sie dastand** — die Hälfte, an der
// #2298
// entstand: der Wert verlor beim Einfahren seine tiefere Einrückung, und der
// Schreiber darüber hatte nichts mehr, womit er sie hätte setzen können.
func TestZusatzdatenKommenMitIhrerTiefeInDenBestand(t *testing.T) {
	ordner := filepath.Join(t.TempDir(), "zusatz.mse-set")
	if err := os.Mkdir(ordner, 0o750); err != nil {
		t.Fatal(err)
	}

	const inhalt = "mse version: 2.0.0\n" +
		"game: magic\n" +
		"stylesheet: nouveau\n" +
		"card:\n" +
		"\tname: Nebelhüterin\n" +
		"\textra_data:\n" +
		"\t\tmagic-nouveau:\n" +
		"\t\t\tkeyword_icons: flying, haste\n"

	if err := os.WriteFile(filepath.Join(ordner, "set"), []byte(inhalt), 0o600); err != nil {
		t.Fatal(err)
	}

	speicher := &speicherDouble{}
	if _, err := service.NeuerImporteur(speicher).Set(t.Context(), ordner); err != nil {
		t.Fatalf("Import: %v", err)
	}

	if len(speicher.set.Cards) != 1 {
		t.Fatalf("%d Karten, erwartet 1", len(speicher.set.Cards))
	}

	wert, da := speicher.set.Cards[0].Value("extra data")
	if !da {
		t.Fatal("extra_data fehlt")
	}

	if erwartet := "magic-nouveau:\n\tkeyword_icons: flying, haste"; wert != erwartet {
		t.Errorf("extra_data %q, erwartet %q", wert, erwartet)
	}
}

// **Ein älterer Bestand verdirbt die Datei nicht** — die Kehrseite von
// #2298.
//
// Wer ein Set vor dieser Korrektur eingelesen hat, trägt den Wert flach im
// Bestand. Der Schreiber vergleicht ihn mit dem, was die Quelldatei sagt; seit
// die Einrückung beim Lesen bleibt, unterscheiden sich beide — und ein
// Vergleich, der nur auf Gleichheit sieht, schriebe die flache Fassung über
// die richtige.
func TestEinFlacherAltwertLaesstDieZusatzdatenDerQuelleStehen(t *testing.T) {
	const quelle = "mse version: 2.0.0\n" +
		"game: probe\n" +
		"card:\n" +
		"\tname: Nebelhüterin\n" +
		"\textra_data:\n" +
		"\t\tmagic-nouveau:\n" +
		"\t\t\tkeyword_icons: flying, haste\n"

	bestand := quellenDouble{
		leserDouble: leserDouble{
			koepfe: []domain.SetHead{{ID: 7, Name: "Sternengarten", Game: "probe"}},
			karten: []*domain.CardRow{{ID: 1, Number: 0, Values: []domain.Value{
				{Field: "name", Text: "Nebelhüterin"},
				{Field: "extra data", Text: "magic-nouveau:\nkeyword_icons: flying, haste"},
			}}},
		},
		quelle: &domain.Quelle{Datei: "set", Inhalt: []byte(quelle)},
	}

	ort := filepath.Join(t.TempDir(), "wieder.mse-set")

	if _, err := service.NeuerSetschreiber(bestand, &ortsschreiber{}).
		Schreibe(t.Context(), 7, ort); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	gelesen, lesefehler := os.ReadFile(filepath.Join(ort, "set")) // #nosec G304 -- aus t.TempDir()
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if string(gelesen) != quelle {
		t.Errorf("Abweichung:\n%q\nerwartet\n%q", gelesen, quelle)
	}
}

// **Und eine absichtlich entfernte Einrückung geht in die Datei** — Befund 1
// des Zweitblicks über E/F.
//
// Die Wache darüber gilt nur für `extra_data:`: dort ist die flache Form kein
// Wert, sondern ein Lesefehler, MSE2 läse die Vorlage als leer. Bei einem
// Text ist sie ein Wert, den jemand gesetzt haben kann — und wer ihn sieht,
// sieht ihn flach.
func TestEineEntfernteEinrueckungImTextGehtInDieDatei(t *testing.T) {
	const quelle = "mse version: 2.0.0\n" +
		"game: probe\n" +
		"card:\n" +
		"\tname: Nebelhüterin\n" +
		"\tnotes:\n" +
		"\t\tA\n" +
		"\t\t\tB\n"

	bestand := quellenDouble{
		leserDouble: leserDouble{
			koepfe: []domain.SetHead{{ID: 7, Name: "Sternengarten", Game: "probe"}},
			karten: []*domain.CardRow{{ID: 1, Number: 0, Values: []domain.Value{
				{Field: "name", Text: "Nebelhüterin"},
				{Field: "notes", Text: "A\nB"},
			}}},
		},
		quelle: &domain.Quelle{Datei: "set", Inhalt: []byte(quelle)},
	}

	ort := filepath.Join(t.TempDir(), "wieder.mse-set")

	if _, err := service.NeuerSetschreiber(bestand, &ortsschreiber{}).
		Schreibe(t.Context(), 7, ort); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	gelesen, lesefehler := os.ReadFile(filepath.Join(ort, "set")) // #nosec G304 -- aus t.TempDir()
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if !strings.Contains(string(gelesen), "\tnotes:\n\t\tA\n\t\tB\n") {
		t.Errorf("die Änderung fehlt:\n%q", gelesen)
	}
}
