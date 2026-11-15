package msepaket

import "fmt"

// Die `auto replace:`-Blöcke eines Spielpakets.
//
// # Warum das eine eigene Datei ist
//
// `ersetzung.go` hält die **Regel** und den Ersetzer darüber; hier steht, wie
// die Regeln aus einem Paket kommen. Beides in einer Datei hieße, die Frage
// „was ist eine Regel" mit „wo steht sie" zu vermischen (§2).
//
// # Der Anlass
//
// Bis zum 2026-08-23 gab es `AutoErsetzung` und `NeuerErsetzer` — und **keine
// Stelle, die sie aus einem Paket liest**
// (#897).
// `magic.mse-game/auto_replace` führt rund 500 Regeln, und keine erreichte je
// die Anwendung.

// AutoErsetzungen gibt die Regeln aus den `auto replace:`-Blöcken des Kopfes.
//
// # Was aus MSE2 übernommen ist
//
// **`whole word` steht ohne Angabe auf `true`** — `AutoReplace::AutoReplace`,
// `src/data/word_list.cpp`, Zeilen 38–42, Fassung f9d9356 vom 2021-07-06.
//
// **Eine Regel ohne Muster ist abgeschaltet** — `after_reading(AutoReplace&)`,
// ebenda Zeilen 51–55: „if (ar.match.empty()) ar.enabled = false". Sie wird
// hier gar nicht erst aufgenommen: eine abgeschaltete Regel, die mitreist,
// wäre ein Zustand, den niemand einschalten kann.
//
// **`enabled: false` schaltet eine Regel ab.** `AutoReplace` liest das Feld aus
// der Datei (`IMPLEMENT_REFLECTION_NO_SCRIPT`, `src/data/word_list.cpp`, Zeilen
// 44–49), und `tryAutoReplace` prüft es, bevor eine Regel greift
// (`src/gui/value/text.cpp`, Zeile 1038) — beides Fassung f9d9356 vom
// 2021-07-06.
//
// # Zwei Schalter, nicht einer
//
// Es sind **zwei** Schalter: `use_auto_replace` ist der globale
// (`src/data/settings.hpp`, Zeile 83), `enabled` der je Regel — und
// `GameSettings::initDefaults` schiebt die Regelzeiger des **Spiels** in seine
// Liste (`src/data/settings.cpp`, Zeilen 88–104), sodass der Zustand des
// Pakets mitreist.
//
// **Der globale Schalter fehlt**, und das ist benannt statt begründet: es
// gibt keine Einstellung, mit der jemand die Autoersetzung abschaltet. Das ist
// eine Lücke, kein Vorsatz.
// **Die Befunde fallen hier weg**, und das ist heute noch keine Antwort auf
// §8. Wohin sie gehören, entscheidet
// #932: der
// Weg von der Regelquelle bis zur Oberfläche hat noch keinen Fehlerkanal.
// Diese Fassung nennt die Lücke, statt sie unsichtbar zu machen.
func (p *Paket) AutoErsetzungen() []AutoErsetzung {
	regeln, _ := p.AutoErsetzungenMitBefunden()

	return regeln
}

// AutoErsetzungenMitBefunden gibt die Regeln und was beim Lesen auffiel.
//
// **Zwei Rückgaben, weil ein unbekannter Boolwert kein Abbruch ist** — MSE
// warnt und liest weiter (`Reader::handle(bool&)`, `src/util/io/reader.cpp`,
// Zeilen 324–332, Fassung f9d9356 vom 2021-07-06). Fremde Daten sind der
// Normalfall (§8): ein Paket von 2007 wird nicht abgewiesen, sein Befund wird
// benannt.
func (p *Paket) AutoErsetzungenMitBefunden() ([]AutoErsetzung, []Befund) {
	var (
		regeln  []AutoErsetzung
		befunde []Befund
	)

	for _, knoten := range p.Kopf.Alle("auto replace") {
		muster := WertOderBlock(knoten.Ersten("match"))
		if muster == "" {
			continue
		}

		an, anBefund := wahrheit(knoten, "enabled")
		befunde = append(befunde, anBefund...)

		if !an {
			continue
		}

		ganzesWort, wortBefund := wahrheit(knoten, "whole word")
		befunde = append(befunde, wortBefund...)

		regeln = append(regeln, AutoErsetzung{
			Muster:     muster,
			Ersatz:     WertOderBlock(knoten.Ersten("replace")),
			GanzesWort: ganzesWort,
		})
	}

	return regeln, befunde
}

