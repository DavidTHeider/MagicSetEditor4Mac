package domain_test

// **Was ein indexierbares Dokument über sich weiß** — Werkstück A15 von
// Phase 5.3
// (#1514).
//
// # Was hier gemessen wird
//
// **Dass eine Fundstelle auflösbar bleibt.** Das Ziel des Issues steht in
// einem Satz: „mit auflösbarer Herkunft". Ein Treffer, der nicht sagt, aus
// welcher Quelle, welchem Stand und welchem Zeichenbereich er kommt, ist eine
// Behauptung — und genau die soll die Recherche in A16 nicht machen.
//
// **Und dass ein Abrufdatum kein Gültigkeitsdatum ist.** Das Issue sagt es
// ausdrücklich; die zwei auseinanderzuhalten ist der Unterschied zwischen
// „diese Regel gilt seit …" und „ich habe das am … geladen".

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// einDokument ist ein vollständiges Oracle-Dokument.
func einDokument() domain.Retrievaldokument {
	return domain.Retrievaldokument{
		Kennung:        "scryfall-oracle:abc:0",
		Quelle:         domain.QuelleOracle,
		Bezugsadresse:  "https://api.scryfall.com/cards/abc",
		Spiel:          "magic",
		Kennungssystem: domain.SystemScryfall,
		Fachkennung:    "abc",
		Semantik:       domain.SemantikKonzept,
		Sprache:        "en",
		Titel:          "Lightning Bolt",
		Text:           "Lightning Bolt deals 3 damage to any target.",
		Quellrevision:  "2026-09-01",
		Quelldatum:     "2026-09-01",
		Abgerufen:      time.Date(2026, 9, 12, 9, 0, 0, 0, time.UTC),
		Inhaltshash:    "sha256:abc",
		Abschnitt:      0,
		Von:            0,
		Bis:            44,
		Sichtbarkeit:   domain.SichtbarOeffentlich,
	}
}

// **Ein vollständiges Dokument geht durch.**
func TestEinVollstaendigesRetrievaldokumentGehtDurch(t *testing.T) {
	t.Parallel()

	if fehler := einDokument().Pruefe(); fehler != nil {
		t.Fatalf("Pruefe: %v", fehler)
	}
}

// **Und jedes unvollständige nicht.**
//
// **Jedes Feld hier steht im Issue.** Es zählt sie in einem Satz auf, und
// jedes davon beantwortet eine Frage, die ein Treffer sonst offenlässt.
func TestEinUnvollstaendigesRetrievaldokumentWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	for name, kaputt := range map[string]func(*domain.Retrievaldokument){
		"ohne Kennung":        func(d *domain.Retrievaldokument) { d.Kennung = "" },
		"ohne Quelle":         func(d *domain.Retrievaldokument) { d.Quelle = "" },
		"erfundene Quelle":    func(d *domain.Retrievaldokument) { d.Quelle = "geraten" },
		"ohne Spiel":          func(d *domain.Retrievaldokument) { d.Spiel = "" },
		"ohne Sprache":        func(d *domain.Retrievaldokument) { d.Sprache = "" },
		"ohne Text":           func(d *domain.Retrievaldokument) { d.Text = "" },
		"nur Leerraum":        func(d *domain.Retrievaldokument) { d.Text = "   " },
		"ohne Revision":       func(d *domain.Retrievaldokument) { d.Quellrevision = "" },
		"ohne Abrufzeit":      func(d *domain.Retrievaldokument) { d.Abgerufen = time.Time{} },
		"ohne Hash":           func(d *domain.Retrievaldokument) { d.Inhaltshash = "" },
		"negativer Abschnitt": func(d *domain.Retrievaldokument) { d.Abschnitt = -1 },
		"verdrehter Bereich":  func(d *domain.Retrievaldokument) { d.Von, d.Bis = 40, 10 },
		"erfundene Sicht":     func(d *domain.Retrievaldokument) { d.Sichtbarkeit = "halb" },
		"erfundenes System":   func(d *domain.Retrievaldokument) { d.Kennungssystem = "geraten" },
		"erfundene Semantik":  func(d *domain.Retrievaldokument) { d.Semantik = "geraten" },
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			dokument := einDokument()
			kaputt(&dokument)

			if fehler := dokument.Pruefe(); !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("%q kam durch: %v", name, fehler)
			}
		})
	}
}

// **Ein eigenes Dokument gehört einem Profil — ein öffentliches keinem.**
//
// # Warum das eine Prüfung ist und keine Vereinbarung
//
// **Weil beides still schiefgehen kann.** Ein eigener Settext ohne Profil
// wäre für alle sichtbar; ein Oracle-Text mit Profil wäre für alle anderen
// unsichtbar, obwohl er keinem gehört. Das Issue verlangt: „private Texte und
// Vektoren bleiben profilgebunden".
func TestDieSichtbarkeitBindetDasProfil(t *testing.T) {
	t.Parallel()

	eigen := einDokument()
	eigen.Quelle = domain.QuelleEigeneKarte
	eigen.Sichtbarkeit = domain.SichtbarProfil

	if fehler := eigen.Pruefe(); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("ein eigenes Dokument ohne Profil kam durch: %v", fehler)
	}

	eigen.ProfilID = 7

	if fehler := eigen.Pruefe(); fehler != nil {
		t.Fatalf("Pruefe: %v", fehler)
	}

	oeffentlich := einDokument()
	oeffentlich.ProfilID = 7

	if fehler := oeffentlich.Pruefe(); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("ein öffentliches Dokument mit Profil kam durch: %v", fehler)
	}
}

