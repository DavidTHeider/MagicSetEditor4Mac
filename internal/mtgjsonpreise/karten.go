package mtgjsonpreise

// **Die fünf Ebenen unter einer Kennung** — Werkstück A3 von Phase 5.3
// (#1442).
//
// Aus `lesen.go` herausgelöst: dort steht, **wie** gelesen wird — strömend,
// mit Deckel —, hier, **was** die Gestalt bedeutet.

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// preisliste ist, was ein Anbieter je Karte und Markt führt.
//
// **Die Feldnamen sind die des Anbieters** (§16, zweite Schicht): `currency`,
// `retail`, `buylist`. Ihre Bedeutung steht eine Ebene weiter, in [reihenAus].
type preisliste struct {
	Currency string `json:"currency"`

	// **Als `json.Number` und nicht als `float64`**: die Ziffernfolge geht
	// durch [domain.MarktpreisAus] und wird zur kleinsten Einheit. Ein `float64`
	// dazwischen wäre Rundungsdrift in einer Geldrechnung, und A3 verbietet
	// sie ausdrücklich.
	Retail  map[string]map[string]json.Number `json:"retail"`
	Buylist map[string]map[string]json.Number `json:"buylist"`
}

// dieMaerkte ordnet MTGJSONs Marktnamen unseren zu.
//
// **Dieselben Wörter**, und das ist die Entscheidung aus
// [0093](../../docs/entscheidungen/0093-eine-preisreihe-hat-sieben-teile.md):
// eine Übersetzung machte jede Zeile unlesbar für den, der beide Seiten
// vergleicht.
func dieMaerkte() map[string]domain.Preismarkt {
	return map[string]domain.Preismarkt{
		"paper": domain.MarktPapier,
		"mtgo":  domain.MarktOnline,
	}
}

// dieAusfuehrungen ordnet MTGJSONs Ausführungsnamen unseren zu.
//
// **`normal` heißt bei Scryfall `nonfoil`**, und das ist der eine Name, der
// wirklich verschieden ist. Die Ausführung steht in beiden Quellen für
// dasselbe; zwei Namen dafür in einer Tafel wären zwei Reihen für eine Sache.
func dieAusfuehrungen() map[string]domain.Finish {
	return map[string]domain.Finish{
		"normal": domain.FinishNonfoil,
		"foil":   domain.FinishFoil,
		"etched": domain.FinishEtched,
	}
}

// karten liest die Kennungen und gibt die Stände der gewollten.
func karten(
	zerteiler *json.Decoder, crosswalk map[string]string, grenzen Grenzen,
) ([]domain.Preisstand, []domain.Befund, error) {
	staende := []domain.Preisstand{}

	var befunde []domain.Befund

	for zerteiler.More() {
		marke, fehler := zerteiler.Token()
		if fehler != nil {
			return nil, befunde, domain.MitGrund(domain.ErrValidation, fehler,
				"eine Kennung der Preisdatei lesen")
		}

		// **Eine Marke im Objekt ist immer ein Text.**
		//
		// `json.Decoder.Token` gibt innerhalb eines Objekts entweder einen
		// Schlüssel als `string` oder die schließende Klammer — und die fängt
		// `More()` ab. Eine Prüfung darauf wäre ein toter Zweig, der behauptet, es
		// gebe einen Fall, den niemand betreten kann (§3).
		kennung, _ := marke.(string) //nolint:errcheck // siehe darüber

		printing, gewollt := crosswalk[kennung]

		// **Was der Bestand nicht kennt, wird übersprungen** — und zwar
		// **ohne** es zu lesen: die Datei führt jede Karte, die es gibt, und
		// der Bestand nur die, die jemand besitzt.
		if !gewollt {
			if fehler := ueberspringen(zerteiler); fehler != nil {
				return nil, befunde, fehler
			}

			continue
		}

		neue, karteBefunde, lesefehler := eineKarte(zerteiler, kennung, printing)
		if lesefehler != nil {
			return nil, befunde, lesefehler
		}

		befunde = append(befunde, karteBefunde...)
		staende = append(staende, neue...)

		if len(staende) > grenzen.HoechstensStaende {
			return nil, befunde, fmt.Errorf(
				"%w: die Preisdatei führt über %d Stände für diesen Bestand",
				domain.ErrValidation, grenzen.HoechstensStaende)
		}
	}

	return staende, befunde, nil
}

// ueberspringen liest einen Wert, ohne ihn zu behalten.
func ueberspringen(zerteiler *json.Decoder) error {
	var weg json.RawMessage
	if fehler := zerteiler.Decode(&weg); fehler != nil {
		return domain.MitGrund(domain.ErrValidation, fehler,
			"einen Eintrag der Preisdatei überspringen")
	}

	return nil
}

