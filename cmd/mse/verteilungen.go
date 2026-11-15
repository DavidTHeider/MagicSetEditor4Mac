// **Drei Verteilungen für Befehle, die den Setpfad nicht brauchen**
// (#1165).
//
// # Warum sie hier stehen und nicht in `lauf.go`
//
// Jene Datei stand an der Zeilengrenze aus §2 — und der Schnitt folgt der
// Frage: dort die Befehle **am Set**, hier die am **Besitz**, an einer
// **Paketdatei** und an einem **fremden Dienst**. Alle drei haben gemeinsam,
// dass sie mit dem geöffneten Set nichts zu tun haben; ein Exemplar hängt an
// keinem ([0052](../../docs/entscheidungen/0052-die-sammlung-ist-nicht-das-set.md)),
// eine Paketdatei liegt auf der Platte, und Scryfall kennt den Bestand nicht.
package main

import "context"

// sammlungsbefehle beantwortet die Befehle am Besitz und am Druckbezug.
//
// **Der zweite Rückgabewert sagt, ob es einer war.** Ohne ihn müsste die
// Verteilung darüber die Namen ein zweites Mal führen — und zwei Aufzählungen
// derselben Sache laufen auseinander.
func sammlungsbefehle(
	ctx context.Context, befehl string, rest []string, bestandsort string,
	an *schreiber,
) (code int, getroffen bool) {
	switch befehl {
	case "sammlung":
		return sammlungsbefehl(ctx, bestandsort, an, rest), true

	case "masse-aendern":
		return masseAendernbefehl(ctx, bestandsort, an, rest), true

	case "masse-loeschen":
		return masseLoeschenbefehl(ctx, bestandsort, an, rest), true

	case "preise-alle":
		return preiseAllebefehl(ctx, bestandsort, an, rest), true

	// **Die Kartenskripte des Spielpakets** (#1337).
	case "kartenskripte":
		return kartenskriptebefehl(ctx, bestandsort, an, rest), true

	case "kartenskript":
		return kartenskriptbefehl(ctx, bestandsort, an, rest), true

	// **Die Auszählungsdimensionen des Spielpakets** (#1338). Ohne Namen
	// steht das Angebot da, mit Namen die Auszählung.
	case "auszaehlung":
		return auszaehlungbefehl(ctx, bestandsort, an, rest), true

	// **Decks und Formate** (#1293, Vermerk 0061). Ein Deck verweist auf
	// Karten, es kopiert sie nicht — und ein Format ist ein Datensatz, keine
	// eingebaute Liste.
	case "decks":
		return decksbefehl(ctx, bestandsort, an, rest), true

	case "deck":
		return deckbefehl(ctx, bestandsort, an, rest), true

	case "formate", "format":
		return formatbefehl(ctx, bestandsort, an, rest), true

	// **Turniere** (#1294, Vermerk 0062). Ein Turnier ist ein Verlauf über
	// die Zeit; die Rangliste wird daraus gerechnet und nirgends abgelegt.
	case "turniere":
		return turnierbefehl(ctx, bestandsort, an, nil), true

	case "turnier":
		return turnierbefehl(ctx, bestandsort, an, rest), true

	// **Die Symbolbibliothek** (DOM-02, #1697) — Symbole des Profils, die
	// sich in jedes Set einsetzen lassen.
	case "symbolbibliothek":
		return symbolbibliothekbefehl(ctx, bestandsort, an, rest), true

	// **Die eigenen Übersetzungen** (LOC-02, #1688).
	case "sprachpaket":
		return sprachpaketbefehl(ctx, bestandsort, an, rest), true

	// **Die Autoersetzungen eines Spiels** (#1341, Vermerk 0060).
	case "autoersetzung":
		return autoersetzungbefehl(ctx, bestandsort, an, rest), true

	case "besitz":
		return besitzbefehl(ctx, bestandsort, an, rest), true

	case "lagerorte":
		return lagerortebefehl(ctx, bestandsort, an), true

	case "dubletten":
		return dublettenbefehl(ctx, bestandsort, an, rest), true

	case "wuensche":
		return wuenschebefehl(ctx, bestandsort, an), true

	case "sammlung-ausgeben":
		return sammlungAusgebenbefehl(ctx, bestandsort, an, rest), true

	case "sammlung-einlesen":
		return sammlungEinlesenbefehl(ctx, bestandsort, an, rest), true

	case "sammlung-vorschau":
		return sammlungVorschaubefehl(ctx, bestandsort, an), true

	case "liste-deuten":
		return listeDeutenbefehl(ctx, bestandsort, an), true

	case "liste-einlesen":
		return listeEinlesenbefehl(ctx, bestandsort, an, rest), true

	case "liste-vorschau":
		return listeVorschaubefehl(ctx, bestandsort, an, rest), true

	case "exemplar":
		return exemplarbefehl(ctx, bestandsort, an, rest), true

	case "exemplar-aendern":
		return exemplarAendernbefehl(ctx, bestandsort, an, rest), true

	case "exemplar-loeschen":
		return exemplarLoeschenbefehl(ctx, bestandsort, an, rest), true

	case "preise":
		return preisebefehl(ctx, bestandsort, an, rest), true

	case "preise-holen":
		return preiseHolenbefehl(ctx, bestandsort, an, rest), true

	case "preise-bulk":
		return preisbulkbefehl(ctx, bestandsort, an, rest), true

	case "ki-modelle":
		return kimodellebefehl(ctx, bestandsort, an, rest), true
	case "ki-modell-pruefen":
		return kimodellpruefenbefehl(ctx, bestandsort, an, rest), true
	case "ki-modell-vergessen":
		return kimodellvergessenbefehl(ctx, bestandsort, an, rest), true
	case "ki-vorschlag":
		return kivorschlagbefehl(ctx, bestandsort, an, rest), true
	case "ki-einschaetzung":
		return kieinschaetzungbefehl(ctx, bestandsort, an, rest), true
	case "ki-batch":
		return kibatchbefehl(ctx, bestandsort, an, rest), true
	case "ki-gruppe":
		return kigruppebefehl(ctx, bestandsort, an, rest), true
	case "ki-erkennung":
		return kierkennungbefehl(ctx, bestandsort, an, rest), true
	case "anbieterstand":
		return anbieterstandbefehl(ctx, bestandsort, an, rest), true
	case "fremdverlauf":
		return fremdverlaufbefehl(ctx, bestandsort, an, rest), true
	case "markt":
		return marktbefehl(ctx, bestandsort, an, rest), true
	case "preisverlauf":
		return preisverlaufbefehl(ctx, bestandsort, an, rest), true

	case "portfolio":
		return portfoliobefehl(ctx, bestandsort, an, rest), true

	// **Der Verweis auf einen echten Druck** (#1167).
	//
	// **Hier und nicht bei den Fremdbefehlen**: er hängt am **Bestand** — an
	// einer Karte, die es gibt —, und die Fremdbefehle bekommen bewusst
	// keinen Bestandsort. Was ihn mit der Sammlung verbindet, ist der
	// Bezugspunkt: beide zeigen auf einen **Druck**.
	case "verweis":
		return verweisbefehl(ctx, bestandsort, an, rest), true

	case "verweise":
		return setverweisbefehl(ctx, bestandsort, an, rest), true

	// **Marken an Karten** (#1208) — dieselbe Bauform wie `verweis`: eine
	// Sache mit drei Zeitformen, nicht drei Sachen.
	case "marke":
		return markenbefehl(ctx, bestandsort, an, rest), true

	case "marken":
		return setmarkenbefehl(ctx, bestandsort, an, rest), true

	// **Und an jedem anderen Fachobjekt** (UX-12, #1684).
	case "objektmarke":
		return objektmarkebefehl(ctx, bestandsort, an, rest), true

	case "objektmarken":
		return objektmarkenbefehl(ctx, bestandsort, an, rest), true

	case "marken-filtern":
		return markenfilterbefehl(ctx, bestandsort, an, rest), true

	case "marke-vergeben":
		return markevergabebefehl(ctx, bestandsort, an, rest), true

	case "markenkatalog":
		return markenkatalogbefehl(ctx, bestandsort, an), true

	case "marke-umbenennen":
		return markeumbenennbefehl(ctx, bestandsort, an, rest), true

	case "marke-tilgen":
		return marketilgbefehl(ctx, bestandsort, an, rest), true

	// **Die Zwischenablage** (#1168) — ohne Fenster sind es die Ströme.
	case "ablage-kopieren":
		return ablagekopierbefehl(ctx, bestandsort, an, rest), true

	case "ablage-ausschneiden":
		return ablageschnittbefehl(ctx, bestandsort, an, rest), true

	case "ablage-einfuegen":
		return ablageeinfuegebefehl(ctx, bestandsort, an, rest), true

	// **Suchen und Ersetzen** (#1168) — die Vorschau ist die Vorgabe.
	case "ersetzen":
		return ersetzbefehl(ctx, bestandsort, an, rest), true

	// **Die Gestalt der Kartenliste** (#1168) — Spalten und Ordnung.
	case "kartenliste":
		return kartenlistenbefehl(ctx, bestandsort, an, rest), true

	// **Der Schnellfilter** (U5, #1416) — dieselbe Sprache wie in MSE2.
	case "kartenfilter":
		return kartenfilterbefehl(ctx, bestandsort, an, rest), true
	}

	return 0, false
}

