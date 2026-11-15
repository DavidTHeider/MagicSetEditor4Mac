package app

// **Aufmerksamkeit aus echten Befunden** — Werkstück U11 von Phase 5.1
// (#1422).
//
// # Drei Quellen, die es an allen vier Flächen schon gibt
//
// Die Vorbehalte eines Sets (`Vorbehalte`), die fehlenden Karten eines Decks
// (`DeckAbgleichen`) und der Skriptbefund einer Paketdatei
// (`PaketdateiPruefen`) sind je Objekt an jeder Fläche zu haben. Die
// Übersicht **sammelt** sie nur — über alle Sets, Decks und bearbeitbaren
// Pakete des Profils — und ordnet sie nach `domain.OrdneHinweise`. Deshalb
// steht sie nur am Fenster (`FAEHIGKEITEN.md`).
//
// # Eine Quelle, die scheitert, hält die anderen nicht an
//
// Dieselbe Regel wie beim Dienstkatalog: ein Deck, das sich nicht abgleichen
// lässt, wird als **Befund** genannt, und die Sets stehen trotzdem da. Nur
// was für alle gilt — kein Profil, kein Bestand — scheitert ganz.

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// HinweisDTO ist ein Hinweis, wie das Fenster ihn sieht.
type HinweisDTO struct {
	Art    string           `json:"art"`
	Objekt ArbeitsobjektDTO `json:"objekt"`
	Anzahl int              `json:"anzahl"`

	// Grund ist der deutsche Satz — der Rückfall, wenn das Fenster den
	// Schlüssel nicht kennt (§9).
	Grund string `json:"grund"`

	// **Schlüssel, Werte und Ort daneben**
	// (#1963):
	// bis dahin stand nur der deutsche Satz da, und das Fenster zeigte ihn
	// auch in englischer Sprache. `werte` ist eine Liste und nie `null`.
	Ort        string   `json:"ort"`
	Schluessel string   `json:"schluessel"`
	Werte      []string `json:"werte"`
}

// DurchgesehenDTO sagt, worüber geschaut wurde.
type DurchgesehenDTO struct {
	Sets   int `json:"sets"`
	Decks  int `json:"decks"`
	Pakete int `json:"pakete"`
}

// AufmerksamkeitAntwort ist die Übersicht der Hinweise samt Umschlag.
type AufmerksamkeitAntwort struct {
	Hinweise     []HinweisDTO    `json:"hinweise"`
	Durchgesehen DurchgesehenDTO `json:"durchgesehen"`
	// Befunde nennt, welche Quelle nicht antwortete — je eine Zeile.
	Befunde []string `json:"befunde"`
	Fehler  *Fehler  `json:"fehler"`
}

// Aufmerksamkeit sammelt die Hinweise des Profils.
func (a *App) Aufmerksamkeit(ctx context.Context) AufmerksamkeitAntwort {
	ctx = a.imProfil(ctx)

	antwort := AufmerksamkeitAntwort{Hinweise: []HinweisDTO{}, Befunde: []string{}}

	if a.bestand == nil {
		antwort.Fehler = Uebersetze(ctx, errors.New("aufmerksamkeit: der Bestand ist nicht eingerichtet"))

		return antwort
	}

	// Die Reihenfolge ist die der Quellen — Sets, Decks, Pakete —, in den
	// Hinweisen wie in den Befunden.
	sets := a.setHinweise(ctx)
	decks := a.deckHinweise(ctx)
	pakete := a.paketHinweise(ctx)

	antwort.Durchgesehen.Sets = len(sets.durchgesehen)
	antwort.Durchgesehen.Decks = decks.gezaehlt
	antwort.Durchgesehen.Pakete = pakete.gezaehlt

	hinweise := make([]domain.Hinweis, 0, len(sets.hinweise)+len(decks.hinweise)+len(pakete.hinweise))
	hinweise = append(hinweise, sets.hinweise...)
	hinweise = append(hinweise, decks.hinweise...)
	hinweise = append(hinweise, pakete.hinweise...)

	antwort.Befunde = append(antwort.Befunde, sets.befunde...)
	antwort.Befunde = append(antwort.Befunde, decks.befunde...)
	antwort.Befunde = append(antwort.Befunde, pakete.befunde...)

	for _, eines := range domain.OrdneHinweise(hinweise) {
		antwort.Hinweise = append(antwort.Hinweise, HinweisDTO{
			Art: string(eines.Art),
			Objekt: ArbeitsobjektDTO{
				Art: string(eines.Objekt.Art), Kennung: eines.Objekt.Kennung, Name: eines.Objekt.Name,
			},
			Anzahl: eines.Anzahl, Grund: eines.Grund,
			Ort: eines.Ort, Schluessel: eines.Schluessel,
			Werte: append([]string{}, eines.Werte...),
		})
	}

	return antwort
}

// hinweisernte ist, was eine Quelle beisteuert — Hinweise, Gezähltes und die
// Befunde über das, was sich nicht durchsehen ließ.
type hinweisernte struct {
	befunde      []string
	hinweise     []domain.Hinweis
	durchgesehen []int64
	gezaehlt     int
}

