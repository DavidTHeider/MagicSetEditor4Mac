package storage_test

// **Woran der Index scheitern darf** — Werkstück A15 von Phase 5.3
// (#1514).
//
// # Warum sie eine eigene Datei haben
//
// **Weil jeder dieser Wege eine Zusage ist.** Der Index ist ein **abgeleiteter**
// Bestand — er darf jederzeit kaputtgehen, und dann soll er es sagen und nicht
// die Hälfte eines Standes zeigen. Ein Weg, den kein Lauf betreten hat, ist
// eine Zusage, die niemand geprüft hat (§10).

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Ein unvollständiges Manifest legt keine Generation an.**
func TestEinUnvollstaendigesManifestLegtNichtsAn(t *testing.T) {
	index := storage.NeuerRetrievalindex(frischeDatenbank(t))

	kaputt := einIndexmanifest()
	kaputt.Modell = ""

	if _, fehler := index.Anlegen(probekontext(t),
		kaputt); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("das kaputte Manifest kam durch: %v", fehler)
	}
}

// **In eine Generation, die es nicht gibt, wird nicht geschrieben.**
func TestInEineUnbekannteGenerationWirdNichtGeschrieben(t *testing.T) {
	index := storage.NeuerRetrievalindex(frischeDatenbank(t))

	if fehler := index.Schreiben(probekontext(t), 99,
		einIndexdokument("eines", "ein Text"),
		[]float32{1, 0, 0}); !errors.Is(fehler, domain.ErrNotFound) {
		t.Fatalf("die unbekannte Generation nahm ein Dokument an: %v", fehler)
	}
}

// **Und eine, die es nicht gibt, lässt sich nicht aktivieren.**
func TestEineUnbekannteGenerationLaesstSichNichtAktivieren(t *testing.T) {
	index := storage.NeuerRetrievalindex(frischeDatenbank(t))

	if fehler := index.Aktivieren(probekontext(t),
		99); !errors.Is(fehler, domain.ErrNotFound) {
		t.Fatalf("die unbekannte Generation wurde aktiv: %v", fehler)
	}
}

// **Ein Abbruch räumt seine Generation weg — und die aktive bleibt.**
//
// **Das ist der Abbruch aus dem Issue**: „Crash/Disk-full/Abbruch vor und
// nach Umschaltung erzeugen nie gemischte Vektoren oder verlorene
// Quelldaten." Wer verwirft, verliert den halben Neubau und behält den
// ganzen alten Stand.
func TestEinAbbruchRaeumtSeineGenerationWeg(t *testing.T) {
	index := mitIndex(t, einIndexdokument("alt", "der alte Text"))

	neu, fehler := index.Anlegen(probekontext(t), einIndexmanifest())
	if fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	if geschrieben := index.Schreiben(probekontext(t), neu,
		einIndexdokument("halb", "ein halber Stand"),
		[]float32{1, 0, 0}); geschrieben != nil {
		t.Fatalf("Schreiben: %v", geschrieben)
	}

	if verworfen := index.Verwerfen(probekontext(t), neu); verworfen != nil {
		t.Fatalf("Verwerfen: %v", verworfen)
	}

	// **Der alte Stand steht noch.**
	treffer, fehler := index.Lexikalisch(probekontext(t), "text", 10)
	if fehler != nil {
		t.Fatalf("Lexikalisch: %v", fehler)
	}

	if len(treffer) != 1 || treffer[0].Dokument.Kennung != "alt" {
		t.Fatalf("die Treffer sind %+v", treffer)
	}

	// **Und die aktive lässt sich nicht verwerfen** — sonst stünde niemand
	// mehr für die Suche bereit.
	aktiv, fehler := index.Manifest(probekontext(t))
	if fehler != nil {
		t.Fatalf("Manifest: %v", fehler)
	}

	if verworfen := index.Verwerfen(probekontext(t), 1); verworfen != nil {
		t.Fatalf("Verwerfen: %v", verworfen)
	}

	if _, fehler := index.Manifest(probekontext(t)); fehler != nil {
		t.Fatalf("die aktive Generation ist fort: %v — sie war %+v", fehler, aktiv)
	}
}

