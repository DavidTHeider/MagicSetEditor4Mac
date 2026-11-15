package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// einstellungsprobe ist die Naht zu `service`, im Arbeitsspeicher.
type einstellungsprobe struct {
	werte  map[string]string
	fehler error

	// nutzerwerte sind die Einstellungen des Menschen — Werkstück 9 (#1351).
	nutzerwerte map[string]string
}

func (e *einstellungsprobe) Alle(context.Context) (map[string]string, error) {
	return e.werte, e.fehler
}

// Stand gibt die Wahl **und** das, was gilt — Werkstück 9 (#1351), Punkt 9.
//
// **Die Verschmelzung geht durch die Domäne und nicht an ihr vorbei.** Ein
// Doppel, das `Wirksam` gleich `Gewaehlt` setzte, wäre großzügiger als der
// Betrieb: ein Profil, das erbt, gäbe dann `erben` an die Oberfläche weiter,
// und kein Lauf hier merkte es (§3).
func (e *einstellungsprobe) Stand(context.Context) (map[string]service.Wahl, error) {
	if e.fehler != nil {
		return nil, e.fehler
	}

	stand := make(map[string]service.Wahl, len(e.werte))

	for _, schluessel := range domain.Einstellungsreihe() {
		gewaehlt, gesetzt := e.werte[schluessel]
		if !gesetzt {
			gewaehlt = domain.EinstellungsVorgabe(schluessel)
		}

		stand[schluessel] = service.Wahl{
			Gewaehlt: gewaehlt,
			Wirksam:  domain.Verschmolzen(schluessel, gewaehlt, e.nutzerwerte[schluessel]),
		}
	}

	return stand, nil
}

// Setzen schreibt — **und prüft vorher, wie der Betrieb es tut**.
//
// **Hier stand die Prüfung bis zum 2026-09-21 nicht**, und damit war dieses
// Doppel großzügiger als `service.Einstellungsverwalter.Setzen`, das
// `domain.PruefeEinstellung` vorschaltet. Genau davor warnt §3: wo eine
// Vorrichtung mehr durchlässt als der Betrieb, prüft ein Lauf sie und nicht
// den Weg. Gefunden beim Bau von UX-14.AK-01 (#1686) — ein Fenstermaß von 20
// Punkten kam durch, obwohl es die Domäne abweist.
func (e *einstellungsprobe) Setzen(_ context.Context, schluessel, wert string) error {
	if e.fehler != nil {
		return e.fehler
	}

	if fehler := domain.PruefeEinstellung(schluessel, wert); fehler != nil {
		return fehler
	}

	if e.werte == nil {
		e.werte = make(map[string]string)
	}

	e.werte[schluessel] = wert

	return nil
}

