package app

// **Die Nähte um die Sammlung** — aus `anwendung.go` herausgelöst, weil jene
// mit der Datenpflege (O8, #1432) über die Zeilengrenze aus §2 ging.
//
// **Der Schnitt folgt der Frage** (§2): dort steht, woraus die Anwendung
// besteht, hier, **was sie über Papier weiß** — Exemplare, Preise, Listen und
// die Wege, die sie ein- und ausfahren.
//
// **Sie erzeugt nichts.** Wie `anwendung.go` erklärt sie nur; der Composition
// Root bleibt `app.go`.

// papiernaehte sind die Nähte um die Sammlung.
//
// **Nicht `sammlungsnaehte`** — den Namen hat `sammlungsnaht.go` für die
// Verdrahtung. Zwei Dinge mit demselben Namen sind eine Verwechslung, die man
// erst bemerkt, wenn eine falsch verdrahtet ist.
type papiernaehte struct {
	// sammlung führt die Exemplare, die jemand besitzt
	// (#1165).
	//
	// **Nichts davon gehört in ein Set.** Ein Exemplar ist ein Stück Papier,
	// eine Karte ein Entwurf ([0052](../../docs/entscheidungen/0052-die-sammlung-ist-nicht-das-set.md)).
	sammlung Sammlungsquelle

	// adressbuch führt die Kontakte aus einem Adressbuch (ID-04, #1703).
	adressbuch Adressbuchquelle

	// sammlungsfenster ist der Ausschnitt einer Sicht — **derselbe Dienst wie
	// an den drei Flächen ohne Fenster** (Befund 1 des Zweitblicks über Phase
	// 5.2). Vorher rechnete das Fenster seinen Ausschnitt selbst, und damit gab
	// es zwei Wahrheiten über dieselbe Frage: die hier, die alles las, und die
	// in `service`, die zweimal schmal fragt.
	sammlungsfenster Sammlungsausschnitt

	// preise holt und führt die Preisstände zu einem Druck (#1165).
	preise Preisquelle

	// listeneinfuhr liest eine Kartenliste in die Sammlung
	// (#1292).
	//
	// **Sie kann fehlen**, wie die Preise: sie braucht ein Zuhause für den
	// Zwischenspeicher von Scryfall. Die Sammlung zu führen braucht das nicht.
	listeneinfuhr Listeneinfuhrquelle

	// deckeinfuhr liest eine Deckliste in ein **Deck** (T8, #1473).
	//
	// **Neben [App.listeneinfuhr] und nicht darin**: jene schreibt Besitz,
	// diese Plätze. Das Issue trennt beides ausdrücklich.
	deckeinfuhr Deckeinfuhrquelle

	// sammlungsausfuhr gibt die Sammlung aus und liest ein CSV zurück (#1292).
	//
	// **Sie braucht kein Netz** — anders als [App.listeneinfuhr]. Wer beide
	// zusammenlegte, machte das Ausgeben von einem Zuhause abhängig, das es
	// dafür nicht braucht (§16, Regel 5).
	sammlungsausfuhr Sammlungsausfuhrquelle

	// sammlungssichten beantwortet Besitz, Lagerorte, Dubletten und Wünsche
	// (#1292, Vermerk 0057).
	sammlungssichten Sammlungssichtenquelle

	// kartenskripte führt die `add cards script:`-Blöcke (#1337).
	kartenskripte Kartenskriptquelle

	// auszaehlung zählt nach den `statistics dimension:`-Blöcken (#1338).
	auszaehlung Auszaehlquelle

	// autoersetzungspflege verwaltet die eigenen `auto replace:`-Regeln
	// (#1341, Vermerk 0060).
	autoersetzungspflege Autoersetzungspflegequelle

	// masse ändert und löscht über eine Auswahl und holt Preise in Schüben
	// (#1292).
	masse Massenquelle

	// fehlkartenwunsch setzt, was einem Deck fehlt, auf die Wunschliste
	// (DOM-03.AK-02, #1698) — **hier und nicht bei den Decks**, weil er
	// Papier schreibt: Wünsche sind Zeilen der Sammlung.
	fehlkartenwunsch Fehlkartenwunschnaht

	// setdeck legt aus einem Set ein Deck an (FMT-04, #1693).
	setdeck Setdecknaht
}
