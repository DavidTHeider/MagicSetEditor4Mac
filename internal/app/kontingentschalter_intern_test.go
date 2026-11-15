package app

// **Die Kontingente lassen sich abschalten** — QA-08.AK-03
// (#1739).

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Ein Server liest den Schalter aus dem gezeigten Profil** — wie Spur und
// Zwischenspeicher, beim Start.
func TestDieGrenzenEinesServersFolgenDemSchalter(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	db, fehler := storage.Oeffnen(t.Context(), filepath.Join(t.TempDir(), "bestand.db"))
	if fehler != nil {
		t.Fatal(fehler)
	}

	t.Cleanup(func() { _ = db.Close() }) //nolint:errcheck // Aufräumen

	if migriert := storage.Migriere(t.Context(), db, nil); migriert != nil {
		t.Fatal(migriert)
	}

	// **Ohne Wahl gelten sie** — die Vorgabe schützt.
	if grenzen, lesefehler := GrenzenUeber(t.Context(), db); lesefehler != nil || grenzen.OhneKontingente {
		t.Errorf("ohne Wahl: ohne=%v, %v", grenzen.OhneKontingente, lesefehler)
	}

	gezeigt, fehler := storage.MitGezeigtemProfil(t.Context(), db)
	if fehler != nil {
		t.Fatal(fehler)
	}

	if gesetzt := einstellungsverwalterUeber(db).Setzen(gezeigt, domain.GrenzenKontingente, domain.KontingenteAus); gesetzt != nil {
		t.Fatal(gesetzt)
	}

	grenzen, fehler := GrenzenUeber(t.Context(), db)
	if fehler != nil || !grenzen.OhneKontingente {
		t.Fatalf("nach „aus“: ohne=%v, %v", grenzen.OhneKontingente, fehler)
	}

	// **Und was ein Server daraus baut, zählt nicht mehr** — weit über den Burst.
	jetzt := time.Date(2026, time.October, 2, 12, 0, 0, 0, time.UTC)
	kontingente := NeueKontingente(grenzen, func() time.Time { return jetzt })

	for stelle := range domain.Vorgabegrenzen().Principal.Burst * 3 {
		if ok, _ := kontingente.Abbuchen("a", domain.KostenLesen, false); !ok {
			t.Fatalf("Ruf %d angehalten", stelle)
		}
	}

	if auskunft := GrenzenauskunftVon(grenzen); auskunft.Kontingente != domain.KontingenteAus ||
		auskunft.Principal.Burst != domain.Vorgabegrenzen().Principal.Burst {
		t.Errorf("Auskunft: %+v", auskunft)
	}

	// **Ein Lesefehler lässt die Vorgabe gelten** — mit Kontingenten — und
	// kommt zurück, wie bei Spur und Zwischenspeicher.
	if geschlossen := db.Close(); geschlossen != nil {
		t.Fatal(geschlossen)
	}

	if grenzen, lesefehler := GrenzenUeber(t.Context(), db); lesefehler == nil || grenzen.OhneKontingente {
		t.Errorf("ohne Antwort: ohne=%v, %v", grenzen.OhneKontingente, lesefehler)
	}
}

// **Das Fenster nennt, was ein Dienst jetzt anwenden würde** — über seinen
// Bestand, an der Bindung wie an der Anwendung.
func TestDasFensterNenntDenKontingentschalter(t *testing.T) {
	anwendung, _ := gestartetMitWeiche(t, map[string]string{domain.GrenzenKontingente: domain.KontingenteAus})

	if antwort := anwendung.Grenzen(t.Context()); antwort.Fehler != nil ||
		antwort.Grenzen.Kontingente != domain.KontingenteAus {
		t.Errorf("Anwendung: %+v", antwort)
	}

	if antwort := NeueBindungen(anwendung).Grenzen(t.Context()); antwort.Fehler != nil ||
		antwort.Grenzen.Kontingente != domain.KontingenteAus {
		t.Errorf("Bindung: %+v", antwort)
	}

	// **Und ohne Bestand die Vorgabe** — mit Kontingenten.
	if antwort := NeueApp("").Grenzen(t.Context()); antwort.Fehler != nil ||
		antwort.Grenzen.Kontingente != domain.KontingenteAn {
		t.Errorf("ohne Bestand: %+v", antwort)
	}

	// **Ein Bestand, der nicht antwortet, sagt es** — neben der Vorgabe.
	if fehler := anwendung.db.Close(); fehler != nil {
		t.Fatal(fehler)
	}

	if antwort := anwendung.Grenzen(t.Context()); antwort.Fehler == nil ||
		antwort.Grenzen.Kontingente != domain.KontingenteAn {
		t.Errorf("ohne Antwort: %+v", antwort)
	}
}
