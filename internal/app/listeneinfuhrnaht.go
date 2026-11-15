// **Die Listeneinfuhr für alle vier Flächen** — Werkstück 2 von Phase 4.5
// (#1292).
//
// §16: dieselbe Fähigkeit, überall dieselbe Antwort.
//
// # Warum eine eigene Naht und nicht ein Anbau an `Sammlungsnaht`
//
// Weil sie eine **andere Quelle** braucht: die Sammlung kommt aus dem Bestand,
// die Deutung aus Scryfall. `SammlungsnahtUeber` scheitert schon heute ohne
// Zuhause; wer die Einfuhr dort anhängte, machte das Führen einer Sammlung von
// einem Dienst abhängig, den es zum Führen nicht braucht (§16, Regel 5).

package app

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// Listentreffer ist ein Druck, der zu einer Zeile passt.
type Listentreffer struct {
	Printing   string `json:"printing"`
	Name       string `json:"name"`
	Setcode    string `json:"setcode"`
	Setname    string `json:"setname"`
	Nummer     string `json:"nummer"`
	Seltenheit string `json:"seltenheit"`
	Bild       string `json:"bild"`
}

// Listendeutungszeile ist eine gelesene Zeile samt ihren Drucken.
type Listendeutungszeile struct {
	// Roh ist die Zeile, wie sie dastand.
	Roh          string `json:"roh"`
	Zeilennummer int    `json:"zeilennummer"`

	Menge   int    `json:"menge"`
	Name    string `json:"name"`
	Setcode string `json:"setcode"`
	Nummer  string `json:"nummer"`
	Finish  string `json:"finish"`

	// Treffer sind die passenden Drucke — höchstens `service.Deutungsgrenze`.
	//
	// **Ohne `omitempty`.** Eine leere Liste muss als `[]` ankommen und nicht
	// als `null`: die Oberfläche zählt sie, und `null.length` ist ein Absturz
	// (§3, „zwischen den Prüfebenen").
	Treffer []Listentreffer `json:"treffer"`

	// Drucke ist die Zahl der passenden Drucke — auch wenn `Treffer` kürzer ist.
	Drucke int `json:"drucke"`

	// Eindeutig sagt, ob diese Zeile ohne Rückfrage übernommen werden kann.
	//
	// **Sie wird mitgeschickt und nicht drüben gerechnet.** Die Regel „genau
	// ein Druck" gehört in eine Schicht; vier Flächen, die sie je selbst
	// rechnen, rechnen sie irgendwann verschieden.
	Eindeutig bool `json:"eindeutig"`

	Befund string `json:"befund"`

	// Weitere sagt, dass es mehr Drucke gibt, als eine Seite nennt.
	Weitere bool `json:"weitere"`

	// Fehler ist der Ausfall dieser Zeile — **übersetzbar**, damit das
	// Fenster nicht den deutschen `Befund` zeigt (G01). Nil heißt: die
	// Quelle hat geantwortet, auch mit „kein Druck".
	Fehler *Fehler `json:"fehler"`
}

// Listendeutung ist, was aus einer eingefügten Liste zu machen wäre.
type Listendeutung struct {
	Zeilen  []Listendeutungszeile `json:"zeilen"`
	Befunde []string              `json:"befunde"`

	// AlterSekunden ist das Alter der ältesten Antwort.
	//
	// **Als Sekunden und nicht als `time.Duration`.** Über die Naht wäre eine
	// Dauer eine Zahl in Nanosekunden, und niemand auf der anderen Seite
	// wüsste, welche Einheit gemeint ist.
	AlterSekunden int64 `json:"alterSekunden"`
}

// Listeneinfuhrzeile ist eine Zeile, für die entschieden ist, welcher Druck es ist.
type Listeneinfuhrzeile struct {
	Printing string `json:"printing"`
	Menge    int    `json:"menge"`
	Zustand  string `json:"zustand"`
	Sprache  string `json:"sprache"`
	Finish   string `json:"finish"`
	Name     string `json:"name"`
	Setcode  string `json:"setcode"`
	Nummer   string `json:"nummer"`
}

// Einfuhrbericht sagt, was eingelesen wurde.
type Einfuhrbericht struct {
	Angelegt []int64 `json:"angelegt"`
	Erhoeht  int     `json:"erhoeht"`

	// Ausgelassen zählt die Zeilen, die eine vorhandene trafen und beim
	// Auslassen stehen blieben (DOM-03.AK-01, #1698).
	Ausgelassen int `json:"ausgelassen"`

	Stueck  int      `json:"stueck"`
	Befunde []string `json:"befunde"`
}

