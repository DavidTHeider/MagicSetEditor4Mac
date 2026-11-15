package app

// **Die eingestellte Adresse erreicht den Klienten** — Werkstück 3 von A18
// (#1533).
//
// # Was hier geprüft wird
//
// **Dass aus der Einstellung eine Adresse am Bauplan wird.** Vorher rief
// `cloudnaehte` jede Naht mit `""`, und der Klient nahm seine
// `Vorgabeadresse` — ein Gateway war nicht erreichbar, und eine Einstellung
// dafür gab es nicht.
//
// **Und dass leer die Vorgabe bleibt.** Wer nichts einstellt, spricht mit dem
// Anbieter; ein Umbau, der das änderte, schickte jeden bestehenden Nutzer
// woandershin.

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/openai"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// adressstand meldet einen festen Einstellungsstand.
type adressstand struct {
	Einstellungsfuehrer

	stand map[string]service.Wahl
}

func (a adressstand) Stand(context.Context) (map[string]service.Wahl, error) {
	return a.stand, nil
}

// mitDienstadresse baut eine Anwendung, deren Bauplan die Adresse mitschreibt.
func mitDienstadresse(t *testing.T, adresse string) (
	anwendung *App, gesehen *[]string,
) {
	t.Helper()

	gesehen = &[]string{}

	anwendung = NeueApp("")
	anwendung.MitEinstellungen(adressstand{stand: map[string]service.Wahl{
		domain.DienstEinstellung(domain.DienstOpenAI): {Wirksam: domain.DienstAn},
		domain.DienstAdresse(domain.DienstOpenAI):     {Wirksam: adresse},
	}})
	anwendung.MitKIAnbietern(
		func(string) KILaufzeitquelle { return stummeQuelle{} },
		map[string]func(adresse string) KILaufzeitquelle{
			// **Sie gibt eine Quelle zurück und nicht `nil`**: ein
			// Bauplan, der nichts baut, sähe in einem Lauf wie ein
			// fehlender aus.
			//
			//nolint:unparam // die Rückgabe verlangt die Schnittstelle; geprüft wird das Argument
			openaianbieter: func(gerufen string) KILaufzeitquelle {
				*gesehen = append(*gesehen, gerufen)

				return stummeQuelle{}
			},
		})

	return anwendung, gesehen
}

// **Eine eingestellte Adresse kommt am Bauplan an.**
func TestEineEingestellteDienstadresseKommtAmBauplanAn(t *testing.T) {
	t.Parallel()

	const gateway = "https://gateway.firma.intern/v1"

	anwendung, gesehen := mitDienstadresse(t, gateway)

	anbieter, fehler := anwendung.kianbieter(t.Context())
	if fehler != nil {
		t.Fatalf("kianbieter: %v", fehler)
	}

	for _, einer := range anbieter {
		if einer.name != openaianbieter {
			continue
		}

		if einer.adresse != gateway {
			t.Fatalf("die Liste nennt %q", einer.adresse)
		}

		einer.bauen()
	}

	if len(*gesehen) != 1 || (*gesehen)[0] != gateway {
		t.Fatalf("der Bauplan bekam %v", *gesehen)
	}
}

// **Und ohne Einstellung bleibt es bei der Vorgabe des Anbieters.**
//
// **Die Gegenprobe** (§10): ein Bauplan, der immer die eingestellte Adresse
// bekäme, hielte den Lauf darüber, und jeder bestehende Nutzer spräche nach
// diesem Umbau mit niemandem.
func TestOhneEinstellungBleibtDieVorgabedesAnbieters(t *testing.T) {
	t.Parallel()

	anwendung, gesehen := mitDienstadresse(t, "")

	anbieter, fehler := anwendung.kianbieter(t.Context())
	if fehler != nil {
		t.Fatalf("kianbieter: %v", fehler)
	}

	for _, einer := range anbieter {
		if einer.name == openaianbieter {
			if einer.adresse != openai.Vorgabeadresse {
				t.Fatalf("die Liste nennt %q statt der Vorgabe %q",
					einer.adresse, openai.Vorgabeadresse)
			}

			einer.bauen()
		}
	}

	if len(*gesehen) != 1 || (*gesehen)[0] != openai.Vorgabeadresse {
		t.Fatalf("der Bauplan bekam %v", *gesehen)
	}
}

