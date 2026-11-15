package openai_test

// **Der Bildweg von OpenAI** — Werkstück A12 von Phase 5.3
// (#1451).
//
// # Was hier gemessen wird
//
// **Dass Bytes ankommen und eine Adresse nicht.** Die Entscheidung steht in
// [0102](../../docs/entscheidungen/0102-wo-ein-erzeugtes-bild-liegt.md): ein
// Bild erreicht dieses Programm als Bytes in der Antwort oder gar nicht. Für
// die `gpt-image`-Familie ist das der dokumentierte Normalfall — kommt doch
// eine Adresse, ist das **Drift** und wird laut.
//
// **Und dass die zwei Wege zwei verschiedene Rümpfe haben.** Erzeugen schickt
// kein Bild, Bearbeiten schickt eine Daten-URI in `images[0].image_url` —
// nicht `image`, das ist xAI (0101 §8, 0102 §5).

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"image"
	"image/color"
	"image/png"
	"maps"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/openai"
)

// einPNG malt ein winziges echtes Bild.
func einPNG(t *testing.T) []byte {
	t.Helper()

	flaeche := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	flaeche.Set(0, 0, color.NRGBA{R: 255, A: 255})

	var aus bytes.Buffer
	if fehler := png.Encode(&aus, flaeche); fehler != nil {
		t.Fatalf("das Prüfbild ließ sich nicht schreiben: %v", fehler)
	}

	return aus.Bytes()
}

// einBilddienst nimmt den Rumpf entgegen und antwortet, wie gesagt.
func einBilddienst(
	t *testing.T, antwort string, gesehen map[string]any, weg *string,
) *openai.Klient {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, gestellt *http.Request) {
			if weg != nil {
				*weg = gestellt.URL.Path
			}

			if gesehen != nil {
				var rumpf map[string]any
				if fehler := json.NewDecoder(gestellt.Body).Decode(&rumpf); fehler != nil {
					t.Errorf("den Rumpf lesen: %v", fehler)
				}

				// **Hineingeschrieben statt zurückgegeben**: der Aufrufer
				// hält die Karte und liest sie nach dem Ruf.
				maps.Copy(gesehen, rumpf)
			}

			_, _ = schreiben.Write([]byte(antwort)) //nolint:errcheck // die Gegenstelle
		}))

	t.Cleanup(server.Close)

	return openai.NeuerKlient(server.URL, netz.KlientUeber(http.DefaultClient),
		openai.Zugang{Schluessel: "geheim"})
}

// **Ein erzeugtes Bild kommt als Bytes zurück.**
func TestEinErzeugtesBildKommtAlsBytesZurueck(t *testing.T) {
	t.Parallel()

	bild := einPNG(t)

	var weg string

	gesehen := map[string]any{}

	klient := einBilddienst(t, `{"created": 1, "output_format": "png", "data": [
		{"b64_json": "`+base64.StdEncoding.EncodeToString(bild)+`",
		 "revised_prompt": "ein Blitz über einer Ebene, Ölgemälde"}],
		"usage": {"input_tokens": 12, "output_tokens": 300, "total_tokens": 312}}`,
		gesehen, &weg)

	roh, fehler := klient.Bild(t.Context(), openai.Bildauftrag{
		Alias: "gpt-image-2.5-flare", Auftrag: "ein Blitz", Groesse: "1024x1024",
	})
	if fehler != nil {
		t.Fatalf("Bild: %v", fehler)
	}

	if !bytes.Equal(roh.Inhalt, bild) {
		t.Errorf("es kamen %d Bytes statt %d", len(roh.Inhalt), len(bild))
	}

	if roh.Medientyp != "image/png" {
		t.Errorf("der Medientyp ist %q", roh.Medientyp)
	}

	// **Die Umschrift ist Herkunft.** Der Dienst darf den Prompt umschreiben;
	// was gezeichnet wurde, ist dann nicht, was ein Mensch getippt hat.
	if roh.Umschrift != "ein Blitz über einer Ebene, Ölgemälde" {
		t.Errorf("die Umschrift ist %q", roh.Umschrift)
	}

	if weg != "/v1/images/generations" {
		t.Errorf("der Weg war %q", weg)
	}

	// **Erzeugen schickt kein Bild.**
	if _, dabei := gesehen["images"]; dabei {
		t.Error("der Erzeugungsruf trug ein Eingangsbild")
	}

	// **Und fragt nicht nach einer Adresse.** `response_format` nimmt diese
	// Modellfamilie gar nicht an (0102 §1); es zu schicken wäre ein 400.
	if _, dabei := gesehen["response_format"]; dabei {
		t.Error("der Ruf trug response_format")
	}
}