// Listeneinfuhrnaht ist die Einfuhr, wie eine Fläche sie ruft.
type Listeneinfuhrnaht struct{ einfuhr *service.Listeneinfuhr }

// Sammlungsausfuhrnaht gibt die Sammlung als Text — **ohne Netz**.
//
// **Eine eigene Naht neben [ListeneinfuhrUeber]**: das Ausgeben liest den
// Bestand und schreibt Text. Wer es an die Einfuhr hängte, machte es von einem
// Zuhause für Scryfalls Zwischenspeicher abhängig, das es dafür nicht braucht
// (§16, Regel 5).
type Sammlungsausfuhrnaht struct{ ausfuhr *service.Sammlungsausfuhr }

// AusfuhrUeber baut sie über einer Datenbank.
func AusfuhrUeber(db *sql.DB) *Sammlungsausfuhrnaht {
	return &Sammlungsausfuhrnaht{
		// **Mit der Herkunftsnaht** — Werkstück 12 von Phase 4.12 (#1351):
		// eine eingelesene Zeile trägt danach, woher sie kam. **Die eine
		// Stelle**, an der die Ausfuhr gebaut wird; sie hier zu vergessen
		// hieße, dass jede Fläche still ohne Herkunft einliest (§10).
		ausfuhr: service.NeueSammlungsausfuhr(storage.NeuerSammlungsspeicher(db)).
			MitHerkunft(storage.NeueHerkunftsablage(db)).
			// **Und mit dem Druckregister** — Werkstück A2 von Phase 5.3
			// (#1441): ohne es läse die Einfuhr nur den Alias, und eine
			// portable Referenz träfe keinen Druck.
			MitDrucken(storage.NeueDrucke(db)),
	}
}

// Ausgeben schreibt die Sammlung in der genannten Form — `liste` oder `csv`.
func (s *Sammlungsausfuhrnaht) Ausgeben(ctx context.Context, form string) (
	string, error,
) {
	return s.ausfuhr.Ausgeben(ctx, service.Ausfuhrform(form))
}

// CSVEinlesen liest eine CSV-Sammlung in den Bestand.
//
// **Ohne Deutung und ohne Rückfrage**: ein CSV trägt die Druckkennung je
// Zeile, es ist nichts zu entscheiden. Das ist der Rückweg der eigenen
// Ausfuhr, nicht der Weg für eine fremde Liste mit Namen.
//
// **`zusammenfuehrung` sagt, was mit einer vorhandenen Zeile geschieht** —
// `addieren`, `auslassen`, leer heißt addieren (DOM-03.AK-01, #1698).
func (s *Sammlungsausfuhrnaht) CSVEinlesen(ctx context.Context, text, quelle, zusammenfuehrung string) (
	*Einfuhrbericht, error,
) {
	bericht, err := s.ausfuhr.EinfuegenMit(ctx, text, quelle, service.Zusammenfuehrung(zusammenfuehrung))
	if err != nil {
		return nil, err
	}

	return alsEinfuhrbericht(bericht), nil
}

// alsEinfuhrbericht bildet den Bericht des Dienstes ab — **eine** Stelle für
// Liste und CSV.
func alsEinfuhrbericht(bericht service.Einfuhrbericht) *Einfuhrbericht {
	return &Einfuhrbericht{
		Angelegt:    bericht.Angelegt,
		Erhoeht:     bericht.Erhoeht,
		Ausgelassen: bericht.Ausgelassen,
		Stueck:      bericht.Stueck,
		Befunde:     befundtexte(bericht.Befunde),
	}
}

// ListeneinfuhrUeber baut sie über einer Datenbank.
func ListeneinfuhrUeber(db *sql.DB) (*Listeneinfuhrnaht, error) {
	// **Derselbe Klient wie überall, über die eine Stelle**: er bringt die
	// Drosseln mit, und zwei unter derselben Grenze halten sie nicht ein
	// (§16, Regel 4). Seit Phase 4.8 wird dort auch der Dienstschalter
	// geprüft — eine eigene Konstruktion hier wäre ein Loch in der Zusage.
	klient, err := Scryfallklient(NeueEinstellungswahl(storage.NeueEinstellungen(db)).MitRiegel(AutorisierungAus(db)))
	if err != nil {
		return nil, err
	}

	// **Und derselbe Klient beantwortet beide Fragen** (#1348): `Druckesuche`
	// für eine Zeile ohne Set, `Sammlung` für bis zu 75 mit Set. Zwei Klienten
	// hätten zwei Drosseln unter derselben Grenze.
	return &Listeneinfuhrnaht{
		einfuhr: service.NeueListeneinfuhr(klient, storage.NeuerSammlungsspeicher(db)).
			MitSammelsuche(klient),
	}, nil
}

