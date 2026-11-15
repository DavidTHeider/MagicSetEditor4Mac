package laufbeleg

import (
	"fmt"
	"regexp"
	"strings"
)

// testlauf ist, was der Strom über einen Test gesagt hat.
type testlauf struct {
	gestartet bool
	ende      string
	dauer     float64
	fehler    []string
	ausgabe   []string
	abbruch   string
}

// paketlauf ist, was der Strom über ein Paket gesagt hat.
type paketlauf struct {
	zwischenspeicher bool
	baufehler        string
	ende             string
	abbruch          string
}

// sammlung hält den gelesenen Strom, nach Paket und Fall.
type sammlung struct {
	pakete     map[string]*paketlauf
	tests      map[Fall]*testlauf
	reihe      []Fall
	bau        map[string][]string
	kennzahlen map[string]string
	fehler     []string
}

// kennzahlzeile findet `kennzahl name=wert` in einer Zeile von `t.Log`.
var kennzahlzeile = regexp.MustCompile(`kennzahl ([A-Za-z0-9_.\-]+)=(\S+)`)

// fundstelle ist der Vorsatz, den `t.Log` vor eine Zeile setzt.
var fundstelle = regexp.MustCompile(`^\s*\S+\.go:\d+: `)

// sammle liest einen Strom.
func sammle(ereignisse []Ereignis) *sammlung {
	gesammelt := &sammlung{
		pakete:     map[string]*paketlauf{},
		tests:      map[Fall]*testlauf{},
		bau:        map[string][]string{},
		kennzahlen: map[string]string{},
		fehler:     []string{},
	}

	for _, ereignis := range ereignisse {
		switch {
		case ereignis.Action == "build-output":
			pfad := ohneKlammer(ereignis.ImportPath)
			gesammelt.bau[pfad] = append(gesammelt.bau[pfad],
				strings.TrimRight(ereignis.Output, "\n"))
		case ereignis.Action == "build-fail":
			pfad := ohneKlammer(ereignis.ImportPath)
			gesammelt.paket(pfad).baufehler = gesammelt.ersteBauzeile(pfad)
		case ereignis.Test == "":
			gesammelt.paketereignis(ereignis)
		default:
			gesammelt.testereignis(ereignis)
		}
	}

	return gesammelt
}

// ohneKlammer nimmt den Zusatz, den ein Bauereignis trägt:
// `probe.test/kaputt [probe.test/kaputt.test]` ist das Paket `probe.test/kaputt`.
func ohneKlammer(pfad string) string {
	vorne, _, _ := strings.Cut(pfad, " [")

	return vorne
}

// ersteBauzeile gibt die erste Zeile des Übersetzers, die keine Überschrift ist.
func (s *sammlung) ersteBauzeile(pfad string) string {
	for _, zeile := range s.bau[pfad] {
		if !strings.HasPrefix(zeile, "# ") && strings.TrimSpace(zeile) != "" {
			return strings.TrimSpace(zeile)
		}
	}

	return "ohne Meldung"
}

// paket gibt den Lauf eines Pakets und legt ihn an.
func (s *sammlung) paket(pfad string) *paketlauf {
	lauf, gibt := s.pakete[pfad]
	if !gibt {
		lauf = &paketlauf{}
		s.pakete[pfad] = lauf
	}

	return lauf
}

// paketereignis liest ein Ereignis ohne Test.
func (s *sammlung) paketereignis(ereignis Ereignis) {
	lauf := s.paket(ereignis.Package)

	switch ereignis.Action {
	case "output":
		// **Nur diese Zeile verrät den Zwischenspeicher** — gemessen, nicht
		// aus der Dokumentation: die Ereignisse davor tragen eine Zeit.
		if strings.Contains(ereignis.Output, "(cached)") {
			lauf.zwischenspeicher = true
		}

		s.abbruchIn(ereignis.Output, &lauf.abbruch)
	case "pass", "fail", "skip":
		lauf.ende = ereignis.Action

		if ereignis.FailedBuild != "" && lauf.baufehler == "" {
			lauf.baufehler = s.ersteBauzeile(ohneKlammer(ereignis.FailedBuild))
		}
	}
}

// testereignis liest ein Ereignis eines Tests oder Unterfalls.
//
// **Ein Unterfall zählt zu seinem obersten Test**: dort steht das Ergebnis,
// das zählt. Seine Ausgabe wird mitgelesen — eine Kennzahl oder ein Absturz
// im Unterfall gehört dem Ganzen.
//
// **Und ein Ereignis eines Tests belegt, dass sein Paket lief** — auch ohne
// `start` davor, etwa in einem abgeschnittenen Strom.
func (s *sammlung) testereignis(ereignis Ereignis) {
	s.paket(ereignis.Package)

	oberster, _, unterfall := strings.Cut(ereignis.Test, "/")
	schluessel := Fall{Paket: ereignis.Package, Test: oberster}

	lauf, gibt := s.tests[schluessel]
	if !gibt {
		lauf = &testlauf{}
		s.tests[schluessel] = lauf
		s.reihe = append(s.reihe, schluessel)
	}

	switch ereignis.Action {
	case "output":
		s.ausgabeLesen(lauf, ereignis)
	case "run":
		lauf.gestartet = lauf.gestartet || !unterfall
	case "pass", "fail", "skip":
		if !unterfall {
			lauf.ende = ereignis.Action
			lauf.dauer = ereignis.Elapsed
		}
	}
}

