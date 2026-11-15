// **Die Wege, die an keinem Set hängen** —
// #1169.
//
// **Eine eigene Datei**, weil `endpunktliste_test.go` mit der
// Bestandssicherung über die Zeilengrenze aus §2 ging. Der Schnitt folgt der
// Frage und nicht der Zeilenzahl: dort stehen die Wege **am Bestand** — Sets,
// Karten, Stile, Pakete —, hier die, die **daneben** stehen: was jemand
// besitzt, was ein Druck kostet, welchen Druck eine Karte meint, was in der
// Zwischenablage liegt.

package rest_test

import "net/http"

// endpunkteDaneben sind die Wege, die an keinem Set hängen.
//
// **Die Nutzerwege stehen seit dem 2026-09-05 nebenan** — mit den drei
// Kontoarten kam diese Datei über die Zeilengrenze aus §2, und der Schnitt
// folgt der Frage: hier steht, was am **Bestand** vorbei geht, dort, wer ihn
// benutzt.
func endpunkteDaneben() []endpunkt {
	return append(append(append(append(append(endpunkteNutzer(), endpunkteBetrieb()...), endpunkteMarken()...),
		endpunktePaketpflege()...), []endpunkt{
		// **Der Platz** — Werkstück 4 von Phase 4.13 (#1352). Er hängt an
		// keinem Set und an keinem Profil: gezählt werden Dateien neben dem
		// Bestand.
		{
			name: "Platz lesen", methode: http.MethodGet, weg: "/platz",
			erfolg: http.StatusOK,
		},
		{
			name: "Platz freigeben", methode: http.MethodDelete,
			weg: "/platz/vorsicherungen", erfolg: http.StatusOK,
		},

		// **MTG JSON** (#1291). Das Lesen legt Sets erst an und hängt deshalb
		// an keinem; die Spieleliste ist eine Auskunft über das Programm.
		{
			name: "MTG JSON lesen", methode: http.MethodPost, weg: "/mtgjson",
			rumpf: `{"pfad":"/tmp/alle.json","spiel":"magic"}`, erfolg: http.StatusCreated,
		},
		{
			name: "MTG-JSON-Spiele", methode: http.MethodGet,
			weg: "/mtgjson/spiele", erfolg: http.StatusOK, ohneFehlerweg: true,
		},

		// **Die Grenzen** (#1429): eine Auskunft über den Prozess, an keinem Set.
		{
			name: "Grenzen", methode: http.MethodGet,
			weg: "/grenzen", erfolg: http.StatusOK,
		},

		// **Die Schnittstellen** (OPS-09, #1907): rein lesend.
		{
			name: "Schnittstellen", methode: http.MethodGet,
			weg: "/schnittstellen", erfolg: http.StatusOK,
		},

		// **Die Schalter** (OPS-10, #1908): rein lesend.
		{
			name: "Schalter", methode: http.MethodGet,
			weg: "/schalter", erfolg: http.StatusOK,
		},

		// **Die Betriebswege** (O9 und O6) stehen in `endpunktebetrieb_test.go`.
		// **Die Paketpflege** (#1291, FMT-02 #1691) steht in
		// `endpunktepaketpflege_test.go`.
		{
			name: "MSE2-Orte", methode: http.MethodGet,
			weg: "/mse2/orte", erfolg: http.StatusOK, ohneFehlerweg: true,
		},

		{
			// **Der Optionswert hängt an der Vorlage, nicht am Set** (#1291):
			// er gilt über Sets hinweg.
			name: "Vorlagenoption setzen", methode: http.MethodPut,
			weg:   "/vorlagen/probe-liste.mse-export-template/optionen/trenner",
			rumpf: `{"wert":"; "}`, erfolg: http.StatusOK,
		},

		// **Die Sammlung** (#1165). Sie steht **neben** `/sets`: was jemand
		// besitzt, ist ein Druck und kein Entwurf.
		{
			name: "Sammlung lesen", methode: http.MethodGet,
			weg: "/sammlung", erfolg: http.StatusOK,
		},
		{
			name: "Sammlung eines Drucks lesen", methode: http.MethodGet,
			weg: "/sammlung/drucke/probe-druck", erfolg: http.StatusOK,
		},
		{
			name: "Exemplar anlegen", methode: http.MethodPost, weg: "/sammlung",
			rumpf: `{"printing":"probe-druck","menge":1}`, erfolg: http.StatusCreated,
		},
		{
			name: "Exemplar ändern", methode: http.MethodPut, weg: "/sammlung/1",
			rumpf:  `{"printing":"probe-druck","menge":2,"fassung":1}`,
			erfolg: http.StatusOK,
		},
		{
			// **Die Fassung steht in der Abfrage**, weil ein `DELETE` keinen
			// Rumpf hat (Befund 10 des Nachlaufs über Phase 3.4).
			name: "Exemplar löschen", methode: http.MethodDelete,
			weg: "/sammlung/1?fassung=1", erfolg: http.StatusOK,
		},

		// **Die Listeneinfuhr** (#1292). Beide `POST`, obwohl das Deuten
		// nichts ändert: eine Liste von dreihundert Zeilen passt in keine
		// Adresse.
		{
			name: "Liste deuten", methode: http.MethodPost,
			weg:   "/sammlung/liste/deuten",
			rumpf: `{"text":"4 Lightning Bolt"}`, erfolg: http.StatusOK,
		},
		{
			name: "Liste einlesen", methode: http.MethodPost,
			weg:    "/sammlung/liste/einlesen",
			rumpf:  `{"zeilen":[{"printing":"probe-druck","menge":1}]}`,
			erfolg: http.StatusOK,
		},

		// **Die Massenvorgänge** (#1292). Alle `POST`: die Zeilen samt
		// Fassungen gehören in den Rumpf, und ein `DELETE` hat keinen.
		{
			name: "Masse ändern", methode: http.MethodPost, weg: "/sammlung/masse",
			rumpf:  `{"zeilen":[{"id":1,"fassung":1}],"lagerort":"Kiste 4"}`,
			erfolg: http.StatusOK,
		},
		{
			name: "Masse löschen", methode: http.MethodPost,
			weg:    "/sammlung/masse/loeschen",
			rumpf:  `{"zeilen":[{"id":1,"fassung":1}]}`,
			erfolg: http.StatusOK,
		},
		{
			name: "Preise in Masse", methode: http.MethodPost,
			weg:    "/sammlung/masse/preise",
			rumpf:  `{"hoechstensStunden":24}`,
			erfolg: http.StatusOK,
		},

		// **Die vier Fragen an eine geordnete Sammlung** (#1292, Vermerk
		// 0057). Alle `GET`: sie lesen.
		{
			name: "Besitz eines Drucks", methode: http.MethodGet,
			weg: "/sammlung/besitz/probe-druck", erfolg: http.StatusOK,
		},
		{
			name: "Lagerorte lesen", methode: http.MethodGet,
			weg: "/sammlung/lagerorte", erfolg: http.StatusOK,
		},

		// **Die Kartenskripte eines Sets** (#1337). Der zweite legt Karten an
		// und gibt deshalb 201.
		{
			name: "Kartenskripte lesen", methode: http.MethodGet,
			weg: "/sets/1/kartenskripte", erfolg: http.StatusOK,
		},
		{
			name: "Kartenskript ausführen", methode: http.MethodPost,
			weg: "/sets/1/kartenskripte", rumpf: `{"name":"Zwei Farben"}`,
			erfolg: http.StatusCreated,
		},

		// **Die Auszählungsdimensionen eines Sets** (#1338). Beide `GET`:
		// sie rechnen über jede Karte und ändern nichts.
		{
			name: "Auszähldimensionen lesen", methode: http.MethodGet,
			weg: "/sets/1/auszaehlung", erfolg: http.StatusOK,
		},
		{
			name: "Auszählung lesen", methode: http.MethodGet,
			weg: "/sets/1/auszaehlung/farbe", erfolg: http.StatusOK,
		},

		// **Die Parameterhilfen eines Schlüsselworts** (#1342).
		{
			name: "Parameterhilfen lesen", methode: http.MethodGet,
			weg: "/sets/1/schluesselwoerter/2/parameter", erfolg: http.StatusOK,
		},

		// **Ein Bild als Symbol nachzeichnen** (#1340).
		{
			name: "Symbol nachzeichnen", methode: http.MethodPost,
			weg: "/sets/1/symbol-nachzeichnen", rumpf: `{"bildpfad":"/tmp/k.png"}`,
			erfolg: http.StatusOK,
		},

		// **Die Autoersetzungen eines Spiels** (#1341, Vermerk 0060).
		{
			name: "Autoersetzungen lesen", methode: http.MethodGet,
			weg: "/sets/1/autoersetzungen", erfolg: http.StatusOK,
		},
		{
			name: "Autoersetzung setzen", methode: http.MethodPut,
			weg: "/sets/1/autoersetzungen/(W)", rumpf: `{"ersatz":"x"}`,
			erfolg: http.StatusNoContent,
		},
		{
			name: "Autoersetzung löschen", methode: http.MethodDelete,
			weg: "/sets/1/autoersetzungen/(W)", erfolg: http.StatusNoContent,
		},
		{
			name: "Autoersetzungen zurücksetzen", methode: http.MethodDelete,
			weg: "/sets/1/autoersetzungen", erfolg: http.StatusNoContent,
		},
		{
			name: "Autoersetzung schalten", methode: http.MethodPut,
			weg: "/sets/1/autoersetzungen", rumpf: `{"an":false}`,
			erfolg: http.StatusNoContent,
		},
		{
			name: "Dubletten lesen", methode: http.MethodGet,
			weg: "/sammlung/dubletten?behalten=1", erfolg: http.StatusOK,
		},
		{
			// **Ohne `behalten` gilt 1.** Der Zweig bliebe sonst unbetreten.
			name: "Dubletten ohne Zahl", methode: http.MethodGet,
			weg: "/sammlung/dubletten", erfolg: http.StatusOK,
		},
		{
			name: "Wunschliste lesen", methode: http.MethodGet,
			weg: "/sammlung/wuensche", erfolg: http.StatusOK,
		},

		// **Die Ausfuhr** (#1292). `GET` beim Ausgeben, weil die Form die
		// Antwort **auswählt** und nichts ändert.
		{
			name: "Sammlung ausgeben", methode: http.MethodGet,
			weg: "/sammlung/ausgeben?form=csv", erfolg: http.StatusOK,
		},
		{
			// **Ohne `form` gilt `csv`.** Der Zweig bliebe sonst unbetreten —
			// und eine Vorgabe, die kein Lauf nimmt, ist eine Behauptung.
			name: "Sammlung ausgeben ohne Form", methode: http.MethodGet,
			weg: "/sammlung/ausgeben", erfolg: http.StatusOK,
		},
		{
			name: "Sammlung einlesen", methode: http.MethodPost,
			weg:    "/sammlung/einlesen",
			rumpf:  `{"text":"printing,menge\nprobe-druck,1\n"}`,
			erfolg: http.StatusOK,
		},

		// **Die Preise** (#1165). `POST` beim Holen, weil es eine Momentaufnahme
		// erzeugt und dafür einen fremden Dienst fragt.
		{
			// **Die Kartensuche** (#1167) — `GET`, weil sie nichts ändert.
			name: "Scryfall durchsuchen", methode: http.MethodGet,
			weg: "/scryfall/suche?frage=bolt", erfolg: http.StatusOK,
		},
		{
			// **Der Dienstkatalog** (#1440) — `GET`, er schickt nichts.
			name: "Dienstkatalog lesen", methode: http.MethodGet,
			weg: "/dienstkatalog", erfolg: http.StatusOK,
		},
		{
			name: "Preise lesen", methode: http.MethodGet,
			weg: "/drucke/probe-druck/preise", erfolg: http.StatusOK,
		},
		{
			name: "Preisverlauf lesen", methode: http.MethodGet,
			weg: "/drucke/probe-druck/preisverlauf?quelle=scryfall&markt=paper" +
				"&art=retail&waehrung=EUR&finish=nonfoil", erfolg: http.StatusOK,
		},
		{
			name: "Marktangebote lesen", methode: http.MethodGet,
			weg: "/drucke/probe-druck/angebote?quelle=cardtrader", erfolg: http.StatusOK,
		},
		{
			name: "Fremdverlauf holen", methode: http.MethodPost,
			weg: "/drucke/probe-druck/fremdverlauf?quelle=scrydex", erfolg: http.StatusOK,
		},
		{
			name: "Anbieterstand lesen", methode: http.MethodGet,
			weg: "/anbieter/scrydex/stand", erfolg: http.StatusOK,
		},
		{
			name: "Preise holen", methode: http.MethodPost,
			weg: "/drucke/probe-druck/preise", erfolg: http.StatusOK,
		},
		{
			name: "Preise einer ganzen Quelle holen", methode: http.MethodPost,
			weg: "/preise/bulk/mtgjson/heute", erfolg: http.StatusOK,
		},
		{
			name: "Portfolio lesen", methode: http.MethodGet,
			weg: "/portfolio", erfolg: http.StatusOK,
		},

		// **Die Verweise auf echte Drucke** (#1167). `GET` zeigt, `PUT` setzt,
		// `DELETE` löst — und ohne Verweis antwortet `GET` mit `200` und `null`,
		// weil eine Karte ohne Verweis der Normalfall ist und kein Fehler.
		{
			name: "Verweis lesen", methode: http.MethodGet,
			weg: "/karten/2/verweis", erfolg: http.StatusOK,
		},
		{
			name: "Verweis setzen", methode: http.MethodPut,
			weg: "/karten/2/verweis", rumpf: `{"printing":"probe-druck"}`,
			erfolg: http.StatusOK,
		},
		{
			name: "Verweis lösen", methode: http.MethodDelete,
			weg: "/karten/2/verweis", erfolg: http.StatusOK,
		},
		{
			name: "Setverweise lesen", methode: http.MethodGet,
			weg: "/sets/1/verweise", erfolg: http.StatusOK,
		},

		// **Marken an Karten** (#1208). Die Marke steht im **Weg** und nicht
		// im Rumpf: `PUT` und `DELETE` sprechen dieselbe Sache an, und ein
		// `DELETE` hat keinen Rumpf.
		{
			name: "Marken eines Sets lesen", methode: http.MethodGet,
			weg: "/sets/1/marken", erfolg: http.StatusOK,
		},
		{
			name: "Marken einer Karte lesen", methode: http.MethodGet,
			weg: "/sets/1/karten/2/marken", erfolg: http.StatusOK,
		},
		{
			name: "Marke setzen", methode: http.MethodPut,
			weg: "/sets/1/karten/2/marken/fertig", erfolg: http.StatusOK,
		},
		{
			name: "Marke entfernen", methode: http.MethodDelete,
			weg: "/sets/1/karten/2/marken/fertig", erfolg: http.StatusOK,
		},

		// **Die Zwischenablage** (#1168). `POST` bei allen dreien, weil die
		// Kartenliste beziehungsweise der Inhalt in den Rumpf gehört.
		{
			name: "Karten kopieren", methode: http.MethodPost,
			weg: "/sets/1/ablage/kopieren", rumpf: `{"karten":[2]}`,
			erfolg: http.StatusOK,
		},
		{
			name: "Karten ausschneiden", methode: http.MethodPost,
			weg: "/sets/1/ablage/ausschneiden", rumpf: `{"karten":[2]}`,
			erfolg: http.StatusOK,
		},
		{
			name: "Karten einfügen", methode: http.MethodPost,
			weg: "/sets/1/ablage/einfuegen", rumpf: `{"inhalt":"game_name: probe\ncard:\n"}`,
			erfolg: http.StatusOK,
		},

		// **Suchen und Ersetzen** (#1168). `POST` bei beiden, weil die Frage in
		// den Rumpf gehört — die Vorschau ändert trotzdem nichts.
		{
			name: "Ersetzung vorschauen", methode: http.MethodPost,
			weg: "/sets/1/ersetzung/vorschau", rumpf: `{"text":"a","ersatz":"b"}`,
			erfolg: http.StatusOK,
		},
		{
			name: "Ersetzung ausführen", methode: http.MethodPost,
			weg: "/sets/1/ersetzung", rumpf: `{"text":"a","ersatz":"b"}`,
			erfolg: http.StatusOK,
		},

		// **Die Gestalt der Kartenliste** (#1168). `PUT` bei beiden
		// Schreibwegen, weil sie die **ganze** Wahl ersetzen.
		{
			name: "Kartenliste lesen", methode: http.MethodGet,
			weg: "/sets/1/kartenliste", erfolg: http.StatusOK,
		},
		// **Der Schnellfilter** — Werkstück U5 von Phase 5.1 (#1416).
		{
			name: "Karten filtern", methode: http.MethodGet,
			weg: "/sets/1/kartenfilter?anfrage=bear", erfolg: http.StatusOK,
		},
		{
			name: "Listenspalten setzen", methode: http.MethodPut,
			weg:    "/sets/1/kartenliste/spalten",
			rumpf:  `{"spalten":[{"feld":"name","sichtbar":true}]}`,
			erfolg: http.StatusOK,
		},
		{
			name: "Listensortierung setzen", methode: http.MethodPut,
			weg:    "/sets/1/kartenliste/sortierung",
			rumpf:  `{"feld":"name","aufsteigend":true}`,
			erfolg: http.StatusOK,
		},
	}...), kiendpunkte()...)
}
