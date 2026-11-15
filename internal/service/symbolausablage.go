// **Die Symbolquelle, die auch in die Ablage sieht** —
// #890,
// Werkstück 4 von Phase X.
//
// # Warum es sie braucht
//
// Eine gezeichnete `.mse-symbol` liegt im **Bildverzeichnis** des Sets — sie
// kommt erst beim Sichern ins Paket ([0029](../../docs/entscheidungen/0029-bilder-liegen-neben-dem-set.md)).
// Die Symbolquelle des Planbauers kennt nur Pakete; ein Verweis auf eine
// Datei, die es dort noch nicht gibt, gäbe eine leere Gestalt, und die Karte
// zeigte MSEs Standardquadrat.
//
// **Das ist derselbe Weg, den ein eingesetztes Bild schon nimmt**:
// `Vorlage.Bilderverzeichnis` steht neben `Vorlage.Setpaket`, und `bildebene`
// bekommt beide. Der Symbolweg bekam bis hierher nur das Paket.

package service

import (
	"path/filepath"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// SymbolquelleAusAblage sieht zuerst in der Ablage nach und fragt sonst weiter.
//
// **Zuerst die Ablage**, denn was jemand gerade gezeichnet hat, ist neuer als
// das, was im Paket steht. Dieselbe Reihenfolge wie bei den Bildern.
//
// **Und der Rückfall qualifiziert den Verweis.** `set.symbol` trägt in einer
// echten Datei einen **schlichten** Namen — `symbol1.mse-symbol` —, und die
// Datei liegt im `.mse-set` daneben; die Paketquelle verlangt aber einen
// Setpaketverweis oder einen Pfad mit führendem `/`. Ohne diesen Schritt zeigte
// ein **importiertes** Set sein Symbol nicht, und nur ein hier gezeichnetes kam
// an — angestrichen vom Zweitblick über Phase X.
//
// **Ohne Ort und ohne Quelle dahinter gibt sie leer zurück** — beides kommt
// vor: eine Anwendung ohne Benutzerverzeichnis, ein Ausfuhrlauf ohne
// Bildquelle. Ein Fehler hier nähme der Karte alles andere mit (§8).
func SymbolquelleAusAblage(ort string, dahinter Symbolquelle, setpaket string) Symbolquelle {
	return ablagensymbolquelle{ort: ort, dahinter: dahinter, setpaket: setpaket}
}

// ablagensymbolquelle hält Ort, Rückfallquelle und den Ort des Setpakets.
type ablagensymbolquelle struct {
	ort      string
	dahinter Symbolquelle
	setpaket string
}

func (a ablagensymbolquelle) Symbolgestalt(verweis string) msepaket.Symbolzeichnung {
	if zeichnung, gefunden := a.ausDerAblage(verweis); gefunden {
		return zeichnung
	}

	if a.dahinter == nil {
		return msepaket.Symbolzeichnung{}
	}

	return a.dahinter.Symbolgestalt(a.qualifiziert(verweis))
}

// qualifiziert hängt den Ort des Setpakets an einen schlichten Namen.
//
// **Dieselbe Regel wie in der Ausfuhr** ([setsymbolquelle]): ein Verweis, der
// seinen Ort schon trägt, bleibt; ein Pfad mit führendem `/` meint ein anderes
// Paket. Nur der schlichte Name bekommt den Ort des Sets.
func (a ablagensymbolquelle) qualifiziert(verweis string) string {
	name := strings.TrimSpace(verweis)

	if a.setpaket == "" || name == "" || strings.HasPrefix(name, "/") {
		return verweis
	}

	if _, _, ist := msepaket.ZerlegeSetpaketverweis(name); ist {
		return verweis
	}

	return msepaket.SetpaketVerweis(a.setpaket, name)
}

// ausDerAblage liest eine Datei, die neben dem Set liegt.
//
// **Nur ein schlichter Name.** Ein Verweis mit einem Schrägstrich meint ein
// Paket — `/spiel.mse-game/symbol` —, und ein `..` führt aus dem Verzeichnis
// heraus. Beides ist hier nichts zu suchen; der Name kommt aus einem Feldwert,
// den eine fremde Datei trägt (§8).
func (a ablagensymbolquelle) ausDerAblage(verweis string) (msepaket.Symbolzeichnung, bool) {
	name := strings.TrimSpace(verweis)
	if a.ort == "" || name == "" || name != filepath.Base(name) || name == ".." {
		return msepaket.Symbolzeichnung{}, false
	}

	zeichnung, _, err := msepaket.LiesSymbolzeichnung(filepath.Join(a.ort, name))
	if err != nil {
		return msepaket.Symbolzeichnung{}, false
	}

	return zeichnung, true
}
