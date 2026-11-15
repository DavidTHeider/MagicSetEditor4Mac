package openai_test

// **Wenn die Bildantwort nicht die dokumentierte ist** — Werkstück A12 von
// Phase 5.3
// (#1451).
//
// # Was hier gemessen wird
//
// **Drift, und zwar laut** (§16). Die Dokumentation sagt für die
// `gpt-image`-Familie Bytes zu; kommt etwas anderes — eine Adresse, ein
// leeres Feld, ein Fehler im 200er-Rumpf —, dann ist das kein Bild und wird
// auch nicht so behandelt.
//
// **Und dass nichts zweimal kostet.** Ein 500 kann bedeuten, dass gerechnet
// und abgerechnet wurde; wiederholt wird deshalb nichts von selbst.

import (
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/openai"
)

// **Weil es ihn nicht gibt** ([0102](../../docs/entscheidungen/0102-wo-ein-erzeugtes-bild-liegt.md)
// §4). Wer die Adresse holte, hätte eine Frist, einen Umleiter und einen
// Schlüssel, der mitreisen könnte — für einen Fall, den die Dokumentation
// für diese Modellfamilie ausschließt.
func TestEineAdresseStattDerBytesIstDrift(t *testing.T) {
	t.Parallel()

	klient := einBilddienst(t,
		`{"created": 1, "data": [{"url": "https://beispiel.invalid/bild.png"}]}`,
		nil, nil)

	_, fehler := klient.Bild(t.Context(), openai.Bildauftrag{
		Alias: "gpt-image-2.5-flare", Auftrag: "ein Blitz",
	})
	if !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("die Adresse kam durch: %v", fehler)
	}

	if !strings.Contains(fehler.Error(), "Adresse") {
		t.Errorf("die Meldung nennt die Adresse nicht: %s", fehler.Error())
	}
}

// **Eine Antwort ohne Bild ist ein Fehler und keine leere Datei.**
func TestEineAntwortOhneBildIstEinFehler(t *testing.T) {
	t.Parallel()

	for name, antwort := range map[string]string{
		"gar keine Liste":  `{"created": 1}`,
		"eine leere Liste": `{"created": 1, "data": []}`,
		"leeres Base64":    `{"created": 1, "data": [{"b64_json": ""}]}`,
		"kaputtes Base64":  `{"created": 1, "data": [{"b64_json": "!!!"}]}`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			klient := einBilddienst(t, antwort, nil, nil)

			if _, fehler := klient.Bild(t.Context(), openai.Bildauftrag{
				Alias: "gpt-image-2.5-flare", Auftrag: "ein Blitz",
			}); fehler == nil {
				t.Fatalf("%q kam durch", name)
			}
		})
	}
}

// **Ein Fehler im 200er-Rumpf ist ein Fehler.**
//
// Derselbe Fall wie bei der Responses-API: die Anfrage ist angekommen, der
// Lauf misslungen. Wer ihn als leere Antwort führte, hätte ein Bild ohne
// Bytes und keine Ahnung, woran es lag.
func TestEinFehlerImRumpfIstEinFehler(t *testing.T) {
	t.Parallel()

	klient := einBilddienst(t, `{"created": 1, "error":
		{"code": "moderation_blocked", "message": "so nicht"}}`, nil, nil)

	_, fehler := klient.Bild(t.Context(), openai.Bildauftrag{
		Alias: "gpt-image-2.5-flare", Auftrag: "ein Blitz",
	})
	if !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("der Fehler kam durch: %v", fehler)
	}

	if !strings.Contains(fehler.Error(), "moderation_blocked") {
		t.Errorf("die Meldung nennt den Code nicht: %s", fehler.Error())
	}
}

