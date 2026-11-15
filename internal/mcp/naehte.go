// Die **Nähte** des MCP-Servers — die Anwendungsfälle, die er anbietet.
//
// **Eine eigene Datei**, seit `werkzeugliste.go` mit der Autoersetzung über
// die Zeilengrenze aus §2 ging
// (#1341).
// Der Schnitt folgt der Frage: dort steht, **welche** Werkzeuge es gibt und in
// welcher Reihenfolge, hier, **woran** sie hängen.

package mcp

// Naehte bündelt die Anwendungsfälle, die der Server anbietet.
//
// **Als Verbund und nicht als fünfzehn Parameter**
// (#1107).
// Die Liste war auf fünfzehn Stellen gewachsen, und in ihr übersetzte ein
// vertauschtes Paar noch: die meisten Doubles erfüllen mehrere dieser
// Schnittstellen. Ein benanntes Feld lässt sich nicht vertauschen.
type Naehte struct {
	Leser     Leser
	Schreiber Schreiber
	Ausfuhr   Ausfuhr
	Abgleich  Fremdabgleich
	Bilder    Fremdbilder
	Auskunft  Fremdauskuenfte

	// Suche ist die Kartensuche bei Scryfall (#1167).
	//
	// **`nil` ist erlaubt** — dann gibt es das Werkzeug nicht, so wie bei
	// jeder anderen Naht, die eine Fläche nicht einrichtet.
	Suche Fremdkartensuche

	// Katalog ist die Auskunft über jede Fähigkeit der fremden Dienste —
	// Werkstück A1 von Phase 5.3 (#1440). `nil` ist erlaubt wie bei der Suche.
	Katalog Fremdkatalog

	// Grenzen nennt, was dieser Prozess an Rufen und Parallelität zulässt —
	// Werkstück O5 von Phase 5.2 (#1429). `nil` ist erlaubt.
	Grenzen Grenzenauskunft

	// Peer nennt die eigene Identität — Werkstück N1 von Phase 5.4 (#1454).
	//
	// **Nur lesend**: anlegen und ablösen sind Handlungen für einen
	// Menschen, und eine Rotation entwertet jede Bestätigung, die ein
	// Kontakt gegeben hat.
	Peer Peerauskunft

	// Kontakte nennt die bekannten Gegenüber und lässt sie sperren —
	// Werkstück N1 von Phase 5.4 (#1454).
	//
	// **Sperren ja, freigeben nein**: der Grund steht im Kopf von
	// `peerkontakt.go`.
	Kontakte Kontaktauskunft

	// Gesundheit sagt, ob dieser Prozess lebt und arbeiten kann — Werkstück O9
	// von Phase 5.2 (#1433). `nil` ist erlaubt.
	Gesundheit Gesundheitsauskunft

	// Metriken nennt, was dieser Prozess gewirkt hat — mit endlich vielen
	// Reihen (O9, #1433).
	Metriken Metrikauskunft

	// Meta sagt, welches Programm antwortet (O9, #1433).
	Meta Metaauskunft

	// Vertraeglichkeit sagt, ob die neueste Freigabe diesen Bestand öffnen
	// kann (O6, #1430).
	Vertraeglichkeit Vertraeglichkeitsauskunft

	// Ruecksprung sagt, was zurückgehen kann und was eine Rückkehr kostet —
	// Werkstück O7 von Phase 5.2
	// (#1431).
	Ruecksprung Ruecksprungauskunft

	// Datenpflege sagt, was im Bestand herumliegt — Werkstück O8 von Phase 5.2
	// (#1432).
	//
	// **Nicht `Pflege`**: den Namen hat die Paketpflege weiter unten. Zwei
	// Dinge, die beide „Pflege" heißen, sind eine Verwechslung, die man erst
	// bemerkt, wenn eine falsch verdrahtet ist.
	Datenpflege Pflegeauskunft

	// Ruecksetzung sagt, was eine Rücksetzung anfassen würde — Werkstück O4
	// von Phase 5.2
	// (#1428).
	//
	// **Nur die Vorschau**, und der Grund steht im Kopf von
	// `ruecksetzung.go`.
	Ruecksetzung Ruecksetzungsauskunft

	// Kosten sagt, wie viel an einem Anbieterkonto freigegeben ist —
	// Werkstück O12 von Phase 5.2
	// (#1439).
	//
	// **Nur lesend**, und der Grund steht im Kopf von `kostenjournal.go`.
	Kosten Kostenauskunft

	// Index ist der lokale Suchindex — Werkstück A15 von Phase 5.3
	// (#1514).
	Index Indexauskunft

	// Recherche fragt den Wissensbestand — Werkstück A16 derselben Phase
	// (#1515).
	//
	// **Das Kontextpaket steht nicht darin**, und der Grund steht im Kopf
	// von `recherche.go`: hier ist der Aufrufer das Modell.
	Recherche Rechercheauskunft

	// Verweise bezieht entworfene Karten auf echte Drucke (#1167).
	Verweise Fremdverweise

	// Marken hängt Wörter an Karten
	// (#1208),
	// Objektmarken an jedes andere Fachobjekt (UX-12, #1684).
	Marken       Fremdmarken
	Objektmarken Objektmarken

	// Ablage kopiert Karten zwischen Sets (#1168).
	Ablage Fremdablage

	// Ersetzung sucht und ersetzt über ein Set (#1168).
	Ersetzung Fremdersetzung

	// Liste gibt die Spalten der Kartenliste und ihre Ordnung (#1168).
	Liste Fremdliste

	// Nutzer führt die Nutzer dieses Bestands und ihre Profile
	// (#1351).
	Nutzer Nutzerpflege

	// Token führt die Zugriffstoken technischer Profile — Werkstück 10
	// desselben Issues.
	Token Tokenpflege

	// Vorgaben führt die Profilvorgaben, die auf **Zeilen** zeigen —
	// Werkstück 15 desselben Issues.
	Vorgaben Vorgabenpflege

	// Freigaben sagt, wer in welches fremde Profil darf — Werkstück 11
	// desselben Issues.
	Freigaben Freigabenpflege

	// Links sind die Wege nach draußen zu einer Karte — Werkstück T8 von
	// Phase 5.6
	// (#1473).
	Links Linkauskunft

	// Konto ist das eigene Konto bei einem Marktplatz — Werkstück T11 von
	// Phase 5.6 (#1476).
	Konto Kontoauskunft

	// Handel ist der Stand eines Handelsvorgangs — Werkstück T1 von Phase
	// 5.6 (#1466),
	// angeschlossen mit
	// #1637.
	// **Lesend und sonst nichts** — der Grund steht in `handelsstand.go`.
	Handel Handelsauskunft

	// Platz sagt, wie viel Platz was belegt — Werkstück 4 von Phase 4.13
	// (#1352).
	//
	// **Sie hängt an keiner Verbindung**: gezählt werden Dateien, und die
	// liegen auch dann da, wenn der Bestand nicht aufgeht.
	Platz Platzauskunft

	// Probleme prüft die Karten eines Sets und nennt den Stand (PERF02,
	// #2352).
	Probleme Problemquelle

	// Deckuebernahme holt ein Deck aus einem fremden Profil — Werkstück 12
	// desselben Issues.
	Deckuebernahme Deckuebernahmepflege

	Sammeldaten Fremdsammeldaten
	Anlage      Fremdanlage
	Saetze      Fremdsaetze
	Konsole     Skriptkonsole
	Wortlisten  Wortlisten
	Woerter     Schluesselwoerter
	Pack        Pack
	Anleger     Setanleger
	Loescher    Setloescher
	Angaben     Setangabenquelle

	// **Und die Stile und der Bestand** — seit
	// #1107.
	Stile   Stilquelle
	Bestand Bestandsquelle

	// Zusatz liest und setzt die Zusatzfelder einer Karte (#2298).
	//
	// **`nil` ist erlaubt** — dann gibt es die Werkzeuge nicht.
	Zusatz Zusatzquelle

	// Kartenbilder setzt ein Bild von der Platte in ein Kartenfeld
	// (#1291).
	//
	// **`nil` ist erlaubt** — dann gibt es das Werkzeug nicht.
	Kartenbilder Kartenbilder

	// Listeneinfuhr liest eine Kartenliste in die Sammlung (#1292).
	Listeneinfuhr Listeneinfuhr

	// Deckeinfuhr liest eine Deckliste in ein **Deck** — Werkstück T8 von
	// Phase 5.6 (#1473).
	Deckeinfuhr Deckeinfuhr

	// Sammlungsausfuhr gibt die Sammlung aus und liest ein CSV zurück (#1292).
	Sammlungsausfuhr Sammlungsausfuhr

	// Sammlungssichten beantwortet Besitz, Lagerorte, Dubletten und Wünsche
	// (#1292, Vermerk 0057).
	Sammlungssichten Sammlungssichten

	// Massenvorgaenge ändert und löscht über eine Auswahl und holt Preise in
	// Schüben (#1292).
	Massenvorgaenge Massenvorgaenge

	// Kartenskripte nennt und führt die `add cards script:`-Blöcke des
	// Spielpakets (#1337).
	Kartenskripte Kartenskripte

	// Auszaehlung zählt ein Set nach den `statistics dimension:`-Blöcken
	// seines Spielpakets aus (#1338).
	Auszaehlung Auszaehlung

	// Parameterhilfen nennt beide Menüs zu einem Schlüsselwort (#1342).
	Parameterhilfen Parameterhilfen

	// Autoersetzungen zeigt und pflegt die `auto replace:`-Regeln eines
	// Spiels (#1341, Vermerk 0060).
	Autoersetzungen Autoersetzungen

	// Decks bedient Decks und Formate (#1293, Phase 4.6).
	Decks Decks

	// Turniere führt ein Turnier über die Zeit (#1294, Phase 4.7).
	Turniere Turniere

	// Symbolbibliothek führt die Symbole eines Profils (DOM-02, #1697).
	Symbolbibliothek Symbolbibliothek

	// Sprachpakete führt die eigenen Übersetzungen (LOC-02, #1688).
	Sprachpakete Sprachpakete

	// Symbolnachzeichnung macht aus einer Bilddatei das Symbol eines Sets
	// (#1340).
	Symbolnachzeichnung Symbolnachzeichnung

	// Vorlagen führt die Exportvorlagen aus
	// (#1291).
	//
	// **`nil` ist erlaubt** — dann gibt es die drei Werkzeuge nicht.
	Vorlagen Vorlagenausgabe

	// Probelauf rechnet den Druckbogen und die Exportvorlage durch, ohne zu
	// schreiben (UXR-F7, #2015).
	//
	// **`nil` ist erlaubt** — dann gibt es die zwei Werkzeuge nicht.
	Probelauf Probelauf

	// Pflege bindet Pakete ein, nimmt sie heraus und übernimmt aus MSE2
	// (#1291).
	//
	// **`nil` ist erlaubt** — dann gibt es die vier Werkzeuge nicht.
	Pflege Paketpflege

	// **Und die Paketwerkstatt** — seit
	// #1164.
	Werkstatt Paketwerkstatt

	// **Und der Symboleditor** — seit
	// #1391.
	// Neun Fähigkeiten, die bis dahin nur das Fenster hatte, obwohl ihre
	// Rechnungen reine Funktionen sind.
	Symbolwerkstatt Symbolwerkstatt

	// **Und die Sammlung samt Preisen** — seit
	// #1165.
	Sammlung Sammlungswerkstatt
	Preise   Preiswerkstatt

	// **Und die Bestandssicherung** — seit
	// #1169.
	//
	// **`nil` ist erlaubt**, und dann gibt es die beiden Werkzeuge nicht — so
	// wie bei jeder anderen Naht, die eine Fläche nicht einrichtet.
	Sicherung Bestandssicherung

	// Dateien zeigt, was in einer Datei liegt — UX-13 (#1685). **`nil` ist
	// erlaubt**, dann gibt es das Werkzeug nicht.
	Dateien Dateiansicht

	// Setvorschau sagt, was das Einfahren einer Datei täte — FMT-03 (#1692).
	Setvorschau Setvorschau

	// **Und der portable Umschlag** — seit
	// #1457.
	//
	// **`nil` ist erlaubt**, und dann gibt es die drei Werkzeuge nicht — so
	// wie bei jeder anderen Naht, die eine Fläche nicht einrichtet.
	Umschlag Umschlagwerkstatt

	// **Und die Partien** — seit Werkstück G1 von Phase 5.5
	// (#1458).
	//
	// **`nil` ist erlaubt**, und dann gibt es die vier Werkzeuge nicht.
	Partien Partienwerkstatt

	// **Und der Verlauf einer Partie** — seit Werkstück G5 von Phase 5.5
	// (#1462).
	//
	// **Eine eigene Naht und kein fünftes Feld an [Partienwerkstatt]**: der
	// Verlauf beantwortet, was eine Partie **war**, jene, was sie **tut** —
	// und wer beides zusammenlegte, zwänge jede Fläche, die nur eines von
	// beidem bietet, zu einer halben Umsetzung.
	Partienverlauf Verlaufswerkstatt

	// **Und der Netzweg** — seit
	// #1571.
	Netzweg Netzweg

	// Schnittstellen sagt, wie Kommandozeile und Werkzeugserver
	// eingerichtet sind (OPS-09, #1907) — nur lesend.
	Schnittstellen Schnittstellenauskunft

	// Schalter sagt, was abgeschaltet ist (OPS-10, #1908) — nur lesend.
	Schalter Schalterauskunft

	// **Und die Regelfreigabe** — seit Werkstück G4 von Phase 5.5
	// (#1461).
	//
	// **`nil` ist erlaubt**, und dann gibt es das Werkzeug nicht.
	Regelfreigabe Regelquellenleser
}
