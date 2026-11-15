package app

import (
	"context"
	"os"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// SetDTO ist ein Set, wie das Frontend es sieht.
//
// # Warum die Kennung eine Zeichenkette ist
//
// `storage.Setkopf` trägt `int64`; JavaScript kennt ganze Zahlen genau nur bis
// 2^53. Eine Kennung, die dort ankommt und nicht mehr dieselbe ist, fällt erst
// auf, wenn jemand ein Set nicht mehr öffnen kann — und dann sucht er überall,
// nur nicht in der Zahlendarstellung. Die Umwandlung ist billig, die Suche
// danach nicht.
//
// # Warum hier nirgends `omitempty` steht
//
// Ein leeres Feld mit `omitempty` verschwindet aus der Antwort; im Frontend ist
// es dann `undefined` statt `""`, und eine **leere Liste** wird zu `null`.
// AGENTS.md §3 nennt genau das als die Lücke zwischen den Prüfebenen: die
// Go-Tests sehen die Struct-Marken nicht, die Frontend-Tests gehen an der
// Serialisierung vorbei, beide sind grün — und die Seite bleibt leer.
type SetDTO struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Spiel    string `json:"spiel"`
	Stil     string `json:"stil"`
	Herkunft string `json:"herkunft"`
	Karten   int    `json:"karten"`

	// HerkunftFehlt sagt, dass die Datei, aus der dieses Set kam, nicht mehr
	// dasteht (#891).
	//
	// **Das Set bleibt trotzdem offen.** Es liegt im Bestand, nicht in der
	// Datei; die Herkunft ist eine Angabe darüber, woher es einmal kam. Wer
	// sie sieht und die Datei sucht, soll aber nicht raten müssen.
	//
	// **Leer heißt nicht »fehlt«**: ein Set, das in der Anwendung entstand,
	// hat gar keine Herkunft.
	HerkunftFehlt bool `json:"herkunftFehlt"`
}

// KarteDTO ist eine Karte mit ihren Feldwerten.
type KarteDTO struct {
	ID     string    `json:"id"`
	Nummer int       `json:"nummer"`
	Werte  []WertDTO `json:"werte"`
}

// WertDTO ist ein Feldwert.
type WertDTO struct {
	Feld string `json:"feld"`
	Wert string `json:"wert"`
}

// VorbehaltDTO ist ein roh mitgeführter Bereich — was beim Import nicht
// gedeutet wurde und trotzdem mitreist.
type VorbehaltDTO struct {
	Datei  string `json:"datei"`
	Zeile  int    `json:"zeile"`
	Inhalt string `json:"inhalt"`
}

// Bestandsleser ist, was `app` vom Speicher braucht.
//
// Die Schnittstelle steht hier und nicht in `storage`: `app` sagt, was es
// braucht, statt zu nehmen, was da ist — dieselbe Richtung wie bei den
// Repositories in `service` (AGENTS.md §2).
// **Über Domänentypen.** Eine Zusage über `[]storage.Setkopf` zwänge `app`,
// `storage` zu importieren — und das darf nur der Composition Root. Der
// Schichtenwächter hat genau das gefunden, als hier zuerst die Speichertypen
// standen; dieselbe Klasse wie `RawBlock` in Phase 2.
type Bestandsleser interface {
	Setkoepfe(ctx context.Context) ([]domain.SetHead, error)
	Karten(ctx context.Context, setID int64) ([]*domain.CardRow, error)
	Rohbloecke(ctx context.Context, setID int64) ([]domain.RawBlock, error)
}

// SetsAntwort ist die Form, in der die Setliste antwortet.
//
// # Warum ein Umschlag und kein zweiter Rückgabewert
//
// Wails nimmt bei zwei Rückgabewerten den ersten als Wert und den zweiten **nur
// dann** als Fehler, wenn er `error` erfüllt (`binding/boundMethod.go`). Ein
// `*Fehler` erfüllt es nicht — er wäre stillschweigend verworfen worden, und
// das Frontend hätte bei jedem Fehler eine leere Liste ohne Meldung bekommen.
//
// Die erzeugte Typdatei behauptet dabei `Promise<T|Fehler>`; zur Laufzeit
// kommt nur `T`. Eine Zusage, die der Typ macht und die Laufzeit nicht hält,
// ist schlimmer als gar keine.
//
// `*Fehler` zu einem `error` zu machen wäre der andere Weg — dann rejectet
// Wails mit `err.Error()`, also mit einer **Zeichenkette**. Code, Schlüssel und
// Feldbezug gingen verloren, und genau die verlangt §8.
type SetsAntwort struct {
	Sets   []SetDTO `json:"sets"`
	Fehler *Fehler  `json:"fehler"`
}

// KartenAntwort trägt Karten oder einen Fehler.
type KartenAntwort struct {
	Karten []KarteDTO `json:"karten"`
	Fehler *Fehler    `json:"fehler"`
}

// VorbehalteAntwort trägt Vorbehalte oder einen Fehler.
type VorbehalteAntwort struct {
	Vorbehalte []VorbehaltDTO `json:"vorbehalte"`
	Fehler     *Fehler        `json:"fehler"`
}

