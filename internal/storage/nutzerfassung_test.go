package storage_test

// **Das Angelegte trägt die Fassung des Bestands** — Befund aus dem
// Vertragsrundgang (B01, #2220): `NutzerAnlegen` und `ProfilAnlegen` gaben
// `0` zurück, der Bestand führt `1` (Spaltenvorgabe aus 0040). Wer die Antwort
// an `NutzerAendern` oder `ProfilAendern` weiterreichte, bekam „Stand
// überholt" — für einen Stand, den niemand überholt hatte.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestEinAngelegterNutzerTraegtDieFassungDesBestands(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))

	angelegt, err := ablage.NutzerAnlegen(probekontext(t), domain.Nutzer{Name: "Zweite"})
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	alle, err := ablage.Nutzer(probekontext(t))
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	for _, gelesen := range alle {
		if gelesen.ID == angelegt.ID && gelesen.Fassung != angelegt.Fassung {
			t.Errorf("angelegt mit Fassung %d, der Bestand führt %d", angelegt.Fassung, gelesen.Fassung)
		}
	}

	angelegt.Name = "Zweite, umbenannt"
	if _, err := ablage.NutzerAendern(probekontext(t), angelegt); err != nil {
		t.Errorf("die Antwort des Anlegens taugt nicht zum Ändern: %v", err)
	}
}

func TestEinAngelegtesProfilTraegtDieFassungDesBestands(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))

	angelegt, err := ablage.ProfilAnlegen(probekontext(t), domain.Profil{NutzerID: 1, Name: "Zweites"})
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	angelegt.Name = "Zweites, umbenannt"
	if _, err := ablage.ProfilAendern(probekontext(t), angelegt); err != nil {
		t.Errorf("die Antwort des Anlegens taugt nicht zum Ändern: %v", err)
	}
}
