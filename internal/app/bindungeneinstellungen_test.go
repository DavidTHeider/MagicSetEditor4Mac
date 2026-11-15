package app_test

// **Die Einstellungsbindungen** — derselbe Vertragstest, eine andere Frage.
//
// **Eine eigene Datei, weil `bindungenlesen_test.go` mit ihr über 400 Zeilen
// ginge** (§2). Der Schnitt folgt der Frage: dort stehen die Bindungen über
// Sets, Karten und Bilder, hier die über das, was ein Mensch **einstellt**.
//
// **Und hier steht die Zahl, die mit jeder neuen Einstellung wächst** — sie
// ist gerechnet und nicht abgeschrieben, und ihre Geschichte steht im
// Kommentar darunter.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestJedeEinstellungsbindungErreichtIhreAnwendung(t *testing.T) {
	stand := bindungsprobe(t)
	anwendung, bindungen := stand.Anwendung, stand.Bindungen

	t.Run("Einstellungen", func(t *testing.T) {
		anwendung.MitEinstellungen(&einstellungsprobe{})

		antwort := bindungen.Einstellungen(t.Context())
		if antwort.Fehler != nil {
			t.Fatalf("%+v", antwort.Fehler)
		}

		// **Gerechnet und nicht abgeschrieben**: sechzehn der Anwendung plus
		// ein Schalter je Dienst (Phase 4.8). Die sechste kam mit der
		// Einführung als Pop-up dazu (#1325), die siebte mit der
		// Anzeigeschrift, die **drei Vorgaben** mit Werkstück 9 von
		// Phase 4.12 (#1351) — und je Anbieter eine **Kennung** mit der
		// Nacharbeit zu Werkstück 4 desselben Issues. **Vier für den
		// Arbeitsplatz und zwei Schranken** kamen mit Werkstück O3 von
		// Phase 5.2 (#1427).
		// **Und die Seitenleiste** — U14 von Phase 5.1 (#1490), **zwei für
		// die KI-Laufzeit** mit A7 (#1446), **je Dienst mit Gateway seine
		// Adresse** mit A18 (#1533) und **sechs fürs Netz** mit N5 (#1571).
		//
		// **Und fünf Schranken für den Handel** — `T1` von Phase 5.6
		// (#1466):
		// zwei Schalter und drei Obergrenzen.
		//
		// **Und eine für die Kontoschau** — `T11` derselben Phase
		// (#1476):
		// wer einen Marktplatz eingeschaltet hat, hat damit nicht sein
		// eigenes Konto freigegeben.
		//
		// **Und drei fürs Fenster** — UX-14.AK-01
		// (#1686):
		// wie es aufgeht, und die zwei Maße für „wie zuletzt".
		//
		// **Und eine fürs Kartenformular** — UX-05.AK-01
		// (#1677):
		// wie viel davon dasteht.
		//
		// **Und drei für Darstellung, Zeitzone und Vorgabewährung** —
		// LOC-03.AK-03
		// (#1689).
		//
		// **Und die Beschriftung des Tisches** — seit dem 2026-09-24: ob und wie
		// groß Zonennamen und Sitzköpfe auf der Leinwand stehen.
		//
		// **Und die Belegung von ⌘0 bis ⌘9** — UX-02.AK-03
		// (#1674).
		//
		// **Und die Lage der Leiste** — UX-14.AK-02
		// (#1686).
		//
		// **Und vier für die Gestaltung** — UX-09
		// (#1681):
		// Farbwelt, Reichweite der Schrift, Fensterbild und eigene Farben.
		// **Und drei gemerkte Ansichten** — UX-11.AK-04
		// (#1683).
		// **Und drei Ordner** — OPS-02
		// (#1719).
		// **Und neun des Protokolls** — OPS-04 (#1721).
		// **Und fünf der Messung** — OPS-05 (#1722), eine Grenze — OPS-08.
		// **Und die gespeicherten Filter** — UX-12 (#1684).
		// **Und die eigene Übersetzung** — LOC-02 (#1688).
		// **Und die Schnittstelle der Suche** — NET-02 (#1706), vier Zeugnispfade — NET-03 (#1707).
		// **Und je Schalter einer** — OPS-10 (#1908), aus dem Verzeichnis.
		// **Und der Kontingentschalter** — QA-08.AK-03 (#1739).
		// **Und Apples Gerätemodell** — AI-05 (#2042).
		// **Und das Modell des Vorrangs** — #2054.
		// **Und die Lauschwahl der eigenen Dienste** — G08 (#2272).
		// **Und der Ordner des freiwilligen Plans** — OPS01 (#2347).
		// **Und der zuletzt gewählte Bildordner** — AP-10 (#2367).
		// **Und das Ziel der Ausgabe je Set** — AP-10 (#2367).
		if erwartet := 81 + len(domain.Dienste()) + len(domain.Fragen()) +
			len(domain.Anbieter()) + dienstemitAdresse() + len(domain.Schalterverzeichnis()); len(
			antwort.Einstellungen) != erwartet {
			t.Errorf("%d statt %d: %+v",
				len(antwort.Einstellungen), erwartet, antwort.Einstellungen)
		}
	})

	t.Run("EinstellungSetzen", func(t *testing.T) {
		anwendung.MitEinstellungen(&einstellungsprobe{})

		if antwort := bindungen.EinstellungSetzen(t.Context(), "erscheinungsbild", "dunkel"); antwort.Fehler != nil {
			t.Fatalf("%+v", antwort.Fehler)
		}
	})
}
