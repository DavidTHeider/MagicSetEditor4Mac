package schriften

import (
	"path/filepath"
	"strings"

	"golang.org/x/image/font/sfnt"
)

// Der Name, den eine Schrift über sich selbst führt.
//
// # Warum das nötig wurde
//
// `magic-m15.mse-style` verlangt `Beleren Bold` und `MPlantin`. Im bezogenen
// `M15-Magic-Pack` liegen sie als `beleren-bold_P1.01.ttf` und `mplantin.ttf`.
// Die zweite fand der Dateinamensvergleich, die erste nie — und damit trug eine
// gezeichnete Karte keinen Text
// (#240).
//
// Der Dateiname ist eine Angabe des Menschen, der die Datei abgelegt hat; der
// Name **in** der Datei ist die Angabe der Schrift über sich selbst. Genau den
// meint eine Vorlage: MSE reicht ihn an das Betriebssystem weiter, und das sieht
// ebenfalls in die Datei.
//
// # Warum keine Liste
//
// Eine Zuordnung „Dateiname → Schriftname" wäre eine Tabelle, die mit jeder
// neuen Sammlung veraltet — und `M15-Magic-Pack` allein führt über hundert
// Schriften.
//
// # Was das kostet
//
// **Jede Datei muss geöffnet werden.** Deshalb steht dieser Weg *hinter* dem
// Dateinamensvergleich: der ist billig und trifft in den meisten Fällen zu.
// Erst wenn er nichts findet, wird gelesen — und das Ergebnis bleibt für die
// Sitzung stehen ([Suche.namensindex]).

// NameAus gibt den **vollen** Namen, den eine Schriftdatei über sich führt.
func NameAus(roh []byte) (string, bool) {
	namen := NamenAus(roh)
	if len(namen) == 0 {
		return "", false
	}

	return namen[0], true
}

// NamenAus gibt **alle** Namen, unter denen eine Schriftdatei zu finden sein
// soll — der volle Name zuerst, dann die Familie.
//
// # Warum beide, und nicht nur der volle
//
// **MSE kennt kein `bold name:`.** Kursiv bekommt eine eigene Schriftdatei
// (`italic_name`), Fett ist ein **Attribut**:
//
//	REFLECT(name);
//	REFLECT(weight);
//	REFLECT(italic_name);
//
// `src/data/font.cpp`, Zeilen 119–134, Fassung f9d9356 vom 2021-07-06. Ein
// Stil schreibt `name: Beleren Small Caps` samt `weight: bold`, und wx sucht
// den passenden Schnitt **der Familie**.
//
// Die Datei im Pack heißt entsprechend:
//
//	belerensmallcaps-bold.ttf
//	    Familie        "Beleren Small Caps"
//	    Unterfamilie   "Bold"
//	    voller Name    "Beleren Small Caps Bold"
//
// **Der volle Name steht vorn**, denn beim Eintragen gewinnt der erste: wer
// „Beleren Small Caps Bold" verlangt, meint diesen Schnitt und nicht
// irgendeinen der Familie.
//
// # Und Familie mit Schnitt
//
// **Fett ist der Schnitt „Bold" der Familie**, auch wenn kein Name das so
// zusammenschreibt. `matrixb.ttf` führt die Familie „Matrix", den Schnitt
// „Bold" und den vollen Namen „Matrix-Bold"; diese Suche fragt nach
// „Matrix Bold" (0023) und fiel bis F02
// (#2210)
// auf den aufrechten Schnitt zurück. Deshalb steht jede Datei auch unter
// Familie und Schnitt zusammen — einmal aus den Namen 1 und 2, einmal aus den
// typographischen 16 und 17.
//
// **Zwei Schnitte derselben Familie belegen denselben Familiennamen**, und
// den bekommt der aufrechte — siehe [Suche.eintragen].
func NamenAus(roh []byte) []string {
	schrift, err := sfnt.Parse(roh)
	if err != nil {
		return nil
	}

	return namenDer(schrift)
}

// namenAusDatei gibt dieselben Namen wie [NamenAus], **ohne die Datei ganz zu
// lesen**: `sfnt.ParseReaderAt` holt sich nur die Tabellen, die es braucht —
// und dazu, ob die Datei ein fetter Schnitt ist.
//
// Bis E03 (#2240)
// las der Index jede Schrift seiner Orte ganz — `Apple Color Emoji.ttc` mit
// 192 MB eingeschlossen, rund 860 MB je frischem Zeichenwerk.
func namenAusDatei(pfad string) (namen []string, fett bool) {
	datei, err := schriftOeffnen(pfad)
	if err != nil {
		return nil, false
	}

	defer datei.Close() //nolint:errcheck // nur gelesen

	schrift, err := sfnt.ParseReaderAt(datei)
	if err != nil {
		return nil, false
	}

	return namenDer(schrift), fetterSchnitt(schrift)
}