// ausgabeLesen liest eine Zeile eines Tests.
func (s *sammlung) ausgabeLesen(lauf *testlauf, ereignis Ereignis) {
	if ereignis.OutputType == "frame" {
		return
	}

	zeile := strings.TrimSpace(fundstelle.ReplaceAllString(ereignis.Output, ""))
	if zeile == "" {
		return
	}

	if s.abbruchIn(zeile, &lauf.abbruch) {
		paket := s.paket(ereignis.Package)
		if paket.abbruch == "" {
			paket.abbruch = lauf.abbruch
		}
	}

	if treffer := kennzahlzeile.FindStringSubmatch(zeile); treffer != nil {
		s.kennzahl(treffer[1], treffer[2])
	}

	if strings.HasPrefix(ereignis.OutputType, "error") {
		lauf.fehler = append(lauf.fehler, zeile)
	} else {
		lauf.ausgabe = append(lauf.ausgabe, zeile)
	}
}

// abbruchIn merkt sich die erste Zeile, mit der ein Prozess abbricht.
//
// **Ein Absturz ist kein roter Fall**: danach läuft nichts mehr in diesem
// Paket, und was fehlt, fehlt ohne eigenes Ereignis. Eine Zeitüberschreitung
// meldet nicht einmal das Ende ihres Falls — gemessen am 2026-09-22.
func (s *sammlung) abbruchIn(zeile string, ziel *string) bool {
	zeile = strings.TrimSpace(zeile)
	if !strings.HasPrefix(zeile, "panic: ") && !strings.HasPrefix(zeile, "fatal error: ") {
		return false
	}

	if *ziel == "" {
		*ziel = zeile
	}

	return true
}

// kennzahl hält eine Kennzahl fest — und nennt einen Widerspruch.
func (s *sammlung) kennzahl(name, wert string) {
	bisher, gibt := s.kennzahlen[name]
	if gibt && bisher != wert {
		s.fehler = append(s.fehler, fmt.Sprintf(
			"kennzahl %s steht zweimal da, mit %s und mit %s", name, bisher, wert))

		return
	}

	s.kennzahlen[name] = wert
}

// bewerte sagt, was ein Fall gezeigt hat.
func (s *sammlung) bewerte(einer Fall, art Art) Fallergebnis {
	ergebnis := Fallergebnis{Fall: einer, Art: art}
	paket := s.pakete[einer.Paket]
	lauf := s.tests[einer]

	if lauf != nil {
		ergebnis.Dauer = lauf.dauer
	}

	ergebnis.Ausfuehrung, ergebnis.Ergebnis, ergebnis.Grund = s.stand(paket, lauf)

	return ergebnis
}

// stand gibt Ausführung, Ergebnis und Grund — in der Reihenfolge, in der
// eine Antwort die nächste ausschließt.
func (s *sammlung) stand(paket *paketlauf, lauf *testlauf) (ausfuehrung, ergebnis Status, grund string) {
	switch {
	case paket == nil:
		return NichtGelaufen, NichtGelaufen, "das Paket lief nicht"
	case paket.baufehler != "":
		return Gescheitert, Gescheitert, "Übersetzung gescheitert: " + paket.baufehler
	case paket.zwischenspeicher:
		return NichtGelaufen, NichtGelaufen,
			"aus dem Zwischenspeicher — gemessen wurde nichts"
	case lauf == nil || !lauf.gestartet:
		if paket.abbruch != "" {
			return NichtGelaufen, NichtGelaufen,
				"nicht gelaufen — das Paket brach vorher ab: " + paket.abbruch
		}

		return NichtGelaufen, NichtGelaufen, "nicht gelaufen"
	case lauf.abbruch != "":
		return Gescheitert, Gescheitert, "abgebrochen: " + lauf.abbruch
	case lauf.ende == "":
		return Gescheitert, Gescheitert, "kein Ende gemeldet — der Prozess brach ab"
	case lauf.ende == "pass":
		return Bestanden, Bestanden, ""
	case lauf.ende == "skip":
		return Uebersprungen, Uebersprungen, grundAus(nil, lauf.ausgabe, "ohne Grund")
	default:
		// **Rot, aber gemessen**: der Fall lief bis zu seinem Urteil.
		return Bestanden, Gescheitert, grundAus(lauf.fehler, lauf.ausgabe, "ohne Meldung")
	}
}

// grundAus gibt die aussagekräftige Zeile: **die erste Fehlerzeile**, denn
// die nennt, was zuerst schiefging — **sonst die letzte Ausgabe**, denn
// `t.Skip` schreibt seinen Grund nach allem, was davor geloggt wurde.
func grundAus(fehler, ausgabe []string, sonst string) string {
	switch {
	case len(fehler) > 0:
		return fehler[0]
	case len(ausgabe) > 0:
		return ausgabe[len(ausgabe)-1]
	default:
		return sonst
	}
}
