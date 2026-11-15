package domain_test

// **Die Fehlerwege des Kampfes** — Werkstück M6 der Regelmaschine
// (#1608).
//
// # Warum sie einzeln stehen
//
// **Weil jeder von ihnen eine eigene Auskunft gibt** (§8): „geht nicht" sagt
// niemandem, ob seine Kreatur getappt ist, ob sie ihm nicht gehört oder ob
// der Schritt der falsche ist.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Jeder Weg, auf dem kein Angreifer erklärt wird, sagt warum.**
func TestJederFehlerwegDerAngreifererklaerungSagtWarum(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		bauen    func(t *testing.T) domain.Partie
		wer      string
		angriffe []domain.Angriff
		wort     string
	}{
		"der falsche Schritt": {
			bauen: vorDemKampf, wer: "sitz-1", wort: "Angreiferschritt",
			angriffe: nil,
		},

		"der falsche Sitz": {
			bauen: imAngreiferschritt, wer: "sitz-2", wort: "CR 506.2",
		},

		"zweimal erklärt": {
			bauen: schonErklaert, wer: "sitz-1", wort: "schon erklärt",
		},

		"eine Kreatur, die es nicht gibt": {
			bauen: imAngreiferschritt, wer: "sitz-1",
			angriffe: []domain.Angriff{{
				Kreatur: "gibtsnicht", Ziel: "sitz-2",
			}},
			wort: "gibt es in dieser Partie nicht",
		},

		"eine fremde Kreatur": {
			bauen: imAngreiferschritt, wer: "sitz-1",
			angriffe: []domain.Angriff{{
				Kreatur: "sitz-2/baer/0", Ziel: "sitz-2",
			}},
			wort: "liegt nicht auf dem Schlachtfeld",
		},

		"ein Land als Angreifer": {
			bauen: mitLandImKampf, wer: "sitz-1",
			angriffe: []domain.Angriff{{
				Kreatur: "sitz-1/wald/0", Ziel: "sitz-2",
			}},
			wort: "ist keine Kreatur",
		},

		"ein Ziel, das es nicht gibt": {
			bauen: imAngreiferschritt, wer: "sitz-1",
			angriffe: []domain.Angriff{{
				Kreatur: "sitz-1/riese/0", Ziel: "gibtsnicht",
			}},
			wort: "kein zulässiges Angriffsziel",
		},

		"sich selbst angreifen": {
			bauen: imAngreiferschritt, wer: "sitz-1",
			angriffe: []domain.Angriff{{
				Kreatur: "sitz-1/riese/0", Ziel: "sitz-1",
			}},
			wort: "kein zulässiges Angriffsziel",
		},

		"eine gegnerische Kreatur als Ziel": {
			bauen: imAngreiferschritt, wer: "sitz-1",
			angriffe: []domain.Angriff{{
				Kreatur: "sitz-1/riese/0", Ziel: "sitz-2/baer/0",
			}},
			wort: "kein zulässiges Angriffsziel",
		},

		"die eigene Kreatur als Ziel": {
			bauen: mitZweiterEigenerKreatur, wer: "sitz-1",
			angriffe: []domain.Angriff{{
				Kreatur: "sitz-1/riese/0", Ziel: "sitz-1/zweiter/0",
			}},
			wort: "kein zulässiges Angriffsziel",
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, fehler := fall.bauen(t).
				AngreiferErklaeren(fall.wer, fall.angriffe)

			if fehler == nil {
				t.Fatal("es kam kein Fehler")
			}

			if !strings.Contains(fehler.Error(), fall.wort) {
				t.Errorf("der Fehler lautet %q", fehler)
			}
		})
	}
}