// werkstattbefehle beantwortet die drei Befehle der Paketwerkstatt.
//
// **Ohne Bestandsort**, und das ist die Aussage: eine Paketdatei liegt auf der
// Platte. Wer ihr eine Datenbank mitgäbe, machte eine Abhängigkeit sichtbar,
// die es nicht gibt (#1164).
func werkstattbefehle(
	ctx context.Context, befehl string, rest []string, an *schreiber,
) (code int, getroffen bool) {
	switch befehl {
	case "symbol-rechnen":
		return symbolRechnenBefehl(an, rest), true

	case "paketdateien":
		return paketdateienbefehl(ctx, an, rest), true

	case "paketdatei":
		return paketdateibefehl(ctx, an, rest), true

	case "paketdatei-pruefen":
		return paketdateipruefenbefehl(ctx, an, rest), true

	case "paketdatei-setzen":
		return paketdateisetzenbefehl(ctx, an, rest), true

	// **Der Dateibaum** (FMT-02, #1691).
	case "paketdatei-anlegen":
		return paketdateianlegenbefehl(ctx, an, rest), true

	case "paketdatei-einfuegen":
		return paketdateieinfuegenbefehl(ctx, an, rest), true

	case "paketdatei-umbenennen":
		return paketdateiumbenennenbefehl(ctx, an, rest), true

	case "paketdatei-entfernen":
		return paketdateientfernenbefehl(ctx, an, rest), true
	}

	return 0, false
}

