package storage_test

// **Der Deltaimport** — Werkstück A15 von Phase 5.3
// (#1514).
//
// **Was hier geprüft wird**: eine Generation lässt sich aus der aktiven
// aufbauen, ohne jeden Abschnitt neu einzubetten — und was in `ausser` steht,
// kommt **nicht** mit. Wer es danach nicht schreibt, hat es entfernt.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Eine neue Generation übernimmt, was sich nicht geändert hat.**
func TestEineNeueGenerationUebernimmtDasUnveraenderte(t *testing.T) {
	index := mitIndex(t,
		einIndexdokument("bolt", "Lightning Bolt deals 3 damage to any target."),
		einIndexdokument("shock", "Shock deals 2 damage to any target."),
		einIndexdokument("alt", "Old Text that will be replaced."))

	naechste, fehler := index.Anlegen(probekontext(t), einIndexmanifest())
	if fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	// **»alt« wird übergangen und danach nicht geschrieben** — damit ist es
	// weg. »shock« wird übergangen und neu geschrieben — damit ist es
	// geändert.
	uebernommen, fehler := index.Uebernehmen(probekontext(t), naechste,
		[]string{"alt", "shock"})
	if fehler != nil {
		t.Fatalf("Uebernehmen: %v", fehler)
	}

	if uebernommen != 1 {
		t.Fatalf("übernommen wurde %d", uebernommen)
	}

	if geschrieben := index.Schreiben(probekontext(t), naechste,
		einIndexdokument("shock", "Shock deals 2 damage. Now with new text."),
		[]float32{0, 1, 0}); geschrieben != nil {
		t.Fatalf("Schreiben: %v", geschrieben)
	}

	if aktiviert := index.Aktivieren(probekontext(t), naechste); aktiviert != nil {
		t.Fatalf("Aktivieren: %v", aktiviert)
	}

	// **Der übernommene Abschnitt ist lexikalisch da** — samt seinem Text.
	treffer, fehler := index.Lexikalisch(probekontext(t), "lightning", 10)
	if fehler != nil {
		t.Fatalf("Lexikalisch: %v", fehler)
	}

	if len(treffer) != 1 || treffer[0].Dokument.Kennung != "bolt" {
		t.Fatalf("die Treffer sind %+v", treffer)
	}

	// **Und sein Vektor ist mitgewandert** — sonst wäre er neu einzubetten,
	// und genau das soll der Deltaimport sparen.
	semantisch, fehler := index.Semantisch(probekontext(t), []float32{1, 0, 0}, 10)
	if fehler != nil {
		t.Fatalf("Semantisch: %v", fehler)
	}

	if len(semantisch) != 2 {
		t.Fatalf("semantisch kamen %+v", semantisch)
	}

	// **»alt« ist weg.**
	weg, fehler := index.Lexikalisch(probekontext(t), "replaced", 10)
	if fehler != nil {
		t.Fatalf("Lexikalisch: %v", fehler)
	}

	if len(weg) != 0 {
		t.Fatalf("der entfernte Abschnitt kam durch: %+v", weg)
	}

	// **Und »shock« trägt den neuen Text.**
	neu, fehler := index.Lexikalisch(probekontext(t), "new", 10)
	if fehler != nil {
		t.Fatalf("Lexikalisch: %v", fehler)
	}

	if len(neu) != 1 || neu[0].Dokument.Kennung != "shock" {
		t.Fatalf("die Treffer sind %+v", neu)
	}
}

// **Ohne aktive Generation gibt es nichts zu übernehmen.**
func TestOhneAktiveGenerationGibtEsNichtsZuUebernehmen(t *testing.T) {
	index := storage.NeuerRetrievalindex(frischeDatenbank(t))

	generation, fehler := index.Anlegen(probekontext(t), einIndexmanifest())
	if fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	if _, fehler := index.Uebernehmen(probekontext(t), generation, nil); fehler == nil {
		t.Fatal("ohne aktive Generation kam die Übernahme durch")
	}
}

