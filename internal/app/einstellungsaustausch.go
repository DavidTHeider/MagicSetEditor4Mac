// **Einstellungen ausführen und einlesen, am Fenster** — Werkstück O3 von
// Phase 5.2 (#1427).
//
// # Warum der Text durch das Fenster geht und nicht der Pfad
//
// **Weil eine Datei zwischen Vorschau und Übernahme eine andere sein kann.**
// Wer den Pfad zweimal läse, zeigte einen Diff über die Datei von vorhin und
// schriebe die von jetzt. Der Text, den die Vorschau gezeigt hat, ist der,
// über den entschieden wurde — und genau der geht zurück.
//
// **Und weil es die drei Flächen ohne Fenster gleich macht**: dort gibt es
// keinen Dialog, und der Text ist das, was ein Befehl übergibt.

package app

import (
	"context"
	"fmt"
	"os"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Einstellungsaustauscher ist die Naht zu `internal/service`.
type Einstellungsaustauscher interface {
	Ausfuhr(ctx context.Context, abschnitte []string) (domain.Austausch, error)
	Vorschau(ctx context.Context, roh []byte) (service.Einfuhrvorschau, error)
	Uebernehmen(
		ctx context.Context, roh []byte, revision string, nur []string,
	) (service.Einfuhrergebnis, error)
}

// **Und es gibt kein `MitEinstellungsaustausch`.**
//
// Die anderen Nähte tragen einen Setzer, weil ein Test sie durch ein Doppel
// ersetzt. Hier braucht das keiner: die Läufe gehen über eine **angelaufene**
// Anwendung, und ein Setzer, den niemand ruft, wäre eine Naht für einen Fall,
// den es nicht gibt (§2). Gesetzt wird sie in `verdrahtung.go`.

// EinstellungsausfuhrAntwort trägt den Text eines Austauschs.
type EinstellungsausfuhrAntwort struct {
	// Inhalt ist das JSON, wie es in eine Datei ginge.
	Inhalt string  `json:"inhalt"`
	Fehler *Fehler `json:"fehler"`
}

// EinfuhrzeileDTO ist eine Zeile des Diffs.
type EinfuhrzeileDTO struct {
	Ebene      string `json:"ebene"`
	Schluessel string `json:"schluessel"`
	Vorher     string `json:"vorher"`
	Nachher    string `json:"nachher"`
	Herkunft   string `json:"herkunft"`

	// Wirkung sagt, wann der neue Wert gilt — OPS-01.AK-02 (#1718).
	Wirkung string `json:"wirkung"`
}

// EinfuhrbefundDTO ist, was an einer Zeile nicht stimmt.
type EinfuhrbefundDTO struct {
	Ebene      string `json:"ebene"`
	Schluessel string `json:"schluessel"`
	Grund      string `json:"grund"`
}

// Einstellungssicht ist, was eine Datei ändern würde — ohne Fehlerfeld.
//
// **Getrennt von der Antwort**, weil sie zwei Verbraucher hat: das Fenster
// bekommt sie in einer Antwort mit `Fehler`, die drei Flächen ohne Fenster
// bekommen sie mit einem `error` daneben. Zwei gleichlautende Verbünde wären
// zwei Verträge über dieselbe Sache.
type Einstellungssicht struct {
	// Inhalt ist der Text, über den entschieden wird — er geht bei der
	// Übernahme zurück.
	Inhalt string `json:"inhalt"`

	Fassung  int    `json:"fassung"`
	Revision string `json:"revision"`

	Zeilen       []EinfuhrzeileDTO  `json:"zeilen"`
	Befunde      []EinfuhrbefundDTO `json:"befunde"`
	Unveraendert int                `json:"unveraendert"`
}

// EinfuhrvorschauAntwort ist sie samt Fehler — für das Fenster.
type EinfuhrvorschauAntwort struct {
	Sicht  Einstellungssicht `json:"sicht"`
	Fehler *Fehler           `json:"fehler"`
}

// Einstellungsbericht ist, was eine Übernahme bewirkt hat.
type Einstellungsbericht struct {
	Uebernommen int                `json:"uebernommen"`
	Revision    string             `json:"revision"`
	Zeilen      []EinfuhrzeileDTO  `json:"zeilen"`
	Befunde     []EinfuhrbefundDTO `json:"befunde"`
}

// EinfuhrAntwort ist er samt Fehler — für das Fenster.
type EinfuhrAntwort struct {
	Bericht Einstellungsbericht `json:"bericht"`
	Fehler  *Fehler             `json:"fehler"`
}

// EinstellungenAusfuehren gibt den Austausch als Text.
func (a *App) EinstellungenAusfuehren(
	ctx context.Context, abschnitte []string,
) EinstellungsausfuhrAntwort {
	ctx = a.imProfil(ctx)

	austausch, fehler := a.austausch.Ausfuhr(ctx, abschnitte)
	if fehler != nil {
		return EinstellungsausfuhrAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	roh, packfehler := auskunftPacken(austausch)
	if packfehler != nil {
		return EinstellungsausfuhrAntwort{Fehler: Uebersetze(ctx, fmt.Errorf(
			"den Austausch packen: %w", packfehler))}
	}

	return EinstellungsausfuhrAntwort{Inhalt: string(roh)}
}

// EinstellungenSchreiben legt den Austausch an einen benannten Ort.
//
// **Eingerückt**, wie die Auskunft über einen Menschen: eine Datei, die
// jemand mitnimmt, liest ein Mensch, bevor er sie weitergibt — und wer sehen
// soll, dass kein Geheimnis darin steht, muss sie lesen können.
func (a *App) EinstellungenSchreiben(
	ctx context.Context, abschnitte []string, ziel string,
) SchreibAntwort {
	ctx = a.imProfil(ctx)

	if ziel == "" {
		return SchreibAntwort{Fehler: Uebersetze(ctx, fmt.Errorf(
			"%w: es wurde kein Ziel genannt", domain.ErrValidation))}
	}

	antwort := a.EinstellungenAusfuehren(ctx, abschnitte)
	if antwort.Fehler != nil {
		return SchreibAntwort{Fehler: antwort.Fehler}
	}

	// **Modus 0600, auch über einer vorhandenen Datei** (SEC-20, #2133).
	if err := service.SchreibeAtomarVertraulich(ziel, []byte(antwort.Inhalt)); err != nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, fmt.Errorf(
			"%w: der Austausch ließ sich nicht schreiben: %s",
			domain.ErrValidation, err.Error()))}
	}

	return SchreibAntwort{}
}

