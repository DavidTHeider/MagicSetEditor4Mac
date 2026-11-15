package domain

// Verlaufseintrag ist eine Änderung samt ihrem Weg zurück.
//
// # Warum die Entität hier steht und nicht in `storage`
//
// Weil `service` ihre Schnittstelle formulieren muss, ohne `storage` zu kennen
// (AGENTS.md §2). Die strukturelle Typisierung trägt eine Zuordnung nur, wenn
// beide Seiten denselben Typ nennen können — und das können sie nur, wenn er in
// einer Schicht liegt, die beide sehen dürfen.
type Verlaufseintrag struct {
	ID      int64
	KarteID int64

	// Zeitpunkt sagt, wann der Eintrag entstand — RFC 3339 mit Nanosekunden,
	// in UTC, so wie der Bestand ihn schreibt. **Seit U6 (#1417) gelesen**:
	// die Spalte stand seit Phase 2 in der Tafel und erreichte den Eintrag
	// nie; eine Verlaufsliste ohne Uhr wäre ein Protokoll ohne Zeit.
	Zeitpunkt string

	// Handlung sagt, was geschehen ist — [HandlungFeld], [HandlungKarte]
	// oder [HandlungReihenfolge].
	//
	// **Die Spalte steht seit Phase 2 da**, und bis zum 2026-08-20 trug sie
	// nur `feld`. Seit dem Löschen einer Karte
	// (#621)
	// trägt sie, worauf sie gewartet hat.
	Handlung string

	// VorherKarte und NachherKarte sind gesetzt, wenn [Verlaufseintrag.Handlung]
	// [HandlungKarte] ist — sonst nil.
	//
	// **Zwei Zeiger statt eines Wahrheitswerts**: ein Schritt kann eine Karte
	// entfernen (vorher da, nachher nicht) oder anlegen (umgekehrt), und beide
	// Seiten tragen ihren eigenen Stand.
	VorherKarte  *Kartenstand
	NachherKarte *Kartenstand

	// VorherStellen und NachherStellen sind gesetzt, wenn
	// [Verlaufseintrag.Handlung] [HandlungReihenfolge] ist — sonst nil.
	VorherStellen  *Stellenstand
	NachherStellen *Stellenstand

	// VorherWort und NachherWort sind gesetzt, wenn
	// [Verlaufseintrag.Handlung] [HandlungSchluesselwort] ist — sonst nil
	// (#1077).
	VorherWort  *Schluesselwortstand
	NachherWort *Schluesselwortstand

	Feld string

	// **`GabVorher` unterscheidet „war leer" von „gab es nicht".** Wer beides
	// zusammenwirft, kann eine Rücknahme nicht mehr ausführen: eine leere Zeile
	// ist nicht dasselbe wie keine Zeile
	// ([0010](../../docs/entscheidungen/0010-mcp-als-eigener-prozess.md)).
	Vorher    string
	GabVorher bool

	Nachher    string
	GabNachher bool

	// StelleVorher und StelleNachher sagen, an welcher Stelle das Feld in der
	// Karte stand — vor und nach der Änderung.
	//
	// **Sie gehören zum Änderungsprotokoll, nicht zur Datenbank.** Der Bestand
	// liest die Feldwerte einer Karte in dieser Reihenfolge, und die
	// Quelldatei wird danach geschrieben (#48). Eine Rücknahme, die das Feld
	// ans Ende setzt, macht aus einer zurückgenommenen Änderung eine bleibende
	// (#160).
	//
	// **Zeiger, weil „keine Stelle" nicht die 0 ist.** Die 0 ist die erste
	// Stelle; ein Feld, das es nicht gab, hat gar keine. Ein Eintrag aus einer
	// Sitzung vor #160 führt beide als `nil` — und wird behandelt wie ein Feld
	// ohne Stelle, also hinten angehängt.
	StelleVorher  *int64
	StelleNachher *int64

	// HerkunftVorher und HerkunftNachher sagen, aus welcher Adresse der Wert
	// stammte — leer: von hier
	// (#1036).
	//
	// **Damit eine Rücknahme sie mitführen kann.** Ohne sie stellte sie den
	// Wert wieder her und seine Nennung nicht — und über einem
	// wiederhergestellten eigenen Bild bliebe eine fremde stehen.
	HerkunftVorher  string
	HerkunftNachher string

	Zurueckgenommen bool

	// Urheber sagt, wer geschrieben hat — Fenster oder Server (#52).
	Urheber Urheber

	// Herkunft sagt, **woher der Wert kam** — leer heißt: von hier.
	//
	// **Eine andere Frage als [Verlaufseintrag.Urheber]**, und deshalb eine
	// eigene Angabe. Der Urheber nennt die **Fläche**: ein von Hand getippter
	// Regeltext und einer, den jemand von Scryfall übernommen hat, kommen
	// beide über dasselbe Fenster. Wer sie zusammenwürfe, könnte eine
	// Übernahme nach dem Schreiben nicht mehr von einer Eingabe unterscheiden.
	//
	// **§16, Regel 6 verlangt sie**: „Nichts wird still übernommen. Was von
	// außen in ein Kartenfeld will, zeigt vorher, was sich ändert, und trägt
	// **danach seine Herkunft**." Der Vorschlag trug sie schon; die
	// geschriebene Zeile bis zum 2026-08-23 nicht.
	//
	// **Was drinsteht, ist eine Adresse**, keine Bezeichnung: die Kartenseite
	// bei Scryfall, die MTG-JSON-Datei, das bezogene Paket. Ein Name allein
	// („Scryfall") ließe später niemanden nachsehen.
	Herkunft string

	// Gruppe hält eine Massenänderung als **einen** Schritt zusammen; leer heißt
	// „steht für sich".
	//
	// **Ein Schritt und nicht eine Zeile ist die Einheit der Rücknahme.** Wer
	// dreihundert Karten in einem Zug ändert, nimmt sie in einem Zug zurück —
	// und die Grenze des Verlaufs zählt ebenfalls Schritte, sonst zerschnitte
	// sie eine Massenänderung und stellte einen Stand her, den es nie gab.
	Gruppe string
}

