package flaechen_test

import (
	"net/http"
	"testing"
)

// **Die Grenzen über jede Fläche** — Werkstück O5 von Phase 5.2 (#1429).
func TestDieGrenzenUeberJedeFlaeche(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	an := starteDienst(t, umgebung)

	fuerAlleDreiMit(t, "die Grenzen",
		[3]string{"rendern\t2", `"rendern": 2`, `"rendern":2`},
		func() string { return mussKommandozeile(t, umgebung, "grenzen") },
		func() string { return mussMCP(t, umgebung, "lesen.grenzen", map[string]any{}) },
		func() string { return mussREST(t, an, http.MethodGet, "/grenzen", "") })

	fuerAlleDreiMit(t, "die Reserve fürs Fenster",
		[3]string{"Reserve 120", `"reserve": 120`, `"reserve":120`},
		func() string { return mussKommandozeile(t, umgebung, "grenzen") },
		func() string { return mussMCP(t, umgebung, "lesen.grenzen", map[string]any{}) },
		func() string { return mussREST(t, an, http.MethodGet, "/grenzen", "") })
}
