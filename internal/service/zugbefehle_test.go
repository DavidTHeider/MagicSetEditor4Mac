package service_test

// **Die fünf Befehle, die zum Zug gehören** — Werkstück G9 von Phase 5.5
// (#1592).
//
// # Wogegen hier geprüft wird
//
//   - **Passen gibt weiter, und alle nacheinander beenden den Abschnitt**
//     (CR 117.3d, CR 117.4).
//   - **Wer nicht an der Reihe ist, darf nichts stellen, was Priorität
//     braucht** (CR 117.1) — und die Buchführung geht trotzdem.
//   - **Leben und Vergiftungsmarken ändern sich um einen Betrag**, nicht auf
//     einen Wert: „drei weniger" überlebt eine gekreuzte Antwort (§7,
//     Regel 2).
//   - **Markierter Schaden liegt am Objekt** (CR 120.3).

import (
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Passen gibt die Priorität weiter.**
func TestPassenGibtDiePrioritaetWeiter(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})

	ereignisse, danach, fehler := kasten.Auslegen(eineZugpartie(t),
		domain.Spielbefehl{
			PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
			Art: service.BefehlPassen,
		})
	if fehler != nil {
		t.Fatalf("passen: %v", fehler)
	}

	switch {
	case danach.Zug.Prioritaet != "sitz-2":
		t.Errorf("Priorität hat %q", danach.Zug.Prioritaet)
	case len(ereignisse) != 1:
		t.Fatalf("%d Ereignisse", len(ereignisse))
	case ereignisse[0].Oeffentlich["prioritaet"] != "sitz-2":
		t.Errorf("das Ereignis nennt %v", ereignisse[0].Oeffentlich)
	}
}

// **Wer nicht an der Reihe ist, passt nicht** — CR 117.3d.
func TestWerNichtDranIstPasstNichtUeberDenSandkasten(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})

	_, _, fehler := kasten.Auslegen(eineZugpartie(t), domain.Spielbefehl{
		PartieID: "p1", Absender: "sitz-2", Kennung: "b1",
		Art: service.BefehlPassen,
	})

	if fehler == nil {
		t.Error("ein Passen außer der Reihe kam durch")
	}
}

// **Wer nicht an der Reihe ist, darf nur die Buchführung.**
//
// **Das ist die Abnahme des Issues**: „der andere sieht, dass er nicht am
// Zug ist, und kann nichts stellen, was Priorität braucht."
func TestOhnePrioritaetBleibtNurDieBuchfuehrung(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})
	partie := eineZugpartie(t)

	erlaubt := kasten.Erlaubt(partie, "sitz-2")

	// **Vierzehn und nicht zwanzig**: die Legendenwahl und die Reihung
	// beantworten eine Frage, und in dieser Partie steht keine
	// (`ohneUngestellteFragen`). **Und vom Beginn steht nichts mehr offen**
	// (#1839): er ist vorbei, bevor in dieser Partie gespielt wird (#1851).
	if len(erlaubt) != 14 {
		t.Fatalf("%d Befehle erlaubt: %v", len(erlaubt), erlaubt)
	}

	for _, art := range erlaubt {
		switch art {
		case service.BefehlAbwerfen, service.BefehlLeben,
			service.BefehlVergiftung, service.BefehlSchaden,
			// **Mana nachtragen ist Buchführung** (M3, #1605) — wie der
			// Lebensstand daneben. Das **Grundland** steht ausdrücklich
			// nicht hier: es tappt etwas am Tisch, und CR 605.3a verlangt
			// dafür Priorität.
			service.BefehlMana,

			// **Angreifen und Blocken brauchen keine Priorität** (M6,
			// #1608): CR 508.1 und CR 509.1 sind zugbasierte Handlungen mit
			// einer Wahl darin, und die Priorität kommt **danach**
			// (CR 508.2, CR 509.2). Wer darf, prüft die Domäne — der aktive
			// Spieler greift an, der Verteidiger blockt.
			service.BefehlAngreifer, service.BefehlBlocker,

			// **Ein Wurf ändert nichts am Tisch** (M15, #1617) — er
			// braucht deshalb keine Priorität. Was er bewirkt, steht auf
			// einer Karte, und die zu wirken braucht sie sehr wohl.
			service.BefehlMuenzwurf, service.BefehlWuerfeln,

			// **Die Legendenwahl (M12, #1614) und die Reihung (M9,
			// #1611) stehen hier nicht**, und das ist keine Auslassung:
			// beide brauchen keine Priorität — aber sie beantworten eine
			// **offene Frage**, und in dieser Partie steht keine. Sie
			// fallen deshalb schon vorher heraus
			// (`ohneUngestellteFragen`); dass sie ohne Priorität gehen,
			// misst `TestEineAltePartieBleibtEinSandkasten` nebenan.
			//
			// **Das Heimholen** (M19, #1620): CR 903.9a hängt an
			// einer zustandsbasierten Prüfung, und die läuft, bevor
			// jemand Priorität bekommt (CR 704.3).
			service.BefehlHeimholen,

			// **Und die Bezeichnungen** (M20, #1621): was sie vergibt,
			// sind eingebaute ausgelöste Fähigkeiten ohne Quelle
			// (CR 725.2, CR 726.2, CR 728.1).
			service.BefehlMonarch, service.BefehlInitiative,
			service.BefehlStrahlung,

			service.BefehlAufgeben:
		default:
			t.Errorf("%q braucht Priorität und steht trotzdem da", art)
		}
	}

	// **Und wer an der Reihe ist, darf alles** — bis auf die zwei Antworten
	// und die vier Schritte des Beginns, der vorbei ist (fünfundvierzig seit
	// der Geste aus P3, #2371).
	if alle := kasten.Erlaubt(partie, "sitz-1"); len(alle) != 45 {
		t.Errorf("dem aktiven Sitz sind %d Befehle erlaubt", len(alle))
	}
}

