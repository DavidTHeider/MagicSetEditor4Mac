package ollama_test

// **Die Fehlerwege** — Werkstück A7 von Phase 5.3
// (#1446).
//
// # Wofür diese Datei da ist
//
// **Damit jeder betreten worden ist** (§10). Ein Fehlerweg, den nie jemand
// gegangen ist, ist eine Behauptung: er sieht aus, als fange er etwas ab, und
// niemand weiß, ob er es tut.

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/ollama"
)

// stummerServer antwortet auf jeden Weg mit demselben Status und Rumpf.
func stummerServer(t *testing.T, status int, rumpf string) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			schreiben.WriteHeader(status)
			_, _ = schreiben.Write([]byte(rumpf)) //nolint:errcheck // die Gegenstelle
		}))

	t.Cleanup(server.Close)

	return server
}

// **Eine Laufzeit, die nicht antwortet, meldet sich als Fehler.**
func TestEineStummeLaufzeitIstEinFehler(t *testing.T) {
	t.Parallel()

	klient := klientFuer(stummerServer(t, http.StatusInternalServerError, ""))

	if _, fehler := klient.Fassung(t.Context()); fehler == nil {
		t.Error("Fassung kam ohne Fehler durch")
	}

	if _, _, fehler := klient.Modelle(t.Context()); fehler == nil {
		t.Error("Modelle kam ohne Fehler durch")
	}

	if _, _, fehler := klient.Vermessen(t.Context(), "x"); fehler == nil {
		t.Error("Vermessen kam ohne Fehler durch")
	}

	if _, fehler := klient.Einbetter(domain.Vektorraum{Modell: "x", Vertrag: derProbevertrag()}).
		Einbetten(t.Context(), []string{"eins"}); fehler == nil {
		t.Error("Einbetten kam ohne Fehler durch")
	}
}

// **Und eine, die kein JSON schickt, ebenso.**
func TestEineUnlesbareAntwortIstEinFehler(t *testing.T) {
	t.Parallel()

	if _, fehler := klientFuer(stummerServer(t, http.StatusOK, "kein JSON")).
		Fassung(t.Context()); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("der Fehler ist %v", fehler)
	}
}

// **Eine Adresse, die keine ist, wird vor dem Ruf abgewiesen.**
func TestEineUnbrauchbareAdresseWirdVorherAbgewiesen(t *testing.T) {
	t.Parallel()

	if _, fehler := ollama.NeuerKlient("://", netz.KlientUeber(http.DefaultClient)).
		Fassung(t.Context()); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("der Fehler ist %v", fehler)
	}
}

// **Ohne Adresse wird die Schleifenadresse genommen.**
//
// # Warum der Lauf keine Verbindung braucht
//
// **Weil der Transport die Anfrage abfängt, bevor sie eine wird.** Geprüft
// wird, **wohin** dieses Paket greift — nicht, ob dort jemand ist. Ein Test,
// der wirklich auf 11434 klopfte, prüfte den Stand dieses Rechners (§3).
func TestOhneAdresseGiltDieSchleifenadresse(t *testing.T) {
	t.Parallel()

	horcher := &adresshorcher{}

	if _, fehler := ollama.NeuerKlient("", netz.KlientUeber(
		&http.Client{Transport: horcher})).Fassung(t.Context()); fehler == nil {
		t.Fatal("der abgefangene Ruf kam ohne Fehler durch")
	}

	if horcher.gesehen != ollama.Vorgabeadresse+ollama.Fassungsweg {
		t.Errorf("gegriffen wurde nach %q", horcher.gesehen)
	}
}

// adresshorcher schreibt die Adresse auf und lässt keine Verbindung zu.
type adresshorcher struct{ gesehen string }

func (a *adresshorcher) RoundTrip(gestellt *http.Request) (*http.Response, error) {
	a.gesehen = gestellt.URL.String()

	return nil, errors.New("hier wird nicht gewählt")
}

