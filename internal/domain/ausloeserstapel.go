package domain

// **Wie eine ausgelöste Fähigkeit auf den Stapel kommt** — CR 603.3,
// Werkstück M9 (#1611).
//
// # Der Schnitt folgt der Frage
//
// **Nebenan steht, *was* auslöst und *was es tut*** — die vier eingebauten
// Fähigkeiten aus CR 725.2, CR 726.2 und CR 728.1. **Hier steht, wie aus
// einer wartenden Fähigkeit ein Objekt auf dem Stapel wird**, und das ist
// dieselbe Frage für jede von ihnen — und für jede, die `M8`
// (#1610)
// noch bringt.
//
// # Zwei Reihenfolgen, und nur eine gehört dem Kern
//
// **Zwischen Spielern entscheidet die Regel** (CR 603.3b mit CR 101.4): der
// aktive Spieler legt zuerst, dann die anderen in Zugreihenfolge.
//
// **Innerhalb eines Spielers entscheidet er** (CR 603.3b). Ein Kern, der die
// Auslösereihenfolge nähme, träfe eine Entscheidung, die ihm nicht gehört —
// dieselbe Gestalt wie der Abwurf (CR 514.1) und die Legendenwahl
// (CR 704.5j): eine **offene Frage im Zustand**, keine Rückfrage im Fenster
// (§16).

import "fmt"

// ausloeserAufDenStapel legt die wartenden Fähigkeiten auf den Stapel —
// CR 603.3.
//
// **In APNAP-Reihenfolge** (CR 603.3b mit CR 101.4): der aktive Spieler
// zuerst, dann die anderen reihum. Weil jede oben aufkommt (CR 405.2), wird
// die des aktiven Spielers **zuletzt** verrechnet — das ist die Regel und
// keine Umkehrung.
func (p Partie) ausloeserAufDenStapel() Partie {
	for {
		wer := p.legtAlsNaechstes()

		// **Was niemandem mehr gehört, fällt weg** — CR 800.4: ein Sitz,
		// der die Partie verlassen hat, legt nichts mehr auf den Stapel.
		if wer == "" {
			p.Ausgeloest = nil
			p.Ausloeserreihung = ""

			return p
		}

		wartend := p.wartendeVon(wer)

		// **Ab zwei entscheidet er selbst** — CR 603.3b. Bis er geantwortet
		// hat, geht hier nichts weiter, und niemand bekommt Priorität.
		if len(wartend) > 1 {
			p.Ausloeserreihung = wer

			return p
		}

		p = p.legeAusloeser(wartend[0])
	}
}

// legtAlsNaechstes gibt den ersten Sitz in APNAP-Reihenfolge, der noch etwas
// liegen hat — CR 603.3b mit CR 101.4.
func (p Partie) legtAlsNaechstes() string {
	for _, wer := range p.apnap() {
		if len(p.wartendeVon(wer)) > 0 {
			return wer
		}
	}

	return ""
}

// wartendeVon gibt die wartenden Fähigkeiten eines Sitzes — in der
// Reihenfolge, in der sie ausgelöst haben.
func (p Partie) wartendeVon(wer string) []Ausloeser {
	var seine []Ausloeser

	for _, einer := range p.Ausgeloest {
		if einer.Beherrscher == wer {
			seine = append(seine, einer)
		}
	}

	return seine
}

// legeAusloeser nimmt eine Fähigkeit aus der Warteschlange und legt sie auf
// den Stapel.
//
// **Nur die erste mit dieser Kennung**, und das ist kein Detail: eine Partie
// aus dem Bestand kann zwei Wartende ohne Kennung tragen — geschrieben von
// einer Fassung, die noch keine vergab (§8). Wer alle Gleichen nähme,
// verschluckte die zweite still.
func (p Partie) legeAusloeser(einer Ausloeser) Partie {
	uebrig := make([]Ausloeser, 0, len(p.Ausgeloest))
	genommen := false

	for _, andere := range p.Ausgeloest {
		if !genommen && andere.Kennung == einer.Kennung {
			genommen = true

			continue
		}

		uebrig = append(uebrig, andere)
	}

	p.Ausgeloest = uebrig

	return p.faehigkeitAufDenStapel(einer)
}

