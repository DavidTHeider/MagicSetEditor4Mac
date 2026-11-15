package app

// **Der Vorrat räumt ganz, wenn er voll ist** — derselbe Zug wie bei der
// Bildquelle und dem Ebenenspeicher: wer die Grenze erreicht, hat keinen
// Arbeitssatz mehr, der hineinpasst (#2005).

import (
	"database/sql"
	"testing"
)

func TestDerWerkvorratRaeumtAnSeinerGrenze(t *testing.T) {
	vorrat := neuerWerkvorrat(2)

	gebaut := 0
	bauen := func() *Zeichenwerk {
		gebaut++

		return &Zeichenwerk{}
	}

	erster, zweiter, dritter := &sql.DB{}, &sql.DB{}, &sql.DB{}

	behalten := vorrat.werk(erster, "stand", bauen)
	vorrat.werk(zweiter, "stand", bauen)

	if vorrat.werk(erster, "stand", bauen) != behalten || gebaut != 2 {
		t.Fatalf("unter der Grenze neu gebaut — %d Bauten", gebaut)
	}

	// **Ein anderer Stand desselben Bestands ersetzt, er füllt nicht.**
	ersetzt := vorrat.werk(zweiter, "neuer stand", bauen)
	if ersetzt == behalten || vorrat.werk(erster, "stand", bauen) != behalten {
		t.Fatal("ein neuer Stand verdrängte den anderen Bestand")
	}

	vorrat.werk(dritter, "stand", bauen)

	if vorrat.werk(erster, "stand", bauen) == behalten {
		t.Error("über der Grenze blieb das alte Werk stehen")
	}
}
