// **Der zweite Faktor — an jeder Fläche** — Werkstück 7 von Phase 4.12
// (#1351).
//
// # Die Abnahme
//
// **Einrichten, bestätigen, und danach geht der Profilwechsel nicht mehr ohne
// Code.** Das ist die Zusage aus dem Issue — „der Profilwechsel verlangt einen
// zweiten Faktor" —, und sie wird hier gegen die **gebauten** Programme
// gefahren: jeder Ruf startet ein eigenes über demselben Bestand.
//
// **Der Code wird hier gerechnet**, mit einer zweiten Umsetzung derselben
// zwölf Zeilen aus RFC 4226 §5.3. Ein Lauf, der `internal/totp` ruft, prüfte
// die Rechnung gegen sich selbst.
package flaechen_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha1" //#nosec G505 -- RFC 6238 §1.2: HMAC-SHA-1 ist das Verfahren
	"encoding/base32"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schluesselbund"
)

// bestandMitBund gibt eine Umgebung mit eigenem Bestand **und eigenem
// Schlüsselbund**.
//
// # Warum der Bund eigens angelegt wird
//
// **Weil die gebauten Programme sonst in den echten schreiben.** Sie laufen
// als eigene Prozesse; ein Feld auf der `App` erreicht sie nicht. Ohne diesen
// Weg legte jeder Lauf Einträge im Schlüsselbund dieses Rechners ab — und
// wäre grün, **weil** er wirklich funktioniert hat (§3).
//
// **Und in der Pipeline hinge er.** Dort ist `HOME` ein frisches
// Verzeichnis; `security` findet keinen Anmeldebund und antwortet nicht, bis
// die Frist ihn erschlägt — gemessen am 2026-09-05: zehn Sekunden je Ruf,
// dann „signal: killed".
func bestandMitBund(t *testing.T) []string {
	t.Helper()

	return append(eigenerBestand(t), schluesselbund.Bundort+"="+probebund(t))
}

// probebund legt einen eigenen Schlüsselbund an und gibt seinen Ort.
//
// **Herausgezogen aus [bestandMitBund]**, weil ein Lauf beides braucht: die
// Saat **und** einen eigenen Bund. Ohne ihn griffe er in den echten
// Schlüsselbund dieses Rechners (§3).
func probebund(t *testing.T) string {
	t.Helper()

	ort := filepath.Join(t.TempDir(), "probe.keychain")

	// **Mit eigener Umgebung** (C06-Rest, #2284) — `security` erbte sonst das
	// echte Zuhause. Einmal gebaut: das Aufräumen läuft, wenn `t.TempDir`
	// nichts Neues mehr hergibt.
	umgebung := securityUmgebung(t)

	//nolint:gosec // das Programm steht als Literal, der Ort liegt in t.TempDir()
	anlegen := exec.CommandContext(t.Context(), "/usr/bin/security",
		"create-keychain", "-p", "leer", ort)
	anlegen.Env = umgebung

	if aus, err := anlegen.CombinedOutput(); err != nil {
		t.Fatalf("den Probebund anlegen: %v — %s", err, aus)
	}

	t.Cleanup(func() {
		// **Ein eigener Kontext**, weil der des Laufs beim Aufräumen schon
		// abgebrochen ist — dann räumte nichts ab.
		ctx, fertig := context.WithTimeout(context.Background(), 10*time.Second)
		defer fertig()

		loeschen := exec.CommandContext(ctx, "/usr/bin/security", "delete-keychain", ort) //nolint:gosec // dito
		loeschen.Env = umgebung
		// Ein Probebund, der stehen bleibt, liegt in einem Verzeichnis, das
		// der Lauf ohnehin abräumt.
		_ = loeschen.Run() //nolint:errcheck // siehe darüber
	})

	// **Aufgeschlossen, damit `security` nicht nach einem Passwort fragt.**
	//
	aufschliessen := exec.CommandContext(t.Context(), "/usr/bin/security", //nolint:gosec // dito
		"unlock-keychain", "-p", "leer", ort)
	aufschliessen.Env = umgebung
	_ = aufschliessen.Run() //nolint:errcheck // dito

	return ort
}

