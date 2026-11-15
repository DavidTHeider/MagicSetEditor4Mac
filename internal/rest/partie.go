// **Die Partie über den Dienst** — Werkstück G1 von Phase 5.5
// (#1458).
//
// # Fünf Wege, und jeder trägt eine Sitzkennung
//
// **Weil es die Frage „wer sieht das?" ohne sie nicht gibt.** Über HTTP ist
// sie ein Abfrageparameter statt eines angemeldeten Fensters — und das ist
// der ganze Unterschied. Wer sie wegließe, bekäme die Sicht eines Zuschauers
// und hielte sie für die eines Spielers.
//
// # Und die Befehlskennung kommt vom Rufer
//
// **Sie steht im Rumpf und wird hier nicht erzeugt.** Nur der Rufer weiß, ob
// er noch einmal dasselbe meint oder etwas Neues; eine Kennung aus diesem
// Dienst machte aus jedem Wiederholungsversuch einen zweiten Zug (§7,
// Regel 4).

package rest

import "context"

// Partienwerkstatt ist, was die Endpunkte dieser Datei brauchen.
type Partienwerkstatt interface {
	// PartieAnlegen nimmt seit P1a (#2335) die Abwahl mit — die Bereiche,
	// die die Runde selbst übernimmt. **Leer heißt: keine.**
	PartieAnlegen(ctx context.Context, spiel, wirt string,
		sitze []Sitzwunsch, abgewaehlt []string) (Partieangaben, error)
	Partie(ctx context.Context, kennung, fuer string) (Partieangaben, error)

	// Partien nennt die Partien dieses Profils, zuletzt geänderte zuerst
	// (#1600).
	Partien(ctx context.Context) ([]Partiezeilenangaben, error)
	PartieBefehl(ctx context.Context, befehl Spielbefehl) (Befehlsangaben, error)
	PartieStand(ctx context.Context, kennung, fuer, stand string) (
		Partieangaben, error)

	// Die vier Wege der Verwaltung — PLAY-03.AK-05
	// (#1711).
	PartieLoeschen(ctx context.Context, kennung string) (
		Partieverwaltungsangaben, error)
	PartieKopieren(ctx context.Context, kennung string) (
		Partieverwaltungsangaben, error)
	PartieAusfuhr(ctx context.Context, kennung string) (
		Partieverwaltungsangaben, error)
	PartieEinfuhr(ctx context.Context, inhalt string) (
		Partieverwaltungsangaben, error)
}

// Partiezeilenangaben ist eine Partie, wie eine Liste sie zeigt.
//
// **Ohne Objekte** — eine Liste ist keine Sicht.
type Partiezeilenangaben struct {
	ID      string `json:"id"`
	Fassung string `json:"fassung"`

	Spiel string `json:"spiel"`
	Stand string `json:"stand"`

	Sitze []Zeilensitzangaben `json:"sitze"`

	Objekte int `json:"objekte"`
	Zug     int `json:"zug"`

	// EigenerSitz ist der Sitz, der an das Profil dieses Zugangs gebunden
	// ist — leer heißt: keiner (UXR-F16,
	// #2031).
	// **Erweiternd** im Sinne von
	// [0033](../../docs/entscheidungen/0033-was-ueber-fassungen-hinweg-zugesagt-wird.md).
	EigenerSitz string `json:"eigenerSitz"`

	// Dran sagt, ob die Partie gerade auf den eigenen Sitz wartet.
	Dran bool `json:"dran"`
}

// Zeilensitzangaben ist ein Platz, wie eine Liste ihn zeigt.
type Zeilensitzangaben struct {
	Kennung string `json:"kennung"`
	Name    string `json:"name"`
}

