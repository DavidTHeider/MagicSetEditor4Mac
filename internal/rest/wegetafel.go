package rest

// **Die Wegetafel des Dienstes** — abgetrennt am 2026-08-26, weil
// `endpunkte.go` über die Zeilengrenze aus §2 ging.
//
// # Der Schnitt folgt der Frage
//
// Hier steht **welcher Weg wohin führt**, daneben in `endpunkte.go`, **was
// dort geschieht**. Die Tafel wächst mit jeder Fähigkeit; die Behandler
// wachsen mit dem, was eine Fähigkeit an Übersetzung braucht.
//
// **Und sie ist der Ort, an dem §16 nachzusehen ist**: was hier fehlt, fehlt
// einer Fläche allein.

import (
	"net/http"
)

// wegeEintragen legt die Endpunkte an.
//
// **Acht, und es sind dieselben acht wie auf der Kommandozeile.** Das ist
// keine Bequemlichkeit, sondern §16: dieselbe Fähigkeit, überall dieselbe
// Antwort. Was hier fehlte, fehlte einer Fläche allein — und genau das soll
// die Paritätsmatrix aus Phase T verhindern.
//
// **Die Methode gehört zum Muster** (`GET /sets`), seit Go 1.22. Ein Aufruf
// mit falscher Methode bekommt damit 405 vom Weg selbst und nicht von uns.
func (d *Dienst) wegeEintragen(weg *http.ServeMux) {
	weg.HandleFunc("GET /sets", d.sets)
	// **Die Wege über den Bestand selbst** — nebenan, der Grund steht dort.
	d.betriebswegeEintragen(weg)

	weg.HandleFunc("DELETE /sets/{setID}", d.setLoeschen)
	weg.HandleFunc("GET /sets/{setID}/karten", d.karten)
	weg.HandleFunc("PUT /sets/{setID}/karten/{karteID}/felder/{feld}", d.feld)

	// **Der Lebenszyklus einer Karte und der Volltext**
	// (#1107).
	//
	// **`/tausch` steht vor `{karteID}`** — nicht in der Reihenfolge hier,
	// sondern in Gos Musterordnung: ein wörtlicher Abschnitt schlägt einen
	// Platzhalter, und deshalb wird `tausch` nie als Kartenkennung gelesen.
	weg.HandleFunc("POST /sets/{setID}/karten", d.karteAnlegen)
	weg.HandleFunc("POST /sets/{setID}/karten/tausch", d.kartenTauschen)
	weg.HandleFunc("POST /sets/{setID}/felder", d.felderSetzen)
	weg.HandleFunc("POST /sets/{setID}/felder/vorschau", d.felderVorschau)
	weg.HandleFunc("POST /sets/{setID}/karten/{karteID}/kopie", d.karteKopieren)
	weg.HandleFunc("DELETE /sets/{setID}/karten/{karteID}", d.karteLoeschen)
	weg.HandleFunc("GET /sets/{setID}/felder", d.felder)
	weg.HandleFunc("GET /suche", d.suche)

	// **Die Setangaben, die Vorbehalte und die mitgelieferten Sets**
	// (#1107).
	weg.HandleFunc("GET /sets/{setID}/angaben", d.angaben)
	weg.HandleFunc("PUT /sets/{setID}/angaben/{feld}", d.angabeSetzen)
	weg.HandleFunc("GET /sets/{setID}/vorbehalte", d.vorbehalte)
	weg.HandleFunc("GET /sets/{setID}/probleme", d.problemstand)
	weg.HandleFunc("POST /sets/{setID}/probleme", d.problemePruefen)
	weg.HandleFunc("GET /beispielsets", d.beispielsets)

	// **Stile, Pakete, Kennzahlen, Verlauf und Einstellungen**
	// (#1107).
	weg.HandleFunc("GET /sets/{setID}/stile", d.stile)
	weg.HandleFunc("GET /sets/{setID}/stile/{paket}/folgen", d.stilfolgen)
	weg.HandleFunc("PUT /sets/{setID}/stil", d.stilWechseln)
	weg.HandleFunc("GET /sets/{setID}/stiloptionen", d.stiloptionen)
	weg.HandleFunc("PUT /sets/{setID}/stiloptionen/{feld}", d.stiloptionSetzen)
	weg.HandleFunc("DELETE /sets/{setID}/stiloptionen/{feld}", d.stiloptionZuruecksetzen)
	weg.HandleFunc("GET /sets/{setID}/karten/{karteID}/zusatzfelder", d.zusatzfelder)
	weg.HandleFunc("PUT /sets/{setID}/karten/{karteID}/zusatzfelder/{feld}", d.zusatzwertSetzen)
	weg.HandleFunc("GET /pakete", d.pakete)
	weg.HandleFunc("POST /pakete", d.paketBeziehen)

	// **Die Paketpflege** (#1291) — einbinden, übernehmen, entfernen. Jede
	// eigene Adresse, weil `POST /pakete` **bezieht** und zwei Fähigkeiten
	// hinter einer Adresse niemand lesen kann (#891).
	weg.HandleFunc("PUT /pakete/eingebunden", d.paketInstallieren)
	weg.HandleFunc("POST /pakete/neu", d.paketAnlegen)

	// **Plan, Ersetzen und Zurück** — DOM-01 (#1696).
	weg.HandleFunc("GET /pakete/plan", d.paketPlan)
	weg.HandleFunc("PUT /pakete/ersetzt", d.paketErsetzen)
	weg.HandleFunc("POST /pakete/{art}/{name}/zurueck", d.paketZurueck)
	weg.HandleFunc("POST /pakete/{art}/{name}/kopie", d.paketKopieren) // FMT-02 (#1691)

	// **Die Stände der bezogenen Bäume** — G04 (#2251). **Unter eigenem
	// Pfad**: unter `/pakete/` träfe `/pakete/bezuege/dateien/vorschau` auch
	// `/pakete/{paket}/dateien/{datei...}`, und der Wegweiser weist zwei
	// gleich genaue Muster ab.
	weg.HandleFunc("GET /paketbezuege", d.paketBezuege)
	weg.HandleFunc("GET /paketbezuege/{name}/vorschau", d.paketBezugVorschau)
	weg.HandleFunc("POST /paketbezuege/{name}/aktualisierung", d.paketBezugAktualisieren)
	weg.HandleFunc("POST /paketbezuege/{name}/zurueck", d.paketBezugZurueck)
	d.weitergabewege(weg)
	weg.HandleFunc("GET /sets/{setID}/paketbedarf", d.paketBedarf)

	// **Nutzer, Profile und was an ihnen hängt** stehen in `wegeleute.go` —
	// abgetrennt am 2026-09-08, als `GET /kontext` diese Datei über 400 Zeilen
	// brachte (§2). Die Wächter lesen beide Dateien.
	d.leutewegeEintragen(weg)

	// **Die Übernahme unter `/decks`** — dort entsteht sie (#1351,
	// Werkstück 12).
	weg.HandleFunc("POST /decks/uebernehmen", d.deckUebernehmen)
	weg.HandleFunc("GET /profile/{profilID}/decks", d.deckwahl)

	// **Der Platz** — Werkstück 4 von Phase 4.13 (#1352). Er hängt an keinem
	// Profil: gezählt werden Dateien neben dem Bestand.
	weg.HandleFunc("GET /platz", d.platz)
	weg.HandleFunc("DELETE /platz/{posten}", d.platzFreigeben)

	weg.HandleFunc("GET /profile/{profilID}/freigaben", d.freigaben)
	weg.HandleFunc("PUT /profile/{profilID}/freigaben/{nutzerID}",
		d.freigabeSetzen)
	weg.HandleFunc("DELETE /profile/{profilID}/freigaben/{nutzerID}",
		d.freigabeLoeschen)

	weg.HandleFunc("PUT /profile/{profilID}", d.profilAendern)
	weg.HandleFunc("DELETE /profile/{profilID}", d.profilLoeschen)

	// **Die vier Kontoarten** — Werkstück 6 (#1351). Sie hängen an keinem
	// Nutzer, denn sie **machen** einen: der Pfad nennt deshalb die Sache
	// (`/konten`) und nicht ihren Besitzer.
	weg.HandleFunc("POST /konten", d.kontoEintragen)
	weg.HandleFunc("POST /konten/anmeldung", d.kontoAnmelden)
	weg.HandleFunc("POST /konten/anonym", d.kontoAnonym)

	// **Der zweite Faktor hängt am Nutzer**, und der Pfad sagt es —
	// Werkstück 7 (#1351).
	weg.HandleFunc("GET /nutzer/{nutzerID}/faktoren", d.faktoren)
	weg.HandleFunc("POST /nutzer/{nutzerID}/faktoren", d.faktorEinrichten)
	weg.HandleFunc("PUT /nutzer/{nutzerID}/faktoren/totp", d.faktorBestaetigen)
	weg.HandleFunc("DELETE /nutzer/{nutzerID}/faktoren/{art}", d.faktorEntfernen)
	weg.HandleFunc("POST /pakete/uebernommen", d.paketUebernehmen)
	weg.HandleFunc("DELETE /pakete/{art}/{name}", d.paketEntfernen)
	weg.HandleFunc("GET /mse2/orte", d.mse2Orte)

	// **Ein Kartenbild von der Platte** (#1291). Am Feld und nicht an der
	// Karte: das Bildfeld ist eines von mehreren.
	weg.HandleFunc("PUT /sets/{setID}/karten/{karteID}/felder/{feld}/bild",
		d.kartenbildSetzen)

	// **Die Exportvorlagen** (#1291). Die Liste hängt am Set — sie richtet
	// sich nach dessen Spiel; der Optionswert hängt an der **Vorlage**, denn
	// er gilt über Sets hinweg.
	weg.HandleFunc("GET /sets/{setID}/vorlagen", d.vorlagen)
	weg.HandleFunc("POST /sets/{setID}/vorlagen", d.vorlageAusfuehren)
	weg.HandleFunc("PUT /vorlagen/{vorlage}/optionen/{feld}", d.vorlagenoptionSetzen)

	// **Der Probelauf** (UXR-F7, #2015). `POST`, weil die Wahl in den Rumpf
	// gehört — wie bei der Ersetzungsvorschau; er liest nur.
	weg.HandleFunc("POST /sets/{setID}/bogen/probelauf", d.bogenprobelauf)
	weg.HandleFunc("POST /sets/{setID}/vorlagen/probelauf", d.vorlagenprobelauf)
	weg.HandleFunc("GET /sets/{setID}/statistik", d.statistik)
	weg.HandleFunc("GET /sets/{setID}/statistik/{feld}", d.verteilung)
	weg.HandleFunc("GET /sets/{setID}/verlauf", d.verlauf)
	weg.HandleFunc("POST /sets/{setID}/zuruecknehmen", d.zuruecknehmen)
	weg.HandleFunc("POST /sets/{setID}/wiederherstellen", d.wiederherstellen)
	// **Die Aktualisierung** (#1100, Werkstück 7): dieselbe Fähigkeit wie
	// `mse aktualisierung`, und aus demselben Grund ohne Rücksicht auf die
	// Einstellung — wer fragt, hat gefragt.
	weg.HandleFunc("GET /aktualisierung", d.fassung)
	weg.HandleFunc("POST /aktualisierung/laden", d.fassungLaden)

	weg.HandleFunc("GET /einstellungen", d.einstellungen)
	weg.HandleFunc("PUT /einstellungen/{schluessel}", d.einstellungSetzen)

	// **Der Austausch** — Werkstück O3 von Phase 5.2 (#1427). Die zwei
	// `POST` tragen ihren Text im Rumpf; warum keiner davon ein `PUT` ist,
	// steht in `einstellungsaustausch.go`.
	weg.HandleFunc("GET /einstellungen/ausfuhr", d.einstellungsausfuhr)
	weg.HandleFunc("POST /einstellungen/vorschau", d.einstellungsvorschau)
	weg.HandleFunc("POST /einstellungen/einfuhr", d.einstellungseinfuhr)
	weg.HandleFunc("POST /sets", d.einfahren)
	weg.HandleFunc("POST /sets/neu", d.setNeu)
	weg.HandleFunc("POST /sets/{setID}/sicherung", d.sichern)
	weg.HandleFunc("POST /karten/{karteID}/bild", d.zeichnen)
	weg.HandleFunc("POST /sets/{setID}/bogen", d.bogen)

	// **MTG JSON hängt am Format, nicht am Set** (#1291): das Lesen legt
	// Sets erst an. Das Schreiben hängt an einem — deshalb zwei Gestalten.
	weg.HandleFunc("GET /mtgjson/spiele", d.mtgjsonSpiele)
	weg.HandleFunc("POST /mtgjson", d.mtgjsonLesen)
	weg.HandleFunc("POST /sets/{setID}/mtgjson", d.mtgjsonSchreiben)

	d.wegeAnGanzenDateien(weg)
	weg.HandleFunc("GET /sets/{setID}/karten/{karteID}/scryfall", d.scryfallVorschlag)
	weg.HandleFunc("POST /sets/{setID}/karten/{karteID}/scryfall", d.scryfallUebernehmen)
	weg.HandleFunc("POST /sets/{setID}/karten/{karteID}/scryfall/bild", d.scryfallBild)

	// **Der Verweis hängt an der Karte, nicht am Set** (#1167): er sagt,
	// welchen Druck **diese** Karte meint. Der Weg über das Set daneben ist
	// die Liste — eine Abfrage je Karte wären bei dreihundert Karten
	// dreihundert Anfragen für eine Ansicht.
	weg.HandleFunc("GET /karten/{karteID}/verweis", d.verweisZeigen)
	weg.HandleFunc("PUT /karten/{karteID}/verweis", d.verweisSetzen)
	weg.HandleFunc("DELETE /karten/{karteID}/verweis", d.verweisLoesen)
	weg.HandleFunc("GET /sets/{setID}/verweise", d.setverweiseZeigen)

	// **Marken an Karten** (#1208).
	weg.HandleFunc("GET /sets/{setID}/marken", d.setmarkenZeigen)
	weg.HandleFunc("GET /sets/{setID}/karten/{karteID}/marken", d.kartenmarkenZeigen)
	weg.HandleFunc("PUT /sets/{setID}/karten/{karteID}/marken/{marke}", d.markeSetzen)
	weg.HandleFunc("DELETE /sets/{setID}/karten/{karteID}/marken/{marke}", d.markeEntfernen)

	// **Die Zwischenablage** (#1168). `POST` bei allen dreien, weil die
	// Kartenliste beziehungsweise der Inhalt in den Rumpf gehört: eine Auswahl
	// von dreihundert Karten in einer Abfrage wäre eine Adresse, die keine
	// Zwischenstelle weiterreicht.
	weg.HandleFunc("POST /sets/{setID}/ablage/kopieren", d.ablageKopieren)
	weg.HandleFunc("POST /sets/{setID}/ablage/ausschneiden", d.ablageAusschneiden)
	weg.HandleFunc("POST /sets/{setID}/ablage/einfuegen", d.ablageEinfuegen)

	// **Suchen und Ersetzen** (#1168) — `POST` bei beiden, weil die Frage in
	// den Rumpf gehört; die Vorschau ändert trotzdem nichts.
	weg.HandleFunc("POST /sets/{setID}/ersetzung/vorschau", d.ersetzungVorschau)
	weg.HandleFunc("POST /sets/{setID}/ersetzung", d.ersetzungAusfuehren)

	// **Die Gestalt der Kartenliste** (#1168) — lesen, Spalten setzen,
	// Sortierung setzen. `PUT` bei beiden Schreibwegen, weil sie die **ganze**
	// Wahl ersetzen und nicht eine Spalte anhängen.
	weg.HandleFunc("GET /sets/{setID}/kartenliste", d.kartenliste)
	weg.HandleFunc("GET /sets/{setID}/kartenfilter", d.kartenfilter)
	weg.HandleFunc("PUT /sets/{setID}/kartenliste/spalten", d.listenspaltenSetzen)
	weg.HandleFunc("PUT /sets/{setID}/kartenliste/sortierung", d.listensortierungSetzen)

	// **Ohne Kartenkennung** (#1032): die Karte gibt es noch nicht. Angelegt
	// wird **im Set**, und das Set ist die Sammlung, an die etwas angehängt
	// wird.
	weg.HandleFunc("POST /sets/{setID}/karten/scryfall", d.scryfallKarte)

	// **Ohne Setkennung** (#1033): das Set entsteht hier.
	weg.HandleFunc("POST /sets/scryfall", d.scryfallSatz)

	// **Diese beiden hängen an keinem Set** (#888): eine Regelauskunft und
	// eine Manakostenprüfung lesen nichts aus dem Bestand.
	// **Ein Ausdruck im Kontext einer Karte** (#1070). `POST`, weil ein
	// Ausdruck kein Wegbestandteil ist — geändert wird nichts.
	weg.HandleFunc("POST /karten/{karteID}/skript", d.skript)
	weg.HandleFunc("GET /karten/{karteID}/skriptnamen", d.skriptnamen)

	// **Die Paketwerkstatt** (#1164). `{datei...}` und nicht `{datei}`: ein
	// Name im Paket kann Schrägstriche tragen — `magic-mana-large/symbol`.
	weg.HandleFunc("GET /pakete/{paket}/dateien", d.paketdateien)
	weg.HandleFunc("GET /pakete/{paket}/dateien/{datei...}", d.paketdateiLesen)
	weg.HandleFunc("PUT /pakete/{paket}/dateien/{datei...}", d.paketdateiSchreiben)
	weg.HandleFunc("POST /pakete/{paket}/dateien", d.paketdateiAnlegen)                // FMT-02 (#1691)
	weg.HandleFunc("PATCH /pakete/{paket}/dateien/{datei...}", d.paketdateiUmbenennen) // FMT-02 (#1691)
	weg.HandleFunc("DELETE /pakete/{paket}/dateien/{datei...}", d.paketdateiEntfernen) // FMT-02 (#1691)
	weg.HandleFunc("POST /paketdateien/pruefung", d.paketdateiPruefen)
	weg.HandleFunc("GET /sets/{setID}/symbol", d.symbolLesen)
	weg.HandleFunc("PUT /sets/{setID}/symbol", d.symbolSetzen)
	weg.HandleFunc("POST /symbol/{rechnung}", d.symbolRechnen)
	weg.HandleFunc("GET /karten/{karteID}/symbolpalette/{feld}", d.symbolpalette)

	// **Die Sammlung** (#1165). Sie steht **neben** `/sets` und nicht darin:
	// was jemand besitzt, ist ein Druck und kein Entwurf.
	weg.HandleFunc("GET /sammlung", d.sammlung)
	weg.HandleFunc("GET /sammlung/drucke/{printing}", d.sammlungFuer)
	weg.HandleFunc("POST /sammlung", d.exemplarAnlegen)
	weg.HandleFunc("PUT /sammlung/{id}", d.exemplarAendern)
	weg.HandleFunc("DELETE /sammlung/{id}", d.exemplarLoeschen)

	// **Die Listeneinfuhr** (#1292) — `POST` auch, wo nichts geschrieben wird.
	weg.HandleFunc("POST /sammlung/liste/deuten", d.listeDeuten)
	weg.HandleFunc("POST /sammlung/liste/vorschau", d.listeVorschauen)
	weg.HandleFunc("POST /sammlung/liste/einlesen", d.listeEinlesen)
	// **Die Ausfuhr** (#1292) — die Form steht in der Abfrage.
	weg.HandleFunc("GET /sammlung/ausgeben", d.sammlungAusgeben)
	weg.HandleFunc("POST /sammlung/vorschau", d.sammlungCSVVorschauen)
	weg.HandleFunc("POST /sammlung/einlesen", d.sammlungCSVEinlesen)

	// **Die vier Fragen an eine geordnete Sammlung** (#1292, Vermerk 0057).
	// Alle `GET`: sie lesen.
	weg.HandleFunc("GET /sammlung/besitz/{printing}", d.sammlungBesitz)
	weg.HandleFunc("GET /sammlung/lagerorte", d.sammlungLagerorte)

	// **Die Kartenskripte eines Sets** (#1337). Der zweite legt Karten an und
	// ist deshalb ein `POST`: ein zweiter Ruf legt weitere an.
	weg.HandleFunc("GET /sets/{setID}/kartenskripte", d.kartenskripte)
	weg.HandleFunc("POST /sets/{setID}/kartenskripte", d.kartenskriptAusfuehren)

	// **Die Auszählungsdimensionen eines Sets** (#1338). Beide `GET`: sie
	// rechnen und ändern nichts.
	weg.HandleFunc("GET /sets/{setID}/auszaehlung", d.auszaehldimensionen)
	weg.HandleFunc("GET /sets/{setID}/auszaehlung/{name}", d.auszaehlung)

	// **Die Parameterhilfen eines Schlüsselworts** (#1342). `GET`: sie
	// sagen, was sich einfügen lässt, und fügen nichts ein.
	weg.HandleFunc("GET /sets/{setID}/schluesselwoerter/{wortID}/parameter",
		d.parameterhilfen)

	// **Die Autoersetzungen eines Spiels** (#1341, Vermerk 0060). `PUT` auf
	// das Muster, weil es der Schlüssel ist; `DELETE` auf die Sammlung setzt
	// alles zurück.
	weg.HandleFunc("POST /sets/{setID}/symbol-nachzeichnen", d.symbolNachzeichnen)
	weg.HandleFunc("GET /sets/{setID}/autoersetzungen", d.autoersetzungen)
	weg.HandleFunc("PUT /sets/{setID}/autoersetzungen/{muster}",
		d.autoersetzungSetzen)
	weg.HandleFunc("DELETE /sets/{setID}/autoersetzungen/{muster}",
		d.autoersetzungLoeschen)
	weg.HandleFunc("DELETE /sets/{setID}/autoersetzungen",
		d.autoersetzungenZuruecksetzen)
	weg.HandleFunc("PUT /sets/{setID}/autoersetzungen", d.autoersetzungSchalten)
	// **Decks und Formate** (#1293, Vermerk 0061) stehen in `deckwegetafel.go`.
	d.deckwege(weg)

	// **Turniere, Symbolbibliothek, eigene Übersetzungen** (#1294, DOM-02,
	// LOC-02) — je in einer eigenen Datei.
	d.turnierwege(weg)
	d.symbolbibliothekswege(weg)
	d.sprachwege(weg)

	weg.HandleFunc("GET /sammlung/dubletten", d.sammlungDubletten)
	weg.HandleFunc("GET /sammlung/wuensche", d.sammlungWuensche)

	// **Die Massenvorgänge** (#1292). Alle `POST`: die Zeilen samt Fassungen
	// gehören in den Rumpf, und ein `DELETE` hat keinen.
	weg.HandleFunc("POST /sammlung/masse", d.masseAendern)
	weg.HandleFunc("POST /sammlung/masse/loeschen", d.masseLoeschen)
	weg.HandleFunc("POST /sammlung/masse/preise", d.massePreise)

	// **Die Preise** (#1165). `POST` beim Holen, weil es eine Momentaufnahme
	// **erzeugt** und dafür einen fremden Dienst fragt — ein `GET` müsste
	// zwischenspeicherbar sein, und dieser Weg ist es nicht.
	weg.HandleFunc("GET /drucke/{printing}/preise", d.preise)
	weg.HandleFunc("GET /drucke/{printing}/preisverlauf", d.preisverlauf)
	weg.HandleFunc("GET /drucke/{printing}/angebote", d.marktangebote)
	weg.HandleFunc("POST /drucke/{printing}/fremdverlauf", d.fremdpreisverlauf)
	weg.HandleFunc("GET /anbieter/{dienst}/stand", d.anbieterstand)
	weg.HandleFunc("GET /ki/modelle", d.kimodelle)
	weg.HandleFunc("POST /karten/{karteID}/ki-vorschlag", d.kivorschlag)
	weg.HandleFunc("POST /ki/batch", d.kibatch)
	weg.HandleFunc("POST /ki/gruppe", d.kigruppe)
	weg.HandleFunc("POST /ki/erkennung", d.kierkennung)
	weg.HandleFunc("POST /karten/{karteID}/ki-einschaetzung", d.kieinschaetzung)
	weg.HandleFunc("POST /ki/modelle/pruefen", d.kimodellpruefen)
	weg.HandleFunc("POST /ki/modelle/vergessen", d.kimodellvergessen)
	weg.HandleFunc("POST /ki/geraetemodell/sitzung", d.kigeraetsitzung)
	weg.HandleFunc("POST /preise/bulk/{quelle}/{umfang}", d.preisbulk)
	weg.HandleFunc("POST /drucke/{printing}/preise", d.preiseHolen)
	weg.HandleFunc("GET /portfolio", d.portfolio)

	// **Das Zufallspack** (#889, Befund 7 des Zweitblicks über Phase W).
	// Ziehen ist ein `POST`, obwohl es nichts ändert: es ist kein Abruf einer
	// Ressource, sondern eine **Erzeugung** — und der Samen macht sie
	// wiederholbar, nicht zwischenspeicherbar.
	weg.HandleFunc("GET /sets/{setID}/packarten", d.packarten)
	weg.HandleFunc("POST /sets/{setID}/packarten/{art}/ziehung", d.packziehen)

	// **Die Auswahllisten eines Feldes** (#1074) — zeigen und wählen, in
	// dieser Reihenfolge (§16 Regel 6).
	weg.HandleFunc("GET /karten/{karteID}/felder/{feld}/wortlisten", d.wortlisten)
	weg.HandleFunc(
		"POST /sets/{setID}/karten/{karteID}/felder/{feld}/wortwahl", d.wortwaehlen)

	// **Die Schlüsselwörter eines Sets** (#1062) — sehen, anlegen, ändern,
	// löschen. Sie hängen am **Set** und an keiner Karte.
	weg.HandleFunc("GET /sets/{setID}/schluesselwoerter", d.schluesselwoerter)
	weg.HandleFunc("POST /sets/{setID}/schluesselwoerter", d.schluesselwortAnlegen)
	weg.HandleFunc(
		"PUT /sets/{setID}/schluesselwoerter/{wortID}/felder/{feld}",
		d.schluesselwortSetzen)
	weg.HandleFunc(
		"DELETE /sets/{setID}/schluesselwoerter/{wortID}", d.schluesselwortLoeschen)

	weg.HandleFunc("GET /scryfall/regeln", d.scryfallRegeln)

	// **Die Kartensuche** (#1167) — der Weg, über den jemand eine Karte
	// findet, deren Namen er nicht auswendig kennt.
	weg.HandleFunc("GET /scryfall/suche", d.scryfallSuche)

	// **Der Dienstkatalog** (#1440) — was die fremden Dienste hier können
	// und ob sie es jetzt dürfen. Ohne Parameter, ohne Nebenwirkung.
	weg.HandleFunc("GET /dienstkatalog", d.dienstkatalog)

	d.peerwege(weg)
	d.objektmarkenwege(weg)

	weg.HandleFunc("GET /scryfall/manakosten", d.scryfallManakosten)
	weg.HandleFunc("GET /scryfall/sammeldaten", d.sammeldatenZeigen)
	weg.HandleFunc("POST /scryfall/sammeldaten", d.sammeldatenHolen)
}
