// **Die Preise an der Kommandozeile** — Werkstück 3 von Phase 3.4
// (#1165).
//
// # Der Vorbehalt steht in der Ausgabe
//
// Die Quelle sagt über ihre Preise ausdrücklich, dass sie für nichts einsteht.
// Jede Zeile trägt deshalb **Quelle, Zeitpunkt und Alter** — auch die, die nur
// aus dem Bestand liest.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// preisebefehl gibt die abgelegten Preise eines Drucks.
func preisebefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) < 1 {
		an.klage("preise braucht die Kennung des Drucks.\n%s", hilfe)

		return CodeBedienung
	}

	return mitSammlung(ctx, bestandsort, an, func(ctx context.Context, naht *app.Sammlungsnaht) int {
		staende, fehler := naht.Preise(ctx, rest[0])
		if fehler != nil {
			an.klageUeber(fehler)

			return CodeFachlich
		}

		return preiseSchreiben(an, staende, nil)
	})
}

// preiseHolenbefehl fragt Scryfall und legt die Stände ab.
func preiseHolenbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) < 1 {
		an.klage("preise-holen braucht die Kennung des Drucks.\n%s", hilfe)

		return CodeBedienung
	}

	return mitSammlung(ctx, bestandsort, an, func(ctx context.Context, naht *app.Sammlungsnaht) int {
		staende, befunde, fehler := naht.PreiseHolen(ctx, rest[0])
		if fehler != nil {
			// **Die Befunde kommen auch im Fehlerfall heraus.** Sonst
			// verschwände „drei Preise waren nicht zu deuten" hinter dem einen
			// Satz „es ging nicht".
			for _, befund := range befunde {
				an.klage("%s\n", befund)
			}

			an.klageUeber(fehler)

			return CodeFachlich
		}

		return preiseSchreiben(an, staende, befunde)
	})
}

// preisverlaufbefehl gibt eine Reihe in der Zeit.
//
// # Brechend, und zwar mit Absicht (§16)
//
// Der Befehl hieß `preisverlauf <druck> <waehrung> <finish>`, und die drei
// fehlenden Teile der Reihe ergänzte die Naht dahinter mit „Scryfall, Papier,
// Verkauf". Solange es eine Quelle gab, war das richtig geraten; mit MTGJSON
// daneben ist es eine **Preisstrategie**, und A3 (#1442) verbietet sie.
//
// **Zustand und Sprache bleiben wahlfrei**: leer ist dort eine Angabe
// („unbekannt"), und heute füllt sie keine Quelle.
func preisverlaufbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) < 6 {
		an.klage("preisverlauf braucht Druck, Quelle, Markt, Preisart, "+
			"Währung und Ausführung, wahlweise Zustand und Sprache.\n%s", hilfe)

		return CodeBedienung
	}

	reihe := app.PreisreiheDTO{
		Printing: rest[0], Quelle: rest[1], Markt: rest[2],
		Art: rest[3], Waehrung: rest[4], Finish: rest[5],
	}

	if len(rest) > 6 {
		reihe.Zustand = rest[6]
	}

	if len(rest) > 7 {
		reihe.Sprache = rest[7]
	}

	return mitSammlung(ctx, bestandsort, an, func(ctx context.Context, naht *app.Sammlungsnaht) int {
		staende, fehler := naht.Preisverlauf(ctx, reihe)
		if fehler != nil {
			an.klageUeber(fehler)

			return CodeFachlich
		}

		return preiseSchreiben(an, staende, nil)
	})
}

// preisbulkbefehl holt eine ganze Preisdatei und legt ab, was darin steht.
//
// **Quelle und Umfang werden genannt und nicht erraten** (A3, #1442): es gibt
// heute eine Quelle, und trotzdem steht sie im Befehl — eine Vorgabe wäre eine
// Preisstrategie, die aussieht wie keine.
func preisbulkbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) < 2 {
		an.klage("preise-bulk braucht die Quelle (»mtgjson«) und den Umfang "+
			"(»heute« oder »neunzigTage«).\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.PreisbulkAnwendungUeber(db).Preisbulk(ctx, rest[0], rest[1])
		if antwort.Fehler != nil {
			an.klageUeber(antwort.Fehler.AlsFehler())

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort.BulkbilanzDTO)
		}

		an.sag("quelle\t%s\n", antwort.Quelle)
		an.sag("umfang\t%s\n", antwort.Umfang)
		an.sag("drucke\t%d\n", antwort.Drucke)
		an.sag("staende\t%d\n", antwort.Staende)

		for _, befund := range antwort.Befunde {
			an.sag("befund\t%s\n", befund)
		}

		return an.abschluss()
	})
}

