// **Was gilt, und woher es kommt** — Werkstück O3 von Phase 5.2
// (#1427).
//
// # Warum die Herkunft dazugehört
//
// **Weil „dunkel" zweierlei heißen kann.** Steht es im Profil, hat es jemand
// für dieses Profil gewählt; steht es beim Menschen, gilt es überall, wo er
// arbeitet; steht es nirgends, ist es die Vorgabe dieser Anwendung. Wer den
// Wert allein sieht, weiß nicht, was ein Zurücksetzen bewirkt — und ein
// Import weiß nicht, ob er etwas überschreibt oder etwas bestätigt.
//
// # Warum die **gesetzten** Werte hereinkommen und nicht die gewählten
//
// **Weil eine Vorgabe keine Aussage ist.** Wer die aufgefüllte Abbildung
// faltete, ließe die Vorgabe eines Profils gegen die Wahl eines Menschen
// antreten — und bei einer Richtlinie gewänne sie, sobald sie strenger ist.
// Eine Ebene, auf der niemand etwas eingetragen hat, sagt nichts.

package domain

// Geltung ist ein Schlüsselwert in drei Gestalten.
//
// **Alle drei, weil drei Fragen daran hängen**: die Auswahlliste zeigt
// `Gewaehlt`, die Anwendung richtet sich nach `Wirksam`, und `Herkunft` sagt,
// wessen Entscheidung das war.
type Geltung struct {
	// Gewaehlt ist, was im Profil steht — die Vorgabe, wo nichts steht.
	Gewaehlt string

	// Wirksam ist, was gilt.
	Wirksam string

	// Herkunft ist `profil`, `nutzer` oder `vorgabe`.
	Herkunft string
}

// WirksamerStand nimmt die **gesetzten** Werte beider Ebenen und gibt je
// Schlüssel der Profilreihe, was gewählt ist, was gilt und woher es kommt.
func WirksamerStand(profil, nutzer map[string]string) map[string]Geltung {
	reihe := Einstellungsreihe()

	stand := make(map[string]Geltung, len(reihe))
	for _, schluessel := range reihe {
		stand[schluessel] = geltungVon(schluessel, profil[schluessel],
			gesetztIn(profil, schluessel), nutzer[schluessel])
	}

	return mitNurLokal(stand)
}

// gesetztIn sagt, ob eine Abbildung diesen Schlüssel führt.
func gesetztIn(werte map[string]string, schluessel string) bool {
	_, da := werte[schluessel]

	return da
}

// geltungVon urteilt über einen Schlüssel.
func geltungVon(schluessel, imProfil string, gesetzt bool, beimNutzer string) Geltung {
	gewaehlt := imProfil
	if !gesetzt {
		gewaehlt = EinstellungsVorgabe(schluessel)
	}

	if IstRichtlinie(schluessel) {
		return richtliniengeltung(schluessel, gewaehlt, imProfil, beimNutzer)
	}

	// **Die alte Regel, unverändert** (`uebersteuerung.go`): das Profil
	// sticht, sonst der Mensch, sonst die Vorgabe. Was O3 hinzufügt, ist
	// nicht der Wert, sondern der Satz darüber, wem er gehört.
	geltung := Geltung{
		Gewaehlt: gewaehlt,
		Wirksam:  Verschmolzen(schluessel, gewaehlt, beimNutzer),
		Herkunft: GeltungVorgabe,
	}

	switch {
	case gesetzt && imProfil != Erben:
		geltung.Herkunft = GeltungProfil
	case gewaehlt == Erben && beimNutzer != "":
		geltung.Herkunft = GeltungNutzer
	}

	return geltung
}

// richtliniengeltung faltet statt zu stechen.
func richtliniengeltung(schluessel, gewaehlt, imProfil, beimNutzer string) Geltung {
	wert, stelle := verschaerftMitStelle(schluessel, imProfil, beimNutzer)
	if stelle < 0 {
		return Geltung{
			Gewaehlt: gewaehlt,
			Wirksam:  EinstellungsVorgabe(schluessel),
			Herkunft: GeltungVorgabe,
		}
	}

	herkunft := GeltungProfil
	if stelle == 1 {
		herkunft = GeltungNutzer
	}

	return Geltung{Gewaehlt: gewaehlt, Wirksam: wert, Herkunft: herkunft}
}

// mitNurLokal legt die übergreifende Schranke über den Stand.
//
// **Eine Regel über mehrere Schlüssel, und deshalb hier und nicht in der
// Faltung**: „nur lokal" sagt nichts über sich selbst, sondern über jeden
// Dienstschalter — und wer das je Schlüssel entschiede, müsste den ganzen
// Stand kennen, um eine Zeile zu beantworten.
//
// # Welche Schlüssel sie trifft
//
// **Genau die Schalter der fremden Dienste** — `dienst.<name>` für jeden
// Dienst aus [Dienste], erkannt wie in der Faltung ([istDienstschalter]).
// Das ist, was [RichtlinieNurLokal] sagt: jeder fremde Dienst ist aus. **Was
// darunter steht, schaltet nichts** und behält seinen Wert: die Adresse eines
// Anbieters (`dienst.<name>.adresse`, #1540) und die vier Pfade des
// Dienstzeugnisses (`dienst.tls.*`, NET-03).
//
// **Bis zum 2026-10-02 traf sie jeden Schlüssel auf `dienst.`** (#2004):
// ein Dienst jenseits der Schleife suchte unter »nur lokal« ein Zertifikat
// namens „aus" und startete nicht, und die Adresse eines Anbieters hieß im
// Stand „aus". Gefunden im Lastlauf QA-08.
func mitNurLokal(stand map[string]Geltung) map[string]Geltung {
	schranke, gibtEs := stand[RichtlinieNurLokal]
	if !gibtEs || schranke.Wirksam != NurLokalAn {
		return stand
	}

	for schluessel, geltung := range stand {
		if !istDienstschalter(schluessel) || geltung.Wirksam == DienstAus {
			// **Ein bereits abgeschalteter Dienst behält seine Herkunft.**
			// Wer ihn selbst ausgeschaltet hat, soll das sehen — die Schranke
			// darüber ändert an seiner Entscheidung nichts.
			continue
		}

		geltung.Wirksam = DienstAus

		// **Die Herkunft ist die der Schranke** und nicht `vorgabe`: das Nein
		// kommt von dem, der „nur lokal" gesagt hat.
		geltung.Herkunft = schranke.Herkunft

		stand[schluessel] = geltung
	}

	return stand
}
