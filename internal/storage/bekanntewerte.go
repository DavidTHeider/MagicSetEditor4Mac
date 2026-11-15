package storage

import (
	"context"
	"slices"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// MitBekanntenWerten gibt einen Kontext, in dem ein Feld nur geschrieben wird,
// wenn es noch einen dieser Werte trägt — U02b
// (#2179).
//
// **Dieselbe Gestalt wie [MitProfil]**: die Fassade des Fensters setzt ihn, und
// jeder Weg darunter reicht den Kontext ohnehin durch. MCP, CLI und REST
// setzen ihn nicht und schreiben ohne diese Bedingung — **den Stand, aus dem
// ein Wert gerechnet ist, hält aber jeder Weg** ([Kartenschreiber.FelderSetzenVonAussen],
// Gesamtreview über 0.9.0, Befund 2).
func MitBekanntenWerten(ctx context.Context, werte []string) context.Context {
	return context.WithValue(ctx, bekanntschluessel{}, slices.Clone(werte))
}

// bekanntschluessel ist der Schlüssel der bekannten Werte im Kontext.
type bekanntschluessel struct{}

// keineBedingung steht im Kontext eines nachgezogenen Werts an der Stelle
// der bekannten — die Prüfung in [fremderWert] erkennt dann keine Liste.
type keineBedingung struct{}

// ohneBekannte gibt den Kontext ohne Bedingung — für einen Wert, den das
// Wertskript eines anderen Felds rechnete ([domain.Value.Nachgezogen]): das
// Fenster kannte ihn nicht und konnte ihn nicht kennen.
func ohneBekannte(ctx context.Context) context.Context {
	return context.WithValue(ctx, bekanntschluessel{}, keineBedingung{})
}

// fremderWert prüft den gespeicherten Wert eines Feldes gegen die bekannten —
// **in der Transaktion, die schreibt**, nach dem Lesen und vor dem Schreiben.
// Ohne Bedingung im Kontext gibt es nichts zu prüfen.
//
// **Ein fehlendes Feld zählt als leer**: das Fenster zeigt es als leeres Feld
// und kennt es deshalb als `""`.
func fremderWert(ctx context.Context, feld, gespeichert string) error {
	bekannt, bedingt := ctx.Value(bekanntschluessel{}).([]string)
	if !bedingt || slices.Contains(bekannt, gespeichert) {
		return nil
	}

	return domain.NewFieldError(feld, domain.SchluesselFremderWert,
		"In %q steht inzwischen ein anderer Wert — jemand anderes hat ihn geändert.", feld)
}

// feldwert liest ein Feld aus dem Stand eines Schlüsselworts — das Gegenstück
// zu [mitFeld].
func feldwert(stand domain.Schluesselwortstand, feld string) string {
	switch feld {
	case "wort":
		return stand.Wort
	case "muster":
		return stand.Muster
	case "erinnerung":
		return stand.Erinnerung
	case "regeln":
		return stand.Regeln
	default:
		return stand.Modus
	}
}
