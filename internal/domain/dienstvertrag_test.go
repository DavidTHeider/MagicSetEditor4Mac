package domain_test

// **Der Providervertrag je Fähigkeit** — Werkstück A1 von Phase 5.3
// (#1440).
//
// Was hier geprüft wird, ist die **Gestalt** des Katalogs: dass jeder Vertrag
// vollständig ist, dass nur ein Dienst mit Vertrag zur Wahl steht und dass ein
// Nachweis Lauf, Datum und Fassung nennt. **Ob die Zahlen stimmen**, hält
// `internal/werkzeuge/dienstvertraege_test.go` gegen die Klientenpakete —
// eine Adresse im Katalog, die das Paket nicht benutzt, wäre eine Behauptung.

import (
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

var datum = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

func TestJederVertragIstVollstaendig(t *testing.T) {
	t.Parallel()

	vertraege := domain.Dienstvertraege()
	if len(vertraege) == 0 {
		t.Fatal("der Katalog ist leer")
	}

	// **Doppelte vor den Unterläufen zählen** — die laufen parallel, und ein
	// gemeinsames `map` darunter wäre das Wettrennen, das `-race` meldet.
	gesehen := map[string]bool{}

	for _, vertrag := range vertraege {
		name := string(vertrag.Dienst) + "/" + string(vertrag.Frage)
		if gesehen[name] {
			t.Errorf("%s steht zweimal im Katalog", name)
		}

		gesehen[name] = true
	}

	for _, vertrag := range vertraege {
		name := string(vertrag.Dienst) + "/" + string(vertrag.Frage)

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if !slices.Contains(domain.Dienste(), vertrag.Dienst) {
				t.Errorf("der Dienst %q hat keinen Schalter", vertrag.Dienst)
			}

			if vertrag.Frage == "" {
				t.Error("die Frage fehlt")
			}

			if vertrag.Wirkung != domain.WirkungExternSenden &&
				vertrag.Wirkung != domain.WirkungExternSchreiben {
				t.Errorf("die Wirkung %q ist keine nach draußen", vertrag.Wirkung)
			}

			for feld, wert := range map[string]string{
				"Basisadresse": vertrag.Basisadresse,
				"Bedingungen":  vertrag.Bedingungen,
				"Beleg":        vertrag.Beleg,
			} {
				if !strings.HasPrefix(wert, "https://") {
					t.Errorf("%s ist %q — erwartet eine prüfbare https-Adresse", feld, wert)
				}
			}

			for feld, wert := range map[string]string{
				"Zugang": vertrag.Zugang, "Tarif": vertrag.Tarif,
				"Grenze": vertrag.Grenze, "Frist": vertrag.Frist,
			} {
				if strings.TrimSpace(wert) == "" {
					t.Errorf("%s ist leer", feld)
				}
			}

			if !datum.MatchString(vertrag.Abgerufen) {
				t.Errorf("Abgerufen ist %q, erwartet JJJJ-MM-TT", vertrag.Abgerufen)
			}

			if len(vertrag.Endpunkte) == 0 {
				t.Error("der Vertrag nennt keinen Endpunkt")
			}
		})
	}
}

func TestDieSucheKoennenAlleVier(t *testing.T) {
	t.Parallel()

	erwartet := []domain.Dienst{
		domain.DienstScryfall, domain.DienstMTGAPI,
		domain.DienstMTGJSON, domain.DienstScrydex,
	}

	if dienste := domain.DiensteFuer(domain.FrageSuche); !slices.Equal(dienste, erwartet) {
		t.Errorf("die Suche können %v, erwartet %v", dienste, erwartet)
	}

	for _, dienst := range erwartet {
		if _, gibts := domain.VertragFuer(dienst, domain.FrageSuche); !gibts {
			t.Errorf("%s steht zur Wahl, hat aber keinen Vertrag für die Suche", dienst)
		}
	}
}