// **Eine Regelstelle behält ihre Nummer und ihr Gültigkeitsdatum.**
//
// **Das Abrufdatum ist nicht das Gültigkeitsdatum** — das Issue sagt es in
// diesen Worten. Wer sie gleichsetzte, hätte einen Regelstand, der sich beim
// Nachladen ändert, ohne dass die Regel sich geändert hat.
func TestEineRegelstelleBehaeltNummerUndGueltigkeit(t *testing.T) {
	t.Parallel()

	regel := einDokument()
	regel.Quelle = domain.QuelleRegelwerk
	regel.Kennungssystem = ""
	regel.Fachkennung = ""
	regel.Semantik = ""
	regel.Regelnummer = "205.4a"
	regel.Quelldatum = "2026-07-26"
	regel.Abgerufen = time.Date(2026, 9, 12, 9, 0, 0, 0, time.UTC)

	if fehler := regel.Pruefe(); fehler != nil {
		t.Fatalf("Pruefe: %v", fehler)
	}

	// **Ohne Nummer ist es keine Regelstelle.**
	ohne := regel
	ohne.Regelnummer = ""

	if fehler := ohne.Pruefe(); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("die Regelstelle ohne Nummer kam durch: %v", fehler)
	}

	// **Und ohne Gültigkeitsdatum auch nicht.** Ein Regelwerk ohne Stand ist
	// ein Regelwerk, von dem niemand sagen kann, ob es gilt.
	ohne = regel
	ohne.Quelldatum = ""

	if fehler := ohne.Pruefe(); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("die Regelstelle ohne Gültigkeitsdatum kam durch: %v", fehler)
	}
}

// **Ein Ruling behält seinen Urheber.**
//
// **`wotc` und `scryfall` sind nicht dasselbe** — das unterscheidet der
// vorhandene Adapter schon (`internal/scryfall/rulings.go`), und das Issue
// verlangt es hier: „Kein pauschales ‚offiziell'".
func TestEinRulingBehaeltSeinenUrheber(t *testing.T) {
	t.Parallel()

	for name, urheber := range map[string]domain.Rulingurheber{
		"von Wizards":  domain.UrheberWotc,
		"von Scryfall": domain.UrheberScryfall,
		"unbekannt":    domain.UrheberUnbekannt,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ruling := einDokument()
			ruling.Quelle = domain.QuelleRuling
			ruling.Urheber = urheber
			ruling.Quelldatum = "2026-07-26"

			if fehler := ruling.Pruefe(); fehler != nil {
				t.Fatalf("Pruefe: %v", fehler)
			}
		})
	}

	// **Ohne Urheber ist es kein Ruling** — „offiziell" wäre geraten.
	ohne := einDokument()
	ohne.Quelle = domain.QuelleRuling
	ohne.Quelldatum = "2026-07-26"

	if fehler := ohne.Pruefe(); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("das Ruling ohne Urheber kam durch: %v", fehler)
	}
}

// **Die Quellen stehen als Aufzählung da.**
func TestDieRetrievalquellenStehenAlsAufzaehlungDa(t *testing.T) {
	t.Parallel()

	quellen := domain.Retrievalquellen()
	if len(quellen) != 5 {
		t.Fatalf("die Quellen sind %v", quellen)
	}

	// **Eine Kopie und keine Sicht auf das Original.**
	domain.Retrievalquellen()[0] = "verbogen"

	if domain.Retrievalquellen()[0] == "verbogen" {
		t.Error("die Aufzählung ließ sich von außen ändern")
	}
}

// **Ein Inhaltshash schützt nichts — und sagt das.**
//
// Das Issue: „Inhalts-Hash ist kein Geheimnisschutz." Er sagt, ob zwei
// Abschnitte denselben Text tragen, und sonst nichts.
func TestDerInhaltshashIstEinAbdruckUndKeinSchutz(t *testing.T) {
	t.Parallel()

	erster := domain.Retrievalhash("Lightning Bolt deals 3 damage to any target.")
	zweiter := domain.Retrievalhash("Lightning Bolt deals 3 damage to any target.")

	if erster != zweiter {
		t.Fatalf("derselbe Text gab zwei Abdrücke: %q und %q", erster, zweiter)
	}

	if !strings.HasPrefix(erster, "sha256:") {
		t.Fatalf("der Abdruck nennt sein Verfahren nicht: %q", erster)
	}

	if anderer := domain.Retrievalhash("etwas anderes"); anderer == erster {
		t.Fatal("zwei Texte gaben denselben Abdruck")
	}
}
