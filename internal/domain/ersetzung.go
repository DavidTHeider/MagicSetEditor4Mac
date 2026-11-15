package domain

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// **Suchen und Ersetzen über einen Feldwert** — Werkstück 2 von Phase 3.7
// (#1168).
//
// # ported: die Übereinstimmung selbst
//
//	String v = untag(value().value());
//	if (!find.caseSensitive()) v.LowerCase();
//	…
//	if (find.wholeWord()) {
//	  if (!is_word_end(s, pos - 1) || !is_word_end(s, pos + find.findString().size()))
//	    return false;
//	}
//
// `src/gui/value/text.cpp`, Zeilen 1199–1246, Fassung f9d9356 vom 2021-07-06.
// Drei Dinge stehen darin, und alle drei gelten hier: **gesucht wird im Text,
// nicht in den Auszeichnungen**; ohne ausdrückliches Verlangen ist groß gleich
// klein; ein ganzes Wort steht zwischen Rand, Leerzeichen und Satzzeichen.
//
// # adapted: die Reichweite
//
// **MSE2 ersetzt eine Stelle, nicht ein Set.** `CardsPanel::doReplaceAll` ist
// dort `return false; // TODO` (`src/gui/set/cards_panel.cpp`, Zeilen 452–455).
// Gesucht wurde am 2026-08-27 über `src/` nach `ReplaceAll`, `replace_all` und
// `wxFR_REPLACEALL`: die Schaltfläche gibt es im Dialog, den Weg dahinter
// nicht. Was **hier** über ein Set läuft, hat deshalb keine Vorlage — die
// Übereinstimmung ist portiert, die Reichweite ist neu.
//
// **Und daneben steht bei MSE2 der Grund, warum**:
//
//	// TODO: string might have changed when doing replace all
//
// (`src/gui/value/text.cpp`, Zeile 1223). Wer beim Ersetzen weitersucht,
// sucht in einem Text, den er selbst gerade geändert hat. Hier stehen die
// Stellen **vor** der ersten Ersetzung fest, und geschrieben wird von hinten
// nach vorn — dann verschiebt keine Ersetzung die nächste.

// Ersetzungsfrage ist, was gesucht und wodurch es ersetzt wird.
type Ersetzungsfrage struct {
	// Text ist der gesuchte Text — **roh, so wie jemand ihn tippt**. Er wird
	// gegen den entmarkten Wert gehalten, also ohne Auszeichnungen.
	Text string

	// Ersatz tritt an seine Stelle; leer heißt streichen.
	Ersatz string

	// Feld beschränkt auf ein Kartenfeld; leer heißt: über alle.
	Feld string

	// Stand ist der Fingerabdruck, den die Vorschau mitgab —
	// #1241.
	//
	// **Leer heißt: ohne Prüfung.** Das ist der Weg der Kommandozeile, von MCP
	// und vom Dienst — dort gibt es keinen Bestätigungsschritt, den etwas
	// binden könnte. Ihn zu erzwingen bräche drei Flächen für eine Zusage, die
	// dort niemand gibt.
	//
	// **Nicht leer heißt: er muss noch gelten**, sonst `ErrConflict`.
	// Optimistisch und nicht sperrend: eine Sperre hielte den Bestand über die
	// Zeit, die jemand zum Lesen einer Vorschau braucht.
	Stand string

	// GrossKleinBeachten macht die Suche empfindlich für Groß- und
	// Kleinschreibung.
	//
	// **Als Verlangen und nicht als Vorgabe** — MSE liest `wxFR_MATCHCASE`,
	// und die Fahne ist aus, solange niemand sie setzt.
	GrossKleinBeachten bool

	// GanzesWort verlangt eine Wortgrenze an beiden Enden.
	GanzesWort bool
}

// GrenzeErsetzungsvorschau ist, wie viele Stellen die Vorschau nennt.
//
// **Sie gilt der Anzeige und nicht der Änderung.** Ausgeführt wird alles;
// gezeigt wird so viel, wie ein Mensch ansehen kann und eine Antwort tragen
// soll — bei 3.788 Karten und einem häufigen Wort wären es sonst Tausende.
const GrenzeErsetzungsvorschau = 200

// Ersetzungsstelle ist ein Feldwert, der sich ändern würde.
type Ersetzungsstelle struct {
	KarteID int64

	// Nummer ist die Stelle der Karte im Set — damit die Anzeige sie
	// wiederfindet, ohne die Liste zu befragen.
	Nummer int

	Feld string

	// Vorher und Nachher sind der Wert **ohne Auszeichnungen**.
	//
	// **Ohne**, weil ein `<i>` in einer Vorschau wie ein Fehler aussieht — es
	// ist Werkzeug und kein Text. Geschrieben wird trotzdem der rohe Wert
	// samt seinen Auszeichnungen; das ist die Aufgabe von
	// [Ersetzungsfrage.Ersetze].
	Vorher  string
	Nachher string

	// Treffer ist die Zahl der Fundstellen in **diesem** Wert.
	Treffer int
}

