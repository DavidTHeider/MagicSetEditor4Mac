// Package schriften findet Schriftdateien auf dem Rechner.
//
// # Warum ein eigenes Paket
//
// `internal/render` sagt, **was** es braucht (`render.Schriftquelle`), und weiß
// nicht, wo Schriften liegen — das ist eine Eigenschaft der Plattform, keine des
// Rasterers. Dieses Paket ist die andere Hälfte: es kennt die Orte und nichts
// vom Zeichnen.
//
// # Gesucht, nicht mitgeliefert
//
// Schriften sind lizenziert (`DRITTANBIETER.md`, AGENTS.md §11). Eine
// mitzuliefern hieße, ihre Lizenz zu prüfen und einzuhalten, und für `MPlantin`
// oder `Beleren` ginge das nicht. Wer eine Vorlage benutzt, hat ihre Schriften
// installiert — oder bekommt einen benannten Befund.
package schriften

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// ErrNichtGefunden sagt, dass keine Datei zu diesem Namen passt.
var ErrNichtGefunden = errors.New("keine Schriftdatei zu diesem Namen")

// Endungen sind die Schriftformate, die gelesen werden.
//
// `.ttc` ist eine **Sammlung** mehrerer Schriften in einer Datei; auf macOS
// liegt der halbe Grundbestand so, `Helvetica.ttc` unter ihnen. Gelesen wird
// daraus die erste Schnittform (`render.SchriftAusDatei`).
//
// **`.dfont` fehlt mit Absicht**: ein Ressourcenbündel aus der Zeit vor Mac OS X,
// das einen eigenen Weg bräuchte. Einen zu bauen, den niemand geprüft hat, wäre
// eine Zusage ohne Deckung — `docs/PORTIERUNG.md` führt die Lücke.
var Endungen = []string{".ttf", ".otf", ".ttc"}

