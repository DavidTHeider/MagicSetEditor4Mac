// **Die laufende Fassung wird über keine Fläche geholt**
// (#1100).
//
// # Warum dieser Fall eine eigene Datei hat
//
// Der erste Flächenlauf hatte nur **einen** Erfolgsfall, und der lieferte
// immer die künstlich neuere Fassung `99.9.9`. Damit belegte er Prüfsummen-
// und Zielortparität — nicht dieselbe **Fachentscheidung**. Und genau dort lag
// eine Abweichung: Fenster und Kommandozeile verglichen die Fassungen, der
// Dienst rief unmittelbar durch und lud auch dieselbe.
//
// Angestrichen vom Gesamtreview vor 0.2.0, samt der Begründung, warum der
// Wächter es nicht sah.
//
// # Was hier **nicht** gleich sein muss
//
// **Der Rückgabewert.** „Es gibt nichts zu holen" ist auf einer Kommandozeile
// kein Fehlschlag — `mse aktualisierung` sagt es und endet mit 0. Im Dienst
// ist es einer, denn ein `POST`, der nichts anlegt, hat keinen Ort zu nennen.
//
// **Gleich ist die Zusage**: es wird nichts abgelegt. Wer hier denselben Code
// verlangte, verlangte dieselbe **Aufsicht** statt derselben Fähigkeit
// ([0044](../../docs/entscheidungen/0044-was-ein-werkzeug-schreiben-darf.md)).
package flaechen_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

func TestDieLaufendeFassungWirdUeberKeineFlaecheGeholt(t *testing.T) {
	for name, versuchen := range map[string]func(*testing.T, []string){
		"Kommandozeile": gleichaufDerKommandozeile,
		"Dienst":        gleichImDienst,
	} {
		t.Run(name, func(t *testing.T) {
			umgebung := append(mitEigenerSaat(t),
				"MSE_FREIGABE_ADRESSE="+gleichaltrigeGegenstelle(t))

			ort := downloadverzeichnis(t, umgebung)

			versuchen(t, umgebung)

			eintraege, err := os.ReadDir(ort)
			if err != nil {
				t.Fatalf("lesen: %v", err)
			}

			if len(eintraege) != 0 {
				t.Errorf("es blieb etwas liegen: %+v", eintraege)
			}
		})
	}
}

// gleichaltrigeGegenstelle meldet **dieselbe** Fassung, die gerade läuft.
func gleichaltrigeGegenstelle(t *testing.T) string {
	t.Helper()

	dienst := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			if _, fehler := schreiben.Write([]byte(
				`{"tag_name":"v` + app.Fassung + `","assets":[]}`,
			)); fehler != nil {
				t.Errorf("schreiben: %v", fehler)
			}
		}))

	t.Cleanup(dienst.Close)

	return dienst.URL
}

// gleichaufDerKommandozeile erwartet eine Auskunft und keinen Ladevorgang.
func gleichaufDerKommandozeile(t *testing.T, umgebung []string) {
	t.Helper()

	aus, fehler, code := kommandozeile(t, umgebung, "aktualisierung", "--laden")
	if code != 0 {
		t.Fatalf("der Code ist %d: %s / %s", code, aus, fehler)
	}

	if !strings.Contains(aus, "laufende Fassung") {
		t.Errorf("die Ausgabe sagt es nicht: %s", aus)
	}
}

// gleichImDienst erwartet einen fachlichen Fehlschlag.
func gleichImDienst(t *testing.T, umgebung []string) {
	t.Helper()

	an := starteDienst(t, umgebung)

	status, rumpf := an.ruf(t, http.MethodPost, "/aktualisierung/laden", "")
	if status < 400 {
		t.Fatalf("POST /aktualisierung/laden: %d (%s)", status, rumpf)
	}
}
