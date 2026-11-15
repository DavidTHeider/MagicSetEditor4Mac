package storage_test

// **Ein technisches Profil im Bestand** — Werkstück 10 von Phase 4.12
// (#1351),
// Punkt 12.
//
// **Ohne `t.Parallel()`**, wie jeder Lauf in diesem Paket: sie teilen eine
// Datenquelle im Arbeitsspeicher, und nebenläufig meldet SQLite „database
// table is locked".

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Was beim Anlegen mitgegeben wird, steht danach im Bestand.**
//
// **Und das ist keine Selbstverständlichkeit gewesen**: `ProfilAnlegen` schrieb
// bis zu diesem Werkstück nur `nutzer_id` und `name` — und gab die übrigen
// Angaben trotzdem zurück, als stünden sie da. Wer den Rückgabewert prüfte,
// sah nichts; wer neu las, fand sie leer. Angestrichen hat es dieser Lauf.
func TestEinAngelegtesProfilBehaeltSeineAngaben(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))
	ctx := t.Context()

	angelegt, fehler := ablage.ProfilAnlegen(ctx, domain.Profil{
		NutzerID:          1,
		Name:              "Nachtlauf",
		Subslogan:         "Bedient sich selbst",
		Kurzebeschreibung: "Kurz.",
		Langebeschreibung: "Lang.",
		Badge:             "profil-2.png",
		Technisch:         true,
		Ablauf:            "2027-01-01T00:00:00Z",
	})
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	profile, lesefehler := ablage.Profile(ctx, 1)
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	var gelesen domain.Profil

	for _, eines := range profile {
		if eines.ID == angelegt.ID {
			gelesen = eines
		}
	}

	if gelesen.ID == 0 {
		t.Fatalf("das angelegte Profil %d steht nicht in %+v", angelegt.ID, profile)
	}

	faelle := map[string]struct{ ist, soll string }{
		"Subslogan":         {gelesen.Subslogan, "Bedient sich selbst"},
		"Kurzebeschreibung": {gelesen.Kurzebeschreibung, "Kurz."},
		"Langebeschreibung": {gelesen.Langebeschreibung, "Lang."},
		"Badge":             {gelesen.Badge, "profil-2.png"},
		"Ablauf":            {gelesen.Ablauf, "2027-01-01T00:00:00Z"},
	}

	for name, fall := range faelle {
		if fall.ist != fall.soll {
			t.Errorf("%s = %q, erwartet %q", name, fall.ist, fall.soll)
		}
	}

	if !gelesen.Technisch {
		t.Error("das Profil ist nicht technisch")
	}
}

// **Ein Profil lässt sich zum technischen machen und zurück.**
func TestTechnischUndAblaufLassenSichAendern(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))
	ctx := t.Context()

	profile, fehler := ablage.Profile(ctx, 1)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	geaendert := profile[0]
	geaendert.Technisch = true
	geaendert.Ablauf = "2026-12-24T18:00:00Z"

	if _, err := ablage.ProfilAendern(ctx, geaendert); err != nil {
		t.Fatalf("ändern: %v", err)
	}

	nachher, lesefehler := ablage.Profile(ctx, 1)
	if lesefehler != nil {
		t.Fatalf("neu lesen: %v", lesefehler)
	}

	if !nachher[0].Technisch {
		t.Error("Technisch = false")
	}

	if nachher[0].Ablauf != "2026-12-24T18:00:00Z" {
		t.Errorf("Ablauf = %q", nachher[0].Ablauf)
	}

	t.Run("und wieder zurück", func(t *testing.T) {
		// **Der Beleg, dass die Spalte wirklich geschrieben wird.** Ein
		// `UPDATE`, das `technisch` gar nicht nennt, ließe den Lauf oben
		// scheitern — eines, das es auf 1 festnagelte, käme durch (§10).
		zurueck := nachher[0]
		zurueck.Technisch = false
		zurueck.Ablauf = ""

		if _, err := ablage.ProfilAendern(ctx, zurueck); err != nil {
			t.Fatalf("zurückändern: %v", err)
		}

		wieder, err := ablage.Profile(ctx, 1)
		if err != nil {
			t.Fatalf("wieder lesen: %v", err)
		}

		if wieder[0].Technisch || wieder[0].Ablauf != "" {
			t.Errorf("Technisch = %v, Ablauf = %q",
				wieder[0].Technisch, wieder[0].Ablauf)
		}
	})
}
