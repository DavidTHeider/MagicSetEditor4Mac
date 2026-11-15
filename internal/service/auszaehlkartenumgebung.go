// **Die Umgebung einer Karte in der Auszählung** — dieselben Schritte wie
// beim Zeichnen, ohne die Inhaltsmaße.
//
// # Warum eine eigene Datei
//
// `auszaehldimension.go` stand an der Zeilengrenze aus §2, als der
// Schlüsselwortvermerk je Karte dazukam (Befund 5 des Zweitblicks über Phase
// 4.11). Der Schnitt folgt der Frage: *was eine Dimension rechnet*, steht
// dort; *worüber sie rechnet*, hier.

package service

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// auszaehlumgebung baut die Umgebung einer Karte.
//
// **Dieselben Schritte wie [kartenumgebungVorbereiten] und [nachbereiten]**,
// ohne die Inhaltsmaße: erst der Stil, dann die Werte der Karte, dann die
// Vorgaben, die Vorgabeskripte und die Wertskripte. Die Reihenfolge ist die
// von MSE (`SetScriptManager::updateAll`) und nicht wählbar — jeder Schritt
// liest, was der vorige gesetzt hat.
//
// **Mit einer Kartenliste steht `set` als das Set mit seinen Karten da**
// ([setMitKarten]) — die Ausfuhr rechnet damit `card_number`, das die Stelle
// der Karte im Set zählt
// (#2006).
// Die Auszählung gibt keine und setzt ihre Liste erst danach.
//
// **Die Felder und die Befunde der Vorgabe- und Wertskripte kommen mit**: die
// Auszählung liest `card` aus der Umgebung und lässt die Befunde fallen, die
// Ausfuhr nimmt die Felder **vor** der Umwandlung in Farben und Bilder und
// meldet die Befunde.
func auszaehlumgebung(
	ctx context.Context, spiel *msescript.Umgebung, vorlage Vorlage,
	karte *domain.CardRow, vermerk *Schluesselwortverwendung,
	kartenliste []msescript.Eintrag,
) (*msescript.Umgebung, []msescript.Eintrag, []domain.Befund) {
	// **Der Vermerk gehört zur Karte, auch wenn er am Spiel hängt.**
	vermerk.Leere()

	umgebung := spiel.Darueber()

	// **Der Stil zuerst**, denn jedes `default:` beginnt mit `stylesheet; `.
	// Ohne Bildquelle: `exists_in_package` antwortet dann mit Nein, und keine
	// der fünfzig Dimensionen fragt danach.
	belegeVorlage(ctx, umgebung, vorlage, nil, karte)

	if kartenliste != nil {
		set, _ := umgebung.Suche("set")
		umgebung.Setze("set", setMitKarten(set, kartenliste))
	}

	felder := make([]msescript.Eintrag, 0, len(karte.Values)+1)
	for _, wert := range karte.Values {
		felder = append(felder, msescript.Eintrag{
			Name: umgebungsname(wert.Field), Wert: msescript.Text(wert.Text),
		})
	}

	// **`notes` steht an jeder Karte, auch an einer ohne** — die Regel steht
	// bei [mitNotizen]. Drei der fünfzig Dimensionen lesen es.
	felder = mitNotizen(felder, karte)

	// **Die Karte steht in der Umgebung, bevor die Vorgaben rechnen.**
	//
	// Jedes `default:` liest `card.<feld>` — `card color` etwa rechnet
	// `card_color(casting_cost: card.casting_cost, …)`. Stünde dort die
	// leere Gestalt, käme „colorless" heraus, und weil das ein **nicht
	// leerer** Wert ist, ließe [mitVorgabeskripten] ihn danach stehen.
	//
	// **Gemessen am 2026-08-31**: mit dieser Zeile an der falschen Stelle
	// gaben **alle** 3.761 Karten von `White.mse-set` „Colorless" statt
	// „White".
	umgebung.Setze("card",
		kartenabbildungMit(felder, vorlage.Kartenfelder, vorlage.Setpaket,
			Skriptkartenkennung(karte)))

	felder = mitVorgaben(ctx, felder, vorlage.Kartenfelder, umgebung)
	umgebung.Setze("card",
		kartenabbildungMit(felder, vorlage.Kartenfelder, vorlage.Setpaket,
			Skriptkartenkennung(karte)))

	felder, vorgabebefunde := mitVorgabeskripten(ctx, felder, vorlage.Kartenfelder, umgebung)
	felder = normalisiereMehrfachwahlen(felder, vorlage.Kartenfelder)
	umgebung.Setze("card",
		kartenabbildungMit(felder, vorlage.Kartenfelder, vorlage.Setpaket,
			Skriptkartenkennung(karte)))

	felder, wertbefunde := mitWertskripten(ctx, felder, vorlage.Kartenfelder, umgebung, vermerk,
		func(stand []msescript.Eintrag) {
			umgebung.Setze("card",
				kartenabbildungMit(stand, vorlage.Kartenfelder, vorlage.Setpaket,
					Skriptkartenkennung(karte)))
		})
	felder = normalisiereMehrfachwahlen(felder, vorlage.Kartenfelder)
	umgebung.Setze("card",
		kartenabbildungMit(felder, vorlage.Kartenfelder, vorlage.Setpaket,
			Skriptkartenkennung(karte)))

	return umgebung, felder, append(vorgabebefunde, wertbefunde...)
}

// auszaehlbefunde meldet die gescheiterten Karten — **als einen Befund**.
//
// **Nicht einen je Karte.** Ein Skript, das an einer Karte scheitert,
// scheitert meist an allen; dreihundert gleichlautende Meldungen sind keine
// Auskunft. Was zählt, ist die Zahl der betroffenen Karten und der Grund.
func auszaehlbefunde(name string, gescheitert int, ursache error) []domain.Befund {
	if gescheitert == 0 {
		return nil
	}

	return []domain.Befund{{
		Ort: name,
		Meldung: fmt.Sprintf(
			"die Dimension %q ließ sich für %d Karten nicht rechnen: %v",
			name, gescheitert, ursache),
	}}
}
