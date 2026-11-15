// **Die Liste aller Endpunkte.**
//
// Eine eigene Datei, weil sie eine eigene Frage beantwortet: `vorrichtung_test.go`
// baut die Doubles — hier steht, **wogegen** sie geprüft werden. Ein Wächter
// hält die Liste gegen den Quelltext; ein Endpunkt, der hier fehlt, hat keine
// der vier Abweisungsklassen.
package rest_test

import "net/http"

var alleEndpunkte = append(append(append(append(endpunkteAmBestand(),
	endpunkteDaneben()...), endpunkteDecks()...), endpunkteTurniere()...),
	append(append(append(append(endpunktePeer(), endpunkteKontakte()...), endpunkteEinfuhrvorschau()...),
		append(endpunktePakete(), endpunkteProbelauf()...)...), append(append(endpunkteSymbolbibliothek(), endpunkteAuftraege()...), endpunkteSprache()...)...)...)

// endpunkteAmBestand sind die Wege, die an einem Set oder einer Karte hängen.
func endpunkteAmBestand() []endpunkt {
	// **Und die Wege am Set dahinter** (§2): sie stehen in
	// `endpunktesatz_test.go`, weil die Datei sonst über die Zeilengrenze
	// ginge — der Schnitt folgt der Frage, nicht der Zahl.
	return append([]endpunkt{
		{name: "sets lesen", methode: http.MethodGet, weg: "/sets", erfolg: http.StatusOK},
		{
			name: "karten lesen", methode: http.MethodGet,
			weg: "/sets/1/karten", erfolg: http.StatusOK,
		},
		{
			name: "feld setzen", methode: http.MethodPut,
			weg: "/sets/1/karten/2/felder/name", rumpf: `{"wert":"neu"}`, erfolg: http.StatusOK,
		},
		{
			name: "set einfahren", methode: http.MethodPost, weg: "/sets",
			rumpf: `{"pfad":"/tmp/x.mse-set"}`, erfolg: http.StatusCreated,
		},
		{
			// **`POST /sets/neu` und nicht `POST /sets`** (#891): jener fährt eine
			// **Datei** ein. Zwei Fähigkeiten hinter einer Adresse wären eine
			// Zusage, die niemand lesen kann.
			name: "Set anlegen", methode: http.MethodPost, weg: "/sets/neu",
			rumpf: `{"name":"Neu","spiel":"probe"}`, erfolg: http.StatusCreated,
		},
		{
			// **Der Name steht in der Abfrage** und aus demselben Grund wie die
			// Fassung beim Exemplar: ein `DELETE` hat keinen Rumpf (#1218).
			name: "Set löschen", methode: http.MethodDelete,
			weg: "/sets/1?name=Probeset", erfolg: http.StatusOK,
		},
		{
			name: "set sichern", methode: http.MethodPost,
			weg: "/sets/1/sicherung", rumpf: `{"ziel":"/tmp/y.mse-set"}`, erfolg: http.StatusCreated,
		},
		{
			name: "karte zeichnen", methode: http.MethodPost,
			weg: "/karten/2/bild", rumpf: `{"ziel":"/tmp/z.png"}`, erfolg: http.StatusCreated,
		},
		{
			// **Der Druckbogen** (#1291). Er hängt am Set und legt eine Datei
			// an — `POST /sets/{id}/bogen`, wie die Sicherung eine Zeile
			// darüber.
			name: "Bogen legen", methode: http.MethodPost,
			weg: "/sets/1/bogen", rumpf: `{"ziel":"/tmp/b.pdf"}`, erfolg: http.StatusCreated,
		},
		{
			// **Die Exportvorlagen** (#1291). Die Liste hängt am Set, denn sie
			// richtet sich nach dessen Spiel.
			name: "Vorlagen lesen", methode: http.MethodGet,
			weg: "/sets/1/vorlagen", erfolg: http.StatusOK,
		},
		{
			name: "Vorlage ausführen", methode: http.MethodPost, weg: "/sets/1/vorlagen",
			rumpf:  `{"paket":"probe-liste.mse-export-template","ziel":"/tmp/r.txt"}`,
			erfolg: http.StatusCreated,
		},
		{
			// **Ein Kartenbild von der Platte** (#1291) — am Feld, denn das
			// Bildfeld ist eines von mehreren.
			name: "Kartenbild setzen", methode: http.MethodPut,
			weg: "/sets/1/karten/2/felder/kunst/bild", rumpf: `{"pfad":"/tmp/q.png"}`,
			erfolg: http.StatusOK,
		},
		{
			// **MTG JSON schreiben** (#1291) — am Set, wie die Sicherung.
			name: "MTG JSON schreiben", methode: http.MethodPost,
			weg: "/sets/1/mtgjson", rumpf: `{"spiel":"magic","ziel":"/tmp/r.json"}`,
			erfolg: http.StatusCreated,
		},
		{
			// **`POST` und trotzdem lesend** (#1070): ein Ausdruck ist kein
			// Wegbestandteil, geändert wird nichts.
			name: "Skript auswerten", methode: http.MethodPost,
			weg: "/karten/2/skript", rumpf: `{"ausdruck":"1 + 1"}`, erfolg: http.StatusOK,
		},
		{
			// **`GET` und nicht `POST`** (#1164): eine Aufzählung ohne Eingabe —
			// anders als beim Ausdruck darüber gibt es nichts zu übermitteln.
			name: "Skriptnamen lesen", methode: http.MethodGet,
			weg: "/karten/2/skriptnamen", erfolg: http.StatusOK,
		},
		{
			// **Die Paketwerkstatt** (#1164) — drei Wege, einer davon schreibend.
			name: "Paketdateien lesen", methode: http.MethodGet,
			weg: "/pakete/probe.mse-style/dateien", erfolg: http.StatusOK,
		},
		{
			name: "Paketdatei lesen", methode: http.MethodGet,
			weg: "/pakete/probe.mse-style/dateien/style", erfolg: http.StatusOK,
		},
		{
			// **`PUT` und trotzdem kein blindes Setzen**: der Stand ist die
			// Bedingung, nicht der Inhalt.
			name: "Paketdatei schreiben", methode: http.MethodPut,
			weg:   "/pakete/probe.mse-style/dateien/style",
			rumpf: `{"inhalt":"x","stand":"s"}`, erfolg: http.StatusOK,
		},
		{
			// **Die Symbolpalette** (#1391) — sie hängt an einer Karte und
			// ihrer Vorlage, nicht an einer Zeichnung; deshalb steht sie
			// unter `/karten/…`.
			name: "Symbolpalette lesen", methode: http.MethodGet,
			//
			// **Ein Feldname ohne Leerzeichen**: der übliche heißt
			// `casting cost`, und der gehört in der Adresse kodiert. Dieser
			// Lauf baut seine Anfragen roh zusammen; ein Leerzeichen darin
			// zerbricht sie, bevor der Dienst sie sieht.
			weg: "/karten/2/symbolpalette/name", erfolg: http.StatusOK,
		},
		{
			// **Der Symboleditor** (#1391) — lesen, ablegen, rechnen.
			//
			// **Der Name der Rechnung steht im Weg**, damit ein
			// Zugriffsprotokoll sagt, was geschah; die Argumente stehen im
			// Rumpf, weil ein Symbol verschachtelt ist.
			name: "Symbol lesen", methode: http.MethodGet,
			weg: "/sets/1/symbol", erfolg: http.StatusOK,
		},
		{
			name: "Symbol ablegen", methode: http.MethodPut,
			weg:   "/sets/1/symbol",
			rumpf: `{"teile":[],"unter":"probe.mse-symbol"}`, erfolg: http.StatusOK,
		},
		{
			name: "Symbol rechnen", methode: http.MethodPost,
			weg:   "/symbol/grundform",
			rumpf: `{"art":"Rechteck"}`, erfolg: http.StatusOK,
		},
		{
			// **`POST` und trotzdem lesend** (#1324, Werkstück B). Der zu
			// prüfende Text steht im Rumpf — er kann mehrere hundert Zeilen
			// haben, und in einer Adresse hätte er nichts zu suchen. Geändert
			// wird dabei nichts; die Methode allein sagt das hier nicht, und
			// deshalb steht es in `FAEHIGKEITEN.md` und in der Erklärung.
			//
			// **Und der Weg hängt an keinem Paket**: geprüft wird, was
			// mitkommt, nicht was installiert ist.
			// **Und er kann nicht scheitern** — `Werkstatt.Pruefen` hat kein
			// Fehlerergebnis, weil der zu prüfende Text mitkommt: nichts
			// wird gelesen, nichts geschrieben. Der Lauf dreht die Zusage um
			// und prüft, dass trotz feindlicher Naht die Antwort kommt.
			name: "Paketdatei prüfen", methode: http.MethodPost,
			weg:   "/paketdateien/pruefung",
			rumpf: `{"inhalt":"mse version: 0.3.8\n"}`, erfolg: http.StatusOK,
			ohneFehlerweg: true,
		},
		{
			// **Die Schlüsselwörter eines Sets** (#1062) — sie hängen am Set und
			// an keiner Karte.
			name: "Schlüsselwörter lesen", methode: http.MethodGet,
			weg: "/sets/1/schluesselwoerter", erfolg: http.StatusOK,
		},
		{
			name: "Schlüsselwort anlegen", methode: http.MethodPost,
			weg: "/sets/1/schluesselwoerter", rumpf: `{}`, erfolg: http.StatusCreated,
		},
		{
			// **`PUT` und nicht `POST`**: gesetzt wird ein Feld auf einen Wert.
			name: "Schlüsselwort setzen", methode: http.MethodPut,
			weg:   "/sets/1/schluesselwoerter/2/felder/muster",
			rumpf: `{"wert":"Beben"}`, erfolg: http.StatusOK,
		},
		{
			name: "Schlüsselwort löschen", methode: http.MethodDelete,
			weg: "/sets/1/schluesselwoerter/2", erfolg: http.StatusOK,
		},
		{
			// **Die Packarten eines Sets** (#889) — lesend.
			name: "Packarten lesen", methode: http.MethodGet,
			weg: "/sets/1/packarten", erfolg: http.StatusOK,
		},
		{
			// **`POST` und nicht `GET`**: es ist kein Abruf einer Ressource,
			// sondern eine **Erzeugung**. Der Samen macht sie wiederholbar, nicht
			// zwischenspeicherbar.
			name: "Pack ziehen", methode: http.MethodPost,
			weg: "/sets/1/packarten/alles/ziehung", erfolg: http.StatusOK,
		},
		{
			// **Die Auswahllisten eines Feldes** (#1074) — lesend, ohne Set:
			// welche Wörter ein Feld anbietet, hängt an der Karte.
			name: "Wortlisten lesen", methode: http.MethodGet,
			weg: "/karten/2/felder/type/wortlisten", erfolg: http.StatusOK,
		},
		{
			// **`POST` und nicht `PUT`** (#1074): ein Präfix wird umgeschaltet,
			// zweimal dasselbe gibt nicht dasselbe Ergebnis.
			name: "Wort wählen", methode: http.MethodPost,
			weg:   "/sets/1/karten/2/felder/type/wortwahl",
			rumpf: `{"bereich":0,"weg":[0]}`, erfolg: http.StatusOK,
		},
		{
			name: "Scryfall zeigen", methode: http.MethodGet,
			weg: "/sets/1/karten/2/scryfall?name=Prüfstein", erfolg: http.StatusOK,
		},
		{
			name: "Scryfall übernehmen", methode: http.MethodPost,
			weg:   "/sets/1/karten/2/scryfall",
			rumpf: `{"name":"Prüfstein","felder":["rule text"]}`, erfolg: http.StatusOK,
		},
		{
			name: "Scryfall-Bild einsetzen", methode: http.MethodPost,
			weg:   "/sets/1/karten/2/scryfall/bild",
			rumpf: `{"feld":"image","name":"Prüfstein","art":"normal"}`, erfolg: http.StatusOK,
		},
		{
			// **Ohne Kartenkennung** (#1032): die Karte gibt es noch nicht.
			name: "Scryfall-Karte anlegen", methode: http.MethodPost,
			weg:   "/sets/1/karten/scryfall",
			rumpf: `{"name":"Prüfstein"}`, erfolg: http.StatusCreated,
		},
		{
			// **Ohne Setkennung** (#1033): das Set entsteht hier.
			name: "Scryfall-Satz einlesen", methode: http.MethodPost,
			weg:   "/sets/scryfall",
			rumpf: `{"code":"lea","spiel":"magic"}`, erfolg: http.StatusCreated,
		},
		{
			name: "Sammeldaten zeigen", methode: http.MethodGet,
			weg: "/scryfall/sammeldaten", erfolg: http.StatusOK,
		},
		{
			name: "Sammeldaten holen", methode: http.MethodPost,
			weg: "/scryfall/sammeldaten", erfolg: http.StatusOK,
		},
		{
			name: "Scryfall-Regeln lesen", methode: http.MethodGet,
			weg: "/scryfall/regeln?name=Prüfstein", erfolg: http.StatusOK,
		},
		{
			name: "Manakosten prüfen", methode: http.MethodGet,
			weg: "/scryfall/manakosten?kosten=%7B2%7D%7BR%7D", erfolg: http.StatusOK,
		},
		// **Der Lebenszyklus einer Karte und der Volltext**
		// (#1107).
		//
		// Sie standen am MCP-Server seit #591 und #886 und hier nicht — ohne dass
		// jemand entschieden hätte, dass sie hier nicht hingehören.
		{
			name: "Karte anlegen", methode: http.MethodPost,
			weg: "/sets/1/karten", erfolg: http.StatusCreated,
		},
		{
			// **`/kopie` als eigener Weg** und nicht `POST /sets/1/karten` mit
			// einer Vorlage im Rumpf: eine Adresse, die zwei Dinge tut, je
			// nachdem was im Rumpf steht, lässt sich nicht lesen.
			name: "Karte kopieren", methode: http.MethodPost,
			weg: "/sets/1/karten/2/kopie", erfolg: http.StatusCreated,
		},
		{
			name: "Karte löschen", methode: http.MethodDelete,
			weg: "/sets/1/karten/2", erfolg: http.StatusOK,
		},
		{
			// **Ein Feld auf vielen Karten** (#1390) — `POST` und nicht
			// `PUT`: `PUT` sagt „dieser eine Ort trägt danach diesen Wert",
			// und hier tragen ihn viele.
			name: "Felder setzen", methode: http.MethodPost,
			weg:   "/sets/1/felder",
			rumpf: `{"karteIDs":[2],"feld":"artist","wert":"X"}`, erfolg: http.StatusOK,
		},
		{
			// **Die Vorschau dazu** (U6, #1417) — `POST`, weil sie einen Rumpf
			// braucht; sie schreibt nichts.
			name: "Felder vorschauen", methode: http.MethodPost,
			weg:   "/sets/1/felder/vorschau",
			rumpf: `{"karteIDs":[2],"feld":"artist","wert":"X"}`, erfolg: http.StatusOK,
		},
		{
			name: "Karten tauschen", methode: http.MethodPost,
			weg: "/sets/1/karten/tausch", rumpf: `{"eine":"2","andere":"3"}`,
			erfolg: http.StatusOK,
		},
		{
			name: "Felder lesen", methode: http.MethodGet,
			weg: "/sets/1/felder", erfolg: http.StatusOK,
		},
		{
			// **Der Verlauf** (U6, #1417) — der jüngste Schritt zuerst.
			name: "Verlauf lesen", methode: http.MethodGet,
			weg: "/sets/1/verlauf", erfolg: http.StatusOK,
		},
		{
			name: "Volltext suchen", methode: http.MethodGet,
			weg: "/suche?text=Feuerball", erfolg: http.StatusOK,
		},
		// **Die Setangaben, die Vorbehalte und die mitgelieferten Sets**
		// (#1107).
		{
			name: "Setangaben lesen", methode: http.MethodGet,
			weg: "/sets/1/angaben", erfolg: http.StatusOK,
		},
		{
			name: "Setangabe setzen", methode: http.MethodPut,
			weg: "/sets/1/angaben/title", rumpf: `{"wert":"Probe"}`, erfolg: http.StatusOK,
		},
		{
			name: "Vorbehalte lesen", methode: http.MethodGet,
			weg: "/sets/1/vorbehalte", erfolg: http.StatusOK,
		},
		{
			// **Ohne Set im Weg**: die mitgelieferten Sets sind keine Sets des
			// Bestands, sondern Dateien im Bündel.
			name: "Beispielsets nennen", methode: http.MethodGet,
			weg: "/beispielsets", erfolg: http.StatusOK,
		},

		{name: "eine Datei ansehen", methode: http.MethodGet, weg: "/dateien/ansicht?pfad=/tmp/liste.txt", erfolg: http.StatusOK}, // UX-13
		// **Der portable Umschlag** (N4, #1457). `POST /umschlag` legt einen
		// an, `GET /umschlag` sieht hinein, und die **Übernahme** ist ein
		// eigener Weg — sie ist die fachliche Freigabe, die das Issue
		// verlangt, und ein `POST /umschlag`, das beides täte, nähme sie
		// entgegen, nachdem sie gegenstandslos geworden wäre.
		{
			name: "Umschlag schnüren", methode: http.MethodPost, weg: "/umschlag",
			rumpf: `{"setID":"1","ziel":"/tmp/x.msepackage"}`, erfolg: http.StatusCreated,
		},
		{
			name: "in einen Umschlag sehen", methode: http.MethodGet,
			weg: "/umschlag?pfad=/tmp/x.msepackage", erfolg: http.StatusOK,
		},
		{
			name: "einen Umschlag übernehmen", methode: http.MethodPost,
			weg: "/umschlag/uebernahme", rumpf: `{"pfad":"/tmp/x.msepackage"}`,
			erfolg: http.StatusCreated,
		},
		// **Der Netzweg** (N5, #1571) — rein lesend, und er hängt an keinem
		// Set: deshalb `/netz` und nicht `/…/netz`.
		{
			name: "den Netzweg ansehen", methode: http.MethodGet,
			weg: "/netz", erfolg: http.StatusOK,
		},
		// **Die Regelfreigabe** (G4, #1461) — rein lesend, und sie hängt an
		// keinem Set: sie rechnet über ein Regelprofil und eine Strenge.
		//
		// **Die Angaben stehen in der Abfrage** und nicht im Pfad: sie sind
		// eine Angabe zur Frage und kein Gegenstand, den man adressiert.
		{
			name: "die Regelfreigabe ansehen", methode: http.MethodGet,
			weg:    "/regelfreigabe?spiel=magic&strenge=sandkasten&sitze=2",
			erfolg: http.StatusOK,
		},
		// **Wer arbeitet, und der Wechsel dorthin** — Werkstück 17 von Phase
		// 4.12 (#1400).
		//
		// **`PUT /nutzer` und nicht `PUT /nutzer/{id}`** — derselbe Grund wie
		// bei `PUT /profil`: der Weg benennt den **Zeiger**, nicht den Nutzer.
		{
			name: "wer arbeitet", methode: http.MethodGet,
			weg: "/werarbeitet", erfolg: http.StatusOK,
		},
		{
			name: "der Arbeitskontext", methode: http.MethodGet,
			weg: "/kontext", erfolg: http.StatusOK,
		},
		{
			name: "den arbeitenden Nutzer wechseln", methode: http.MethodPut,
			weg: "/nutzer", rumpf: `{"nutzerID":"1"}`, erfolg: http.StatusOK,
		},
	}, append(append(endpunkteAmSet(), endpunktePartien()...), endpunkteBestandssicherung()...)...)
}
