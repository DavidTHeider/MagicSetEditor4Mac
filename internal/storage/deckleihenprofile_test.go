package storage_test

// **Eine Leihe von einem Profil ins andere** — Befund 1 des Zweitblicks über
// Phase 6.1.
//
// Wer zwei Profile in einem Bestand führt, hat zwei Peeridentitäten (0066)
// und kann von einem ins andere leihen. Dann liegt **dieselbe** Leihe zweimal
// im Bestand — beim Verleiher mit `verliehen`, beim Entleiher ohne —, und
// jede Seite schreibt ihren eigenen Stand.

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestEineLeiheZwischenZweiProfilenBleibtZweiZeilen(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	verleiher := probekontext(t)

	if _, fehler := db.ExecContext(verleiher,
		`INSERT INTO profile (id, nutzer_id, name) VALUES (2, 1, 'Zweites')`); fehler != nil {
		t.Fatalf("das zweite Profil: %v", fehler)
	}

	entleiher := storage.MitProfil(context.WithoutCancel(verleiher), 2)
	ablage := storage.NeueLeihablage(db)

	angeboten := einLeihstand(func(stand *domain.Leihstand) { stand.Verliehen = true })
	if fehler := ablage.LeiheSchreiben(verleiher, angeboten); fehler != nil {
		t.Fatalf("anbieten: %v", fehler)
	}

	// **Der Entleiher nimmt an und gibt sie gleich zurück** — ein Stand,
	// der sich vom Verleiher unterscheidet.
	angenommen := einLeihstand(func(stand *domain.Leihstand) { stand.Widerrufen = true })
	if fehler := ablage.LeiheSchreiben(entleiher, angenommen); fehler != nil {
		t.Fatalf("annehmen: %v", fehler)
	}

	beimVerleiher, fehler := ablage.Leihe(verleiher, "l1")
	if fehler != nil {
		t.Fatalf("beim Verleiher lesen: %v", fehler)
	}

	if !beimVerleiher.Verliehen || beimVerleiher.Widerrufen {
		t.Errorf("der Verleiher sieht %+v — der Entleiher hat in sein Profil geschrieben",
			beimVerleiher)
	}

	beimEntleiher, fehler := ablage.Leihe(entleiher, "l1")
	if fehler != nil {
		t.Fatalf("beim Entleiher lesen: %v — er hat keine eigene Zeile bekommen", fehler)
	}

	if beimEntleiher.Verliehen || !beimEntleiher.Widerrufen {
		t.Errorf("der Entleiher sieht %+v", beimEntleiher)
	}
}
