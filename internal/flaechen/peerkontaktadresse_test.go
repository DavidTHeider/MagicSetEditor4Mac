package flaechen_test

// **Wo ein Gegenüber erreichbar ist — über jede Fläche** — Werkstück N2 von
// Phase 5.4
// (#1455).
//
// # Was hier gemessen wird
//
// **Dass die Adresse an beiden schreibenden Flächen dasselbe tut** (§16) — und
// dass das **Löschen** dazugehört. „Ich kenne keine mehr" ist eine Aussage;
// eine Fläche, die einen leeren Wert still überspränge, ließe die alte stehen,
// und der nächste Versuch ginge an einen Rechner, der inzwischen jemand
// anderem gehört.
//
// **Und dass das Nein überall dasselbe ist.** Eine krumme Adresse endet, und
// zwar an jeder Fläche.
//
// # Was hier **nicht** gemessen wird
//
// **Dass MCP sie zeigt** — es zeigt sie nicht, und das ist eine Entscheidung:
// eine Adresse sagt etwas über das Netz dieses Menschen und nichts über den
// Kontakt. Der Eintrag dazu steht in `docs/FAEHIGKEITEN.md`.

import (
	"net/http"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

func TestDieAdresseEinesGegenueberUeberJedeFlaeche(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	an := starteDienst(t, umgebung)

	fremd, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	kennung := string(fremd.Kennung())

	mussKommandozeile(t, umgebung, "kontakt-aufnehmen", kennung, "Annas Mac")

	// **Ohne Eintrag steht keine da** — und zwar als leeres Feld und nicht
	// als fehlendes.
	if aus := mussREST(t, an, http.MethodGet, "/peer/kontakte",
		""); !strings.Contains(aus, `"adresse":""`) {
		t.Fatalf("vorher steht da: %s", aus)
	}

	if aus := mussKommandozeile(t, umgebung, "kontakte"); strings.Contains(
		aus, "unter\t") {
		t.Fatalf("die Kommandozeile nennt schon eine: %q", aus)
	}

	// **Eingetragen wird an der Kommandozeile, gesehen im Dienst.**
	mussKommandozeile(
		t, umgebung, "kontakt-adresse", kennung, "192.168.1.7:41234")

	if aus := mussREST(t, an, http.MethodGet, "/peer/kontakte",
		""); !strings.Contains(aus, `"adresse":"192.168.1.7:41234"`) {
		t.Fatalf("der Dienst sieht sie nicht: %s", aus)
	}

	if aus := mussKommandozeile(t, umgebung, "kontakte"); !strings.Contains(
		aus, "unter\t192.168.1.7:41234") {
		t.Fatalf("die Kommandozeile sieht sie nicht: %q", aus)
	}

	// **Und geändert wird über den Dienst, gesehen an der Kommandozeile** —
	// beide Richtungen, damit keine Fläche eine Einbahn ist.
	mussREST(t, an, http.MethodPut, "/peer/kontakte/"+kennung+"/adresse",
		`{"adresse":"[2001:db8::1]:41234"}`)

	if aus := mussKommandozeile(t, umgebung, "kontakte"); !strings.Contains(
		aus, "unter\t[2001:db8::1]:41234") {
		t.Fatalf("nach der Änderung steht da %q", aus)
	}

	// **Eine krumme Adresse endet — an beiden.**
	if _, _, code := kommandozeile(t, umgebung,
		"kontakt-adresse", kennung, "frag mal David"); code == 0 {
		t.Error("die Kommandozeile nahm eine krumme Adresse an")
	}

	if status, _ := an.ruf(t, http.MethodPut,
		"/peer/kontakte/"+kennung+"/adresse",
		`{"adresse":"frag mal David"}`); status < 400 {
		t.Errorf("der Dienst nahm eine krumme Adresse an: %d", status)
	}

	// **Und die alte steht danach noch da** — ein abgewiesener Schreibversuch
	// darf nichts hinterlassen.
	if aus := mussKommandozeile(t, umgebung, "kontakte"); !strings.Contains(
		aus, "unter\t[2001:db8::1]:41234") {
		t.Fatalf("nach der Ablehnung steht da %q", aus)
	}

	// **Eine leere löscht sie** — über den Dienst gegeben, an der
	// Kommandozeile gesehen.
	mussREST(t, an, http.MethodPut, "/peer/kontakte/"+kennung+"/adresse",
		`{"adresse":""}`)

	if aus := mussKommandozeile(t, umgebung, "kontakte"); strings.Contains(
		aus, "unter\t") {
		t.Fatalf("nach dem Löschen steht sie noch da: %q", aus)
	}

	// **Und ebenso an der Kommandozeile** — das leere Argument ist eine
	// Angabe und kein vergessenes.
	mussKommandozeile(t, umgebung, "kontakt-adresse", kennung, "beispiel:443")
	mussKommandozeile(t, umgebung, "kontakt-adresse", kennung, "")

	if aus := mussREST(t, an, http.MethodGet, "/peer/kontakte",
		""); !strings.Contains(aus, `"adresse":""`) {
		t.Fatalf("nach dem zweiten Löschen steht da: %s", aus)
	}
}
