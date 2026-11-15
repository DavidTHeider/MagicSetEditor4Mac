package domain_test

// **Die Kennung, die ein Anbieter braucht, lässt sich eintragen** —
// Nacharbeit zu Werkstück 4 von Phase 4.12
// (#1351).
//
// # Der Befund
//
// **Die Anmeldung über einen Anbieter war unerreichbar.** Die Oberfläche sagt
// „registriert dort eine eigene Anwendung und trägt ihre Kennung in den
// Einstellungen ein"; `app.zugangsquelleUeber` liest dafür
// `anbieter.<name>.kennung` — und `PruefeEinstellung` wies genau diesen
// Schlüssel ab: „ist keine Einstellung dieser Anwendung."
//
// **Damit ließ sich die Kennung an keiner der vier Flächen setzen**, und der
// ganze Anmeldeweg konnte nie in Betrieb gehen. Gefunden vom Auftraggeber am
// 2026-09-06 an der laufenden Anwendung — er suchte die Stelle, auf die der
// Satz verweist, und fand nichts.
//
// **Das ist §10s erste Dimension**: löst der Beitrag ein, wofür es die
// Anwendung gibt — und woran ist das zu sehen? Vier Anbieter, ein Horcher auf
// der Schleifenadresse, PKCE, `state` und `nonce`, und **kein Weg hinein**.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Jeder Anbieter hat seine Kennung als Einstellung.**
//
// **Über [domain.Anbieter] und nicht über eine abgeschriebene Liste**: eine
// zweite Aufzählung liefe beim fünften Anbieter auseinander, und niemand sähe
// es — die Falle, die in `einstellung.go` schon zweimal zugeschnappt ist.
func TestJederAnbieterHatSeineKennung(t *testing.T) {
	t.Parallel()

	for _, anbieter := range domain.Anbieter() {
		schluessel := domain.AnbieterKennung(anbieter)

		if fehler := domain.PruefeEinstellung(schluessel, "Ov23xyz"); fehler != nil {
			t.Errorf("%s: %v", schluessel, fehler)
		}

		if bereich := domain.EinstellungsBereich(schluessel); bereich != domain.BereichAnmeldung {
			t.Errorf("%s steht im Bereich %q", schluessel, bereich)
		}
	}
}

// **Und ein Anbieter, den es nicht gibt, bekommt keine.**
//
// **Die Gegenprobe** (§10): ohne sie wäre eine Prüfung, die **jeden**
// Schlüssel durchlässt, ebenso grün — und das wäre schlimmer als die Lücke,
// die hier geschlossen wird.
func TestEinErfundenerAnbieterBekommtKeineKennung(t *testing.T) {
	t.Parallel()

	if fehler := domain.PruefeEinstellung(
		domain.AnbieterKennung("erfunden"), "x"); fehler == nil {
		t.Error("ein erfundener Anbieter bekam eine Kennung")
	}
}

// **Die Kennung ist freier Text und hat keine Werteliste.**
//
// **Weil jeder Anbieter eine andere Gestalt vergibt**: GitHub schreibt
// `Ov23…`, Google eine Adresse unter `apps.googleusercontent.com`, Microsoft
// eine UUID. Eine Prüfung darauf wäre eine Regel, die dieses Programm sich
// ausdenkt — und die erste Änderung beim Anbieter machte jede gültige Kennung
// unbrauchbar (§11: was in einem fremden System gilt, weiß das fremde System).
func TestDieAnbieterkennungIstFreierText(t *testing.T) {
	t.Parallel()

	schluessel := domain.AnbieterKennung(domain.Anbieter()[0])

	for _, wert := range []string{
		"Ov23liAbCdEf",
		"123456789-abc.apps.googleusercontent.com",
		"11111111-2222-3333-4444-555555555555",
	} {
		if fehler := domain.PruefeEinstellung(schluessel, wert); fehler != nil {
			t.Errorf("%q wurde abgewiesen: %v", wert, fehler)
		}
	}
}
