// **Ein Set, das in dieser Anwendung entstanden ist, hat keine Quelldatei** —
// Befund **F02** der nativen Prüfung
// (#1746).
//
// # Warum es diese Datei gibt
//
// [0011](../../docs/entscheidungen/0011-die-quelldatei-wird-mitgefuehrt.md)
// sagt es seit dem 2026-08-15: *„Ein Set, das in der Anwendung entsteht, hat
// keine. Dann wird zusammengesetzt, wie `msepaket.Schreibe` es für Knoten ohne
// Rohzeile tut. Die Byte-Zusage gilt dort nicht — es gibt nichts, womit zu
// vergleichen wäre."*
//
// **Umgesetzt war das nie.** `Setschreiber.Schreibe` wies ein solches Set mit
// `ErrNotFound` ab, und ein Lauf hielt die Abweisung fest — begründet mit
// derselben Byte-Zusage, die hier gar nicht auf dem Spiel steht. Was das
// gekostet hat, steht in der nativen Prüfung von v0.5.0: **⌘S und „Sichern
// unter…" meldeten beide „gesuchten Eintrag nicht gefunden", und es entstand
// keine Datei.**
//
// # Warum ein Gerüst und keine fertige Datei
//
// **Weil die drei Angleicher schon da sind.** `angleichen`,
// `setwerteAngleichen` und `schluesselwoerterAngleichen` bringen einen Baum auf
// den Stand des Bestands — geschrieben für den Fall „es gab schon eine Datei".
// Sie tun dasselbe an einem leeren Gerüst, und zwei Wege, die eine Datei
// erzeugen, liefen auseinander.
//
// Deshalb steht hier **nur**, was ein Gerüst ausmacht: die Kopfzeilen und ein
// leerer `set info:`-Block. Alles andere kommt von dort, wo es auch sonst
// herkommt.

package service

import (
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// standardfassung ist die MSE-Fassung, die ein frisches Set trägt.
//
// **Keine Erfindung, sondern eine Aussage über das, was hier geschrieben
// wird.** MSE2 setzt die Zeile bei jedem Speichern (`src/util/io/writer.cpp`,
// Zeile 26, Fassung f9d9356 vom 2021-07-06); sie sagt, nach welchem
// Formatstand die Datei aufgebaut ist, und nicht, wer sie erzeugt hat.
//
// **Sie greift nur, wo der Bestand nichts führt.** Ein importiertes Set bringt
// seine eigene mit; die Spalte `mse_fassung` steht seit Migration 0032 da und
// ist bei einem selbst angelegten Set leer.
const standardfassung = "2.0.0"

// setgeruest baut den Baum eines Sets, das aus keiner Datei stammt.
//
// **Die Reihenfolge ist die aus `docs/FORMATE.md`** und nicht eine beliebige:
// `mse version` zuerst, dann `game`, dann `stylesheet`. MSE2 liest zwar nach
// Schlüssel; ein Mensch, der zwei Dateien nebeneinanderlegt, liest nach Lage.
func setgeruest(zeile domain.SetHead, format domain.Setkopf) (*msepaket.Baum, error) {
	// **Ohne Spiel entsteht nichts.** Ein Paket ohne `game:` öffnet in MSE2
	// nicht, und eines zu erfinden wäre eine Falschangabe — dieselbe
	// Überlegung wie bei [domain.Set.Origin]: leer heißt leer.
	if zeile.Game == "" {
		return nil, fmt.Errorf(
			"%w: das Set %d nennt kein Spiel — ohne eines ist kein Paket zu schreiben",
			domain.ErrValidation, zeile.ID)
	}

	fassung := format.MSEFassung
	if fassung == "" {
		fassung = standardfassung
	}

	wurzel := []*msepaket.Knoten{
		kopfzeile("mse version", fassung),
		kopfzeile("game", zeile.Game),
	}

	// **Ein Stil ist nicht zwingend, und deshalb steht er nur da, wenn es ihn
	// gibt.** `Importeur.Neu` verlangt ihn; ein Set aus einer anderen Quelle
	// könnte ohne dastehen, und eine leere `stylesheet:`-Zeile ist schlechter
	// als keine — MSE2 suchte dann ein Paket mit leerem Namen.
	if zeile.Style != "" {
		wurzel = append(wurzel, kopfzeile("stylesheet", zeile.Style))
	}

	wurzel = append(wurzel, setzusaetze(format)...)

	// **Der Block steht leer da, und das ist kein halber Zustand.**
	// `setwerteAngleichen` gibt zurück, ohne etwas zu tun, wenn es ihn nicht
	// gibt — der Titel eines frisch angelegten Sets stünde dann nirgends.
	wurzel = append(wurzel, &msepaket.Knoten{
		Zeilenart:  msepaket.ArtEintrag,
		Schluessel: "set info",
		Zeilenende: "\n",
	})

	return &msepaket.Baum{Datei: "set", Wurzel: wurzel}, nil
}

// setzusaetze sind die Kopfangaben, die nur dastehen, wenn sie etwas sagen.
//
// **Leer heißt: nicht geschrieben.** In jedem kanonischen Set steht
// `apprentice code` leer und `version control` auf `none`; eine Zeile, die
// nichts aussagt, in eine frische Datei zu setzen hieße, dem Set eine Angabe
// anzudichten, die niemand gemacht hat.
func setzusaetze(format domain.Setkopf) []*msepaket.Knoten {
	var zusaetze []*msepaket.Knoten

	for _, angabe := range []struct{ schluessel, wert string }{
		{"game version", format.Spielfassung},
		{"stylesheet version", format.Stilfassung},
		{"version control", format.Versionskontrolle},
		{"apprentice code", format.Apprenticecode},
	} {
		if angabe.wert != "" {
			zusaetze = append(zusaetze, kopfzeile(angabe.schluessel, angabe.wert))
		}
	}

	return zusaetze
}

// kopfzeile baut einen Wurzelknoten ohne Rohfassung.
//
// **Ohne `Roh`, und das ist der Punkt**: `msepaket.Schreibe` setzt eine Zeile
// aus Schlüssel und Wert zusammen, sobald keine Rohfassung dasteht. Eine
// erfundene Rohzeile wäre eine Behauptung darüber, wie die Datei einmal
// aussah — und es gab sie nie.
func kopfzeile(schluessel, wert string) *msepaket.Knoten {
	return &msepaket.Knoten{
		Zeilenart:  msepaket.ArtEintrag,
		Schluessel: schluessel,
		Wert:       wert,
		Zeilenende: "\n",
	}
}
