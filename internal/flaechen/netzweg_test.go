// **Der Netzweg über jede Fläche ohne Fenster** — Werkstück N5 von Phase 5.4
// (#1571),
// Werkstück 2.
//
// # Was hier gemessen wird
//
// **Dass dieselbe Einstellung überall gilt** — und das ist der Punkt, der
// leicht durchfällt: `mse`, `mse-mcp` und `mse-rest` sind Klienten wie das
// Fenster, und eine Einstellung, die nur das Fenster erreichte, wäre eine
// halbe.
//
// **Und dass die Einstellung die Umgebung schlägt.** A18
// (#1533)
// hat diese Zusage wörtlich verlangt und offengelassen; hier steht sie als
// Messung.
package flaechen_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Derselbe Netzweg an jeder Fläche — und die Einstellung schlägt die
// Umgebung.**
func TestDenNetzwegUeberJedeFlaeche(t *testing.T) {
	// **Mit einem Proxy in der Umgebung**, damit die zweite Hälfte der
	// Zusage etwas zu schlagen hat.
	umgebung := append(mitEigenerSaat(t), "HTTPS_PROXY=http://aus.der.umgebung:3128")

	an := starteDienst(t, umgebung)

	// **Ohne Einstellung gilt die Umgebung** — an allen dreien.
	fuerAlleDreiMit(t, "der Proxy aus der Umgebung",
		[3]string{"aus.der.umgebung:3128", "aus.der.umgebung:3128", "aus.der.umgebung:3128"},
		func() string { return mussKommandozeile(t, umgebung, "netzweg") },
		func() string { return mussMCP(t, umgebung, "lesen.netzweg", map[string]any{}) },
		func() string { return mussREST(t, an, http.MethodGet, "/netz", "") })

	// **Und die Herkunft sagt es auch** — ein Wert ohne sie wäre eine
	// Auskunft, die nicht sagt, ob man sie ändern kann.
	if aus := mussKommandozeile(t, umgebung, "netzweg"); !strings.Contains(aus, "umgebung") {
		t.Errorf("die Herkunft fehlt: %s", aus)
	}

	// **Jetzt die Einstellung** — gesetzt an **einer** Fläche.
	mussKommandozeile(t, umgebung, "einstellungen", "setzen",
		"netz.proxy.https", "http://aus.der.einstellung:3128")

	// **Und sie gilt an allen dreien.** Das ist die Zusage, und sie steht
	// hier als Messung statt als Satz.
	fuerAlleDreiMit(t, "der Proxy aus der Einstellung",
		[3]string{
			"aus.der.einstellung:3128", "aus.der.einstellung:3128",
			"aus.der.einstellung:3128",
		},
		func() string { return mussKommandozeile(t, umgebung, "netzweg") },
		func() string { return mussMCP(t, umgebung, "lesen.netzweg", map[string]any{}) },
		func() string { return mussREST(t, an, http.MethodGet, "/netz", "") })

	// **Die Gegenprobe**: der Wert aus der Umgebung steht nicht mehr da.
	// Ohne sie sagte der Lauf nur, dass irgendwo „einstellung" vorkommt.
	if aus := mussKommandozeile(t, umgebung, "netzweg"); strings.Contains(
		aus, "aus.der.umgebung") {
		t.Errorf("die Umgebung steht noch im Weg: %s", aus)
	}
}

// **Eine abgewiesene Einstellung wird überall gleich abgewiesen.**
//
// **Ein Proxy ohne Schema ist der Fall, der zählt**: `httpproxy` liest ihn
// als **Rechnernamen mit Port**, und wer `http://` vergisst, bekäme
// stillschweigend etwas anderes, als er meinte.
func TestEinProxyOhneSchemaWirdUeberallAbgewiesen(t *testing.T) {
	umgebung := mitEigenerSaat(t)

	an := starteDienst(t, umgebung)

	abweisungGleich(t, "ein Proxy ohne Schema", abweisung{
		kern:   "Schema",
		klasse: klasseValidierung,
		cli: func() (string, int) {
			aus, fehler, code := kommandozeile(t, umgebung, "einstellungen",
				"setzen", "netz.proxy.https", "proxy.firma:3128")

			return aus + fehler, code
		},
		mcp: func() (string, bool) {
			return mcpRuf(t, umgebung, "schreiben.einstellung", map[string]any{
				"schluessel": "netz.proxy.https", "wert": "proxy.firma:3128",
			})
		},
		rest: func() (int, string) {
			return an.ruf(t, http.MethodPut, "/einstellungen/netz.proxy.https",
				`{"wert":"proxy.firma:3128"}`)
		},
		fenster: func() *app.Fehler {
			return fensterAus(t, umgebung).EinstellungSetzen(t.Context(), "netz.proxy.https", "proxy.firma:3128").Fehler
		},
	})
}
