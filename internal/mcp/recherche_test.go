package mcp_test

// **Die Recherche über MCP** — Werkstück A16 von Phase 5.3
// (#1515).
//
// # Wogegen hier geprüft wird
//
// **Dass der Grund einer leeren Antwort ankommt.** Ein Modell, das „keine
// Treffer" sieht, antwortet aus dem Gedächtnis; eines, das „ohne-index"
// sieht, sagt, was fehlt.
//
// **Und dass es das Kontextpaket hier nicht gibt** — mit der Gegenprobe, dass
// die zwei Werkzeuge, die es geben soll, wirklich dastehen.

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// rechercheprobe hält, was die Werkzeuge bekommen sollen.
type rechercheprobe struct {
	ergebnis mcp.Rechercheergebnis
	fehler   error

	// gesehen hält, womit gefragt wurde.
	fall        string
	frage       string
	grenze      int
	sprache     string
	korpusstand string
	fortsetzung string
	karteID     int64

	// freigegeben hält, welche privaten Korpora freigegeben wurden — A13.4
	// (#1452). **Festgehalten und nicht verworfen**: eine Freigabe, die das
	// Werkzeug fallen ließe, sähe an der Antwort genauso aus wie eine, die
	// ankam.
	freigegeben []string
}

func (p *rechercheprobe) Recherche(
	_ context.Context, fall, frage string, grenze int,
	sprache, korpusstand, fortsetzung string, freigegeben []string,
) (mcp.Rechercheergebnis, error) {
	p.fall, p.frage, p.grenze = fall, frage, grenze
	p.sprache, p.korpusstand, p.fortsetzung = sprache, korpusstand, fortsetzung
	p.freigegeben = freigegeben

	return p.ergebnis, p.fehler
}

func (p *rechercheprobe) AehnlicheKarten(
	_ context.Context, karteID int64, grenze int, sprache, fortsetzung string,
	freigegeben []string,
) (mcp.Rechercheergebnis, error) {
	p.karteID, p.grenze = karteID, grenze
	p.sprache, p.fortsetzung = sprache, fortsetzung
	p.freigegeben = freigegeben

	return p.ergebnis, p.fehler
}

func TestOhneNahtGibtEsKeinRecherchewerkzeug(t *testing.T) {
	t.Parallel()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{}) {
		if werkzeug.Name == "lesen.recherche" ||
			werkzeug.Name == "lesen.aehnlichekarten" {
			t.Fatalf("%s stand ohne Naht da", werkzeug.Name)
		}
	}
}

// **Über MCP gibt es kein Kontextpaket — und die zwei anderen schon.**
//
// **Die Gegenprobe zum Nein** (§10): ohne sie prüfte der Lauf eine
// Werkzeugliste, die er gar nicht erreicht.
func TestUeberMCPGibtEsKeinKontextwerkzeug(t *testing.T) {
	t.Parallel()

	gefunden := map[string]bool{}
	for _, werkzeug := range mcp.Werkzeuge(
		mcp.Naehte{Recherche: &rechercheprobe{}}) {
		gefunden[werkzeug.Name] = true

		// **Gesucht wird das Paket und nicht das Wort.** `lesen.kontext`
		// gibt es längst — es nennt den **Arbeitskontext**, also wer gerade
		// arbeitet, und hat mit einem Kontextpaket nichts zu tun außer der
		// Silbe.
		if strings.Contains(werkzeug.Name, "kontext"+"paket") {
			t.Errorf("%s stellt über MCP ein Kontextpaket zusammen",
				werkzeug.Name)
		}
	}

	for _, name := range []string{"lesen.recherche", "lesen.aehnlichekarten"} {
		if !gefunden[name] {
			t.Errorf("%s fehlt — dann sagt der Lauf nichts über das Nein", name)
		}
	}
}

