package domain_test

// **Die Angaben eines Profils** — Werkstück 9 von Phase 4.12
// (#1351),
// Punkte 10 bis 13.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Alle vier Angaben sind freiwillig.**
//
// **Pflicht bleibt der Name.** Ein Profil ohne Beschreibung ist ein Profil;
// eines ohne Namen wäre eine Zeile ohne Aufschrift.
func TestEinProfilBrauchtNurSeinenNamen(t *testing.T) {
	t.Parallel()

	if err := (domain.Profil{NutzerID: 1, Name: "Turnier"}).Validate(); err != nil {
		t.Errorf("gemeldet wurde: %v", err)
	}
}

// **Ein Badge ist ein Dateiname und kein Pfad.**
//
// **Dieselbe Zusage wie beim Profilbild** (Werkstück 8): der Bestand führt
// den Namen, die Datei liegt unter `beiwerk/nutzer/`. Wer einen Pfad
// hineinschriebe, zeigte aus dem Bildverzeichnis heraus.
func TestEinProfilbadgeIstEinDateiname(t *testing.T) {
	t.Parallel()

	faelle := map[string]string{
		"mit Schrägstrich": "../../etc/passwd",
		"mit Rückstrich":   `bilder\badge.png`,
		"nur zwei Punkte":  "..",
		"absoluter Pfad":   "/tmp/badge.png",
	}

	for name, badge := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fehler := (domain.Profil{NutzerID: 1, Name: "Turnier", Badge: badge}).Validate()
			if !errors.Is(fehler, domain.ErrValidation) {
				t.Errorf("gemeldet wurde: %v", fehler)
			}
		})
	}

	// **Und ein gewöhnlicher Name kommt durch** — sonst prüfte der Lauf
	// darüber nur, dass gar nichts durchkommt (§10).
	if err := (domain.Profil{
		NutzerID: 1, Name: "Turnier", Badge: "profil-7.png",
	}).Validate(); err != nil {
		t.Errorf("der gewöhnliche Name wurde abgewiesen: %v", err)
	}
}

// **Der Subslogan hat eine Grenze.**
//
// **Weil er in eine Zeile unter den Namen gehört.** Ohne Grenze stünde dort
// ein Absatz, und die Zeile darunter verschöbe sich für jeden.
func TestEinSubsloganHatEineGrenze(t *testing.T) {
	t.Parallel()

	zulang := make([]byte, domain.HoechstlaengeSubslogan+1)
	for i := range zulang {
		zulang[i] = 'a'
	}

	fehler := (domain.Profil{
		NutzerID: 1, Name: "Turnier", Subslogan: string(zulang),
	}).Validate()
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("gemeldet wurde: %v", fehler)
	}

	// **Genau auf der Grenze kommt es durch** — die Gegenprobe zur Zeile
	// darüber: ein Wächter, der jede Länge abwiese, wäre auch grün.
	if err := (domain.Profil{
		NutzerID: 1, Name: "Turnier", Subslogan: string(zulang[:len(zulang)-1]),
	}).Validate(); err != nil {
		t.Errorf("die erlaubte Länge wurde abgewiesen: %v", err)
	}
}