// **Ohne Profil im Kontext gibt es keine Suche.**
//
// **Weil ohne eines nicht zu sagen ist, wessen Texte gemeint sind.** Eine
// Suche, die dann alles zeigte, wäre der Fehler, den die Sichtprüfung
// verhindert.
func TestOhneProfilGibtEsKeineSuche(t *testing.T) {
	index := mitIndex(t, einIndexdokument("eines", "ein Text"))

	if _, fehler := index.Lexikalisch(t.Context(), "text", 10); fehler == nil {
		t.Error("die lexikalische Suche lief ohne Profil")
	}

	if _, fehler := index.Semantisch(t.Context(), []float32{1, 0, 0}, 10); fehler == nil {
		t.Error("die semantische Suche lief ohne Profil")
	}
}

// **Eine leere Frage findet nichts und meldet nichts.**
//
// **Das Issue nennt sie als Abnahme**: „leere/überlange Anfrage". Ein Fehler
// wäre hier falsch — wer nichts eingibt, hat nichts gesucht (§6).
func TestEineLeereFrageFindetNichts(t *testing.T) {
	index := mitIndex(t, einIndexdokument("eines", "ein Text"))

	for name, frage := range map[string]string{
		"leer":         "",
		"nur Leerraum": "   ",
	} {
		t.Run(name, func(t *testing.T) {
			treffer, fehler := index.Lexikalisch(probekontext(t), frage, 10)
			if fehler != nil {
				t.Fatalf("Lexikalisch: %v", fehler)
			}

			if len(treffer) != 0 {
				t.Fatalf("die Treffer sind %+v", treffer)
			}
		})
	}

	// **Und ein leerer Vektor ebenso.**
	treffer, fehler := index.Semantisch(probekontext(t), nil, 10)
	if fehler != nil {
		t.Fatalf("Semantisch: %v", fehler)
	}

	if len(treffer) != 0 {
		t.Fatalf("die Treffer sind %+v", treffer)
	}
}

// **Die Grenze begrenzt wirklich.**
func TestDieGrenzeBegrenzt(t *testing.T) {
	index := mitIndex(t,
		einIndexdokument("a", "Lightning eins"),
		einIndexdokument("b", "Lightning zwei"),
		einIndexdokument("c", "Lightning drei"))

	lexikalisch, fehler := index.Lexikalisch(probekontext(t), "lightning", 2)
	if fehler != nil {
		t.Fatalf("Lexikalisch: %v", fehler)
	}

	if len(lexikalisch) != 2 {
		t.Fatalf("die lexikalischen Treffer sind %+v", lexikalisch)
	}

	semantisch, fehler := index.Semantisch(probekontext(t), []float32{1, 0, 0}, 2)
	if fehler != nil {
		t.Fatalf("Semantisch: %v", fehler)
	}

	if len(semantisch) != 2 {
		t.Fatalf("die semantischen Treffer sind %+v", semantisch)
	}
}

// **Ein Nullvektor ist kein Fehler und keine Ähnlichkeit.**
//
// **Er kommt vor**: ein Text, den der Encoder nicht deuten kann, gibt ihn.
// Eine Division durch null wäre ein Absturz, eine erfundene Eins eine Lüge.
func TestEinNullvektorIstKeineAehnlichkeit(t *testing.T) {
	index := storage.NeuerRetrievalindex(frischeDatenbank(t))

	generation, fehler := index.Anlegen(probekontext(t), einIndexmanifest())
	if fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	if geschrieben := index.Schreiben(probekontext(t), generation,
		einIndexdokument("leer", "ein Text"),
		[]float32{0, 0, 0}); geschrieben != nil {
		t.Fatalf("Schreiben: %v", geschrieben)
	}

	if aktiviert := index.Aktivieren(probekontext(t), generation); aktiviert != nil {
		t.Fatalf("Aktivieren: %v", aktiviert)
	}

	treffer, fehler := index.Semantisch(probekontext(t), []float32{1, 0, 0}, 10)
	if fehler != nil {
		t.Fatalf("Semantisch: %v", fehler)
	}

	if len(treffer) != 1 {
		t.Fatalf("die Treffer sind %+v", treffer)
	}
}