// Deuten liest die Liste und ordnet jede Zeile zu — **ohne zu schreiben**.
func (l *Listeneinfuhrnaht) Deuten(ctx context.Context, text string) (
	*Listendeutung, error,
) {
	gedeutet, err := l.einfuhr.Deuten(ctx, text)
	if err != nil {
		return nil, err
	}

	deutung := &Listendeutung{
		Zeilen:        make([]Listendeutungszeile, 0, len(gedeutet.Zeilen)),
		Befunde:       befundtexte(gedeutet.Befunde),
		AlterSekunden: int64(gedeutet.Alter.Seconds()),
	}

	for _, zeile := range gedeutet.Zeilen {
		deutung.Zeilen = append(deutung.Zeilen, alsDeutungszeile(ctx, zeile))
	}

	return deutung, nil
}

// Einlesen schreibt die entschiedenen Zeilen.
//
// **`zusammenfuehrung` wie beim CSV** (DOM-03.AK-01, #1698).
func (l *Listeneinfuhrnaht) Einlesen(
	ctx context.Context, zeilen []Listeneinfuhrzeile, zusammenfuehrung string,
) (*Einfuhrbericht, error) {
	bericht, err := l.einfuhr.EinlesenMit(ctx, alsListenzeilen(zeilen),
		service.Zusammenfuehrung(zusammenfuehrung))

	// **Der Bericht geht samt Fehler hinaus** (§7, Regel 3).
	//
	// Der Dienst gibt bei einem Systemfehler ausdrücklich den Bericht über die
	// **schon geschriebenen** Zeilen zurück; hier stand `return nil, err` und
	// warf ihn weg. Wer bei Zeile 217 von 300 abbrach, erfuhr nicht, dass 216
	// stehen — und ein zweiter Versuch erhöhte sie noch einmal.
	return alsEinfuhrbericht(bericht), err
}

// alsListentreffer bildet einen gefundenen Druck auf sein DTO ab.
//
// **Eine Stelle und nicht zwei**: die Decklisteneinfuhr (T8, #1473) zeigt
// dieselben Treffer zur Wahl, und zwei Abbildungen desselben Verbunds liefen
// auseinander.
func alsListentreffer(einer service.Suchtreffer) Listentreffer {
	return Listentreffer{
		Printing:   einer.Kennung,
		Name:       einer.Name,
		Setcode:    einer.Setcode,
		Setname:    einer.Setname,
		Nummer:     einer.Nummer,
		Seltenheit: einer.Seltenheit,
		Bild:       einer.Bild,
	}
}

// alsDeutungszeile bildet eine Zeile auf ihr DTO ab.
func alsDeutungszeile(ctx context.Context, zeile service.Zeilendeutung) Listendeutungszeile {
	treffer := make([]Listentreffer, 0, len(zeile.Treffer))

	for _, einer := range zeile.Treffer {
		treffer = append(treffer, alsListentreffer(einer))
	}

	var fehler *Fehler
	if zeile.Fehler != nil {
		fehler = Uebersetze(ctx, zeile.Fehler)
	}

	return Listendeutungszeile{
		Roh:          zeile.Zeile.Roh,
		Zeilennummer: zeile.Zeile.Zeilennummer,
		Menge:        zeile.Zeile.Menge,
		Name:         zeile.Zeile.Name,
		Setcode:      zeile.Zeile.Setcode,
		Nummer:       zeile.Zeile.Nummer,
		// **Leer heißt `nonfoil`** — über die eine Stelle, an der die Domäne
		// das sagt (`Exemplar.Ausfuehrung`), wie jedes DTO der Sammlung (C01).
		Finish:    string(domain.Exemplar{Finish: zeile.Zeile.Finish}.Ausfuehrung()),
		Treffer:   treffer,
		Drucke:    zeile.Drucke,
		Eindeutig: zeile.Eindeutig(),
		Befund:    zeile.Befund,
		Weitere:   zeile.Weitere,
		Fehler:    fehler,
	}
}
