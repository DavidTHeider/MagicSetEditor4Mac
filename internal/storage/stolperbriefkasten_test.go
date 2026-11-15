package storage

// **Die Fehlerwege des Briefkastens** — Werkstück `T5` von Phase 5.6
// (#1470).
//
// # Warum es dafür Haken braucht
//
// **Weil eine krumme Uhr kein echter Treiber hergibt.** Der Briefkasten liest
// seine eigene Uhr aus der Zeile zurück; steht dort etwas, das kein Zeitpunkt
// ist, muss das eine Meldung geben und keinen Nullwert, der später wie „vor
// langer Zeit behandelt" aussieht — und damit wie ein Eintrag, den man
// wegräumen darf.

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Die Stellen, an denen der Briefkasten stolpern kann.
const (
	beimBriefkastenscan stolperstelle = iota + 2500
	beimBriefkastenuhr
)

func init() {
	sql.Register("stolper-briefkasten-scan",
		stolpertreiber{stelle: beimBriefkastenscan})
	sql.Register("stolper-briefkasten-uhr",
		stolpertreiber{stelle: beimBriefkastenuhr})
}

// briefkastenabfrage beantwortet die eine Abfrage des Briefkastens.
func briefkastenabfrage(
	stelle stolperstelle, abfrage string,
) (driver.Rows, bool) {
	if !strings.HasPrefix(strings.TrimSpace(abfrage),
		"SELECT abdruck, ergebnis, behandelt") {
		return nil, false
	}

	spalten := []string{"abdruck", "ergebnis", "behandelt"}

	switch stelle {
	case beimBriefkastenscan:
		return &krummeZeile{spalten: spalten}, true

	case beimBriefkastenuhr:
		// **Eine Uhr, die keine ist.** Genau so sähe eine Zeile aus, die
		// jemand von Hand geschrieben hat.
		return &werteZeile{
			spalten: spalten,
			werte:   []driver.Value{"abdruck-1", "ja", "keine Uhr"},
		}, true
	}

	return nil, false
}

func TestEineKrummeBriefkastenzeileMeldetSich(t *testing.T) {
	for name, treiber := range map[string]string{
		"eine Zeile, die sich nicht lesen lässt": "stolper-briefkasten-scan",
		"eine Uhr, die keine ist":                "stolper-briefkasten-uhr",
	} {
		t.Run(name, func(t *testing.T) {
			db := stolperbestand(t, treiber)

			if _, fehler := NeueBriefkastenablage(db).
				Nachschlagen(probekontext(t), derSchluesselIntern); fehler == nil {
				t.Error("die krumme Zeile kam durch")
			}
		})
	}
}

// **Ohne Profil gibt es keinen Briefkasten** (0066) — auf beiden Wegen.
func TestOhneProfilGibtEsKeinenBriefkasten(t *testing.T) {
	ablage := NeueBriefkastenablage(migrierteDatenbank(t))

	if _, fehler := ablage.Nachschlagen(
		context.Background(), derSchluesselIntern); fehler == nil {
		t.Error("das Nachschlagen kam ohne Profil durch")
	}

	if fehler := ablage.Vermerken(context.Background(), "v-1",
		domain.Inboxeintrag{Schluessel: derSchluesselIntern}); fehler == nil {
		t.Error("das Vermerken kam ohne Profil durch")
	}
}

// derSchluesselIntern ist der Schlüssel dieser Läufe.
const derSchluesselIntern = "1\x00v-1\x00vorschlag\x002"