// **Jeder Weg, auf dem kein Blocker erklärt wird, sagt warum.**
func TestJederFehlerwegDerBlockererklaerungSagtWarum(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		bauen   func(t *testing.T) domain.Partie
		wer     string
		bloecke []domain.Block
		wort    string
	}{
		"der falsche Schritt": {
			bauen: vorDemKampf, wer: "sitz-2", wort: "Blockerschritt",
		},

		"der aktive Spieler": {
			bauen: mitAngriff, wer: "sitz-1", wort: "CR 506.2",
		},

		"zweimal erklärt": {
			bauen: schonGeblockt, wer: "sitz-2", wort: "schon erklärt",
		},

		"eine Kreatur, die es nicht gibt": {
			bauen: mitAngriff, wer: "sitz-2",
			bloecke: []domain.Block{{
				Kreatur: "gibtsnicht", Angreifer: "sitz-1/riese/0",
			}},
			wort: "gibt es in dieser Partie nicht",
		},

		"ein Angreifer, der nicht angreift": {
			bauen: mitAngriff, wer: "sitz-2",
			bloecke: []domain.Block{{
				Kreatur: "sitz-2/baer/0", Angreifer: "sitz-2/baer/0",
			}},
			wort: "greift gar nicht an",
		},

		"ein getappter Blocker": {
			bauen: mitGetapptemBlocker, wer: "sitz-2",
			bloecke: []domain.Block{{
				Kreatur: "sitz-2/baer/0", Angreifer: "sitz-1/riese/0",
			}},
			wort: "CR 509.1a",
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, fehler := fall.bauen(t).BlockerErklaeren(fall.wer, fall.bloecke)

			if fehler == nil {
				t.Fatal("es kam kein Fehler")
			}

			if !strings.Contains(fehler.Error(), fall.wort) {
				t.Errorf("der Fehler lautet %q", fehler)
			}
		})
	}
}

// vorDemKampf gibt eine Partie im Versorgungsschritt — vor jedem Kampf.
func vorDemKampf(t *testing.T) domain.Partie {
	t.Helper()

	return bisZumSchritt(t, einePartie().ZugBeginnen("sitz-1"),
		domain.SchrittVersorgung)
}

// imAngreiferschritt gibt den Kampftisch, wie er ist.
func imAngreiferschritt(t *testing.T) domain.Partie {
	t.Helper()

	return einKampftisch(t)
}

// schonErklaert gibt eine Partie mit bereits erklärten Angreifern.
func schonErklaert(t *testing.T) domain.Partie {
	t.Helper()

	partie, fehler := einKampftisch(t).AngreiferErklaeren("sitz-1", nil)
	if fehler != nil {
		t.Fatalf("erste Erklärung: %v", fehler)
	}

	return partie
}

// schonGeblockt gibt eine Partie mit bereits erklärten Blockern.
func schonGeblockt(t *testing.T) domain.Partie {
	t.Helper()

	partie, fehler := mitAngriff(t).BlockerErklaeren("sitz-2", nil)
	if fehler != nil {
		t.Fatalf("erste Erklärung: %v", fehler)
	}

	return partie
}

// mitLandImKampf legt ein Land auf das eigene Schlachtfeld.
func mitLandImKampf(t *testing.T) domain.Partie {
	t.Helper()

	partie := einKampftisch(t)

	sitze := make([]domain.Sitz, len(partie.Sitze))
	copy(sitze, partie.Sitze)
	sitze[0].Deck.Karten = append(sitze[0].Deck.Karten, domain.Abzugskarte{
		Kennung: "wald", Name: "Forest", Menge: 1,
		Seiten: []domain.Abzugsseite{{Felder: map[string]string{
			"super type": "Basic Land", "sub type": "Forest",
		}}},
	})
	partie.Sitze = sitze

	partie.Objekte = append(partie.Objekte, domain.Spielobjekt{
		Kennung: "sitz-1/wald/0", Art: domain.ObjektKarte,
		Beherrscher: "sitz-1", Eigner: "sitz-1",
		Zone: domain.ZoneTisch, Karte: "wald",
	})

	return partie
}

// mitZweiterEigenerKreatur legt eine zweite eigene Kreatur hin.
func mitZweiterEigenerKreatur(t *testing.T) domain.Partie {
	t.Helper()

	partie := einKampftisch(t)

	partie.Objekte = append(partie.Objekte, domain.Spielobjekt{
		Kennung: "sitz-1/zweiter/0", Art: domain.ObjektKarte,
		Beherrscher: "sitz-1", Eigner: "sitz-1",
		Zone: domain.ZoneTisch, Karte: "kleiner",
	})

	return partie
}

