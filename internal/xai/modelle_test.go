package xai_test

// **Was xAI über seine Modelle sagt** — Werkstück A10 von Phase 5.3
// (#1449).
//
// # Der Fund, um den sich dieses Werkstück dreht
//
// **xAI sagt mehr als OpenAI und weniger als Anthropic.**
// `/v1/language-models` nennt je Modell `input_modalities` und
// `output_modalities` — damit stehen **Text** und **Bildlesen** fest.
// **Strukturierte Ausgabe steht nicht dabei**, obwohl die Schnittstelle sie
// kann; sie bleibt deshalb ungeprüft und braucht die Messung aus A9.
//
// **Und das ist der Grund, warum dieses Werkstück nicht von OpenAI abschreibt**
// (so steht es im Issue): derselbe Anbieter hat zwei Endpunkte mit **zwei
// verschiedenen** Bildformen, und seine Modellliste hat eine Gestalt, die es
// bei keinem anderen gibt.

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/xai"
)

// **Aus den Modalitäten werden belegte Fähigkeiten.**
func TestAusDenModalitaetenWerdenBelegteFaehigkeiten(t *testing.T) {
	t.Parallel()

	klient := xai.NeuerKlient(eineGegenstelle(t).URL,
		netz.KlientUeber(http.DefaultClient), xai.Zugang{Schluessel: "geheim"})

	modelle, _, fehler := klient.Modelle(t.Context())
	if fehler != nil {
		t.Fatalf("Modelle: %v", fehler)
	}

	if len(modelle) != 1 {
		t.Fatalf("die Modelle sind %+v", modelle)
	}

	modell := modelle[0]

	if !modell.Kann(domain.KITextgenerierung) {
		t.Error("»text« unter den Ausgaben belegt keine Textgenerierung")
	}

	if !modell.Kann(domain.KIVision) {
		t.Error("»image« unter den Eingaben belegt kein Bildlesen")
	}

	// **Und die feste Form gerade nicht.** Die Liste sagt nichts darüber,
	// und was nicht dasteht, ist nicht belegt (0097) — dafür gibt es die
	// Messung aus A9.
	if modell.Kann(domain.KIStrukturierteGenerierung) {
		t.Error("die feste Form gilt als belegt, obwohl die Liste nichts sagt")
	}

	// **Der Beleg nennt das Feld, aus dem er kommt.**
	for _, beleg := range modell.Geprueft {
		if !strings.Contains(beleg.Beleg, "modalities") ||
			beleg.Datum != "2025-04-04" {
			t.Errorf("der Beleg ist %+v", beleg)
		}
	}
}

// **Ein Modell, das keinen Text ausgibt, gehört nicht in diesen Katalog.**
//
// **Und es verschwindet nicht still**: es wird ein Befund. Ein Videomodell
// gibt es, und diese Anwendung kann damit nichts anfangen — das ist zweierlei,
// und beides gehört gesagt.
func TestEinModellOhneTextausgabeWirdEinBefund(t *testing.T) {
	t.Parallel()

	klient := xai.NeuerKlient(eineGegenstelle(t).URL,
		netz.KlientUeber(http.DefaultClient), xai.Zugang{Schluessel: "geheim"})

	_, befunde, fehler := klient.Modelle(t.Context())
	if fehler != nil {
		t.Fatalf("Modelle: %v", fehler)
	}

	if len(befunde) != 1 || befunde[0].Datei != "grok-stumm" {
		t.Fatalf("die Befunde sind %+v", befunde)
	}

	if !strings.Contains(befunde[0].Meldung, "video") {
		t.Errorf("der Befund nennt die Ausgabe nicht: %q", befunde[0].Meldung)
	}
}

// **Fingerabdruck und Fassung reisen im Beleg mit.**
//
// # Warum nicht in der Herkunft
//
// **Weil eine Kette Lizenz und Dateiprüfsummen verlangt** ([domain.KIGlied]),
// und xAI nennt beides nicht. Eine halbe Kette ist schlimmer als keine
// (0097) — die Herkunft bleibt deshalb leer.
//
// **Der Fingerabdruck ist trotzdem das Nächste, was ein Anbieter bisher zu
// »was läuft hier wirklich« gesagt hat.** Wer den Katalog liest, sieht ihn;
// ändert er sich, liegt hinter demselben Namen etwas anderes.
func TestFingerabdruckUndFassungReisenImBelegMit(t *testing.T) {
	t.Parallel()

	klient := xai.NeuerKlient(eineGegenstelle(t).URL,
		netz.KlientUeber(http.DefaultClient), xai.Zugang{Schluessel: "geheim"})

	modelle, _, fehler := klient.Modelle(t.Context())
	if fehler != nil {
		t.Fatalf("Modelle: %v", fehler)
	}

	if !modelle[0].Herkunft.Unbekannt() {
		t.Errorf("die Herkunft ist %+v", modelle[0].Herkunft)
	}

	if !strings.Contains(modelle[0].Laufzeit, "fp_0123456789") ||
		!strings.Contains(modelle[0].Laufzeit, "1.0.0") {
		t.Errorf("die Laufzeit ist %q", modelle[0].Laufzeit)
	}
}

