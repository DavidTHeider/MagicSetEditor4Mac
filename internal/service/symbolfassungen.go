package service

import (
	"context"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **Eine Symbolschrift führt denselben Code mehrfach.**
//
// `magic-mana-large` trägt `T` und `Q` je dreimal, dazu die Hybridsymbole in
// je zwei Fassungen — insgesamt 54 `enabled:`-Bedingungen über 167 Symbole.
// MSE geht die Liste in **Dateireihenfolge** durch und nimmt den ersten
// Eintrag, dessen Bedingung erfüllt ist und dessen Code am Anfang des Restes
// steht (`SymbolFont::split`, Zeilen 191–230, Fassung f9d9356 vom 2021-07-06).
//
// Bis zum 2026-08-17 nahm dieses Programm den ersten Eintrag überhaupt und
// zeichnete damit `mana_t_older.png`, wo MSE `mana_t.png` zeichnet
// (#293).

// aktiveSymbolschriften behält je Schrift die Fassungen, die hier gelten.
//
// # Warum das einmal je Karte geschieht und nicht je Symbol
//
// Die Bedingungen fragen nach Funktionen, die das Initskript des Stils anlegt
// — `mana_t()`, `guild_mana()`, `use_small_v_mana()`. Sie hängen an der Karte,
// nicht an der Stelle im Text; einmal ausgewertet gilt die Antwort für jedes
// Feld. Und was danach übrig ist, ist eine Symbolschrift wie jede andere:
// jeder Code kommt einmal vor, und der erste Treffer ist der richtige.
//
// # Ausdruckscodes bleiben stehen
//
// 32-mal in `magic-mana-large`, darunter der letzte Eintrag der Datei als
// Auffangausdruck. Sie werden **nicht** hier aussortiert, sondern beim Zerlegen
// als Ausdruck geprüft (`symbolmuster.go`) — ohne sie hätte die Schrift für
// eine schlichte `1` kein Symbol, denn die Ziffern selbst hängen dort an
// `enabled: {ancestral_mana()}`.
func aktiveSymbolschriften(
	ctx context.Context, schriften map[string]msepaket.Symbolschrift,
	umgebung *msescript.Umgebung,
) (map[string]msepaket.Symbolschrift, []domain.Befund) {
	if len(schriften) == 0 {
		return schriften, nil
	}

	aktive := make(map[string]msepaket.Symbolschrift, len(schriften))

	var befunde []domain.Befund

	for name, schrift := range schriften {
		gewaehlt := make([]msepaket.Symbol, 0, len(schrift.Symbole))

		for i := range schrift.Symbole {
			symbol := schrift.Symbole[i]

			gilt, meldung := bedingungGilt(ctx, symbol.Bedingung, umgebung)
			if meldung != "" {
				befunde = append(befunde, domain.Befund{
					Ort: "Symbolschrift " + name,
					Meldung: "die Bedingung des Symbols " + symbol.Code +
						" ließ sich nicht ausrechnen: " + meldung +
						"; die Fassung gilt",
				})
			}

			if gilt {
				gewaehlt = append(gewaehlt, symbol)
			}
		}

		schrift.Symbole = gewaehlt
		aktive[name] = schrift
	}

	return aktive, entdoppelt(befunde)
}

// bedingungGilt wertet ein `enabled:` aus.
//
// **Ohne Bedingung gilt die Fassung.** Und **eine, die scheitert, gilt
// ebenfalls**: MSE führt `enabled` als `Scriptable<bool>` mit dem Vorgabewert
// `true`, und eine gescheiterte Auswertung lässt den alten Wert stehen
// (`SymbolInFont::SymbolInFont`, Zeile 103). Die andere Wahl — im Zweifel
// wegwerfen — nähme einem Code womöglich seine einzige Fassung und ließe die
// Karte ohne Symbol dastehen.
//
// Der zweite Rückgabewert ist die Meldung, wenn es schiefging.
func bedingungGilt(
	ctx context.Context, bedingung string, umgebung *msescript.Umgebung,
) (gilt bool, meldung string) {
	roh := strings.TrimSpace(bedingung)
	if roh == "" {
		return true, ""
	}

	rumpf, geklammert := skriptrumpf(roh)
	if !geklammert {
		// **Eine Bedingung ohne Klammern ist ein Wert, kein Ausdruck.** MSE
		// liest `enabled: false` genauso wie `enabled: { false }`.
		rumpf = roh
	}

	wert, err := msescript.AuswertenQuelle(ctx, rumpf, umgebung)
	if err != nil {
		return true, err.Error()
	}

	gilt, brauchbar := wert.AlsWahrheit()
	if !brauchbar {
		return true, "der Wert " + wert.AlsText() + " ist kein Wahrheitswert"
	}

	return gilt, ""
}

// entdoppelt wirft gleichlautende Befunde weg.
//
// **Weil eine Schrift 54 Bedingungen trägt und sie alle dieselbe Funktion
// fragen.** Fehlt `guild_mana`, stünde derselbe Satz 18-mal da — und eine
// Liste, in der ein Befund 18-mal steht, liest niemand zu Ende.
func entdoppelt(befunde []domain.Befund) []domain.Befund {
	if len(befunde) < 2 {
		return befunde
	}

	gesehen := make(map[string]bool, len(befunde))
	einmalig := make([]domain.Befund, 0, len(befunde))

	for _, befund := range befunde {
		schluessel := befund.Ort + "\x00" + befund.Meldung
		if gesehen[schluessel] {
			continue
		}

		gesehen[schluessel] = true

		einmalig = append(einmalig, befund)
	}

	return einmalig
}
