package service_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein gescheitertes Skript sagt, was schon auf der Platte liegt** —
// #750.
//
// # Der Befund
//
// `copy_file`, `write_text_file`, `write_set_file` und die Bilder schreiben
// **während** der Auswertung. Scheitert ein späterer Ausdruck, kehrt die
// Ausfuhr zurück, bevor die Hauptdatei entsteht — und die Beigaben bleiben
// liegen. Die Oberfläche zeigte einen Fehler und sonst nichts; für die
// Nutzenden sah es aus, als sei nichts geschehen.
//
// Angestrichen vom Zweitblick durch Codex am 2026-08-21.
//
// # Warum nicht aufgeräumt wird
//
// **Weil MSE es auch nicht tut.** Seine Dateifunktionen schreiben sofort
// (`src/script/functions/export.cpp`, Zeile 368), die Hauptdatei entsteht erst
// nach dem Skript (`src/gui/html_export_window.cpp`, Zeile 73). Fassung f9d9356
// vom 2021-07-06. §15.4: keine stillschweigende Änderung bestehender Semantik —
// und ein Umbau auf ein Nebenverzeichnis bräche eine Vorlage, die ihren eigenen
// Pfad über `&{directory}` in die Ausgabe schreibt.
//
// **Falsch war die Zusage, nicht das Verhalten.** `Fuehre` sagte „ein Fehler im
// Skript ist ein Fehler und kein halbes Ergebnis" — das gilt für die
// **Ausgabedatei** und nicht für die Beigaben.
func TestEinGescheitertesSkriptNenntDieGeschriebenenBeigaben(t *testing.T) {
	t.Parallel()

	mitschrift := &beigabenmitschrift{}

	_, _, err := probelauf(service.Ausfuhrmittel{
		Beigaben: mitschrift, Verzeichnis: "spoiler-files",
	}).
		Fuehre(t.Context(), vorlageMit(
			`write_text_file(file: "a.txt", "erste Zeile") + `+
				`write_text_file(file: "b.txt", "zweite") + gibtesnicht()`),
			eineKarte(), nil, nil)
	if err == nil {
		t.Fatal("der Aufruf ging durch")
	}

	meldung := err.Error()

	// **Die Zahl und das Verzeichnis** — beides, denn eine Zahl ohne Ort sagt
	// nicht, wo man nachsieht.
	//
	// **Als ganzer Satz gesucht und nicht als Ziffer.** Der erste Anlauf
	// prüfte auf `"2"`, und das stand ohnehin in der Spaltenangabe der
	// Fehlermeldung — der Test war grün, bevor es die Zahl gab (§10).
	if !strings.Contains(meldung, `2 Datei(en) liegen bereits in "spoiler-files"`) {
		t.Errorf("die Meldung nennt den halben Stand nicht: %s", meldung)
	}
}

// **Und ohne geschriebene Beigabe steht der Satz nicht da.**
//
// Ein Hinweis, der immer erscheint, ist keiner: wer ihn jedes Mal liest, sieht
// ihn beim einen Mal nicht, an dem er zählt.
func TestOhneBeigabeNenntDerFehlerKeine(t *testing.T) {
	t.Parallel()

	_, _, err := probelauf(service.Ausfuhrmittel{
		Beigaben: &beigabenmitschrift{}, Verzeichnis: "spoiler-files",
	}).
		Fuehre(t.Context(), vorlageMit(`gibtesnicht()`), eineKarte(), nil, nil)
	if err == nil {
		t.Fatal("der Aufruf ging durch")
	}

	if strings.Contains(err.Error(), "spoiler-files") {
		t.Errorf("die Meldung spricht von Beigaben, obwohl keine geschrieben "+
			"wurde: %s", err.Error())
	}
}
