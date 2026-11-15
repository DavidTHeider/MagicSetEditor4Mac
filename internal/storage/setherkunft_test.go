package storage_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Eine Quelle, ein Set — als Zusage der Datenbank** —
// #1042,
// Befund 8 des Nachlaufs über Phase V.
//
// # Warum der Code allein nicht reicht
//
// Die Regel (#886) wurde geprüft, indem **vor** dem Netzabruf nachgesehen und
// **danach** in einer eigenen Transaktion geschrieben wurde. Zwischen beidem
// liegt ein Netzgang von Sekunden.
//
// Fenster, `mse-mcp` und `mse-rest` sind **drei Prozesse an derselben Datei**
// ([0045](../../docs/entscheidungen/0045-zwei-programme-an-einem-bestand.md)).
// Zwei gleichzeitige Importe derselben Quelle sahen beide „nicht vorhanden"
// und legten zwei Sets an — und danach fand die Regel zwei und wusste nicht,
// welches gemeint ist.
//
// **Eine Bedingung im Schema kennt dieses Fenster nicht.** Sie gilt für jeden
// Prozess, jede Fassung und jeden Weg, der schreibt.
func TestEineQuelleEinSet(t *testing.T) {
	db := frischeDatenbank(t)
	speicher := storage.NeuerSetSpeicher(db)

	if _, fehler := speicher.SetAnlegen(probekontext(t), &domain.Set{
		Name: "Erstes", Game: "magic", Origin: "scryfall:lea",
	}, nil); fehler != nil {
		t.Fatalf("das erste Set: %v", fehler)
	}

	_, fehler := speicher.SetAnlegen(probekontext(t), &domain.Set{
		Name: "Zweites", Game: "magic", Origin: "scryfall:lea",
	}, nil)
	if fehler == nil {
		t.Fatal("dieselbe Quelle wurde ein zweites Mal angelegt")
	}

	if !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("Fehler %v — erwartet war ein Konflikt", fehler)
	}
}

// **Zwei Sets aus derselben Datei sind kein Duplikat** — und Sets ohne Quelle
// erst recht nicht.
//
// # Die beiden Fälle, an denen eine schlichte Bedingung zerbräche
//
// **Die leere Herkunft.** `sets.herkunft` ist `NOT NULL DEFAULT ”`, und jedes
// selbst angelegte Set trägt sie. Ein `UNIQUE (herkunft)` ließe davon genau
// eines zu — die Anwendung könnte kein zweites neues Set mehr anlegen.
//
// **Ein Dateipfad.** Eine MTG-JSON-Datei kann **mehrere** Sets führen;
// `testdaten/mtgjson/alle.json` führt zwei, und beide tragen denselben Pfad.
// „Eine Quelle, ein Set" stimmt für Dateien also **nicht** — gesehen, nicht
// überlegt: ein Index über alle nichtleeren Herkünfte machte
// `TestEinGeglueckterImportNenntSeineSetsAuch` rot.
//
// Der Index ist deshalb auf die eine Gestalt beschränkt, für die der Satz ohne
// Ausnahme gilt: `WHERE herkunft LIKE 'scryfall:%'`.
func TestNichtJedeHerkunftIstEindeutig(t *testing.T) {
	db := frischeDatenbank(t)
	speicher := storage.NeuerSetSpeicher(db)

	for _, fall := range []struct{ name, herkunft string }{
		{"Erstes ohne Quelle", ""},
		{"Zweites ohne Quelle", ""},
		{"Erstes aus der Datei", "/tmp/alle.json"},
		{"Zweites aus derselben Datei", "/tmp/alle.json"},
	} {
		if _, fehler := speicher.SetAnlegen(probekontext(t), &domain.Set{
			Name: fall.name, Game: "magic", Origin: fall.herkunft,
		}, nil); fehler != nil {
			t.Fatalf("%s: %v", fall.name, fehler)
		}
	}
}
