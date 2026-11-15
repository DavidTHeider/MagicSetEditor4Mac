package service

import (
	"path/filepath"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// Symbolschriften lädt die `.mse-symbol-font`-Pakete, die ein Stil nennt.
//
// # Warum vorab und nicht auf Zuruf
//
// Der Planbauer öffnet keine Pakete — er bekommt ein Modell und rechnet
// (AGENTS.md §2). Wer ihm eine Paketwurzel hineinreichte, machte aus ihm einen
// zweiten Paketleser, und die Naht zwischen `msepaket` und `service` wäre weg.
//
// # Jede Schrift einmal
//
// `magic-m15.mse-style` nennt `magic-mana-large` an mehreren Feldern. Ein
// Paket je Feld zu öffnen kostete das Mehrfache und brächte dasselbe Modell.
//
// # Eine fehlende Schrift ist ein Befund
//
// Nicht ein Abbruch (§8): ein Stil, dessen Symbolschrift auf diesem Rechner
// fehlt, zeichnet alles andere. Der Befund entsteht dort, wo er sichtbar wird —
// am Feld, in `symbolebenen`, und nicht hier: hier ist noch nicht bekannt, ob
// die Karte das Feld überhaupt füllt.
func (s *Stilspeicher) Symbolschriften(
	stil msepaket.Stil,
) (map[string]msepaket.Symbolschrift, []domain.Befund) {
	gebraucht := symbolschriftnamen(stil)
	if len(gebraucht) == 0 {
		return nil, nil
	}

	schriften := make(map[string]msepaket.Symbolschrift, len(gebraucht))

	var befunde []domain.Befund

	for _, name := range gebraucht {
		schrift, weitere, gefunden := s.eineSymbolschrift(name)
		befunde = append(befunde, weitere...)

		if gefunden {
			schriften[name] = schrift
		}
	}

	return schriften, befunde
}

// Symbolschrift öffnet eine Symbolschrift nach ihrem Namen — ohne Endung.
//
// **Für die Ausfuhr** (#658): `symbols_to_html` nennt `magic-mana-small`, und
// dort steht kein Stil daneben, aus dem sich die Schrift ableiten ließe.
//
// **Was hier verlorengeht, steht dabei** (§10): die Befunde des Öffnens — ein
// nicht einbindbares Modul, ein unlesbarer `symbol:`-Block. Der Ausfuhrlauf
// hat keinen Kanal dafür; was er hat, ist ein Fehler, wenn die Schrift **gar
// nicht** dasteht, und der ist der Fall, an dem eine Vorlage scheitert. Ein
// halb gelesenes Symbolschriftpaket zeichnet die Symbole, die es trägt.
func (s *Stilspeicher) Symbolschrift(name string) (msepaket.Symbolschrift, bool) {
	schrift, _, gefunden := s.eineSymbolschrift(name)

	return schrift, gefunden
}

// eineSymbolschrift öffnet ein Symbolschriftpaket unter den Wurzeln.
func (s *Stilspeicher) eineSymbolschrift(
	name string,
) (schrift msepaket.Symbolschrift, befunde []domain.Befund, gefunden bool) {
	dateiname := name + "." + string(msepaket.ArtSymbolschrift)

	for _, wurzel := range s.wurzeln {
		paket, err := msepaket.Oeffne(filepath.Join(wurzel, dateiname))
		if err != nil {
			continue
		}

		// **Erst die Module.** Ein Symbolschriftpaket bindet sie ein wie jedes
		// andere — `magic-mana-large` nennt `magic-modules.mse-include`.
		befunde = append(befunde,
			ausMsepaket(paket.ModuleEinbinden(msepaket.NeuerPaketleser(s.wurzeln...)))...)

		gelesen, modellbefunde := paket.Symbolmodell()
		befunde = append(befunde, ausMsepaket(modellbefunde)...)
		_ = paket.Schliessen() //nolint:errcheck // das Modell ist vollständig kopiert

		return gelesen, befunde, true
	}

	// **Kein Befund hier.** Ob das eine Lücke ist, entscheidet die Karte: ein
	// Feld, das leer bleibt, braucht seine Symbolschrift nicht. Gemeldet wird
	// in `symbolebenen`, wo der Wert dasteht.
	return msepaket.Symbolschrift{}, befunde, false
}
