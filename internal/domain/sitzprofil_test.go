package domain_test

// **Ein Sitz, an ein Profil gebunden** — Werkstück G8 von Phase 5.5
// (#1465).
//
// # Wogegen hier geprüft wird
//
//   - Die Bindung ist **freiwillig**: ein Sitz ohne Profil bleibt ein Sitz.
//   - **Null findet niemanden** — sonst bekäme jemand die Sicht eines
//     ungebundenen Sitzes, den er nicht gewählt hat.
//   - Die **Sicht** ändert sich dadurch nicht: sie rechnet je Sitz, und was
//     einem anderen gehört, steht auch mit Profil nicht darin.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestSichtFuerProfil(t *testing.T) {
	t.Parallel()

	partie := domain.Partie{Wirt: "anna", Sitze: []domain.Sitz{
		{Kennung: "anna", Name: "Anna", ProfilID: 1},
		{Kennung: "bert", Name: "Bert", ProfilID: 2},
		{Kennung: "cem", Name: "Cem"},
	}}

	faelle := map[string]struct {
		profil   int64
		erwartet string
	}{
		"das erste Profil":          {profil: 1, erwartet: "anna"},
		"das zweite":                {profil: 2, erwartet: "bert"},
		"eines, das niemand hat":    {profil: 9, erwartet: "anna"},
		"die Null bekommt den Wirt": {profil: 0, erwartet: "anna"},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if gesehen := partie.SichtFuerProfil(fall.profil); gesehen != fall.erwartet {
				t.Fatalf("erwartet %q, bekam %q", fall.erwartet, gesehen)
			}
		})
	}
}

// **Der ungebundene Sitz ist nicht der Rückfall.**
//
// **Sonst wäre die Gegenprobe oben blind**: dort ist der Wirt zugleich der
// erste Sitz, und wer „den ersten ungebundenen" zurückgäbe statt des Wirts,
// käme bei „eines, das niemand hat" auf `cem` — hier fällt das auf.
func TestEinFremdesProfilBekommtDenWirtUndNichtDenFreienSitz(t *testing.T) {
	t.Parallel()

	partie := domain.Partie{Wirt: "bert", Sitze: []domain.Sitz{
		{Kennung: "anna", Name: "Anna"},
		{Kennung: "bert", Name: "Bert", ProfilID: 2},
	}}

	if gesehen := partie.SichtFuerProfil(9); gesehen != "bert" {
		t.Fatalf("ein fremdes Profil sitzt bei %q", gesehen)
	}
}

// **Ein Sitz ohne Profil ist ein Sitz** — die Bindung ist freiwillig.
func TestEineParteiOhneProfileGiltTrotzdem(t *testing.T) {
	t.Parallel()

	partie := einePartie()

	if fehler := partie.Pruefe(); fehler != nil {
		t.Fatalf("eine Partie ohne Profile wurde abgewiesen: %v", fehler)
	}

	partie.Sitze[0].ProfilID = 1

	if fehler := partie.Pruefe(); fehler != nil {
		t.Fatalf("eine Partie mit einem Profil wurde abgewiesen: %v", fehler)
	}
}

// **Die Sicht nennt das Profil jedes Sitzes** — damit das Fenster weiß,
// welchen es zeigen soll.
//
// **Und mehr ändert sie nicht**: was einem anderen Sitz gehört, steht auch mit
// Profil nicht darin. Das ist die dritte Zusage des Issues („private
// Information verschwindet beim Wechsel"), und sie hängt an [Partie.Sicht] —
// nicht am Profil.
func TestDieSichtNenntDasProfilJedesSitzes(t *testing.T) {
	t.Parallel()

	partie := einePartie()
	partie.ID = "p1"
	partie.Sitze[0].ProfilID = 2

	sicht := partie.Sicht("sitz-1")

	switch {
	case sicht.Sitze[0].ProfilID != 2:
		t.Errorf("das Profil des ersten Sitzes ist %d", sicht.Sitze[0].ProfilID)
	case sicht.Sitze[1].ProfilID != 0:
		t.Errorf("der ungebundene Sitz trägt %d", sicht.Sitze[1].ProfilID)
	}
}
