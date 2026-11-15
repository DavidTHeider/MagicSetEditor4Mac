package rest

// **Die Nähte der Paket- und Bildwege** — aus `naehte.go` herausgelöst, weil
// jene mit dem Rücksprung (O7, #1431) über die Zeilengrenze aus §2 ging.
//
// **Der Schnitt folgt der Frage** (§2): dort steht, was der Dienst über
// **Karten und Sets** braucht, hier, was er über **Pakete und Bilder**
// braucht — zwei Gebiete, die sich unabhängig voneinander ändern.

import "context"

// Paketkatalog ist, was die zwei Paketendpunkte brauchen.
type Paketkatalog interface {
	Liste(ctx context.Context) (pakete []Paket, befunde []string, fehler error)
	Beziehen(ctx context.Context, adresse, stand, unterverzeichnis, name string) (Bezug, error)

	// PaketInstallieren bindet ein Paket ein, das schon auf der Platte liegt
	// (#1291).
	//
	// **Mit Vorsilbe**, obwohl `Liste` und `Beziehen` daneben keine haben: die
	// Doubles der Tests erfüllen **jede** Naht dieses Pakets, und
	// `Uebernehmen` gab es dort schon — für Scryfall. Ein Name, der zweimal
	// dieselbe Signatur bräuchte, ließe sich nicht erfüllen.
	PaketInstallieren(ctx context.Context, quelle string) (Paket, error)

	// PaketAnlegen schreibt ein neues Paket aus einem Gerüst
	// (#1350).
	PaketAnlegen(ctx context.Context, art, name, vollname, spiel string) (Paket, error)

	// PaketKopieren legt eine eigene Kopie eines Pakets an — FMT-02
	// (#1691).
	PaketKopieren(ctx context.Context, art, name, neu string) (Paket, error)

	// PaketEntfernen nimmt ein Paket von der Platte und aus dem Katalog (#1291).
	//
	// **`pfad` ist die Bestätigung** und wird durchgereicht, nicht hier
	// geprüft: die Wache gehört dorthin, wo die Naht zusammengesetzt wird
	// (0044) — dieselbe Gestalt wie `freierOrt`.
	PaketEntfernen(ctx context.Context, art, name, pfad string) error

	// PaketUebernehmen holt alle Vorlagenpakete aus einem Verzeichnis (#1291).
	PaketUebernehmen(ctx context.Context, ort string) (Uebernahme, error)

	// PaketWeitergeben und PaketEinlesen gehen über `.mse-installer` —
	// FMT-02.AK-04 (#1691).
	PaketWeitergeben(ctx context.Context, pakete []string, ziel string) (string, error)
	PaketEinlesen(ctx context.Context, datei string) (Uebernahme, error)

	// MSE2Orte nennt die üblichen Orte einer MSE2-Installation (#1291).
	MSE2Orte() []string

	// PaketPlan, PaketErsetzen und PaketZurueck — der Plan vor einer
	// Änderung und der Tausch einer Fassung (DOM-01, #1696).
	PaketPlan(ctx context.Context, vorgang, art, name, quelle string) (Paketplan, error)
	PaketErsetzen(ctx context.Context, quelle string, bestaetigt bool) (Paket, error)
	PaketZurueck(ctx context.Context, art, name string) (Paket, error)

	// PaketBedarf nennt, was einem Set an Paketen fehlt — DOM-01.AK-04.
	PaketBedarf(ctx context.Context, setID int64) ([]Paketkonflikt, error)

	// Die Stände der bezogenen Bäume — G04 (#2251): nennen, eine
	// Aktualisierung vorher zeigen, sie ausführen und zurücktauschen.
	PaketBezuege(ctx context.Context, fragen bool) ([]Bezugsstand, error)
	PaketBezugVorschau(ctx context.Context, name, stand string) (Bezugsvorschau, error)
	PaketBezugAktualisieren(ctx context.Context, name, stand string, bestaetigt bool) (Bezug, error)
	PaketBezugZurueck(ctx context.Context, name string) (Bezug, error)
}

