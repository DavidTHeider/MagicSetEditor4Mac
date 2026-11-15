package werkzeuge_test

// **Der Vertragskatalog hält gegen die Pakete** — Werkstück A1 von Phase 5.3
// (#1440),
// entschieden in 0082.
//
// # Was er hält
//
// Für jeden [domain.Dienstvertrag]: die Basisadresse **ist** die `Adresse`
// des Klientenpakets, Zeichen für Zeichen; jeder genannte Endpunkt kommt im
// Quelltext des Pakets vor (der feste Teil vor einem Platzhalter wie `{id}`);
// jede Adresse in Bedingungen und Beleg ist `https://`, und das Abrufdatum
// ist ein Datum.
//
// # Was er nicht hält
//
// **Ob Tarif, Grenze und Frist noch stimmen.** Das steht beim Anbieter, und
// dafür trägt der Vertrag sein Abrufdatum. Und **ob ein Endpunkt wirklich
// gerufen wird** — er sieht die Zeichenkette, nicht den Aufruf.

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/anthropic"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/cardtrader"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtgapi"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtggraphql"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtgjsonpreise"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/openai"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scrydex"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/xai"
)

// paketzeile ist ein Paket mit seiner Adresse.
type paketzeile struct {
	paket   string
	adresse string
}

// klientenpakete ordnet jedem Dienst sein Paket und dessen Adresse zu.
//
// **Die Vorgabe je Dienst steht unter der leeren Frage**, eine Abweichung je
// Frage daneben. Ein Dienst kann hier **mehrere** Pakete haben: MTGJSON wird
// über `mtggraphql` gesucht und über `mtgjsonpreise` nach Preisen gefragt,
// und das sind zwei Verträge mit zwei Adressen und zwei Grenzen (A4, #1443).
//
// **Vorher hieß der Schlüssel nur `Dienst`** — damit galt für jede Frage
// dieselbe Adresse, und der Preisvertrag von MTGJSON wurde gegen den
// GraphQL-Dienst gehalten.
var klientenpakete = map[struct {
	dienst domain.Dienst
	frage  domain.Frage
}]paketzeile{
	{domain.DienstScryfall, ""}: {"scryfall", scryfall.Adresse},
	{domain.DienstMTGAPI, ""}:   {"mtgapi", mtgapi.Adresse},
	{domain.DienstMTGJSON, ""}:  {"mtggraphql", mtggraphql.Adresse},
	{domain.DienstScrydex, ""}:  {"scrydex", scrydex.Adresse},

	{domain.DienstMTGJSON, domain.FragePreise}: {
		"mtgjsonpreise", mtgjsonpreise.Vorgabeadresse,
	},
	{domain.DienstCardTrader, ""}: {"cardtrader", cardtrader.Vorgabeadresse},

	// **Der erste, der keine Karten liefert** (A8, #1447) — und trotzdem
	// derselbe Vertrag: eine Basisadresse, Endpunkte, Bedingungen und ein
	// Abrufdatum. Genau daran hängt, dass der Kasten in den Einstellungen
	// auch über ihn etwas sagt.
	{domain.DienstAnthropic, ""}: {"anthropic", anthropic.Vorgabeadresse},

	// **Der zweite, der keine Karten liefert** (A9, #1448).
	{domain.DienstOpenAI, ""}: {"openai", openai.Vorgabeadresse},
	{domain.DienstXAI, ""}:    {"xai", xai.Vorgabeadresse},
}

// paketFuer gibt das Paket eines Vertrags — die Abweichung je Frage vor der
// Vorgabe des Dienstes.
func paketFuer(vertrag domain.Dienstvertrag) (paketzeile, bool) {
	if zeile, kennt := klientenpakete[struct {
		dienst domain.Dienst
		frage  domain.Frage
	}{vertrag.Dienst, vertrag.Frage}]; kennt {
		return zeile, true
	}

	zeile, kennt := klientenpakete[struct {
		dienst domain.Dienst
		frage  domain.Frage
	}{vertrag.Dienst, ""}]

	return zeile, kennt
}

var abrufdatum = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

func TestJederVertragHaeltGegenSeinPaket(t *testing.T) {
	t.Parallel()

	quelltexte := map[string]string{}

	for _, vertrag := range domain.Dienstvertraege() {
		paket, kennt := paketFuer(vertrag)
		if !kennt {
			t.Errorf("%s hat kein Klientenpaket in klientenpakete", vertrag.Dienst)

			continue
		}

		if vertrag.Basisadresse != paket.adresse {
			t.Errorf("%s/%s nennt %q — das Paket %s sagt %q",
				vertrag.Dienst, vertrag.Frage, vertrag.Basisadresse, paket.paket, paket.adresse)
		}

		quelle, gelesen := quelltexte[paket.paket]
		if !gelesen {
			quelle = paketquelltext(t, paket.paket)
			quelltexte[paket.paket] = quelle
		}

		for _, endpunkt := range vertrag.Endpunkte {
			fest := endpunkt
			if stelle := strings.Index(endpunkt, "{"); stelle >= 0 {
				fest = endpunkt[:stelle]
			}

			if !strings.Contains(quelle, fest) {
				t.Errorf("%s/%s nennt den Endpunkt %q — %q steht nicht in internal/%s",
					vertrag.Dienst, vertrag.Frage, endpunkt, fest, paket.paket)
			}
		}

		for feld, adresse := range map[string]string{
			"Bedingungen": vertrag.Bedingungen, "Beleg": vertrag.Beleg,
		} {
			if !strings.HasPrefix(adresse, "https://") {
				t.Errorf("%s/%s: %s ist %q, keine prüfbare Adresse", vertrag.Dienst, vertrag.Frage, feld, adresse)
			}
		}

		if !abrufdatum.MatchString(vertrag.Abgerufen) {
			t.Errorf("%s/%s: Abgerufen ist %q", vertrag.Dienst, vertrag.Frage, vertrag.Abgerufen)
		}
	}
}

// paketquelltext gibt den Quelltext eines Pakets ohne Tests — als eine Zeichenkette.
func paketquelltext(t *testing.T, paket string) string {
	t.Helper()

	dateien, err := filepath.Glob(filepath.Join("..", "..", "internal", paket, "*.go"))
	if err != nil || len(dateien) == 0 {
		t.Fatalf("internal/%s lesen: %v (%d Dateien)", paket, err, len(dateien))
	}

	var gesamt strings.Builder

	for _, pfad := range dateien {
		if strings.HasSuffix(pfad, "_test.go") {
			continue
		}

		roh, lesefehler := os.ReadFile(pfad) //nolint:gosec // Quelldatei im Repositorium
		if lesefehler != nil {
			t.Fatalf("%s lesen: %v", pfad, lesefehler)
		}

		gesamt.Write(roh)
		gesamt.WriteByte('\n')
	}

	return gesamt.String()
}