// **Die Uhr des Klienten steht im Beleg.**
//
// **Damit ein Beleg datierbar ist, ohne den Tag des Laufs zu nennen**: ein
// Test, der `time.Now` läse, wäre morgen ein anderer.
func TestDieUhrDesKlientenStehtImBeleg(t *testing.T) {
	t.Parallel()

	fest := time.Date(2026, 9, 11, 12, 0, 0, 0, time.UTC)

	modelle, _, fehler := klientFuer(gegenstelle(t)).
		MitUhr(func() time.Time { return fest }).Modelle(t.Context())
	if fehler != nil {
		t.Fatalf("Modelle: %v", fehler)
	}

	if datum := modelle[0].Geprueft[0].Datum; datum != "2026-09-11" {
		t.Errorf("das Datum ist %q", datum)
	}
}

// **Ein Schaufehler bricht die Modellliste ab.**
//
// **Und macht sie nicht kürzer**: eine Liste, der ein Modell fehlt, weil eine
// Anfrage scheiterte, sähe aus wie eine Liste, in der es nicht installiert ist.
func TestEinSchaufehlerBrichtDieListeAb(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, gestellt *http.Request) {
			switch gestellt.URL.Path {
			case ollama.Fassungsweg:
				_, _ = schreiben.Write([]byte(`{"version":"0.12.3"}`)) //nolint:errcheck // die Gegenstelle
			case ollama.Modellweg:
				schreibeDatei(t, schreiben, "tags.json")
			default:
				schreiben.WriteHeader(http.StatusInternalServerError)
			}
		}))
	t.Cleanup(server.Close)

	if _, _, fehler := klientFuer(server).Modelle(t.Context()); fehler == nil {
		t.Fatal("der Schaufehler kam ohne Fehler durch")
	}
}

// **Eine Modellliste, die keine ist, ebenso.**
func TestEineUnlesbareModelllisteIstEinFehler(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, gestellt *http.Request) {
			if gestellt.URL.Path == ollama.Fassungsweg {
				_, _ = schreiben.Write([]byte(`{"version":"0.12.3"}`)) //nolint:errcheck // die Gegenstelle

				return
			}

			schreiben.WriteHeader(http.StatusBadGateway)
		}))
	t.Cleanup(server.Close)

	if _, _, fehler := klientFuer(server).Modelle(t.Context()); fehler == nil {
		t.Fatal("die unlesbare Liste kam ohne Fehler durch")
	}
}

// **Eine fehlende Lizenz wird benannt und nicht weggelassen.**
//
// **Weil ein leeres Lizenzfeld die Herkunftskette ungültig machte** — A6
// verlangt sie je Glied. „Von Ollama nicht genannt" ist eine Auskunft; leer
// wäre ein Loch.
func TestEineFehlendeLizenzWirdBenannt(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, gestellt *http.Request) {
			switch gestellt.URL.Path {
			case ollama.Fassungsweg:
				_, _ = schreiben.Write([]byte(`{"version":"0.12.3"}`)) //nolint:errcheck // die Gegenstelle
			case ollama.Modellweg:
				schreibeDatei(t, schreiben, "tags.json")
			default:
				_, _ = schreiben.Write([]byte( //nolint:errcheck // die Gegenstelle
					`{"capabilities":["completion"]}`))
			}
		}))
	t.Cleanup(server.Close)

	modelle, _, fehler := klientFuer(server).Modelle(t.Context())
	if fehler != nil {
		t.Fatalf("Modelle: %v", fehler)
	}

	lizenz := modelle[0].Herkunft.Glieder[0].Lizenz

	if !strings.Contains(lizenz, "nicht genannt") {
		t.Errorf("die Lizenz ist %q", lizenz)
	}

	// **Und die Kette bleibt gültig** — das ist der Grund für die Zeile.
	if fehler := modelle[0].Herkunft.Pruefe(); fehler != nil {
		t.Errorf("die Kette ist ungültig: %v", fehler)
	}
}

// **Eine Systemanweisung geht als eigene Nachricht hinaus.**
func TestEineSystemanweisungGehtAlsEigeneNachricht(t *testing.T) {
	t.Parallel()

	gesehen := gesehenerRumpf(t)

	if _, fehler := klientFuer(gesehen.server).Rede(t.Context(), ollama.Frage{
		Alias: "gemma4:8b", System: "Du bist knapp.", Auftrag: "x",
	}); fehler != nil {
		t.Fatalf("Rede: %v", fehler)
	}

	rumpf := string(<-gesehen.rumpf)

	if !strings.Contains(rumpf, `"role":"system"`) ||
		!strings.Contains(rumpf, "Du bist knapp.") {
		t.Errorf("der Rumpf ist %s", rumpf)
	}
}

