package service_test

// **Die Ränder der Belegauswahl** — Werkstück A16 von Phase 5.3
// (#1515).
//
// **Drei Ränder, die der Weg über den Dienst nicht erreicht**: ein leerer
// Text, ein Treffer ohne Abschnittstext, und ein Manifest, das seine eigene
// Prüfung nicht besteht.

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein leerer Text kostet nichts.**
//
// **Und nicht ein Token**: die Aufrundung machte aus null Zeichen eines, und
// ein Budget zählte Belege mit, die es nicht gibt.
func TestEinLeererTextKostetNichts(t *testing.T) {
	t.Parallel()

	if kosten := service.Tokenschaetzung(""); kosten != 0 {
		t.Fatalf("ein leerer Text kostet %d", kosten)
	}

	// **Und ein Zeichen kostet eines** — aufgerundet, nicht abgerundet.
	if kosten := service.Tokenschaetzung("x"); kosten != 1 {
		t.Errorf("ein Zeichen kostet %d", kosten)
	}

	if kosten := service.Tokenschaetzung("12345"); kosten != 2 {
		t.Errorf("fünf Zeichen kosten %d", kosten)
	}
}

// **Ein Treffer ohne Abschnittstext trägt seinen Titel.**
//
// **Und nicht nichts**: ein Beleg, dessen Ausschnitt leer wäre, fiele aus dem
// Paket, und die Karte, die ihn trägt, wäre nicht belegt — obwohl sie
// gefunden wurde.
func TestEinTrefferOhneTextTraegtSeinenTitel(t *testing.T) {
	t.Parallel()

	ohneText := domain.Retrievaltreffer{
		Rang: 1,
		Dokument: domain.Retrievaldokument{
			Kennung: "regel:x", Quelle: domain.QuelleRegelwerk,
			Titel: "Ein Titel ohne Text", Sprache: "de",
			Quellrevision: "1", Abgerufen: dieRechercheuhr(),
		},
	}

	paket, fehler := einKontextdienst(&recherchequellendoppel{
		stand: einRegelstand(), treffer: []domain.Retrievaltreffer{ohneText},
	}).Zusammenstellen(t.Context(), einKontextwunsch())
	if fehler != nil {
		t.Fatalf("Zusammenstellen: %v", fehler)
	}

	if len(paket.Manifest.Belege) != 1 ||
		!strings.Contains(paket.Manifest.Belege[0].Ausschnitt, "Ein Titel") {
		t.Fatalf("die Belege sind %+v", paket.Manifest.Belege)
	}

	// **Und der Ausschnitt trägt keinen Zeilenumbruch ins Leere.**
	if strings.HasSuffix(paket.Manifest.Belege[0].Ausschnitt, "\n") {
		t.Errorf("der Ausschnitt endet mit einem Umbruch: %q",
			paket.Manifest.Belege[0].Ausschnitt)
	}
}

// **Ein Manifest, das seine eigene Prüfung nicht besteht, geht nicht hinaus.**
//
// # Der Fall, für den es die Zeile gibt
//
// **Eine Profilquelle, die null meldet und keinen Fehler.** Das ist die
// Gestalt eines Adapters, der nichts findet und es nicht sagt — und ein Paket
// ohne Profil gehörte niemandem. Die Prüfung des Manifests ist die letzte
// Stelle, an der das auffällt.
func TestEinManifestOhneProfilGehtNichtHinaus(t *testing.T) {
	t.Parallel()

	dienst := service.NeuerKontextdienst(
		einRecherchedienst(&recherchequellendoppel{stand: einRegelstand()}, nil),
		profildoppel{keines: true}).MitUhr(dieRechercheuhr)

	if _, fehler := dienst.Zusammenstellen(
		t.Context(), einKontextwunsch()); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Fatalf("das Paket ohne Profil kam durch: %v", fehler)
	}
}