// **Der ganze Weg über die Kommandozeile.**
func TestDerGanzeFaktorwegAnDerKommandozeile(t *testing.T) {
	t.Parallel()

	umgebung := bestandMitBund(t)

	// **Ohne Faktor wechselt es ohne Nachweis** — die Abnahme aus dem Issue:
	// „ein bestehender Aufruf ohne Anmeldung bricht nicht."
	mussKommandozeile(t, umgebung, "profil", "1")

	eingerichtet := einrichtungAus(t,
		mussKommandozeile(t, umgebung, "faktor-neu", "1", "totp", "--json"))

	if !strings.HasPrefix(eingerichtet.Adresse, "otpauth://totp/") {
		t.Errorf("die Adresse ist %q", eingerichtet.Adresse)
	}

	if len(eingerichtet.Codes) != 10 {
		t.Errorf("es kamen %d Wiederherstellungscodes", len(eingerichtet.Codes))
	}

	mussKommandozeile(t, umgebung, "faktor-bestaetigen", "1",
		bestaetigungskennwort(t, eingerichtet.Geheimnis))

	// **Und jetzt geht der Wechsel ohne Nachweis nicht mehr.**
	if _, _, code := kommandozeile(t, umgebung, "profil", "1"); code != 1 {
		t.Errorf("der Wechsel ging ohne zweiten Faktor durch (%d)", code)
	}

	mussKommandozeile(t, umgebung, "profil", "1",
		einmalkennwort(t, eingerichtet.Geheimnis))

	// **Und ein Wiederherstellungscode geht auch** — genau einmal.
	mussKommandozeile(t, umgebung, "profil", "1", eingerichtet.Codes[0])

	if _, _, code := kommandozeile(t, umgebung, "profil", "1",
		eingerichtet.Codes[0]); code != 1 {
		t.Errorf("derselbe Wiederherstellungscode ging zweimal durch (%d)", code)
	}

	mussKommandozeile(t, umgebung, "faktor-loeschen", "1", "totp")
	mussKommandozeile(t, umgebung, "profil", "1")
}

// **Über MCP geht nur lesen und löschen — einrichten nicht.**
//
// **Die zweite Hälfte ist die wichtigere** (§10): eine Liste dessen, was
// fehlt, liest sich wie eine Zusage, und niemand prüft nach.
func TestUeberMCPGehtNurLesenUndLoeschen(t *testing.T) {
	t.Parallel()

	umgebung := eigenerBestand(t)

	if roh := mussMCP(t, umgebung, "lesen.faktoren",
		map[string]any{"nutzerID": "1"}); !strings.Contains(roh, "faktoren") {
		t.Errorf("die Antwort ist %q", roh)
	}

	// **Nichts im Bereich, das ein Geheimnis trüge.** Als Klasse und nicht
	// als Aufzählung erdachter Namen (§10).
	angemeldet := angemeldeteWerkzeuge(t, umgebung)

	for _, name := range angemeldet {
		if strings.HasPrefix(name, faktorbereich) && name != faktorloeschen {
			t.Errorf("%s ist angemeldet — im Bereich %s gehört über MCP nur %s, "+
				"weil alles andere ein Geheimnis in einen Werkzeugaufruf trüge",
				name, faktorbereich, faktorloeschen)
		}
	}

	if !slices.Contains(angemeldet, faktorloeschen) {
		t.Fatalf("die Werkzeugliste kennt nicht einmal das Löschen (%d Namen)",
			len(angemeldet))
	}
}

// **Und derselbe Weg über den Dienst.**
func TestDerGanzeFaktorwegUeberDenDienst(t *testing.T) {
	t.Parallel()

	umgebung := bestandMitBund(t)
	an := starteDienst(t, umgebung)

	eingerichtet := einrichtungAus(t, mussREST(t, an, http.MethodPost,
		"/nutzer/1/faktoren", `{"art":"totp"}`))

	mussREST(t, an, http.MethodPut, "/nutzer/1/faktoren/totp",
		`{"code":"`+bestaetigungskennwort(t, eingerichtet.Geheimnis)+`"}`)

	// **Ohne Nachweis wechselt jetzt nichts mehr.**
	status, antwort := an.ruf(t, http.MethodPut, "/profil", `{"profilID":"1"}`)
	if status < 400 {
		t.Errorf("der Wechsel ging ohne zweiten Faktor durch: %d (%s)", status, antwort)
	}

	mussREST(t, an, http.MethodPut, "/profil",
		`{"profilID":"1","nachweis":"`+einmalkennwort(t, eingerichtet.Geheimnis)+`"}`)

	// **Und das Geheimnis steht in keiner weiteren Antwort.**
	stand := mussREST(t, an, http.MethodGet, "/nutzer/1/faktoren", "")
	if strings.Contains(stand, eingerichtet.Geheimnis) {
		t.Errorf("das Geheimnis steht in der Übersicht: %s", stand)
	}

	mussREST(t, an, http.MethodDelete, "/nutzer/1/faktoren/totp", "")
	mussREST(t, an, http.MethodPut, "/profil", `{"profilID":"1"}`)
}

