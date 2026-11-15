package domain_test

// **Die Basisadresse je Dienst** — Werkstück 3 von A18
// (#1533).
//
// # Warum es sie braucht
//
// **Weil die Adressen der Cloudanbieter fest verdrahtet waren.**
// `kianbieterliste.go` rief alle drei Nähte mit `""`, und der Klient nahm dann
// seine `Vorgabeadresse`. Ein Gateway — LiteLLM, OpenRouter, Azure, ein
// Firmengateway — war damit nicht erreichbar, und ein billiger Testendpunkt
// ebenso wenig.
//
// **Dieselbe Gestalt wie `ki.adresse`** für Ollama: leer heißt „die Vorgabe
// des Anbieters", und eine fremde Adresse ist ein **Datenexport an einen
// Dritten**, auch wenn dahinter derselbe Anbieter steht.

import (
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Jeder Dienst mit einer einstellbaren Adresse hat seine Einstellung.**
func TestJederDienstMitAdresseHatSeineEinstellung(t *testing.T) {
	t.Parallel()

	reihe := domain.Einstellungsreihe()

	for _, dienst := range domain.Dienste() {
		if !dienst.AdresseEinstellbar() {
			continue
		}

		schluessel := domain.DienstAdresse(dienst)

		if !slices.Contains(reihe, schluessel) {
			t.Errorf("%q steht nicht in der Einstellungsreihe — dann lässt "+
				"sich die Adresse von %q nirgends setzen", schluessel, dienst)
		}
	}
}

// **Und die drei Cloudanbieter sind darunter.**
//
// **Der benannte Fall neben der Klasse**: ohne ihn wäre der Lauf darüber auch
// grün, wenn **kein** Dienst eine einstellbare Adresse hätte.
func TestDieDreiCloudanbieterHabenEineEinstellbareAdresse(t *testing.T) {
	t.Parallel()

	for _, dienst := range []domain.Dienst{
		domain.DienstAnthropic, domain.DienstOpenAI, domain.DienstXAI,
	} {
		if !dienst.AdresseEinstellbar() {
			t.Errorf("%q hat keine einstellbare Adresse — dann ist kein "+
				"Gateway erreichbar", dienst)
		}
	}
}

// **Der Schlüssel nennt den Dienst und nicht die Angabeart.**
func TestDerAdressschluesselNenntDenDienst(t *testing.T) {
	t.Parallel()

	if gefunden := domain.DienstAdresse(domain.DienstOpenAI); gefunden !=
		"dienst.openai.adresse" {
		t.Fatalf("der Schlüssel ist %q", gefunden)
	}
}

// **Eine Adresse wird geprüft — und leer heißt: die Vorgabe des Anbieters.**
func TestEineDienstadresseWirdGeprueft(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		wert  string
		haelt bool
	}{
		"leer heißt Vorgabe": {"", true},
		"ein Gateway":        {"https://gateway.firma.intern/v1", true},
		"Klartext im Netz":   {"http://gpu.intern:8000", true},
		"ohne Schema":        {"gateway.firma.intern", false},
		"ein fremdes Schema": {"ftp://gateway.firma.intern", false},
		"ohne Rechner":       {"https://", false},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fehler := domain.PruefeEinstellung(
				domain.DienstAdresse(domain.DienstOpenAI), fall.wert)

			if fall.haelt != (fehler == nil) {
				t.Fatalf("Pruefe(%q): %v", fall.wert, fehler)
			}
		})
	}
}

// **Eine fremde Adresse stuft den Dienst als nicht-lokal ein.**
//
// **Das ist die Zusage, wegen der es die Einstellung überhaupt geben darf**:
// ein Gateway ist ein **Datenexport an einen Dritten**, auch wenn dahinter
// derselbe Anbieter steht. Wer das verschwiege, machte aus einer Bequemlichkeit
// eine stille Weitergabe.
func TestEineFremdeDienstadresseIstEinDatenexport(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		adresse string
		ort     domain.KIOrt
	}{
		"die Vorgabe des Anbieters": {"", domain.KIOrtCloud},
		"ein Gateway im Netz":       {"https://gateway.firma.intern", domain.KIOrtCloud},
		"auf diesem Rechner":        {"http://127.0.0.1:4000", domain.KIOrtLokal},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if gefunden := domain.OrtFuerDienstadresse(fall.adresse); gefunden != fall.ort {
				t.Fatalf("der Ort ist %q, erwartet %q", gefunden, fall.ort)
			}
		})
	}
}

// **Ein Dienst ohne einstellbare Adresse hat auch keine Einstellung.**
//
// **Die Gegenprobe** (§10): eine Reihe, die für **jeden** Dienst eine Zeile
// trüge, hielte den Lauf ganz oben, ohne dass `AdresseEinstellbar` etwas
// entschiede.
func TestEinDienstOhneEinstellbareAdresseHatKeineEinstellung(t *testing.T) {
	t.Parallel()

	ohne := 0

	for _, dienst := range domain.Dienste() {
		if dienst.AdresseEinstellbar() {
			continue
		}

		ohne++

		if slices.Contains(domain.Einstellungsreihe(),
			domain.DienstAdresse(dienst)) {
			t.Errorf("%q hat keine einstellbare Adresse und trotzdem eine "+
				"Einstellung", dienst)
		}
	}

	if ohne == 0 {
		t.Fatal("jeder Dienst hat eine einstellbare Adresse — dann sagt " +
			"dieser Lauf nichts")
	}
}