// Bezugsstand ist ein bezogener Baum samt Lage seiner Quelle (G04).
type Bezugsstand struct {
	Name         string `json:"name"`
	Adresse      string `json:"adresse"`
	Stand        string `json:"stand"`
	Geprueft     string `json:"geprueft"`
	Pakete       int    `json:"pakete"`
	Vorige       bool   `json:"vorige"`
	VorigerStand string `json:"vorigerStand"`
	Lage         string `json:"lage"`
	Spitze       string `json:"spitze"`
	SpitzeZeit   string `json:"spitzeZeit"`
	Vorsprung    int    `json:"vorsprung"`

	// Grund nennt, warum die Lage `unbekannt` ist — leer sonst.
	Grund string `json:"grund"`
}

// Bezugsvorschau ist, was eine Aktualisierung täte (G04).
type Bezugsvorschau struct {
	Name               string           `json:"name"`
	Von                string           `json:"von"`
	Zu                 string           `json:"zu"`
	Lage               string           `json:"lage"`
	Vorsprung          int              `json:"vorsprung"`
	Rueckstand         int              `json:"rueckstand"`
	Pakete             []Paketaenderung `json:"pakete"`
	Uebrige            int              `json:"uebrige"`
	Gekuerzt           bool             `json:"gekuerzt"`
	Sets               []string         `json:"sets"`
	BestaetigungNoetig bool             `json:"bestaetigungNoetig"`
	Gruende            []string         `json:"gruende"`
	VorigeErsetzt      bool             `json:"vorigeErsetzt"`
}

// Paketaenderung ist ein Paket, das eine Aktualisierung berührt.
type Paketaenderung struct {
	Paket     string `json:"paket"`
	Neu       int    `json:"neu"`
	Geaendert int    `json:"geaendert"`
	Entfernt  int    `json:"entfernt"`
}

// Kartenbilder setzt ein Bild aus einer Datei in ein Kartenfeld (#1291).
type Kartenbilder interface {
	AusDatei(ctx context.Context, setID, karteID int64, feld, pfad string) (Bildmasse, error)
}

// Bildmasse sagt, was gelesen und worauf zugeschnitten wurde.
type Bildmasse struct {
	Feld         string `json:"feld"`
	QuelleBreite int    `json:"quelleBreite"`
	QuelleHoehe  int    `json:"quelleHoehe"`
	ZielBreite   int    `json:"zielBreite"`
	ZielHoehe    int    `json:"zielHoehe"`
}

// Uebernahme ist der Ausgang einer Übernahme aus einer MSE2-Installation.
//
// **Alle drei Listen ohne `omitempty`** (§3): „nichts übergangen" ist eine
// Aussage, `null` ist keine.
type Uebernahme struct {
	Uebernommen []Paket  `json:"uebernommen"`
	Uebergangen []string `json:"uebergangen"`
	Befunde     []string `json:"befunde"`
}

// Paket ist ein installiertes Paket.
type Paket struct {
	Art      string `json:"art"`
	Name     string `json:"name"`
	Version  string `json:"version"`
	Pfad     string `json:"pfad"`
	Herkunft string `json:"herkunft"`

	// Die Übersicht — DOM-01.AK-01 (#1696), **erweiternd** nach §16:
	// Prüfsumme des Bezugs, Herkunftsart (mitgeliefert, bezogen, uebernommen,
	// eigen), ob es vom Stand beim Einbinden abweicht, was es braucht, wer es
	// braucht und wie viele Sets es benutzen.
	Pruefsumme   string   `json:"pruefsumme"`
	Herkunftsart string   `json:"herkunftsart"`
	Geaendert    bool     `json:"geaendert"`
	Braucht      []string `json:"braucht"`
	GebrauchtVon []string `json:"gebrauchtVon"`
	Benutzt      int      `json:"benutzt"`
}

// Bezug ist der Ausgang eines Paketbezugs.
//
// **Die übergangenen Einträge reisen mit** (§8): ein Archiv mit einem Pfad
// über `..` wird nicht ausgepackt, und wer das nicht sähe, hielte ein halbes
// Paket für ein ganzes.
type Bezug struct {
	Ort         string   `json:"ort"`
	Stand       string   `json:"stand"`
	Pruefsumme  string   `json:"pruefsumme"`
	Uebergangen []string `json:"uebergangen"`
}