// Sets gibt alle Sets, zuletzt geöffnete zuerst.
//
// **Immer eine Liste, nie `null`.** Ein leeres Ergebnis ist `[]`, und das
// Frontend darf darüber laufen, ohne vorher zu prüfen.
func (a *App) Sets(ctx context.Context) SetsAntwort {
	ctx = a.imProfil(ctx)

	koepfe, err := a.leser().Setkoepfe(ctx)
	if err != nil {
		return SetsAntwort{Sets: []SetDTO{}, Fehler: Uebersetze(ctx, err)}
	}

	sets := make([]SetDTO, 0, len(koepfe))
	for _, kopf := range koepfe {
		sets = append(sets, SetDTO{
			ID:       strconv.FormatInt(kopf.ID, 10),
			Name:     kopf.Name,
			Spiel:    kopf.Game,
			Stil:     kopf.Style,
			Herkunft: kopf.Origin,
			Karten:   kopf.Cards,

			HerkunftFehlt: fehltDieHerkunft(kopf.Origin),
		})
	}

	return SetsAntwort{Sets: sets}
}

// Karten gibt die Karten eines Sets in ihrer Reihenfolge.
func (a *App) Karten(ctx context.Context, setID string) KartenAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := kennungAus(setID)
	if befund != nil {
		return KartenAntwort{Karten: []KarteDTO{}, Fehler: befund}
	}

	gelesen, err := a.leser().Karten(ctx, kennung)
	if err != nil {
		return KartenAntwort{Karten: []KarteDTO{}, Fehler: Uebersetze(ctx, err)}
	}

	karten := make([]KarteDTO, 0, len(gelesen))
	return KartenAntwort{Karten: append(karten, kartenDTO(gelesen)...)}
}

// kartenDTO setzt Karten für die Oberfläche um.
//
// **Ausgeführt seit #1072**: das Zufallspack gibt gezogene Karten zurück, und
// eine zweite Umsetzung daneben liefe beim ersten neuen Feld auseinander.
func kartenDTO(gelesen []*domain.CardRow) []KarteDTO {
	karten := make([]KarteDTO, 0, len(gelesen))

	for _, karte := range gelesen {
		werte := make([]WertDTO, 0, len(karte.Values))
		for _, wert := range karte.Values {
			werte = append(werte, WertDTO{Feld: wert.Field, Wert: wert.Text})
		}

		karten = append(karten, KarteDTO{
			ID:     strconv.FormatInt(karte.ID, 10),
			Nummer: karte.Number,
			Werte:  werte,
		})
	}

	return karten
}

// Vorbehalte gibt, was beim Import eines Sets nicht gedeutet wurde.
//
// Sichtbar zu machen ist der Zweck: ein Set, von dem 300 Karten lesbar sind und
// zwei nicht, wird mit 300 Karten und zwei benannten Befunden geöffnet (§8) —
// und die zwei sollen jemand ansehen können.
func (a *App) Vorbehalte(ctx context.Context, setID string) VorbehalteAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := kennungAus(setID)
	if befund != nil {
		return VorbehalteAntwort{Vorbehalte: []VorbehaltDTO{}, Fehler: befund}
	}

	bloecke, err := a.leser().Rohbloecke(ctx, kennung)
	if err != nil {
		return VorbehalteAntwort{Vorbehalte: []VorbehaltDTO{}, Fehler: Uebersetze(ctx, err)}
	}

	vorbehalte := make([]VorbehaltDTO, 0, len(bloecke))
	for _, block := range bloecke {
		vorbehalte = append(vorbehalte, VorbehaltDTO{
			Datei: block.File, Zeile: block.Line, Inhalt: block.Text,
		})
	}

	return VorbehalteAntwort{Vorbehalte: vorbehalte}
}

// kennungAus liest eine Kennung aus der Zeichenkette, die das Frontend schickt.
//
// Eine unlesbare Kennung ist ein **Validierungsfehler**, kein „nicht gefunden":
// wer `abc` schickt, hat nicht ein Set verfehlt, sondern gar keines genannt.
func kennungAus(text string) (int64, *Fehler) {
	return benannteKennungAus("id", text)
}

// benannteKennungAus ist dasselbe mit dem Feldnamen, den der Aufrufer meint.
//
// **Seit eine Kartenänderung zwei Kennungen trägt**
// (#808):
// „id" sagte, dass **eine** von beiden unlesbar war, und ließ offen, welche.
// §8 verlangt den Fehler am verursachenden Feld — bei zwei gleichartigen
// Angaben ist der gemeinsame Name keiner.
func benannteKennungAus(feld, text string) (int64, *Fehler) {
	kennung, err := strconv.ParseInt(text, 10, 64)
	if err != nil {
		return 0, Nutzerfehler(domain.NewFieldError(feld, domain.SchluesselValidierung,
			"Die Kennung %q ist keine Zahl.", text))
	}

	return kennung, nil
}

// leser gibt den Bestandsleser.
//
// **Gebaut wird er im Composition Root**, nicht hier: `storage` zu importieren
// ist `app.go` vorbehalten (AGENTS.md §2), und der Schichtenwächter zählt nach.
func (a *App) leser() Bestandsleser { return a.bestand }

// MitBestand setzt den Bestandsleser — die Naht für Tests und für den
// Composition Root.
func (a *App) MitBestand(leser Bestandsleser) { a.bestand = leser }

// fehltDieHerkunft sagt, ob eine genannte Quelldatei nicht mehr dasteht.
//
// **Eine leere Herkunft fehlt nicht.** Ein Set, das in der Anwendung entstand,
// hat keine; das ist kein Mangel, sondern eine Aussage (#891).
func fehltDieHerkunft(pfad string) bool {
	if pfad == "" {
		return false
	}

	_, err := os.Stat(pfad)

	return err != nil
}