// fremdbefehle beantwortet die vier Befehle ohne Bestand.
//
// **Ohne Bestandsort, und das ist die Aussage**: sie fragen einen fremden
// Dienst und fassen die Datenbank nicht an. Ihn zu öffnen, nur um ihn nicht zu
// benutzen, machte sie von einem Zustand abhängig, den sie nicht haben.
func fremdbefehle(
	ctx context.Context, befehl string, rest []string, bestandsort string,
	an *schreiber,
) (code int, getroffen bool) {
	switch befehl {
	case "scryfall-regeln":
		return scryfallregelnbefehl(ctx, bestandsort, an, rest), true

	case "scryfall-mana":
		return scryfallmanabefehl(ctx, bestandsort, an, rest), true

	case "aktualisierung":
		return aktualisierungsbefehl(ctx, an, rest), true

	// **Der Platz steht bei den Befehlen ohne Bestand**, und das ist die
	// Aussage: gezählt werden Dateien, und die liegen auch dann da, wenn der
	// Bestand nicht aufgeht (Werkstück 4 von Phase 4.13, #1352).
	case "platz":
		return platzbefehl(ctx, an, bestandsort, rest), true

	case "scryfall-suche":
		return scryfallsuchbefehl(ctx, bestandsort, an, rest), true

	case "scryfall-sammeldaten":
		return scryfallsammelbefehl(ctx, bestandsort, an, rest), true
	}

	return 0, false
}
