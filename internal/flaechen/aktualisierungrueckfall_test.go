package flaechen_test

// **Eine alte Freigabe unter neuem Tag, über jede Fläche** — SEC03
// (#2356),
// Vermerk 0194, die Zeile „Rollback" der Signaturmatrix.
//
// Die Gegenstelle nennt `v99.9.9` und liefert die Prüfsummen, die Signatur
// und das Abbild einer **anderen** Fassung — alle drei passen zueinander, und
// die Signatur ist gültig. Vor SEC03 lud jede Fläche das Abbild. Die Prüfung
// steht im Klienten, an der einen Stelle, die alle Flächen nehmen; dieser
// Lauf hält, dass keine an ihr vorbeigeht (§16).

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
)

// rueckfallgegenstelle bietet unter `v99.9.9` das echt signierte Abbild von
// 0.7.0 an.
func rueckfallgegenstelle(t *testing.T) string {
	t.Helper()

	inhalt := []byte(strings.Repeat("alt", 3000))
	summe := sha256.Sum256(inhalt)
	summen := hex.EncodeToString(summe[:]) + "  ./MagicSetEditor4Mac-0.7.0.dmg\n"
	signatur := probeSigniert(t, summen)

	dienst := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			var rumpf []byte

			switch {
			case strings.HasSuffix(anfrage.URL.Path, "/latest"):
				eigen := "http://" + anfrage.Host

				type beigabe struct {
					Name    string `json:"name"`
					Groesse int    `json:"size"`
					Adresse string `json:"url"`
				}

				if fehler := json.NewEncoder(schreiben).Encode(struct {
					Marke    string    `json:"tag_name"`
					Seite    string    `json:"html_url"`
					Beigaben []beigabe `json:"assets"`
				}{
					Marke: "v99.9.9",
					Seite: "https://beispiel.test/v",
					Beigaben: []beigabe{
						{Name: "MagicSetEditor4Mac-0.7.0.dmg", Groesse: len(inhalt), Adresse: eigen + "/abbild"},
						{Name: "pruefsummen.txt", Groesse: len(summen), Adresse: eigen + "/summen"},
						{Name: "pruefsummen.txt.asc", Groesse: len(signatur), Adresse: eigen + "/signatur"},
					},
				}); fehler != nil {
					t.Errorf("schreiben: %v", fehler)
				}

				return
			case strings.HasSuffix(anfrage.URL.Path, "/signatur"):
				rumpf = []byte(signatur)
			case strings.HasSuffix(anfrage.URL.Path, "/summen"):
				rumpf = []byte(summen)
			default:
				schreiben.Header().Set("Content-Length", strconv.Itoa(len(inhalt)))
				rumpf = inhalt
			}

			if _, fehler := schreiben.Write(rumpf); fehler != nil {
				t.Logf("der Klient ging weg: %v", fehler)
			}
		}))
	t.Cleanup(dienst.Close)

	return dienst.URL
}

func TestEineAlteFreigabeUnterNeuemTagWirdUeberJedeFlaecheAbgewiesen(t *testing.T) {
	for name, abweisen := range map[string]func(*testing.T, []string){
		"Kommandozeile": abweisungAufDerKommandozeile,
		"Dienst":        abweisungImDienst,
	} {
		t.Run(name, func(t *testing.T) {
			umgebung := mitProbeanker(append(mitEigenerSaat(t),
				"MSE_FREIGABE_ADRESSE="+rueckfallgegenstelle(t)))

			ort := downloadverzeichnis(t, umgebung)

			abweisen(t, umgebung)

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
