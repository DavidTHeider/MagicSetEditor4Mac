package app

// **Die Menüs, die macOS erwartet** — Anwendung und Bearbeiten, in der
// Sprache des Fensters (#255).
//
// Eine eigene Datei, weil [Menueleiste] mit ihnen über die Zeilengrenze aus
// §2 ginge. Der Schnitt folgt der Frage: dort, was diese Anwendung in die
// Leiste stellt; hier, was macOS an seinen gewohnten Stellen erwartet. Die
// Reihenfolge ist die von macOS selbst, damit dieselbe Handlung überall an
// derselben Stelle steht.

// anwendungsmenue hängt das Anwendungsmenü an — vorn, wie macOS es verlangt.
//
// **„Über" führt auf den Abschnitt der Hilfe**, der Lizenz und Rechtelage
// nennt (§11: sie steht „in der Anwendung selbst unter »Über«"), und nicht auf
// die Standardtafel von Wails, die davon nichts weiß.
func anwendungsmenue(leiste *Menue, handlungen Menuhandlungen, texte Menuetexte) {
	anwendung := leiste.Untermenue(texte.Anwendung)

	anwendung.Befehl(texte.Ueber, nil, rueckruf(handlungen.UeberOeffnen))
	anwendung.Trennlinie()
	anwendung.Befehl(texte.Einstellungen, Befehlstaste(","),
		rueckruf(handlungen.EinstellungenOeffnen))
	anwendung.Trennlinie()
	anwendung.Rolleneintrag(RolleDienste, texte.Dienste)
	anwendung.Trennlinie()
	anwendung.Rolleneintrag(RolleAusblenden, texte.Ausblenden)
	anwendung.Rolleneintrag(RolleAndereAusblenden, texte.AndereAusblenden)
	anwendung.Rolleneintrag(RolleAlleEinblenden, texte.AlleEinblenden)
	anwendung.Trennlinie()
	anwendung.Rolleneintrag(RolleBeenden, texte.Beenden)
}

// bearbeitenmenue hängt das Bearbeitenmenü an.
//
// **Jeder Eintrag hängt an der Responder-Kette** — er wirkt im Textfeld, das
// den Fokus hat, und weiß von Karten nichts. Die fachliche Rücknahme steht
// deshalb im Setmenü und ohne Kürzel (der Vermerk bei
// [Menuhandlungen.Rueckgaengig] sagt, warum ⌘Z sie trotzdem erreicht).
//
// **Ohne „Einsetzen und Stil anpassen"**, das macOS sonst dort führt: Wails
// gibt der Rolle keinen Selektor (`menuitem_selectors_darwin.go`,
// v3.0.0-beta.23), und ein Eintrag, der nichts tut, ist schlimmer als keiner.
func bearbeitenmenue(leiste *Menue, texte Menuetexte) {
	bearbeiten := leiste.Untermenue(texte.Bearbeiten)

	bearbeiten.Rolleneintrag(RolleWiderrufen, texte.Widerrufen)
	bearbeiten.Rolleneintrag(RolleWiederholen, texte.Wiederholen)
	bearbeiten.Trennlinie()
	bearbeiten.Rolleneintrag(RolleAusschneiden, texte.Ausschneiden)
	bearbeiten.Rolleneintrag(RolleKopieren, texte.Kopieren)
	bearbeiten.Rolleneintrag(RolleEinsetzen, texte.Einsetzen)
	bearbeiten.Rolleneintrag(RolleLoeschen, texte.Loeschen)
	bearbeiten.Rolleneintrag(RolleAllesAuswaehlen, texte.AllesMarkieren)
	bearbeiten.Trennlinie()

	sprachausgabe := bearbeiten.Untermenue(texte.Sprachausgabe)
	sprachausgabe.Rolleneintrag(RolleSprechenStarten, texte.SprechenStarten)
	sprachausgabe.Rolleneintrag(RolleSprechenStoppen, texte.SprechenStoppen)
}