// **Die möglichen Werte kommen mit** (#245).
//
// Ohne sie baute das Fenster die Auswahlliste selbst — und dann stünden die
// erlaubten Werte an zwei Stellen, von denen die im Fenster die ungeprüfte
// wäre. Ein `grau` käme dort durch und im Bestand nicht an; sichtbar wäre nur,
// dass nichts geschieht.
func TestJedeEinstellungBringtIhreWerteMit(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitEinstellungen(&einstellungsprobe{
		werte: map[string]string{domain.EinstellungErscheinungsbild: "dunkel"},
	})

	antwort := anwendung.Einstellungen(t.Context())
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	// **Sechzehn feste seit Werkstück O3 von Phase 5.2**
	// (#1427):
	// sechs für die Einstellungsseite, **drei Vorgaben**, eine für die
	// Bildausgabe, **vier für den Arbeitsplatz** und **zwei Schranken** —
	// dazu ein Schalter je Dienst und eine Wahl je Frage. Welche wohin
	// gehört, sagt `Bereich`.
	//
	// **Die Zahl wird gerechnet, nicht abgeschrieben** — die vier Dienste
	// stehen in `domain.Dienste()`, und wer einen fünften aufnimmt, soll
	// nicht hier eine Zahl nachziehen müssen, die er nicht sieht.
	//
	// **Und je Anbieter eine Kennung** — Nacharbeit zu Werkstück 4 von
	// Phase 4.12 (#1351): dieselbe Überlegung, dieselbe Quelle.
	// **Und die Seitenleiste** — Werkstück U14 von Phase 5.1 (#1490): der
	// siebzehnte feste Schlüssel, auf Geräteebene.
	//
	// **Und zwei für die KI-Laufzeit** — Werkstück A7 von Phase 5.3 (#1446):
	// ob eine gefragt werden darf (Profil) und wo sie lauscht (Gerät). Zwei
	// Schlüssel für zwei Fragen; einer für beide hieße, dass ein Rechner
	// ohne Laufzeit auch keine Adresse behalten kann.
	// **Und je Dienst mit Gateway seine Adresse** — Werkstück 3 von A18
	// (#1533).
	// **Gezählt und nicht abgeschrieben**: welche Dienste eine tragen
	// dürfen, sagt `AdresseEinstellbar`, und eine feste Zahl hier liefe
	// beim vierten Anbieter auseinander.
	//
	// **Sechs für das Netz** kamen mit N5 (#1571): zwei Proxyadressen, die
	// Ausnahmeliste und drei Fristen.
	// **Und fünf Schranken für den Handel** — `T1` von Phase 5.6
	// (#1466):
	// zwei Schalter und drei Obergrenzen.
	//
	// **Und eine für die Kontoschau** — `T11` derselben Phase (#1476).
	//
	// **Und drei fürs Fenster** — UX-14.AK-01
	// (#1686):
	// wie es aufgeht, und die zwei Maße für „wie zuletzt".
	//
	// **Und eine fürs Kartenformular** — UX-05.AK-01
	// (#1677):
	// wie viel davon dasteht.
	//
	// **Und eine für den Tisch** — PLAY-04.AK-03
	// (#1712):
	// ob er die Karten selbst ordnet.
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
	// (#1681).
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
		t.Fatalf("%d Einstellungen, erwartet %d",
			len(antwort.Einstellungen), erwartet)
	}

	for _, eintrag := range antwort.Einstellungen {
		if eintrag.Bereich == "" {
			t.Errorf("%s nennt keinen Bereich", eintrag.Schluessel)
		}
	}

	// **Die Reihenfolge ist die der Anzeige** und steht in `app`, nicht im
	// Fenster: von oben nach unten nach Tragweite.
	if antwort.Einstellungen[0].Schluessel != domain.EinstellungSprache {
		t.Errorf("%+v", antwort.Einstellungen)
	}

	// **Jede aufzählbare bringt ihre Werte mit** — Werkstück 9 (#1351),
	// Punkt 8: seither gibt es eine, deren Wert ein **Pfad** ist, und ein
	// Verzeichnis lässt sich nicht aufzählen. Sie bringt eine **leere** Liste
	// und kein `null`; das hält `TestKeineBindungGibtEineListeAlsNull`.
	//
	// **Und je Anbieter eine Kennung**, aus demselben Grund: GitHub schreibt
	// `Ov23…`, Google eine Adresse, Microsoft eine UUID — eine Aufzählung
	// wäre eine Regel, die dieses Programm sich ausdenkt (Nacharbeit zu
	// Werkstück 4, #1351).
	//
	// **Als Aufzählung und nicht als Regel über Namen**: dass ein Schlüssel
	// freien Text nimmt, ist eine **Entscheidung** — an einem Präfix
	// festgemacht bliebe die nächste Auswahlliste versehentlich leer, ohne
	// dass es jemand merkt.
	//
	// **Drei kamen mit Werkstück O3 von Phase 5.2 dazu**
	// (#1427):
	// zwei Spaltenbreiten und eine Kostengrenze. Sie sind **nicht** ungeprüft
	// — für sie gilt eine Regel statt einer Liste, und die hält
	// `internal/domain/einstellungsreihe_test.go`.
	frei := freieEinstellungen()

	for _, eintrag := range antwort.Einstellungen {
		if frei[eintrag.Schluessel] {
			continue
		}

		if len(eintrag.Werte) == 0 {
			t.Errorf("%s bringt keine Werte mit", eintrag.Schluessel)
		}
	}

	// **Nach dem Schlüssel und nicht nach der Stelle.** Hier stand
	// `Einstellungen[1]`, und das war das Erscheinungsbild nur, solange
	// nichts vor ihm stand — mit der Darstellung (LOC-03, #1689) rückte es
	// eine Stelle weiter, und der Lauf verglich die falsche Zeile.
	//
	// **Und sie muss dastehen** — sonst prüfte die Schleife nichts und wäre
	// trotzdem grün.
	gefunden := false

	for _, eintrag := range antwort.Einstellungen {
		if eintrag.Schluessel != domain.EinstellungErscheinungsbild {
			continue
		}

		gefunden = true

		if eintrag.Wert != "dunkel" {
			t.Errorf("%+v", eintrag)
		}
	}

	if !gefunden {
		t.Errorf("das Erscheinungsbild steht nicht in der Antwort")
	}
}

func TestEinLesefehlerErreichtDasFenster(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitEinstellungen(&einstellungsprobe{fehler: errors.New("die Platte schweigt")})

	if antwort := anwendung.Einstellungen(t.Context()); antwort.Fehler == nil {
		t.Error("kein Fehler")
	}
}

func TestEineEinstellungLaesstSichSetzen(t *testing.T) {
	probe := &einstellungsprobe{}
	anwendung := app.NeueApp("")
	anwendung.MitEinstellungen(probe)

	antwort := anwendung.EinstellungSetzen(t.Context(), domain.EinstellungSchriftgroesse, "gross")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if probe.werte[domain.EinstellungSchriftgroesse] != "gross" {
		t.Errorf("%v", probe.werte)
	}
}

func TestEinSchreibfehlerErreichtDasFenster(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitEinstellungen(&einstellungsprobe{fehler: errors.New("die Platte schweigt")})

	if antwort := anwendung.EinstellungSetzen(t.Context(), "sprache", "en"); antwort.Fehler == nil {
		t.Error("kein Fehler")
	}
}

// dienstemitAdresse zählt, wie viele Dienste hinter ein Gateway dürfen.
//
// **Gezählt und nicht abgeschrieben** — Werkstück 3 von A18 (#1533): eine
// feste Zahl liefe beim vierten Anbieter auseinander, und der Lauf wäre dann
// rot, ohne dass jemand etwas falsch gemacht hätte.
func dienstemitAdresse() int {
	mit := 0

	for _, dienst := range domain.Dienste() {
		if dienst.AdresseEinstellbar() {
			mit++
		}
	}

	return mit
}