// **Die Namen daneben stehen im Beleg und werden keine Einträge.**
//
// **Weil es dasselbe Modell ist.** Drei Zeilen für einen Fingerabdruck wären
// drei Entscheidungen für dieselbe Sache — und wer `grok-probe-latest` wählt,
// bekommt heute dieses und morgen ein anderes.
func TestDieNamenDanebenWerdenKeineEintraege(t *testing.T) {
	t.Parallel()

	klient := xai.NeuerKlient(eineGegenstelle(t).URL,
		netz.KlientUeber(http.DefaultClient), xai.Zugang{Schluessel: "geheim"})

	modelle, _, fehler := klient.Modelle(t.Context())
	if fehler != nil {
		t.Fatalf("Modelle: %v", fehler)
	}

	aliase := make([]string, 0, len(modelle))
	for _, modell := range modelle {
		aliase = append(aliase, modell.Alias)
	}

	if slices.Contains(aliase, "grok-probe-latest") {
		t.Errorf("ein Zweitname wurde ein Eintrag: %v", aliase)
	}

	if !strings.Contains(modelle[0].Geprueft[0].Beleg, "grok-probe-latest") {
		t.Errorf("der Beleg nennt die Zweitnamen nicht: %q",
			modelle[0].Geprueft[0].Beleg)
	}
}

// eineGegenstelle antwortet wie die Schnittstelle von xAI.
func eineGegenstelle(t *testing.T) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, gestellt *http.Request) {
			if strings.HasSuffix(gestellt.URL.Path, "language-models") {
				schreibeDatei(t, schreiben, "modelle.json")

				return
			}

			schreibeDatei(t, schreiben, "antwort.json")
		}))

	t.Cleanup(server.Close)

	return server
}

// schreibeDatei schickt eine Datei aus `testdaten`.
func schreibeDatei(t *testing.T, schreiben http.ResponseWriter, name string) {
	t.Helper()

	// **Der Name kommt aus dem Lauf, nicht von außen** — ein fester Ordner
	// im Repositorium, und jeder Aufrufer steht daneben.
	roh, fehler := os.ReadFile(filepath.Join("testdaten", name)) //nolint:gosec // siehe oben
	if fehler != nil {
		t.Fatalf("%s lesen: %v", name, fehler)
	}

	_, _ = schreiben.Write(roh) //nolint:errcheck // die Gegenstelle
}

// **Ein Modell ohne lesbares Datum bleibt draußen — und wird ein Befund.**
//
// # Warum das nicht der Sparsamkeit wegen geprüft wird
//
// **Ein Beleg ohne Tag sagt nicht, seit wann er gilt** — und
// `service.NeuerModellkatalog` weist ihn ab. Wer ihn durchließe, verlöre
// **den ganzen Katalog** dieses Anbieters an einem einzigen Eintrag; genau
// das ist in A8 bei Anthropic passiert.
func TestEinModellOhneLesbaresDatumWirdEinBefund(t *testing.T) {
	t.Parallel()

	modelle, befunde := ausDieserListe(t,
		`{"models":[{"id":"grok-zeitlos","created":0,`+
			`"input_modalities":["text"],"output_modalities":["text"]}]}`)

	if len(modelle) != 0 {
		t.Fatalf("die Modelle sind %+v", modelle)
	}

	if len(befunde) != 1 || !strings.Contains(befunde[0].Meldung, "created") {
		t.Fatalf("die Befunde sind %+v", befunde)
	}
}

// **Ein Modell ohne Zweitnamen und Fingerabdruck kommt ohne Klammer durch.**
//
// **Der Fall ist der Normalfall bei einem neuen Modell** — und eine leere
// Klammer im Beleg wäre eine Auskunft, die nichts sagt.
func TestEinModellOhneZweitnamenKommtOhneKlammerDurch(t *testing.T) {
	t.Parallel()

	modelle, _ := ausDieserListe(t,
		`{"models":[{"id":"grok-nackt","created":1743724800,`+
			`"input_modalities":["text"],"output_modalities":["text"]}]}`)

	if len(modelle) != 1 {
		t.Fatalf("die Modelle sind %+v", modelle)
	}

	if strings.Contains(modelle[0].Geprueft[0].Beleg, "(") {
		t.Errorf("der Beleg trägt eine leere Klammer: %q",
			modelle[0].Geprueft[0].Beleg)
	}

	if modelle[0].Laufzeit != "xAI grok-nackt" {
		t.Errorf("die Laufzeit ist %q", modelle[0].Laufzeit)
	}
}

// ausDieserListe fragt einen Dienst, der genau diese Liste führt.
func ausDieserListe(t *testing.T, liste string) (
	[]domain.KIModell, []domain.Befund,
) {
	t.Helper()

	klient := xai.NeuerKlient(einDienstDer(t, http.StatusOK, liste),
		netz.KlientUeber(http.DefaultClient), xai.Zugang{Schluessel: "geheim"})

	modelle, befunde, fehler := klient.Modelle(t.Context())
	if fehler != nil {
		t.Fatalf("Modelle: %v", fehler)
	}

	return modelle, befunde
}
