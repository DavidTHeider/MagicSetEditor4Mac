package app

// **Teilen über die Ziele des Systems** — NET-04
// (#1708),
// entschieden in
// [0155](../../docs/entscheidungen/0155-teilen-ist-eine-kopie.md).
//
// # Drei Schritte, und der mittlere gehört dem Fenster
//
//  1. **Vorbereiten** legt eine Kopie ab und sagt, was darin steht und was
//     fehlt (AK-02) — [App.TeilenVorbereiten].
//  2. **Anbieten** öffnet den Teilen-Dialog von macOS: Mail, Nachrichten,
//     AirDrop und was sonst eingerichtet ist. Wer das Ziel wählt und den
//     Versand bestätigt, tut es im zuständigen Dienst — [App.TeilenAnbieten].
//  3. **Verwerfen** räumt die Kopie weg (AK-04) — [App.TeilenVerwerfen].
//
// **Nur das Fenster hat einen Teilen-Dialog.** Die Flächen ohne Fenster
// geben dieselben Dateien über ihre Ausfuhrwege aus (`mse deck liste`,
// `mse sammlung-ausgeben`, …); die Zeile in `docs/FAEHIGKEITEN.md` sagt es.

import (
	"context"
	"sync"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Teilnaht zeigt den Teilen-Dialog des Systems für eine Datei oder einen
// Text — `text` geht vor, wenn beides da ist.
type Teilnaht func(ctx context.Context, pfad, text string) error

// Teilablage legt die Kopien ab — [service.Teilstelle].
type Teilablage interface {
	Ordner() (kennung, ordner string, err error)
	Ablegen(kennung, dateiname string, inhalt []byte) (pfad string, groesse int64, err error)
	Verwerfen(kennung string) error
}

// teilfelder halten die Teilungen dieser Sitzung.
type teilfelder struct {
	schloss   sync.Mutex
	stelle    Teilablage
	naht      Teilnaht
	teilungen map[string]teilung

	// generation zählt das Räumen (Nachlauf über Phase 6.7, N1): eine
	// Vorbereitung, die vor einem Wechsel begann und danach endet, trägt
	// nichts mehr ein.
	generation uint64
}

// teilung ist, was zu einer Kennung gehört.
type teilung struct {
	gut  domain.Teilgut
	pfad string

	// profil ist das, unter dem sie vorbereitet wurde — eine Teilung gehört
	// ihrem Profil (Zweitblick über Phase 6.7, Z8).
	profil int64
}

// TeilgutDTO ist, was vor dem Teilen gezeigt wird.
type TeilgutDTO struct {
	Kennung   string `json:"kennung"`
	Art       string `json:"art"`
	Titel     string `json:"titel"`
	Dateiname string `json:"dateiname"`
	Groesse   int64  `json:"groesse"`

	// Text ist, was als Text geht — bei einer Einladung der Code.
	Text string `json:"text"`

	// Mitgehendes und Verluste sind Schlüssel, die das Fenster in seine
	// Sprache setzt (AK-02).
	Mitgehendes []string `json:"mitgehendes"`
	Verluste    []string `json:"verluste"`

	// ZuGrossFuerMail empfiehlt einen anderen Weg als Mail (AK-03).
	ZuGrossFuerMail bool `json:"zuGrossFuerMail"`
}

// TeilenAntwort trägt das Teilgut oder den Fehler.
type TeilenAntwort struct {
	Teilgut *TeilgutDTO `json:"teilgut"`
	Fehler  *Fehler     `json:"fehler"`
}

// MitTeilablage setzt, wo die Kopien liegen — die Zusammensetzung tut es.
func (a *App) MitTeilablage(stelle Teilablage) *App {
	a.teilen.schloss.Lock()
	defer a.teilen.schloss.Unlock()

	a.teilen.stelle = stelle

	return a
}

// MitTeilnaht setzt den Teilen-Dialog — das Fenster tut es beim Start.
func (a *App) MitTeilnaht(naht Teilnaht) *App {
	a.teilen.schloss.Lock()
	defer a.teilen.schloss.Unlock()

	a.teilen.naht = naht

	return a
}

// TeilenVorbereiten legt eine Kopie ab und sagt, was darin steht.
//
// **`bezug` sagt, welches Stück**: die Kennung einer Karte, eines Sets
// oder eines Decks, bei einer Einladung ihr Code, bei der Sammlung nichts.
func (a *App) TeilenVorbereiten(ctx context.Context, art, bezug string) TeilenAntwort {
	ctx = a.imProfil(ctx)

	teilart, fehler := domain.PruefeTeilart(art)
	if fehler != nil {
		return TeilenAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	// **Das Profil und die Generation vom Anfang** (N1): wer erst nach dem
	// Bauen nachsähe, gäbe eine Kopie des vorigen Profils dem nächsten.
	profil := a.profil.Load()

	a.teilen.schloss.Lock()
	stelle, generation := a.teilen.stelle, a.teilen.generation
	a.teilen.schloss.Unlock()

	if stelle == nil {
		return TeilenAntwort{Fehler: Uebersetze(ctx, ohneTeilen())}
	}

	kennung, ordner, fehler := stelle.Ordner()
	if fehler != nil {
		return TeilenAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	gut, pfad, fehler := a.teilgutBauen(ctx, stelle, teilart, bezug, kennung, ordner)
	if fehler != nil {
		_ = stelle.Verwerfen(kennung) //nolint:errcheck // aufgeräumt wird bestenfalls; der Fehler davor zählt

		return TeilenAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	a.teilen.schloss.Lock()
	if a.teilen.generation != generation || a.profil.Load() != profil {
		a.teilen.schloss.Unlock()

		_ = stelle.Verwerfen(kennung) //nolint:errcheck // bestenfalls; geräumt hat der Wechsel

		return TeilenAntwort{Fehler: Uebersetze(ctx, teilungFort())}
	}

	if a.teilen.teilungen == nil {
		a.teilen.teilungen = map[string]teilung{}
	}

	a.teilen.teilungen[kennung] = teilung{gut: gut, pfad: pfad, profil: profil}
	a.teilen.schloss.Unlock()

	return TeilenAntwort{Teilgut: &TeilgutDTO{
		Kennung: kennung, Art: string(gut.Art), Titel: gut.Titel,
		Dateiname: gut.Dateiname, Groesse: gut.Groesse, Text: gut.Text,
		Mitgehendes: mitLeererListe(gut.Art.Mitgehendes()), Verluste: mitLeererListe(gut.Art.Verluste()),
		ZuGrossFuerMail: gut.ZuGrossFuerMail(),
	}}
}

// TeilenAnbieten öffnet den Teilen-Dialog für eine vorbereitete Teilung.
func (a *App) TeilenAnbieten(ctx context.Context, kennung string) TeilenAntwort {
	ctx = a.imProfil(ctx)

	vorbereitet, naht, fehler := a.teilungZu(kennung)
	if fehler != nil {
		return TeilenAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	if naht == nil {
		return TeilenAntwort{Fehler: Uebersetze(ctx, ohneTeilen())}
	}

	if fehler := naht(ctx, vorbereitet.pfad, vorbereitet.gut.Text); fehler != nil {
		return TeilenAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	return TeilenAntwort{}
}

// TeilenVerwerfen räumt eine Teilung weg — auch eine, die es nicht mehr gibt.
func (a *App) TeilenVerwerfen(ctx context.Context, kennung string) TeilenAntwort {
	ctx = a.imProfil(ctx)

	if fehler := domain.PruefeTeilkennung(kennung); fehler != nil {
		return TeilenAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	// **Eine fremde wird nicht angefasst** (Z8) — eine unbekannte schon gar
	// nicht: dann gibt es nichts zu räumen, und das ist kein Fehler.
	a.teilen.schloss.Lock()
	vorbereitet, gibt := a.teilen.teilungen[kennung]

	if gibt && vorbereitet.profil != a.profil.Load() {
		a.teilen.schloss.Unlock()

		return TeilenAntwort{Fehler: Uebersetze(ctx, teilungFort())}
	}

	stelle := a.teilen.stelle
	a.teilen.schloss.Unlock()

	if !gibt || stelle == nil {
		return TeilenAntwort{}
	}

	// **Vergessen erst nach dem geglückten Räumen** (Nachlauf, N4): sonst
	// fände der nächste Versuch nichts mehr, und die Kopie bliebe liegen.
	if fehler := stelle.Verwerfen(kennung); fehler != nil {
		return TeilenAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	a.teilen.schloss.Lock()
	delete(a.teilen.teilungen, kennung)
	a.teilen.schloss.Unlock()

	return TeilenAntwort{}
}

// teilungZu liest eine Teilung samt Dialog — oder sagt, warum es keine gibt.
func (a *App) teilungZu(kennung string) (teilung, Teilnaht, error) {
	if fehler := domain.PruefeTeilkennung(kennung); fehler != nil {
		return teilung{}, nil, fehler
	}

	a.teilen.schloss.Lock()
	defer a.teilen.schloss.Unlock()

	// **Eine fremde ist keine** — dieselbe Antwort wie eine verworfene, damit
	// ein anderes Profil nicht erfährt, dass es sie gibt.
	vorbereitet, gibt := a.teilen.teilungen[kennung]
	if !gibt || vorbereitet.profil != a.profil.Load() {
		return teilung{}, nil, teilungFort()
	}

	return vorbereitet, a.teilen.naht, nil
}

// teilungenRaeumen verwirft jede Teilung samt Kopie — beim Wechsel, Sperren
// und Abmelden, wie alles, was einem Profil gehört (Z8).
//
// **Bestenfalls**: eine Kopie, die sich nicht wegräumen lässt, räumt der
// nächste Start nach Ablauf; vergessen ist sie hier trotzdem.
func (a *App) teilungenRaeumen() {
	a.teilen.schloss.Lock()
	stelle, teilungen := a.teilen.stelle, a.teilen.teilungen
	a.teilen.teilungen = nil
	a.teilen.generation++
	a.teilen.schloss.Unlock()

	for kennung := range teilungen {
		_ = stelle.Verwerfen(kennung) //nolint:errcheck // siehe darüber
	}
}

// teilungFort ist die Antwort auf eine Kennung, die es hier nicht (mehr) gibt.
func teilungFort() error {
	return domain.MitGrund(domain.ErrNotFound, nil,
		"diese Teilung ist verworfen oder abgelaufen — bereite sie neu vor")
}

// ohneTeilen sagt, dass es hier keinen Teilen-Dialog gibt.
func ohneTeilen() error {
	return domain.MitGrund(domain.ErrValidation, nil,
		"Teilen gibt es nur im Fenster — gib die Datei über ihren Ausfuhrweg aus")
}

// mitLeererListe macht aus `nil` eine leere Liste — die Zusage aller
// Antworten dieses Pakets.
func mitLeererListe(liste []string) []string {
	if liste == nil {
		return []string{}
	}

	return liste
}
