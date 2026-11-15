package domain

// **Abkürzungen** — Werkstück M1 der Regelmaschine
// (#1603).
//
// # Die Zahl, um die es geht
//
// **Zehn der zwölf Abschnitte vergeben Priorität** (alle außer Enttappen,
// CR 502.4, und Aufräumen, CR 514.3). Bei zwei Sitzen sind das **zwanzig
// Passen je Zug**, in denen nichts geschieht — und genau so steht es seit G9
// als Grenze in der Regelmatrix. CR 732.1 sagt, dass am Tisch niemand so
// spielt.
//
// # Was die CR beschreibt und was hier daraus wird
//
// **CR 732.2a**: wer Priorität hat, schlägt eine Folge von Entscheidungen
// **für alle** vor, die sich aus dem Stand vorhersagen lässt. **CR 732.2b**:
// jeder andere darf sie abkürzen, indem er eine Stelle benennt, an der er
// anders entscheidet. **CR 732.2c**: danach springt das Spiel ans Ende.
//
// **Hier wird der Vorschlag ausgeführt und nicht verhandelt.** Das ist die
// Grenze, und sie gehört ausgesprochen (§10): am gemeinsamen Tisch verabreden
// sich zwei Menschen mit einem Satz, und genau das ist dieser Befehl. **Über
// das Netz** (G5) kann das Gegenüber nicht dazwischengehen — was geschehen
// ist, steht im Ereignisstrom, und die Rücknahme gibt es. Wer 732.2b wirklich
// braucht, baut eine Verhandlung; dieses Werkstück baut sie nicht.
//
// # Wo sie anhält
//
// **An allem, was keine Wahl mehr vorhersagen lässt**: etwas liegt auf dem
// Stapel, jemand muss abwerfen, die Partie ist zu Ende — oder der Zug ist
// vorbei. Der Zug ist die natürliche Grenze: was im nächsten geschieht, hat
// niemand vorgeschlagen.

import "fmt"

// Abkuerzen reicht durch, bis etwas geschieht — CR 732 (Fassung 2026-08-19).
//
// **Der Ruf gehört dem, der Priorität hat** (CR 732.2a); alle passen, und
// zwar so lange, bis der Zug vorbei ist oder etwas dazwischenkommt.
func (p Partie) Abkuerzen(wer string) (Partie, error) {
	if p.Stand == StandBeendet {
		return p, fmt.Errorf(
			"%w: die Partie ist zu Ende — es gibt nichts abzukürzen",
			ErrValidation)
	}

	if p.Zug.Prioritaet == "" || wer != p.Zug.Prioritaet {
		return p, fmt.Errorf(
			"%w: %q schlägt keine Abkürzung vor — die Priorität hat %q "+
				"(CR 732.2a)", ErrValidation, wer, p.Zug.Prioritaet)
	}

	bisZug := p.Zug.Nummer

	// **Die Schleife endet, und zwar ohne Obergrenze.** Jeder vollständige
	// Durchgang lässt alle Sitze passen; [Partie.Passen] schaltet dann einen
	// Abschnitt weiter, und nach zwölf steigt die Zugnummer — dann hält
	// [Partie.haeltAn]. Eine Zahl daneben wäre ein Zweig, den kein Lauf
	// betreten kann, und das ist ein Versprechen, das niemand einlöst (§10).
	//
	// **Und `Passen` kann hier nicht scheitern**: es weist aus zwei Gründen
	// ab — eine leere Priorität und ein fremder Sitz —, und `haeltAn`
	// schließt den ersten aus, während der zweite nicht entstehen kann, weil
	// hier die Priorität selbst eingesetzt wird. Ein Fehlerweg dafür wäre
	// derselbe unbetretene Zweig.
	for !p.haeltAn(bisZug) {
		p, _ = p.Passen(p.Zug.Prioritaet) //nolint:errcheck // siehe oben
	}

	return p, nil
}