// Sitzwunsch ist ein Platz am Tisch samt Deck.
type Sitzwunsch struct {
	Kennung string `json:"kennung"`
	Name    string `json:"name"`
	DeckID  string `json:"deckID"`

	// Commander ist die Kartenkennung des Commanders — CR 903.3 (M19,
	// #1620).
	// **Leer heißt: keiner.**
	//
	// **Erweiternd** im Sinne von
	// [0033](../../docs/entscheidungen/0033-was-ueber-fassungen-hinweg-zugesagt-wird.md).
	Commander string `json:"commander,omitempty"`

	// ProfilID bindet den Sitz an ein Profil **desselben Nutzers** (G8,
	// #1465; an dieser Fläche seit UXR-F21, #2041) — **leer heißt: keines**.
	//
	// Der Sitz am Profil dieses Zugangs ist in `eigenerSitz` und `dran` der
	// Liste der eigene. Ob die Kennung eine ist und ob es sie gibt, entscheidet
	// `internal/app` für alle vier Flächen (§2); ein Profil, das es nicht gibt,
	// ist ein **Befund** und kein Fehler, wie am Fenster (§8).
	//
	// **Erweiternd** im Sinne von
	// [0033](../../docs/entscheidungen/0033-was-ueber-fassungen-hinweg-zugesagt-wird.md).
	ProfilID string `json:"profilID,omitempty"`
}

// Spielbefehl ist ein Wunsch an eine Partie.
type Spielbefehl struct {
	PartieID string         `json:"partieID"`
	Absender string         `json:"absender"`
	Kennung  string         `json:"kennung"`
	Art      string         `json:"art"`
	Nutzlast map[string]any `json:"nutzlast,omitempty"`
}

// Partieangaben sind eine Partie, wie ein Sitz sie sehen darf.
type Partieangaben struct {
	ID      string `json:"id"`
	Fassung string `json:"fassung"`

	Spiel string `json:"spiel"`
	Stand string `json:"stand"`

	Fuer string `json:"fuer"`
	Wirt string `json:"wirt"`

	Sitze   []Partiesitz   `json:"sitze"`
	Objekte []Partieobjekt `json:"objekte"`

	// Mischpflichten sind die verborgenen Zonen, die jemand angesehen und
	// seitdem nicht gemischt hat — **für jeden dieselben**.
	Mischpflichten []Mischpflicht `json:"mischpflichten"`

	// Legendenwahlen sind die offenen Fragen aus CR 704.5j — **für jeden
	// dieselben** (M12,
	// #1614).
	//
	// **Eine Pflicht, die niemand sieht, erfüllt niemand.** Bis jemand
	// geantwortet hat, steht das Spiel; wer den Grund nicht sähe, hielte es
	// für einen Ausfall. Derselbe Satz wie beim Zugstand darunter, und
	// derselbe Grund.
	//
	// **Erweiternd** im Sinne von
	// [0033](../../docs/entscheidungen/0033-was-ueber-fassungen-hinweg-zugesagt-wird.md):
	// ein neues Feld, kein bestehendes geändert.
	Legendenwahlen []Legendenwahl `json:"legendenwahlen"`

	// Zug ist, wo die Partie im Zug steht — der Grund, warum er erst seit
	// `M6` hier steht, bei [Partiezug].
	Zug Partiezug `json:"zug"`

	// Abgewaehlt sind die Regelbereiche, die die Runde selbst übernimmt —
	// eingefroren mit der Partie (P1a,
	// #2335).
	// **Für jeden Sitz dieselben**: sie sind Teil der Vereinbarung und kein
	// Geheimnis. **Eine leere Liste und kein `null`.** **Erweiternd** im
	// Sinne von
	// [0033](../../docs/entscheidungen/0033-was-ueber-fassungen-hinweg-zugesagt-wird.md).
	Abgewaehlt []string `json:"abgewaehlt"`

	// Vorrat sind die Kartenkennungen des **eigenen** Decks — damit die
	// Bilder da sind, bevor die Karte kommt (G2,
	// #1459).
	//
	// **Leer für jeden, der nicht an diesem Sitz sitzt.** Die Reihenfolge
	// der Bibliothek bleibt geheim; welche Karten im eigenen Deck liegen,
	// hat der Eigner selbst aufgeschrieben.
	Vorrat []string `json:"vorrat"`

	Erlaubt []string `json:"erlaubt"`

	// Handlungen: je Objekt, was dieser Sitz daran tun kann (P3, #2371).
	Handlungen []Objekthandlung `json:"handlungen"`

	// Befunde sind Auffälligkeiten, die die Partie **nicht** unbenutzbar
	// machen (G8, #1465; an dieser Fläche seit UXR-F21, #2041): ein Sitz,
	// dessen Profil es nicht gibt, bleibt ein Sitz, und die Bindung wird
	// **benannt** statt verschwiegen — wie am Fenster. **Eine leere Liste und
	// kein `null`.** **Erweiternd** im Sinne von
	// [0033](../../docs/entscheidungen/0033-was-ueber-fassungen-hinweg-zugesagt-wird.md).
	Befunde []string `json:"befunde"`

	// Nachladefehler: angelegt oder gewechselt **ist**, nur die Sicht
	// danach war nicht zu lesen (#1970) — dann stehen nur Kennung und Sicht
	// da. Erweiternd (0033).
	Nachladefehler *Befund `json:"nachladefehler,omitempty"`
}