// Der Bereich der Faktorwerkzeuge und das einzige, das es darin gibt.
const (
	faktorbereich  = "schreiben.faktor."
	faktorloeschen = faktorbereich + "loeschen"
)

// einrichtung ist, was `faktor-neu` und `POST /nutzer/…/faktoren` geben.
type einrichtung struct {
	Adresse   string   `json:"adresse"`
	Geheimnis string   `json:"geheimnis"`
	Codes     []string `json:"codes"`
}

// einrichtungAus liest sie aus einer Antwort.
func einrichtungAus(t *testing.T, roh string) einrichtung {
	t.Helper()

	var gelesen einrichtung

	if err := json.Unmarshal([]byte(roh), &gelesen); err != nil {
		t.Fatalf("die Einrichtung lesen: %v — %s", err, roh)
	}

	if gelesen.Geheimnis == "" {
		t.Fatalf("die Einrichtung kam ohne Geheimnis: %s", roh)
	}

	return gelesen
}

// einmalkennwort rechnet den Code zu einem base32-Geheimnis.
//
// **Eine zweite Umsetzung von RFC 4226 §5.3** — siehe den Kopf dieser Datei.
func einmalkennwort(t *testing.T, geheimnis string) string {
	t.Helper()

	return kennwortZu(t, geheimnis, time.Now())
}

// bestaetigungskennwort ist der Code des **vorigen** Schritts — noch im
// Fenster, aber ein anderer als der jetzige.
//
// **Denn ein Code gilt einmal** (SEC-09, #2150): wer mit dem jetzigen
// bestätigt, kann sich damit nicht gleich danach anmelden — und genau das
// taten diese Läufe, bis der Einmalverbrauch kam.
//
// **Und der vorige Schritt gilt nur bis zum Ende des jetzigen.** Liegt die
// Grenze nah, wartet der Lauf sie ab (`wartenBisZurGrenze`).
func bestaetigungskennwort(t *testing.T, geheimnis string) string {
	t.Helper()

	time.Sleep(wartenBisZurGrenze(time.Now()))

	return kennwortZu(t, geheimnis, time.Now().Add(-30*time.Second))
}

// wartenBisZurGrenze sagt, wie lange bis zum nächsten Schritt zu warten ist —
// null, solange noch mindestens zehn Sekunden bleiben. So viel braucht die
// Kommandozeile unter der Last der Pipeline höchstens bis zur Prüfung.
func wartenBisZurGrenze(jetzt time.Time) time.Duration {
	rest := 30*time.Second - time.Duration(jetzt.UnixNano()%int64(30*time.Second))
	if rest >= 10*time.Second {
		return 0
	}

	return rest
}

// kennwortZu rechnet den Code zum Schritt von `zeitpunkt`.
func kennwortZu(t *testing.T, geheimnis string, zeitpunkt time.Time) string {
	t.Helper()

	roh, fehler := base32.StdEncoding.WithPadding(base32.NoPadding).
		DecodeString(geheimnis)
	if fehler != nil {
		t.Fatalf("das Geheimnis lesen: %v", fehler)
	}

	schritte := make([]byte, 8)
	binary.BigEndian.PutUint64(schritte, uint64(zeitpunkt.Unix())/30) //#nosec G115 -- Zeit vor 1970 gibt es hier nicht

	rechner := hmac.New(sha1.New, roh)
	rechner.Write(schritte)
	summe := rechner.Sum(nil)

	ab := summe[len(summe)-1] & 0x0f
	zahl := binary.BigEndian.Uint32(summe[ab:ab+4]) & 0x7fffffff

	return fmt.Sprintf("%06d", zahl%1_000_000)
}