// namenDer liest voller Name, Familie mit Schnitt und Familie aus einer
// geöffneten Schrift — in dieser Reihenfolge, die genaueren zuerst.
func namenDer(schrift *sfnt.Font) []string {
	var puffer sfnt.Buffer

	name := func(kennung sfnt.NameID) string {
		gelesen, err := schrift.Name(&puffer, kennung)
		if err != nil {
			return ""
		}

		return gelesen
	}

	familie, schnitt := name(sfnt.NameIDFamily), name(sfnt.NameIDSubfamily)
	typoFamilie, typoSchnitt := name(sfnt.NameIDTypographicFamily), name(sfnt.NameIDTypographicSubfamily)

	var namen []string

	for _, kandidat := range []string{
		name(sfnt.NameIDFull),
		zusammen(familie, schnitt),
		zusammen(typoFamilie, typoSchnitt),
		familie,
		typoFamilie,
	} {
		if kandidat != "" && !enthaelt(namen, kandidat) {
			namen = append(namen, kandidat)
		}
	}

	return namen
}

// zusammen schreibt Familie und Schnitt so, wie diese Suche fett fragt:
// „Matrix" und „Bold" werden „Matrix Bold". Fehlt eines, gibt es keinen Namen.
func zusammen(familie, schnitt string) string {
	if familie == "" || schnitt == "" {
		return ""
	}

	return familie + " " + schnitt
}

// fetterSchnitt sagt, ob eine Schrift ein fetter Schnitt ist — gelesen am
// Namen ihres Schnitts, dem typographischen zuerst.
//
// **Am Namen und nicht am Gewicht aus `OS/2`**: diese Suche fragt fett über
// den Namen (0023), und dieselbe Lesart gehört auf beide Seiten.
func fetterSchnitt(schrift *sfnt.Font) bool {
	var puffer sfnt.Buffer

	for _, kennung := range []sfnt.NameID{sfnt.NameIDTypographicSubfamily, sfnt.NameIDSubfamily} {
		if schnitt, err := schrift.Name(&puffer, kennung); err == nil && schnitt != "" {
			return strings.Contains(strings.ToLower(schnitt), "bold")
		}
	}

	return false
}

// enthaelt sagt, ob ein Name schon dasteht.
func enthaelt(namen []string, gesucht string) bool {
	for _, name := range namen {
		if name == gesucht {
			return true
		}
	}

	return false
}

// namensindex baut die Zuordnung „eigener Name → Datei" über alle Orte.
//
// **Einmal je Sitzung und erst bei Bedarf.** Über die Systemschriften und ein
// bezogenes Paket zu laufen und jede Datei zu öffnen kostet spürbar; wer nie
// eine Schrift unter einem abweichenden Namen sucht, zahlt es nie.
//
// **Ohne die Pakete.** Der Index deckt die Orte ab; die Pakete haben ihren
// eigenen Weg ([Suche.paketdatei]), weil ihre Wurzel ein ganzer Baum ist und
// nicht eine Liste von Verzeichnissen.
func (s *Suche) namensindex() map[string]string {
	if s.namen != nil {
		return s.namen
	}

	s.namen = map[string]string{}

	for _, ort := range s.orte {
		for _, datei := range schriftdateien(ort) {
			s.merkeNamen(filepath.Join(ort, datei))
		}
	}

	return s.namen
}

// merkeNamen liest eine Schriftdatei und trägt ihren eigenen Namen ein.
func (s *Suche) merkeNamen(pfad string) {
	// Der Pfad kommt aus der Verzeichnisliste eines Ortes dieser Suche, nicht
	// aus einer Eingabe — dieselbe Lage wie in [Suche.Suche].
	s.eintragen(s.namen, pfad)
}

// eintragen trägt eine Datei unter **jedem Namen** ein, den sie führt —
// voller Name, Familie mit Schnitt und Familie. Der Grund steht bei
// [NamenAus]: MSE kennt kein `bold name:`, und ein Stil nennt die **Familie**
// (#502).
//
// **Der erste gewinnt.** Die Orte stehen in der Reihenfolge ihrer
// Vorrangigkeit, und was jemand selbst installiert hat, ist die Schrift, die er
// meint (#140).
//
// **Mit einer Ausnahme: der aufrechte Schnitt schlägt den fetten.** Beide
// führen denselben Familiennamen; MSE reicht für einen Stil ohne `weight: bold`
// `wxFONTWEIGHT_NORMAL` weiter (`Font::toWxFont`, `src/data/font.cpp`,
// Zeilen 83–110, Fassung f9d9356 vom 2021-07-06), und das System nimmt den
// aufrechten — gleich, welcher im Verzeichnis vorn steht. Fett fragt diese
// Suche ohnehin unter „Familie Bold" ([Suche.schnitt]).
func (s *Suche) eintragen(index map[string]string, pfad string) {
	// **Bei jedem Eintragen neu gelesen**: ein Index wird neu gebaut, wenn sich
	// eine Paketwurzel ändert, und an demselben Pfad kann dann ein anderer
	// Schnitt liegen.
	namen, fett := namenAusDatei(pfad)
	if fett {
		s.fettePfade[pfad] = true
	} else {
		delete(s.fettePfade, pfad)
	}

	for _, name := range namen {
		schluessel := strings.ToLower(name)
		if bisher := index[schluessel]; bisher == "" || (s.fettePfade[bisher] && !fett) {
			index[schluessel] = pfad
		}
	}
}
