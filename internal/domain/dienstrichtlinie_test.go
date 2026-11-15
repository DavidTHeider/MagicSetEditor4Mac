package domain_test

// **Ein Dienstschalter ist eine Richtlinie, seine Adresse nicht** —
// #1540.
//
// # Der Befund
//
// **Eine eingestellte Anbieteradresse galt nicht.** `findeRichtlinie` nahm
// **jeden** Schlüssel unter `dienst.` als An/Aus-Schranke — die Regel entstand,
// als dort nur Schalter lagen. Werkstück 3 von A18
// (#1533)
// hat Adressen danebengelegt, und `http://127.0.0.1:8099` steht in keiner
// Reihung `{aus, an}`: der Wert fiel auf die Vorgabe zurück, und die Naht baute
// den Klienten mit `https://api.openai.com`.
//
// **Kein Lauf sah es**, weil die Go-Läufe den Stand als fertige Karte setzen
// und die Frontendläufe die Naht belegen — die Faltung dazwischen prüfte
// niemand an einem freien Wert.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Eine gesetzte Anbieteradresse ist wirksam.**
func TestEineGesetzteDienstadresseIstWirksam(t *testing.T) {
	t.Parallel()

	const adresse = "http://127.0.0.1:8099"

	schluessel := domain.DienstAdresse(domain.DienstOpenAI)

	stand := domain.WirksamerStand(
		map[string]string{schluessel: adresse}, map[string]string{})

	geltung := stand[schluessel]

	if geltung.Wirksam != adresse {
		t.Errorf("wirksam ist %q, gewählt %q — erwartet beides %q",
			geltung.Wirksam, geltung.Gewaehlt, adresse)
	}

	if geltung.Herkunft != domain.GeltungProfil {
		t.Errorf("die Herkunft ist %q, erwartet %q",
			geltung.Herkunft, domain.GeltungProfil)
	}
}

// **Und der Schalter daneben bleibt eine Richtlinie.**
//
// **Die Gegenprobe zur Zeile darüber** (§10): wer die Präfixregel ganz
// entfernte, machte den Lauf oben grün — und nähme jedem Dienstschalter die
// Schranke, ohne dass es jemand merkte.
func TestEinDienstschalterBleibtEineRichtlinie(t *testing.T) {
	t.Parallel()

	for _, dienst := range domain.Dienste() {
		schalter := domain.DienstEinstellung(dienst)

		if !domain.IstRichtlinie(schalter) {
			t.Errorf("%q ist keine Richtlinie mehr", schalter)
		}

		if adresse := domain.DienstAdresse(dienst); domain.IstRichtlinie(adresse) {
			t.Errorf("%q gilt als Richtlinie — eine Adresse ist keine", adresse)
		}
	}
}

// **Ein Import kann einen eingeschalteten Dienst nicht einschalten lassen.**
//
// **Das ist die Zusage, die die Präfixregel trägt**: `aus` ist strenger als
// `an`, und wer eine Datei einliest, lockert nichts. Sie steht hier neben dem
// Lauf, der die Adresse ausnimmt — damit sichtbar bleibt, was die Ausnahme
// **nicht** anfasst.
func TestEinNutzerLockertKeinenDienstschalter(t *testing.T) {
	t.Parallel()

	schalter := domain.DienstEinstellung(domain.DienstOpenAI)

	stand := domain.WirksamerStand(
		map[string]string{schalter: domain.DienstAus},
		map[string]string{schalter: domain.DienstAn})

	if stand[schalter].Wirksam != domain.DienstAus {
		t.Errorf("wirksam ist %q — der Mensch hat die Schranke gelockert",
			stand[schalter].Wirksam)
	}
}
