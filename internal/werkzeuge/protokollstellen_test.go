package werkzeuge

// **Wer protokollieren darf, und warum** — aus `protokollmerkmale_test.go`
// gelöst, als die Liste die Datei über die Zeilengrenze aus §2 hob (#2500).
// Die Regel steht dort; hier steht, wo sie eine Ausnahme kennt.

// protokollstellen sind die Dateien, die protokollieren dürfen —
// **jede mit ihrem Grund**.
var protokollstellen = map[string]string{
	"../../main.go":                "legt das Protokoll an und meldet den Bestand",
	"../app/protokoll.go":          "ist das Protokoll",
	"../app/protokoll_darwin.go":   "die Senke ins Unified Log",
	"../app/sicherungsplanlauf.go": "eine Sicherung nach Plan — die Datei und, wie viele eigene dabei gingen (OPS01)",
	"../app/restepflege.go":        "räumt beim Start liegengebliebene eigene Reste — wie viele, und warum keiner ging (OPS04, #2350)",
	"../app/betriebsband.go":       "zählt einen Webprozess, den WebKit beendet hat und Wails neu lud — sonst sagt es niemand (OPS02, #2348)",
	"../app/protokoll_andere.go":   "dieselbe Senke, wo es keines gibt",
	"../app/kindprotokoll.go":      "die Zeilen eines vom Fenster gestarteten Dienstes (Nachlauf über Phase 6.4, N3)",
	"../app/dienstwerk.go":         "setzt den Empfänger der Zeilen eines Dienstes, den es startet",
	"../app/protokollmerkmale.go":  "die Liste der erlaubten Merkmale",
	"../app/fensterfehler.go":      "schreibt einen Fehler, den das Fenster beim Zeichnen gefangen hat, mit der Kennung, die es anzeigt — sonst stünde sie nur dort (F01, #2340)",
	"../app/sitzung.go":            "meldet, wenn die festgehaltene Sitzung beim Start nicht aufgeht (ID-02, #1701) — ohne Geheimnis und ohne Nutzerkennung",
	"../app/protokollweiche.go": "die Weiche zwischen Schreibern und Kanälen (OPS-04, #1721) — " +
		"sie ist das Protokoll, nur nach den Einstellungen umschaltbar",
	"../app/protokolleinrichtung.go": "schreibt auf, dass die Einstellungen des Protokolls " +
		"sich nicht lesen ließen (#1721) — dann gilt die Vorgabe, und nur das Protokoll sagt es",
	"../fenster/abnahme_darwin.go": "die Abnahmeleitung, nur im Abnahmebau (#1727): " +
		"öffnet sie sich nicht, gibt es kein Fenster, das es sagen könnte — die " +
		"Probe bekommt keine Antwort, und der Grund steht hier",
	"../fenster/abnahme_hand_darwin.go": "dieselbe Leitung: eine Webansicht, die sich " +
		"für die Abnahme nicht umstellen lässt, sagt es hier und nicht am Fenster",
	"../vorschauweg/schutzweiche.go": "nimmt die Verstoßberichte der Schutzrichtlinie an und " +
		"schreibt sie — mit Anweisung, Ursprung und Stelle, ohne Pfad (SEC-17, #2164)",
	"../fenster/schutznaht.go": "reicht das Protokoll an die Schutzweiche weiter; selbst " +
		"schreibt die Hülle keine Zeile (SEC-17)",
	"../fenster/schutznaht_server.go": "reicht das Protokoll im Servermodus nur durch — " +
		"dort trägt die Seite keine Schutzrichtlinie (SEC-17)",
	"../fenster/probepanik_abnahme_darwin.go": "nur im Abnahmebau: meldet, dass ein synchroner Ruf " +
		"nach einer eingespeisten Panik zurückkam — die Zeile, die `syncpanik09` liest (#2398)",
	"../fenster/riegel.go": "schreibt eine Panik in einem Rückruf der Hülle und was Wails selbst " +
		"fängt, je mit Vorgangskennung — sonst endete sie still oder in os.Exit (F02, #2341)",
	"../fenster/wails.go": "reicht das Protokoll der Anwendung an Wails weiter (ENG-02) " +
		"— sonst schriebe Wails 3 an ihm vorbei auf die Fehlerausgabe, die bei " +
		"einer gestarteten `.app` niemand sieht; selbst schreibt die Hülle keine Zeile",
	"../app/nutzerbilder.go": "ein Bild, das nach einem Wechsel nicht mehr " +
		"gezeigt wird und sich nicht wegräumen lässt: der Wechsel ist " +
		"geschrieben (§7, Regel 3), eine Meldung am Fenster widerspräche ihm " +
		"— die Datei ohne Verweis steht im Protokoll (Zweitblick über Phase " +
		"6.1, Befund 3)",
	"../app/befehl.go": "eine Quelle der Befehle, die scheitert, hält den Start nicht auf — die Zentrale sagt, dass sie leer ist, und der Grund steht hier",
	"../app/nebenlaufzusammenbruch.go": "eine hereinkommende Verbindung, die nicht " +
		"zustande kommt, ist **kein Ereignis für einen Menschen** (N2, " +
		"#1455): sie kann von jedem ausgelöst werden, der die Adresse " +
		"kennt, und wer daraus eine Meldung am Fenster machte, hätte eine " +
		"Fläche, die ein Fremder bedient. Sie steht deshalb im Protokoll — " +
		"dort sucht sie, wer wissen will, warum ein Gegenüber nicht " +
		"durchkommt. **Und seit BK-05a** (#2111) der Grund eines Nebenlaufs von " +
		"Partie und Peer, der zusammenbrach — der Tisch pausiert, der Horchposten " +
		"steht still, und nur das Protokoll sagt, warum",
	"../app/partiewege.go": "ein Zug erreicht den ganzen Tisch (#1627), und " +
		"eine Leitung dorthin kann zugegangen sein. **Der Zug ist dann " +
		"trotzdem geschehen** (§7, Regel 3): er steht im Bestand, und der " +
		"Absender bekommt seine Antwort. Was übrig bleibt, ist eine " +
		"Auskunft für den, der hinterher fragt, warum ein Gegenüber den " +
		"Tisch anders sieht — und **kein Ereignis für einen Menschen**: " +
		"auslösen kann es jeder, dessen Leitung gerade endet",
	"../app/partiebeitritt.go": "der Tisch eines Wirts läuft, **nachdem** der " +
		"Ruf vorbei ist, der ihn betreten hat (G5, #1462) — dieselbe Lage " +
		"wie bei der Aufnahme eine Zeile weiter, nur andersherum. Und was " +
		"ein Wirt schickt, ist **kein Ereignis für einen Menschen**: eine " +
		"abgewiesene Nachricht kann jeder auslösen, an dessen Tisch man " +
		"sitzt",
	"../app/partienennung.go": "ein Gast wartet auf das erste Wort des " +
		"Wirts, **nachdem** der Ruf vorbei ist, der ihn hingesetzt hat " +
		"(UX-08.AK-04, #1680) — dieselbe Lage wie beim Tisch eine Zeile " +
		"weiter. Dass keines kam, ist **kein Ereignis für einen Menschen**: " +
		"das Fenster sagt es dort, wo der Tisch fehlt, und die Leitung " +
		"endet wie bei jedem Abriss",
	"../app/schalter.go": "ein Schalter ist geschrieben, und sein Stand ließ sich " +
		"danach nicht lesen (OPS-10b, #1978): die Läufe enden trotzdem mit dem " +
		"geschriebenen Wert (§7, Regel 3), und nur das Protokoll sagt, warum " +
		"die übrigen Schalter dabei als an galten",
	"../app/partieaufnahme.go": "eine Gastsitzung läuft, **nachdem** der Ruf " +
		"vorbei ist, der sie angenommen hat (G5, #1462) — es gibt keine " +
		"Antwort mehr, an die sich etwas hängen ließe. Und was ein Gast " +
		"schickt, ist **kein Ereignis für einen Menschen**: eine abgewiesene " +
		"Nachricht kann jeder auslösen, der eine Sitzung hat, und wer daraus " +
		"eine Meldung am Fenster machte, hätte eine Fläche, die ein Fremder " +
		"bedient. Dieselbe Lage wie beim Horchposten, und dieselbe Antwort: " +
		"ins Protokoll",
	"../app/einloesung.go": "eine Einlösung läuft in der Sitzung, die sie " +
		"brachte, und die Meldung beim Einladenden nach einer Annahme, die " +
		"schon geschrieben ist (§7, Regel 3) — der Satz geht an die Fläche, " +
		"der Grund bleibt hier (§8, NET-01b, #1954)",
	"../app/handelsaufnahme.go": "eine Handelssitzung läuft, **nachdem** " +
		"der Ruf vorbei ist, der sie angenommen hat (#1634) — dieselbe Lage " +
		"wie bei der Partieaufnahme darüber. Und was ein Gegenüber schickt, " +
		"ist **kein Ereignis für einen Menschen**: eine abgewiesene " +
		"Handelsnachricht kann jeder auslösen, dem der Handel freigegeben " +
		"ist, und ein Horcher hat ohnehin kein Fenster, an dem etwas " +
		"erschiene",

	"../app/dienstposten.go": "ein Dienst, der nach dem Starten stirbt " +
		"oder dessen Erreichbarkeit sich nicht rechnen lässt, meldet sich " +
		"**nach** dem Ruf, der ihn gestartet hat (N6, #1572) — es gibt " +
		"keine Antwort mehr, an die es sich hängen ließe. Dieselbe Lage " +
		"wie beim Horchposten darüber, und dieselbe Antwort: ins Protokoll",

	"../service/handelsgegenbeleg.go": "ein Gegenbeleg, der nicht passt, wird " +
		"am Beleg festgehalten und **nicht** als Fehler gemeldet (#2500, " +
		"Vermerk 0202) — sonst endete die Sitzung. Er kommt über eine " +
		"Leitung, an der kein Mensch steht; hier steht, dass es geschah, " +
		"und welcher Mangel, nicht der Satz",
	"../service/aktualisierungsvorpruefung.go": "eine Sicherungsprobe, die " +
		"scheitert, ist ein **Hindernis** und kein Fehler des Wegs (O6, " +
		"#1430): die Vorprüfung soll sagen, was im Weg steht, statt " +
		"abzubrechen. Der technische Grund erreicht keine Oberfläche (§8) und " +
		"steht deshalb im Protokoll — unter der Vorgangskennung dieses Rufs",
	"../service/rueckspielabschluss.go": "der Vermerk über eine Rückkehr hält " +
		"den Weg nicht auf (§7, Regel 3) — getauscht ist getauscht. Was dabei " +
		"schiefgeht, steht im Protokoll; die Marke daneben bleibt liegen, und " +
		"der nächste Start trägt es nach (Befund 10 des Zweitblicks über " +
		"Phase 5.2)",
	"../service/epochenmarke.go": "eine liegengebliebene Rückkehr wird beim " +
		"Start nachgetragen (Befund 10 des Zweitblicks über Phase 5.2). Was " +
		"dabei schiefgeht, hält den Start nicht auf — die Anwendung ist " +
		"benutzbar, und ein Programm, das deswegen nicht startet, nimmt den " +
		"Nutzenden auch den Weg, es von Hand zu richten. Der Satz fürs " +
		"Fenster kommt als Rückgabewert; der technische Grund steht im " +
		"Protokoll",
	"../service/ruecksprung.go": "ein Vergleich vor der Rückkehr, der nicht zu " +
		"bekommen ist, hält den Tausch nicht auf (O7, #1431) — er wird zur " +
		"**Handprüfung**, und die Epoche sagt das. Der technische Grund " +
		"erreicht keine Oberfläche (§8) und steht deshalb im Protokoll, unter " +
		"der Vorgangskennung dieses Rufs",
	"../app/weiterarbeiten.go": "ein Vermerk für die Übersicht, der nicht gelingt, hält den Weg nicht " +
		"an (#1422, §7 Regel 3) — er steht im Protokoll",
	"../app/kibudget.go": "der Kostenposten eines KI-Auftrags wird " +
		"geschlossen, nachdem der Anbieter geantwortet hat (#1542). " +
		"Misslingt das, ist der Auftrag trotzdem geschehen (§7, Regel 3): " +
		"wer die Meldung an die Antwort hängte, machte aus einem geglückten " +
		"Vorschlag einen gescheiterten — wegen einer Buchung, die niemand " +
		"mehr rückgängig machen kann. Der Grund steht im Protokoll",
	"../app/fehler.go": "schreibt den rohen Fehler auf, den die Oberfläche nicht bekommt (#857) — " +
		"und seit Werkstück 1 von Phase 4.13 auch den **verdeckten Grund** eines " +
		"eingeordneten Fehlers, den `domain.MitGrund` aus dem Satz genommen hat (#1352)",
	"../app/start.go": "das Paketverzeichnis, dessen Fehler den Start nicht aufhält (#857) — " +
		"aus `app.go` mitgezogen, als der Start an der Zeilengrenze aus §2 abgetrennt wurde",
	"../app/ablagennaht.go": "das Bilderverzeichnis, dessen Fehler den Start nicht aufhält (#857) — mit `app.go` umgezogen, als es die Symbolablage dazubekam (#890)",
	"../app/peersitzungnaht.go": "eine Peerleitung, die sich beim " +
		"Niederlegen beschwert — der Profilwechsel findet trotzdem statt, " +
		"und die Klage wäre sonst fort (N1, #1454)",
	"../service/sicherungspakete.go": "ein eigenes Paket, das sich nach dem Tausch nicht " +
		"einsetzen ließ — der Bestand ist dann schon zurückgespielt (OPS-06)",
	"../service/bestandsrueckspielen.go": "schreibt auf, wo der Rückweg nach einem gescheiterten " +
		"Zurückspielen endet — beide Stände liegen dann als Datei da, und nur das Protokoll sagt es (#940)",
	"../service/rueckspielmarke.go": "schreibt auf, dass ein abgebrochenes Zurückspielen beim Start " +
		"vollendet wurde — es geschieht ohne Rückfrage, und nur das Protokoll sagt, dass es geschah (#1162)",
	"../service/paketeinlesen.go": "schreibt auf, warum ein Paket aus einem Installer draußen blieb " +
		"(FMT-02.AK-04, #1691) — der Befund trägt nur den Satz, der Grund steht hier",
	"../service/paketbezug.go": "schreibt auf, dass der bisherige Paketstand liegen blieb — " +
		"der Bezug ist geglückt und wird deswegen nicht verworfen (§7, Regel 3), aber ein " +
		".vorher von über einem Gigabyte darf nicht still den Platz verbrauchen (#1170)",
	"../mcp/rufvollzug.go": "schreibt den Grund eines Werkzeugs auf, das zusammenbrach (BK-05b-MCP, " +
		"#2113) — das Gegenüber bekommt nur „zusammengebrochen“, denn der Grund kann sagen, " +
		"was im Bestand steht; der Prozess schreibt über den schwärzenden Schreiber (`app.Serverprotokoll`)",
	"../app/lebenszyklus.go": "schreibt auf, was ein Schluss in seiner Frist nicht schaffte (BK-07a, " +
		"BK-02, #2115, #2110) — das Fenster ist dann zu, und nur das Protokoll sagt, dass die " +
		"Fenstergröße verloren ging oder ein Lauf den Bestand geschlossen vorfand",
	"../app/dienstnachdemaufstieg.go": "der Neustart des Dienstes vor der Wanderung, geglückt oder nicht — " +
		"er bricht laufende Anfragen ab, und sonst sagt es niemand (M02, #2346)",
	"../app/fassungswache.go": "ein Prozess endet, weil eine neuere Fassung gewandert hat — die Antwort " +
		"bekommt nur der eine Rufer, das Protokoll erklärt das Ende (M02, #2346)",
	"../app/raeumschritte.go": "schreibt auf, welcher Schritt des Räumens in Panik endete — die " +
		"übrigen laufen weiter (Nachlauf über Phase 1 von 0.9.0, Befund 2)",
	"../app/profilnaht.go": "schreibt auf, dass der Profilzeiger sich nicht lesen ließ (#1351) — " +
		"der Start läuft weiter, weil die erste Abfrage es deutlicher sagt als eine Meldung " +
		"beim Anlaufen, und dann ist das Protokoll die einzige Spur",
	"../app/scryfallstart.go": "die Scryfall-Zugänge, deren Fehler den Start nicht aufhalten (#857) — " +
		"aus `start.go` mitgezogen, als die Datei mit der Kartensuche an die Zeilengrenze kam (#1167)",
	"../app/nurlesen.go": "schreibt auf, dass die nur lesende Verbindung nicht " +
		"zustande kam (#1351, Werkstück 16) — der Start läuft weiter, weil eine " +
		"Anwendung, die deswegen gar nicht aufgeht, einen kleinen Mangel gegen " +
		"einen großen tauscht (§8). **Was das kostet, gehört gesagt**: dann wird " +
		"geschrieben, wo vielleicht nicht geschrieben werden sollte — und nur " +
		"das Protokoll sagt es",
	"../app/verdrahtung.go": "das Bilderverzeichnis und die Kartenbilder, deren " +
		"Fehler den Start nicht aufhalten (#857) — mit der Verdrahtung aus " +
		"`start.go` umgezogen, als sie zweimal laufen können musste (#1351, " +
		"Werkstück 16)",
	"../app/wiederanlauf.go": "meldet beim Start eine Sitzung, die nicht sauber endete (F04, #2343) — " +
		"das Fenster sagt es auch, das Protokoll hält es für das Diagnosepaket fest",
	"../app/deckstart.go": "die Formatprüfung, deren Fehler den Start nicht aufhält (#1293) — " +
		"sie braucht Scryfall, und wer kein Zuhause hat, bekommt keinen Klienten. Das Fenster " +
		"geht trotzdem auf, und die Meldung kommt beim ersten Versuch",
}
