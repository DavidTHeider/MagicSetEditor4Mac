package flaechen_test

// **Die Kontingente lassen sich abschalten — und alles andere bleibt** —
// QA-08.AK-03 (#1739).
//
// # Warum am gebauten Dienst
//
// **Weil der Schalter an drei Stellen wirkt, die kein Paket zusammen sieht**:
// die Einstellung im Bestand, das Lesen beim Start in der Wurzel des Dienstes
// und die Auskunft an jeder Fläche. Ein Lauf je Paket wäre grün, auch wenn
// die Wurzel den Schalter nie läse.
//
// # Was er nicht hält
//
// **Die Plätze je Ressourcenklasse, die Antwortgrenze und die Fristen** —
// sie hängen an keinem Zweig, den der Schalter berührt (`internal/service`,
// `internal/rest`), und ihre eigenen Läufe stehen dort. Hier stehen die zwei
// Zusagen, die ein Abschalten am ehesten mitreißen könnte: der Zugang und die
// Sperre gegen das Raten.

import (
	"net/http"
	"strconv"
	"testing"
)

func TestDieKontingenteLassenSichAbschalten(t *testing.T) {
	umgebung := mitEigenerSaat(t)

	// **Mit Kontingenten**: der Burst ist 120 Lesen, danach wartet ein Ruf —
	// mit einer Wartezeit, die eine Zahl ist. **Nicht genau der 121.**: der
	// Zulauf ist zwei Einheiten je Sekunde, und ein langsamer Lauf bekommt
	// unterwegs eine oder zwei dazu.
	mit := starteDienst(t, umgebung)
	if angehalten, warten := lesenBisZumNein(t, mit, 140); angehalten <= 120 || warten < 1 {
		t.Fatalf("mit Kontingenten: angehalten beim Ruf %d, Retry-After %d", angehalten, warten)
	}

	// **Abgeschaltet über die Kommandozeile** — dieselbe Einstellung, die das
	// Fenster setzt; sie gilt ab dem nächsten Start eines Dienstes.
	mussKommandozeile(t, umgebung, "einstellungen", "setzen", "grenzen.kontingente", "aus")

	ohne := starteDienst(t, umgebung)

	fuerAlleDreiMit(t, "abgeschaltete Kontingente",
		[3]string{"kontingente\taus", `"kontingente": "aus"`, `"kontingente":"aus"`},
		func() string { return mussKommandozeile(t, umgebung, "grenzen") },
		func() string { return mussMCP(t, umgebung, "lesen.grenzen", map[string]any{}) },
		func() string { return mussREST(t, ohne, http.MethodGet, "/grenzen", "") })

	if angehalten, _ := lesenBisZumNein(t, ohne, 400); angehalten != 0 {
		t.Errorf("ohne Kontingente angehalten beim Ruf %d", angehalten)
	}

	// **Der Zugang bleibt** — ohne Marke kein Ruf.
	if status, antwort := ohne.ohneMarke().ruf(t, http.MethodGet, "/grenzen", ""); status != http.StatusUnauthorized {
		t.Errorf("ohne Marke: %d %s", status, antwort)
	}

	// **Und die Sperre gegen das Raten bleibt**: drei freie Versuche, dann
	// wartet dieser Nachweis — ein abgeschaltetes Kontingent ist keine
	// Einladung zum Raten.
	ratend := ohne.mitNachweis("mse_falsch_geraten")
	for versuch := range 3 {
		if status, _ := ratend.ruf(t, http.MethodGet, "/grenzen", ""); status != http.StatusUnauthorized {
			t.Fatalf("Versuch %d: %d", versuch+1, status)
		}
	}

	ratend.ruf(t, http.MethodGet, "/grenzen", "")

	if status, koepfe := ratend.kopfruf(t, http.MethodGet, "/grenzen", ""); status != http.StatusTooManyRequests ||
		koepfe.Get("Retry-After") == "" {
		t.Errorf("der Ratende ist nicht gesperrt: %d %v", status, koepfe)
	}

	// **Und wer die Marke hat, merkt davon nichts.**
	mussREST(t, ohne, http.MethodGet, "/grenzen", "")
}

// lesenBisZumNein fragt bis zu `hoechstens` Mal die Grenzen und gibt den
// Ruf, der abgewiesen wurde — 0, wenn keiner — samt seiner Wartezeit.
func lesenBisZumNein(t *testing.T, an dienst, hoechstens int) (angehalten, warten int) {
	t.Helper()

	for ruf := 1; ruf <= hoechstens; ruf++ {
		status, koepfe := an.kopfruf(t, http.MethodGet, "/grenzen", "")
		if status == http.StatusTooManyRequests {
			sekunden, _ := strconv.Atoi(koepfe.Get("Retry-After")) //nolint:errcheck // 0 heißt: fehlt

			return ruf, sekunden
		}

		if status != http.StatusOK {
			t.Fatalf("Ruf %d: %d", ruf, status)
		}
	}

	return 0, 0
}
