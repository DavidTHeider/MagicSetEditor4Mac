package domain_test

// **Woher ein Paket stammt und wer es braucht** — DOM-01.AK-01
// (#1696).

import (
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestJedesPaketHatEineHerkunftsart(t *testing.T) {
	t.Parallel()

	const verwaltet = "/Nutzer/pakete"

	faelle := map[string]struct {
		paket domain.Paket
		art   domain.Paketherkunftsart
	}{
		"bezogen — mit Prüfsumme, gleich wo": {
			domain.Paket{Pfad: "/Nutzer/bezug/m15/magic.mse-game", Herkunft: "https://example.org/m15", Pruefsumme: "sha256:ab"},
			domain.PaketBezogen,
		},
		"übernommen — mit Herkunft, ohne Prüfsumme": {
			domain.Paket{Pfad: verwaltet + "/magic.mse-game", Herkunft: "/Applications/MSE2/data"},
			domain.PaketUebernommen,
		},
		"eigen — im verwalteten Verzeichnis, ohne Herkunft": {
			domain.Paket{Pfad: verwaltet + "/eigen.mse-style"},
			domain.PaketEigen,
		},
		"mitgeliefert — außerhalb, ohne Herkunft": {
			domain.Paket{Pfad: "/Programme/MagicSetEditor4Mac.app/saat/probe.mse-game"},
			domain.PaketMitgeliefert,
		},
		"ein Nachbarverzeichnis mit demselben Anfang ist nicht das verwaltete": {
			domain.Paket{Pfad: verwaltet + "-alt/eigen.mse-style"},
			domain.PaketMitgeliefert,
		},
	}

	// **Ohne verwaltetes Verzeichnis gibt es nichts Eigenes** — eine Fläche,
	// die keines kennt, hält jedes Paket ohne Herkunft für mitgeliefert.
	if art := (domain.Paket{Pfad: "/x/eigen.mse-style"}).HerkunftsartIn(""); art != domain.PaketMitgeliefert {
		t.Errorf("ohne Verzeichnis: %q", art)
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if art := fall.paket.HerkunftsartIn(verwaltet); art != fall.art {
				t.Errorf("Herkunftsart %q, erwartet %q", art, fall.art)
			}
		})
	}
}

// **Wer ein Paket braucht** — die Umkehrung von `depends on:`, sortiert und
// ohne das Paket selbst.
func TestGebrauchtVonNenntJedenAbhaengigen(t *testing.T) {
	t.Parallel()

	katalog := []domain.Paket{
		{Art: "mse-game", Name: "magic"},
		{Art: "mse-style", Name: "magic-m15", Braucht: []domain.Bedarf{{Paket: "magic.mse-game"}}},
		{Art: "mse-style", Name: "magic-alt", Braucht: []domain.Bedarf{
			{Paket: "magic.mse-game"}, {Paket: "magic-mana-small.mse-symbol-font"},
		}},
		{Art: "mse-symbol-font", Name: "magic-mana-small"},
	}

	if von := domain.GebrauchtVon(katalog, katalog[0]); !slices.Equal(von, []string{"magic-alt.mse-style", "magic-m15.mse-style"}) {
		t.Errorf("das Spiel brauchen %v", von)
	}

	if von := domain.GebrauchtVon(katalog, katalog[1]); len(von) != 0 {
		t.Errorf("den Stil braucht niemand, und doch %v", von)
	}

	if datei := katalog[3].Dateiname(); datei != "magic-mana-small.mse-symbol-font" {
		t.Errorf("der Dateiname ist %q", datei)
	}
}
