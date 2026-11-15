package domain

// **Ausgelöste Fähigkeiten** — Werkstück M9 der Regelmaschine
// (#1611).
//
// # Welche überhaupt auslösen können
//
// **Die eingebauten.** CR 603.1 beschreibt eine ausgelöste Fähigkeit als
// **Kartentext**, und Kartentexte gibt es ohne `M8`
// (#1610)
// nicht. **Drei stehen aber im Regelwerk selbst** und haben ausdrücklich
// **keine Quelle** — die CR nennt das eine Ausnahme zu CR 113.8:
//
//   - **CR 725.2**, zwei für den Monarchen: er zieht im Endschritt, und
//     wer ihm Kampfschaden zufügt, wird selbst Monarch.
//   - **CR 728.1**, eine für die Radioaktivitätsmarken.
//
// **Damit trägt dieses Werkstück seinen eigenen Gegenstand.** Ohne sie wäre
// es ein Mechanismus ohne einen einzigen Fall — ein Paket ohne Aufrufer
// (§10).
//
// # Die drei Regeln, nach denen es läuft
//
// **CR 603.2**: eine Fähigkeit löst aus und **tut nichts**. Sie steht in der
// Warteschlange, nicht auf dem Stapel.
//
// **CR 603.3**: sie kommt auf den Stapel, wenn das nächste Mal jemand
// Priorität bekäme — also in [Partie.abschnittBetreten] und nach jedem
// Verrechnen, an denselben Stellen wie die zustandsbasierten Aktionen
// (CR 704.3).
//
// **CR 603.3b mit CR 101.4**: der aktive Spieler legt zuerst, dann die
// anderen in Zugreihenfolge. Und weil der Stapel von oben abgeräumt wird
// (CR 405.2), wird seine **zuletzt** verrechnet.

import "fmt"

// Ausloeserart sagt, welche eingebaute Fähigkeit ausgelöst hat.
//
// **Eine Aufzählung und kein Text**: was eine Fähigkeit tut, steht im
// Regelwerk und nicht in dieser Partie. Ein Text hier wäre eine zweite
// Quelle für dieselbe Regel.
type Ausloeserart string

const (
	// AusloeserMonarchzug — CR 725.2, erste der beiden.
	AusloeserMonarchzug Ausloeserart = "monarch.zug"

	// AusloeserMonarchwechsel — CR 725.2, zweite der beiden.
	AusloeserMonarchwechsel Ausloeserart = "monarch.wechsel"

	// AusloeserInitiative — CR 726.2, die zweite der drei.
	//
	// **Nur die zweite.** Die erste und die dritte lassen den Träger „in
	// den Undercity vorstoßen" (CR 701.49), und ein Verlies ist eine Karte
	// — die gibt es ohne `M8` nicht. Das steht als Grenze in der
	// Regelmatrix und nicht als Auslassung (§10).
	AusloeserInitiative Ausloeserart = "initiative.wechsel"

	// AusloeserStrahlung — CR 728.1.
	AusloeserStrahlung Ausloeserart = "strahlung"
)

// Ausloeser ist eine Fähigkeit, die ausgelöst hat und wartet — CR 603.2.
type Ausloeser struct {
	// Art sagt, welche Regel ausgelöst hat.
	Art Ausloeserart

	// Beherrscher ist, wer sie kontrolliert — CR 603.3.
	//
	// **Bei den eingebauten sagt es die Regel**: die des Monarchen gehören
	// dem, der zur Zeit des Auslösens Monarch war (CR 725.2), die der
	// Marken dem aktiven Spieler (CR 728.1).
	Beherrscher string

	// Wen ist, auf wen sie wirkt — leer, wo sie den Beherrscher meint.
	Wen string

	// Kennung unterscheidet zwei wartende Fähigkeiten voneinander.
	//
	// **Weil die Reihung sie nennen muss** (CR 603.3b): zwei Auslöser
	// derselben Art und desselben Sitzes sind sonst dasselbe Wort, und ein
	// Befehl könnte nicht sagen, welcher zuerst gelegt wird.
	//
	// **Sie wird beim Auslösen vergeben** ([Partie.ausloesen]) und gilt nur,
	// solange die Fähigkeit wartet — auf dem Stapel trägt das Objekt seine
	// eigene (CR 603.3).
	Kennung string
}

