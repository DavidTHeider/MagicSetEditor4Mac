package werkzeuge_test

// **Bezug und Quelltext an derselben, öffentlichen Stelle** — D01
// (#2331),
// Vermerk [0190](../../docs/entscheidungen/0190-das-repositorium-wird-oeffentlich.md).
//
// # Wogegen hier geprüft wird
//
// **Gegen einen Bezugsweg, den nur erreicht, wer schon Zugang hat** —
// Befund A-08: die Aktualisierungsprüfung fragte ein privates Repositorium,
// und ein Gast bekam 404. Ein Quelltextangebot, das auf einen privaten Ort,
// eine Adresse der Programmierschnittstelle oder einen SSH-Zugang zeigt, ist
// dasselbe in anderer Gestalt.
//
// # Was er hält
//
//   - Jeder Verweis im Abschnitt „Bezug und Quelltext" von `docs/RELEASE.md`
//     und im Vermerk zeigt auf die Freigabeseite dieses Repositoriums — und
//     zwar auf dieselbe, die `aktualisierung.Repositorium` fragt.
//   - `scripts/release.sh` legt das Quelltextpaket **vor** `shasum` an, nimmt
//     es in `pruefsummen.txt` auf (und damit unter die Signatur) und nennt es
//     im Vorschlag für `gh release create`.
//
// # Was er nicht hält
//
//   - **Dass die Seite erreichbar ist.** Das ist sie, seit das Repositorium
//     am 2026-11-15 öffentlich wurde — belegt ohne Zugang am selben Tag
//     (Vermerk 0190, Nachtrag); ein Lauf, der das Netz fragt, gehört nicht in
//     die Pipeline.
//   - **Dass das Paket vollständig ist.** Was darin liegt, prüft
//     `scripts/releasequelltext.test.sh` an einem nachgestellten Baum.

import (
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/aktualisierung"
)

// kanal ist die Freigabeseite — aus dem Klienten, nicht abgeschrieben.
var kanal = "https://github.com/" + aktualisierung.Repositorium + "/releases"

// bezugsadressen findet die Adressen in spitzen Klammern.
var bezugsadressen = regexp.MustCompile(`<([a-z]+:[^>\s]+|git@[^>\s]+)>`)

// bezugsfehler sagt, warum ein Verweis kein öffentlicher Bezugsweg ist — oder
// gibt "".
func bezugsfehler(verweis string) string {
	adresse, err := url.Parse(verweis)

	switch {
	case err != nil || adresse.Scheme == "":
		return "keine Adresse — etwa ein SSH-Zugang, den nur erreicht, wer einen Schlüssel hat"
	case adresse.Scheme != "https":
		return "nicht über https"
	case adresse.User != nil || adresse.RawQuery != "":
		return "trägt Anmeldung oder Zusatz"
	case adresse.Host != "github.com":
		return "nicht die Freigabeseite — " + adresse.Host + " verlangt für ein Repositorium ohne Zugang eine Anmeldung"
	case verweis != kanal && !strings.HasPrefix(verweis, kanal+"/"):
		return "nicht unter " + kanal
	}

	return ""
}

func TestEinBezugswegIstOeffentlichOderFalsch(t *testing.T) {
	t.Parallel()

	for verweis, oeffentlich := range map[string]bool{
		kanal:                 true,
		kanal + "/tag/v0.9.0": true,
		kanal + "/download/v0.9.0/MagicSetEditor4Mac-0.9.0-quelltext.tar.gz":                 true,
		"git@github.com:" + aktualisierung.Repositorium + ".git":                             false,
		"https://api.github.com/repos/" + aktualisierung.Repositorium + "/releases/assets/1": false,
		"https://github.com/" + aktualisierung.Repositorium + "/settings":                    false,
		"http://github.com/" + aktualisierung.Repositorium + "/releases":                     false,
		kanal + "?token=geheim":                      false,
		"https://github.com/jemand/anderes/releases": false,
	} {
		if grund := bezugsfehler(verweis); (grund == "") != oeffentlich {
			t.Errorf("%s: öffentlich erwartet %v, Grund %q", verweis, oeffentlich, grund)
		}
	}
}

// **Der Vertrag steht an der Freigabe und nicht nur im Vermerk.**
func TestDasQuelltextangebotStehtAnDerFreigabe(t *testing.T) {
	t.Parallel()

	vermerk := lies(t, "docs/entscheidungen/0190-das-repositorium-wird-oeffentlich.md")

	for _, muss := range []string{kanal, "Herausgeber", "Verantwortlich", "Signatur", "-quelltext.tar.gz"} {
		if !strings.Contains(vermerk, muss) {
			t.Errorf("der Vermerk 0190 nennt %q nicht", muss)
		}
	}

	_, abschnitt, gefunden := strings.Cut(lies(t, "docs/RELEASE.md"), "\n## Bezug und Quelltext\n")
	if !gefunden {
		t.Fatal("docs/RELEASE.md hat keinen Abschnitt »Bezug und Quelltext«")
	}

	abschnitt, _, _ = strings.Cut(abschnitt, "\n## ")

	// **Die Bezugswege sind die Adressen in spitzen Klammern** — die Verweise
	// auf Issues und Vermerke daneben sind Belege und kein Bezug.
	gesehen := 0

	for _, treffer := range bezugsadressen.FindAllStringSubmatch(abschnitt, -1) {
		gesehen++

		if grund := bezugsfehler(treffer[1]); grund != "" {
			t.Errorf("Bezug und Quelltext: %s — %s", treffer[1], grund)
		}
	}

	if gesehen < 2 || !strings.Contains(abschnitt, "-quelltext.tar.gz") {
		t.Errorf("der Abschnitt nennt nicht Freigabeseite, Fassungsseite und "+
			"Quelltextpaket (%d Bezugswege)", gesehen)
	}

	freigabe := lies(t, "scripts/release.sh")
	quelltext := strings.Index(freigabe, "quelltext_schreiben ")
	summen := strings.Index(freigabe, "shasum -a 256")

	switch {
	case !strings.Contains(freigabe, `source "$WURZEL/scripts/releasequelltext.sh"`):
		t.Error("scripts/release.sh bindet releasequelltext.sh nicht ein")
	case quelltext < 0 || summen < 0 || quelltext > summen:
		t.Error("scripts/release.sh legt das Quelltextpaket nicht vor den Prüfsummen an — es stünde nicht unter der Signatur")
	case !regexp.MustCompile(`shasum -a 256 [^\n]*\./\*\.tar\.gz`).MatchString(freigabe):
		t.Error("pruefsummen.txt nimmt das Quelltextpaket nicht auf")
	case !regexp.MustCompile(`gh release create[^\n]*tar\.gz`).MatchString(freigabe):
		t.Error("der Vorschlag für gh release create lässt das Quelltextpaket weg")
	}
}
