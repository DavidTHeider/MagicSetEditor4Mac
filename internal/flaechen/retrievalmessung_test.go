//go:build messlauf

package flaechen_test

// **Der Messlauf über den Retrievalindex** — Werkstück 5 von A15, Phase 5.3
// (#1514).
//
// # Warum er hinter einem Tag steht
//
// **Weil er eine laufende Einbettungslaufzeit braucht.** Das Issue verlangt
// es ausdrücklich: „separat gekennzeichnete Messung, keine Download-/GPU-
// Pflicht in regulärer CI". Ohne den Tag `messlauf` wird diese Datei nicht
// einmal übersetzt.
//
//	go test -tags messlauf ./internal/flaechen/ -run Retrievalmessung -v
//
// # Und warum er hier steht
//
// **Weil er durch alle Schichten geht** — `storage` baut den Index, `service`
// führt die Hälften zusammen, `ollama` bettet ein. Das ist dieselbe Lage wie
// bei den Flächenläufen daneben, nur über die Tiefe statt über die Breite.
//
// # Was er misst und was nicht
//
// **Er misst die Maschinerie an einem selbst geschriebenen Korpus** (§11).
// Was er **nicht** misst, ist die Güte an echten Kartentexten: 120 eigene
// Dokumente sind kein Oracle-Korpus, und eine Zahl von hier sagt nichts über
// die Suche in einem echten Bestand. Sie sagt, ob Volltext, Vektoren und ihre
// Zusammenführung das tun, was sie sollen — und das ist die Frage, für die es
// das Frageset gibt.

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/ollama"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// dasModell ist das Einbettungsmodell dieses Laufs.
//
// **Über die Umgebung austauschbar**, damit ein zweites Modell denselben Lauf
// fährt, ohne dass jemand hier etwas ändert.
func dasModell() string {
	if gewaehlt := os.Getenv("MSE_MESSMODELL"); gewaehlt != "" {
		return gewaehlt
	}

	return "all-minilm"
}

// probendokument ist die Gestalt im Korpus.
type probendokument struct {
	Kennung     string                   `json:"kennung"`
	Quelle      string                   `json:"quelle"`
	Spiel       string                   `json:"spiel"`
	Sprache     string                   `json:"sprache"`
	Titel       string                   `json:"titel"`
	Text        string                   `json:"text"`
	Regelnummer string                   `json:"regelnummer"`
	Quelldatum  string                   `json:"quelldatum"`
	Merkmale    domain.Retrievalmerkmale `json:"merkmale"`
}

// probenkorpus ist die ganze Datei.
type probenkorpus struct {
	Revision  string           `json:"revision"`
	Datum     string           `json:"datum"`
	Dokumente []probendokument `json:"dokumente"`
}

// probenfrage ist eine Frage samt Urteilen.
type probenfrage struct {
	Kennung  string           `json:"kennung"`
	Klasse   string           `json:"klasse"`
	Richtung string           `json:"richtung"`
	Frage    string           `json:"frage"`
	Halte    bool             `json:"halte"`
	Urteile  []service.Urteil `json:"urteile"`
}

// probenfrageset ist die ganze Datei.
type probenfrageset struct {
	Fassung string        `json:"fassung"`
	Datum   string        `json:"datum"`
	Korpus  string        `json:"korpus"`
	Fragen  []probenfrage `json:"fragen"`
}

// gelesenAus liest eine JSON-Datei aus `testdaten/retrieval`.
func gelesenAus[T any](t *testing.T, name string) T {
	t.Helper()

	roh, fehler := os.ReadFile(filepath.Join("..", "..", "testdaten", "retrieval", name)) //nolint:gosec // eine Datei aus testdaten
	if fehler != nil {
		t.Fatalf("%s lesen: %v", name, fehler)
	}

	var ziel T
	if fehler := json.Unmarshal(roh, &ziel); fehler != nil {
		t.Fatalf("%s deuten: %v", name, fehler)
	}

	return ziel
}

// messbestand öffnet einen frischen Bestand im Arbeitsspeicher.
//
// **Eine eigene Vorrichtung und kein Griff in `internal/storage`**: die
// dortige liegt in einem Testpaket und ist von hier nicht erreichbar.
func messbestand(t *testing.T) *sql.DB {
	t.Helper()

	db, fehler := storage.Oeffnen(t.Context(), storage.ImArbeitsspeicher)
	if fehler != nil {
		t.Fatalf("die Datenbank ließ sich nicht öffnen: %v", fehler)
	}

	t.Cleanup(func() {
		if fehler := db.Close(); fehler != nil {
			t.Errorf("die Datenbank ließ sich nicht schließen: %v", fehler)
		}
	})

	if fehler := storage.Migriere(t.Context(), db, nil); fehler != nil {
		t.Fatalf("die Migration schlug fehl: %v", fehler)
	}

	return db
}

// messkontext trägt das erste Profil — das, das die Wanderung 0034 anlegt.
func messkontext(t *testing.T) context.Context {
	t.Helper()

	return storage.MitProfil(t.Context(), 1)
}

