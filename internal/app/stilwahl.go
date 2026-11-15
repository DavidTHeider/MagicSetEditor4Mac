package app

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Stilwaehler beantwortet die Fragen der Stilauswahl — die Naht zu `service`.
type Stilwaehler interface {
	Stile(ctx context.Context, setID int64) ([]service.Stilangebot, error)
	Folgen(ctx context.Context, setID int64, paket string) (*service.Folgen, error)
	Wechseln(ctx context.Context, setID int64, paket string) error
	Optionen(ctx context.Context, setID int64) ([]service.Stiloption, error)
	Kartenfelder(ctx context.Context, setID int64) ([]domain.Field, error)

	// Spielskript ist das `init script:` des Spiels — die Kartenliste
	// braucht es, um berechnete Spalten zu füllen (#1238).
	Spielskript(ctx context.Context, setID int64) (string, error)

	// Setfelder sind die `set field:`-Blöcke des Spiels — was ein **Set** an
	// Angaben führt (#567).
	Setfelder(ctx context.Context, setID int64) ([]domain.Field, error)
	OptionSetzen(ctx context.Context, setID int64, feld, wert string) error

	// OptionZuruecksetzen entfernt den gespeicherten Wert einer Option.
	//
	// **Eine eigene Naht und kein leerer Wert an [Stilwaehler.OptionSetzen]**:
	// leer ist bei einer Textoption ein gültiger Wert, und zwei Bedeutungen an
	// einem Aufruf wären später nicht mehr auseinanderzuhalten.
	OptionZuruecksetzen(ctx context.Context, setID int64, feld string) error

	// Stilpaket gibt ein Stilpaket im Spiel des Sets — leer heißt sein
	// eigener Stil. Die Ausfuhr baut daraus `set.stylesheet` und
	// `card.stylesheet` (#2039).
	Stilpaket(ctx context.Context, setID int64, paket string) (msepaket.Stil, error)
}

// StilDTO ist ein wählbarer Stil.
type StilDTO struct {
	Paket string `json:"paket"`

	// Anzeigename ist, was über dem Stil steht — der Kurzname,
	// großgeschrieben wie MSE2 (UX-05, S07.NAT-02, #1677).
	Anzeigename string `json:"anzeigename"`

	Spiel        string `json:"spiel"`
	Version      string `json:"version"`
	EigenesSpiel bool   `json:"eigenesSpiel"`
	Gewaehlt     bool   `json:"gewaehlt"`
}

// StileAntwort trägt die wählbaren Stile.
type StileAntwort struct {
	Stile  []StilDTO `json:"stile"`
	Fehler *Fehler   `json:"fehler"`
}

// FolgenDTO sagt, was ein Wechsel kostet.
//
// **Angaben, kein Satz** (§9): der Satz steht in den Sprachkatalogen des
// Frontends. Ein deutscher Fließtext von hier wäre in der englischen Oberfläche
// deutsch.
type FolgenDTO struct {
	UngedeuteteFelder []string `json:"ungedeuteteFelder"`
	VerloreneOptionen []string `json:"verloreneOptionen"`
	Spielwechsel      bool     `json:"spielwechsel"`
}

// FolgenAntwort trägt die Folgen eines geplanten Wechsels.
type FolgenAntwort struct {
	Folgen *FolgenDTO `json:"folgen"`
	Fehler *Fehler    `json:"fehler"`
}