// Ersetzungsvorschau ist, was eine Ersetzung ändern würde.
type Ersetzungsvorschau struct {
	Stellen []Ersetzungsstelle

	// Karten und Treffer zählen **alles**, auch was die Grenze abschneidet.
	Karten  int
	Treffer int

	// Weitere sagt, dass die Grenze gegriffen hat.
	Weitere bool

	// OhneVorlage sagt, dass die Feldarten nicht zu lesen waren — dann galt
	// **jedes** Feld als Text.
	//
	// **Es steht in der Antwort und nicht in einem Fehler** (§8): ein Set aus
	// MTG JSON ohne sein Spielpaket soll bearbeitbar bleiben, und die
	// Vorschau nennt jedes Feld beim Namen. Wer den Hinweis sieht, sieht sie
	// genauer an — ein `choice:`-Feld wäre sonst nicht als solches erkennbar.
	OhneVorlage bool

	// Stand ist der **Fingerabdruck der gezeigten Stellen** —
	// #1241.
	//
	// # Wozu
	//
	// Zwischen Vorschau und Bestätigung kann eine Karte durch den Editor, eine
	// Rücknahme, den Dienst oder den **unabhängigen MCP-Prozess** geändert
	// werden. `Ausfuehren` liest den Bestand dann neu und rechnet einen neuen
	// Änderungsplan: der Knopf bestätigte unter Umständen **andere Treffer als
	// die angezeigten**.
	//
	// **Die Absicht war gebunden, der Quellzustand nicht.**
	//
	// # Was er deckt
	//
	// Karte, Feld und den **vorherigen** Wert jeder gezeigten Stelle. Damit
	// merkt er sowohl eine geänderte Karte als auch eine, die dazugekommen
	// ist und einen Treffer trägt.
	//
	// # Was er nicht deckt
	//
	// **Eine Änderung, die keinen Treffer berührt.** Wer an einer ganz anderen
	// Karte ein ganz anderes Feld ändert, hält die Vorschau nicht auf — und
	// soll es nicht: sie zeigt, was sich ändert, und das ist unverändert
	// richtig.
	Stand string
}

// Feldwechsel ist **ein** Feldwert, der sich ändern soll.
//
// **Warum er in der Domäne steht**: `service` muss die Schnittstelle des
// Schreibers formulieren, ohne `storage` zu kennen (§2) — und beide Seiten
// können denselben Typ nur nennen, wenn er in einer Schicht liegt, die beide
// sehen dürfen. Derselbe Grund wie bei [Verlaufseintrag].
type Feldwechsel struct {
	KarteID int64
	Feld    string
	Wert    string

	// Vorher ist der Text, aus dem Wert gerechnet ist — der Speicher schreibt
	// nur, wenn er noch dasteht (#2534, Nachlauf über das Gesamtreview über
	// 0.9.0, Befund 2). Leer heißt: das Feld war leer.
	Vorher string
}

// Trefferstelle ist ein Fund im **rohen** Wert — Bytestellen, keine Zeichen.
//
// **Roh und nicht entmarkt**, damit die Vorschau ihn zeigen und der Schreiber
// ihn austauschen kann, ohne ein zweites Mal zu suchen.
type Trefferstelle struct {
	Anfang int
	Ende   int
}

// Validate prüft die Frage.
//
// **Ein leerer Ersatz ist kein Befund**: „das Wort überall streichen" ist eine
// gewöhnliche Absicht.
func (f Ersetzungsfrage) Validate() []*FieldError {
	if strings.TrimSpace(f.Text) == "" {
		return []*FieldError{{
			Field:   "text",
			Message: "Ohne Suchtext gibt es nichts zu ersetzen.",
		}}
	}

	return nil
}

// Stellen gibt jede Fundstelle im rohen Wert — überschneidungsfrei, von vorn.
func (f Ersetzungsfrage) Stellen(roh string) []Trefferstelle {
	if strings.TrimSpace(f.Text) == "" {
		return nil
	}

	text, grenzen, enden := entmarkt(roh, !f.GrossKleinBeachten)
	gesucht := f.Text

	if !f.GrossKleinBeachten {
		gesucht = strings.ToLower(gesucht)
	}

	var stellen []Trefferstelle

	for stelle := 0; stelle+len(gesucht) <= len(text); {
		if !strings.HasPrefix(text[stelle:], gesucht) || !f.amWortrand(text, stelle, len(gesucht)) {
			stelle++

			continue
		}

		// **Das Ende ist das des getroffenen Zeichens, nicht der Anfang des
		// nächsten** — Befund 9 des Zweitblicks über Phase 3.7 und die Klasse
		// dahinter. Zwischen beiden können Auszeichnungen stehen: bei
		// `a<i>b</i>` und `a → X` verschlang der erste Anlauf das `<i>`.
		stellen = append(stellen, Trefferstelle{
			Anfang: grenzen[stelle], Ende: enden[stelle+len(gesucht)-1],
		})

		// **Weiter hinter dem Treffer** und nicht ein Zeichen weiter: `aa` in
		// `aaa` steht zweimal, und die zweite Ersetzung ginge in die erste.
		stelle += len(gesucht)
	}

	return stellen
}