// EinstellungenSichern fragt nach einem Ziel und schreibt dorthin.
func (a *App) EinstellungenSichern(
	ctx context.Context, abschnitte []string,
) SchreibAntwort {
	ctx = a.imProfil(ctx)

	ziel, err := a.waehleDatei(ctx, Dateiauswahl{
		Titel:     DialogtexteFuer(a.sprache).EinstellungenSichern,
		Filter:    DialogtexteFuer(a.sprache).JSONFilter,
		Muster:    "*.json",
		Vorschlag: DialogtexteFuer(a.sprache).EinstellungenVorschlag,
		Speichern: true,
	})
	if err != nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, err)}
	}

	// **Ein abgebrochener Dialog ist kein Fehler.** Wer die Wahl schließt,
	// hat nichts falsch gemacht — dieselbe Antwort wie bei jedem anderen
	// Speicherweg dieser Anwendung.
	if ziel == "" {
		return SchreibAntwort{}
	}

	return a.EinstellungenSchreiben(ctx, abschnitte, ziel)
}

// EinstellungenVorschau sagt, was ein Text ändern würde.
func (a *App) EinstellungenVorschau(
	ctx context.Context, inhalt string,
) EinfuhrvorschauAntwort {
	ctx = a.imProfil(ctx)

	sicht, fehler := einstellungssichtAus(a.austausch.Vorschau(ctx, []byte(inhalt)))
	if fehler != nil {
		return leereVorschau(Uebersetze(ctx, fehler))
	}

	sicht.Inhalt = inhalt

	return EinfuhrvorschauAntwort{Sicht: sicht}
}

// einstellungssichtAus bildet die Vorschau des Dienstes ab.
//
// **Als Funktion und nicht in jedem Aufrufer**: das Fenster und die drei
// Flächen ohne Fenster nehmen dieselbe Abbildung, und zwei liefen auseinander.
func einstellungssichtAus(vorschau service.Einfuhrvorschau, fehler error) (Einstellungssicht, error) {
	if fehler != nil {
		return leereEinstellungssicht(), fehler
	}

	return Einstellungssicht{
		Fassung:      vorschau.Fassung,
		Revision:     vorschau.Revision,
		Zeilen:       einfuhrzeilenDTO(vorschau.Zeilen),
		Befunde:      einfuhrbefundeDTO(vorschau.Befunde),
		Unveraendert: vorschau.Unveraendert,
	}, nil
}