// **Der Grund einer leeren Antwort kommt an.**
func TestDasRecherchewerkzeugNenntDenGrund(t *testing.T) {
	t.Parallel()

	werkzeug := gesundheitswerkzeug(t, mcp.Naehte{
		Recherche: &rechercheprobe{ergebnis: mcp.Rechercheergebnis{
			Treffer: []mcp.Recherchetreffer{}, Modus: "ohne-index",
			Leere: "ohne-index", Grund: "Es ist kein Suchindex aktiv.",
		}},
	}, "lesen.recherche")

	antwort, fehler := werkzeug.Ruf(t.Context(),
		[]byte(`{"fall":"regeln","frage":"trampeln"}`))
	if fehler != nil {
		t.Fatalf("Ruf: %v", fehler)
	}

	if !strings.Contains(antwort, `"leere": "ohne-index"`) {
		t.Fatalf("Antwort %q", antwort)
	}
}

// **Die Frage kommt vollständig an — samt Filter und Zeiger.**
func TestDasRecherchewerkzeugReichtDieFrageDurch(t *testing.T) {
	t.Parallel()

	probe := &rechercheprobe{}

	werkzeug := gesundheitswerkzeug(t, mcp.Naehte{Recherche: probe},
		"lesen.recherche")

	if _, fehler := werkzeug.Ruf(t.Context(), []byte(
		`{"fall":"rulings","frage":"trampeln","grenze":3,"sprache":"de",`+
			`"korpusstand":"2026-08-01","fortsetzung":"zeiger"}`,
	)); fehler != nil {
		t.Fatalf("Ruf: %v", fehler)
	}

	if probe.fall != "rulings" || probe.frage != "trampeln" ||
		probe.grenze != 3 || probe.sprache != "de" ||
		probe.korpusstand != "2026-08-01" || probe.fortsetzung != "zeiger" {
		t.Fatalf("die Probe sah %+v", probe)
	}
}

// **Ohne Trefferzahl gilt zehn.**
func TestOhneTrefferzahlGiltZehn(t *testing.T) {
	t.Parallel()

	probe := &rechercheprobe{}

	if _, fehler := gesundheitswerkzeug(t, mcp.Naehte{Recherche: probe},
		"lesen.recherche").Ruf(t.Context(),
		[]byte(`{"frage":"trampeln"}`)); fehler != nil {
		t.Fatalf("Ruf: %v", fehler)
	}

	if probe.grenze != 10 {
		t.Fatalf("die Grenze ist %d", probe.grenze)
	}
}

// **Die Ähnlichkeitsfrage reicht die Karte durch.**
func TestDasAehnlichkeitswerkzeugReichtDieKarteDurch(t *testing.T) {
	t.Parallel()

	probe := &rechercheprobe{}

	if _, fehler := gesundheitswerkzeug(t, mcp.Naehte{Recherche: probe},
		"lesen.aehnlichekarten").Ruf(t.Context(),
		[]byte(`{"karte":42,"grenze":5,"sprache":"en","fortsetzung":"z"}`),
	); fehler != nil {
		t.Fatalf("Ruf: %v", fehler)
	}

	if probe.karteID != 42 || probe.grenze != 5 || probe.sprache != "en" ||
		probe.fortsetzung != "z" {
		t.Fatalf("die Probe sah %+v", probe)
	}
}

// **Ein Fehler der Naht kommt heraus — und keine halbe Antwort.**
func TestEinFehlerDerRecherchenahtKommtHeraus(t *testing.T) {
	t.Parallel()

	naehte := mcp.Naehte{Recherche: &rechercheprobe{fehler: errors.New("aus")}}

	for _, name := range []string{"lesen.recherche", "lesen.aehnlichekarten"} {
		if _, fehler := gesundheitswerkzeug(t, naehte, name).Ruf(t.Context(),
			[]byte(`{"frage":"x","karte":1}`)); fehler == nil {
			t.Errorf("%s kam trotz Fehler durch", name)
		}
	}
}

// **Und eine Eingabe, die kein JSON ist, ebenso.**
func TestEineUnlesbareRechercheeingabeIstEinFehler(t *testing.T) {
	t.Parallel()

	naehte := mcp.Naehte{Recherche: &rechercheprobe{}}

	for _, name := range []string{"lesen.recherche", "lesen.aehnlichekarten"} {
		if _, fehler := gesundheitswerkzeug(t, naehte, name).Ruf(t.Context(),
			[]byte(`kein json`)); fehler == nil {
			t.Errorf("%s las eine Eingabe, die keine ist", name)
		}
	}
}