// Ersetze gibt den Wert mit allen Fundstellen ausgetauscht.
//
// **Von hinten nach vorn**, damit keine Ersetzung die Stellen der nächsten
// verschiebt — das ist die Antwort auf MSEs `// TODO: string might have
// changed when doing replace all`.
func (f Ersetzungsfrage) Ersetze(roh string) string {
	stellen := f.Stellen(roh)
	ersatz := marke(f.Ersatz)
	neu := roh

	for i := len(stellen) - 1; i >= 0; i-- {
		// **Marken im geschnittenen Bereich werden ausgeglichen** — Befund 1
		// des Nachlaufs über Phase 3.7. Geht ein Treffer über eine Marke
		// hinweg, nähme ein stumpfer Schnitt sie mit und ließe ihre
		// Gegenmarke stehen; der Grund und die Fundstelle in MSE2 stehen bei
		// [Markenausgleich].
		davor, danach := Markenausgleich(neu, stellen[i].Anfang, stellen[i].Ende)

		neu = neu[:stellen[i].Anfang] + davor + ersatz + danach + neu[stellen[i].Ende:]
	}

	// **Und danach vereinfacht**, wie MSEs `simplify_tagged`: ein Paar, das
	// nichts mehr umschließt, ist Unrat.
	return OhneLeerePaare(neu)
}

// amWortrand sagt, ob ein Fund an beiden Enden an einer Wortgrenze steht.
//
// **Ohne Verlangen ist jede Stelle eine.** MSE fragt `is_word_end` nur unter
// `if (find.wholeWord())`.
func (f Ersetzungsfrage) amWortrand(text string, stelle, laenge int) bool {
	return !f.GanzesWort || (randVor(text, stelle) && randNach(text, stelle+laenge))
}

// randVor sagt, ob **vor** einer Stelle eine Wortgrenze steht.
//
// # ported, mit einer berichtigten Grenze (§15, Regeln 4 und 7)
//
//	bool is_word_end(const String& s, size_t pos) {
//	  if (pos == 0 || pos >= s.size()) return true;
//	  Char c = s.GetChar(pos);
//	  return isSpace(c) || isPunct(c);
//	}
//	…
//	if (!is_word_end(s, pos - 1) || !is_word_end(s, pos + find.findString().size()))
//	  return false;
//
// `src/gui/value/text.cpp`, Zeilen 1192–1197 und 1202, Fassung f9d9356 vom
// 2021-07-06.
//
// **Die Abweichung**: `is_word_end(s, pos - 1)` fragt „steht vor dem Fund eine
// Grenze?". Bei einem Fund an Stelle 0 läuft `pos - 1` über und fällt in
// `pos >= s.size()` — richtig, der Rand ist eine Grenze. **Bei einem Fund an
// Stelle 1 aber ist `pos - 1` gleich 0**, und dann greift `pos == 0` und gibt
// `true` zurück, **ohne das Zeichen anzusehen**. In MSE2 gilt damit `ab` in
// `xab` als ganzes Wort. Nachgelesen, nicht vermutet: der Zweig steht
// ausdrücklich vor dem `GetChar`.
//
// **Hier ist der Rand die einzige Ausnahme**, und das ist keine Bequemlichkeit:
// MSEs Suche zeigt **einen** Fund, den ein Mensch ansieht und überspringt.
// Diese Ersetzung läuft über ein ganzes Set und schriebe den falschen Fund
// dreihundertmal, ohne dass jemand hinsieht. Ein übernommener Fehler wiegt
// unterschiedlich schwer, je nachdem, wie oft er ausgeführt wird.
//
// **Und gefragt wird nach der Rune, nicht nach dem Byte.** MSEs `GetChar`
// zählt Zeichen; ein `text[stelle]` in Go läse bei einem Umlaut das halbe.
func randVor(text string, stelle int) bool {
	if stelle <= 0 || stelle > len(text) {
		return true
	}

	zeichen, _ := utf8.DecodeLastRuneInString(text[:stelle])

	return istWortgrenze(zeichen)
}

// randNach sagt, ob **an** einer Stelle eine Wortgrenze steht — MSEs
// `is_word_end(s, pos + len)`.
func randNach(text string, stelle int) bool {
	if stelle < 0 || stelle >= len(text) {
		return true
	}

	zeichen, _ := utf8.DecodeRuneInString(text[stelle:])

	return istWortgrenze(zeichen)
}

// istWortgrenze ist MSEs `isSpace(c) || isPunct(c)`.
func istWortgrenze(zeichen rune) bool {
	return unicode.IsSpace(zeichen) || istSatzzeichen(zeichen)
}

// istSatzzeichen ist wx' `isPunct` — Zeichensetzung und Symbole.
//
// **Beides**, denn `wxIsPunct` folgt der C-Bibliothek: alles Druckbare, das
// weder Buchstabe noch Ziffer noch Leerzeichen ist.
func istSatzzeichen(zeichen rune) bool {
	return unicode.IsPunct(zeichen) || unicode.IsSymbol(zeichen)
}
