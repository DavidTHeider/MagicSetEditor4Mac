// Die `keyword:`-Blöcke eines Sets
// (#1062).
//
// # Der Befund
//
// Bis hierher fielen sie in den Rohweg: sie überlebten den Rundlauf Byte für
// Byte und waren im Bestand **unsichtbar** — keine Suche fand sie, keine Seite
// zeigte sie, und pflegen ließ sich keines.
//
// # Warum sie nicht bei den Spielschlüsselwörtern stehen
//
// `msepaket.Schluesselwoerter` liest den Kopf eines **Pakets** und gibt
// `msepaket.Schluesselwort`. Ein Set führt dieselben fünf Felder, aber im
// Bestand: mit Kennung, anlegbar, löschbar. Das ist [domain.Keyword], und die
// Übersetzung steht hier — an derselben Stelle, an der auch Karten und
// Setangaben aus dem Baum kommen.

package service

import (
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// schluesselwortAus macht aus einem `keyword:`-Block ein Schlüsselwort.
//
// **Fünf Felder, und mehr steht nicht in der Datei** —
// `IMPLEMENT_REFLECTION(Keyword)`, `src/data/keyword.cpp`, Zeilen 86–93,
// Fassung f9d9356 vom 2021-07-06. Der Ausdruck, die Parameterliste und
// `valid` werden beim Vorbereiten **gerechnet**.
//
// **Ein leeres `match` wird nicht abgewiesen.** MSE legt ein neues
// Schlüsselwort genauso an — leer in jedem Feld — und schaltet es beim
// Vorbereiten ab (`Keyword::prepare`, Zeilen 205–209). Wer es hier verwürfe,
// verlöre beim Speichern einen Block, den jemand angefangen hat.
func schluesselwortAus(knoten *msepaket.Knoten) *domain.Keyword {
	return &domain.Keyword{
		Word:     msepaket.WertOderBlock(knoten.Ersten("keyword")),
		Match:    msepaket.WertOderBlock(knoten.Ersten("match")),
		Reminder: msepaket.WertOderBlock(knoten.Ersten("reminder")),
		Rules:    msepaket.WertOderBlock(knoten.Ersten("rules")),
		Mode:     msepaket.WertOderBlock(knoten.Ersten("mode")),
	}
}
