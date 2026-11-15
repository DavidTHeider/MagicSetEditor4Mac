package flaechen_test

// **Ein abgewiesener Gegenbeleg, über beide Flächen** —
// #2500.
//
// Festgehalten wird er von der Post, wenn über eine Leitung ein Beleg ankommt,
// der nicht passt (`service/handelsgegenbeleg.go`); hier wird er in den
// Bestand gelegt und gelesen. **Gemessen wird, dass Kommandozeile und Dienst
// dasselbe sagen** — Absender, Abdruck, Mangel, Grund und Uhr.

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// befundGleich hält einen Befund am Beleg fest und liest ihn über beide
// Flächen.
func befundGleich(t *testing.T, umgebung []string, an dienst) {
	t.Helper()

	db, fehler := storage.Oeffnen(t.Context(), bestandsortAus(t, umgebung))
	if fehler != nil {
		t.Fatalf("den Bestand öffnen: %v", fehler)
	}

	ctx, profilfehler := storage.MitGezeigtemProfil(t.Context(), db)
	if profilfehler != nil {
		t.Fatalf("das Profil lesen: %v", profilfehler)
	}

	abgewiesen := storage.NeueAbschlussablage(db).Abweisen(ctx, domain.Gegenbelegbefund{
		Vorgang: derHandel, Von: dasGegenueberImHandel, Abdruck: "abdruck-drueben",
		Mangel: domain.MangelZeilen, Grund: "der Beleg trägt 1 Zeilen und meiner 2",
	})

	_ = db.Close() //nolint:errcheck // nur zum Aufräumen

	if abgewiesen != nil {
		t.Fatalf("abweisen: %v", abgewiesen)
	}

	type befund struct{ Von, Abdruck, Mangel, Grund, Empfangen string }

	var ueberCLI, ueberREST struct{ Befund *befund }

	for ziel, roh := range map[*struct{ Befund *befund }]string{
		&ueberCLI: mussKommandozeile(t, umgebung, "--json", "handelsbeleg", derHandel),
		&ueberREST: mussREST(t, an, http.MethodGet,
			strings.ReplaceAll(pfadAus(derBelegweg), "{vorgang}", derHandel), ""),
	} {
		if fehler := json.Unmarshal([]byte(roh), ziel); fehler != nil {
			t.Fatalf("kein Beleg: %v\n%s", fehler, roh)
		}
	}

	if ueberCLI.Befund == nil || ueberREST.Befund == nil {
		t.Fatalf("ein Befund fehlt: Kommandozeile %v, Dienst %v", ueberCLI.Befund, ueberREST.Befund)
	}

	if *ueberCLI.Befund != *ueberREST.Befund || ueberCLI.Befund.Mangel != "zeilen" ||
		ueberCLI.Befund.Von != dasGegenueberImHandel || ueberCLI.Befund.Empfangen == "" {
		t.Errorf("die Kommandozeile liest %+v, der Dienst %+v", *ueberCLI.Befund, *ueberREST.Befund)
	}
}
