// **Die Angaben des Sets gehen zurück in die Datei** —
// #890.
//
// # Der Befund
//
// Der Setschreiber glich **Karten** und **Schlüsselwörter** ab — die Angaben
// des Sets selbst nicht. `SetwertSetzen` schrieb in den Bestand, und beim
// Sichern blieb in der Datei stehen, was vorher dort stand: Titel,
// Urheberrecht, Setkürzel, **und das Symbol**.
//
// **Gefunden hat es die Abnahme von Phase X**, nicht ein Test: ein
// gezeichnetes Symbol wurde abgelegt, als `set.symbol` eingetragen, das Set
// gesichert — und die Datei trug `symbol:` leer. Nachgemessen an `copyright`:
// im Bestand `PROBE-MARKE`, in der Datei unverändert.
//
// # Was hier **nicht** geschieht
//
// **Es wird keine Zeile entfernt.** Beim Kartenblock gibt es dafür eine
// Geste — ein Feld lässt sich leeren, und `zeilenEntfernen` bildet das ab.
// Für die Setangaben gibt es keine: die Felder stehen im Spielpaket, und was
// dort steht, verschwindet nicht dadurch, dass jemand ein Set speichert.
//
// **Und es wird kein `set info:` erfunden.** Ein Block, den niemand
// angeordnet hat, wäre eine Änderung an einer fremden Datei — die Bytezusage
// gilt für alles, was niemand angefasst hat
// ([0002](../../docs/entscheidungen/0002-sqlite-als-arbeitsformat.md)).

package service

import (
	"slices"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// setwerteAngleichen bringt den `set info:`-Block auf den Stand des Bestands.
func setwerteAngleichen(baum *msepaket.Baum, werte []domain.Value) {
	block := baum.Ersten("set info")
	if block == nil {
		return
	}

	// **Die setweiten Stiloptionen gehören nicht hierher** (F05, #2247): sie
	// stehen im eigenen Block `styling:` ([stilblockAngleichen]). Bis zum
	// 2026-10-10 hängte dieser Weg sie als Zeile an `set info:`, und MSE2 wies
	// sie beim Öffnen mit „Unexpected key: 'styling'" ab. **Eine solche Zeile
	// aus einer früheren Abschrift wird dabei entfernt** — sie hat in MSE2
	// keine Bedeutung, und hier läse der Import sie als zweiten Stilblock.
	werte = slices.DeleteFunc(slices.Clone(werte), func(wert domain.Value) bool {
		return wert.Field == Setstilfeld
	})
	block.Kinder = slices.DeleteFunc(block.Kinder, verirrterStilblock)

	setwerteSetzen(block, werte)

	block.Kinder = setzeilenErgaenzen(block.Kinder, werte)
}

// setwerteSetzen schreibt die Werte in die Zeilen, die es schon gibt.
//
// **Wie beim Kartenblock**: verglichen wird gegen den wirksamen Wert, und eine
// geänderte Zeile verliert ihre Rohfassung, damit `msepaket.Schreibe` sie neu
// zusammensetzt. Was gleich bleibt, behält seine Bytes.
func setwerteSetzen(block *msepaket.Knoten, werte []domain.Value) {
	for _, kind := range block.Kinder {
		if kind.Zeilenart != msepaket.ArtEintrag {
			continue
		}

		wert, da := setwertVon(werte, kind.Schluessel)
		if !da || wieInDerQuelle(wert, kind) {
			continue
		}

		// **Der alte Block geht mit, wenn er der Wert war** — derselbe Grund
		// wie an der Karte (#704): ein mehrzeiliger Wert liegt in den
		// eingerückten Kindern, und beide zu behalten schriebe ihn doppelt.
		if kind.Wert == "" {
			kind.Kinder = nil
		}

		kind.Roh = ""
		kind.Wert = wert
	}
}

// setzeilenErgaenzen setzt die Angaben ein, für die keine Zeile dasteht.
//
// **Ans Ende des Blocks**, nicht an die Stelle des Vorgängers. Beim
// Kartenblock gibt `card_values.reihenfolge` die Stelle her; die Setangaben
// stehen im Spielpaket in einer Reihenfolge, die mit der Datei nichts zu tun
// hat — eine daraus abgeleitete Stelle wäre eine Erfindung.
//
// **Ein Symbol ist genau dieser Fall**: `set.symbol` steht in vielen Sets gar
// nicht, und ein gezeichnetes muss trotzdem hinein.
func setzeilenErgaenzen(kinder []*msepaket.Knoten, werte []domain.Value) []*msepaket.Knoten {
	for _, wert := range werte {
		if wert.Text == "" || zeileVon(kinder, wert.Field) >= 0 {
			continue
		}

		neue := &msepaket.Knoten{
			Zeilenart: msepaket.ArtEintrag,
			// **Kanonisch geschrieben** — so schreibt MSE selbst (#396).
			Schluessel: msepaket.Kanonisch(wert.Field),
			Wert:       wert.Text,
			Zeilenende: zeilenendeVon(kinder),
		}

		// **Die letzte Zeile bekommt ihr Ende** — sonst kleben beide
		// zusammen; derselbe Grund wie am Kartenblock.
		if letzte := len(kinder) - 1; letzte >= 0 {
			neue.Zeilenende = kinder[letzte].Zeilenende
			kinder[letzte].Zeilenende = zeilenendeVon(kinder)
		}

		kinder = slices.Insert(kinder, len(kinder), neue)
	}

	return kinder
}

// setwertVon sucht einen Wert nach seinem Feldnamen.
//
// **Kanonisch verglichen** (#396): in der Datei steht `masterpiece_symbol`, im
// Modell `masterpiece symbol`.
func setwertVon(werte []domain.Value, schluessel string) (string, bool) {
	gesucht := msepaket.Kanonisch(schluessel)

	for _, wert := range werte {
		if msepaket.Kanonisch(wert.Field) == gesucht {
			return wert.Text, true
		}
	}

	return "", false
}

// verirrterStilblock erkennt die Zeile `styling:` unter `set info:`, die
// diese Anwendung bis zum 2026-10-10 dorthin schrieb (F05, #2247).
//
// **An ihrer Gestalt und nicht am Namen allein**: ein Spiel darf ein Setfeld
// `styling` erklären, und das bliebe stehen. Erkannt wird die flache Form
// `<paket>/<option>: <wert>` in jeder Zeile darunter — die schreibt nur der
// alte Weg, MSE2 nie.
func verirrterStilblock(kind *msepaket.Knoten) bool {
	if kind.Zeilenart != msepaket.ArtEintrag || msepaket.Kanonisch(kind.Schluessel) != Setstilfeld {
		return false
	}

	zeilen := strings.Split(strings.TrimSpace(wertText(kind)), "\n")

	for _, zeile := range zeilen {
		vorne, _, hat := strings.Cut(zeile, ":")
		if !hat || !strings.Contains(vorne, "/") {
			return false
		}
	}

	return len(zeilen) > 0 && zeilen[0] != ""
}
