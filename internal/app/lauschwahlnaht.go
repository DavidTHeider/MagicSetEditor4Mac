package app

// **Die Lauschwahl an den Diensten** — G08
// (#2272).
//
// `mse-rest` und `mse-mcp` nehmen `--adresse` wiederholt; ohne sie gilt für
// den REST-Dienst die Einstellung `netz.lauschen` — so startet auch der
// LaunchAgent, der keine Adresse mitgibt, dort, wo gewählt wurde. Die Regeln
// stehen an **einer** Stelle, und beide Programme fragen sie hier.

import (
	"context"
	"flag"
	"net/netip"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Lauschwahlflagge meldet `--adresse` wiederholbar an und gibt, was genannt
// wurde — **leer, wenn nichts**: was dann gilt, entscheidet der Aufrufer.
func Lauschwahlflagge(fahne *flag.FlagSet, erklaerung string) func() []string {
	var genannt []string

	fahne.Func("adresse", erklaerung, func(wert string) error {
		genannt = append(genannt, wert)

		return nil
	})

	return func() []string { return genannt }
}

// LauschwahlAusGenannten liest, was an `--adresse` stand — jede auch als
// Liste mit Komma.
func LauschwahlAusGenannten(genannt []string) (domain.Lauschwahl, error) {
	return domain.LauschwahlAus(strings.Join(genannt, ","))
}

// LauschwahlFuerDienst gibt die genannten Adressen oder, ohne sie, die
// Einstellung.
func (a *App) LauschwahlFuerDienst(ctx context.Context, genannt []string) (domain.Lauschwahl, error) {
	ctx = a.imProfil(ctx)

	if len(genannt) > 0 {
		return LauschwahlAusGenannten(genannt)
	}

	stand, fehler := a.einstellungen.Stand(ctx)
	if fehler != nil {
		return domain.Lauschwahl{}, fehler
	}

	return domain.LauschwahlAus(stand[domain.NetzLauschen].Wirksam)
}

// Rechnerteile gibt die Adressen ohne Port und ohne Klammern — die Namen,
// unter denen sich ein Dienst auf ihnen ansprechen lässt (`Host`).
//
// **Seit N6 durfte ein Dienst auf `::1` binden, und er wies jede eigene
// Anfrage ab**: die Herkunftsprüfung kannte nur `127.0.0.1` und
// `localhost`. Was ein Lauscher hält, ist ein Name, unter dem er gemeint ist.
func Rechnerteile(adressen []string) []string {
	teile := make([]string, 0, len(adressen))

	for _, adresse := range adressen {
		if wirt, gut := netip.ParseAddrPort(adresse); gut == nil {
			teile = append(teile, wirt.Addr().WithZone("").String())

			continue
		}

		teile = append(teile, adresse)
	}

	return teile
}

// NachAussen gibt die Adressen, die den Rechner verlassen — für sie gilt das
// Zeugnis (0115).
func NachAussen(bindungen []domain.Bindung) []string {
	var draussen []string

	for _, bindung := range bindungen {
		if bindung.VerlaesstDenRechner() {
			draussen = append(draussen, bindung.Wirkadresse())
		}
	}

	return draussen
}

// dienstStartenUeberWahl startet einen Dienst aus dem Fenster auf der
// genannten Wahl oder der Einstellung.
func (a *App) dienstStartenUeberWahl(
	ctx context.Context, art domain.Dienstart, genannt []string, port int,
) (Dienstzustand, error) {
	ctx = a.imProfil(ctx)

	wahl, fehler := a.LauschwahlFuerDienst(ctx, genannt)
	if fehler != nil {
		return Dienstzustand{}, fehler
	}

	return a.dienstStartenAuf(ctx, art, wahl.Bindungen(port))
}