// **Weil ein Alias kein Modell ist.** Wer `gpt-image-2.5-flare` ruft, bekommt
// den Stand, den der Dienst heute darunter führt; was geantwortet hat, steht
// in der Antwort. Die Herkunft soll das Zweite nennen und nicht das Erste
// (A11, #1450).
func TestDasModellDerAntwortSchlaegtDenAlias(t *testing.T) {
	t.Parallel()

	bild := einPNG(t)

	klient := einBilddienst(t, `{"created": 1, "model": "gpt-image-2.5-flare-2026-09-08",
		"data": [{"b64_json": "`+base64.StdEncoding.EncodeToString(bild)+`"}]}`,
		nil, nil)

	roh, fehler := klient.Bild(t.Context(), openai.Bildauftrag{
		Alias: "gpt-image-2.5-flare", Auftrag: "ein Blitz",
	})
	if fehler != nil {
		t.Fatalf("Bild: %v", fehler)
	}

	if roh.Modell != "gpt-image-2.5-flare-2026-09-08" {
		t.Fatalf("das Modell ist %q", roh.Modell)
	}
}

// **`output_format` sagt `png`, und es kommt ein JPEG.** Wer der Angabe
// glaubte, legte ein JPEG unter `.png` ab — und der Vorrat, der die Signatur
// liest, widerspräche der Ablage, die dem Namen glaubt.
func TestDieSignaturSchlaegtDieAngabe(t *testing.T) {
	t.Parallel()

	jpeg := []byte("\xff\xd8\xff\xe0 und was ein JPEG sonst so hat")

	klient := einBilddienst(t, `{"created": 1, "output_format": "png",
		"data": [{"b64_json": "`+base64.StdEncoding.EncodeToString(jpeg)+`"}]}`,
		nil, nil)

	roh, fehler := klient.Bild(t.Context(), openai.Bildauftrag{
		Alias: "gpt-image-2.5-flare", Auftrag: "ein Blitz",
	})
	if fehler != nil {
		t.Fatalf("Bild: %v", fehler)
	}

	if roh.Medientyp != "image/jpeg" {
		t.Fatalf("der Medientyp ist %q statt image/jpeg", roh.Medientyp)
	}
}

// **Abgewiesen wird hier nicht.** Was keine Bildsignatur trägt, weist der
// Vorrat ab ([service.Bildvorrat.Legen]); zweimal dieselbe Prüfung an zwei
// Stellen liefe irgendwann auseinander. Dieses Paket sagt nur, was es sieht.
func TestOhneSignaturBleibtDieAngabe(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		form     string
		erwartet string
	}{
		"mit Angabe":  {"webp", "image/webp"},
		"ohne Angabe": {"", ""},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			rumpf := `{"created": 1, "data": [{"b64_json": "` +
				base64.StdEncoding.EncodeToString([]byte("weder noch")) + `"}]`
			if fall.form != "" {
				rumpf += `, "output_format": "` + fall.form + `"`
			}

			roh, fehler := einBilddienst(t, rumpf+"}", nil, nil).Bild(
				t.Context(), openai.Bildauftrag{
					Alias: "gpt-image-2.5-flare", Auftrag: "ein Blitz",
				})
			if fehler != nil {
				t.Fatalf("Bild: %v", fehler)
			}

			if roh.Medientyp != fall.erwartet {
				t.Fatalf("der Medientyp ist %q statt %q", roh.Medientyp, fall.erwartet)
			}
		})
	}
}

// **Gezählt wird, nicht behauptet**: der Lauf sieht genau eine Anfrage.
func TestEinGescheiterterBildauftragWirdNichtWiederholt(t *testing.T) {
	t.Parallel()

	rufe := 0

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			rufe++

			schreiben.WriteHeader(http.StatusInternalServerError)
		}))

	t.Cleanup(server.Close)

	klient := openai.NeuerKlient(server.URL, netz.KlientUeber(http.DefaultClient),
		openai.Zugang{Schluessel: "geheim"})

	if _, fehler := klient.Bild(t.Context(), openai.Bildauftrag{
		Alias: "gpt-image-2.5-flare", Auftrag: "ein Blitz",
	}); fehler == nil {
		t.Fatal("der 500er kam als Erfolg zurück")
	}

	if rufe != 1 {
		t.Fatalf("es gingen %d Anfragen hinaus statt einer", rufe)
	}
}
