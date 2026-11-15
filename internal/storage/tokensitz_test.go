package storage_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Die Sitzbindung eines Tokens geht in den Bestand und zurück** —
// #1819.
func TestEineSitzbindungGehtInDenBestandUndZurueck(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueZugriffstokenablage(db)

	gebunden := probetoken("Am Sitz")
	gebunden.Sitzbindung = domain.Sitzbindung{Partie: "p1", Sitz: "anna"}

	angelegt, fehler := ablage.TokenAnlegen(t.Context(), gebunden, "gehütet:abc")
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	liste, lesefehler := ablage.Token(t.Context(), 1)
	if lesefehler != nil || len(liste) != 1 || liste[0].Sitzbindung != gebunden.Sitzbindung {
		t.Fatalf("die Liste: %+v (%v)", liste, lesefehler)
	}

	einzeln, _, einzelfehler := ablage.TokenMitKennung(t.Context(), angelegt.ID)
	if einzelfehler != nil || einzeln.Sitzbindung != gebunden.Sitzbindung {
		t.Errorf("einzeln: %+v (%v)", einzeln, einzelfehler)
	}

	// **Eine halbe Bindung hält auch der Bestand ab** — für einen Weg, der
	// an der Domäne vorbeiführt.
	if _, fehler := db.ExecContext(t.Context(),
		`UPDATE zugriffstoken SET sitz = '' WHERE id = ?`, angelegt.ID); fehler == nil {
		t.Error("eine Partie ohne Sitz ging in den Bestand")
	}
}

// **Ändern und Rotieren lassen die Bindung stehen** — sie entsteht mit dem
// Token, und ein Werkzeug, dessen Token jemand umbenennt, soll dadurch nicht
// an einen anderen Sitz rücken.
func TestEineSitzbindungUeberstehtAendernUndRotieren(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueZugriffstokenablage(db)
	bindung := domain.Sitzbindung{Partie: "p1", Sitz: "anna"}

	gebunden := probetoken("Am Sitz")
	gebunden.Sitzbindung = bindung

	angelegt, fehler := ablage.TokenAnlegen(t.Context(), gebunden, "gehütet:abc")
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	angelegt.Bezeichnung = "Umbenannt"
	angelegt.Sitzbindung = domain.Sitzbindung{Partie: "p2", Sitz: "bert"}

	// **Der Bestand zählt, nicht das Echo**: dass der Rückgabewert die
	// Bindung von vorher trägt, sorgt der Dienst (`Tokenverwaltung.Aendern`).
	if _, aenderfehler := ablage.TokenAendern(t.Context(), angelegt); aenderfehler != nil {
		t.Fatalf("ändern: %v", aenderfehler)
	}

	if gelesen, _, lesefehler := ablage.TokenMitKennung(t.Context(),
		angelegt.ID); lesefehler != nil || gelesen.Sitzbindung != bindung {
		t.Errorf("nach dem Ändern: %+v (%v)", gelesen.Sitzbindung, lesefehler)
	}

	rotiert, rotierfehler := ablage.TokenRotieren(t.Context(), angelegt.ID,
		"gehütet:def", "2026-09-26T00:00:00Z")
	if rotierfehler != nil || rotiert.Sitzbindung != bindung {
		t.Errorf("rotiert: %+v (%v)", rotiert.Sitzbindung, rotierfehler)
	}
}