// mitGetapptemBlocker tappt den Bären.
func mitGetapptemBlocker(t *testing.T) domain.Partie {
	t.Helper()

	partie := mitAngriff(t)

	objekte := make([]domain.Spielobjekt, len(partie.Objekte))
	copy(objekte, partie.Objekte)

	for stelle := range objekte {
		if objekte[stelle].Kennung == "sitz-2/baer/0" {
			objekte[stelle].Getappt = true
		}
	}

	partie.Objekte = objekte

	return partie
}

// **Ein Planeswalker ist ein zulässiges Angriffsziel** — CR 508.1a.
//
// **Der Zweig, den dieser Lauf betritt**: die Typprüfung in
// `verteidigerSteht`. Die Läufe daneben treffen sie nur auf ihrem
// **Nein**-Weg — und ein Weg, der jedes fremde Permanent abwiese, käme
// genauso durch (§10).
func TestEinPlaneswalkerIstEinZulaessigesZiel(t *testing.T) {
	t.Parallel()

	partie := einKampftisch(t)

	sitze := make([]domain.Sitz, len(partie.Sitze))
	copy(sitze, partie.Sitze)
	sitze[1].Deck.Karten = append(sitze[1].Deck.Karten, domain.Abzugskarte{
		Kennung: "walker", Name: "Wanderer", Menge: 1,
		Seiten: []domain.Abzugsseite{{Felder: map[string]string{
			"super type": "Legendary Planeswalker", "sub type": "Jace",
			"loyalty": "4",
		}}},
	})
	partie.Sitze = sitze

	partie.Objekte = append(partie.Objekte, domain.Spielobjekt{
		Kennung: "sitz-2/walker/0", Art: domain.ObjektKarte,
		Beherrscher: "sitz-2", Eigner: "sitz-2",
		Zone: domain.ZoneTisch, Karte: "walker",
	})

	if _, fehler := partie.AngreiferErklaeren("sitz-1",
		[]domain.Angriff{{
			Kreatur: "sitz-1/riese/0", Ziel: "sitz-2/walker/0",
		}}); fehler != nil {
		t.Errorf("ein Planeswalker wurde abgewiesen: %v", fehler)
	}
}

// **Ein Objekt außerhalb des Schlachtfelds ist kein Ziel** — CR 508.1a.
func TestEinObjektAusserhalbDesTischesIstKeinZiel(t *testing.T) {
	t.Parallel()

	partie := einKampftisch(t)

	objekte := make([]domain.Spielobjekt, len(partie.Objekte))
	copy(objekte, partie.Objekte)

	for stelle := range objekte {
		if objekte[stelle].Kennung == "sitz-2/baer/0" {
			objekte[stelle].Zone = domain.ZoneHand
		}
	}

	partie.Objekte = objekte

	if _, fehler := partie.AngreiferErklaeren("sitz-1",
		[]domain.Angriff{{
			Kreatur: "sitz-1/riese/0", Ziel: "sitz-2/baer/0",
		}}); fehler == nil {
		t.Error("eine Karte auf der Hand wurde angegriffen")
	}
}

// **Ohne Regeltext hat niemand Erstschlag** — der leere Weg von `regeltext`.
func TestOhneRegeltextGibtEsKeinenErstschlag(t *testing.T) {
	t.Parallel()

	// **Ein Spielstein ohne Karte** (CR 111): er hat keinen Abzug und damit
	// keinen Text — und trotzdem darf die Frage nach Erstschlag nicht
	// umfallen (§8).
	partie := mitAngriff(t)

	partie.Objekte = append(partie.Objekte, domain.Spielobjekt{
		Kennung: "sitz-2/stein/0", Art: domain.ObjektSpielstein,
		Beherrscher: "sitz-2", Eigner: "sitz-2",
		Zone: domain.ZoneTisch,
	})

	danach, fehler := partie.BlockerErklaeren("sitz-2", nil)
	if fehler != nil {
		t.Fatalf("blocker erklären: %v", fehler)
	}

	if danach.Zug.Kampf.Schadensschritte != 0 {
		t.Errorf("gezählt sind %d Schadensschritte",
			danach.Zug.Kampf.Schadensschritte)
	}
}
