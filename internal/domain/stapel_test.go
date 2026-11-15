package domain_test

// **Der Stapel** — Werkstück M1 der Regelmaschine
// (#1603).
//
// # Wogegen hier geprüft wird
//
//   - CR 405.2 — was zuletzt daraufkommt, liegt oben.
//   - CR 405.5 und CR 608.1 — passen alle nacheinander, verrechnet das
//     **oberste**; ist der Stapel leer, endet der Abschnitt.
//   - CR 608.2n — ein Spontan- oder Hexereizauber geht danach in den
//     Friedhof seines **Eigners**.
//   - CR 608.3a — ein bleibendes Objekt kommt auf das Schlachtfeld, unter
//     die Kontrolle dessen, der es dorthin gebracht hat.
//   - CR 117.3b — danach bekommt der **aktive** Spieler Priorität, nicht
//     der, der gewirkt hat.
//   - CR 500.2 — ein Abschnitt endet, wenn der Stapel **leer** ist und alle
//     nacheinander passen. G9 hatte nur die zweite Hälfte.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// eineStapelpartie gibt zwei Sitze, einen laufenden Zug und zwei Karten im
// Abzug des zweiten Sitzes.
func eineStapelpartie(t *testing.T) domain.Partie {
	t.Helper()

	partie := eineKartenpartie()
	partie.Sitze[0].Deck.Karten = append(partie.Sitze[0].Deck.Karten,
		domain.Abzugskarte{
			Kennung: "k2", Name: "Blitz", Menge: 1,
			Seiten: []domain.Abzugsseite{{
				Name: "Blitz",
				Felder: map[string]string{
					"name": "Blitz", "casting cost": "{R}",
					"super type": "Instant",
				},
			}},
		})

	partie.Objekte = []domain.Spielobjekt{
		{
			Kennung: "bär", Art: domain.ObjektKarte, Karte: "k1",
			Beherrscher: "sitz-1", Eigner: "sitz-1",
			Zone: domain.ZoneStapel, Ordnung: 2,
		},
		{
			Kennung: "blitz", Art: domain.ObjektKarte, Karte: "k2",
			Beherrscher: "sitz-2", Eigner: "sitz-1",
			Zone: domain.ZoneStapel, Ordnung: 1,
		},
	}

	return partie.ZugBeginnen("sitz-1")
}

// **Was oben liegt, verrechnet zuerst** — CR 405.2 mit CR 608.1.
func TestDasObersteVerrechnetZuerst(t *testing.T) {
	t.Parallel()

	partie := eineStapelpartie(t)

	oben := partie.StapelOben()
	if oben == nil || oben.Kennung != "blitz" {
		t.Fatalf("oben liegt %+v", oben)
	}
}

// **Ein leerer Stapel hat kein Oberstes.**
func TestEinLeererStapelHatKeinOberstes(t *testing.T) {
	t.Parallel()

	partie := eineStapelpartie(t)
	partie.Objekte = nil

	if oben := partie.StapelOben(); oben != nil {
		t.Fatalf("auf dem leeren Stapel liegt %+v", oben)
	}
}

// **Ein Spontanzauber geht in den Friedhof seines Eigners** — CR 608.2n.
//
// **Des Eigners und nicht des Beherrschers**: der Blitz wird von `sitz-2`
// beherrscht und gehört `sitz-1`. Ein Kontrollwechsel gibt kein Eigentum.
func TestEinSpontanzauberGehtInDenFriedhofSeinesEigners(t *testing.T) {
	t.Parallel()

	danach := eineStapelpartie(t).Verrechnen()

	blitz := objektMit(t, danach, "blitz")

	switch {
	case blitz.Zone != domain.ZoneFriedhof:
		t.Errorf("der Blitz liegt in %q", blitz.Zone)
	case blitz.Beherrscher != "sitz-1":
		t.Errorf("den Blitz beherrscht %q — im Friedhof ist es sein Eigner",
			blitz.Beherrscher)
	}
}

// **Ein bleibendes Objekt kommt auf das Schlachtfeld** — CR 608.3a, und zwar
// unter die Kontrolle dessen, der es dorthin gebracht hat.
func TestEinBleibendesObjektKommtAufDenTisch(t *testing.T) {
	t.Parallel()

	partie := eineStapelpartie(t)
	partie.Objekte = partie.Objekte[:1]

	danach := partie.Verrechnen()

	baer := objektMit(t, danach, "bär")

	switch {
	case baer.Zone != domain.ZoneTisch:
		t.Errorf("der Bär liegt in %q", baer.Zone)
	case baer.Beherrscher != "sitz-1":
		t.Errorf("den Bären beherrscht %q", baer.Beherrscher)
	}
}

// **Ein leerer Stapel verrechnet nichts** — und gibt dieselbe Partie.
func TestEinLeererStapelVerrechnetNichts(t *testing.T) {
	t.Parallel()

	partie := eineStapelpartie(t)
	partie.Objekte = nil

	if danach := partie.Verrechnen(); len(danach.Objekte) != 0 {
		t.Fatalf("aus nichts wurden %d Objekte", len(danach.Objekte))
	}
}