// **Ein Schlüssel für ein Gateway geht nicht an den Anbieter.**
//
// # Der Fall, um den es geht
//
// **Jemand stellt ein Firmengateway ein, trägt dort seinen Schlüssel ein und
// leert die Adresse später wieder.** Ohne diese Trennung ginge derselbe
// Schlüssel danach an `api.openai.com` — eine Weitergabe an jemanden, für den
// er nie ausgestellt wurde.
func TestEinSchluesselFuerEinGatewayGehtNichtAnDenAnbieter(t *testing.T) {
	t.Parallel()

	const gateway = "https://gateway.firma.intern/v1"

	buch := neueZugangsmitschrift()

	anwendung := NeueApp("")
	anwendung.bund = buch
	anwendung.marktbund = buch

	// **Der Schlüssel wird unter der Gatewayadresse abgelegt.**
	if fehler := buch.Schreibe(t.Context(), Schluesselbunddienst,
		domain.ZugangKontoFuer(domain.DienstOpenAI, domain.ZugangSchluessel,
			gateway), "geheim-fuers-gateway"); fehler != nil {
		t.Fatalf("ablegen: %v", fehler)
	}

	if gefunden := anwendung.cloudschluesselFuer(domain.DienstOpenAI,
		gateway); gefunden != "geheim-fuers-gateway" {
		t.Fatalf("das Gateway bekam %q", gefunden)
	}

	// **Und der Anbieter bekommt ihn nicht.**
	if gefunden := anwendung.cloudschluesselFuer(domain.DienstOpenAI,
		""); gefunden != "" {
		t.Fatalf("der Anbieter bekam den Schlüssel des Gateways: %q", gefunden)
	}

	// **Auch ein zweites Gateway nicht.**
	if gefunden := anwendung.cloudschluesselFuer(domain.DienstOpenAI,
		"https://openrouter.ai/api/v1"); gefunden != "" {
		t.Fatalf("ein fremdes Gateway bekam den Schlüssel: %q", gefunden)
	}
}

// **Und ein Schlüssel ohne eingestellte Adresse erreicht den Anbieter weiter.**
//
// **Die Gegenprobe** (§10): eine Trennung, die auch die Vorgabe verschöbe,
// ließe jeden eingetragenen Schlüssel nach diesem Umbau verschwinden — ohne
// dass jemand wüsste, warum.
func TestOhneAdresseErreichtDerSchluesselDenAnbieterWeiter(t *testing.T) {
	t.Parallel()

	buch := neueZugangsmitschrift()

	anwendung := NeueApp("")
	anwendung.bund = buch
	anwendung.marktbund = buch

	if antwort := anwendung.ZugangSetzen(t.Context(),
		string(domain.DienstOpenAI), domain.ZugangSchluessel,
		"geheim"); antwort.Fehler != nil {
		t.Fatalf("ZugangSetzen: %+v", antwort.Fehler)
	}

	if gefunden := anwendung.cloudschluesselFuer(domain.DienstOpenAI,
		""); gefunden != "geheim" {
		t.Fatalf("der Anbieter bekam %q — geschrieben nach %v",
			gefunden, buch.geschrieben)
	}
}