// setHinweise zählt die Vorbehalte je Set — die Blöcke, die niemand deutet.
func (a *App) setHinweise(ctx context.Context) hinweisernte {
	koepfe, err := a.bestand.Setkoepfe(ctx)
	if err != nil {
		return hinweisernte{befunde: []string{"Sets: " + err.Error()}}
	}

	var ernte hinweisernte

	for _, kopf := range koepfe {
		ernte.durchgesehen = append(ernte.durchgesehen, kopf.ID)

		bloecke, err := a.bestand.Rohbloecke(ctx, kopf.ID)
		if err != nil {
			ernte.befunde = append(ernte.befunde, kopf.Name+": "+err.Error())

			continue
		}

		if len(bloecke) == 0 {
			continue
		}

		// **„Zeile" ist ein Wort und kein Bezeichner** (#1963): der Satz
		// steht im Katalog unter [domain.GrundVorbehalt], Datei und Zeile
		// reisen als Werte.
		ernte.hinweise = append(ernte.hinweise, domain.Hinweis{
			Art:        domain.HinweisSetVorbehalte,
			Objekt:     domain.Arbeitsobjekt{Art: domain.ArbeitsartSet, Kennung: strconv.FormatInt(kopf.ID, 10), Name: kopf.Name},
			Anzahl:     len(bloecke),
			Grund:      fmt.Sprintf("%s, Zeile %d", bloecke[0].File, bloecke[0].Line),
			Schluessel: domain.GrundVorbehalt,
			Werte:      []string{bloecke[0].File, strconv.Itoa(bloecke[0].Line)},
		})
	}

	return ernte
}

// deckHinweise zählt je Deck, was in der Sammlung fehlt.
func (a *App) deckHinweise(ctx context.Context) hinweisernte {
	if a.decks == nil || a.deckabgleich == nil {
		return hinweisernte{}
	}

	decks, err := a.decks.Decks(ctx, "")
	if err != nil {
		return hinweisernte{befunde: []string{"Decks: " + err.Error()}}
	}

	ernte := hinweisernte{gezaehlt: len(decks)}

	for _, deck := range decks {
		stand, err := a.deckabgleich.Fuer(ctx, deck.ID)
		if err != nil {
			ernte.befunde = append(ernte.befunde, deck.Name+": "+err.Error())

			continue
		}

		if stand.Fehlend == 0 {
			continue
		}

		grund := ""
		for _, zeile := range stand.Zeilen {
			if zeile.Fehlend > 0 {
				grund = zeile.Bezeichnung

				break
			}
		}

		ernte.hinweise = append(ernte.hinweise, domain.Hinweis{
			Art:    domain.HinweisDeckFehlend,
			Objekt: domain.Arbeitsobjekt{Art: domain.ArbeitsartDeck, Kennung: strconv.FormatInt(deck.ID, 10), Name: deck.Name},
			Anzahl: stand.Fehlend,
			Grund:  grund,
		})
	}

	return ernte
}

// paketHinweise prüft die Skripte jeder Paketdatei — nur der bearbeitbaren.
func (a *App) paketHinweise(ctx context.Context) hinweisernte {
	if a.pakete == nil || a.paketdateien == nil {
		return hinweisernte{}
	}

	pakete, err := a.pakete.Liste(ctx)
	if err != nil {
		return hinweisernte{befunde: []string{"Pakete: " + err.Error()}}
	}

	var ernte hinweisernte

	for _, paket := range pakete {
		// **Über den Ort, nicht über den Namen.** Die Werkstatt löst ein Paket
		// in ihrem Verzeichnis auf; die mitgelieferte Saat liegt im Bündel, und
		// ihr Name allein ergäbe dort „nicht gefunden" — die Sichtprüfung am
		// 2026-09-09 führte so sieben Saatpakete als Befund. Mit dem Ort sagt
		// die Werkstatt, was zutrifft: liegt nicht im Paketverzeichnis.
		texte, err := a.paketdateien.Texte(ctx, paketort(paket))
		if err != nil {
			// **Mitgelieferte Pakete sind nicht bearbeitbar** — die Werkstatt
			// weist sie ab, und das ist kein Befund der Übersicht.
			if errors.Is(err, domain.ErrValidation) {
				continue
			}

			ernte.befunde = append(ernte.befunde, paket.Name+": "+err.Error())

			continue
		}

		ernte.gezaehlt++

		hinweis := domain.Hinweis{
			Art:    domain.HinweisPaketBefund,
			Objekt: domain.Arbeitsobjekt{Art: domain.ArbeitsartPaket, Kennung: paket.Name, Name: paket.Name},
		}

		for _, text := range texte {
			fehler := service.PaketSkriptfehler(text.Inhalt)
			if fehler == nil {
				continue
			}

			hinweis.Anzahl++

			// **Der erste Befund ist der Grund — mit Schlüssel und Werten**
			// (#1963), denselben, die die Prüfung der Paketdatei dem Fenster
			// gibt. Der Name der Datei ist ein Bezeichner und steht als Ort
			// davor.
			if hinweis.Anzahl == 1 {
				hinweis.Grund = strings.TrimSpace(text.Name + ": " + fehler.Meldung)
				hinweis.Ort, hinweis.Schluessel = text.Name, string(fehler.Schluessel)
				hinweis.Werte = fehler.Werte
			}
		}

		if hinweis.Anzahl == 0 {
			continue
		}

		ernte.hinweise = append(ernte.hinweise, hinweis)
	}

	return ernte
}

// Aufmerksamkeit — die Fensterbindung.
func (b *Bindungen) Aufmerksamkeit(ctx context.Context) AufmerksamkeitAntwort {
	return b.app(ctx).Aufmerksamkeit(b.ruf(ctx))
}

// paketort ist, worüber die Werkstatt ein Paket findet: sein Ort auf der
// Platte, und nur ohne ihn der Name.
func paketort(paket domain.Paket) string {
	if paket.Pfad != "" {
		return paket.Pfad
	}

	return paket.Name
}
