package storage_test

// **Verweist irgendein Set auf eine Quelle?** — für das Aufräumen des
// Eingangs am Dienst (#1944). Über **alle** Profile: eine Datei, die ein
// fremdes Profil noch braucht, darf nicht fort.

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestEineQuelleIstVerwiesenAuchAusEinemFremdenProfil(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	fremd := storage.MitProfil(t.Context(), zweitesProfil(t, storage.NeueNutzerablage(db), "Fremd"))

	if _, err := storage.NeuerSetSpeicher(db).SetAnlegen(fremd, &domain.Set{
		Name: "Fremdes", Game: "magic", Origin: "/dienst/eingang/abc.mse-set",
	}, nil); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	leser := storage.NeuerSetLeser(db)

	for pfad, erwartet := range map[string]bool{
		"/dienst/eingang/abc.mse-set": true,
		"/dienst/eingang/def.mse-set": false,
		"":                            false,
	} {
		if verwiesen, err := leser.HerkunftVerwiesen(t.Context(), pfad); err != nil || verwiesen != erwartet {
			t.Errorf("%q: %v, %v — erwartet %v", pfad, verwiesen, err, erwartet)
		}
	}

	// **Ein Fehler ist kein „nein"** — sonst räumte der Eingang eine Quelle
	// ab, weil die Datenbank gerade nicht antwortete.
	abgebrochen, abbrechen := context.WithCancel(t.Context())
	abbrechen()

	if _, err := leser.HerkunftVerwiesen(abgebrochen, "/x"); err == nil {
		t.Error("eine abgebrochene Frage verwies still auf nichts")
	}
}