// Mischpflicht nennt Sitz und Zone einer offenen Mischpflicht.
type Mischpflicht struct {
	Sitz string `json:"sitz"`
	Zone string `json:"zone"`
}

// Legendenwahl nennt Sitz und Namen einer offenen Wahl — CR 704.5j.
type Legendenwahl struct {
	Sitz string `json:"sitz"`
	Name string `json:"name"`
}

// Partiezug ist, wo die Partie im Zug steht — CR 500.1 (Fassung 2026-08-19).
//
// # Warum er erst seit `M6` hier steht
//
// **Weil er an keiner Fläche außer dem Fenster und der Kommandozeile
// stand.** G9 (#1592)
// hat den Zugbegriff gebaut, und `internal/app` gab ihn heraus — MCP und
// dieser Dienst reichten ihn nicht weiter. Gemessen am 2026-09-17 mit dem
// Lauf, der beide Flächen gegeneinander hält: die Kommandozeile sagte „Zug 1,
// Versorgungsschritt", diese Fläche sagte „Zug 0".
//
// **Das ist §16**: dieselbe Fähigkeit, überall dieselbe Antwort — und ein
// Wächter, der nur Bindungen zählt, sieht so etwas nicht.
//
// **Erweiternd** im Sinne von
// [0033](../../docs/entscheidungen/0033-was-ueber-fassungen-hinweg-zugesagt-wird.md):
// ein neues Feld, kein bestehendes geändert.
type Partiezug struct {
	// Nummer ist der wievielte Zug — **null heißt: es geht nicht zugweise
	// zu**, und dann steht am Tisch kein Takt.
	Nummer int `json:"nummer"`

	Aktiv   string `json:"aktiv"`
	Starter string `json:"starter"`
	Phase   string `json:"phase"`
	Schritt string `json:"schritt"`

	// Prioritaet ist, wer gerade handeln darf — **leer heißt: niemand**
	// (CR 502.4, CR 514.3).
	Prioritaet string `json:"prioritaet"`

	// Abwurf ist, wie viele Karten der aktive Spieler noch abwerfen muss —
	// CR 514.1.
	Abwurf int `json:"abwurf"`

	// Angriffe und Bloecke sind der Kampfstand — CR 508.1, CR 509.1
	// (#1623).
	//
	// **Wer blocken will, muss wissen, wer angreift** (CR 509.1a).
	//
	// **Erweiternd** im Sinne von
	// [0033](../../docs/entscheidungen/0033-was-ueber-fassungen-hinweg-zugesagt-wird.md):
	// zwei neue Felder, kein bestehendes geändert.
	Angriffe []Kampfpaar `json:"angriffe"`
	Bloecke  []Kampfpaar `json:"bloecke"`
}

// Kampfpaar ist eine Kreatur und das, wogegen sie steht.
//
// **Ein Verbund für beide Richtungen**: bei einem Angriff steht das Ziel
// (CR 508.1a), bei einem Block der Angreifer (CR 509.1a).
type Kampfpaar struct {
	Kreatur   string `json:"kreatur"`
	Ziel      string `json:"ziel,omitempty"`
	Angreifer string `json:"angreifer,omitempty"`
}