// Reihung beantwortet die Frage aus CR 603.3b — die genannte Fähigkeit kommt
// als Nächste auf den Stapel.
//
// **Als Nächste und nicht als Erste**: wer zuerst gelegt wird, liegt unten
// (CR 405.2) und wird **zuletzt** verrechnet. Das ist die Wahl, die CR 603.3b
// meint, und sie sieht am Tisch anders herum aus, als sie hier heißt.
func (p Partie) Reihung(wer, kennung string) (Partie, error) {
	if p.Ausloeserreihung == "" || p.Ausloeserreihung != wer {
		return p, fmt.Errorf(
			"%w: %q ordnet gerade keine Fähigkeiten (CR 603.3b)",
			ErrValidation, wer)
	}

	for _, einer := range p.wartendeVon(wer) {
		if einer.Kennung != kennung {
			continue
		}

		p.Ausloeserreihung = ""
		p = p.legeAusloeser(einer).ausloeserAufDenStapel()

		// **Und danach bekommt jemand Priorität** — CR 603.3 stellt das
		// Legen genau davor, und beide Stellen, die es auslösen
		// ([Partie.abschnittBetreten] und [Partie.nachDemPassen]), tun
		// danach dasselbe.
		p.Zug.Prioritaet = p.prioritaetFuer(p.Zug.Abschnitt, p.Zug.Aktiv)

		return p, nil
	}

	return p, fmt.Errorf(
		"%w: %q wartet nicht unter den Fähigkeiten von %q (CR 603.3b)",
		ErrValidation, kennung, wer)
}

// apnap gibt die Sitze in APNAP-Reihenfolge — CR 101.4.
//
// **Ohne die Gegangenen** (CR 800.4): wer nicht mehr im Spiel ist, legt
// nichts auf den Stapel.
func (p Partie) apnap() []string {
	reihe := make([]string, 0, len(p.Sitze))

	anfang := 0

	for stelle, sitz := range p.Sitze {
		if sitz.Kennung == p.Zug.Aktiv {
			anfang = stelle

			break
		}
	}

	for um := range len(p.Sitze) {
		sitz := p.Sitze[(anfang+um)%len(p.Sitze)]
		if !p.istDraussen(sitz.Kennung) {
			reihe = append(reihe, sitz.Kennung)
		}
	}

	return reihe
}

// faehigkeitAufDenStapel macht aus einer wartenden Fähigkeit ein Objekt.
//
// **Ein Objekt ohne Karte** — CR 603.3 sagt es wörtlich: der Beherrscher legt
// die Fähigkeit als **Objekt, das keine Karte ist**, auf den Stapel. Deshalb
// trägt sie keine Kartenkennung, und deshalb geht sie beim Verrechnen
// nirgendwohin (CR 608.2n).
//
// **Und sie liegt obenauf** (CR 603.3) — dafür sorgt [Partie.zonenwechsel]
// mit CR 405.2.
//
// # Der Name ist eine Abweichung, und sie steht hier
//
// **CR 603.3 sagt, die Fähigkeit habe „den Text der Fähigkeit, die sie
// erzeugt hat, und sonst keine Merkmale".** Einen Text gibt es hier nicht —
// die vier eingebauten stehen im Regelwerk und nicht auf einer Karte. Der
// Name trägt deshalb die **Art**, und das ist die einzige Gestalt, in der
// dieses Programm „welche Fähigkeit" sagen kann. Ein Name im Sinne von
// CR 201 ist es nicht.
func (p Partie) faehigkeitAufDenStapel(einer Ausloeser) Partie {
	kennung := fmt.Sprintf("%s/faehigkeit/%s/%d",
		einer.Beherrscher, einer.Art, p.Zug.Nummer)

	for lauf := 0; ; lauf++ {
		versuch := fmt.Sprintf("%s/%d", kennung, lauf)

		if !p.kennungVergeben(versuch) {
			kennung = versuch

			break
		}
	}

	objekte := append(append([]Spielobjekt{}, p.Objekte...), Spielobjekt{
		Kennung: kennung, Art: ObjektFaehigkeit,
		Beherrscher: einer.Beherrscher, Eigner: einer.Beherrscher,
		Zone: ZoneTisch, Name: string(einer.Art), Seit: p.Zug.Nummer,
	})

	p.Objekte = objekte

	return p.zonenwechsel(len(objekte)-1, ZoneStapel).merkeZiel(
		len(objekte)-1, einer.Wen)
}

// merkeZiel legt fest, auf wen eine Fähigkeit wirkt.
//
// **Als Ziel und nicht als eigenes Feld**: CR 115 nennt das Ziel eines
// Objekts auf dem Stapel, und der Kern führt es dort schon
// ([Spielobjekt.Ziele]).
func (p Partie) merkeZiel(stelle int, wen string) Partie {
	if wen == "" {
		return p
	}

	p.Objekte[stelle].Ziele = []Ziel{{Kennung: wen}}

	return p
}

// kennungVergeben sagt, ob es ein Objekt mit dieser Kennung gibt.
func (p Partie) kennungVergeben(kennung string) bool {
	for _, objekt := range p.Objekte {
		if objekt.Kennung == kennung {
			return true
		}
	}

	return false
}