// **Ein verbogener Bestand wird gemeldet und nicht gedeutet.**
//
// # Warum diese Läufe an der Datei ansetzen
//
// **Ohne die Tafel der Generationen geht keiner der vier Wege.**
//
// **Vier in einem Lauf**, weil sie alle dieselbe Tafel brauchen und jeder
// von ihnen eine eigene Meldung trägt.
func TestOhneGenerationentafelGehtNichts(t *testing.T) {
	db := frischeDatenbank(t)
	index := storage.NeuerRetrievalindex(db)

	generation, fehler := index.Anlegen(probekontext(t), einIndexmanifest())
	if fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	if _, fehler := db.ExecContext(probekontext(t),
		`DROP TABLE retrieval_generationen`); fehler != nil {
		t.Fatalf("verbiegen: %v", fehler)
	}

	if _, fehler := index.Anlegen(probekontext(t), einIndexmanifest()); fehler == nil {
		t.Error("eine Generation entstand ohne Tafel")
	}

	if fehler := index.Aktivieren(probekontext(t), generation); fehler == nil {
		t.Error("eine Generation wurde ohne Tafel aktiv")
	}

	if fehler := index.Verwerfen(probekontext(t), generation); fehler == nil {
		t.Error("eine Generation ließ sich ohne Tafel verwerfen")
	}

	if fehler := index.Schreiben(probekontext(t), generation,
		einIndexdokument("eines", "ein Text"),
		[]float32{1, 0, 0}); fehler == nil {
		t.Error("ein Dokument kam ohne Tafel durch")
	}
}

// **Und über einer geschlossenen Datenbank erst recht nicht.**
//
// **Das ist der Abbruch mitten im Betrieb**: wer den Bestand schließt,
// während ein Neubau läuft, bekommt eine Meldung und keinen halben Stand.
func TestUeberEinerGeschlossenenDatenbankGehtNichts(t *testing.T) {
	db := frischeDatenbank(t)
	index := storage.NeuerRetrievalindex(db)

	generation, fehler := index.Anlegen(probekontext(t), einIndexmanifest())
	if fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	if fehler := db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	if fehler := index.Schreiben(probekontext(t), generation,
		einIndexdokument("eines", "ein Text"),
		[]float32{1, 0, 0}); fehler == nil {
		t.Error("in die geschlossene Datenbank wurde geschrieben")
	}

	if fehler := index.Aktivieren(probekontext(t), generation); fehler == nil {
		t.Error("in der geschlossenen Datenbank wurde aktiviert")
	}
}

// **Eine Zeile, die sich nicht lesen lässt, ist ein Fehler.**
//
// **Verschiedene Richtungen werden verschieden geordnet.**
//
// **Ohne diesen Lauf betritt der Vergleich seinen Ja-Zweig nie**: alle
// bisherigen Vektoren zeigen in dieselbe Richtung, und dann entscheidet nur
// der Tie-Break.
func TestVerschiedeneRichtungenWerdenGeordnet(t *testing.T) {
	db := frischeDatenbank(t)
	index := storage.NeuerRetrievalindex(db)

	generation, fehler := index.Anlegen(probekontext(t), einIndexmanifest())
	if fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	for kennung, vektor := range map[string][]float32{
		"quer":   {0, 1, 0},
		"nahe":   {1, 0, 0},
		"schief": {1, 1, 0},
	} {
		if geschrieben := index.Schreiben(probekontext(t), generation,
			einIndexdokument(kennung, "Text "+kennung), vektor); geschrieben != nil {
			t.Fatalf("Schreiben: %v", geschrieben)
		}
	}

	if aktiviert := index.Aktivieren(probekontext(t), generation); aktiviert != nil {
		t.Fatalf("Aktivieren: %v", aktiviert)
	}

	treffer, fehler := index.Semantisch(probekontext(t), []float32{1, 0, 0}, 10)
	if fehler != nil {
		t.Fatalf("Semantisch: %v", fehler)
	}

	if len(treffer) != 3 {
		t.Fatalf("die Treffer sind %+v", treffer)
	}

	// **Am nächsten, dann schief, dann quer** — 1, 0,707 und 0.
	if treffer[0].Dokument.Kennung != "nahe" ||
		treffer[1].Dokument.Kennung != "schief" ||
		treffer[2].Dokument.Kennung != "quer" {
		t.Fatalf("die Ordnung ist %q, %q, %q",
			treffer[0].Dokument.Kennung, treffer[1].Dokument.Kennung,
			treffer[2].Dokument.Kennung)
	}
}

// **Eine Zeile mit unbekannter Sichtbarkeit kommt nicht heraus.**
//
// # Warum dieser Lauf über die Abfrage geht und nicht über die Funktion
//
