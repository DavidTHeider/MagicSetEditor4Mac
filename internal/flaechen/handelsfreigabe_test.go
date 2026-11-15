package flaechen_test

// **Kein verbindlicher Handelsschritt ohne die Zustimmung eines Menschen** —
// Werkstück `T10` von Phase 5.6
// (#1475).
//
// # Was hier gemessen wird
//
//   - **Dass keine der drei kopflosen Flächen einen Nachweis ausstellen
//     kann.** Er entsteht am Fenster, von einem Menschen; eine Fläche, die
//     ihn sich selbst gäbe, hebelte aus, wofür es ihn gibt.
//   - **Dass ein erfundenes Freigabezeichen nichts bewirkt** — über
//     Kommandozeile und Dienst, mit derselben fachlichen Absage.
//   - **Und dass MCP diesen Schritt gar nicht anbietet.** Das ist eine
//     Entscheidung und keine Lücke: ein Modell bestätigt keine physische
//     Übergabe.
//
// # Und der geglückte Schritt steht nebenan
//
// **`handelsleitung_test.go` fährt ihn** — der Nachweis entsteht dort am
// Fenster, und die Kommandozeile löst ihn ein
// (#1637).
// Hier stand bis zum 2026-09-19, dass es ihn nicht geben kann: die Schritte
// bis zur bestätigten Übergabe hatten keine Fläche. **Was dieser Lauf misst,
// ist die andere Hälfte** — dass kein erfundenes Zeichen wirkt und dass keine
// kopflose Fläche eines ausstellt.

import (
	"net/http"
	"strings"
	"testing"
)

// derSchrittwegT10 ist der Endpunkt, wie er im Katalog steht.
const derSchrittwegT10 = "POST /handel/{vorgang}/schritt"

// dasErfundeneZeichen ist ein Freigabezeichen, das niemand erteilt hat.
const dasErfundeneZeichen = "f-erfunden"

// **Ein erfundenes Freigabezeichen verbucht nichts.**
func TestEinErfundenesFreigabezeichenVerbuchtNichts(t *testing.T) {
	t.Parallel()

	umgebung := mitEigenerSaat(t)
	an := starteDienst(t, umgebung)

	ausgabe, klage, code := kommandozeile(t, umgebung,
		"--json", "handel", "verbuchen", derUnbekannteHandel,
		dasErfundeneZeichen)
	if code == 0 {
		t.Fatalf("die Kommandozeile verbuchte: %s", ausgabe)
	}

	if strings.TrimSpace(klage+ausgabe) == "" {
		t.Error("die Absage sagt nichts")
	}

	status, rumpf := an.ruf(t, http.MethodPost,
		strings.ReplaceAll(pfadAus(derSchrittwegT10), "{vorgang}",
			derUnbekannteHandel),
		`{"schritt":"verbuchen","wer":"","freigabe":"`+
			dasErfundeneZeichen+`"}`)
	if status == http.StatusOK {
		t.Errorf("der Dienst verbuchte: %s", rumpf)
	}
}

// **Und ohne jedes Zeichen ebenso.**
//
// **Die Gegenprobe zum Lauf darüber**: wäre der Riegel gar nicht da, käme
// beides gleich weit — und der Unterschied zwischen „erfunden" und „gar
// keins" sagte nichts über ihn. Beide enden vor dem Bestand.
func TestOhneFreigabezeichenVerbuchtNichts(t *testing.T) {
	t.Parallel()

	umgebung := mitEigenerSaat(t)

	ausgabe, _, code := kommandozeile(t, umgebung,
		"--json", "handel", "verbuchen", derUnbekannteHandel)
	if code == 0 {
		t.Fatalf("die Kommandozeile verbuchte ohne Freigabe: %s", ausgabe)
	}
}

// **Keine kopflose Fläche stellt einen Nachweis aus.**
//
// **Gemessen an der Abwesenheit**: die Kommandozeile kennt keinen Befehl
// dafür, und der Dienst keinen Weg. Das ist die Zusage aus 0078 — ein
// Nachweis entsteht dort, wo ein Mensch ist.
func TestKeineKopfloseFlaecheStelltEinenNachweisAus(t *testing.T) {
	t.Parallel()

	umgebung := mitEigenerSaat(t)

	hilfe, _, _ := kommandozeile(t, umgebung, "hilfe")

	for _, wort := range []string{"freigabe-erteilen", "handel-freigabe"} {
		if strings.Contains(hilfe, wort) {
			t.Errorf("die Kommandozeile bietet %q an", wort)
		}
	}

	// **Und MCP bewegt keinen Handel** — weder das Verbuchen noch die
	// Freigabe dafür, und seit
	// #1637
	// auch keinen der neun Schritte davor.
	//
	// **Über das Wort „handel" und nicht über eine Namensliste**: eine Liste
	// wäre beim nächsten Werkzeug zu kurz, und niemand sähe es. Was sie
	// nicht trifft, sind die **Sichtfreigaben** aus `T3` — die erlauben
	// einem Gegenüber, etwas zu **sehen**, und sind kein Handelsschritt.
	//
	// **Und was sie trifft, muss lesend sein.** Hier stand bis zum
	// 2026-09-19 „gar nichts", und das war die engere Zusage; sie ist mit
	// `lesen.handel` gefallen. Nachsehen, wo ein Vorgang steht, ist Auskunft
	// über den **eigenen** Bestand — den Vorgang **bewegen** kann diese
	// Fläche nicht, und genau das misst der Lauf jetzt.
	agent := starteMCPProzess(t, umgebung)

	gesehen := 0

	for _, name := range werkzeugnamen(t, agent) {
		if !strings.Contains(name, "handel") {
			continue
		}

		gesehen++

		if !strings.HasPrefix(name, "lesen.") {
			t.Errorf("MCP führt das schreibende Werkzeug %q", name)
		}
	}

	// **Sonst wäre ein Lauf über null Werkzeuge grün** — und die Zusage
	// „lesend" sagte nichts (§10).
	if gesehen == 0 {
		t.Error("kein einziges Handelswerkzeug gesehen — der Lauf greift daneben")
	}
}