// ausloesen stellt eine Fähigkeit in die Warteschlange — CR 603.2.
//
// **Und gibt ihr eine Kennung**, unter der die Reihung sie ansprechen kann
// (CR 603.3b). Sie zählt hoch, bis eine frei ist — zwei Wartende mit
// derselben wären für einen Befehl dieselbe.
func (p Partie) ausloesen(einer Ausloeser) Partie {
	for lauf := 0; ; lauf++ {
		einer.Kennung = fmt.Sprintf("%s/%s/%d", einer.Beherrscher, einer.Art, lauf)

		if !p.wartetSchon(einer.Kennung) {
			break
		}
	}

	p.Ausgeloest = append(append([]Ausloeser{}, p.Ausgeloest...), einer)

	return p
}

// wartetSchon sagt, ob eine wartende Fähigkeit diese Kennung trägt.
func (p Partie) wartetSchon(kennung string) bool {
	for _, einer := range p.Ausgeloest {
		if einer.Kennung == kennung {
			return true
		}
	}

	return false
}

// faehigkeitVerrechnen führt eine eingebaute Fähigkeit aus — CR 608.2.
//
// **Und danach hört sie auf zu existieren** — CR 608.2n mit CR 603.3: sie
// ist keine Karte und geht in keine Zone.
func (p Partie) faehigkeitVerrechnen(stelle int) Partie {
	objekt := p.Objekte[stelle]

	wen := objekt.Beherrscher
	if len(objekt.Ziele) > 0 {
		wen = objekt.Ziele[0].Kennung
	}

	switch Ausloeserart(objekt.Name) {
	// **CR 725.2**, erste: der Monarch zieht.
	case AusloeserMonarchzug:
		p = p.ziehtEine(objekt.Beherrscher)

	// **CR 725.2**, zweite: wer Kampfschaden zufügte, wird Monarch.
	case AusloeserMonarchwechsel:
		p.Monarch = wen

	// **CR 726.2**, zweite: wer Kampfschaden zufügte, nimmt die Initiative.
	case AusloeserInitiative:
		p.Initiative = wen

	// **CR 728.1**: mahlen, Leben je Nichtland, Marken herunter.
	case AusloeserStrahlung:
		p = p.strahlungsschaden(objekt.Beherrscher)
	}

	return p.ohneObjektAn(stelle)
}

// ohneObjektAn nimmt ein Objekt aus der Partie — es hört auf zu existieren.
func (p Partie) ohneObjektAn(stelle int) Partie {
	objekte := make([]Spielobjekt, 0, len(p.Objekte)-1)
	objekte = append(objekte, p.Objekte[:stelle]...)
	objekte = append(objekte, p.Objekte[stelle+1:]...)

	p.Objekte = objekte

	return p
}

// strahlungsschaden führt die Fähigkeit der Radioaktivitätsmarken aus —
// CR 728.1.
//
// **Drei Schritte in einem**: so viele Karten mahlen, wie Marken daliegen;
// je gemahlenem **Nichtland** ein Leben; und ebenso viele Marken herunter.
//
// **Gezählt wird das Gemahlene und nicht das Gewollte**: eine Bibliothek mit
// zwei Karten und drei Marken gibt zwei — CR 701.17a sagt, dass so viele
// gehen, wie da sind.
func (p Partie) strahlungsschaden(wer string) Partie {
	// **Den Sitz gibt es** — die Fähigkeit ist von ihm ausgelöst worden
	// (CR 728.1), und ein Sitz verschwindet nicht aus der Liste, wenn er
	// die Partie verlässt (CR 104.3). Ein Zweig dafür wäre einer, den kein
	// Weg betritt (§10).
	marken := 0

	for _, sitz := range p.Sitze {
		if sitz.Kennung == wer {
			marken = sitz.Strahlung
		}
	}

	// **Null heißt: nichts zu tun.** Zwischen Auslösen und Verrechnen kann
	// jemand die Marken wegnehmen, und CR 728.1 trägt die Bedingung als
	// Nachprüfung beim Verrechnen.
	if marken <= 0 {
		return p
	}

	verloren := 0

	for range marken {
		stelle := obersteDerBibliothek(p.Objekte, wer)
		if stelle < 0 {
			break
		}

		nichtland := !p.istLand(p.Objekte[stelle])

		p = p.zonenwechsel(stelle, ZoneFriedhof)
		p.Objekte[stelle].Beherrscher = p.Objekte[stelle].Eigner

		if nichtland {
			verloren++
		}
	}

	if verloren == 0 {
		return p
	}

	sitze := make([]Sitz, len(p.Sitze))
	copy(sitze, p.Sitze)

	for stelle := range sitze {
		if sitze[stelle].Kennung != wer {
			continue
		}

		// **Leben je Nichtland und ebenso viele Marken herunter** —
		// CR 728.1 nennt beides in einem Satz, und beide Zahlen sind
		// dieselbe.
		sitze[stelle].Leben -= verloren
		sitze[stelle].Strahlung = max(0, sitze[stelle].Strahlung-verloren)
	}

	p.Sitze = sitze

	return p
}

