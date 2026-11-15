package flaechen_test

// **Eine vergessene Messung zeigt sich an der Messung** — C04-Rest aus dem
// Nachlauf über Phase C
// (#2284).
//
// Der Rollenlauf prüfte beim Vergessen nur Erfolg oder Misserfolg; eine
// Messung lag gar nicht da, und das Vergessen war ein Leerlauf. Jetzt liegt
// eine im Profil der Lage: wer darf, nimmt sie weg, wer nicht darf, lässt sie
// stehen.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// merkeEineMessung legt im gezeigten Profil eine Messung von `gpt-probe` ab
// und gibt das Profil.
func merkeEineMessung(t *testing.T, umgebung []string) int64 {
	t.Helper()

	db, fehler := storage.Oeffnen(t.Context(), bestandsortAus(t, umgebung))
	if fehler != nil {
		t.Fatalf("den Bestand öffnen: %v", fehler)
	}

	defer func() { _ = db.Close() }() //nolint:errcheck // gelesen und geschrieben ist es

	imProfil, profilfehler := storage.MitGezeigtemProfil(t.Context(), db)
	if profilfehler != nil {
		t.Fatalf("das Profil: %v", profilfehler)
	}

	profil, _ := storage.ProfilAus(imProfil) //nolint:errcheck // eben gesetzt

	if fehler := storage.NeueModellprobenablage(db).Merken(imProfil, profil, domain.KIModellprobe{
		Anbieter: "openai", Alias: "gpt-probe", Faehigkeit: domain.KIStrukturierteGenerierung,
		Datum: "2026-10-10", Beleg: "Probeauftrag der Rollenläufe", Laufzeit: "Responses-API /v1/responses",
		Modell: "gpt-probe-2026-09-01",
	}); fehler != nil {
		t.Fatalf("die Messung merken: %v", fehler)
	}

	return profil
}

// pruefeMessung sieht nach, ob die Messung noch dasteht — genau dann, wenn
// die Lage nicht schreiben durfte.
func pruefeMessung(t *testing.T, umgebung []string, lage rollenlage, profil int64) {
	t.Helper()

	db, fehler := storage.Oeffnen(t.Context(), bestandsortAus(t, umgebung))
	if fehler != nil {
		t.Fatalf("den Bestand öffnen: %v", fehler)
	}

	defer func() { _ = db.Close() }() //nolint:errcheck // nur gelesen

	alle, lesefehler := storage.NeueModellprobenablage(db).Alle(storage.MitProfil(t.Context(), profil), profil)
	if lesefehler != nil {
		t.Fatalf("die Messungen lesen: %v", lesefehler)
	}

	steht := false
	for _, probe := range alle {
		steht = steht || (probe.Anbieter == "openai" && probe.Alias == "gpt-probe")
	}

	if steht == lage.darf {
		t.Errorf("die Messung steht noch: %v — erwartet %v", steht, !lage.darf)
	}
}