// WertOderBlock gibt den Wert eines Knotens — auch wenn er darunter steht.
//
// **Ein `match:` kann mehrzeilig sein**, und im echten Paket ist es das:
//
//	auto replace:
//		match:
//			 -
//		replace:
//			 —
//
// Der umschließende Abstand ist die ganze Regel; hinter dem Doppelpunkt wäre
// er nicht zu schreiben, weil der Zerteiler links kappt — genau wie MSEs
// `Reader::readLine` (`src/util/io/reader.cpp`, Zeile 224, Fassung f9d9356 vom
// 2021-07-06).
//
// **Was das nicht kann**: einen Block, der selbst wieder eingerückte Zeilen
// trägt. `Skripttext` streicht führende Tabulatoren, und eine Regel mit einem
// eingerückten Muster gäbe es entstellt zurück. In `magic.mse-game` kommt das
// nicht vor; käme es, wäre es hier zu sehen und nicht stillschweigend falsch.
// WertOderBlock ist **seit #1062 ausgeführt**: `internal/service` deutet die
// `keyword:`-Blöcke eines Sets und braucht dieselbe Regel. Eine zweite Fassung
// daneben liefe beim ersten mehrzeiligen Erinnerungstext auseinander.
func WertOderBlock(knoten *Knoten) string {
	if knoten == nil {
		return ""
	}

	if knoten.Wert != "" {
		return knoten.Wert
	}

	return knoten.Skripttext()
}

// wahr sind die Schreibweisen, die MSE als wahr liest.
//
// `Reader::handle(bool&)`, `src/util/io/reader.cpp`, Zeilen 324–332, Fassung
// f9d9356 vom 2021-07-06 — **drei und drei**, nicht eine und eine.
var wahr = map[string]bool{"true": true, "1": true, "yes": true}

// falsch sind die Schreibweisen, die MSE als falsch liest.
var falsch = map[string]bool{"false": true, "0": true, "no": true}

// wahrheit liest einen Wahrheitswert — ohne Angabe wahr, bei Unbekanntem ein
// Befund.
//
// # Nicht nur `"false"`
//
// `angabe.Wert != "false"` kennte **eine** Schreibweise für falsch; ein Paket
// mit `enabled: no` oder `whole word: 0` würde damit anders gelesen als in
// MSE2 (#930).
//
// # Warum ein unbekannter Wert wahr bleibt
//
// MSEs `warning` hält den Leser nicht auf, und der Wert bleibt, was er vor dem
// Lesen war — bei `AutoReplace` also `true` (`src/data/word_list.cpp`, Zeilen
// 38–42, Fassung f9d9356 vom 2021-07-06). Ein Befund sagt es trotzdem: **gedeutet oder benannt, eine dritte
// Antwort gibt es nicht** (§8).
//
// **Beide Felder haben dieselbe Vorgabe**, und beide aus derselben Stelle;
// zwei Helfer wären zwei Orte für dieselbe Regel.
func wahrheit(knoten *Knoten, schluessel string) (bool, []Befund) {
	return wahrheitMit(knoten, schluessel, true)
}

// wahrheitMit ist dasselbe mit gewählter Vorgabe.
//
// # Warum es beide gibt
//
// **Die Vorgabe steht im Altcode und ist nicht überall dieselbe.**
// `whole word` und `enabled` stehen ohne Angabe auf `true`
// (`src/data/word_list.cpp:38–42`, Fassung f9d9356 vom 2021-07-06), `is
// default` eines Schlüsselwortmodus auf `false` und `optional` einer
// Parameterart auf `true` (`src/data/keyword.cpp:20–27`, `keyword.hpp:81`,
// dieselbe Fassung).
func wahrheitMit(knoten *Knoten, schluessel string, vorgabe bool) (bool, []Befund) {
	angabe := knoten.Ersten(schluessel)
	if angabe == nil {
		return vorgabe, nil
	}

	if wahr[angabe.Wert] {
		return true, nil
	}

	if falsch[angabe.Wert] {
		return false, nil
	}

	return vorgabe, []Befund{{
		Stelle: angabe.Stelle,
		Meldung: fmt.Sprintf(
			"%s: %q ist kein Wahrheitswert — erwartet wird true, 1, yes, false, 0 oder no",
			schluessel, angabe.Wert),
	}}
}
