package service

// **Der Ausschnitt eines Treffers aus der Vektorhälfte** — Werkstück A15/A16
// von Phase 5.3.
//
// # Warum dieser Lauf im Paket steht
//
// **Weil `zusammengefuehrt` Verdrahtung ist und keine Fähigkeit.** Sie zu
// exportieren, nur damit ein Test sie sieht, machte aus einer inneren Naht
// eine Fläche (§2) — dieselbe Überlegung wie bei `retrievalnaht_intern_test.go`
// in `internal/app`.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ein Treffer, den nur die Vektorhälfte fand, trägt trotzdem einen
// Ausschnitt.**
//
// # Warum das ein eigener Lauf ist
//
// **Weil er am 2026-09-12 an einem gebauten Index aufgefallen ist**, nicht in
// einem Test: `mse recherche` gab vier Treffer und vier leere Zeilen darunter.
// Der Zusammenführer reicht den Ausschnitt zwischen den Listen weiter — aber
// nur, wenn **eine** ihn hat. Wen nur der Cosinus findet, den kennt die
// Volltexthälfte gar nicht, und FTS5 schneidet die Stelle heraus.
//
// **Ein Treffer ohne Ausschnitt sagt nicht, warum er einer ist.** Das Issue
// verlangt ihn ausdrücklich: „Treffer enthalten … Snippet."
func TestEinTrefferAusDerVektorhaelfteTraegtEinenAusschnitt(t *testing.T) {
	t.Parallel()

	text := "Creature — Human Wizard\nImmer wenn du einen Spontanzauber " +
		"oder eine Hexerei wirkst, erzeuge einen 1/1 blauen " +
		"Elementarwesen-Spielstein."

	zusammen := zusammengefuehrt([]domain.Retrievaltreffer{{
		Rang: 1, Ausschnitt: "",
		Dokument: domain.Retrievaldokument{
			Kennung: "oracle:a:0", Quelle: domain.QuelleOracle,
			Titel: "Sturmrufer", Text: text,
		},
	}})

	if len(zusammen) != 1 {
		t.Fatalf("die Liste ist %+v", zusammen)
	}

	if zusammen[0].Ausschnitt == "" {
		t.Fatal("der Treffer aus der Vektorhälfte trägt keinen Ausschnitt")
	}

	// **Und er ist ein Anfang des Textes und keine Erfindung.**
	if !strings.HasPrefix(text, strings.TrimSuffix(
		zusammen[0].Ausschnitt, Kuerzungszeichen)) {
		t.Errorf("der Ausschnitt %q steht so nicht im Text",
			zusammen[0].Ausschnitt)
	}
}

// **Ein Ausschnitt aus der Volltexthälfte bleibt stehen.**
//
// **Die Gegenprobe** (§10): ohne sie sagte der Lauf darüber nur, dass
// irgendein Ausschnitt entsteht — und nicht, dass der **bessere** gewinnt.
// FTS5 schneidet die **Fundstelle** heraus; ein Anfang des Textes ist der
// Rückfall und nicht der Normalfall.
func TestEinAusschnittAusDerVolltexthaelfteBleibtStehen(t *testing.T) {
	t.Parallel()

	dokument := domain.Retrievaldokument{
		Kennung: "oracle:a:0", Quelle: domain.QuelleOracle,
		Titel: "Sturmrufer", Text: "Ein langer Text, dessen Anfang hier steht.",
	}

	zusammen := zusammengefuehrt(
		[]domain.Retrievaltreffer{{Rang: 1, Dokument: dokument}},
		[]domain.Retrievaltreffer{{
			Rang: 1, Ausschnitt: "…die Fundstelle…", Dokument: dokument,
		}})

	if len(zusammen) != 1 || zusammen[0].Ausschnitt != "…die Fundstelle…" {
		t.Fatalf("der Ausschnitt ist %+v", zusammen)
	}
}

// **Ein langer Text wird gekürzt — und sagt es.**
//
// **Die andere Hälfte der Zeile darüber**: ohne sie stünde die Grenze da,
// ohne je erreicht worden zu sein, und `Kuerzungszeichen` wäre eine Zusage,
// die kein Lauf einlöst.
func TestEinLangerTextWirdGekuerzt(t *testing.T) {
	t.Parallel()

	lang := strings.Repeat("Wort ", Ausschnittzeichen)

	zusammen := zusammengefuehrt([]domain.Retrievaltreffer{{
		Rang: 1,
		Dokument: domain.Retrievaldokument{
			Kennung: "oracle:a:0", Quelle: domain.QuelleOracle, Text: lang,
		},
	}})

	ausschnitt := zusammen[0].Ausschnitt

	if !strings.HasSuffix(ausschnitt, Kuerzungszeichen) {
		t.Fatalf("der Ausschnitt sagt nicht, dass er gekürzt ist: %q",
			ausschnitt)
	}

	// **Gezählt wird in Zeichen und nicht in Bytes** — ein „ä" wäre sonst
	// zwei, und die Grenze verschöbe sich mit der Sprache.
	if len([]rune(ausschnitt)) > Ausschnittzeichen+len(Kuerzungszeichen) {
		t.Errorf("der Ausschnitt ist %d Zeichen lang",
			len([]rune(ausschnitt)))
	}
}

// **Ein Text, der kürzer ist als die Grenze, wird nicht gekürzt.**
func TestEinKurzerTextWirdNichtGekuerzt(t *testing.T) {
	t.Parallel()

	zusammen := zusammengefuehrt([]domain.Retrievaltreffer{{
		Rang: 1,
		Dokument: domain.Retrievaldokument{
			Kennung: "oracle:a:0", Quelle: domain.QuelleOracle, Text: "Kurz.",
		},
	}})

	if zusammen[0].Ausschnitt != "Kurz." {
		t.Fatalf("der Ausschnitt ist %q", zusammen[0].Ausschnitt)
	}
}
