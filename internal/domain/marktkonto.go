package domain

// **Das eigene Konto bei einem Marktplatz** — Werkstück `T11` von Phase 5.6
// (#1476).
//
// # Warum die Gestalt hier steht und nicht im Adapter
//
// **Weil sie die Zuordnung ist, an der §16 die dritte Schicht festmacht.**
// Der Adapter hält die fremden Feldnamen; was davon in diesem Programm gilt,
// steht hier. Wer den Verbund des Anbieters durchreichte, hätte dessen
// Struktur in seiner Fachlogik — und ändert der Anbieter sie, ändert sich
// hier alles.
//
// # Und warum so wenige Felder
//
// **Weil ein Feld, das niemand liest, eine Zusage ohne Einlösung ist.** Der
// Anbieter schickt zu einer Bestellung zwei Dutzend Angaben — Gebühren,
// Zwischensummen, Anschriften. Hier steht, was eine Übersicht braucht und
// was die eine Regel braucht, die daran hängt ([BewegtDenBestand]).

// Marktkonto ist, was ein Marktplatz über den eigenen Zugang sagt.
//
// **Drei Felder von fünf.** `shared_secret` fehlt mit Absicht (siehe oben),
// und `webhook_url` ebenso: dieses Programm nimmt keine Rückrufe entgegen,
// und ein Feld, das nur dasteht, ist eine Zusage ohne Einlösung.
type Marktkonto struct {
	AppID    int64
	Name     string
	NutzerID int64
}

// Eigenangebot ist ein Posten, den das **eigene** Konto zum Verkauf
// anbietet.
//
// **Nicht [Marktangebot]**, und der Unterschied ist die ganze Frage dieses
// Werkstücks: jenes ist ein Posten eines **fremden** Verkäufers und trägt
// Verkäufer und Land; dieses gehört dem eigenen Konto und trägt die eigene
// Notiz. Beides in einen Verbund zu legen hieße, zwei Auskünfte unter einem
// Namen zu führen.
//
// **Es ist kein Verkauf.** Ein Angebot sagt, was im eigenen Lager liegt und
// zu welchem Preis — verkauft ist erst, was in einer Bestellung steht.
type Eigenangebot struct {
	ID          int64
	BlueprintID int64
	Name        string
	Menge       int

	// PreisCent ist der Preis in der kleinsten Einheit, Waehrung seine
	// Währung — **beides zusammen oder keines** (0093).
	PreisCent int64
	Waehrung  string

	// Eigennotiz ist `user_data_field` — der Faden zum eigenen Lager.
	//
	// **Sie kommt mit, weil sie das Einzige ist, was beide Seiten kennen**:
	// der Anbieter kennt keine Exemplarkennung dieses Bestands, und dieser
	// Bestand kennt keine Produktkennung dort.
	Eigennotiz string

	// Beschreibung ist, was der Verkäufer öffentlich dazuschreibt.
	Beschreibung string
}

// Marktbestellung ist ein Kauf, an dem das eigene Konto beteiligt ist.
type Marktbestellung struct {
	ID   int64
	Code string

	// Zustand ist einer der acht ([Bestellzustaende]) — oder ein
	// neunter, den der Anbieter eingeführt hat; dann steht er in den
	// Befunden.
	Zustand Bestellzustand

	// UeberNullkonto ist `via_cardtrader_zero`. **Es entscheidet mit**, ob
	// diese Bestellung den eigenen Bestand bewegt hat
	// ([BewegtDenBestand]).
	UeberNullkonto bool

	// Rolle ist `seller` oder `buyer` — die eigene Rolle in diesem Kauf.
	Rolle string

	// Stueck ist die Zahl der verkauften Posten.
	Stueck int

	// Bezahlt ist der Zeitpunkt der Zahlung, wie der Anbieter ihn schreibt.
	Bezahlt string
}