// TestRetrievalmessung fährt den Lauf und schreibt den Bericht.
func TestRetrievalmessung(t *testing.T) {
	korpus := gelesenAus[probenkorpus](t, "korpus.json")
	frageset := gelesenAus[probenfrageset](t, "frageset.json")

	if frageset.Korpus != korpus.Revision {
		t.Fatalf("das Frageset gehört zu %q, der Korpus ist %q",
			frageset.Korpus, korpus.Revision)
	}

	klient := ollama.NeuerKlient(ollama.Vorgabeadresse,
		netz.NeuerKlient(netz.Fristen{
			Verbindung: 5 * time.Second, Kopf: 30 * time.Second,
			Stillstand: 5 * time.Minute,
		}))

	raum, befunde, fehler := klient.Raum(t.Context(), dasModell())
	if fehler != nil || len(befunde) > 0 {
		t.Fatalf("der Raum von %q: %v, %+v", dasModell(), fehler, befunde)
	}

	db := messbestand(t)
	index := storage.NeuerRetrievalindex(db)
	einbetter := klient.Einbetter(raum)

	bauzeit, dokumente := baueIndex(t, index, einbetter, korpus, raum)

	bericht := messe(t, index, einbetter, frageset)
	bericht.Modell = raum
	bericht.Bauzeit = bauzeit
	bericht.Dokumente = dokumente
	bericht.Korpus = korpus.Revision
	bericht.Frageset = frageset.Fassung
	bericht.Indexbytes = indexgroesse(t, db)
	bericht.Maschine = dieMaschine()

	schreibeBericht(t, bericht)
}

// baueIndex legt eine Generation an und schaltet auf sie um.
func baueIndex(
	t *testing.T, index *storage.Retrievalindex,
	einbetter *ollama.Einbetter, korpus probenkorpus, raum domain.Vektorraum,
) (bauzeit time.Duration, dokumente int) {
	t.Helper()

	begonnen := time.Now()

	manifest := domain.Indexmanifest{
		Schemafassung:  domain.RetrievalschemaFassung,
		Zerteilfassung: service.Zerteilfassung,
		Indexfassung:   domain.RetrievalindexFassung,
		Modell:         raum.Modell, Modellrevision: raum.Modellrevision,
		Artefakthash: raum.Artefakthash, Laufzeit: raum.Laufzeit,
		Abstandsmass: raum.Abstandsmass, Vertrag: raum.Vertrag,
		Korpusstaende: []domain.Korpusstand{{
			Quelle: domain.QuelleOracle, Revision: korpus.Revision,
			Datum: korpus.Datum,
		}, {
			Quelle: domain.QuelleRegelwerk, Revision: korpus.Revision,
			Datum: korpus.Datum,
		}},
		Suchkonfiguration: service.Rangfusionsname,
	}

	generation, fehler := index.Anlegen(messkontext(t), manifest)
	if fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	geschrieben := 0

	for von := 0; von < len(korpus.Dokumente); von += ollama.HoechstensJeStapel {
		bis := min(von+ollama.HoechstensJeStapel, len(korpus.Dokumente))
		teil := korpus.Dokumente[von:bis]

		texte := make([]string, 0, len(teil))
		for _, dokument := range teil {
			texte = append(texte, raum.Vertrag.FuerDokument(
				dokument.Titel+"\n"+dokument.Text))
		}

		vektoren, fehler := einbetter.Einbetten(messkontext(t), texte)
		if fehler != nil {
			t.Fatalf("Einbetten: %v", fehler)
		}

		for stelle, dokument := range teil {
			if fehler := index.Schreiben(messkontext(t), generation,
				alsDokument(dokument, korpus), vektoren[stelle]); fehler != nil {
				t.Fatalf("Schreiben von %q: %v", dokument.Kennung, fehler)
			}

			geschrieben++
		}
	}

	if fehler := index.Aktivieren(messkontext(t), generation); fehler != nil {
		t.Fatalf("Aktivieren: %v", fehler)
	}

	return time.Since(begonnen), geschrieben
}

// alsDokument macht aus einer Probenzeile ein Retrievaldokument.
func alsDokument(roh probendokument, korpus probenkorpus) domain.Retrievaldokument {
	return domain.Retrievaldokument{
		Kennung: roh.Kennung, Quelle: domain.Retrievalquelle(roh.Quelle),
		Spiel: roh.Spiel, Sprache: roh.Sprache,
		Titel: roh.Titel, Text: roh.Text,
		Regelnummer:   roh.Regelnummer,
		Quellrevision: korpus.Revision, Quelldatum: roh.Quelldatum,
		Abgerufen:    time.Now(),
		Inhaltshash:  domain.Retrievalhash(roh.Text),
		Merkmale:     roh.Merkmale,
		Bis:          len(roh.Text),
		Sichtbarkeit: domain.SichtbarOeffentlich,
	}
}
