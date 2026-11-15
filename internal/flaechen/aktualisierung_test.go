// **Dieselbe Aktualisierung über jede Fläche**
// (#1100,
// Werkstück 7).
//
// **Die Erklärung, was diese Läufe halten und was nicht, steht in
// `scryfallsatz_test.go`.**
//
// # Warum der Lauf hierhergehört
//
// §16 verlangt für jede Fähigkeit einen Lauf, der **denselben Arbeitsablauf
// über jede Fläche fährt und dieselben fachlichen Fehler erwartet**. Der
// Fähigkeitskatalog zählt nur, **dass** eine Fähigkeit an einer Fläche
// vorhanden ist — nicht, dass sie dort dasselbe tut.
//
// **Hier ist das mehr als eine Formsache**: der Ladeweg legt eine Datei ins
// Downloadverzeichnis der Nutzenden. Eine Fläche, die eine falsche Prüfsumme
// durchgehen ließe, wäre die eine, über die etwas hereinkommt.
//
// # Was er nicht prüft
//
//   - **Das Fenster.** Es ist kein Prozess, den ein Test startet; sein Weg
//     steht in `frontend/src/fassungsband.test.tsx`.
//   - **MCP.** Die Fähigkeit gibt es dort bewusst nicht; der Grund steht in
//     `docs/FAEHIGKEITEN.md`.
package flaechen_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// freigabeantwort schreibt die Freigabe als JSON.
//
// **Über [json.NewEncoder] und nicht als zusammengefügte Zeichenkette.** Der
// erste Anlauf setzte sie aus Stücken zusammen, und `anfrage.Host` fließt
// darin bis in die Antwort — gosec liest das als Einschleusung und hat recht:
// eine Gestalt, die man von Hand zusammenklebt, hält man von Hand richtig.
func freigabeantwort(
	t *testing.T, schreiben http.ResponseWriter, eigen string, groesse int,
) {
	t.Helper()

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
			{
				Name:    "MagicSetEditor4Mac-99.9.9.dmg",
				Groesse: groesse,
				Adresse: eigen + "/abbild",
			},
			{Name: "pruefsummen.txt", Groesse: 8, Adresse: eigen + "/summen"},
			{
				Name:    "pruefsummen.txt.asc",
				Groesse: 8,
				Adresse: eigen + "/signatur",
			},
		},
	}); fehler != nil {
		t.Errorf("schreiben: %v", fehler)
	}
}

// freigabegegenstelle gibt eine Gegenstelle, die eine Freigabe ausliefert.
//
// **`stimmt` entscheidet, ob die Prüfsummenzeile zum Abbild passt.** Damit
// fährt derselbe Aufbau beide Fälle — den geglückten und den, auf den es
// ankommt.
func freigabegegenstelle(t *testing.T, stimmt bool) string {
	t.Helper()

	inhalt := []byte(strings.Repeat("abbild", 2000))

	// **Signiert mit dem Probeanker** (SEC03, #2356). `mse` und `mse-rest`
	// sind eigene Programme und prüfen gegen den Anker in ihnen; gebaut mit
	// dem Tag `probeanker` nehmen sie den, den dieser Lauf erzeugt hat
	// (`probeanker_test.go`). Hier stand der echte Schlüssel des
	// Herausgebers aus dem Schlüsselbund dieses Rechners.
	summe := sha256.Sum256(inhalt)
	zeile := hex.EncodeToString(summe[:])

	if !stimmt {
		zeile = strings.Repeat("0", 64)
	}

	summen := zeile + "  ./MagicSetEditor4Mac-99.9.9.dmg\n"
	signatur := probeSigniert(t, summen)

	dienst := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, anfrage *http.Request) {
			switch {
			case strings.HasSuffix(anfrage.URL.Path, "/latest"):
				// **Die Adresse kommt aus der Anfrage, nicht aus einer
				// geschlossenen Veränderlichen.** `httptest.NewServer` startet
				// seine Schleife **vor** der Rückgabe; wer danach `dienst.URL` in
				// eine Variable schreibt, die der Behandler liest, hat ein
				// Wettrennen — vom Race-Detektor am 2026-08-26 gefunden.
				eigen := "http://" + anfrage.Host

				freigabeantwort(t, schreiben, eigen, len(inhalt))
			case strings.HasSuffix(anfrage.URL.Path, "/signatur"):
				if _, fehler := schreiben.Write([]byte(signatur)); fehler != nil {
					t.Errorf("schreiben: %v", fehler)
				}
			case strings.HasSuffix(anfrage.URL.Path, "/summen"):
				if _, fehler := schreiben.Write([]byte(summen)); fehler != nil {
					t.Errorf("schreiben: %v", fehler)
				}
			default:
				schreiben.Header().Set("Content-Length", strconv.Itoa(len(inhalt)))

				if _, fehler := schreiben.Write(inhalt); fehler != nil {
					t.Logf("der Klient ging weg: %v", fehler)
				}
			}
		}))

	t.Cleanup(dienst.Close)

	return dienst.URL
}