// portfoliobefehl rechnet die Sammlung zusammen.
func portfoliobefehl(ctx context.Context, bestandsort string, an *schreiber, _ []string) int {
	return mitSammlung(ctx, bestandsort, an, func(ctx context.Context, naht *app.Sammlungsnaht) int {
		zeilen, ohnePreis, fehler := naht.Portfolio(ctx)
		if fehler != nil {
			an.klageUeber(fehler)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, map[string]any{
				"zeilen": zeilen, "ohnePreis": ohnePreis,
			})
		}

		for _, zeile := range zeilen {
			an.sag("%s\t%s\t%d Exemplare\t%s\tvor %d h\n", zeile.Waehrung,
				zeile.Wert, zeile.Exemplare, zeile.Quelle, zeile.AlterStunden)
		}

		// **Die Zahl ohne Preis steht immer da, auch wenn sie null ist.** Eine
		// Summe ohne sie sähe aus wie eine über alles.
		an.sag("ohne Preis: %d\n", ohnePreis)

		return an.abschluss()
	})
}

// preiseSchreiben gibt Preisstände aus — als Text oder als JSON.
func preiseSchreiben(an *schreiber, staende []app.PreisstandDTO, befunde []string) int {
	if an.alsJSON {
		return alsJSONSchreiben(an, map[string]any{
			"staende": staende, "befunde": jeEinTextOhneNil(befunde),
		})
	}

	for _, befund := range befunde {
		an.klage("%s\n", befund)
	}

	for _, stand := range staende {
		an.sag("%s\t%s\t%s\t%s\tvor %d h\n", stand.Waehrung, stand.Finish,
			stand.Wert, stand.Quelle, stand.AlterStunden)
	}

	return an.abschluss()
}

// jeEinTextOhneNil macht aus einer fehlenden Liste eine leere (§3).
func jeEinTextOhneNil(texte []string) []string {
	if texte == nil {
		return []string{}
	}

	return texte
}

// marktbefehl gibt die Angebote eines Marktplatzes zu einem Druck.
//
// **Die Quelle steht im Befehl** — wie bei `preise-bulk` und aus demselben
// Grund (A3, A4): eine Vorgabe wäre eine Marktstrategie, die aussieht wie
// keine.
//
// **Sprache, Ausführung und Zustände sind wahlfrei** und stehen als Flaggen,
// nicht als Stellen: drei leere Stellen hintereinander sind keine Aufschrift
// mehr.
func marktbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) < 2 {
		an.klage("markt braucht die Quelle (»cardtrader«) und den Druck.\n%s", hilfe)

		return CodeBedienung
	}

	wunsch, fehler := marktwunschAus(rest[2:])
	if fehler != nil {
		an.klageUeber(fehler)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.MarktAnwendungUeber(db).Marktangebote(ctx,
			rest[0], rest[1], wunsch.sprache, wunsch.finish, wunsch.zustaende)
		if antwort.Fehler != nil {
			an.klageUeber(antwort.Fehler.AlsFehler())

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort.MarktantwortDTO)
		}

		return marktSchreiben(an, antwort.MarktantwortDTO)
	})
}

// marktwunsch sind die wahlfreien Einschränkungen eines Marktabrufs.
type marktwunsch struct {
	sprache   string
	finish    string
	zustaende []string
}