// **Nur tatsächlich implementierte Fragen** (#1440): ein Universalanbieter
// mit Methoden, die »nicht unterstützt« melden, ist genau das, was der
// Katalog verhindert.
//
// **Sechs Fragen beantwortet nur Scryfall, die Preise beantworten drei.**
// Hier stand bis A4 „die übrigen beantwortet nur Scryfall" samt den Preisen
// in der Aufzählung — und das war schon mit A3 (#1442) nicht mehr wahr: der
// Bulkbezug holt Preise bei MTGJSON, während der Katalog sagte, MTGJSON
// könne hier suchen und sonst nichts. Aufgefallen beim Eintragen von
// CardTrader.
func TestDieUebrigenFragenBeantwortetNurScryfall(t *testing.T) {
	t.Parallel()

	uebrige := []domain.Frage{
		domain.FrageKartendaten, domain.FrageBild, domain.FrageRegeln,
		domain.FrageManakosten, domain.FrageSetangaben, domain.FrageSammeldaten,
	}

	for _, frage := range uebrige {
		if dienste := domain.DiensteFuer(frage); !slices.Equal(
			dienste, []domain.Dienst{domain.DienstScryfall}) {
			t.Errorf("%s beantworten %v, erwartet nur scryfall", frage, dienste)
		}
	}

	// **Die Preise sind die Ausnahme, und sie ist aufgezählt** — wer einen
	// vierten Preisdienst aufnimmt, ändert diese Zeile und sagt damit, dass
	// er es wollte.
	if dienste := domain.DiensteFuer(domain.FragePreise); !slices.Equal(
		dienste, []domain.Dienst{
			domain.DienstScryfall, domain.DienstMTGJSON,
			domain.DienstScrydex, domain.DienstCardTrader,
		}) {
		t.Errorf("die Preise beantworten %v", dienste)
	}

	if fragen := domain.FragenVon(domain.DienstMTGAPI); !slices.Equal(
		fragen, []domain.Frage{domain.FrageSuche}) {
		t.Errorf("magicthegathering.io kann %v, erwartet nur die Suche", fragen)
	}

	// **Zwei Dienste können beides** — suchen und Preise (A3 und A5).
	for _, dienst := range []domain.Dienst{domain.DienstMTGJSON, domain.DienstScrydex} {
		if fragen := domain.FragenVon(dienst); !slices.Equal(
			fragen, []domain.Frage{domain.FrageSuche, domain.FragePreise}) {
			t.Errorf("%s kann %v, erwartet Suche und Preise", dienst, fragen)
		}
	}

	// **CardTrader kann seit T11 zwei** (#1476): die Preise des Marktplatzes
	// und das **eigene** Konto. Zwei Verträge und nicht ein größerer — jener
	// fragt nach fremden Angeboten, dieser nach eigenen.
	if fragen := domain.FragenVon(domain.DienstCardTrader); !slices.Equal(
		fragen, []domain.Frage{domain.FragePreise, domain.FrageKonto}) {
		t.Errorf("CardTrader kann %v, erwartet Preise und Konto", fragen)
	}

	if fragen := domain.FragenVon(domain.DienstScryfall); len(fragen) != len(uebrige)+2 {
		t.Errorf("Scryfall kann %v — erwartet die Suche, die Preise und %d weitere",
			fragen, len(uebrige))
	}
}

// **Eine Wahl gibt es nur, wo mehr als einer antworten kann.** Eine
// Einstellung »frage.bild« mit einem einzigen erlaubten Wert wäre eine
// Auswahlliste ohne Auswahl.
func TestEineWahlGibtEsNurBeiMehrAlsEinemDienst(t *testing.T) {
	t.Parallel()

	if fragen := domain.Fragen(); !slices.Equal(fragen, []domain.Frage{domain.FrageSuche}) {
		t.Errorf("zur Wahl stehen %v, erwartet nur die Suche", fragen)
	}

	if slices.Contains(domain.Einstellungsreihe(), domain.FrageEinstellung(domain.FrageBild)) {
		t.Error("»frage.bild« hat eine Einstellung — mit genau einem erlaubten Wert")
	}
}