// Suche findet Schriften in einer Liste von Verzeichnissen.
//
// **Nebenläufig benutzbar.** Der Rasterer läuft je Vorschau, und eine Vorschau
// wird abgebrochen, nicht abgewartet (§7) — es sind also immer mehrere
// unterwegs. Der erste Anlauf hatte den Zwischenspeicher ohne Schloss, und die
// laufende Anwendung stürzte mit „concurrent map writes" ab. Kein Test hat das
// gesehen: sie rufen nacheinander.
type Suche struct {
	orte []string

	// schloss deckt **jeden** Speicher dieser Suche — `gefunden`, `fehlt`,
	// `ersatz`, die Namensindizes und die Abdrücke.
	//
	// **Lesen und Schreiben getrennt** (PERF02, #2352): eine Karte fragt
	// hundertfach nach Schriften, die meisten schon gefunden oder schon
	// vergeblich gesucht. Unter einem einfachen Schloss warteten acht Arbeiter
	// der Problemprüfung ein Drittel ihrer Zeit aufeinander, und der Plan
	// skalierte nicht über das Dreifache. Geschrieben wird nur, wer wirklich
	// sucht.
	schloss sync.RWMutex

	// gefunden hält, was schon gelesen wurde.
	//
	// **Eine Karte fragt dieselbe Schrift für jede Textebene.** Ohne das läse
	// jede Vorschau die Datei mehrfach von der Platte.
	gefunden map[string]render.Glyphenschrift

	// gemerkteAbdruecke hält die Abdrücke der gefundenen Dateien (G2, #1459).
	//
	// **Neben `gefunden` und nicht darin**: jenes hält die gelesene Schrift,
	// dieses 32 Byte je Name. Ein gemeinsamer Speicher zwänge den einen Zweck
	// in die Grenze des anderen.
	gemerkteAbdruecke map[string]string

	// paketwurzeln sind die Verzeichnisse der bezogenen Pakete. Leer: keines.
	//
	// **Mehrere, seit dem 2026-08-28** (#1170):
	// Saat und bezogene Pakete liegen nebeneinander, und welche Wurzel die
	// Schriften trägt, weiß nur der Aufrufer. Eine einzige zu nehmen hieß, sich
	// für eine zu entscheiden — und der Kanonlauf entschied sich ein Jahr lang
	// für die falsche und zeichnete jede Karte ohne Text.
	paketwurzeln []string

	// standwurzeln sind **alle** genannten Paketwurzeln, auch die, die unter
	// einer früheren liegen und deshalb nicht eigens gegangen werden — ihr
	// Stand zählt trotzdem ([Suche.MitPaketen], #2005).
	standwurzeln []string

	// namen ordnet den **eigenen** Namen einer Schrift ihrer Datei zu.
	// `nil`, solange niemand danach gefragt hat — siehe [Suche.namensindex].
	namen map[string]string

	// paketnamen ist dasselbe für die bezogenen Pakete.
	paketnamen map[string]string

	// paketzeit ist der Stand der Wurzel, zu dem `paketnamen` gehört.
	//
	// **Ohne ihn stand ein leerer Index für die ganze Sitzung fest**, sobald
	// jemand vor dem Bezug einmal eine Karte geöffnet hatte — und seit #263
	// ist genau das der vorgesehene Weg (Zweitblick über Phase G).
	paketzeit string

	// fehlt hält die Namen, zu denen **nichts** gefunden wurde, mit dem
	// Fehler, den der Aufrufer bekam.
	//
	// **Eine fehlende Schrift kostete 49 ms je Textebene.** `paketdatei` geht
	// die Paketwurzel bei jedem Fehlschlag neu; beim installierten Pack sind
	// das 30.220 Dateien, und ohne Treffer läuft er den Baum ganz ab.
	// Gemessen am 2026-08-25 über 197 Blicke auf denselben fehlenden Namen —
	// so viele Textebenen haben zwanzig Karten aus `Black.mse-set`:
	//
	//	ohne diesen Speicher   9,907 s
	//	mit ihm                  398 ms
	//
	// **Über die Suche, wie die Anwendung sie baut**: mit den Schriftorten
	// dieses Rechners *und* der Paketwurzel. Der Faktor ist 25.
	//
	// **Und genau dann trifft es, wenn es am meisten weh tut**: wer die
	// Schriften nicht installiert hat, sieht ohnehin eine unfertige Karte —
	// und wartete zusätzlich am längsten auf sie.
	//
	// # Was der Speicher **nicht** einbringt
	//
	// Sind die Schriften da, ändert er nichts. Dasselbe Profil über zwanzig
	// Karten zählte **13 Läufe von `paketdatei` mit und ohne ihn**: lauter
	// erste Blicke, die schon `gefunden` abfängt. Der Lauf, den ein CPU-Profil
	// dort im Zeichenweg sieht, ist eine **einmalige** Anlaufkosten und keine
	// je Karte — über zwanzig Karten sieht das gleich aus und ist es nicht
	// (AGENTS.md §10, „die Reichweite einer Messung").
	//
	// Dasselbe hat [Suche.Ersatz] seit
	// #515:
	// dort heißt `nil` „keine gefunden", und auch das wird gemerkt. Hier
	// fehlte es eine Ebene höher.
	fehlt map[string]error

	// fehlstand ist der Stand, zu dem `fehlt` gehört — die Änderungszeit der
	// Paketwurzel **und die der Orte**.
	//
	// **Die Orte gehören dazu, und das hat gefehlt.** Der erste Anlauf band
	// nur an `paketstand()`; eine Schrift, die jemand während der Sitzung nach
	// `~/Library/Fonts` legt, ändert den nicht — und blieb bis zum Neustart
	// unauffindbar. Vorher fand [Suche.datei] sie, denn der Weg liest die Orte
	// bei jedem Versuch neu. Angestrichen vom Zweitblick über Phase Ω.
	fehlstand string

	// ersatz merkt sich je Zeichen die Schrift, die es führt — `nil` als Wert
	// heißt „keine gefunden", und auch das wird gemerkt
	// (#515).
	ersatz map[rune]render.Glyphenschrift
	// fettePfade nennt die Dateien beider Namensindizes, die ein fetter
	// Schnitt sind — der aufrechte schlägt sie beim selben Namen
	// ([Suche.eintragen]).
	fettePfade map[string]bool
}

// Neue baut eine Suche über den gegebenen Verzeichnissen.
func Neue(orte []string) *Suche {
	return &Suche{orte: orte, gefunden: map[string]render.Glyphenschrift{}, fettePfade: map[string]bool{}}
}