// **Eine alte Partie bleibt ein Sandkasten.**
//
// **Ihr Profil ist eingefroren** (G4, #1461) und führt `prioritaet` als
// `manuell` — sie soll weiterlaufen, wie sie angefangen hat (§15, Punkt 4).
func TestEineAltePartieBleibtEinSandkasten(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})
	partie := eineZugpartie(t)

	for stelle := range partie.Profil.Matrix.Eintraege {
		if partie.Profil.Matrix.Eintraege[stelle].Bereich ==
			domain.BereichPrioritaet {
			partie.Profil.Matrix.Eintraege[stelle].Grad = domain.UnterstuetzungManuell
		}
	}

	// **Dieselben fünfundvierzig wie für den aktiven Sitz nebenan** — auch
	// für den, der nicht an der Reihe ist (fünfundvierzig seit der Geste aus
	// P3, #2371).
	if alle := kasten.Erlaubt(partie, "sitz-2"); len(alle) != 45 {
		t.Errorf("im Sandkasten sind %d Befehle erlaubt", len(alle))
	}
}

// **Der Abwurf im Aufräumschritt geht über denselben Weg** — CR 514.1.
//
// # Warum der Zug dafür bis zum Aufräumen gefahren wird
//
// **Weil die Pflicht dort entsteht.** Sie zu setzen, indem man sie in den
// Zustand schreibt, prüfte die Vorrichtung statt des Wegs (§3) — hier passen
// beide Sitze durch alle Abschnitte, bis der Aufräumschritt sie stellt.
func TestDerAbwurfGehtUeberDenSandkasten(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})
	partie := eineZugpartie(t)

	// **Acht Karten auf der Hand** — eine über der Grenze von sieben.
	//
	// **Sieben bringt die Starthand mit, eine kommt dazu** (#1851): bis
	// dahin legte der Lauf alle acht dazu, weil das Probedeck ohne Beginn
	// spielte und keine Hand hatte.
	if hand := handDes(partie, "sitz-1"); len(hand) != domain.Starthand {
		t.Fatalf("die Starthand hat %d Karten", len(hand))
	}

	partie.Objekte = append(partie.Objekte, domain.Spielobjekt{
		Kennung:     "sitz-1/hand/a",
		Art:         domain.ObjektKarte,
		Beherrscher: "sitz-1", Eigner: "sitz-1",
		Zone: domain.ZoneHand, Karte: "wald",
	})

	partie = bisZumAufraeumen(t, kasten, partie)

	if partie.Zug.Abwurf != 1 {
		t.Fatalf("abzuwerfen sind %d", partie.Zug.Abwurf)
	}

	auf := ersteAufDerHand(partie, "sitz-1")

	_, danach, fehler := kasten.Auslegen(partie, domain.Spielbefehl{
		PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
		Art: service.BefehlAbwerfen, Nutzlast: map[string]any{"objekt": auf},
	})
	if fehler != nil {
		t.Fatalf("abwerfen: %v", fehler)
	}

	switch {
	case danach.Zug.Abwurf != 0:
		t.Errorf("es sind noch %d offen", danach.Zug.Abwurf)
	case danach.Zug.Nummer != 2:
		t.Errorf("der Zug ging nicht weiter: %d", danach.Zug.Nummer)
	}
}

// bisZumAufraeumen passt durch, bis der Aufräumschritt dasteht.
func bisZumAufraeumen(
	t *testing.T, kasten *service.Sandkasten, partie domain.Partie,
) domain.Partie {
	t.Helper()

	for range 40 {
		if partie.Zug.Abschnitt.Schritt == domain.SchrittAufraeumen {
			return partie
		}

		_, danach, fehler := kasten.Auslegen(partie, domain.Spielbefehl{
			PartieID: "p1", Absender: partie.Zug.Prioritaet, Kennung: "b0",
			Art: service.BefehlPassen,
		})
		if fehler != nil {
			t.Fatalf("passen bei %+v: %v", partie.Zug.Abschnitt, fehler)
		}

		partie = danach
	}

	t.Fatal("der Aufräumschritt kam nicht")

	return partie
}

