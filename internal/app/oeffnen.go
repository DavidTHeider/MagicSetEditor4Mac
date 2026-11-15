package app

import (
	"context"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Importeur liest Quellen in den Bestand — die Naht zu `internal/service`.
type Importeur interface {
	Set(ctx context.Context, pfad string) (*service.Ergebnis, error)
	MTGJSON(ctx context.Context, pfad, spiel string) (*service.Ergebnis, error)

	// Neu legt ein leeres Set an (#891) — der Weg, der **nicht** über eine
	// Datei führt.
	Neu(ctx context.Context, name, spiel, stil string) (int64, error)

	// Vorschau sagt, was [Importeur.Set] täte — FMT-03 (#1692).
	Vorschau(ctx context.Context, pfad string) (*service.Setvorschau, error)
}

// OeffnenAntwort ist der Ausgang eines Öffnens.
type OeffnenAntwort struct {
	SetID  string `json:"setID"`
	Karten int    `json:"karten"`

	// Befunde sind die Stellen, an denen der Import etwas **nicht lesen**
	// konnte.
	//
	// **Befunde und Erfolg schließen einander nicht aus** (§8): ein Set mit 300
	// lesbaren und zwei unlesbaren Karten wird geöffnet, und die zwei stehen
	// hier — mit ihrer Stelle, sonst sucht sie jemand in 3.000 Zeilen.
	//
	// **Nicht zu verwechseln mit `Vorbehalte(setID)`.** Das sind die Rohblöcke:
	// Zeilen, die der Import nicht *deutet* und unverändert mitführt. Der erste
	// Anlauf nannte beides „Vorbehalte", und der erste Test darüber erwartete
	// prompt das eine und bekam das andere. Ein unverstandener Block ist kein
	// Befund — er ist Bestand.
	Befunde []string `json:"befunde"`

	// SchonOffen sagt, dass diese Datei bereits im Bestand liegt.
	//
	// **Gemeldet und nicht verschwiegen**: wer zweimal dieselbe Datei öffnet,
	// bekommt dasselbe Set und die Auskunft, dass es dasselbe ist. Ein zweites
	// anzulegen hieße, dieselben Karten doppelt zu führen — und wer im einen
	// ändert, fände die Änderung im anderen nicht.
	SchonOffen bool `json:"schonOffen"`

	// Angelegt sind die Sets, die dabei entstanden sind.
	//
	// **Auch bei einem Abbruch** (#155):
	// jedes Set einer AllSets-Datei liegt in einer eigenen Transaktion. Was vor
	// dem Fehler durchkam, **steht im Bestand** — und wer es nicht erfährt,
	// findet die Sets später in der Liste und hält sie für vollständig.
	Angelegt []string `json:"angelegt"`

	// Unvollstaendig sagt, dass die Quelle **nicht zu Ende** gelesen wurde.
	//
	// Ein eigener Wert und nicht „Fehler und Angelegt zusammen gelesen": die
	// Oberfläche baut daraus einen Satz, und ein Satz aus zwei Bedingungen ist
	// einer, den irgendwann jemand falsch zusammensetzt.
	Unvollstaendig bool `json:"unvollstaendig"`

	// Abgebrochen sagt, dass niemand eine Datei gewählt hat.
	//
	// Ein eigener Wert und kein leerer Pfad: „nichts gewählt" ist eine Aussage,
	// und eine Oberfläche, die sie von einem Fehler nicht unterscheiden kann,
	// zeigt beim Abbrechen eine Fehlermeldung.
	Abgebrochen bool `json:"abgebrochen"`

	Fehler *Fehler `json:"fehler"`
}

// SetOeffnen liest ein Setpaket in den Bestand.
//
// **Eine Panik kommt als Fehler zurück** — F02 (#2341): über die Bindung fing
// sie bis hierher Wails und gab dem Fenster einen Satz mit Paketpfad, über
// den Finder ([App.DateiOeffnen]) fing sie niemand. Beide Wege führen hier
// durch und bekommen damit denselben Fehler.
func (a *App) SetOeffnen(ctx context.Context, pfad string) (antwort OeffnenAntwort) {
	ctx = a.imProfil(ctx)

	defer func() {
		if wert := recover(); wert != nil {
			antwort = OeffnenAntwort{Befunde: []string{}, Angelegt: []string{}, Fehler: panikFehler(ctx, wert)}
		}
	}()

	return a.setOeffnenIm(ctx, pfad)
}

// setOeffnenIm ist derselbe Weg **in einem Kontext, der schon einer ist**.
//
// # Warum es sie gibt
//
// **Weil [App.imProfil] nicht nur das Profil anhängt, sondern eine
// Vorgangskennung vergibt.** Ein zweiter Ruf darauf im selben Vorgang gäbe
// eine **zweite** Kennung — und dann nennt die Meldung, die ein Mensch sieht,
// eine andere als das Protokoll, in dem der Grund steht (§8).
//
// [App.UmschlagUebernehmen] ist der erste Ruf von innen (N4, #1457): es hat
// seinen Kontext schon ausgestattet und reicht ihn hier hinein.
func (a *App) setOeffnenIm(ctx context.Context, pfad string) OeffnenAntwort {
	leer := OeffnenAntwort{Befunde: []string{}, Angelegt: []string{}}

	if pfad == "" {
		leer.Abgebrochen = true

		return leer
	}

	// **Die Regel „eine Quelle, ein Set" steht im Importeur** (#886). In einer
	// Fassade wäre sie eine Regel je Fassade: die Kommandozeile ginge an ihr
	// vorbei und legte dasselbe Set ein zweites Mal an.
	ergebnis, err := a.importeur.Set(ctx, pfad)

	// **Ein Ergebnis samt Fehler geht samt Fehler hinaus** — Befund 1 des
	// Nachlaufs über Phase 4.4
	// (#1291).
	// `Importeur.Set` hängt einen Schließfehler des Archivs an ein
	// **geglücktes** Einfahren an (`import.go`, das `defer` mit
	// `errors.Join`); das Set steht dann im Bestand, und wer nur den Fehler
	// zeigte, ließe es zurück, ohne dass jemand seine Kennung erführe
	// (§7 Regel 3).
	if ergebnis == nil {
		leer.Fehler = Uebersetze(ctx, err)

		return leer
	}

	if err != nil {
		leer.Fehler = Uebersetze(ctx, err)
	}

	for _, befund := range ergebnis.Befunde {
		leer.Befunde = append(leer.Befunde, befund.String())
	}

	if ergebnis.SetID != 0 {
		a.arbeitVermerken(ctx, domain.ArbeitsartSet, strconv.FormatInt(ergebnis.SetID, 10), domain.VorgangEingefuehrt)
	}

	for _, kennung := range ergebnis.Angelegt {
		leer.Angelegt = append(leer.Angelegt, strconv.FormatInt(kennung, 10))
	}

	leer.SetID = strconv.FormatInt(ergebnis.SetID, 10)
	leer.Karten = ergebnis.Karten
	leer.SchonOffen = ergebnis.SchonOffen

	return leer
}

// MitImporteur setzt die Naht zum Import.
func (a *App) MitImporteur(importeur Importeur) { a.importeur = importeur }

// MitDateiwahl setzt die Naht zum Dateidialog.
func (a *App) MitDateiwahl(waehlen func(context.Context, Dateiauswahl) (string, error)) {
	a.dateiwahl = waehlen
}
