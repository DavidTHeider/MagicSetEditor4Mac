package service_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Kontingente je Principal und global, mit Schutzreserve fürs Fenster** —
// Werkstück O5 von Phase 5.2 (#1429).

func kleineGrenzen() domain.Haushaltsgrenzen {
	grenzen := domain.Vorgabegrenzen()
	grenzen.Principal = domain.Kontingent{Burst: 4, Nachfuellen: 4, Fenster: time.Minute}
	grenzen.Global = domain.Kontingent{Burst: 10, Nachfuellen: 10, Fenster: time.Minute}
	grenzen.Reserve = 4

	return grenzen
}

func TestEinPrincipalVerbrauchtSeinBudgetUndDerAndereArbeitetWeiter(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, time.September, 9, 12, 0, 0, 0, time.UTC)
	kontingente := service.NeueKontingente(kleineGrenzen(), func() time.Time { return start })

	for range 4 {
		if ok, _ := kontingente.Abbuchen("a", domain.KostenLesen, false); !ok {
			t.Fatal("a: Budget zu früh leer")
		}
	}

	ok, warten := kontingente.Abbuchen("a", domain.KostenLesen, false)
	if ok || warten != 15*time.Second {
		t.Errorf("a über dem Budget: ok=%v warten=%v", ok, warten)
	}

	// **B bleibt arbeitsfähig** — sein eigenes Budget ist voll.
	if ok, _ := kontingente.Abbuchen("b", domain.KostenSchreiben, false); !ok {
		t.Error("b durch a ausgesperrt")
	}

	// Und ein Schreiben kostet zwei: nach 4 + 2 sind global noch 4 — die Reserve.
	if ok, warten := kontingente.Abbuchen("c", domain.KostenLesen, false); ok || warten != 6*time.Second {
		t.Errorf("c in der Reserve: ok=%v warten=%v", ok, warten)
	}

	// **Das Fenster darf in die Reserve** — und hat kein eigenes Budget.
	for range 4 {
		if ok, _ := kontingente.Abbuchen("", domain.KostenLesen, true); !ok {
			t.Fatal("das Fenster ist ausgesperrt")
		}
	}

	// Leer ist leer, auch fürs Fenster — mit ehrlicher Wartezeit.
	if ok, warten := kontingente.Abbuchen("", domain.KostenLesen, true); ok || warten != 6*time.Second {
		t.Errorf("Fenster bei leerem Global: ok=%v warten=%v", ok, warten)
	}
}

func TestDieTabelleDerPrincipalsBleibtBegrenzt(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, time.September, 9, 12, 0, 0, 0, time.UTC)
	// Ein weites globales Guthaben — hier geht es um die Tabelle, nicht ums Ganze.
	grenzen := domain.Vorgabegrenzen()
	grenzen.Global = domain.Kontingent{Burst: 100_000, Nachfuellen: 1, Fenster: time.Minute}
	kontingente := service.NeueKontingente(grenzen, func() time.Time { return start })

	for i := range service.HoechstensPrincipals * 3 {
		kontingente.Abbuchen(fmt.Sprintf("p%d", i), domain.KostenLesen, false)
	}

	if umfang := kontingente.Umfang(); umfang != service.HoechstensPrincipals {
		t.Errorf("Tabelle: %d Einträge", umfang)
	}

	// Ein verdrängter Principal beginnt wieder voll — das ist der Preis der
	// Grenze, und er ist bewusst: Speicher ist endlich, ein Guthaben nicht wert.
	if ok, _ := kontingente.Abbuchen("p0", domain.KostenLesen, false); !ok {
		t.Error("verdrängter Principal ohne Guthaben")
	}
}

// **Was nichts kostet, geht immer durch** — #2003. Ein Abbruch nimmt Last;
// im Lastlauf bekam er bei leerem Guthaben ein 429, und der Auftrag lief zu
// Ende. Auch die Reserve fürs Fenster hält ihn nicht auf: sie schützt
// Guthaben, und ein Abbruch nimmt keines.
func TestWasNichtsKostetGehtImmerDurch(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, time.October, 2, 12, 0, 0, 0, time.UTC)
	kontingente := service.NeueKontingente(kleineGrenzen(), func() time.Time { return start })

	// a leert sein Guthaben, das Fenster das globale samt der Reserve — nur es
	// darf hinein.
	for range 4 {
		kontingente.Abbuchen("a", domain.KostenLesen, false)
	}

	for range 6 {
		kontingente.Abbuchen("", domain.KostenLesen, true)
	}

	// **Der Beleg, dass jemand zuhört**: ein Lesen bekommt jetzt ein Nein.
	if ok, _ := kontingente.Abbuchen("a", domain.KostenLesen, false); ok {
		t.Fatal("das leere Guthaben ließ ein Lesen durch")
	}

	for _, principal := range []string{"a", "c"} {
		if ok, warten := kontingente.Abbuchen(principal, domain.KostenAbbruch, false); !ok || warten != 0 {
			t.Errorf("ein Abbruch von %s: ok=%v warten=%v", principal, ok, warten)
		}
	}
}