// Verlaufsschritt ist eine Bewegung des Verlaufs — eine Einzeländerung oder
// eine Massenänderung als Ganzes.
//
// **Der Schritt und nicht die Zeile ist die Einheit der Rücknahme.** Hier, weil
// `service` seine Schnittstelle formulieren muss, ohne `storage` zu kennen (§2).
type Verlaufsschritt struct {
	// Der **erste** Eintrag des Schritts — bei einer Einzeländerung der
	// einzige, bei einer Massenänderung der der ersten Karte, bei einer
	// Bearbeitung mit Nachzug das bearbeitete Feld (#2396).
	Verlaufseintrag

	// Umfang ist die Zahl der Karten, die der Schritt berührt hat. Die
	// Oberfläche soll sagen können, wie viele es waren.
	Umfang int
}

// Urheber sagt, wer eine Änderung geschrieben hat.
//
// **Weil mehrere Prozesse schreiben** ([0010](../../docs/entscheidungen/0010-mcp-als-eigener-prozess.md)):
// das Fenster, der MCP-Server und seit #886 die Kommandozeile, auf derselben
// Datei. Eine Rücknahme, die nicht
// sagt, wessen Änderung sie umkehrt, ist im Mehr-Schreiber-Betrieb eine
// Vermutung (#52).
type Urheber string

// Die vier, die es gibt.
//
// **Der dritte kam mit der Kommandozeile**
// (#886) —
// genau der Fall, den der Satz hier vorgesehen hat: „mehr sollen es nicht
// werden, ohne dass ein dritter Prozess dazukommt". Er ist dazugekommen, und
// eine Rücknahme, die nicht sagt, wessen Änderung sie umkehrt, wäre mit drei
// Schreibern noch mehr Vermutung als mit zweien.
//
// **Der vierte kam mit dem Dienst auf der Rückschleife**
// (#997,
// [0046](../../docs/entscheidungen/0046-doch-ein-dienst-auf-der-rueckschleife.md)),
// und er ist der erste, hinter dem **nicht unbedingt ein Mensch am selben
// Rechner** steht: was über HTTP kommt, hat eine Marke vorgezeigt und sonst
// nichts. Genau deshalb soll ein Verlaufseintrag ihn benennen können.
const (
	UrheberFenster       Urheber = "fenster"
	UrheberServer        Urheber = "server"
	UrheberKommandozeile Urheber = "kommandozeile"
	UrheberDienst        Urheber = "dienst"
)

// Die Handlungen, die ein Verlaufseintrag tragen kann.
//
// **Die Spalte steht seit Phase 2 da und trug bis zum 2026-08-20 nur `feld`.**
// Der Kommentar an [Verlaufseintrag.Handlung] hat gesagt, worauf sie wartet.
const (
	// HandlungFeld ist eine Änderung an einem Feldwert.
	HandlungFeld = "feld"

	// HandlungKarte ist das Anlegen oder Entfernen einer ganzen Karte.
	HandlungKarte = "karte"

	// HandlungReihenfolge ist ein Tausch zweier Karten
	// (#628).
	HandlungReihenfolge = "reihenfolge"

	// HandlungSchluesselwort ist eine Änderung an einem Schlüsselwort des Sets
	// (#1077).
	//
	// **Eine Art für alle drei Handlungen** — anlegen, ändern, löschen —, und
	// nicht drei. Der Unterschied steht in den beiden Ständen: anlegen heißt
	// „vorher nicht da", löschen „nachher nicht da", ändern „beides da".
	// Dieselbe Gestalt wie bei [HandlungKarte], und aus demselben Grund: eine
	// Rücknahme, die drei Arten unterscheiden müsste, unterschiede sie an drei
	// Stellen.
	HandlungSchluesselwort = "schluesselwort"
)