// **Passen alle, verrechnet das Oberste — und der Abschnitt bleibt stehen**
// (CR 405.5, CR 500.2).
func TestBeiVollemStapelWirdVerrechnetStattWeiterzugehen(t *testing.T) {
	t.Parallel()

	partie := eineStapelpartie(t)
	vorher := partie.Zug.Abschnitt

	danach := passenBisAlle(t, partie)

	switch {
	case danach.Zug.Abschnitt != vorher:
		t.Errorf("der Abschnitt ist %+v statt %+v", danach.Zug.Abschnitt, vorher)
	case objektMit(t, danach, "blitz").Zone != domain.ZoneFriedhof:
		t.Error("der Blitz liegt noch auf dem Stapel")
	case len(danach.Zug.Gepasst) != 0:
		t.Errorf("gepasst haben noch %v", danach.Zug.Gepasst)
	}
}

// **Nach dem Verrechnen bekommt der aktive Spieler Priorität** — CR 117.3b,
// und nicht der, der zuletzt gepasst hat.
func TestNachDemVerrechnenHatDerAktiveSpielerPrioritaet(t *testing.T) {
	t.Parallel()

	danach := passenBisAlle(t, eineStapelpartie(t))

	if danach.Zug.Prioritaet != danach.Zug.Aktiv {
		t.Fatalf("die Priorität hat %q, aktiv ist %q",
			danach.Zug.Prioritaet, danach.Zug.Aktiv)
	}
}

// **Ist der Stapel leer, endet der Abschnitt** — CR 500.2, wie seit G9.
func TestBeiLeeremStapelEndetDerAbschnitt(t *testing.T) {
	t.Parallel()

	partie := eineStapelpartie(t)
	partie.Objekte = nil
	vorher := partie.Zug.Abschnitt

	if danach := passenBisAlle(t, partie); danach.Zug.Abschnitt == vorher {
		t.Fatalf("der Abschnitt steht noch bei %+v", vorher)
	}
}

// passenBisAlle lässt jeden Sitz einmal passen.
func passenBisAlle(t *testing.T, partie domain.Partie) domain.Partie {
	t.Helper()

	for range partie.Sitze {
		danach, fehler := partie.Passen(partie.Zug.Prioritaet)
		if fehler != nil {
			t.Fatalf("passen: %v", fehler)
		}

		partie = danach
	}

	return partie
}

// objektMit gibt das Objekt mit dieser Kennung.
func objektMit(
	t *testing.T, partie domain.Partie, kennung string,
) domain.Spielobjekt {
	t.Helper()

	for _, objekt := range partie.Objekte {
		if objekt.Kennung == kennung {
			return objekt
		}
	}

	t.Fatalf("das Objekt %q gibt es nicht", kennung)

	return domain.Spielobjekt{}
}

// **Ein Zug endet vorzeitig** — CR 724.1 (Fassung 2026-08-19).
//
// **Alles auf dem Stapel wird exiliert** (724.1b), die zustandsbasierten
// Aktionen laufen (724.1c), und das Spiel springt in den **Aufräumschritt**
// (724.1d) — die Abschnitte dazwischen fallen aus.
func TestEinZugEndetVorzeitig(t *testing.T) {
	t.Parallel()

	partie := eineStapelpartie(t)
	vorher := partie.Zug.Nummer

	danach := partie.ZugBeenden()

	switch {
	case objektMit(t, danach, "blitz").Zone != domain.ZoneExil:
		t.Errorf("der Blitz liegt in %q",
			objektMit(t, danach, "blitz").Zone)
	case objektMit(t, danach, "bär").Zone != domain.ZoneExil:
		t.Errorf("der Bär liegt in %q", objektMit(t, danach, "bär").Zone)
	case danach.Zug.Nummer != vorher+1:
		t.Errorf("der Zug steht bei %d statt %d",
			danach.Zug.Nummer, vorher+1)
	}
}

// **Der Aufräumschritt läuft dabei wirklich** — CR 724.1d springt dorthin,
// und dort geht der markierte Schaden weg (CR 514.2).
func TestBeimVorzeitigenZugendeLaeuftDasAufraeumen(t *testing.T) {
	t.Parallel()

	partie := eineKartenpartie().ZugBeginnen("sitz-1")
	partie.Objekte = []domain.Spielobjekt{{
		Kennung: "wund", Art: domain.ObjektKarte, Karte: "k1",
		Beherrscher: "sitz-1", Eigner: "sitz-1",
		Zone: domain.ZoneTisch, Schaden: 3,
	}}

	if danach := partie.ZugBeenden(); objektMit(t, danach, "wund").Schaden != 0 {
		t.Fatalf("der Schaden steht bei %d",
			objektMit(t, danach, "wund").Schaden)
	}
}

// **Und die Abwurfpflicht hält ihn an** — CR 514.1 ist eine Wahl, und der
// Aufräumschritt wartet darauf.
func TestEinVorzeitigesZugendeHaeltVorDerAbwurfpflichtAn(t *testing.T) {
	t.Parallel()

	partie := eineKartenpartie().ZugBeginnen("sitz-1")
	partie.Objekte = handvoll("sitz-1", domain.Handgrenze+1)

	if danach := partie.ZugBeenden(); danach.Zug.Abwurf != 1 {
		t.Fatalf("abzuwerfen ist %d", danach.Zug.Abwurf)
	}
}

// **Eine Partie ohne laufenden Zug endet nicht vorzeitig** — sie ist ein
// Sandkasten, und in ihm gibt es keinen Zug, der enden könnte.
func TestOhneLaufendenZugEndetNichtsVorzeitig(t *testing.T) {
	t.Parallel()

	partie := eineKartenpartie()

	if danach := partie.ZugBeenden(); danach.Zug.Nummer != 0 {
		t.Fatalf("der Zug steht bei %d", danach.Zug.Nummer)
	}
}
