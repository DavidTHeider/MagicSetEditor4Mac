package mcp_test

// **Die Zusagen, unter denen ein Werkzeug schreiben darf** — abgetrennt am
// 2026-09-02, weil `werkzeugumfang_test.go` mit der zwölften über die
// Zeilengrenze aus §2 ging.
//
// # Der Schnitt folgt der Frage
//
// Dort steht, **dass** jedes schreibende Werkzeug unter einer Zusage steht;
// hier stehen **die Zusagen** — je Gruppe, warum sie eine eigene ist und was
// sie nicht sagt. Das ist der Teil, der beim Lesen zählt: die Schleife
// daneben ist drei Zeilen.
//
// **Und `zusagen_test.go` daneben ist eine dritte Datei**, nicht dieselbe:
// dort stehen die **zwei**, die am meisten Text brauchen, seit #1291. Diese
// hier führt die ganze Tabelle und ruft jene auf.

// zusagen gibt jedes schreibende Werkzeug, das unter einer Zusage steht.
//
// **Als eine Tabelle und nicht als zwölf Prüfungen.** Der Aufrufer fragt „ist
// dieses Werkzeug zugesagt", und die Antwort hängt nicht davon ab, unter
// welcher — die Gruppen sind für den **Leser** getrennt, nicht für die
// Rechnung.
func zusagen() map[string]bool {
	// **Der erlaubte Umfang, benannt statt gezählt** — die Karten, die ein
	// Werkzeug in seinen Argumenten nennt, und sonst nichts.
	erlaubt := map[string]bool{
		"schreiben.feld":            true,
		"schreiben.karte.anlegen":   true,
		"schreiben.karte.kopieren":  true,
		"schreiben.karte.loeschen":  true,
		"schreiben.karten.tauschen": true,

		// **Der Abgleich mit Scryfall schreibt eine Karte je Aufruf** (#1005)
		// — die, die in seinen Argumenten steht. Er **überschreibt** dabei
		// Feldwerte und steht deshalb nicht in der Liste darunter; erlaubt
		// ist er aus demselben Grund wie `schreiben.feld`: er gibt die
		// vorigen Werte zurück, und damit ist die Änderung umkehrbar (0010).
		"schreiben.scryfall.uebernehmen": true,

		// **Und der Bildweg ebenso** (#1005). Er schreibt **zwei** Felder
		// derselben Karte — das Bildfeld und »illustrator« —, und beide
		// stehen in seinen Argumenten oder folgen zwingend aus ihnen: die
		// Nennungspflicht von Scryfall ist keine zweite Karte.
		"schreiben.scryfall.bild": true,

		// **Und die Wortwahl** (#1074). Sie schreibt **ein** Feld **einer**
		// Karte — beide stehen in ihren Argumenten —, und sie überschreibt
		// dabei: die Stelle, auf die sich die Wahl bezieht. Umkehrbar ist das
		// wie bei `schreiben.feld`, denn es geht durch denselben Weg und
		// derselbe Verlauf hält den vorigen Wert (0010).
		"schreiben.wortwahl": true,

		// **Und das Kartenbild von der Platte** (#1291). Es schreibt **ein**
		// Feld **einer** Karte — beide stehen in seinen Argumenten — und
		// überschreibt dabei, was dort stand; umkehrbar ist das wie bei
		// `schreiben.feld`, denn es geht durch denselben Weg und derselbe
		// Verlauf hält den vorigen Wert (0010).
		"schreiben.karte.bild": true,
	}

	// **Und die, die schreiben, ohne zu überschreiben** (0044) — die
	// Begründung steht bei [ohneUeberschreiben].
	ohneUeberschreiben := ohneUeberschreiben()

	// **Und die, die den Bestand gar nicht anfassen** (#888).
	//
	// # Warum das eine dritte Zusage ist und keine der beiden darüber
	//
	// „Eine Karte je Aufruf" trifft nicht zu — es sind null. „Nichts, was
	// schon da ist" trifft auch nicht zu — die Sammeldatei **wird** ersetzt.
	// Eines der beiden zu behaupten wäre bequem und falsch.
	//
	// **Was diese Zusage sagt**: das Werkzeug schreibt auf die Platte dieser
	// Maschine, aber in **keine** Karte und in **kein** Set. Es ist damit von
	// 0040 nicht betroffen — die Entscheidung misst den Umfang **im
	// Bestand**, und der ist hier leer.
	//
	// **Was sie nicht sagt**: dass es harmlos ist. Es lädt 25 MB und ersetzt
	// eine Datei; deshalb heißt es `schreiben.` und nicht `lesen.`
	// **Und die Exportvorlage** (#1291). Sie schreibt eine Datei an einem
	// freien Ort, wie das Sichern — und der Optionswert ändert **einen**
	// Schlüssel, wie eine Einstellung.
	fuehrtEineVorlageAus := map[string]bool{
		"schreiben.vorlage.ausfuehren": true,
		"schreiben.vorlagenoption":     true,
	}

	ohneBestandswirkung := map[string]bool{
		"schreiben.scryfall.sammeldaten": true,
	}

	// **Und die, die am **Set** schreiben und an keiner Karte** (#1062).
	//
	// # Warum das eine vierte Zusage ist
	//
	// „Eine Karte je Aufruf" trifft nicht zu — es sind null. „Nichts, was
	// schon da ist" auch nicht: `setzen` überschreibt. Und „ohne
	// Bestandswirkung" wäre schlicht falsch.
	//
	// **Was diese Zusage sagt**: das Werkzeug ändert **ein** Schlüsselwort
	// **eines** Sets, und beide stehen in seinen Argumenten. Keine Karte wird
	// angefasst; die Reihenfolge der übrigen bleibt.
	//
	// **Und seit #1077 ist es umkehrbar.** Hier stand das Gegenteil: „für
	// Schlüsselwörter gibt es noch keine Rücknahme". Jetzt gibt es sie —
	// anlegen, ändern und löschen schreiben einen Verlaufsschritt, und ein
	// zurückgenommenes Löschen bringt das Wort an **seine** Stelle zurück.
	// Die Zusage selbst bleibt dieselbe: **ein** Schlüsselwort **eines** Sets.
	//
	// **Und `schreiben.setangabe` gehört dazu** (#1107): es setzt Titel,
	// Setcode oder Copyright — eine Angabe **eines** Sets, und beide stehen in
	// seinen Argumenten. Keine Karte wird angefasst.
	amSetUndKeinerKarte := map[string]bool{
		"schreiben.schluesselwort.anlegen":  true,
		"schreiben.schluesselwort.setzen":   true,
		"schreiben.schluesselwort.loeschen": true,
		"schreiben.setangabe":               true,

		// **Und die drei Stilwerkzeuge** (#1107). Ein Stilwechsel schreibt
		// `set_style` und die Stilwerte — **keine Karte**
		// (`service.Stilwaehler.Wechseln`, nachgelesen am 2026-08-25). Die
		// „ungedeuteten Felder", die `lesen.stilfolgen` nennt, bleiben im
		// Bestand stehen; sie werden vom neuen Stil nur nicht mehr gezeichnet.
		"schreiben.stil":                     true,
		"schreiben.stiloption":               true,
		"schreiben.stiloption.zuruecksetzen": true,
	}

	// **Und das, was ein Set **anlegt** und keines anfasst** (#891).
	//
	// # Warum das eine fünfte Zusage ist
	//
	// Keine der vier trifft zu: es schreibt keine Karte, es überschreibt
	// nichts (es entsteht ein **neues** Set), es hat sehr wohl eine
	// Bestandswirkung, und es ändert kein vorhandenes Set.
	//
	// **Was diese Zusage sagt**: der Aufruf legt **ein** leeres Set an und
	// lässt jedes vorhandene, wie es war. Was danach darin steht, entscheiden
	// die Kartenwerkzeuge — mit ihren eigenen Zusagen.
	//
	// **Was sie nicht sagt**: dass es umkehrbar ist. Die Gegenrichtung ist
	// seit #1218 ein **eigenes** Werkzeug mit eigener Zusage — sie steht
	// weiter unten, und sie ist das Gegenteil einer Rücknahme: sie nimmt
	// nicht diesen einen Aufruf zurück, sondern das ganze Set.
	//
	// **Hier stand bis zum 2026-08-27 „ein Set löschen kann nur das
	// Fenster".** Der Satz war falsch, als er geschrieben wurde: die
	// Fähigkeit gab es nirgends — zwei Zeilen weiter stand es sogar richtig.
	legtEinSetAn := map[string]bool{
		"schreiben.set.neu": true,
	}

	// **Und das, was ein ganzes Set **wegnimmt** (#1218).
	//
	// # Warum das eine eigene Zusage ist
	//
	// Weil es die **einzige unumkehrbare** dieser Fläche ist. Jede andere
	// schreibt etwas, das sich zurückschreiben lässt — ein Feld, eine Karte,
	// eine Zeile der Sammlung, ein Schritt des Verlaufs. Diese nicht: nach
	// ihr gibt es das Set nicht mehr, und es gibt keinen Papierkorb
	// ([0054](../../docs/entscheidungen/0054-ein-set-laesst-sich-loeschen.md)).
	//
	// **Was diese Zusage sagt**: der Aufruf nimmt **ein** Set samt allem, was
	// an ihm hängt, und lässt jedes andere stehen. Er verlangt den **heutigen
	// Namen** als Bestätigung; stimmt er nicht, geschieht nichts. Die
	// **Sammlung** bleibt — sie hängt an Drucken und nicht an Sets (0052).
	//
	// **Was sie nicht sagt**: dass die eingesetzten Bilder mitgehen. Sie
	// liegen auf der Platte, und wenn sie sich nicht wegräumen lassen, steht
	// das als **Befund** in der Antwort — das Set ist trotzdem weg (§7,
	// Regel 3).
	nimmtEinSetWeg := map[string]bool{
		"schreiben.set.loeschen": true,
	}

	// **Und das, was ein Vorlagenpaket von der Platte nimmt** (#1291) — die
	// Begründung steht bei [nimmtEinPaketWeg].
	paketweg := nimmtEinPaketWeg()

	// **Und das, was den Bestand als Ganzes betrifft und kein Set** (#1107).
	//
	// # Warum das eine sechste Zusage ist
	//
	// Eine Einstellung gehört keinem Set. Sie steht im Bestand — also nicht
	// „ohne Bestandswirkung" — und trotzdem ändert sie kein Set, keine Karte
	// und keine Datei.
	//
	// **Was diese Zusage sagt**: der Aufruf ändert **einen** Schlüssel der
	// Anwendung und lässt jedes Set, wie es war. Die erlaubten Werte stehen in
	// `lesen.einstellungen`, und ein anderer wird abgewiesen.
	amBestandUndKeinemSet := map[string]bool{
		"schreiben.einstellung": true,
	}

	// **Und das, was einen Schritt des Verlaufs bewegt** (#1107).
	//
	// # Warum das eine siebte Zusage ist
	//
	// Sie ist die **einzige**, die mehr als eine Karte anfassen darf — und sie
	// darf es, weil sie nichts Neues schreibt: sie stellt einen Stand her, den
	// es schon gab. Eine Massenänderung ist **ein** Schritt (§ und `Umfang`),
	// und ein halb zurückgenommener Schritt wäre schlimmer als ein ganzer.
	//
	// **Was diese Zusage sagt**: der Aufruf bewegt genau **einen** Schritt im
	// Verlauf **eines** Sets — vorwärts oder rückwärts —, und `bewegt` sagt,
	// ob es einen gab.
	//
	// **Was sie nicht sagt**: dass sie über Setgrenzen hinweg wirkt. Der
	// Verlauf wird je Set geführt.
	bewegtEinenSchritt := map[string]bool{
		"schreiben.zuruecknehmen":    true,
		"schreiben.wiederherstellen": true,
	}

	// **Und das, was Dateien holt und keinen Bestand anfasst** (#1107).
	//
	// # Warum das eine achte Zusage ist
	//
	// `schreiben.paket.beziehen` lädt aus dem Netz und legt ein Verzeichnis
	// ab. Es ändert **kein** Set, **keine** Karte und **keine** Einstellung —
	// aber es ist das folgenreichste Werkzeug dieser Liste, weil es den
	// Rechner anfasst und Bytes aus dem Netz nimmt.
	//
	// **Was diese Zusage sagt**: der Aufruf legt **ein** Paket unter dem
	// genannten Namen ab und lässt jedes vorhandene, wie es war — ein
	// vorhandenes ist ein Konflikt und kein Überschreiben
	// (`service.Paketverwalter.Installieren`). Was nicht ausgepackt wurde,
	// steht in `uebergangen`.
	//
	// **Was sie nicht sagt**: dass die Quelle vertrauenswürdig ist. Das
	// entscheidet der Aufrufer, und §11 verlangt dafür eine ausdrückliche
	// Anweisung.
	holtEinPaket := map[string]bool{
		"schreiben.paket.beziehen": true,
	}

	// **Und eines ändert eine Vorlage** —
	// #1164.
	//
	// Es ist das einzige Werkzeug, das **keine Karte** anfasst und trotzdem
	// jede ändert: eine Vorlage gilt für alle Sets, die ihr Paket benutzen.
	// Genau deshalb steht es hier als eigene Zusage und nicht bei einer der
	// anderen — sein Umfang ist ein anderer.
	//
	// **Was es einlöst**: es schreibt **eine** Datei je Aufruf, nur wenn der
	// mitgegebene Stand noch stimmt, und der vorige Stand bleibt als `.bak`
	// neben dem Paket liegen. Damit ist es zurückzunehmen.
	//
	// **Und der Dateibaum** (FMT-02, #1691) löst dieselbe Zusage ein: eine
	// Datei je Aufruf; angelegt wird nur, wo nichts steht; eine Vorlage wird
	// nur mit ihrem Stand umbenannt oder entfernt (`service.mitStand`), und
	// was entfernt wird, liegt danach in der `.bak`.
	aendertEineVorlage := map[string]bool{
		"schreiben.paketdatei":            true,
		"schreiben.paketdatei.anlegen":    true,
		"schreiben.paketdatei.umbenennen": true,
		"schreiben.paketdatei.entfernen":  true,
	}

	// **Und drei ändern die Sammlung** —
	// #1165.
	//
	// # Warum das eine zehnte Zusage ist
	//
	// Sie fassen **keine Karte, kein Set und keine Einstellung** an. Was sie
	// ändern, ist eine Aussage über den **Besitz**: was jemand hat, in welchem
	// Zustand, in welcher Sprache
	// ([0052](../../docs/entscheidungen/0052-die-sammlung-ist-nicht-das-set.md)).
	//
	// **Was diese Zusage sagt**: der Aufruf ändert **eine** Zeile der
	// Sammlung. Dieselbe Karte in demselben Zustand, derselben Sprache und
	// derselben Ausführung gibt es einmal — ein zweites Anlegen ist ein
	// Konflikt und kein Überschreiben.
	//
	// **Was sie nicht sagt**: dass es einen Verlauf gäbe, der sie zurücknimmt.
	// Den gibt es für die Sammlung **nicht** — sie steht neben den
	// Kreativdaten, und deren Verlauf führt sie nicht mit.
	aendertEineSammlungszeile := map[string]bool{
		"schreiben.exemplar":         true,
		"schreiben.exemplaraendern":  true,
		"schreiben.exemplarloeschen": true,
	}

	for name := range kiVorschlagszusage() {
		erlaubt[name] = true
	}

	return zusammengefasst(
		erlaubt, ohneUeberschreiben, ohneBestandswirkung, amSetUndKeinerKarte,
		legtEinSetAn, nimmtEinSetWeg, paketweg, fuehrtEineVorlageAus,
		amBestandUndKeinemSet, bewegtEinenSchritt, holtEinPaket,
		aendertEineVorlage, aendertEineSammlungszeile,
		weitereZusagen(),
	)
}