// **Und der Schlüsselbund trennt nach Adresse.**
//
// # Warum das keine Feinheit ist
//
// **Ein Schlüssel für ein Gateway ist kein Schlüssel für den Anbieter.** Wer
// `openai.adresse` auf ein Firmengateway stellt, dort seinen Schlüssel
// einträgt und die Adresse später leert, schickte denselben Schlüssel an
// `api.openai.com` — eine Weitergabe an jemanden, für den er nie ausgestellt
// wurde.
func TestDerZugangTrenntNachAdresse(t *testing.T) {
	t.Parallel()

	vorgabe := domain.ZugangKontoFuer(domain.DienstOpenAI,
		domain.ZugangSchluessel, "")
	gateway := domain.ZugangKontoFuer(domain.DienstOpenAI,
		domain.ZugangSchluessel, "https://gateway.firma.intern/v1")

	if vorgabe == gateway {
		t.Fatalf("beide liegen unter %q — ein Schlüssel für das Gateway ginge "+
			"damit an den Anbieter", vorgabe)
	}

	// **Die Vorgabe behält ihren alten Ort**, damit ein eingetragener
	// Schlüssel den Umbau überlebt.
	if vorgabe != domain.ZugangKonto(domain.DienstOpenAI, domain.ZugangSchluessel) {
		t.Errorf("die Vorgabe liegt jetzt unter %q statt unter %q", vorgabe,
			domain.ZugangKonto(domain.DienstOpenAI, domain.ZugangSchluessel))
	}

	// **Und zwei Gateways sind zwei Orte.**
	anderes := domain.ZugangKontoFuer(domain.DienstOpenAI,
		domain.ZugangSchluessel, "https://openrouter.ai/api/v1")
	if gateway == anderes {
		t.Errorf("zwei Gateways liegen unter demselben Ort %q", gateway)
	}

	// **Der Ort bleibt ein Kontoname** — kein Schrägstrich, kein Leerzeichen.
	for _, ort := range []string{vorgabe, gateway, anderes} {
		if strings.ContainsAny(ort, " /") {
			t.Errorf("%q taugt nicht als Kontoname", ort)
		}
	}
}

// **Ein Dienst, den es nicht gibt, hat keine einstellbare Adresse.**
//
// **Der Zweig, den keine Aufzählung betritt** (§3): `AdresseEinstellbar`
// sucht in den Dienstzeilen, und wer nicht darin steht, bekommt `false`.
// Erreichbar ist er über eine Kennung aus einer Fläche — dort kommt ein
// Dienstname als Zeichenkette an.
func TestEinUnbekannterDienstHatKeineEinstellbareAdresse(t *testing.T) {
	t.Parallel()

	if domain.Dienst("erfunden").AdresseEinstellbar() {
		t.Fatal("ein erfundener Dienst darf hinter ein Gateway")
	}
}

// **Der Ort im Schlüsselbund lässt sich dort auch ablegen.**
//
// # Der Befund
//
// **`@` ist im Schlüsselbund kein erlaubtes Zeichen.** Die erste Fassung von
// [domain.ZugangKontoFuer] baute `openai.schluessel@6794ca4973d2`, und
// `schluesselbund.Schreibe` wies es ab — der Gateway-Schlüssel ließ sich
// **gar nicht ablegen**, und die Fähigkeit aus Werkstück 3 von A18 war für
// genau den Fall unbenutzbar, für den es sie gibt.
//
// **Mein eigener Lauf sah es nicht**, weil er auf Leerzeichen und
// Schrägstriche prüfte — ein Zeichensatz, den ich mir ausgedacht habe, statt
// den zu nehmen, der gilt. Gefunden hat es die Sichtprüfung, beim Versuch,
// einen Probeschlüssel abzulegen.
//
// **Der erlaubte Satz steht in `internal/schluesselbund`** und wird hier
// nachgebildet — `domain` kennt das Paket nicht (§2), und eine Kante dorthin
// wäre teurer als diese Zeile. Dass beide gleich bleiben, hält
// `internal/werkzeuge/zugangsortzeichen_test.go`.
func TestDerZugangsortTraegtNurErlaubteZeichen(t *testing.T) {
	t.Parallel()

	erlaubt := func(zeichen rune) bool {
		return zeichen >= 'a' && zeichen <= 'z' ||
			zeichen >= 'A' && zeichen <= 'Z' ||
			zeichen >= '0' && zeichen <= '9' ||
			zeichen == '.' || zeichen == '-' || zeichen == '_'
	}

	for name, adresse := range map[string]string{
		"die Vorgabe":      "",
		"ein Gateway":      "https://gateway.firma.intern/v1",
		"mit Port":         "http://127.0.0.1:4000/v1",
		"mit Abfrage":      "https://openrouter.ai/api/v1?x=1",
		"mit Sonderzeiche": "https://gäteway.example/v1",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ort := domain.ZugangKontoFuer(domain.DienstOpenAI,
				domain.ZugangSchluessel, adresse)

			for _, zeichen := range ort {
				if !erlaubt(zeichen) {
					t.Fatalf("%q trägt %q — das weist der Schlüsselbund ab",
						ort, string(zeichen))
				}
			}
		})
	}
}