// zuhauseAus liest das `HOME` aus einer Umgebung — **das letzte**.
//
// # Warum das letzte und nicht das erste
//
// `eigenerBestand` fängt bei [os.Environ] an, und darin steht das **echte**
// `HOME` dieses Rechners; `mitEigenerSaat` hängt danach das des Wegwerfheims
// an. Wer das erste nähme, bekäme das echte.
//
// **Und das ist hier kein Schönheitsfehler.** Der erste Anlauf legte damit
// `~/Downloads` im echten Zuhause an, zählte dessen Inhalt und wäre an den
// Dateien der Nutzenden gescheitert — der Fall aus §3: ein Test, der nach
// draußen greift, ohne dass jemand es wollte. Aufgefallen ist es nur daran,
// dass der Kindprozess das **letzte** `HOME` nahm und die Datei dort nicht
// anlegen konnte.
//
// **Deshalb steht die Probe daneben**: ein `HOME`, das nicht unter dem
// Wegwerfverzeichnis dieses Laufs liegt, beendet den Test, statt ihn woanders
// weiterlaufen zu lassen.
func zuhauseAus(t *testing.T, umgebung []string) string {
	t.Helper()

	for i := len(umgebung) - 1; i >= 0; i-- {
		wert, gefunden := strings.CutPrefix(umgebung[i], "HOME=")
		if !gefunden {
			continue
		}

		if !strings.HasPrefix(wert, os.TempDir()) &&
			!strings.HasPrefix(wert, "/var/folders/") {
			t.Fatalf("das HOME dieses Laufs ist %q — das ist kein Wegwerfheim", wert)
		}

		return wert
	}

	t.Fatal("die Umgebung führt kein HOME")

	return ""
}

// downloadverzeichnis legt `~/Downloads` an und gibt seinen Pfad.
func downloadverzeichnis(t *testing.T, umgebung []string) string {
	t.Helper()

	ort := filepath.Join(zuhauseAus(t, umgebung), "Downloads")
	if err := os.MkdirAll(ort, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	return ort
}

// **Beide Flächen laden dieselbe Freigabe an denselben Ort.**
func TestDieAktualisierungUeberJedeFlaeche(t *testing.T) {
	for name, laden := range map[string]func(*testing.T, []string) string{
		"Kommandozeile": ladenUeberKommandozeile,
		"Dienst":        ladenUeberDienst,
	} {
		t.Run(name, func(t *testing.T) {
			umgebung := mitProbeanker(append(mitEigenerSaat(t),
				"MSE_FREIGABE_ADRESSE="+freigabegegenstelle(t, true)))

			ort := downloadverzeichnis(t, umgebung)

			pfad := laden(t, umgebung)

			if pfad != filepath.Join(ort, "MagicSetEditor4Mac-99.9.9.dmg") {
				t.Errorf("die Datei liegt unter %q", pfad)
			}

			if _, err := os.Stat(pfad); err != nil {
				t.Errorf("die Datei liegt nicht da: %v", err)
			}
		})
	}
}

// **Und beide weisen eine falsche Prüfsumme ab — ohne etwas liegen zu lassen.**
//
// **Das ist der Satz, für den es diesen Lauf gibt.** Eine Fläche, die hier
// durchließe, wäre die, über die etwas auf die Platte der Nutzenden käme.
func TestEineFalschePruefsummeWirdUeberJedeFlaecheAbgewiesen(t *testing.T) {
	for name, abweisen := range map[string]func(*testing.T, []string){
		"Kommandozeile": abweisungAufDerKommandozeile,
		"Dienst":        abweisungImDienst,
	} {
		t.Run(name, func(t *testing.T) {
			umgebung := mitProbeanker(append(mitEigenerSaat(t),
				"MSE_FREIGABE_ADRESSE="+freigabegegenstelle(t, false)))

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

// ladenUeberKommandozeile fährt `mse aktualisierung --laden`.
func ladenUeberKommandozeile(t *testing.T, umgebung []string) string {
	t.Helper()

	aus, fehler, code := kommandozeile(t, umgebung, "aktualisierung", "--laden")
	if code != 0 {
		t.Fatalf("mse aktualisierung --laden: %d (%s / %s)", code, aus, fehler)
	}

	// **Der Pfad steht als letzte Zeile auf dem Ziel** — der Fortschritt geht
	// auf den Fehlerstrom, damit ein Skript ihn weiterreichen kann.
	zeilen := strings.Fields(strings.TrimSpace(aus))

	return zeilen[len(zeilen)-1]
}

// ladenUeberDienst fährt `POST /aktualisierung/laden`.
func ladenUeberDienst(t *testing.T, umgebung []string) string {
	t.Helper()

	an := starteDienst(t, umgebung)

	status, rumpf := an.ruf(t, http.MethodPost, "/aktualisierung/laden", "")
	if status != http.StatusCreated {
		t.Fatalf("POST /aktualisierung/laden: %d (%s)", status, rumpf)
	}

	var antwort struct {
		Pfad string `json:"pfad"`
	}

	if err := json.Unmarshal([]byte(rumpf), &antwort); err != nil {
		t.Fatalf("die Antwort lesen: %v (%s)", err, rumpf)
	}

	return antwort.Pfad
}

// abweisungAufDerKommandozeile erwartet einen fachlichen Fehlschlag.
func abweisungAufDerKommandozeile(t *testing.T, umgebung []string) {
	t.Helper()

	aus, _, code := kommandozeile(t, umgebung, "aktualisierung", "--laden")
	if code == 0 {
		t.Fatalf("es kam durch: %s", aus)
	}
}

// abweisungImDienst erwartet einen fachlichen Fehlschlag.
//
// **Nicht auf einen festen Status.** Eine falsche Prüfsumme ist ein Konflikt,
// eine gleiche Fassung eine Validierung — beide Male ist die Zusage dieselbe:
// **es wird nichts abgelegt.** Wer hier einen Code festschriebe, prüfte die
// Übersetzung und nicht die Zusage.
func abweisungImDienst(t *testing.T, umgebung []string) {
	t.Helper()

	an := starteDienst(t, umgebung)

	status, rumpf := an.ruf(t, http.MethodPost, "/aktualisierung/laden", "")
	if status < 400 {
		t.Fatalf("POST /aktualisierung/laden: %d (%s)", status, rumpf)
	}
}
