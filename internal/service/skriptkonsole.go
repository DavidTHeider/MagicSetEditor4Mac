// Die Skriptkonsole
// (#1070).
//
// # Warum es sie gibt
//
// MSE-Script ist vollständig portiert — und niemand konnte einen Ausdruck
// eingeben. §15 nennt MSEs eigene Kommandozeile ausdrücklich als **Orakel**
// dieses Projekts (`magicseteditor --cli <SET>.mse-set -q --raw`); eine
// eigene Konsole beantwortet dieselben Fragen über **dieses** Programm und
// macht die beiden vergleichbar.
//
// # Was sie nicht ist
//
// **Kein Schreibweg.** Sie wertet aus und gibt zurück; sie ändert keine Karte
// und kein Set. Die Sprache selbst kann es nicht — `internal/schichten` weist
// jeden Zugriff auf das Dateisystem aus `msescript` ab —, und diese Naht legt
// nichts dazu.

package service

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// Konsolenergebnis ist, was ein ausgewerteter Ausdruck zurückgibt.
type Konsolenergebnis struct {
	// Quelltext ist das Ergebnis als **Quelltext** — `"a"` und nicht `a`.
	//
	// So zeigt MSE es (`src/cli/cli_main.cpp`, Zeile 234, Fassung f9d9356 vom
	// 2021-07-06), und der Unterschied ist der Punkt: wer eine Konsole
	// benutzt, will eine Zeichenkette von einem Namen unterscheiden.
	Quelltext string

	// Text ist dasselbe Ergebnis in seiner gewöhnlichen Form — für die, die
	// es weiterverwenden statt es zu lesen.
	Text string

	// Art ist die Art des Werts, für eine Oberfläche, die sie anzeigen will.
	Art string

	// Befunde sind die Vorbehalte beim Aufbau der Umgebung — ein fehlendes
	// Stilpaket etwa. **Sie halten die Auswertung nicht auf** (§8): ein
	// Ausdruck, der die Vorlage nicht braucht, läuft trotzdem.
	Befunde []domain.Befund

	// Eintraege sind die Einträge einer Abbildung, rekursiv — leer für jeden
	// anderen Wert. Das Fenster zeigt sie als Baum (Werkstück U13 von Phase
	// 5.1, #1424).
	Eintraege []Konsoleneintrag
}

// Skriptkonsole wertet Ausdrücke im Kontext eines Sets aus.
type Skriptkonsole struct {
	karten Kartensetleser
	planer *Planbauer
}

// NeueSkriptkonsole baut sie.
//
// **Über den Planbauer und nicht daneben**: die Umgebung eines Ausdrucks ist
// dieselbe, in der eine Karte gezeichnet wird. Eine zweite liefe auseinander,
// und dann beantwortete die Konsole Fragen über ein Programm, das es nicht
// gibt.
func NeueSkriptkonsole(karten Kartensetleser, planer *Planbauer) *Skriptkonsole {
	return &Skriptkonsole{karten: karten, planer: planer}
}

// Auswerten rechnet einen Ausdruck im Kontext einer Karte aus.
//
// # Die Reihenfolge
//
// **Erst die Umgebung, dann der Ausdruck.** MSE zerteilt zuerst und wertet
// danach aus (`cli_main.cpp`, Zeilen 221–234); hier steht die Umgebung
// zuerst, weil sie die Befunde trägt — ein Ausdruck, der an einem fehlenden
// Stilpaket scheitert, soll sagen können, dass es fehlt.
func (k *Skriptkonsole) Auswerten(
	ctx context.Context, karteID int64, ausdruck string,
) (*Konsolenergebnis, error) {
	karte, err := k.karten.KarteLesen(ctx, karteID)
	if err != nil {
		return nil, err
	}

	umgebung, befunde := k.umgebung(ctx, karte)

	wert, fehler := msescript.AuswertenQuelle(ctx, ausdruck, umgebung)
	if fehler != nil {
		// **Umschlossen und nicht geglättet** (#1164). Hier stand `%s`, und
		// damit war die **Stelle** fort, die der Zerteiler seit je meldet:
		// aus „Zeile 2, Spalte 3: hier fehlt ein Wert" wurde ein Satz, in dem
		// zwei Zahlen stehen und den niemand auseinandernehmen kann.
		//
		// **Zwei `%w`** (§8): `ErrValidation` sagt der Fassade, welcher
		// Ausgang es ist, `fehler` trägt die Stelle. `errors.As` findet beide.
		return nil, fmt.Errorf("%w: %w", domain.ErrValidation, fehler)
	}

	return &Konsolenergebnis{
		Quelltext: wert.AlsQuelltext(),
		Text:      wert.AlsText(),
		Art:       string(wert.Art),
		Befunde:   befunde,
		Eintraege: eintraegeVon(wert),
	}, nil
}

// umgebung baut die Umgebung der Karte — mit Vorlage, wenn es eine gibt.
//
// **Ohne Stilpaket bleibt die Bibliothek.** Ein Ausdruck wie `1 + 1` braucht
// keine Vorlage, und ihn abzuweisen, weil ein Paket fehlt, hieße eine Konsole
// zu bauen, die genau dann schweigt, wenn man sie am nötigsten braucht.
func (k *Skriptkonsole) umgebung(ctx context.Context, karte *domain.CardRow) (
	*msescript.Umgebung, []domain.Befund,
) {
	if k.planer == nil {
		return msescript.Bibliothek().Darueber(), befundliste(
			"für die Konsole ist kein Planbauer eingerichtet — " +
				"`card` und `set` stehen nicht zur Verfügung")
	}

	return k.planer.Kartenumgebung(ctx, karte)
}

// Namen gibt, was ein Ausdruck an dieser Karte nennen darf.
//
// # Wofür
//
// Vervollständigung und die Liste in der Konsole — beides Werkstücke von
// Phase 3.3 (#1164).
//
// **Aus der wirklichen Umgebung und nicht aus einer Aufzählung daneben.** Das
// Issue sagt es: „was die Umgebung wirklich kennt, nicht was plausibel
// klingt". Eine gepflegte Liste liefe an dem Tag auseinander, an dem die
// Bibliothek eine Funktion dazubekommt — und die Vervollständigung böte etwas
// an, das nicht läuft.
//
// **Ohne die Wörter des Zerteilers.** `if`, `then`, `else` sind keine Namen
// der Umgebung; wer sie von hier nähme, behauptete, sie stünden darin.
//
// **Die Befunde halten sie nicht auf** (§8): ohne Stilpaket bleibt die
// Bibliothek, und die ist das meiste.
func (k *Skriptkonsole) Namen(ctx context.Context, karteID int64) ([]string, error) {
	karte, err := k.karten.KarteLesen(ctx, karteID)
	if err != nil {
		return nil, err
	}

	umgebung, _ := k.umgebung(ctx, karte)

	return umgebung.AlleNamen(), nil
}