// ersteAufDerHand gibt die erste Karte auf der Hand dieses Sitzes.
func ersteAufDerHand(partie domain.Partie, sitz string) string {
	for _, objekt := range partie.Objekte {
		if objekt.Beherrscher == sitz && objekt.Zone == domain.ZoneHand {
			return objekt.Kennung
		}
	}

	return ""
}

// **Eine Abkürzung reicht durch bis zum Ende des Zuges** — CR 732 (M1,
// #1603).
//
// **Über den Befehl und nicht über die Domäne unmittelbar**: die Abnahme des
// Issues lautet „eine Abkürzung führt durch mehrere Abschnitte und hält an,
// sobald etwas geschieht" — das ist eine Aussage über den **Weg**.
func TestEineAbkuerzungGehtDurchDenBefehl(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})
	partie := eineZugpartie(t)
	vorher := partie.Zug.Nummer

	ereignisse, danach, fehler := kasten.Auslegen(partie, domain.Spielbefehl{
		PartieID: "p1", Absender: partie.Zug.Prioritaet, Kennung: "b1",
		Art: service.BefehlAbkuerzen,
	})
	if fehler != nil {
		t.Fatalf("abkürzen: %v", fehler)
	}

	switch {
	case danach.Zug.Nummer != vorher+1:
		t.Errorf("der Zug steht bei %d statt %d", danach.Zug.Nummer, vorher+1)
	case len(ereignisse) != 1:
		t.Errorf("%d Ereignisse statt einem", len(ereignisse))
	case ereignisse[0].Art != service.BefehlAbkuerzen:
		t.Errorf("das Ereignis heißt %q", ereignisse[0].Art)
	}
}

// **Wer nicht an der Reihe ist, kürzt nicht ab** — CR 732.2a.
func TestOhnePrioritaetWirdNichtAbgekuerzt(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})
	partie := eineZugpartie(t)

	_, _, fehler := kasten.Auslegen(partie, domain.Spielbefehl{
		PartieID: "p1", Absender: "sitz-2", Kennung: "b1",
		Art: service.BefehlAbkuerzen,
	})

	if fehler == nil {
		t.Fatal("der Sitz ohne Priorität hat abgekürzt")
	}
}

// **Ein Zug lässt sich vorzeitig beenden** — CR 724.1 (M1, #1603).
func TestEinZugLaesstSichVorzeitigBeenden(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})
	partie := eineZugpartie(t)
	vorher := partie.Zug.Nummer

	ereignisse, danach, fehler := kasten.Auslegen(partie, domain.Spielbefehl{
		PartieID: "p1", Absender: partie.Zug.Prioritaet, Kennung: "b1",
		Art: service.BefehlZugende,
	})
	if fehler != nil {
		t.Fatalf("zugende: %v", fehler)
	}

	switch {
	case danach.Zug.Nummer != vorher+1:
		t.Errorf("der Zug steht bei %d statt %d", danach.Zug.Nummer, vorher+1)
	case len(ereignisse) != 1 || ereignisse[0].Art != service.BefehlZugende:
		t.Errorf("die Ereignisse: %+v", ereignisse)
	}
}

// **Eine Antwort steht da, sobald die Frage steht** — CR 704.5j und
// CR 603.3b.
//
// # Die Gegenprobe zu `ohneUngestellteFragen`
//
// **Ohne sie sagte der Lauf nebenan nur, dass beide fehlen**, und das wäre
// mit „nie erlaubt" genauso wahr. Hier steht, dass sie **wiederkommen** —
// und zwar bei dem, der gefragt ist, und nur bei ihm.
func TestEineAntwortStehtDaSobaldDieFrageSteht(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})
	partie := eineZugpartie(t)

	partie.Ausloeserreihung = "sitz-2"
	partie.Legendenwahlen = []domain.Legendenwahl{
		{Sitz: "sitz-1", Name: "Held"},
	}

	gefragt := map[string][]string{
		"sitz-1": {service.BefehlLegende},
		"sitz-2": {service.BefehlReihung},
	}

	for wer, erwartet := range gefragt {
		erlaubt := kasten.Erlaubt(partie, wer)

		for _, art := range []string{service.BefehlLegende, service.BefehlReihung} {
			soll := slices.Contains(erwartet, art)

			if slices.Contains(erlaubt, art) != soll {
				t.Errorf("%q darf %q: %v, erwartet %v",
					wer, art, !soll, soll)
			}
		}
	}
}