// Partiesitz ist ein Platz am Tisch — ohne das, was an ihm geheim ist.
type Partiesitz struct {
	Kennung string `json:"kennung"`
	Name    string `json:"name"`
	Karten  int    `json:"karten"`

	// Leben, Vergiftung und Manavorrat machen den Sitz zum **Spieler** —
	// CR 119.1, CR 122.1f, CR 106.4 (Fassung 2026-08-19).
	//
	// # Warum sie erst mit M3 hier stehen
	//
	// **Weil sie an keiner Fläche standen.** G9 hat Leben und
	// Vergiftungsmarken ans Modell gelegt
	// (#1592),
	// `M17` setzt sie durch — und gemessen am 2026-09-17 zeigte **weder das
	// Fenster noch dieser Dienst noch MCP** eine der beiden Zahlen. Eine
	// Regel, die greift, ohne dass jemand sieht, worauf, ist die erste
	// Dimension aus §10.
	//
	// **Erweiternd** im Sinne von
	// [0033](../../docs/entscheidungen/0033-was-ueber-fassungen-hinweg-zugesagt-wird.md):
	// drei neue Felder, kein bestehendes geändert.
	Leben      int `json:"leben"`
	Vergiftung int `json:"vergiftung"`

	// Manavorrat steht als Buchstabenfolge — `RRG`, leer heißt leer.
	Manavorrat string `json:"manavorrat"`
}

// Partieobjekt ist ein Ding auf dem Tisch, so weit es gezeigt werden darf.
type Partieobjekt struct {
	Kennung     string `json:"kennung"`
	Beherrscher string `json:"beherrscher"`
	Zone        string `json:"zone"`

	Karte string `json:"karte,omitempty"`
	Name  string `json:"name,omitempty"`

	// KarteID ist die Zeile im Bestand — leer, wo nicht hineingesehen werden
	// darf (G2, #1459).
	KarteID string `json:"karteID,omitempty"`

	// Farben, Typ und Werte sind, was diese Karte **ist** — leer, wo nicht
	// hineingesehen werden darf (M2, #1604).
	//
	// **Als Text und nicht als Verbund**: `WU`, `legendary creature — Human
	// Wizard`, `2/3`. Diese Fläche ist Text, und wer sie liest, soll lesen
	// können. Hergeleitet wird an **einer** Stelle (`app.EigenschaftenDTO`).
	Farben   string `json:"farben,omitempty"`
	Typzeile string `json:"typzeile,omitempty"`
	Werte    string `json:"werte,omitempty"`

	// Manawert ist die Summe des Manas im Kosten — CR 202.3.
	Manawert int `json:"manawert,omitempty"`

	Getappt bool `json:"getappt"`

	// Geheim heißt: **dieser** Sitz darf nicht hineinsehen.
	Geheim bool `json:"geheim"`
}

// Befehlsangaben sind, was aus einem Spielbefehl wurde.
type Befehlsangaben struct {
	Fassung    string           `json:"fassung"`
	Ereignisse []Partieereignis `json:"ereignisse"`
}

// Partieereignis ist, was geschehen ist.
//
// # Warum es seinen Inhalt trägt und nicht nur seine Art
//
// **Weil sonst eine Fähigkeit hier ins Leere liefe.** `zone.blick` gibt sein
// Ergebnis als **private Ergänzung** zurück — wer nur `art` und `kette`
// bekäme, hätte gesucht und nichts gefunden. Das ist §16: dieselbe Fähigkeit,
// überall dieselbe Antwort. Der erste Anlauf ließ beide Felder weg, und an
// der Kommandozeile stand die Bibliothek, hier nicht.
type Partieereignis struct {
	Folge string `json:"folge"`
	Art   string `json:"art"`

	// Oeffentlich ist, was jeder sehen darf.
	Oeffentlich map[string]any `json:"oeffentlich"`

	// Privat ist die Ergänzung für **den Absender** — leer, wo keine für ihn
	// dabei war.
	Privat map[string]any `json:"privat"`

	Kette string `json:"kette"`
}
