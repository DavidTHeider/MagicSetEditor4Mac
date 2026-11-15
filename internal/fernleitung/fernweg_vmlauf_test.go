//go:build vmlauf

package fernleitung

// **Der Fernweg zwischen zwei Maschinen** — ABN-01.AK-05
// (#2175),
// mit SEC-22 (#2163).
//
// `fernweg_netzlauf_test.go` (SEC-21) stellt Horcher und Wählenden in
// **einen** Prozess. Hier stehen sie in zwei macOS-Gästen mit je eigenem
// Netzstapel, verbunden über tailcat — also über ein DERP-Relais von
// Tailscale im Internet, ohne Konto:
//
//	MSE_VM_ROLLE=a MSE_VM_EINLADUNG=~/einladung.json ./fernleitung.test -test.run Fernweg -test.v
//	MSE_VM_ROLLE=b MSE_VM_EINLADUNG=~/einladung.json ./fernleitung.test -test.run Fernweg -test.v
//
// A horcht und schreibt Einladung und Kennung in die Datei; wer die Gäste
// fährt, trägt sie zu B. B wählt an, schickt eine Nachricht und bekommt sie
// als Echo zurück — der reale Datentransfer. **Zwei Fälle**, wie in SEC-21:
// gewöhnlich und mit `TS_DEBUG_ALWAYS_USE_DERP=1` auf beiden Seiten.
//
// **Der Weg**: bis tailcat v0.6.0 nannte der Server keinen (SEC-21), und der
// Lauf sagte „nicht messbar". Seit v0.7.0 (K01,
// #2187)
// verlangt er ihn: gewöhnlich einen direkten, mit erzwungenem DERP das Relais
// und keinen direkten.
//
// **Was er nicht belegt**: zwei physische Macs in zwei Netzen hinter
// verschiedenen Routern — beide Gäste stehen hier im selben Netz.

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

type vmEinladung struct {
	Einladung domain.Einladung `json:"einladung"`
	Kennung   peer.Kennung     `json:"kennung"`
}

func TestDerFernwegZwischenZweiMaschinen(t *testing.T) {
	rolle, ablage := os.Getenv("MSE_VM_ROLLE"), os.Getenv("MSE_VM_EINLADUNG")
	if (rolle != "a" && rolle != "b") || ablage == "" {
		t.Skip("MSE_VM_ROLLE (a oder b) und MSE_VM_EINLADUNG fehlen — dieser Lauf gehört in zwei Gäste")
	}

	t.Setenv("HOME", t.TempDir())
	t.Logf("Rolle %s, TS_DEBUG_ALWAYS_USE_DERP=%q", rolle, os.Getenv("TS_DEBUG_ALWAYS_USE_DERP"))

	ich := eineIdentitaet(t)

	if rolle == "a" {
		horcheUndAntworte(t, ich, ablage)

		return
	}

	waehleUndPruefe(t, ich, ablage)
}