// haeltAn sagt, ob die Abkürzung hier aufhört.
//
// **Vier Gründe, und jeder ist eine Wahl oder ein Ende**: der Zug ist vorbei,
// die Partie auch, etwas liegt auf dem Stapel (CR 405.5), oder jemand muss
// abwerfen (CR 514.1). Was danach käme, hat niemand vorgeschlagen — und
// CR 732.2a verlangt, dass eine Abkürzung vorhersehbar ist.
func (p Partie) haeltAn(bisZug int) bool {
	return p.Zug.Nummer != bisZug ||
		p.Stand == StandBeendet ||
		p.StapelOben() != nil ||
		p.Zug.Abwurf > 0 ||
		p.Zug.Prioritaet == ""
}

// ZugBeenden beendet den Zug vorzeitig — CR 724.1 (Fassung 2026-08-19).
//
// # Die vier Schritte, und welche hier geschehen
//
//   - **724.1a**: wartende Auslöser hören auf zu existieren. Auslöser gibt es
//     nicht (`M9`), also gibt es auch keine wartenden.
//   - **724.1b**: alles auf dem Stapel wird **exiliert** — auch das, was
//     gerade verrechnet. Das geschieht.
//   - **724.1c**: die zustandsbasierten Aktionen laufen, und **niemand
//     bekommt Priorität**. Das geschieht.
//   - **724.1d**: der laufende Abschnitt endet, und das Spiel springt
//     **geradewegs in den Aufräumschritt** — alles dazwischen fällt aus.
//
// # Warum es einen Befehl dafür gibt und keine Karte
//
// **CR 724.1 fängt mit „manche Karten beenden den Zug" an**, und Karten, die
// etwas tun, gibt es nicht (`M8`). Was hier steht, ist derselbe Weg mit einem
// Menschen als Auslöser — wie `spieler.leben` und `objekt.schaden`: im
// Sandkasten sagt es jemand. Ohne diesen Weg wäre CR 724 eine Zusage ohne
// Einlösung (§10).
//
// # Was nicht geschieht
//
// **CR 724.1d, zweiter Teil**: endet der Zug **im** Aufräumschritt, beginnt
// ein neuer. Das braucht einen Effekt, der dort ausgelöst wird (`M10`). Und
// **CR 724.2** (die Kampfphase beenden) braucht den Kampf (`M6`).
func (p Partie) ZugBeenden() Partie {
	// **Ohne laufenden Zug gibt es nichts zu beenden** — eine Partie aus G1
	// bis G8 ist ein Sandkasten, und in ihm geht es nicht zugweise zu.
	if p.Zug.Nummer == 0 {
		return p
	}

	p = p.stapelExilieren()

	// **CR 724.1c** — und ausdrücklich, bevor der Aufräumschritt betreten
	// wird: dort laufen sie ohnehin noch einmal, aber die Regel nennt sie
	// hier.
	p = p.Zustandspruefung()

	// **CR 724.1d** — geradewegs in den Aufräumschritt. `abschnittBetreten`
	// erledigt den Rest: Schaden weg (CR 514.2), Abwurfpflicht feststellen
	// (CR 514.1), und weiter, sobald sie erfüllt ist (CR 514.3).
	p.Zug.Abschnitt = Abschnitt{Phase: PhaseEnde, Schritt: SchrittAufraeumen}
	p.Zug.Gepasst = nil

	return p.abschnittBetreten()
}

// stapelExilieren schickt alles vom Stapel ins Exil — CR 724.1b.
//
// **Auch das, was gerade verrechnet**, sagt die Regel ausdrücklich. Was keine
// Karte ist, hört danach bei der nächsten zustandsbasierten Prüfung auf zu
// existieren (CR 704.5d); Spielsteine auf dem Stapel gibt es hier nicht.
func (p Partie) stapelExilieren() Partie {
	var geaendert []Spielobjekt

	for stelle, objekt := range p.Objekte {
		if objekt.Zone != ZoneStapel {
			continue
		}

		if geaendert == nil {
			geaendert = make([]Spielobjekt, len(p.Objekte))
			copy(geaendert, p.Objekte)
		}

		// **CR 400.7** (M18, #1619): was ins Exil geht, ist ein neues
		// Objekt — die Ziele des Zaubers gelten nicht mehr.
		geaendert[stelle] = NeuesObjekt(geaendert[stelle], p.Zug.Nummer)
		geaendert[stelle].Zone = ZoneExil
		geaendert[stelle].Beherrscher = geaendert[stelle].Eigner
	}

	if geaendert == nil {
		return p
	}

	p.Objekte = geaendert

	return p
}
