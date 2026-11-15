// **Die Schlüssel, unter denen eine Meldung an der Oberfläche ankommt** —
// aus `fehler.go` herausgelöst am 2026-09-06.
//
// # Warum die Trennung
//
// **Zwei Fragen, und beide wachsen.** Nebenan steht, *welche Fehlerarten es
// gibt*; hier steht, *unter welchem Namen eine Meldung übersetzt wird*. Die
// Zeilengrenze aus §2 war der Anlass — Werkstück 14 von Phase 4.12 hat die
// Datei auf genau 400 Zeilen gebracht —, aber der Schnitt folgt der Frage und
// nicht der Zahl: wer einen Schlüssel nachschlägt, sucht ihn nicht zwischen
// Fehlertypen mit eigenen `Error`-Methoden.

package domain

// Schluessel benennt eine Meldung sprachunabhängig.
//
// Der deutsche Satz reist mit (siehe FieldError) und ist der Rückfall: ein
// Schlüssel, den die Kataloge nicht kennen, kostet eine unübersetzte Meldung,
// nicht eine leere (AGENTS.md §9).
type Schluessel string

// Die angemeldeten Schlüssel.
//
// Wer einen neuen setzt, trägt ihn hier und in beiden Katalogen ein —
// `internal/app` hält beide Seiten zusammen und wird rot, sobald eine fehlt.
const (
	SchluesselUnbekannt Schluessel = "unbekannt"

	// SchluesselUnbekanntMitVorgang ist derselbe Fall **mit dem Griff ins
	// Protokoll** —
	// #1601.
	//
	// **Warum zwei Schlüssel und nicht ein Platzhalter, der leer bleibt.**
	// Der Rückfallsatz verweist auf das Protokoll, und ein Protokoll hat
	// viele Zeilen; die Vorgangskennung steht in jeder Zeile dieses Rufs
	// und ist der Griff, mit dem ein Mensch die richtige findet. **Wo es
	// keine gibt** — ein Lauf ohne Grenze —, wäre ein `{vorgang}` im Satz
	// sichtbar leer; dann gilt der Satz ohne ihn.
	SchluesselUnbekanntMitVorgang Schluessel = "unbekanntMitVorgang"
	SchluesselNichtGefunden       Schluessel = "nichtGefunden"
	SchluesselKonflikt            Schluessel = "konflikt"
	SchluesselPflichtfeld         Schluessel = "pflichtfeld"
	SchluesselValidierung         Schluessel = "validierung"

	// SchluesselValidierungsgrund ist die Validierung **ohne Feld und ohne
	// eigenen Schlüssel** — #1601.
	//
	// # Warum sie einen eigenen braucht
	//
	// **Weil `validierung` auf ein Feld zeigt** („der Hinweis am Feld sagt,
	// woran es liegt"), und diese hat keines: sie entsteht tief im Fachkern,
	// wo niemand weiß, welches Formular gerade offensteht.
	//
	// **Und weil `unbekannt` auf das Protokoll zeigt**, das für eine
	// Validierung nach eigener Entscheidung leer bleibt
	// (`TestErwarteteFehlerStehenNichtImProtokoll`). Bis zum 2026-09-17 fiel
	// sie dorthin, und der Satz schickte Menschen an eine Stelle, an der
	// nichts steht — eine Zusage, die zur Hälfte hält, ist schlimmer als
	// keine (§10).
	//
	// **Der Grund reist als Wert mit**, unter `grund`: der Satz aus dem
	// Fachkern ist für Menschen geschrieben (§8), und der Katalog setzt ihn
	// in seinen eigenen Rahmen ein — so bleibt der Rahmen übersetzt und die
	// Auskunft erhalten.
	SchluesselValidierungsgrund Schluessel = "validierungsgrund"

	// Die zwei für die Massenänderung — Befund 1 des Zweitblicks über
	// Phase 3.5 (#1211).
	//
	// **Zwei und nicht einer**: „lässt sich nicht bearbeiten" und „entsteht
	// aus anderen Feldern" sind zwei Auskünfte mit zwei Abhilfen. Wer sie
	// zusammenlegte, schickte jemanden auf die falsche Suche (§8).
	SchluesselNichtEditierbar Schluessel = "nichtEditierbar"
	SchluesselBerechnetesFeld Schluessel = "berechnetesFeld"

	// Die zwei für das Zurückspielen einer Sicherung
	// (#940).
	//
	// **Sie stehen hier und nicht bei `unbekannt`**, weil eine abgewiesene
	// Sicherung kein unerwarteter Fehler ist: sie ist der vorgesehene Ausgang
	// einer Prüfung, und die Empfangenden können etwas damit anfangen.
	SchluesselSicherungNeuereFassung Schluessel = "sicherungNeuereFassung"
	SchluesselSicherungUnbrauchbar   Schluessel = "sicherungUnbrauchbar"

	// SchluesselSicherungVerworfen gehört zu [ErrSicherungVerworfen]: die eben
	// geschriebene Sicherung bestand ihre Prüfung nicht (F05, #2344).
	SchluesselSicherungVerworfen Schluessel = "sicherungVerworfen"

	// SchluesselSicherungAeltereFassung heißt: die Sicherung steht **vor der
	// Grundlinie** und lässt sich hier nicht mehr nachziehen — J02
	// (#2266,
	// Vermerk 0189). **Ein eigener Rat**: über 0.7.0 geht es weiter.
	SchluesselSicherungAeltereFassung Schluessel = "sicherungAeltereFassung"

	// SchluesselSicherungFremdeFassung heißt: das **Archiv** stammt aus einer
	// anderen Fassung (#1162).
	//
	// **Nicht dasselbe wie `sicherungNeuereFassung`**: der nennt die
	// Schemaschritte im **Bestand**, dieser die Fassung im **Verzeichnis der
	// Sicherung**. Eine künftige Fassung kann das Schema unverändert lassen
	// und die Gestalt des Archivs ändern; dann fängt nur dieser sie.
	SchluesselSicherungFremdeFassung Schluessel = "sicherungFremdeFassung"

	// SchluesselErlaubnis gehört zu [ErrErlaubnis] — Werkstück 11 von
	// Phase 4.12
	// (#1351).
	//
	// **Der einzige Ausgang, der sagt: du bist es, du darfst nur nicht.**
	SchluesselErlaubnis Schluessel = "erlaubnis"

	// SchluesselUnbekannteHerkunftsart und SchluesselHerkunftOhneProfil
	// gehören zu [Herkunft] — Werkstück 12 von Phase 4.12
	// (#1351).
	//
	// **Zwei und nicht einer**: die eine Abweisung heißt „das ist keine
	// Herkunft", die andere „diese Herkunft ist in sich widersprüchlich".
	// Ein gemeinsamer Schlüssel gäbe der Fläche denselben Satz für zwei
	// Fälle, die verschiedene nächste Schritte haben.
	SchluesselUnbekannteHerkunftsart Schluessel = "unbekannteHerkunftsart"
	SchluesselHerkunftOhneProfil     Schluessel = "herkunftOhneProfil"

	// SchluesselNurLesend gehört zu [ErrNurLesend] — Werkstück 16 von Phase
	// 4.12. **Ein eigener Schlüssel**, weil der Satz sagen muss, was zu tun
	// ist: das Profil verlassen oder um eine Rolle bitten.
	SchluesselNurLesend Schluessel = "nurLesend"

	// SchluesselSchluesselFehlt gehört zu [ErrSchluesselFehlt] — Werkstück 14
	// von Phase 4.12.
	//
	// **Ein eigener Schlüssel und kein `unbekannt`**, weil das Fenster hier
	// mehr sagen muss als „ein Fehler": was zu tun ist, hängt daran, ob jemand
	// den Schlüsselbund wiederherstellen kann.
	SchluesselSchluesselFehlt Schluessel = "schluesselFehlt"

	// SchluesselBelegt gehört zu [ErrBelegt] — der einzige Ausgang, der von
	// allein vergeht (#886).
	SchluesselBelegt Schluessel = "belegt"

	// SchluesselBestandWirdGetauscht heißt: der Bestand wird gerade
	// getauscht — zurückgespielt, verschlüsselt oder neu angefangen — und
	// gleich wieder da (A02, #2216).
	//
	// **Nicht `belegt`**: dessen Satz vermutet ein anderes Fenster, und hier
	// ist es dieses.
	SchluesselBestandWirdGetauscht Schluessel = "bestandWirdGetauscht"

	// SchluesselAnwendungEndet heißt: das Programm endet, und der Schritt wird
	// nicht mehr getan — bis A02 kam das als „ein unerwarteter Fehler" an.
	SchluesselAnwendungEndet Schluessel = "anwendungEndet"

	// SchluesselVorgangAbgebrochen heißt: der Vorgang wurde abgebrochen,
	// bevor er fertig war — oder seine Frist lief ab. Bis zum Nachlauf über
	// Phase A kam das als „ein unerwarteter Fehler" an.
	SchluesselVorgangAbgebrochen Schluessel = "vorgangAbgebrochen"

	// SchluesselKarteInDeck heißt: diese Karte liegt in einem Deck, und das
	// Löschen nähme sie auch dort heraus
	// (#1369).
	//
	// **Ein eigener Schlüssel und nicht `konflikt`.** Ein Konflikt heißt
	// „lade neu"; dieser heißt „entscheide". Ohne die Unterscheidung böte
	// keine Fläche das Trotzdem an, weil keine wüsste, wann.
	SchluesselKarteInDeck Schluessel = "karteInDeck"

	// **Die Befunde eines Ordners** — OPS-02
	// (#1719).
	// **Eigene Schlüssel und nicht `validierung`**: der Satz ist die Abhilfe
	// — ein Laufwerk anschließen, einen Ordner statt einer Datei nennen —,
	// und über `validierung` zeigte das Fenster nur „nicht gültig". Sie
	// tragen `ordner` (oder `wert`) als Platzhalter.
	SchluesselOrdnerangabe   Schluessel = "ordnerangabe"
	SchluesselOrdnerFehlt    Schluessel = "ordnerFehlt"
	SchluesselOrdnerIstDatei Schluessel = "ordnerIstDatei"
	SchluesselOrdnerGesperrt Schluessel = "ordnerGesperrt"
	SchluesselOrdnerVoll     Schluessel = "ordnerVoll"

	// SchluesselFalschesSpiel heißt: das Deck gehört zu einem anderen Spiel
	// als das Turnier (Befund 3 des Nachlaufs über Phase 4.7).
	//
	// **Ein eigener Schlüssel und nicht `validierung`**, weil die Abhilfe
	// eine andere ist: nicht „berichtige das Feld", sondern „nimm ein
	// anderes Deck".
	SchluesselFalschesSpiel Schluessel = "falschesSpiel"

	// SchluesselTurnierformatSpiel heißt: das gewählte Format gehört zu einem
	// anderen Spiel als das Turnier — DOM-04.AK-01 (#1699). **Nicht
	// [SchluesselFalschesSpiel]**: dessen Satz spricht vom Deck, und die
	// Abhilfe hier ist ein anderes Format.
	SchluesselTurnierformatSpiel Schluessel = "turnierformatSpiel"

	// SchluesselTurnierAbgeschlossen heißt: das Turnier ist abgeschlossen und
	// nimmt keine Änderung an — DOM-04.AK-04 (#1699). **Ein Konflikt**: die
	// Eingabe ist in Ordnung, der Stand steht ihr entgegen; die Abhilfe ist
	// das Wiedereröffnen.
	SchluesselTurnierAbgeschlossen Schluessel = "turnierAbgeschlossen"

	// SchluesselPaketGeaendert heißt: das Paket weicht vom Stand beim
	// Einbinden ab, und ein Ersetzen braucht eine Bestätigung —
	// DOM-01.AK-03 (#1696). Ein Konflikt, keine Abweisung.
	SchluesselPaketGeaendert Schluessel = "paketGeaendert"

	// SchluesselSitzGegangen heißt: dieser Sitz hat die Partie verlassen und
	// stellt keine Befehle mehr — CR 800.4a (Fassung 2026-09-25), #1979.
	// **Ein Konflikt**: der Befehl ist in Ordnung, der Stand steht ihm
	// entgegen.
	SchluesselSitzGegangen Schluessel = "sitzGegangen"

	// SchluesselFremderWert heißt: in diesem Feld steht inzwischen ein Wert,
	// den der Entwurf nicht kannte — jemand anderes hat ihn geändert, und der
	// Entwurf darüber wäre ein stilles Überschreiben (U02b, #2179). Ein
	// Konflikt am Feld.
	SchluesselFremderWert Schluessel = "fremderWert"

	// SchluesselVorschlagNurGegenseite heißt: aus einem offenen Angebot
	// schlägt nur die Gegenseite vor — die Rolle aus der Matrix für
	// `gegenangebot` (TRADE-02, #1716; #1975).
	SchluesselVorschlagNurGegenseite Schluessel = "vorschlagNurGegenseite"

	// SchluesselTurnierNichtFertig heißt: ein Turnier ohne Runde oder mit
	// einem offenen Tisch lässt sich nicht abschließen.
	SchluesselTurnierNichtFertig Schluessel = "turnierNichtFertig"

	// SchluesselTurniernameLeer heißt: ein Turnier ohne Namen
	// (#1294).
	SchluesselTurniernameLeer Schluessel = "turniernameLeer"

	// SchluesselTurnierspielLeer heißt: ein Turnier ohne Spiel.
	SchluesselTurnierspielLeer Schluessel = "turnierspielLeer"

	// SchluesselPaketnameUngueltig heißt: dieser Name taugt nicht als
	// Paketverzeichnis
	// (#1350).
	//
	// **Ein eigener Schlüssel und nicht `validierung`**, weil die Regel selbst
	// die Abhilfe ist: „Buchstaben, Ziffern, Bindestriche und Unterstriche"
	// sagt, was zu tun ist — „Die Eingabe ist nicht gültig" sagt es nicht.
	//
	// Angestrichen vom Nachlauf über Phase 4.10: das Fenster ersetzte die
	// genaue Meldung durch den allgemeinen Katalogsatz, und ein Test von mir
	// hielt dieses Verhalten fest, statt es zu bestreiten (§6).
	SchluesselPaketnameUngueltig Schluessel = "paketnameUngueltig"

	// SchluesselKeinSetMitStil heißt: kein offenes Set benutzt diesen Stil
	// (#1191).
	//
	// **Er steht neben `nichtGefunden` und nicht darin.** Dieser Code trägt
	// jedes `ErrNotFound` — eine fehlende Karte, ein fehlendes Paket, ein
	// fehlender Stil. Die Vorlagenvorschau macht daran aber einen **leeren
	// Zustand** fest („öffne ein Set, das sie nennt"), und ein anderes
	// `ErrNotFound` schickte damit jemanden an die falsche Stelle.
	// Angestrichen vom Nachlauf des Zweitblicks über Phase 3.3.
	SchluesselKeinSetMitStil Schluessel = "keinSetMitStil"

	// SchluesselSchluesselbundAblage heißt: der Schlüssel ließ sich nicht im
	// Schlüsselbund ablegen (S13.NAT-03, #1706).
	SchluesselSchluesselbundAblage Schluessel = "schluesselbundAblage"

	// SchluesselEinladungZuNeu heißt: die Einladung ist aus einer neueren
	// Fassung (N3, #1456).
	SchluesselEinladungZuNeu Schluessel = "einladungZuNeu"

	// SchluesselFremderNutzer heißt: über diesen Menschen schreibt nur, wer
	// ihn werden kann — ID-03 (#1702).
	SchluesselFremderNutzer Schluessel = "fremderNutzer"

	// SchluesselFernhorchenLaeuft heißt: es ist schon eine Einladung offen
	// (N3, #1456).
	SchluesselFernhorchenLaeuft Schluessel = "fernhorchenLaeuft"

	// SchluesselAnmeldungAbgelaufen heißt: der Rückruf einer Anmeldung kam
	// nicht binnen ihrer Frist (BK-07d, #2118).
	SchluesselAnmeldungAbgelaufen Schluessel = "anmeldungAbgelaufen"

	// Die drei Ausgänge eines Umschlags (N4,
	// #1457).
	//
	// **Drei und nicht einer**, weil ein Mensch an drei verschiedenen Stellen
	// nachsähe: bei der Fassung an der Gegenseite, bei der Unterschrift am
	// Absender, beim Inhalt an der Datei.
	SchluesselUmschlagFassung      Schluessel = "umschlagFassung"
	SchluesselUmschlagUnterschrift Schluessel = "umschlagUnterschrift"
	SchluesselUmschlagInhalt       Schluessel = "umschlagInhalt"

	// SchluesselVorlageFehlt heißt: das Vorlagenpaket dieser Karte ist nicht
	// eingebunden (#1055).
	//
	// **Er steht hier, weil er der Normalfall eines frischen Rechners ist.**
	// Ein Set von Scryfall einlesen geht ohne Vorlage, ein Bild einsetzen
	// nicht — und wer „Es ist ein unerwarteter Fehler aufgetreten" las, wusste
	// nicht, dass ihm ein Paket fehlt.
	SchluesselVorlageFehlt Schluessel = "vorlageFehlt"

	// SchluesselSkriptfehler heißt: der Ausdruck ist nicht auswertbar
	// (#1164).
	//
	// **Er steht hier, weil eine Stelle ohne Aussage nichts nützt.** Der
	// Zerteiler sagt seit je, was fehlt — „etwas fehlt", „eine Zeichenkette
	// ohne schließendes Anführungszeichen", „das Zeichen ; gehört nicht zur
	// Sprache". Über `unbekannt` erreichte davon nichts das Fenster: es
	// bevorzugt den Katalog, und dort steht der Rückfallsatz.
	//
	// **Sein Katalogtext hat einen Platzhalter `{meldung}`** und trägt damit
	// den Satz des Zerteilers. Dass dieser Satz **deutsch bleibt**, auch im
	// englischen Katalog, ist eine offene Stelle und keine Entscheidung —
	// #1178.
	SchluesselSkriptfehler Schluessel = "skriptFehler"

	// SchluesselZweiterFaktor heißt: hier fehlt der zweite Faktor —
	// Werkstück 7 von Phase 4.12
	// (#1351).
	//
	// **Er steht hier, weil die Fläche etwas anderes tun muss.** Ein
	// `validierung` lässt sie dasselbe Formular noch einmal zeigen; dieser
	// sagt ihr, **nach dem Code zu fragen** — und die Art des Faktors reist
	// als Wert mit, damit sie das richtige Feld anbietet.
	SchluesselZweiterFaktor Schluessel = "zweiterFaktor"

	// SchluesselFaktorGesperrt heißt: zu viele falsche Codes, gewartet wird
	// bis `bis` — SEC-10 (#2151, Vermerk 0178).
	SchluesselFaktorGesperrt Schluessel = "faktorGesperrt"

	// **Die Kommandozentrale** — Werkstück U12 von Phase 5.1 (#1423): ein
	// gefährlicher Befehl braucht die Bestätigung, und Argumente sind ein
	// JSON-Objekt oder keine.
	SchluesselBefehlBestaetigen Schluessel = "befehlBestaetigen"
	SchluesselBefehlArgumente   Schluessel = "befehlArgumente"

	// **Was der Bestand selbst sagt** — ENG-05.AK-03 (#1730) und #1812.
	// Warum `konflikt` dafür nicht reicht, steht bei [ErrDoppelt].
	//
	// `doppelterName` steht am Feld eines Namens, `doppeltesExemplar` an einer
	// Sammlungszeile (F08), `doppelt` an allem Übrigen; `konfliktgrund` trägt
	// den Satz des Fachkerns unter `grund` wie `validierungsgrund`.
	SchluesselDoppelt           Schluessel = "doppelt"
	SchluesselDoppelterName     Schluessel = "doppelterName"
	SchluesselDoppeltesExemplar Schluessel = "doppeltesExemplar"
	SchluesselKonfliktgrund     Schluessel = "konfliktgrund"
	SchluesselVoll              Schluessel = "voll"

	// SchluesselGesperrt: auf diesen Ort darf die Anwendung nicht zugreifen
	// ([ErrGesperrt]) — FIX-F02.
	SchluesselGesperrt Schluessel = "gesperrt"

	// **Ein fremder Dienst, der nicht antwortet** — je Fall aus
	// `gegenstelle.go` ein Satz mit seinem Rat, `{wirt}` und, wo eine Frist
	// griff, `{sekunden}` (D02, #2233). `gedrosselt` ohne Wirt: nicht jeder
	// Klient nennt ihn.
	SchluesselNetzName            Schluessel = "netzName"
	SchluesselNetzNichtErreichbar Schluessel = "netzNichtErreichbar"
	SchluesselNetzKeineAntwort    Schluessel = "netzKeineAntwort"
	SchluesselNetzAntwortStockt   Schluessel = "netzAntwortStockt"
	SchluesselNetzAbgerissen      Schluessel = "netzAbgerissen"
	SchluesselNetzProxyAnmeldung  Schluessel = "netzProxyAnmeldung"
	SchluesselGedrosselt          Schluessel = "gedrosselt"

	// SchluesselGedrosseltBis ist eine Drosselung mit bekanntem Ende — der
	// Satz nennt Wirt und Uhrzeit (R01, #2333).
	SchluesselGedrosseltBis Schluessel = "gedrosseltBis"

	// SchluesselKontingent ist das eigene Budget eines Aufrufers an dieser
	// Anwendung — der Satz nennt die `{sekunden}` bis zum nächsten Ruf
	// (Zweitblick über Phase 1 von 0.9.0, Befund 5).
	SchluesselKontingent Schluessel = "kontingent"

	// SchluesselProzessVeraltet ist ein laufender Prozess, unter dem eine
	// neuere Fassung den Bestand gewandert hat — er endet; ein eingerichteter
	// Dienst startet unter launchd mit der neuen Fassung neu (M02, #2346).
	SchluesselProzessVeraltet Schluessel = "prozessVeraltet"

	// SchluesselBestandVomDienstGehalten ist ein Zurückspielen, das nicht
	// allein ist, weil der eingerichtete Dienst den Bestand hält — der Satz
	// sagt, wo er sich widerrufen lässt (M02, #2346).
	SchluesselBestandVomDienstGehalten Schluessel = "bestandVomDienstGehalten"
)