// StiloptionDTO ist eine Stiloption samt Wert.
type StiloptionDTO struct {
	Name string `json:"name"`
	Art  string `json:"art"`

	// Beschreibung kommt aus dem Paket und wird **nicht** übersetzt (§9).
	Beschreibung string `json:"beschreibung"`
	Wert         string `json:"wert"`

	// Gesetzt sagt, ob [StiloptionDTO.Wert] aus dem Bestand kommt oder aus der
	// Vorgabe. Vorgabe ist der gerechnete Wert aus `default:` — **leer**, wenn
	// die Vorlage keinen nennt.
	//
	// **Zwei Angaben für einen Eintrag**: `Vorgabe` sagt, ob es ihn gibt und
	// welche Farbe er zeigt, `Gesetzt`, ob er gewählt ist. MSE blendet ihn ohne
	// `default_script` ganz aus und zeichnet seinen Tupfer aus `default_color`
	// (`src/gui/value/color.cpp`, Zeilen 38 und 107–112, Fassung f9d9356 vom
	// 2021-07-06) — #764.
	Gesetzt bool `json:"gesetzt"`

	// HatVorgabe sagt, ob die Vorlage überhaupt ein `default:` nennt —
	// **nicht dasselbe wie ein nichtleerer Vorgabewert**. Ein Skript darf
	// scheitern; MSE zeigt seinen Eintrag trotzdem
	// (`src/gui/value/color.cpp`, Zeile 38, Fassung f9d9356 vom 2021-07-06).
	HatVorgabe bool   `json:"hatVorgabe"`
	Vorgabe    string `json:"vorgabe"`

	// Moeglichkeiten sind die Wahlen der Option — Wert und Beschriftung.
	//
	// **Eine Liste für jede Art, die eine Auswahl ist**: bei `choice` und
	// `boolean` sind Wert und Beschriftung dasselbe, bei einer `package choice`
	// und bei einer Farbe nicht.
	//
	// **Ohne `omitempty`**: eine leere Liste, die aus der Antwort
	// verschwindet, kommt im Fenster als `undefined` an — §3 nennt das unter
	// „was die Tests hier nicht sehen werden".
	Moeglichkeiten []domain.Wahlmoeglichkeit `json:"moeglichkeiten"`

	// HatWahlbilder sagt, ob es zu den Wahlen Bilder gibt
	// (#1225).
	//
	// **Damit die Oberfläche nur dort nachfragt, wo es etwas zu holen gibt.**
	// Ohne die Auskunft riefe sie [App.Stilwahlbilder] für jede Option — und
	// bekäme bei `magic-m15` neunmal von dreizehn eine leere Liste.
	HatWahlbilder bool `json:"hatWahlbilder"`

	// EigeneFarbe sagt, ob eine Farbe außerhalb der Liste erlaubt ist — bei
	// jeder anderen Art als `color` falsch (#634).
	EigeneFarbe bool `json:"eigeneFarbe"`
}

// StiloptionenAntwort trägt die Optionen des gewählten Stils.
type StiloptionenAntwort struct {
	Optionen []StiloptionDTO `json:"optionen"`
	Fehler   *Fehler         `json:"fehler"`
}

// Stile nennt die Stile, die für ein Set zur Wahl stehen.
func (a *App) Stile(ctx context.Context, setID string) StileAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := kennungAus(setID)
	if befund != nil {
		return StileAntwort{Stile: []StilDTO{}, Fehler: befund}
	}

	angebote, err := a.stile.Stile(ctx, kennung)
	if err != nil {
		return StileAntwort{Stile: []StilDTO{}, Fehler: Uebersetze(ctx, err)}
	}

	stile := make([]StilDTO, 0, len(angebote))
	for _, angebot := range angebote {
		stile = append(stile, StilDTO{
			Paket:        angebot.Paket,
			Anzeigename:  angebot.Anzeigename,
			Spiel:        angebot.Spiel,
			Version:      angebot.Version,
			EigenesSpiel: angebot.EigenesSpiel,
			Gewaehlt:     angebot.Gewaehlt,
		})
	}

	return StileAntwort{Stile: stile}
}

// Stilfolgen sagt, was ein Wechsel kostet — **bevor** er stattfindet (#41).
func (a *App) Stilfolgen(ctx context.Context, setID, paket string) FolgenAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := kennungAus(setID)
	if befund != nil {
		return FolgenAntwort{Fehler: befund}
	}

	folgen, err := a.stile.Folgen(ctx, kennung, paket)
	if err != nil {
		return FolgenAntwort{Fehler: Uebersetze(ctx, err)}
	}

	return FolgenAntwort{Folgen: &FolgenDTO{
		UngedeuteteFelder: folgen.UngedeuteteFelder,
		VerloreneOptionen: folgen.VerloreneOptionen,
		Spielwechsel:      folgen.Spielwechsel,
	}}
}