// **Eine Adresse statt der Bytes ist Drift und wird laut.**
//
// # Warum kein Holweg
//
// **Der Bearbeitungsweg schickt die Daten-URI — und nie eine Adresse.**
func TestDerBearbeitungswegSchicktDieDatenUri(t *testing.T) {
	t.Parallel()

	bild := einPNG(t)

	var weg string

	gesehen := map[string]any{}

	klient := einBilddienst(t, `{"created": 1, "data": [{"b64_json": "`+
		base64.StdEncoding.EncodeToString(bild)+`"}]}`, gesehen, &weg)

	if _, fehler := klient.Bild(t.Context(), openai.Bildauftrag{
		Alias: "gpt-image-2.5-sunburst", Auftrag: "mach es dunkler",
		Quelle: bild,
	}); fehler != nil {
		t.Fatalf("Bild: %v", fehler)
	}

	if weg != "/v1/images/edits" {
		t.Fatalf("der Weg war %q", weg)
	}

	// **`images` ist eine Liste** — bei xAI heißt dasselbe Feld `image` und
	// ist ein Objekt (0102 §5). Wer den einen Vertrag für den anderen hält,
	// bekommt einen Fehler, der nach einem kaputten Bild aussieht.
	liste, ist := gesehen["images"].([]any)
	if !ist || len(liste) != 1 {
		t.Fatalf("images ist %#v", gesehen["images"])
	}

	eintrag, ist := liste[0].(map[string]any)
	if !ist {
		t.Fatalf("der Eintrag ist %#v", liste[0])
	}

	adresse, ist := eintrag["image_url"].(string)
	if !ist || !strings.HasPrefix(adresse, "data:image/png;base64,") {
		t.Fatalf("image_url ist %#v", eintrag["image_url"])
	}

	// **Und keine Dateikennung.** Sie setzte `/v1/files` voraus — einen Weg,
	// der beim Anbieter etwas ablegt (0102 §5).
	if _, dabei := eintrag["file_id"]; dabei {
		t.Error("der Ruf trug eine Dateikennung")
	}
}

// **Eine Quelle, die kein Bild ist, geht gar nicht erst hinaus.**
func TestEineQuelleOhneBildsignaturGehtNichtHinaus(t *testing.T) {
	t.Parallel()

	gefragt := false

	server := httptest.NewServer(http.HandlerFunc(
		func(_ http.ResponseWriter, _ *http.Request) { gefragt = true }))

	t.Cleanup(server.Close)

	klient := openai.NeuerKlient(server.URL, netz.KlientUeber(http.DefaultClient),
		openai.Zugang{Schluessel: "geheim"})

	_, fehler := klient.Bild(t.Context(), openai.Bildauftrag{
		Alias: "gpt-image-2.5-sunburst", Auftrag: "mach es dunkler",
		Quelle: []byte("kein Bild"),
	})
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("die Quelle kam durch: %v", fehler)
	}

	if gefragt {
		t.Error("es ging trotzdem eine Anfrage hinaus")
	}
}

// **Ohne Prompt und ohne Alias geht nichts hinaus.**
func TestOhneAuftragGehtNichtsHinaus(t *testing.T) {
	t.Parallel()

	for name, auftrag := range map[string]openai.Bildauftrag{
		"ohne Alias":   {Auftrag: "ein Blitz"},
		"ohne Prompt":  {Alias: "gpt-image-2.5-flare"},
		"nur Leerraum": {Alias: "gpt-image-2.5-flare", Auftrag: "   "},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			gefragt := false

			server := httptest.NewServer(http.HandlerFunc(
				func(_ http.ResponseWriter, _ *http.Request) { gefragt = true }))

			t.Cleanup(server.Close)

			klient := openai.NeuerKlient(server.URL,
				netz.KlientUeber(http.DefaultClient),
				openai.Zugang{Schluessel: "geheim"})

			if _, fehler := klient.Bild(t.Context(), auftrag); !errors.Is(
				fehler, domain.ErrValidation) {
				t.Fatalf("%q kam durch: %v", name, fehler)
			}

			if gefragt {
				t.Error("es ging trotzdem eine Anfrage hinaus")
			}
		})
	}
}

// **Die Positivliste hat jetzt vier Wege — und keiner legt etwas ab.**
func TestDiePositivlisteHatVierWege(t *testing.T) {
	t.Parallel()

	wege := openai.Wege()

	if len(wege) != 4 {
		t.Fatalf("die Wege sind %v", wege)
	}

	for _, weg := range wege {
		if !strings.HasPrefix(weg, "/v1/") {
			t.Errorf("%q sieht nicht wie ein Weg dieser Schnittstelle aus", weg)
		}

		if strings.Contains(weg, "/files") || strings.Contains(weg, "/batches") {
			t.Errorf("%q legt beim Anbieter etwas ab", weg)
		}
	}
}

// **Nennt die Antwort ein Modell, gilt ihres.**
//
// # Warum das eine eigene Zeile wert ist
//
// **Die Signatur schlägt die Angabe des Dienstes.**
//
// # Der Fall, für den das gebaut ist
//
// **Und ohne Signatur bleibt die Angabe — oder nichts.**
//
// **Ein gescheiterter Bildauftrag wird nicht wiederholt.**
//
// # Warum das ein eigener Lauf ist
//
// **Weil ein zweiter Ruf ein zweites Mal Geld kostet.** Das Issue sagt es
// als Abnahme: „Unklar abgerechneter oder abgebrochener Bildrequest löst
// keinen automatischen zweiten Kostenrequest aus." Ein 500 ist genau dieser
// Fall — der Dienst kann gerechnet und abgerechnet haben, bevor er stolperte.
//
