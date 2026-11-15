package app

// **Eigene Übersetzungen an jeder Fläche** — LOC-02
// (#1688).
//
// Die Anwendung übersetzt nur: zwischen dem Dienst und einer Antwort, die
// Fenster, Kommandozeile, REST und MCP gleich lesen. Geprüft wird im Dienst.

import (
	"context"
	"fmt"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// SprachpaketDTO ist ein Paket mit seinen Zahlen.
type SprachpaketDTO struct {
	Kennung   string `json:"kennung"`
	Name      string `json:"name"`
	Basis     string `json:"basis"`
	Stand     string `json:"stand"`
	Lizenz    string `json:"lizenz"`
	Herkunft  string `json:"herkunft"`
	Geaendert string `json:"geaendert"`

	// Gesamt ist die Zahl der Sätze der Vorlage, Uebersetzt die der
	// übersetzten, Fehlend die ohne Übersetzung, Veraltet die mit einem
	// geänderten deutschen Satz und Unbekannt die, die diese Fassung nicht
	// kennt.
	Gesamt     int `json:"gesamt"`
	Uebersetzt int `json:"uebersetzt"`
	Fehlend    int `json:"fehlend"`
	Veraltet   int `json:"veraltet"`
	Unbekannt  int `json:"unbekannt"`
}

// UebersetzungstextDTO ist ein übersetzter Satz und der deutsche, gegen den
// er übersetzt wurde.
type UebersetzungstextDTO struct {
	Text   string `json:"text"`
	Quelle string `json:"quelle"`
}

// SprachpaketlisteAntwort sind alle Pakete — ohne ihre Sätze.
type SprachpaketlisteAntwort struct {
	Pakete []SprachpaketDTO `json:"pakete"`
	Fehler *Fehler          `json:"fehler"`
}

// SprachpaketAntwort ist ein Paket mit seinen Sätzen und dem, was fehlt.
type SprachpaketAntwort struct {
	Paket           SprachpaketDTO                  `json:"paket"`
	Texte           map[string]UebersetzungstextDTO `json:"texte"`
	FehlendePfade   []string                        `json:"fehlendePfade"`
	VeraltetePfade  []string                        `json:"veraltetePfade"`
	UnbekanntePfade []string                        `json:"unbekanntePfade"`

	// Pfad ist der Ort einer Datei, wo eine im Spiel ist; Abgebrochen sagt,
	// dass niemand einen gewählt hat — kein Fehler.
	Pfad        string `json:"pfad"`
	Abgebrochen bool   `json:"abgebrochen"`

	Fehler *Fehler `json:"fehler"`
}

// SprachpaketdateiAntwort ist eine ausgegebene Übersetzung als Text.
type SprachpaketdateiAntwort struct {
	Name   string  `json:"name"`
	Inhalt string  `json:"inhalt"`
	Fehler *Fehler `json:"fehler"`
}

// SprachpaketangabenDTO sind die Angaben, die sich ändern sollen — `null`
// heißt: bleibt.
type SprachpaketangabenDTO struct {
	Name     *string `json:"name"`
	Stand    *string `json:"stand"`
	Lizenz   *string `json:"lizenz"`
	Herkunft *string `json:"herkunft"`
}

// MitSprachpaketen setzt die Naht.
func (a *App) MitSprachpaketen(pakete *service.Sprachpakete) { a.sprachpakete = pakete }

// errOhneSprachpakete ist der Befund einer fehlenden Naht.
var errOhneSprachpakete = fmt.Errorf("%w: die eigenen Übersetzungen sind nicht eingerichtet",
	domain.ErrValidation)

// Sprachpakete gibt alle eigenen Übersetzungen.
func (a *App) Sprachpakete(ctx context.Context) SprachpaketlisteAntwort {
	ctx = a.imProfil(ctx)

	if a.sprachpakete == nil {
		return SprachpaketlisteAntwort{Pakete: []SprachpaketDTO{}, Fehler: Uebersetze(ctx, errOhneSprachpakete)}
	}

	liste, err := a.sprachpakete.Liste(ctx)
	if err != nil {
		return SprachpaketlisteAntwort{Pakete: []SprachpaketDTO{}, Fehler: Uebersetze(ctx, err)}
	}

	pakete := make([]SprachpaketDTO, 0, len(liste))
	for _, auskunft := range liste {
		pakete = append(pakete, sprachpaketAlsDTO(auskunft))
	}

	return SprachpaketlisteAntwort{Pakete: pakete}
}

// Sprachpaket gibt eine Übersetzung mit ihren Sätzen.
func (a *App) Sprachpaket(ctx context.Context, kennung string) SprachpaketAntwort {
	ctx = a.imProfil(ctx)

	return a.sprachweg(ctx, func(dienst *service.Sprachpakete) (service.Sprachpaketauskunft, error) {
		return dienst.Zeige(ctx, kennung)
	})
}

// SprachpaketAnlegen legt eine Übersetzung über einer Basis an.
func (a *App) SprachpaketAnlegen(ctx context.Context, kennung, name, basis string) SprachpaketAntwort {
	ctx = a.imProfil(ctx)

	return a.sprachweg(ctx, func(dienst *service.Sprachpakete) (service.Sprachpaketauskunft, error) {
		return dienst.Anlegen(ctx, kennung, name, basis)
	})
}

// SprachpaketAendern setzt Angaben und Sätze — ein leerer Satz nimmt die
// Übersetzung zurück.
func (a *App) SprachpaketAendern(
	ctx context.Context, kennung string, angaben SprachpaketangabenDTO, texte map[string]string,
) SprachpaketAntwort {
	ctx = a.imProfil(ctx)

	return a.sprachweg(ctx, func(dienst *service.Sprachpakete) (service.Sprachpaketauskunft, error) {
		return dienst.Aendern(ctx, kennung, service.Sprachpaketangaben{
			Name: angaben.Name, Stand: angaben.Stand, Lizenz: angaben.Lizenz, Herkunft: angaben.Herkunft,
		}, texte)
	})
}

// SprachpaketLoeschen entfernt eine Übersetzung.
func (a *App) SprachpaketLoeschen(ctx context.Context, kennung string) SchreibAntwort {
	ctx = a.imProfil(ctx)

	if a.sprachpakete == nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, errOhneSprachpakete)}
	}

	return SchreibAntwort{Fehler: Uebersetze(ctx, a.sprachpakete.Loeschen(ctx, kennung))}
}