// StilWechseln setzt den Stil eines Sets.
func (a *App) StilWechseln(ctx context.Context, setID, paket string) SchreibAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := kennungAus(setID)
	if befund != nil {
		return SchreibAntwort{Fehler: befund}
	}

	return SchreibAntwort{Fehler: Uebersetze(ctx, a.stile.Wechseln(ctx, kennung, paket))}
}

// Stiloptionen gibt die Optionen des gewählten Stils samt Werten.
func (a *App) Stiloptionen(ctx context.Context, setID string) StiloptionenAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := kennungAus(setID)
	if befund != nil {
		return StiloptionenAntwort{Optionen: []StiloptionDTO{}, Fehler: befund}
	}

	gelesen, err := a.stile.Optionen(ctx, kennung)
	if err != nil {
		return StiloptionenAntwort{Optionen: []StiloptionDTO{}, Fehler: Uebersetze(ctx, err)}
	}

	optionen := make([]StiloptionDTO, 0, len(gelesen))
	// **Über den Index**: eine Stiloption ist über 128 Byte groß, seit ein
	// Farbfeld seine Wahlfarben mitführt.
	for i := range gelesen {
		optionen = append(optionen, alsStiloptionDTO(&gelesen[i]))
	}

	return StiloptionenAntwort{Optionen: optionen}
}

// alsStiloptionDTO macht aus einer Stiloption ihr DTO.
//
// **Eine Abbildung für beide Wege.** Das Fenster ([App.Stiloptionen]) und die
// drei anderen Flächen ([Stilweg.Optionen]) haben verschiedene Leser und
// dieselbe Antwort; zwei Kopien dieser Zuordnung liefen auseinander, sobald
// eine von beiden ein Feld dazubekäme (§10) — und genau das ist beim
// Wahlbild-Schalter beinahe passiert
// (#1225).
func alsStiloptionDTO(option *service.Stiloption) StiloptionDTO {
	// **Eine leere Liste und kein `nil`.** Ein `nil`-Slice käme über die
	// Wails-Bindung als `null` an, und das Frontend prüfte auf zwei Gestalten
	// desselben Nichts.
	moeglichkeiten := option.Moeglichkeiten
	if moeglichkeiten == nil {
		moeglichkeiten = []domain.Wahlmoeglichkeit{}
	}

	return StiloptionDTO{
		Name:           option.Name,
		Art:            string(option.Kind),
		Beschreibung:   option.Description,
		Wert:           option.Wert,
		Gesetzt:        option.Gesetzt,
		HatVorgabe:     option.HatVorgabe,
		Vorgabe:        option.Vorgabe,
		Moeglichkeiten: moeglichkeiten,
		HatWahlbilder:  option.HatWahlbilder,
		EigeneFarbe:    option.Kind == domain.KindColor && option.EigeneFarbe,
	}
}

// StiloptionSetzen schreibt den Wert einer Stiloption.
func (a *App) StiloptionSetzen(ctx context.Context, setID, feld, wert string) SchreibAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := kennungAus(setID)
	if befund != nil {
		return SchreibAntwort{Fehler: befund}
	}

	return SchreibAntwort{Fehler: Uebersetze(ctx, a.stile.OptionSetzen(ctx, kennung, feld, wert))}
}

// StiloptionZuruecksetzen stellt eine Stiloption auf ihre Vorgabe zurück.
func (a *App) StiloptionZuruecksetzen(ctx context.Context, setID, feld string) SchreibAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := kennungAus(setID)
	if befund != nil {
		return SchreibAntwort{Fehler: befund}
	}

	return SchreibAntwort{Fehler: Uebersetze(ctx, a.stile.OptionZuruecksetzen(ctx, kennung, feld))}
}

// MitStilwahl setzt die Naht zur Stilauswahl.
func (a *App) MitStilwahl(waehler Stilwaehler) { a.stile = waehler }