// **Die Freigabe nennt die Adresse, nicht nur den Anbieter.**
//
// # Warum
//
// **Seit A18 kann hinter „openai" ein Gateway stehen** — ein Firmengateway,
// OpenRouter, Azure. Das ist ein **Datenexport an einen Dritten**, auch wenn
// dahinter derselbe Anbieter rechnet. Eine Freigabe, die nur den Namen
// zeigte, verschwiege genau den Unterschied, um den es bei ihr geht.
func TestDieFreigabeNenntDieWirksameAdresse(t *testing.T) {
	t.Parallel()

	const gateway = "https://gateway.firma.intern/v1"

	anwendung := NeueApp("")
	anwendung.MitEinstellungen(adressstand{stand: map[string]service.Wahl{
		domain.DienstAdresse(domain.DienstOpenAI): {Wirksam: gateway},
	}})
	anwendung.MitRecherche(&stummeRecherche{})

	antwort := anwendung.KontextZusammenstellen(t.Context(), "a", "",
		string(domain.FallRegeln), "b", 5, string(domain.DienstOpenAI),
		"gpt", 4096, 1024, 0, nil)

	if antwort.Fehler != nil {
		t.Fatalf("KontextZusammenstellen: %+v", antwort.Fehler)
	}

	if antwort.Freigabe.Zieladresse != gateway {
		t.Fatalf("die Freigabe nennt %q", antwort.Freigabe.Zieladresse)
	}
}

// **Und ein Gateway auf der Schleife rechnet auf diesem Rechner.**
//
// **Die Gegenprobe** (§10): vorher stand für jeden Cloudanbieter ein blankes
// „Cloud" da — richtig, solange die Adresse fest war, und falsch, sobald sie
// einstellbar ist. Ein LiteLLM auf `127.0.0.1` ist ein lokaler Vermittler.
func TestEinGatewayAufDerSchleifeRechnetHier(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")
	anwendung.MitEinstellungen(adressstand{stand: map[string]service.Wahl{
		domain.DienstAdresse(domain.DienstOpenAI): {
			Wirksam: "http://127.0.0.1:4000/v1",
		},
	}})

	if ort := zielOrt(t, anwendung, t.Context(),
		string(domain.DienstOpenAI)); ort != domain.KIOrtLokal {
		t.Fatalf("der Ort ist %q", ort)
	}

	// **Und ohne Einstellung bleibt es Cloud** — die Vorgabe des Anbieters
	// liegt im Netz.
	leer := NeueApp("")
	leer.MitEinstellungen(adressstand{stand: map[string]service.Wahl{}})

	if ort := zielOrt(t, leer, t.Context(),
		string(domain.DienstOpenAI)); ort != domain.KIOrtCloud {
		t.Fatalf("ohne Einstellung ist der Ort %q", ort)
	}
}

// stummeRecherche gibt ein leeres Paket ohne Fehler.
type stummeRecherche struct{}

func (stummeRecherche) Recherchieren(
	context.Context, domain.Rechercheanfrage, string,
) (domain.Rechercheergebnis, error) {
	return domain.Rechercheergebnis{}, nil
}

func (stummeRecherche) Zusammenstellen(
	_ context.Context, wunsch service.Kontextwunsch,
) (service.Kontextpaket, error) {
	return service.Kontextpaket{Manifest: domain.Kontextmanifest{
		PaketID:      "kp-probe",
		Zielanbieter: wunsch.Zielanbieter,
		Zielmodell:   wunsch.Zielmodell,
	}}, nil
}

// stummeQuelle ist eine Laufzeitquelle, die nichts kann und da ist.
//
// **Da zu sein ist ihre Aufgabe**: ein Bauplan, der `nil` gäbe, sähe in einem
// Lauf wie ein fehlender aus — und die Läufe hier prüfen, **womit** er
// gerufen wurde, nicht was er baut.
type stummeQuelle struct{}

func (stummeQuelle) Fassung(context.Context) (string, error) { return "", nil }

func (stummeQuelle) Modelle(context.Context) (
	[]domain.KIModell, []domain.Befund, error,
) {
	return nil, nil, nil
}

func (stummeQuelle) Vorschlagen(
	context.Context, string, string, string, map[string]any,
) (KIRohantwort, error) {
	return KIRohantwort{}, nil
}