// **Ein Nachweis ist ein Lauf mit Datum und Fassung** — kein Gütesiegel.
// Zwei Dienste haben einen (0056, Entscheidung 4): Scryfall und
// magicthegathering.io; für MTGJSON und Scrydex gibt es hier kein Konto.
//
// **Und bei Scryfall nicht jeder Vertrag**: der Lauf ruft weder die
// Bulk-Dateien noch ein Bild. Ein Nachweis gilt je Endpunkt (#1440) — wo der
// Lauf keinen Endpunkt des Vertrags ruft, trägt der Vertrag keinen.
func TestEinNachweisNenntLaufDatumUndFassung(t *testing.T) {
	t.Parallel()

	ohneLauf := []domain.Frage{domain.FrageBild, domain.FrageSammeldaten}

	for _, vertrag := range domain.Dienstvertraege() {
		nachweis := vertrag.Nachweis

		switch vertrag.Dienst {
		case domain.DienstScryfall, domain.DienstMTGAPI:
			if slices.Contains(ohneLauf, vertrag.Frage) {
				if nachweis != nil {
					t.Errorf("%s/%s trägt einen Nachweis — kein Lauf ruft seine Endpunkte",
						vertrag.Dienst, vertrag.Frage)
				}

				continue
			}

			if nachweis == nil {
				t.Errorf("%s/%s hat keinen Nachweis", vertrag.Dienst, vertrag.Frage)

				continue
			}

			if !deckt(nachweis.Endpunkte, vertrag.Endpunkte) {
				t.Errorf("der Nachweis von %s/%s ruft %v — der Vertrag nennt %v",
					vertrag.Dienst, vertrag.Frage, nachweis.Endpunkte, vertrag.Endpunkte)
			}

			if !strings.HasSuffix(nachweis.Lauf, "netzlauf_test.go") {
				t.Errorf("der Lauf %q ist keine Datei hinter dem Bautag netzlauf", nachweis.Lauf)
			}

			if !datum.MatchString(nachweis.Datum) {
				t.Errorf("das Datum %q ist keines", nachweis.Datum)
			}

			if nachweis.Fassung == "" {
				t.Error("der Nachweis nennt keine Fassung")
			}

		default:
			if nachweis != nil {
				t.Errorf("%s hat einen Nachweis — es gibt hier kein Konto dafür", vertrag.Dienst)
			}
		}
	}
}

func TestEinVertragFuerEinUnbekanntesPaarGibtEsNicht(t *testing.T) {
	t.Parallel()

	if _, gibts := domain.VertragFuer(domain.DienstScrydex, domain.FrageBild); gibts {
		t.Error("Scrydex hat einen Vertrag für Bilder — den es nicht gibt")
	}

	if fragen := domain.FragenVon("erfunden"); fragen != nil {
		t.Errorf("ein erfundener Dienst kann %v", fragen)
	}

	if dienste := domain.DiensteFuer("erfunden"); dienste != nil {
		t.Errorf("eine erfundene Frage beantworten %v", dienste)
	}
}

// deckt sagt, ob der Lauf wenigstens einen Endpunkt des Vertrags ruft.
func deckt(gelaufen, genannt []string) bool {
	for _, endpunkt := range genannt {
		if slices.Contains(gelaufen, endpunkt) {
			return true
		}
	}

	return false
}

// **Preise sind mehrstimmig, und deshalb gibt es dafür keine Wahl.**
//
// # Der Unterschied zur Suche
//
// Bei der **Suche** antwortet genau einer: wer „Lightning Bolt" tippt, bekommt
// eine Trefferliste, und zwei Dienste ergäben zwei Listen, zwischen denen
// niemand entschieden hat. Deshalb die Wahl.
//
// Bei **Preisen** antworten alle nebeneinander: Scryfall schätzt, MTGJSON
// sammelt fünf Anbieter, CardTrader nennt Angebote — und jede dieser Aussagen
// ist eine **eigene Reihe**
// ([0093](../../docs/entscheidungen/0093-eine-preisreihe-hat-sieben-teile.md)),
// keine Alternative zu den anderen. Eine Einstellung „nimm für Preise jenen
// Dienst" wäre ein Schalter, den niemand liest — und das ist schlimmer als
// keiner (§10).
func TestFuerPreiseGibtEsKeineWahl(t *testing.T) {
	t.Parallel()

	if !domain.FragePreise.Mehrstimmig() {
		t.Error("die Preise gelten als einstimmig")
	}

	if domain.FrageSuche.Mehrstimmig() {
		t.Error("die Suche gilt als mehrstimmig")
	}

	// **Drei Dienste, und trotzdem keine Wahl** — genau das prüft der Lauf.
	if dienste := domain.DiensteFuer(domain.FragePreise); len(dienste) < 2 {
		t.Fatalf("die Preise beantworten %v — der Lauf prüft nichts", dienste)
	}

	if slices.Contains(domain.Fragen(), domain.FragePreise) {
		t.Error("»frage.preise« steht zur Wahl")
	}

	if slices.Contains(domain.Einstellungsreihe(),
		domain.FrageEinstellung(domain.FragePreise)) {
		t.Error("»frage.preise« hat eine Einstellung")
	}
}