// einstellungsberichtAus bildet das Ergebnis des Dienstes ab.
func einstellungsberichtAus(ergebnis service.Einfuhrergebnis, fehler error) (Einstellungsbericht, error) {
	if fehler != nil {
		return leererEinstellungsbericht(), fehler
	}

	return Einstellungsbericht{
		Uebernommen: ergebnis.Uebernommen,
		Revision:    ergebnis.Revision,
		Zeilen:      einfuhrzeilenDTO(ergebnis.Zeilen),
		Befunde:     einfuhrbefundeDTO(ergebnis.Befunde),
	}, nil
}

// EinstellungenEinlesen fragt nach einer Datei und zeigt, was sie änderte.
func (a *App) EinstellungenEinlesen(ctx context.Context) EinfuhrvorschauAntwort {
	ctx = a.imProfil(ctx)

	quelle, err := a.waehleDatei(ctx, Dateiauswahl{
		Titel:  DialogtexteFuer(a.sprache).EinstellungenOeffnen,
		Filter: DialogtexteFuer(a.sprache).JSONFilter,
		Muster: "*.json",
	})
	if err != nil {
		return leereVorschau(Uebersetze(ctx, err))
	}

	if quelle == "" {
		return leereVorschau(nil)
	}

	//#nosec G304 -- ein Pfad, den ein Mensch im Dialog gewählt hat
	roh, lesefehler := os.ReadFile(quelle)
	if lesefehler != nil {
		return leereVorschau(Uebersetze(ctx, fmt.Errorf(
			"%w: die Datei ließ sich nicht lesen: %s",
			domain.ErrValidation, lesefehler.Error())))
	}

	return a.EinstellungenVorschau(ctx, string(roh))
}

// EinstellungenUebernehmen schreibt die gewählten Zeilen.
func (a *App) EinstellungenUebernehmen(
	ctx context.Context, inhalt, revision string, nur []string,
) EinfuhrAntwort {
	ctx = a.imProfil(ctx)

	bericht, fehler := einstellungsberichtAus(
		a.austausch.Uebernehmen(ctx, []byte(inhalt), revision, nur))
	if fehler != nil {
		return EinfuhrAntwort{Bericht: bericht, Fehler: Uebersetze(ctx, fehler)}
	}

	return EinfuhrAntwort{Bericht: bericht}
}

// leereVorschau gibt eine Antwort mit leeren Listen.
//
// **Leer und nicht `nil`** (§3): ein `nil` durchquert die Naht als `null`, und
// das Fenster bekäme keine Liste statt einer leeren.
func leereVorschau(fehler *Fehler) EinfuhrvorschauAntwort {
	return EinfuhrvorschauAntwort{Sicht: leereEinstellungssicht(), Fehler: fehler}
}

// leereEinstellungssicht gibt eine Vorschau mit leeren Listen.
func leereEinstellungssicht() Einstellungssicht {
	return Einstellungssicht{
		Zeilen: []EinfuhrzeileDTO{}, Befunde: []EinfuhrbefundDTO{},
	}
}

// leererEinstellungsbericht gibt einen Bericht mit leeren Listen.
func leererEinstellungsbericht() Einstellungsbericht {
	return Einstellungsbericht{
		Zeilen: []EinfuhrzeileDTO{}, Befunde: []EinfuhrbefundDTO{},
	}
}

// einfuhrzeilenDTO bildet den Diff ab.
func einfuhrzeilenDTO(zeilen []service.Einfuhrzeile) []EinfuhrzeileDTO {
	abgebildet := make([]EinfuhrzeileDTO, 0, len(zeilen))
	for _, zeile := range zeilen {
		abgebildet = append(abgebildet, EinfuhrzeileDTO{
			Ebene: zeile.Ebene, Schluessel: zeile.Schluessel,
			Vorher: zeile.Vorher, Nachher: zeile.Nachher,
			Herkunft: zeile.Herkunft, Wirkung: zeile.Wirkung,
		})
	}

	return abgebildet
}

// einfuhrbefundeDTO bildet die Befunde ab.
func einfuhrbefundeDTO(befunde []domain.Einfuhrbefund) []EinfuhrbefundDTO {
	abgebildet := make([]EinfuhrbefundDTO, 0, len(befunde))
	for _, befund := range befunde {
		abgebildet = append(abgebildet, EinfuhrbefundDTO{
			Ebene: befund.Ebene, Schluessel: befund.Schluessel,
			Grund: befund.Grund,
		})
	}

	return abgebildet
}
