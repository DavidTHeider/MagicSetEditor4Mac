//go:build fremdpakete

package app

import (
	"strings"
	"testing"
)

// **Die feste Kohorte trägt kein Band** — QA-02.AK-03
// (#1734),
// angestrichen vom Zweitblick über Phase 6.0. Ein Standardfehler und ein
// 95-%-Band setzen eine Zufallsstichprobe voraus; die feste Auswahl ist keine
// (`MESSMETHODIK.md`), und bis dahin schrieb der Kanonlauf trotzdem
// `mse2.band.unten` und `mse2.band.oben` in jeden Beleg. **Beschreibend bleibt
// sie**: Mittel, Median, Quantile — für den Verlauf.
func TestDieFesteKohorteTraegtKeinBand(t *testing.T) {
	t.Parallel()

	schaetzung := kanonschaetzung{
		schichten: 1, ganz: 100, gezogen: 10, mittel: 20,
		standardfehler: 0.5, unten: 19, oben: 21,
	}

	fest := strings.Join(append(schaetzung.alsZeilen(festeKohorte),
		schaetzung.kennzahlen("mse2", festeKohorte)...), "\n")

	for _, verboten := range []string{"Standardfehler", "95-%-Band", "mse2.standardfehler", "mse2.band."} {
		if strings.Contains(fest, verboten) {
			t.Errorf("die feste Kohorte nennt %q:\n%s", verboten, fest)
		}
	}

	for _, noetig := range []string{"kennzahl mse2.kanonmittel=20.0000", "kennzahl mse2.stichprobe=10", "feste Auswahl"} {
		if !strings.Contains(fest, noetig) {
			t.Errorf("die feste Kohorte nennt %q nicht:\n%s", noetig, fest)
		}
	}

	zufall := strings.Join(append(schaetzung.alsZeilen(zufallsKohorte),
		schaetzung.kennzahlen("zufall", zufallsKohorte)...), "\n")

	for _, noetig := range []string{
		"95-%-Band", "kennzahl zufall.standardfehler=0.5000",
		"kennzahl zufall.band.unten=19.0000", "kennzahl zufall.band.oben=21.0000",
	} {
		if !strings.Contains(zufall, noetig) {
			t.Errorf("die Zufallsstichprobe nennt %q nicht:\n%s", noetig, zufall)
		}
	}
}
