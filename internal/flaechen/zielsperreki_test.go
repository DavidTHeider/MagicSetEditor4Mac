// **Ein Token lenkt die KI nicht nach innen** — SEC-P4
// (#2000),
// Befund P4 in `docs/PENTEST.md`.
//
// # Der Angriff, in zwei Zügen
//
// Ein Automat mit einem Token `schreiben` setzt `ki.adresse` auf einen
// Dienst, den nur dieser Rechner erreicht, und fragt danach den
// Modellkatalog — einen **lesenden** Weg, den jedes Token gehen darf. Die
// Anwendung fragte dann die Adresse, die der Automat gewählt hat.
//
// **Gezielt wird über `0.0.0.0`**: auf macOS landet eine Verbindung dorthin
// bei diesem Rechner (gemessen am 2026-10-02), und die Einstufung „auf der
// Schleife" in `internal/domain` kennt die Schreibweise nicht.
//
// # Woran gemessen
//
// **Am Zähler der nachgestellten Laufzeit.** Dazu zwei Gegenproben: ein
// Mensch an der Kommandozeile, der dieselbe Adresse setzt, erreicht sie auch
// nicht — die Grundsperre gilt jedem Klienten —, und die Schleife selbst,
// die eine lokale KI braucht, bleibt erreichbar.
package flaechen_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEinTokenLenktDieKINichtNachInnen(t *testing.T) {
	umgebung := append(eigenerBestand(t), "HOME="+t.TempDir())
	laufzeit := starteLaufzeit(t, `{}`)
	port := laufzeit.adresse[strings.LastIndex(laufzeit.adresse, ":")+1:]
	innen := "http://0.0.0.0:" + port

	mussKommandozeile(t, umgebung, "einstellungen", "setzen", domain.KILaufzeit, "an")

	gezeigt := profilkennungAus(t, mussKommandozeile(t, umgebung, "profil", "--json"))
	automat := strings.TrimSpace(mitEingabe(t, umgebung,
		`{"principal":"Rechner","bezeichnung":"Nachtlauf","scope":"schreiben"}`,
		"token-anlegen", gezeigt))

	an := starteDienst(t, umgebung)
	alsAutomat := an.mitNachweis(automat)

	// **Zug 1: wohin die Anwendung sendet, setzt kein Token** — weder die
	// Laufzeit noch das Gateway eines Anbieters noch den Proxy davor.
	for schluessel, wert := range map[string]string{
		domain.KIAdresse: innen,
		domain.DienstAdresse(domain.DienstAnthropic): innen,
		domain.NetzProxyHTTP:                         innen,
	} {
		if status, antwort := alsAutomat.ruf(t, http.MethodPut,
			"/einstellungen/"+schluessel, `{"wert":"`+wert+`"}`); status < http.StatusBadRequest {
			t.Errorf("ein Token setzte %s auf %s: %d — %s", schluessel, wert, status, antwort)
		}
	}

	// **Zug 2: der lesende Weg fragt dorthin nicht** — auch dann nicht,
	// wenn die Adresse dort steht, weil ein Mensch sie gesetzt hat.
	mussKommandozeile(t, umgebung, "einstellungen", "setzen", domain.KIAdresse, innen)

	vorher := laufzeit.beruehrt.Load()

	if status, antwort := alsAutomat.ruf(t, http.MethodGet, "/ki/modelle", ""); status != http.StatusOK {
		t.Fatalf("GET ki/modelle: %d — %s", status, antwort)
	}

	if nachher := laufzeit.beruehrt.Load(); nachher != vorher {
		t.Errorf("der Katalog fragte %s %d-mal", innen, nachher-vorher)
	}

	// **Die Gegenprobe**: die Schleife, auf der eine lokale KI lauscht,
	// bleibt erreichbar — derselbe Weg, dieselbe Laufzeit.
	mussKommandozeile(t, umgebung, "einstellungen", "setzen", domain.KIAdresse, laufzeit.adresse)

	if status, antwort := alsAutomat.ruf(t, http.MethodGet, "/ki/modelle", ""); status != http.StatusOK {
		t.Fatalf("GET ki/modelle: %d — %s", status, antwort)
	}

	if laufzeit.beruehrt.Load() == vorher {
		t.Error("auch über die Schleife kam keine Anfrage an — der Zähler misst nichts")
	}
}