// marktwunschAus liest die Flaggen hinter Quelle und Druck.
func marktwunschAus(flaggen []string) (marktwunsch, error) {
	var wunsch marktwunsch

	for _, flagge := range flaggen {
		name, wert, hatWert := strings.Cut(flagge, "=")
		if !hatWert {
			return marktwunsch{}, fmt.Errorf(
				"%w: %q ist keine Angabe der Form --name=wert",
				domain.ErrValidation, flagge)
		}

		switch name {
		case "--sprache":
			wunsch.sprache = wert
		case "--finish":
			wunsch.finish = wert
		case "--zustand":
			// **Mehrfach erlaubt**: `--zustand="Near Mint" --zustand=Played`
			// fragt nach beiden. Eine Liste in einer Zeichenkette bräuchte ein
			// Trennzeichen, und CardTraders Zustände tragen Leerzeichen.
			wunsch.zustaende = append(wunsch.zustaende, wert)
		default:
			return marktwunsch{}, fmt.Errorf(
				"%w: die Angabe %q kennt der Befehl markt nicht — es gibt "+
					"--sprache, --finish und --zustand",
				domain.ErrValidation, name)
		}
	}

	return wunsch, nil
}

// marktSchreiben gibt die Angebote als Zeilen.
func marktSchreiben(an *schreiber, antwort app.MarktantwortDTO) int {
	an.sag("quelle\t%s\n", antwort.Quelle)
	an.sag("fassung\t%s\n", antwort.Fassung)

	if antwort.Zugeordnet {
		an.sag("zugeordnet\tja\n")
	}

	for _, angebot := range antwort.Angebote {
		// **Der Urlaub steht dabei und wird nicht weggelassen**: ein Posten,
		// der da ist und nicht kommt, ist eine andere Auskunft als keiner.
		lage := ""
		if angebot.ImUrlaub {
			lage = "\tim Urlaub"
		}

		an.sag("%s %s\t%d Stück\t%s\t%s\t%s\t%s%s\n",
			angebot.Wert, angebot.Waehrung, angebot.Menge, angebot.Finish,
			angebot.Zustand, angebot.Sprache, angebot.Verkaeufer, lage)
	}

	for _, befund := range antwort.Befunde {
		an.sag("befund\t%s\n", befund)
	}

	return an.abschluss()
}

// anbieterstandbefehl nennt das Kontingent an einem Anbieterkonto.
func anbieterstandbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) < 1 {
		an.klage("anbieterstand braucht den Dienst (»scrydex«).\n%s", hilfe)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.AnbieterAnwendungUeber(db).Anbieterstand(ctx, rest[0])
		if antwort.Fehler != nil {
			an.klageUeber(antwort.Fehler.AlsFehler())

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort.AnbieterstandDTO)
		}

		an.sag("dienst\t%s\n", antwort.Dienst)
		an.sag("gesamt\t%d\n", antwort.Gesamt)
		an.sag("rest\t%d\n", antwort.Rest)
		an.sag("verbraucht\t%d\n", antwort.Verbraucht)
		an.sag("ueberzugssatz\t%s\n", antwort.Ueberzugssatz)

		// **Die Frist steht immer da**, auch in der Zeilenform: ohne sie
		// sähe der Rest aus wie eine Sperre.
		an.sag("hoechstens alt\t%d Minuten\n", antwort.HoechstensAltMinuten)

		return an.abschluss()
	})
}

// fremdverlaufbefehl holt den Preisverlauf eines Drucks bei einer Quelle.
func fremdverlaufbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) < 2 {
		an.klage("fremdverlauf braucht die Quelle (»scrydex«) und den Druck.\n%s", hilfe)

		return CodeBedienung
	}

	wunsch, fehler := marktwunschAus(rest[2:])
	if fehler != nil {
		an.klageUeber(fehler)

		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		antwort := app.AnbieterAnwendungUeber(db).Fremdpreisverlauf(ctx,
			rest[0], rest[1], wunsch.finish, wunsch.zustaende)
		if antwort.Fehler != nil {
			an.klageUeber(antwort.Fehler.AlsFehler())

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, antwort.FremdpreisbilanzDTO)
		}

		an.sag("quelle\t%s\n", antwort.Quelle)
		an.sag("fassung\t%s\n", antwort.Fassung)
		an.sag("staende\t%d\n", antwort.Staende)

		if antwort.Zugeordnet {
			an.sag("zugeordnet\tja\n")
		}

		for _, befund := range antwort.Befunde {
			an.sag("befund\t%s\n", befund)
		}

		return an.abschluss()
	})
}