// istLand sagt, ob ein Objekt ein Land ist — CR 305.
func (p Partie) istLand(objekt Spielobjekt) bool {
	eigen, da := p.Eigenschaften(objekt)

	return da && eigen.Typzeile.Ist(TypLand)
}

// monarchZiehtImEndschritt löst die erste Fähigkeit aus CR 725.2 aus.
//
// **Zu Beginn des Endschritts**, und nur, wenn es einen Monarchen gibt
// (CR 725.1).
func (p Partie) monarchZiehtImEndschritt() Partie {
	if p.Monarch == "" || p.istDraussen(p.Monarch) {
		return p
	}

	return p.ausloesen(Ausloeser{
		Art: AusloeserMonarchzug, Beherrscher: p.Monarch,
	})
}

// strahlungInDerHauptphase löst die Fähigkeit aus CR 728.1 aus.
//
// **Zu Beginn der ersten Hauptphase jedes Spielers**, und sie gehört dem
// **aktiven** Spieler — CR 728.1 sagt beides.
func (p Partie) strahlungInDerHauptphase() Partie {
	sitz := p.SitzVon(p.Zug.Aktiv)
	if sitz == nil || sitz.Strahlung <= 0 {
		return p
	}

	return p.ausloesen(Ausloeser{
		Art: AusloeserStrahlung, Beherrscher: p.Zug.Aktiv,
	})
}

// monarchwechselDurchKampfschaden löst die zweite Fähigkeit aus CR 725.2
// aus.
//
// **Wer dem Monarchen Kampfschaden zufügt**, dessen Beherrscher wird
// Monarch — und die Fähigkeit gehört dem, der zur Zeit des Auslösens
// Monarch **war** (CR 725.2).
func (p Partie) monarchwechselDurchKampfschaden(eine zuteilung) Partie {
	if p.Monarch == "" || eine.Sitz != p.Monarch {
		return p
	}

	// **Eine Schleife und kein Nachschlagen**: eine Zuteilung entsteht nur,
	// wo eine Kreatur Stärke hat — also gibt es sie (§10, derselbe Grund
	// wie bei `commanderschadenZaehlen`).
	for _, quelle := range p.Objekte {
		if quelle.Kennung != eine.Von || !p.istKreatur(quelle) ||
			quelle.Beherrscher == p.Monarch {
			continue
		}

		return p.ausloesen(Ausloeser{
			Art:         AusloeserMonarchwechsel,
			Beherrscher: p.Monarch,
			Wen:         quelle.Beherrscher,
		})
	}

	return p
}

// istKreatur sagt, ob ein Objekt eine Kreatur ist — CR 302.
func (p Partie) istKreatur(objekt Spielobjekt) bool {
	eigen, da := p.Eigenschaften(objekt)

	return da && eigen.Typzeile.Ist(TypKreatur)
}

// initiativeDurchKampfschaden löst die zweite Fähigkeit aus CR 726.2 aus.
//
// **Je Beherrscher und nicht je Kreatur.** CR 726.2 fragt nach „einer oder
// mehr Kreaturen, die ein Spieler beherrscht" — zwei Bären desselben Sitzes
// sind **ein** Auslöser. CR 725.2 fragt nach **einer** Kreatur, und dort sind
// zwei Bären zwei; der Unterschied steht so im Regelwerk und ist keiner, den
// dieses Programm gemacht hat.
//
// **Und sie gehört dem, der die Initiative zur Zeit des Auslösens hatte**
// (CR 726.2) — nicht dem, der sie bekommt.
func (p Partie) initiativeDurchKampfschaden(zuteilungen []zuteilung) Partie {
	if p.Initiative == "" {
		return p
	}

	// **Die Reihenfolge ist die der Zuteilungen** und nicht die einer
	// Tabelle: zwei Läufe über dieselbe Partie sollen dieselben Auslöser in
	// derselben Folge geben.
	var nehmer []string

	gesehen := map[string]bool{}

	for _, eine := range zuteilungen {
		if eine.Sitz != p.Initiative {
			continue
		}

		for _, quelle := range p.Objekte {
			if quelle.Kennung != eine.Von || !p.istKreatur(quelle) ||
				gesehen[quelle.Beherrscher] {
				continue
			}

			gesehen[quelle.Beherrscher] = true

			nehmer = append(nehmer, quelle.Beherrscher)
		}
	}

	for _, wer := range nehmer {
		p = p.ausloesen(Ausloeser{
			Art: AusloeserInitiative, Beherrscher: p.Initiative, Wen: wer,
		})
	}

	return p
}