// Suche findet eine Schrift zu ihrem Namen.
//
// # Drei Wege, in dieser Reihenfolge
//
//  1. **Der Dateiname ohne Endung**, ohne Rücksicht auf Groß- und
//     Kleinschreibung — `MPlantin` findet `MPlantin.ttf` und `mplantin.otf`.
//     Billig, und trifft meistens.
//  2. **Die bezogenen Pakete**, ebenso über den Dateinamen.
//  3. **Der Name, den die Schrift über sich selbst führt** — dafür muss jede
//     Datei geöffnet werden, und deshalb steht er hinten.
//
// **Der dritte Weg ist nötig** (#240): `magic-m15.mse-style` verlangt
// `Beleren Bold`, und im bezogenen Pack heißt die Datei
// `beleren-bold_P1.01.ttf`. Ohne ihn trüge eine gezeichnete Karte keinen Text.
//
// **Was auch der dritte Weg nicht findet:** eine Schrift in einer `.ttc`- oder
// `.dfont`-Sammlung. Das ist eine benannte Lücke, und `docs/PORTIERUNG.md`
// führt sie.
func (s *Suche) Suche(name string, fett bool) (render.Schrift, error) {
	schluessel := name
	if fett {
		schluessel += fettzusatz
	}

	if schrift, da, fehler := s.gemerkt(schluessel); da {
		return schrift, fehler
	}

	s.schloss.Lock()
	defer s.schloss.Unlock()

	if schrift, da := s.gefunden[schluessel]; da {
		return schrift, nil
	}

	// **Ein Fehlschlag zählt genauso.** Er kostet den teuersten Weg von allen
	// — siehe das Feld `fehlt`.
	if fehler := s.fehlschlaege()[schluessel]; fehler != nil {
		return nil, fehler
	}

	ort, datei, err := s.schnitt(name, fett)
	if err != nil {
		s.fehlschlaege()[schluessel] = err

		return nil, err
	}

	// **Der Name kommt aus der Verzeichnisliste dieses Ortes**, nicht aus einer
	// Eingabe: `datei` gibt den Ort und den Eintrag, den es dort gesehen hat.
	// Ein Weg mit `..` kann so nicht entstehen.
	//
	// `os.Root` wäre die Zusage im Code statt im Kommentar — und brächte einen
	// Fehlerzweig mit, den kein Aufruf betreten kann: das Verzeichnis ist
	// gerade gelesen worden. Ein Zweig ohne Fall ist eine Behauptung darüber,
	// was im Ernstfall geschieht (AGENTS.md §3).
	roh, err := schriftLesen(filepath.Join(ort, datei))
	if err != nil {
		return nil, fmt.Errorf("die Schriftdatei %s ließ sich nicht lesen: %w", datei, err)
	}

	schrift, err := render.SchriftAusDatei(roh)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", filepath.Join(ort, datei), err)
	}

	s.gefunden[schluessel] = schrift

	return schrift, nil
}

// gemerkt antwortet aus dem Speicher, **unter dem Leseschloss** — eine
// gefundene Schrift oder ein Fehlschlag zum heutigen Stand. `da` heißt: die
// Antwort steht fest; sonst sucht [Suche.Suche] unter dem Schreibschloss.
func (s *Suche) gemerkt(schluessel string) (schrift render.Schrift, da bool, fehler error) {
	s.schloss.RLock()
	defer s.schloss.RUnlock()

	if gefunden, ist := s.gefunden[schluessel]; ist {
		return gefunden, true, nil
	}

	if s.fehlt == nil || s.fehlstand != s.paketstand()+"\x00"+s.ortestand() {
		return nil, false, nil
	}

	if fehlschlag := s.fehlt[schluessel]; fehlschlag != nil {
		return nil, true, fehlschlag
	}

	return nil, false, nil
}

// fehlschlaege gibt die gemerkten Fehlschläge zum **heutigen** Stand der
// Paketwurzel — und räumt sie, sobald der ein anderer ist.
//
// **Ein Eintrag ist nie `nil`**, deshalb genügt der Nullwert der Map als
// Antwort „nichts gemerkt". Ein zweiter Rückgabewert sagte dasselbe noch
// einmal.
//
// # Was das kostet
//
// Ein `os.Stat` je Frage statt eines Laufs über die ganze Wurzel. Dieselbe
// Bindung wie bei [Suche.paketnamensindex] — und **damit dieselbe Grenze**:
// eine Schriftdatei, die jemand während der Sitzung von Hand **in ein schon
// vorhandenes Paket** legt, ändert die Änderungszeit der Wurzel nicht und wird
// bis zum Neustart nicht gesehen. Ein **bezogenes** Paket legt ein
// Verzeichnis in der Wurzel an, ändert sie damit und wird weiterhin sofort
// gefunden — das ist der Weg aus
// #263, und
// `TestEinePaketschriftWirdAuchNachDemBezugGefunden` hält ihn.
//
// **Die Orte zählen mit.** Der Namensindex ([Suche.namensindex]) deckt nur den
// Weg über den **eigenen** Namen einer Schrift; der Weg über den **Dateinamen**
// liest die Orte bei jedem Versuch neu ([Suche.datei]) und fände eine frisch
// installierte Datei — ein Fehlschlag, der nur an die Paketwurzel gebunden
// wäre, verdeckte sie.
func (s *Suche) fehlschlaege() map[string]error {
	stand := s.paketstand() + "\x00" + s.ortestand()

	if s.fehlt == nil || s.fehlstand != stand {
		s.fehlt = map[string]error{}
		s.fehlstand = stand
	}

	return s.fehlt
}