// **Der Einbetter gibt seinen Vertrag heraus.**
//
// **Damit ein Verbraucher den Fingerabdruck lesen kann**, ohne ihn selbst zu
// führen — A15 entscheidet daran über einen Indexneubau.
func TestDerEinbetterGibtSeinenVertrag(t *testing.T) {
	t.Parallel()

	vertrag := derProbevertrag()

	if gelesen := klientFuer(gegenstelle(t)).
		Einbetter(domain.Vektorraum{Modell: "x", Vertrag: vertrag}).Vertrag(); gelesen.Fingerabdruck() !=
		vertrag.Fingerabdruck() {
		t.Errorf("der Vertrag ist %+v", gelesen)
	}
}

// **Ein Cloudschalter, der kein Ja oder Nein ist, bleibt ungewiss.**
func TestEinUnlesbarerCloudschalterBleibtUngewiss(t *testing.T) {
	t.Parallel()

	befund := ollama.LiesCloudstand(einZuhauseMit(t,
		`{"disable_ollama_cloud": "vielleicht"}`))

	if befund.Stand != ollama.CloudUngewiss {
		t.Errorf("der Stand ist %q", befund.Stand)
	}

	if !strings.Contains(befund.Beleg, "vielleicht") {
		t.Errorf("der Beleg nennt den Wert nicht: %q", befund.Beleg)
	}
}

// **Ein Fehler der Laufzeit nennt ihren Grund** — gefunden in AI-02
// (#1743)
// am 2026-10-04: `mse ki-modell-pruefen ollama all-minilm:22m` endete mit
// „Ollama antwortet auf /api/chat mit 400", und Ollama hatte dazu
// geschrieben, warum — „does not support chat". Wer die Zeile las, wusste
// nicht, dass er ein Einbettungsmodell nach Text gefragt hatte.
//
// **Und nur das Feld `error`, gekürzt, auf einer Zeile**: ein fremder
// Rumpf gehört nicht ungelesen in eine Meldung.
func TestEinFehlerDerLaufzeitNenntIhrenGrund(t *testing.T) {
	t.Parallel()

	lang := strings.Repeat("x", 500)

	for name, fall := range map[string]struct{ rumpf, erwartet, nicht string }{
		"der Grund steht in der Meldung": {
			rumpf:    `{"error":"\"all-minilm:22m\" does not support chat"}`,
			erwartet: `mit 400: "all-minilm:22m" does not support chat`,
		},
		"ein Rumpf ohne Feld error bleibt draußen": {
			rumpf: `<html>kaputt</html>`, erwartet: "mit 400", nicht: "kaputt",
		},
		"ein leerer Grund bleibt draußen": {
			rumpf: `{"error":"  "}`, erwartet: "mit 400", nicht: "400:",
		},
		"ein langer Grund wird gekürzt, ein Umbruch geglättet": {
			rumpf: `{"error":"erste\nzweite ` + lang + `"}`, erwartet: "erste zweite xxx", nicht: lang,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(
				func(schreiben http.ResponseWriter, _ *http.Request) {
					schreiben.WriteHeader(http.StatusBadRequest)
					_, _ = schreiben.Write([]byte(fall.rumpf)) //nolint:errcheck // die Gegenstelle
				}))
			t.Cleanup(server.Close)

			_, fehler := klientFuer(server).Rede(t.Context(), ollama.Frage{Alias: "m", Auftrag: "x"})
			if fehler == nil || !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("Fehler %v, erwartet ErrValidation", fehler)
			}

			if !strings.Contains(fehler.Error(), fall.erwartet) {
				t.Errorf("die Meldung %q nennt nicht %q", fehler.Error(), fall.erwartet)
			}

			if fall.nicht != "" && strings.Contains(fehler.Error(), fall.nicht) {
				t.Errorf("die Meldung %q enthält %q", fehler.Error(), fall.nicht)
			}
		})
	}
}
