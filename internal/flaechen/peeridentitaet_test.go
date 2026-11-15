package flaechen_test

// **Die eigene Peeridentität über jede Fläche** — Werkstück N1 von Phase 5.4
// (#1454).
//
// # Was hier gemessen wird
//
// **Dass dieselbe Frage überall dieselbe Antwort bekommt** (§16) — und
// besonders: dass das **Anlegen** an drei Flächen dasselbe tut und der
// **Ablauf** an allen dreien derselbe ist. Ein Lauf je Fläche prüfte nur, dass
// jede für sich antwortet.
//
// # Und warum ein eigener Schlüsselbund
//
// **Weil hier wirklich geschrieben wird.** Ohne `MSE_SCHLUESSELBUND` legte
// dieser Lauf eine Identität im Anmeldebund dieses Rechners an — genau der
// Fehler, den `internal/schluesselbund` in seinem Kopf benennt.

import (
	"net/http"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schluesselbund"
)

func TestDiePeeridentitaetUeberJedeFlaeche(t *testing.T) {
	umgebung := append(mitEigenerSaat(t), schluesselbund.Bundort+"="+probebund(t))
	an := starteDienst(t, umgebung)

	// **Vor dem ersten Anlegen gibt es keine** — an allen drei Flächen, und
	// keine von ihnen legt beim Fragen eine an.
	fuerAlleDreiMit(t, "ohne Identität",
		[3]string{"stand\tohne", `"stand": "ohne"`, `"stand":"ohne"`},
		func() string { return mussKommandozeile(t, umgebung, "peer") },
		func() string { return mussMCP(t, umgebung, "lesen.peeridentitaet", map[string]any{}) },
		func() string { return mussREST(t, an, http.MethodGet, "/peer", "") })

	// **Angelegt wird an der Kommandozeile** — und danach sehen sie alle drei
	// dieselbe Kennung. Das ist die Frage, für die es diesen Lauf gibt: eine
	// Identität, die je Fläche eine andere wäre, wäre keine.
	angelegt := mussKommandozeile(t, umgebung, "peer-anlegen")
	if !strings.Contains(angelegt, "stand\tbereit") {
		t.Fatalf("das Anlegen gab %q", angelegt)
	}

	kennung := zeilenwert(t, angelegt, "kennung")
	if !strings.HasPrefix(kennung, "b") {
		t.Fatalf("die Kennung ist %q", kennung)
	}

	fuerAlleDreiMit(t, "dieselbe Identität",
		[3]string{kennung, kennung, kennung},
		func() string { return mussKommandozeile(t, umgebung, "peer") },
		func() string { return mussMCP(t, umgebung, "lesen.peeridentitaet", map[string]any{}) },
		func() string { return mussREST(t, an, http.MethodGet, "/peer", "") })

	// **Ein zweites Anlegen legt nichts Neues an** — an keiner der drei.
	for _, nochmal := range []func() string{
		func() string { return mussKommandozeile(t, umgebung, "peer-anlegen") },
		func() string { return mussREST(t, an, http.MethodPost, "/peer", "") },
	} {
		if aus := nochmal(); !strings.Contains(aus, kennung) {
			t.Errorf("ein zweites Anlegen gab %q", aus)
		}
	}

	// **Und die Rotation nennt die abgelöste Kennung** — sonst könnte
	// niemand einem Kontakt sagen, welche nicht mehr gilt. **An beiden
	// schreibenden Flächen**, denn beide sagen dasselbe zu.
	rotiert := mussREST(t, an, http.MethodPost, "/peer/rotation", "")
	if !strings.Contains(rotiert, kennung) {
		t.Fatalf("die Rotation im Dienst nennt die alte Kennung nicht: %s", rotiert)
	}

	nachRest := zeilenwert(t, mussKommandozeile(t, umgebung, "peer"), "kennung")
	if nachRest == kennung {
		t.Fatalf("nach der Rotation steht dieselbe Kennung da: %s", nachRest)
	}

	nochmal := mussKommandozeile(t, umgebung, "peer-rotieren")
	if !strings.Contains(nochmal, nachRest) {
		t.Fatalf("die Rotation an der Kommandozeile nennt die alte Kennung "+
			"nicht: %s", nochmal)
	}

	if zeilenwert(t, nochmal, "kennung") == nachRest {
		t.Fatal("die zweite Rotation gab dieselbe Kennung")
	}
}

// zeilenwert holt den Wert einer Zeile `name\twert` aus der Ausgabe.
func zeilenwert(t *testing.T, ausgabe, name string) string {
	t.Helper()

	for _, zeile := range strings.Split(ausgabe, "\n") {
		if teile := strings.SplitN(zeile, "\t", 2); len(teile) == 2 &&
			teile[0] == name {
			return strings.TrimSpace(teile[1])
		}
	}

	t.Fatalf("in %q steht keine Zeile %q", ausgabe, name)

	return ""
}