// horcheUndAntworte ist die Seite A: horchen, einladen, ein Echo geben.
func horcheUndAntworte(t *testing.T, ich peer.Identitaet, ablage string) {
	t.Helper()

	horcher, fehler := Horche(ich)
	if fehler != nil {
		t.Fatalf("Horche: %v", fehler)
	}

	defer func() { _ = horcher.Schliesse() }() //nolint:errcheck // beim Aufräumen

	roh, _ := json.Marshal(vmEinladung{Einladung: horcher.Einladung(), Kennung: ich.Kennung()}) //nolint:errcheck,errchkjson // zwei Zeichenketten
	if schreibfehler := os.WriteFile(ablage, roh, 0o600); schreibfehler != nil {                //nolint:gosec // den Ort nennt, wer den Lauf startet
		t.Fatalf("die Einladung ablegen: %v", schreibfehler)
	}

	// **Nicht die Einladung selbst**: sie trägt den vorab geteilten Schlüssel,
	// und das Protokoll dieses Laufs wird als Beleg abgelegt (Vermerk 0205).
	einladung := horcher.Einladung()
	t.Logf("Einladung abgelegt: %d Zeichen, %q…", len(einladung), einladung[:min(len(einladung), 10)])

	offen, fehler := horcher.Nimm(t.Context())
	if fehler != nil {
		t.Fatalf("Nimm: %v", fehler)
	}

	nachricht, fehler := offen.Empfange(t.Context())
	if fehler != nil {
		t.Fatalf("empfangen: %v", fehler)
	}

	if fehler := offen.Sende(t.Context(), append([]byte("zurück: "), nachricht...)); fehler != nil {
		t.Fatalf("antworten: %v", fehler)
	}

	t.Logf("bedient: %q von %s, %d angebunden", nachricht, offen.Gegenstelle(), horcher.Angebunden())

	nurRelais := os.Getenv("TS_DEBUG_ALWAYS_USE_DERP") != ""
	wege := wartetAufDenWeg(t, horcher, nurRelais)
	t.Logf("Weg: %s", wegAus(wege))

	switch {
	case len(wege) != 1:
		t.Errorf("tailcat nennt %d Wege, erwartet einen", len(wege))
	case nurRelais && (wege[0].Relais == "" || wege[0].Direkt != ""):
		t.Errorf("mit erzwungenem DERP erwartet: Relais und kein direkter Weg — %+v", wege[0])
	case !nurRelais && wege[0].Direkt == "":
		t.Errorf("gewöhnlich erwartet: ein direkter Weg — %+v", wege[0])
	}

	// **Offen lassen, bis B das Echo gelesen hat** — ein Schließen davor
	// nähme ihm die Antwort aus der Leitung.
	time.Sleep(10 * time.Second)
}

// waehleUndPruefe ist die Seite B: die Einladung lesen, wählen, das Echo prüfen.
func waehleUndPruefe(t *testing.T, ich peer.Identitaet, ablage string) {
	t.Helper()

	var gelesen vmEinladung

	for bis := time.Now().Add(2 * time.Minute); ; {
		roh, fehler := os.ReadFile(ablage) //nolint:gosec // dito
		if fehler == nil && json.Unmarshal(roh, &gelesen) == nil && gelesen.Einladung != "" {
			break
		}

		if time.Now().After(bis) {
			t.Fatalf("keine Einladung unter %s", ablage)
		}

		time.Sleep(time.Second)
	}

	anfang := time.Now()

	gewaehlt, fehler := Waehle(t.Context(), gelesen.Einladung, ich, gelesen.Kennung)
	if fehler != nil {
		t.Fatalf("Waehle: %v", fehler)
	}

	defer func() { _ = gewaehlt.Schliesse() }() //nolint:errcheck // beim Aufräumen

	gruss := []byte("Mausbraten von b")
	if sendefehler := gewaehlt.Sende(t.Context(), gruss); sendefehler != nil {
		t.Fatalf("senden: %v", sendefehler)
	}

	antwort, fehler := gewaehlt.Empfange(t.Context())
	if fehler != nil {
		t.Fatalf("empfangen: %v", fehler)
	}

	t.Logf("Echo nach %v: %q (%d Bytes), Gegenstelle %s", time.Since(anfang).Round(time.Millisecond), antwort, len(antwort), gewaehlt.Gegenstelle())

	if !bytes.Equal(antwort, append([]byte("zurück: "), gruss...)) {
		t.Fatalf("das Echo ist %q", antwort)
	}

	if gewaehlt.Gegenstelle() != gelesen.Kennung {
		t.Errorf("die Gegenstelle ist %s, eingeladen hat %s", gewaehlt.Gegenstelle(), gelesen.Kennung)
	}

	// **Verbunden bleiben, solange A nach dem Weg fragt** — ohne Gegenstelle
	// gäbe es keinen, und A wartet bis zu 15 s.
	time.Sleep(20 * time.Second)
}

// wegAus sagt, was über den Weg bekannt ist — oder dass nichts bekannt ist.
func wegAus(wege []Fernweg) string {
	if len(wege) == 0 {
		return "keiner — tailcat nennt die Gegenstelle nicht"
	}

	roh, _ := json.Marshal(wege) //nolint:errcheck,errchkjson // nur zur Anzeige

	return string(roh)
}
