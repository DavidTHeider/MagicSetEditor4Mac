package domain_test

// **Technische Profile** — Werkstück 10 von Phase 4.12
// (#1351),
// Punkt 12.

import (
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ein technisches Profil ist ein Profil mit einer Marke.**
//
// **Und nicht eine eigene Entität.** Es hat einen Nutzer, einen Namen und
// eine Fassung wie jedes andere; was es unterscheidet, ist, **wofür** es da
// ist. Eine zweite Tabelle daneben hätte jede Abfrage über Profile verdoppelt.
func TestEinTechnischesProfilIstEinProfil(t *testing.T) {
	t.Parallel()

	eines := domain.Profil{NutzerID: 1, Name: "Sicherung", Technisch: true}

	if err := eines.Validate(); err != nil {
		t.Errorf("gemeldet wurde: %v", err)
	}
}

// **Sein Ablauf ist ein Zeitpunkt oder gar nichts** — wie beim Token.
func TestDerAblaufEinesProfilsIstEinZeitpunktOderNichts(t *testing.T) {
	t.Parallel()

	mit := domain.Profil{
		NutzerID: 1, Name: "Sicherung", Technisch: true,
		Ablauf: "2026-12-24T18:00:00Z",
	}
	if err := mit.Validate(); err != nil {
		t.Errorf("mit Ablauf abgewiesen: %v", err)
	}

	krumm := domain.Profil{
		NutzerID: 1, Name: "Sicherung", Technisch: true, Ablauf: "Heiligabend",
	}
	if fehler := krumm.Validate(); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("gemeldet wurde: %v", fehler)
	}
}

// **Ein Ablauf ohne technische Marke wird abgewiesen.**
//
// # Warum das eine Regel ist und keine Bequemlichkeit
//
// **Weil ein persönliches Profil nicht abläuft.** Ein Mensch hört nicht auf,
// dieser Mensch zu sein; ein Datum an seiner Arbeitsumgebung wäre eine
// Zusage, die niemand einlösen will — und die beim Ablesen wie eine Frist
// aussähe.
func TestEinAblaufOhneTechnischeMarkeWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	eines := domain.Profil{
		NutzerID: 1, Name: "Mein Profil", Ablauf: "2026-12-24T18:00:00Z",
	}

	if fehler := eines.Validate(); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("gemeldet wurde: %v", fehler)
	}
}

// **Abgelaufen ist, was vor jetzt liegt** — dieselbe Regel wie beim Token.
func TestEinAbgelaufenesProfilSagtEs(t *testing.T) {
	t.Parallel()

	stichtag := time.Date(2026, time.September, 5, 12, 0, 0, 0, time.UTC)

	abgelaufen := domain.Profil{
		NutzerID: 1, Name: "Sicherung", Technisch: true,
		Ablauf: "2026-09-04T12:00:00Z",
	}
	if !abgelaufen.Abgelaufen(stichtag) {
		t.Error("das abgelaufene Profil gilt als gültig")
	}

	// **Und ein persönliches läuft nie ab** — es hat keinen Ablauf.
	if (domain.Profil{NutzerID: 1, Name: "Ich"}).Abgelaufen(stichtag) {
		t.Error("ein persönliches Profil gilt als abgelaufen")
	}
}
