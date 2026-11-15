// Die Liste der Werkzeuge dieses Servers.
//
// **Von `bestand.go` getrennt**, als sie über 400 Zeilen wuchs (§2), später
// noch einmal von `ganzedateien.go`. Der Schnitt folgt der Frage: dort das
// **Was**, hier das **Welche** und in welcher Folge.

package mcp

// Werkzeuge baut die Werkzeugliste über einem Bestand.
//
// **Die lesenden zuerst, die schreibenden danach** — dieselbe Reihenfolge wie
// in `tools/list`, und wer die Liste kürzt, behält damit die harmlosen.
func Werkzeuge(naehte Naehte) []Werkzeug {
	leser, schreiber, ausfuhr := naehte.Leser, naehte.Schreiber, naehte.Ausfuhr
	abgleich, bilder, auskunft := naehte.Abgleich, naehte.Bilder, naehte.Auskunft
	sammeldaten, anlage, saetze := naehte.Sammeldaten, naehte.Anlage, naehte.Saetze
	konsole, wortlisten, woerter := naehte.Konsole, naehte.Wortlisten, naehte.Woerter
	pack, anleger, angaben := naehte.Pack, naehte.Anleger, naehte.Angaben
	loescher, sammlung, preise := naehte.Loescher, naehte.Sammlung, naehte.Preise
	stilquelle, bestand, werkstatt := naehte.Stile, naehte.Bestand, naehte.Werkstatt

	// **Der Lebenszyklus gehört dazu und wird nicht erfragt.** Der erste
	// Anlauf holte ihn über eine Typzusicherung — und ein Aufrufer, der sie
	// nicht erfüllt, verlöre vier Werkzeuge **stillschweigend**. Eine
	// Schnittstelle, die der Übersetzer prüft, sagt es beim Bauen
	// (#591, §10).
	// **Und der Arbeitsablauf** ([0044](../../docs/entscheidungen/0044-was-ein-werkzeug-schreiben-darf.md)):
	// einfahren, sichern, zeichnen. Sie schreiben nur, was sie nicht
	// überschreiben — deshalb dürfen sie mehr als eine Karte anfassen.
	// **Und der Abgleich mit Scryfall** (#1005): zeigen und übernehmen, in
	// dieser Reihenfolge. Für ein Sprachmodell ist die Zweiteilung die
	// wichtigere Hälfte von §16 Regel 6 — es ruft, während niemand hinsieht.
	sammellesend, sammelholend := sammeldatenwerkzeuge(sammeldaten)

	lesendes := append(append(append(lesende(leser), scryfallLesend(abgleich)...),
		scryfallauskunftswerkzeuge(auskunft)...), sammellesend...)

	// **Und die Skriptkonsole** (#1070). Sie steht bei den lesenden, weil sie
	// nichts ändert — und ein Sprachmodell braucht sie, um zu **prüfen**, was
	// eine Vorlage rechnet, statt es zu raten.
	// **Und die Kartensuche** (#1167). Sie steht bei den lesenden, weil sie
	// nichts ändert — und ohne sie kann ein Sprachmodell eine Karte nur
	// finden, deren Namen es schon kennt.
	lesendes = append(lesendes, suchwerkzeug(naehte.Suche)...)

	// **Und der Dienstkatalog** (#1440) — was die Suche und ihre
	// Geschwister **können**, bevor ein Modell sie ruft.
	lesendes = append(lesendes, katalogwerkzeug(naehte.Katalog)...)

	// **Und die Grenzen** (#1429) — was ein Ruf kostet, bevor ein Modell ihn tut.
	lesendes = append(lesendes, grenzenwerkzeug(naehte.Grenzen)...)

	// **Und die eigene Peeridentität** (N1, #1454) — wer ein Austausch
	// vorbereitet, muss sagen können, wer dieses Programm ist.
	lesendes = append(lesendes, peerwerkzeug(naehte.Peer)...)

	// **Und die Gegenüber** (N1, #1454) — lesen hier, sperren unten.
	kontaktlesend, kontaktschreibend := kontaktwerkzeuge(naehte.Kontakte)

	lesendes = append(lesendes, kontaktlesend...)
	lesendes = append(lesendes, gesundheitswerkzeug(naehte.Gesundheit)...)
	lesendes = append(lesendes, metrikwerkzeug(naehte.Metriken)...)
	lesendes = append(lesendes, metawerkzeug(naehte.Meta)...)
	lesendes = append(lesendes, vertraeglichkeitswerkzeug(naehte.Vertraeglichkeit)...)
	lesendes = append(lesendes, ruecksprungwerkzeuge(naehte.Ruecksprung)...)
	lesendes = append(lesendes, pflegewerkzeug(naehte.Datenpflege)...)
	lesendes = append(lesendes, ruecksetzungswerkzeug(naehte.Ruecksetzung)...)
	lesendes = append(lesendes, kostenwerkzeug(naehte.Kosten)...)

	// **Und der lokale Suchindex** (A15, #1514) — Stand und Suche, beide
	// lesend. Der **Aufbau** steht nicht hier: er läuft Minuten, bindet einen
	// Kern und ist von hier aus nicht abbrechbar (`retrievalindex.go`).
	lesendes = append(lesendes, indexwerkzeuge(naehte.Index)...)

	// **Und die Recherche darüber** (A16, #1515) — Frage und Ähnlichkeit,
	// beide lesend. Das **Kontextpaket** steht nicht hier: es budgetiert für
	// ein Zielmodell, und hier ist der Aufrufer das Modell (`recherche.go`).
	lesendes = append(lesendes, recherchewerkzeuge(naehte.Recherche)...)

	// **Und die Verweise** (#1167) — eines lesend, zwei schreibend. Ohne sie
	// holt jeder Abgleich irgendeinen Druck, und ein Modell, das „nimm die
	// Alpha-Fassung" gehört hat, kann es nirgends festhalten.
	verweislesend, verweisschreibend := verweiswerkzeuge(naehte.Verweise)

	lesendes = append(lesendes, verweislesend...)

	// **Und die Marken** (#1208) — zwei lesend, zwei schreibend. Wer ein
	// Modell bittet, ein Set durchzusehen, will das Ergebnis irgendwo stehen
	// haben; ohne Marken bliebe ihm nur eine Antwort, die niemand wiederfindet.
	markenlesend, markenschreibend := markenwerkzeuge(naehte.Marken)

	lesendes = append(lesendes, markenlesend...)

	// **Und die Zwischenablage** (#1168) — eines lesend, zwei schreibend.
	// Ohne sie kann ein Modell eine Karte nur **innerhalb** eines Sets
	// duplizieren; zwischen zwei Sets ging nichts.
	ablagelesend, ablageschreibend := ablagewerkzeuge(naehte.Ablage)

	lesendes = append(lesendes, ablagelesend...)

	// **Und das Ersetzen** (#1168) — eines lesend, eines schreibend. Das
	// lesende ist die **Vorschau**: ein Modell soll zeigen können, was daraus
	// wird, bevor es dreihundert Karten ändert.
	ersetzlesend, ersetzschreibend := ersetzungswerkzeuge(naehte.Ersetzung)

	lesendes = append(lesendes, ersetzlesend...)

	// **Und die Gestalt der Kartenliste** (#1168) — eines lesend, zwei
	// schreibend. `lesen.karten` gibt alle Felder und keine Ordnung; wer eine
	// Liste zeigen will, fragt hier.
	listelesend, listeschreibend := listenwerkzeuge(naehte.Liste)

	lesendes = append(lesendes, listelesend...)

	// **Und die Nutzer** (#1351) — eines lesend, zwei schreibend. Ohne das
	// lesende kennte ein Modell die Kennungen nicht, die das Anlegen eines
	// Profils verlangt.
	nutzerlesend, nutzerschreibend := nutzerwerkzeuge(naehte.Nutzer)

	lesendes = append(lesendes, nutzerlesend...)

	// **Und die Zugriffstoken** (#1351, Werkstück 10) — eines lesend, drei
	// schreibend. **Verwalten heißt nicht vorzeigen**: was ein Modell hier
	// anlegt, weist einen Automaten am REST-Zugang aus, und dorthin reicht
	// MCP nicht.
	tokenlesend, tokenschreibend := tokenwerkzeuge(naehte.Token)

	lesendes = append(lesendes, tokenlesend...)

	// **Und die Profilvorgaben** (#1351, Werkstück 15) — eines lesend, eines
	// schreibend. Ohne das lesende erriete ein Modell die Kennungen, auf die
	// eine Vorgabe zeigt.
	vorgabenlesend, vorgabenschreibend := vorgabenwerkzeuge(naehte.Vorgaben)

	lesendes = append(lesendes, vorgabenlesend...)

	// **Und die Freigaben** (#1351, Werkstück 11) — eines lesend, zwei
	// schreibend. Ohne das lesende wüsste ein Modell nicht, ob der Mensch
	// dahinter überhaupt vergeben darf.
	freigabenlesend, freigabenschreibend := freigabenwerkzeuge(naehte.Freigaben)
	uebernahmelesend, uebernahmeschreibend := deckuebernahmewerkzeug(naehte.Deckuebernahme)

	lesendes = append(append(lesendes, freigabenlesend...), uebernahmelesend...)

	// **Und die Wege nach draußen** (T8, #1473) — **nur lesend**: ein Modell
	// nennt sie und öffnet keine.
	lesendes = append(lesendes, marktlinkwerkzeug(naehte.Links)...)
	lesendes = append(lesendes, marktkontowerkzeug(naehte.Konto)...)
	lesendes = append(lesendes, handelsstandwerkzeug(naehte.Handel)...)

	// **Und der Platz** (Werkstück 4 von Phase 4.13, #1352) — eines lesend,
	// eines schreibend. Ohne das lesende antwortete ein Modell auf „warum ist
	// meine Platte voll?" mit einer Vermutung.
	platzlesend, platzschreibend := platzwerkzeuge(naehte.Platz)

	lesendes = append(lesendes, platzlesend...)

	lesendes = append(lesendes, skriptwerkzeug(konsole)...)
	lesendes = append(lesendes, namenswerkzeug(konsole)...)

	// **Und die Wortlisten** (#1074) — je eines auf jeder Seite. Das lesende
	// steht bei den lesenden, weil es nichts ändert; ohne es erfände ein
	// Modell die Untertypen, die ein Spiel kennt.
	wortlesend, wortschreibend := wortlistenwerkzeuge(wortlisten)

	lesendes = append(lesendes, wortlesend...)

	// **Und die Schlüsselwörter** (#1062) — eines lesend, drei schreibend.
	// Ohne das lesende schriebe ein Modell ein zweites »Fliegend« ins Set, und
	// zwei Schlüsselwörter auf denselben Text geben zwei Erinnerungstexte auf
	// derselben Karte.
	schluessellesend, schluesselschreibend := schluesselwortwerkzeuge(woerter)

	lesendes = append(lesendes, schluessellesend...)

	// **Und das Zufallspack** (#889, Befund 7 des Zweitblicks über Phase W) —
	// beide lesend, weil Ziehen nichts ändert: es rechnet aus einem Samen eine
	// Auswahl. Bis dahin gab es das Pack **nur im Fenster**, ohne dass
	// irgendwo gestanden hätte, warum (§16).
	lesendes = append(lesendes, packwerkzeuge(pack)...)

	// **Und die Setangaben** (#1107) — drei lesend, eines schreibend. Ohne sie
	// kann ein Modell Karten anlegen, aber dem Set keinen Titel geben; und die
	// Vorbehalte sind die Gegenprobe zum Einfahren (§8). Daneben die Probleme
	// eines ganzen Sets (PERF02, #2352), beide lesend.
	angabenlesend, angabenschreibend := setangabenwerkzeuge(angaben)

	lesendes = append(append(lesendes, angabenlesend...), problemwerkzeuge(naehte.Probleme)...)

	// **Und die Stile** (#1107) — drei lesend, drei schreibend. Ohne sie kann
	// ein Modell Karten schreiben, aber nicht sagen, wie sie aussehen sollen.
	stillesend, stilschreibend := stilwerkzeuge(stilquelle, naehte.Zusatz)

	lesendes = append(lesendes, stillesend...)

	// **Und der Bestand** (#1107) — Pakete, Kennzahlen, Verlauf, Einstellungen.
	bestandlesend, bestandschreibend := bestandswerkzeuge(bestand)

	// **Dateien** (#1169, #1457), Netzweg und Regelfreigabe (G4, #1461).
	dateilesend, dateischreibend := werkzeugeAnGanzenDateien(naehte)

	lesendes = append(append(append(append(lesendes, bestandlesend...),
		dateilesend...), netzwegwerkzeug(naehte.Netzweg)...),
		append(append(regelfreigabewerkzeug(naehte.Regelfreigabe),
			schnittstellenwerkzeug(naehte.Schnittstellen)...),
			schalterwerkzeug(naehte.Schalter)...)...)

	// **Und die Paketwerkstatt** (#1164) — zwei lesend, eines schreibend. Die
	// Fachlogik einer Vorlage steht in einer Textdatei; ohne diese Werkzeuge
	// kann ein Modell sie lesen, aber nicht ändern.
	werkstattlesend, werkstattschreibend := geteilt(werkstattwerkzeuge(werkstatt))

	lesendes = append(lesendes, werkstattlesend...)

	// **Und die Sammlung** (#1165) — zwei lesend, drei schreibend. Sie hängt
	// an keinem Set: was jemand besitzt, ist ein **Druck** und kein Entwurf.
	sammlunglesend, sammlungschreibend := geteilt(sammlungswerkzeuge(sammlung))

	lesendes = append(lesendes, sammlunglesend...)

	// **Und die Preise** (#1165) — drei lesend, eines schreibend. Das
	// schreibende ist das einzige hier, das ins **Netz** geht.
	preiselesend, preiseschreibend := geteilt(preiswerkzeuge(preise))

	lesendes = append(lesendes, preiselesend...)

	// **Und der Arbeitsablauf ist seit #1291 nicht mehr nur schreibend.** MTG
	// JSON braucht die Frage, für welche Spiele es eine Zuordnung gibt — und
	// die ändert nichts. Geteilt und nicht zerschnitten: die Werkzeuge stehen
	// beieinander, weil sie einen Ablauf bilden.
	ablauflesend, ablaufschreibend := geteilt(ablauf(ausfuhr))

	lesendes = append(lesendes, ablauflesend...)

	// **Und die Paketpflege** (#1291) — eines lesend, drei schreibend. Ohne
	// sie kann ein Modell die Vorlagen nur **beziehen**, nicht einbinden, was
	// schon auf der Platte liegt.
	pflegelesend, pflegeschreibend := geteilt(paketpflege(naehte.Pflege))

	lesendes = append(lesendes, pflegelesend...)

	// **Und das Kartenbild von der Platte** (#1291). Es steht bei den
	// schreibenden: es ersetzt ein Feld, und der Verlauf hält den vorigen
	// Wert — dieselbe Zusage wie `schreiben.feld`.
	kartenbilder := kartenbildwerkzeug(naehte.Kartenbilder)

	// **Und die Listen rund um die Sammlung** — die Begründung je Gruppe
	// steht bei [sammlungslistenwerkzeuge].
	listenlesend, listenschreibend := sammlungslistenwerkzeuge(naehte)

	lesendes = append(lesendes, listenlesend...)

	// **Und die Kartenskripte** (#1337) — eines lesend, eines schreibend. Der
	// schreibende legt mehrere Karten in **einem** Zug an; ein Zurücknehmen
	// holt alle zusammen zurück.
	skriptlesend, skriptschreibend := kartenskripte(naehte.Kartenskripte)
	lesendes = append(lesendes, skriptlesend...)

	// **Und die Auszählungsdimensionen** (#1338) — beide lesend. Sie rechnen
	// über jede Karte des Sets und ändern nichts.
	lesendes = append(lesendes, auszaehlung(naehte.Auszaehlung)...)

	// **Und die Parameterhilfen der Schlüsselwörter** (#1342) — lesend. Sie
	// sagen, was sich einfügen lässt; eingefügt wird mit
	// `schreiben.schluesselwort.setzen`.
	lesendes = append(lesendes, parameterhilfen(naehte.Parameterhilfen)...)

	// **Und die Autoersetzungen** (#1341, Vermerk 0060) — eines lesend, vier
	// schreibend. Sie schreiben in den **Bestand**, nicht ins Paket.
	autolesend, autoschreibend := autoersetzungen(naehte.Autoersetzungen)

	// **Und die Decks samt Formaten** (#1293) — fünf lesend, zehn schreibend.
	// Ohne sie kann ein Modell kein Deck bauen und keines prüfen, obwohl
	// beides im Bestand steht.
	decklesend, deckschreibend := decks(naehte.Decks)
	turnierlesend, turnierschreibend := turniere(naehte.Turniere)
	nachzeichnend := symbolnachzeichnen(naehte.Symbolnachzeichnung)
	lesendes = append(lesendes, autolesend...)

	// **Und die Massenvorgänge** (#1292) — alle drei schreibend, auch der
	// Preisschub: er trägt Preisstände ein.
	massenschreibend := massenvorgaenge(naehte.Massenvorgaenge)

	// **Und die Exportvorlagen** (#1291) — eines lesend, zwei schreibend.
	// Ohne sie kann ein Modell ein Set nicht in ein fremdes Format bringen,
	// obwohl das Pack zwölf Vorlagen dafür mitbringt.
	vorlagenlesend, vorlagenschreibend := geteilt(exportvorlagen(naehte.Vorlagen))

	// **Und der Probelauf** (UXR-F7, #2015) — beide lesend, ohne zu schreiben.
	lesendes = append(append(lesendes, vorlagenlesend...), probelaufwerkzeuge(naehte.Probelauf)...)

	schreibendes := append(append(append(append(schreibende(schreiber), wortschreibend...),
		lebenszyklus(schreiber)...), ablaufschreibend...),
		scryfallSchreibend(abgleich)...)

	// **Und ein neues, leeres Set** (#891). Es steht bei den schreibenden,
	// weil es schreibt — und ohne es hat ein Modell, das gebeten wird „bau mir
	// ein Set", keinen ersten Schritt.
	schreibendes = append(append(append(append(append(schreibendes,
		setneuwerkzeug(anleger)...), angabenschreibend...),
		stilschreibend...), bestandschreibend...),
		append(append(append(pflegeschreibend, kartenbilder...),
			vorlagenschreibend...), append(listenschreibend, massenschreibend...)...)...)

	// **Und ein Feld auf vielen Karten** (#1390). Es steht bei den
	// schreibenden und trägt eine Zusage, die keine Schleife hat: alles oder
	// nichts.
	schreibendes = append(schreibendes, massenfeldwerkzeug(schreiber)...)
	schreibendes = append(schreibendes, kontaktschreibend...)
	lesendes = append(lesendes, massenvorschauwerkzeug(schreiber)...)
	lesendes = append(lesendes, verlaufswerkzeug(bestand)...)

	// **Und der Symboleditor** (#1391) — acht Rechnungen und das Lesen
	// lesend, das Ablegen schreibend. Die Rechnungen ändern nichts: Teile
	// hinein, Teile heraus.
	symbollesend, symbolschreibend := geteilt(symbolwerkzeuge(naehte.Symbolwerkstatt))

	lesendes = append(lesendes, symbollesend...)
	schreibendes = append(schreibendes, symbolschreibend...)

	// **Und das Kartenskript** (#1337).
	schreibendes = append(schreibendes, skriptschreibend...)

	// **Und die vier schreibenden der Autoersetzung** (#1341).
	schreibendes = append(schreibendes, autoschreibend...)

	// **Und die Deckwerkzeuge** (#1293).
	lesendes = append(lesendes, decklesend...)
	schreibendes = append(schreibendes, deckschreibend...)

	// **Und die Turnierwerkzeuge** (#1294).
	lesendes = append(lesendes, turnierlesend...)
	schreibendes = append(schreibendes, turnierschreibend...)
	schreibendes = append(schreibendes, nachzeichnend...)

	// **Und die Werkzeuge an ganzen Dateien** (#1169, #1457) — sie fassen als
	// einzige eine ganze Datei an. Das Zurückspielen steht ausdrücklich
	// **nicht** dabei; warum, sagt `bestandssicherung.go`.
	schreibendes = append(schreibendes, dateischreibend...)

	// **Und das Bild** (#1005). Es steht bei den schreibenden, weil es
	// schreibt — und es schreibt **zwei** Felder: das Bildfeld und den
	// Künstler. Die Nennungspflicht von Scryfall ist keine Sache der
	// Oberfläche, sondern der Karte.
	schreibendes = append(append(schreibendes, scryfallbildwerkzeug(bilder)...),
		sammelholend...)

	// **Und das Anlegen aus einem Namen** (#1032). Es steht zuletzt, weil es
	// das folgenreichste ist: es legt eine Karte an, die es vorher nicht gab.
	schreibendes = append(schreibendes, schluesselschreibend...)

	schreibendes = append(schreibendes, scryfallkartenwerkzeug(anlage)...)

	// **Und ein ganzer Satz** (#1033). Es steht ganz am Ende, weil es das
	// folgenreichste ist: ein Aufruf legt ein Set mit Hunderten Karten an.
	schreibendes = append(schreibendes, scryfallsatzwerkzeug(saetze)...)

	// **Und das Schreiben einer Paketdatei** (#1164). Es steht bei den
	// schreibenden und weit hinten, weil es die **Vorlage für alle Sets**
	// ändert, die dieses Paket benutzen — mehr Reichweite als jede Karte.
	schreibendes = append(schreibendes, werkstattschreibend...)

	// **Und die Sammlung** (#1165). Sie steht bei den schreibenden, weil sie
	// schreibt — und weit hinten, weil ein Exemplar eine Aussage über den
	// **Besitz** ist und niemand sie versehentlich treffen soll.
	schreibendes = append(append(schreibendes, sammlungschreibend...),
		preiseschreibend...)

	// **Und das Löschen eines ganzen Sets** (#1218). Es steht **nach** allem
	// anderen, weil es das einzige **unumkehrbare** Werkzeug dieser Fläche
	// ist: jedes andere schreibt etwas, das sich zurückschreiben lässt.
	schreibendes = append(schreibendes, setloeschwerkzeug(loescher)...)

	// **Und die beiden schreibenden Verweiswerkzeuge** (#1167). Sie ändern
	// keinen Feldwert — sie ändern, **welchen Druck** der Abgleich holt.
	schreibendes = append(schreibendes, verweisschreibend...)
	schreibendes = append(schreibendes, markenschreibend...)

	// **Und die beiden schreibenden Ablagewerkzeuge** (#1168).
	schreibendes = append(schreibendes, ablageschreibend...)

	// **Und das schreibende Ersetzen** (#1168).
	schreibendes = append(schreibendes, ersetzschreibend...)

	// **Und die beiden schreibenden Listenwerkzeuge** (#1168).
	schreibendes = append(schreibendes, listeschreibend...)
	schreibendes = append(schreibendes, nutzerschreibend...)
	schreibendes = append(schreibendes, tokenschreibend...)
	schreibendes = append(schreibendes, vorgabenschreibend...)
	schreibendes = append(schreibendes, freigabenschreibend...)
	schreibendes = append(schreibendes, platzschreibend...)

	// **Und die Deckübernahme** (#1351, Werkstück 12) — eines schreibend, und
	// seine lesende Hälfte steht bei den Freigaben (UX-08.AK-04, #1680).
	// **Neben den Freigaben und nicht bei den Decks**: sie verlangt dieselbe
	// Rolle wie der Wechsel und fasst zwei Profile an.
	schreibendes = append(schreibendes, uebernahmeschreibend...)

	return append(lesendes, schreibendes...)
}