// **Eine Übernahme ohne Ausnahmen nimmt alles.**
func TestEineUebernahmeOhneAusnahmenNimmtAlles(t *testing.T) {
	index := mitIndex(t,
		einIndexdokument("bolt", "Lightning Bolt deals 3 damage."),
		einEigenesDokument("meins", "Mein eigener Text", 1))

	naechste, fehler := index.Anlegen(probekontext(t), einIndexmanifest())
	if fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	uebernommen, fehler := index.Uebernehmen(probekontext(t), naechste, nil)
	if fehler != nil {
		t.Fatalf("Uebernehmen: %v", fehler)
	}

	if uebernommen != 2 {
		t.Fatalf("übernommen wurde %d", uebernommen)
	}

	if aktiviert := index.Aktivieren(probekontext(t), naechste); aktiviert != nil {
		t.Fatalf("Aktivieren: %v", aktiviert)
	}

	// **Auch das profilgebundene Dokument wandert mit — samt seiner Sicht.**
	treffer, fehler := index.Lexikalisch(probekontext(t), "eigener", 10)
	if fehler != nil {
		t.Fatalf("Lexikalisch: %v", fehler)
	}

	if len(treffer) != 1 ||
		treffer[0].Dokument.Sichtbarkeit != domain.SichtbarProfil {
		t.Fatalf("die Treffer sind %+v", treffer)
	}
}

// **Eine Sperre schlägt den Deltaimport.**
//
// # Der Satz, um den es geht
//
// **„Widerruf/Löschung schlägt ‚letzter guter Index'"** — das Issue sagt es
// als eigene Zeile, und der Deltaimport ist der Fall, in dem es schiefgehen
// könnte: er **kopiert** Zeilen aus der alten Generation, und eine gesperrte
// käme mit, wenn die Sperre an der Zeile hinge statt an der Kennung.
//
// **Sie hängt an der Kennung**, und die wandert mit. Deshalb bleibt der
// Inhalt auch in der neuen Generation gesperrt — ohne dass der Import davon
// wissen muss.
func TestEineSperreSchlaegtDenDeltaimport(t *testing.T) {
	index := mitIndex(t,
		einIndexdokument("bolt", "Lightning Bolt deals 3 damage."),
		einIndexdokument("weg", "This one gets revoked."))

	if fehler := index.Sperren(probekontext(t), "weg", 0); fehler != nil {
		t.Fatalf("Sperren: %v", fehler)
	}

	naechste, fehler := index.Anlegen(probekontext(t), einIndexmanifest())
	if fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	if _, uebernommen := index.Uebernehmen(probekontext(t), naechste,
		nil); uebernommen != nil {
		t.Fatalf("Uebernehmen: %v", uebernommen)
	}

	if aktiviert := index.Aktivieren(probekontext(t), naechste); aktiviert != nil {
		t.Fatalf("Aktivieren: %v", aktiviert)
	}

	// **Der gesperrte Abschnitt kommt auch aus der neuen Generation nicht
	// heraus** — weder lexikalisch noch semantisch.
	for name, suchen := range map[string]func() ([]domain.Retrievaltreffer, error){
		"lexikalisch": func() ([]domain.Retrievaltreffer, error) {
			return index.Lexikalisch(probekontext(t), "revoked", 10)
		},
		"semantisch": func() ([]domain.Retrievaltreffer, error) {
			return index.Semantisch(probekontext(t), []float32{0, 1, 0}, 10)
		},
	} {
		t.Run(name, func(t *testing.T) {
			treffer, gesucht := suchen()
			if gesucht != nil {
				t.Fatalf("%s: %v", name, gesucht)
			}

			for _, einzelner := range treffer {
				if einzelner.Dokument.Kennung == "weg" {
					t.Fatalf("der gesperrte Abschnitt kam %s durch", name)
				}
			}
		})
	}

	// **Und der andere ist noch da** — die Gegenprobe zur Sperre: ein Lauf,
	// in dem gar nichts gefunden wird, sagt nichts über sie.
	treffer, fehler := index.Lexikalisch(probekontext(t), "lightning", 10)
	if fehler != nil || len(treffer) != 1 {
		t.Fatalf("die Treffer sind %+v (%v)", treffer, fehler)
	}
}

// **Zweimal dieselbe Übernahme in dieselbe Generation scheitert.**
//
// **Und das ist die Zusage der Tafel**: ein Abschnitt steht je Generation
// **einmal** (`UNIQUE (generation, kennung)`). Zweimal derselbe wäre ein
// Treffer, der doppelt zählt — und genau das könnte ein wiederaufgenommener
// Deltaimport versuchen.
func TestZweimalDieselbeUebernahmeScheitert(t *testing.T) {
	index := mitIndex(t, einIndexdokument("bolt", "Lightning Bolt deals 3 damage."))

	naechste, fehler := index.Anlegen(probekontext(t), einIndexmanifest())
	if fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	if _, erste := index.Uebernehmen(probekontext(t), naechste, nil); erste != nil {
		t.Fatalf("die erste Übernahme: %v", erste)
	}

	if _, zweite := index.Uebernehmen(probekontext(t), naechste, nil); zweite == nil {
		t.Fatal("die zweite Übernahme kam durch")
	}
}