// eineKarte liest die Märkte einer Kennung.
func eineKarte(zerteiler *json.Decoder, kennung, printing string) (
	[]domain.Preisstand, []domain.Befund, error,
) {
	var maerkte map[string]map[string]preisliste
	if fehler := zerteiler.Decode(&maerkte); fehler != nil {
		return nil, nil, domain.MitGrund(domain.ErrValidation, fehler,
			"die Preise zu %s lesen", kennung)
	}

	staende := []domain.Preisstand{}

	var befunde []domain.Befund

	for marktname, anbieter := range maerkte {
		markt, bekannt := dieMaerkte()[marktname]
		if !bekannt {
			// **Ein unbekannter Markt ist ein Befund und kein Abbruch**
			// (§16): MTGJSON kann einen dritten aufnehmen, und die zwei
			// bekannten sollen deshalb nicht ausfallen.
			befunde = append(befunde, domain.Befund{
				Datei: kennung, Ort: marktname, Art: domain.ArtUmgebung,
				Meldung: fmt.Sprintf(
					"MTGJSON führt den Markt %q, den diese Anwendung nicht kennt. "+
						"Seine Preise werden übergangen.", marktname),
			})

			continue
		}

		for name, liste := range anbieter {
			neue, listenBefunde := reihenAus(kennung, printing, markt, name, liste)
			befunde = append(befunde, listenBefunde...)
			staende = append(staende, neue...)
		}
	}

	return staende, befunde, nil
}

// reihenAus macht aus einer Anbieterliste ihre Stände.
func reihenAus(
	kennung, printing string, markt domain.Preismarkt,
	anbieter string, liste preisliste,
) ([]domain.Preisstand, []domain.Befund) {
	staende := []domain.Preisstand{}

	var befunde []domain.Befund

	for art, nachAusfuehrung := range map[domain.Preisart]map[string]map[string]json.Number{
		domain.PreisartVerkauf: liste.Retail,
		domain.PreisartAnkauf:  liste.Buylist,
	} {
		for ausfuehrungsname, nachTag := range nachAusfuehrung {
			finish, bekannt := dieAusfuehrungen()[ausfuehrungsname]
			if !bekannt {
				befunde = append(befunde, domain.Befund{
					Datei: kennung, Ort: ausfuehrungsname, Art: domain.ArtUmgebung,
					Meldung: fmt.Sprintf(
						"MTGJSON führt die Ausführung %q, die diese Anwendung "+
							"nicht kennt. Ihre Preise werden übergangen.",
						ausfuehrungsname),
				})

				continue
			}

			neue, tagBefunde := tageAus(kennung, printing, markt, anbieter,
				art, finish, liste.Currency, nachTag)

			befunde = append(befunde, tagBefunde...)
			staende = append(staende, neue...)
		}
	}

	return staende, befunde
}

// tageAus macht aus den Tagen einer Ausführung ihre Stände.
func tageAus(
	kennung, printing string, markt domain.Preismarkt, anbieter string,
	art domain.Preisart, finish domain.Finish, waehrung string,
	nachTag map[string]json.Number,
) ([]domain.Preisstand, []domain.Befund) {
	staende := make([]domain.Preisstand, 0, len(nachTag))

	var befunde []domain.Befund

	for tag, betrag := range nachTag {
		beobachtet, tagfehler := time.Parse(time.DateOnly, tag)
		if tagfehler != nil {
			befunde = append(befunde, domain.Befund{
				Datei: kennung, Ort: tag, Art: domain.ArtUmgebung,
				Meldung: fmt.Sprintf(
					"MTGJSON nennt den Tag %q, der kein Datum ist: %v. "+
						"Dieser Preis wird übergangen.", tag, tagfehler),
			})

			continue
		}

		wert, wertfehler := domain.MarktpreisAus(betrag.String(), waehrung)
		if wertfehler != nil {
			befunde = append(befunde, domain.Befund{
				Datei: kennung, Ort: tag, Art: domain.ArtUmgebung,
				Meldung: fmt.Sprintf(
					"MTGJSON nennt für %s den Preis %q, der nicht zu deuten "+
						"ist: %v. Die übrigen Preise werden übernommen.",
					tag, betrag.String(), wertfehler),
			})

			continue
		}

		staende = append(staende, domain.Preisstand{
			Printing: printing,
			// **Der Anbieter steht in der Quelle** (0093): MTGJSON ist ein
			// Sammler, kein Händler — `cardkingdom` und `cardmarket` nennen
			// verschiedene Preise, und eine Quelle für beide wäre eine Reihe
			// für zwei Aussagen.
			Quelle:     Familie + domain.Quellentrenner + domain.Preisquelle(strings.ToLower(anbieter)),
			Markt:      markt,
			Art:        art,
			Waehrung:   waehrung,
			Finish:     finish,
			Lage:       domain.PreisGenannt,
			Wert:       wert,
			Beobachtet: beobachtet,
		})
	}

	return staende, befunde
}
