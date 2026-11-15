package service

import (
	"context"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// wahlbild sucht das Bild, das zum Wert eines Feldes gehört.
//
// # Der Mechanismus, nachgelesen in MSE2
//
// `src/data/field/choice.cpp:171` (Fassung f9d9356 vom 2021-07-06) baut aus
//
//	choice images:
//	     a: {uvw}
//	     b: {xyz}
//
// das Skript
//
//	[a: {uvw}, b: {xyz}][input]() or else nil
//
// — eine Abbildung von Wahlname auf Bildskript, indiziert mit dem **Wert des
// Feldes**, gerufen, und `or else nil`.
//
// # Jeder Name steht zweimal
//
// Einmal wie geschrieben und einmal in der **unkanonischen** Form, in der
// Unterstriche zu Leerzeichen werden (`util/string.cpp:127`). Der Stil schreibt
// `diagonal T`, ein Kartenwert kann `diagonal_T` heißen.
//
// # Ein Wert ohne Wahl zeichnet nichts — und meldet nichts
//
// **`or else nil`** ist der letzte Schritt des Skripts, das MSE baut: ein Wert,
// für den keine Wahl hinterlegt ist, ergibt `nil`, und die Ebene bleibt leer.
// **Das ist kein Mangel, sondern die Bauform** — `card symbol: none` hat kein
// Bild, weil es keines haben soll
// (#270).
func wahlbild(
	ctx context.Context, feld *msepaket.Feldstil, wert string,
	umgebung *msescript.Umgebung,
) (msescript.Bildbaum, []domain.Befund) {
	// **Kein Zweig für „gar keine Wahlbilder".** Der Aufrufer prüft es, bevor
	// er hier hereinkommt; ein Weg, den kein Aufruf betreten kann, ist einer zu
	// viel (§3).
	//
	// **Ein leeres Feld ist dagegen ein Fall, und kein Mangel.** MSE gibt dort `nil` und zeichnet
	// nichts; eine Karte ohne Seltenheit hat eben kein Seltenheitssymbol. Der
	// erste Anlauf meldete auch das und trieb die Befundzahl über die 190 Stile
	// von 1.668 auf 1.797 — eine Meldung, die für jedes ungefüllte Feld
	// erscheint, liest niemand mehr.
	gesucht := unkanonisch(wert)
	if gesucht == "" {
		return msescript.Bildbaum{}, nil
	}

	for stelle := range feld.Wahlbilder {
		gewaehlt := &feld.Wahlbilder[stelle]
		if unkanonisch(gewaehlt.Wahl) != gesucht {
			continue
		}

		// **Der Wert darf ein Skript sein**, so wie jede andere Bildangabe:
		// MSE ruft ihn (`I_CALL, 0`).
		baum, befunde := bildangabe(ctx, feld.Feld, "choice images", gewaehlt.Bild, umgebung)

		return baum, befunde
	}

	return msescript.Bildbaum{}, nil
}

// unkanonisch macht aus Unterstrichen Leerzeichen — `uncanonical_name_form`.
func unkanonisch(name string) string {
	return strings.ReplaceAll(strings.TrimSpace(name), "_", " ")
}