// fettzusatz ist der Namenszusatz, an dem MSE den fetten Schnitt erkennt.
//
//	static const String BOLD_STRING = _(" Bold");
//
// `src/data/font.cpp`, Zeile 82, Fassung f9d9356 vom 2021-07-06.
const fettzusatz = " Bold"

// schnitt sucht die Datei zu Name **und Gewicht**.
//
// # Warum aus einem Gewicht ein Name wird
//
// MSE reicht Familienname und Gewicht an wxWidgets weiter, und das
// Betriebssystem sucht den Schnitt. Diese Suche kennt Namen; sie bildet MSEs
// eigene Regel nach — das abschließende ` Bold` gehört zum Schnitt, nicht zur
// Familie
// ([0023](../../docs/entscheidungen/0023-fett-wird-ueber-den-namen-gesucht.md)).
//
// **Ein Name, der schon auf ` Bold` endet, ist fett.** `magic-m15` schreibt
// `Beleren Bold`; MSE schneidet den Zusatz ab und setzt das Gewicht, und `<b>`
// ändert dort nichts mehr.
//
// **Ohne fetten Schnitt bleibt es beim aufrechten.** Dieselbe
// Bestenfalls-Arbeit wie bei der Kursiven (§8): ein aufrechter Text ist eine
// Karte, ein fehlender ist keine.
func (s *Suche) schnitt(name string, fett bool) (ort, datei string, fehler error) {
	familie, zusatz := strings.CutSuffix(strings.TrimSpace(name), fettzusatz)
	if !fett && !zusatz {
		return s.datei(name)
	}

	if ort, datei, err := s.datei(familie + fettzusatz); err == nil {
		return ort, datei, nil
	}

	return s.datei(familie)
}

// datei sucht die Datei zu einem Schriftnamen und gibt Ort und Dateinamen.
//
// **Getrennt**, weil das Lesen unter einer Wurzel geschieht: der Ort ist die
// Wurzel, der Name der Weg darin.
func (s *Suche) datei(name string) (ort, datei string, fehler error) {
	gesucht := strings.ToLower(strings.TrimSpace(name))
	if gesucht == "" {
		return "", "", fmt.Errorf("%w: der Name ist leer", ErrNichtGefunden)
	}

	for _, ort := range s.orte { //nolint:varnamelen // der Ort ist der Ort
		for _, datei := range schriftdateien(ort) {
			if strings.ToLower(ohneEndung(filepath.Base(datei))) == gesucht {
				return ort, datei, nil
			}
		}
	}

	// **Erst jetzt die Pakete.** Die Reihenfolge steht nicht in einem Kommentar,
	// sondern hier: was jemand selbst installiert hat, ist die Schrift, die er
	// meint (#140).
	if ort, datei, gefunden := s.paketdatei(gesucht); gefunden {
		return ort, datei, nil
	}

	// **Und zuletzt der Name, den die Schrift über sich selbst führt.**
	//
	// `Beleren Bold` liegt in `M15-Magic-Pack` als `beleren-bold_P1.01.ttf`;
	// kein Dateinamensvergleich findet das. Warum das hinten steht und was es
	// kostet, steht bei [Suche.namensindex]
	// (#240).
	if pfad, bekannt := s.namensindex()[gesucht]; bekannt {
		return filepath.Dir(pfad), filepath.Base(pfad), nil
	}

	if pfad, bekannt := s.paketnamensindex()[gesucht]; bekannt {
		return filepath.Dir(pfad), filepath.Base(pfad), nil
	}

	return "", "", fmt.Errorf("%w: %s (gesucht in %s)",
		ErrNichtGefunden, name, strings.Join(append(append([]string{}, s.orte...), s.paketwurzeln...), ", "))
}