// Schluesselwortstand ist ein Schlüsselwort, wie es auf einer Seite eines
// Verlaufsschritts dasteht.
//
// **Mit Kennung**, anders als [Kartenstand]: die des Schlüsselworts steht nicht
// in der Verlaufszeile. `history.karte_id` ist die einer **Karte**, und eine
// Spalte mit zwei Bedeutungen wäre später nicht mehr auseinanderzuhalten — die
// Setangaben führen sie aus demselben Grund als 0.
type Schluesselwortstand struct {
	ID int64 `json:"id"`

	// GabEs sagt, ob das Schlüsselwort auf dieser Seite des Schritts dasteht.
	//
	// **Nicht aus leeren Feldern zu erschließen**: ein frisch angelegtes ist in
	// jedem Feld außer dem Modus leer — so legt MSE eines an
	// (`AddKeywordAction`, `src/data/action/keyword.cpp`, Zeilen 20–32,
	// Fassung f9d9356 vom 2021-07-06).
	GabEs bool `json:"gabEs"`

	// Reihenfolge ist die Stelle in der Liste des Sets.
	//
	// **Sie muss zurück**, weil sie entscheidet, welches Schlüsselwort zuerst
	// auf einen Text passt (`src/data/keyword.cpp`, Zeilen 148–208, Fassung
	// f9d9356 vom 2021-07-06) — und weil `set_keywords` `UNIQUE (set_id,
	// reihenfolge)` trägt. Ein zurückgenommenes Löschen, das ans Ende
	// zurückbrächte, wäre keine Rücknahme.
	Reihenfolge int64 `json:"reihenfolge"`

	Wort       string `json:"wort"`
	Muster     string `json:"muster"`
	Erinnerung string `json:"erinnerung"`
	Regeln     string `json:"regeln"`
	Modus      string `json:"modus"`
}

// Stellenstand sagt, wohin zwei getauschte Karten gehören.
//
// # Warum ein Eintrag und nicht zwei
//
// **`cards` trägt `UNIQUE (set_id, reihenfolge)`.** Zwei Einträge nacheinander
// auszuführen hieße, die Zusage zwischendurch zu verletzen: die erste Karte
// zöge auf eine Stelle, auf der die zweite noch steht. Ein Tausch ist eine
// Bewegung, kein Paar von Bewegungen — und er gehört in einen Eintrag.
//
// # Warum beide Wege dieselbe Gestalt haben
//
// Hin- und Rückweg unterscheiden sich nur darin, **wohin** die beiden Karten
// gehören. Damit braucht das Ausführen keine Richtung zu kennen: es setzt
// jede Karte auf die Stelle, die dasteht.
type Stellenstand struct {
	// Eine und Andere sind die beiden Karten.
	Eine   int64 `json:"eine"`
	Andere int64 `json:"andere"`

	// EineStelle und AndereStelle sind die Stellen, auf die sie gehören.
	EineStelle   int64 `json:"eineStelle"`
	AndereStelle int64 `json:"andereStelle"`
}

// Kartenstand ist eine ganze Karte, wie sie vor oder nach einem Schritt
// dastand.
//
// # Warum die Felder mitreisen
//
// **`card_values` hängt an `ON DELETE CASCADE`.** Nach dem Löschen sind sie
// weg; wer sie nicht im Rückweg führt, bringt eine leere Karte zurück und
// nennt das eine Rücknahme.
type Kartenstand struct {
	// GabEs sagt, ob die Karte auf dieser Seite des Schritts dasteht.
	//
	// **Nicht aus `len(Felder) > 0` zu erschließen**: eine Karte ohne
	// Feldwerte ist eine gültige Karte — `KarteAnlegen` legt genau die an.
	GabEs bool `json:"gabEs"`

	// Reihenfolge ist die Stelle im Set.
	//
	// **Sie muss zurück, weil `cards` `UNIQUE (set_id, reihenfolge)` trägt**
	// und weil die Karte sonst ans Ende rutschte — eine Rücknahme, die die
	// Ordnung ändert, ist keine.
	Reihenfolge int64 `json:"reihenfolge"`

	Felder []Feldstand `json:"felder,omitempty"`
}

// Feldstand ist ein Feldwert samt seiner Stelle in der Karte.
type Feldstand struct {
	Feld        string `json:"feld"`
	Wert        string `json:"wert"`
	Reihenfolge int64  `json:"reihenfolge"`

	// Herkunft ist die Adresse, aus der dieser Wert stammt — leer: von hier
	// (#1036).
	//
	// **Damit eine zurückgeholte Karte ihre Nennungen wiederbekommt.**
	// `wert_herkunft` hängt über `ON DELETE CASCADE` an `cards`; eine
	// gelöschte Karte nimmt ihre Zeilen mit, und ohne diese Angabe im
	// Verlaufseintrag wären sie nicht wiederherzustellen.
	//
	// **Wahlfrei, wie `Reihenfolge` darüber**: ein Eintrag aus einer älteren
	// Sitzung trägt den Schlüssel nicht und kommt als leer an — „von hier".
	Herkunft string `json:"herkunft,omitempty"`
}