// SprachpaketAusgabeSchreiben legt eine Übersetzung an einen Ort.
func (a *App) SprachpaketAusgabeSchreiben(ctx context.Context, kennung, pfad string) SprachpaketAntwort {
	ctx = a.imProfil(ctx)

	antwort := a.sprachweg(ctx, func(dienst *service.Sprachpakete) (service.Sprachpaketauskunft, error) {
		if _, err := dienst.AusgebenNach(ctx, kennung, pfad); err != nil {
			return service.Sprachpaketauskunft{}, err
		}

		return dienst.Zeige(ctx, kennung)
	})

	return mitPfad(antwort, pfad)
}

// SprachpaketAusgabeText gibt eine Übersetzung als Text — für die Flächen,
// an denen keine Datei im Spiel ist.
func (a *App) SprachpaketAusgabeText(ctx context.Context, kennung string) SprachpaketdateiAntwort {
	ctx = a.imProfil(ctx)

	if a.sprachpakete == nil {
		return SprachpaketdateiAntwort{Fehler: Uebersetze(ctx, errOhneSprachpakete)}
	}

	name, inhalt, err := a.sprachpakete.Ausgeben(ctx, kennung)
	if err != nil {
		return SprachpaketdateiAntwort{Fehler: Uebersetze(ctx, err)}
	}

	return SprachpaketdateiAntwort{Name: name, Inhalt: string(inhalt)}
}

// SprachpaketEinlesenLesen nimmt eine Übersetzung von einem Ort auf.
func (a *App) SprachpaketEinlesenLesen(ctx context.Context, pfad string, ersetzen bool) SprachpaketAntwort {
	ctx = a.imProfil(ctx)

	antwort := a.sprachweg(ctx, func(dienst *service.Sprachpakete) (service.Sprachpaketauskunft, error) {
		return dienst.EinlesenAus(ctx, pfad, ersetzen)
	})

	return mitPfad(antwort, pfad)
}

// SprachpaketEinlesenText nimmt eine Übersetzung aus einem Text auf.
func (a *App) SprachpaketEinlesenText(ctx context.Context, inhalt string, ersetzen bool) SprachpaketAntwort {
	ctx = a.imProfil(ctx)

	return a.sprachweg(ctx, func(dienst *service.Sprachpakete) (service.Sprachpaketauskunft, error) {
		return dienst.Einlesen(ctx, []byte(inhalt), ersetzen)
	})
}

// sprachweg ruft den Dienst und übersetzt seine Antwort — **nie `null`**
// in einer Liste, auch nicht im Fehlerfall.
func (a *App) sprachweg(
	ctx context.Context, weg func(*service.Sprachpakete) (service.Sprachpaketauskunft, error),
) SprachpaketAntwort {
	if a.sprachpakete == nil {
		return leeresSprachpaket(Uebersetze(ctx, errOhneSprachpakete))
	}

	auskunft, err := weg(a.sprachpakete)
	if err != nil {
		return leeresSprachpaket(Uebersetze(ctx, err))
	}

	texte := make(map[string]UebersetzungstextDTO, len(auskunft.Paket.Texte))
	for pfad, text := range auskunft.Paket.Texte {
		texte[pfad] = UebersetzungstextDTO(text)
	}

	return SprachpaketAntwort{
		Paket: sprachpaketAlsDTO(auskunft), Texte: texte,
		FehlendePfade: auskunft.Abgleich.Fehlend, VeraltetePfade: auskunft.Abgleich.Veraltet,
		UnbekanntePfade: auskunft.Abgleich.Unbekannt,
	}
}

// leeresSprachpaket ist eine Antwort mit einem Befund und leeren Listen.
func leeresSprachpaket(fehler *Fehler) SprachpaketAntwort {
	return SprachpaketAntwort{
		Texte: map[string]UebersetzungstextDTO{}, FehlendePfade: []string{},
		VeraltetePfade: []string{}, UnbekanntePfade: []string{}, Fehler: fehler,
	}
}

// mitPfad nennt den Ort, wenn der Weg glückte.
func mitPfad(antwort SprachpaketAntwort, pfad string) SprachpaketAntwort {
	if antwort.Fehler == nil {
		antwort.Pfad = pfad
	}

	return antwort
}

// sprachpaketAlsDTO übersetzt die Angaben und Zahlen eines Pakets.
func sprachpaketAlsDTO(auskunft service.Sprachpaketauskunft) SprachpaketDTO {
	paket, abgleich := auskunft.Paket, auskunft.Abgleich

	return SprachpaketDTO{
		Kennung: paket.Kennung, Name: paket.Name, Basis: paket.Basis, Stand: string(paket.Stand),
		Lizenz: paket.Lizenz, Herkunft: paket.Herkunft,
		Geaendert: paket.Geaendert.UTC().Format(time.RFC3339),
		Gesamt:    abgleich.Gesamt, Uebersetzt: abgleich.Uebersetzt, Fehlend: len(abgleich.Fehlend),
		Veraltet: len(abgleich.Veraltet), Unbekannt: len(abgleich.Unbekannt),
	}
}
