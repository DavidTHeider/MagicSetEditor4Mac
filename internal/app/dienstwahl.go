// **Wer wen fragen darf** — Werkstück 2 von Phase 4.8
// (#1321),
// entschieden in [0063](../../docs/entscheidungen/0063-was-ein-dienst-hier-ist.md).

package app

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Dienstwahl sagt, ob ein fremder Dienst gefragt werden darf.
//
// # Warum eine Schnittstelle und keine Abfrage im Bestand
//
// **Weil sie dort steht, wo sie gebraucht wird** (§2). Wer einen Klienten
// baut, muss den Schalter kennen; den **Bestand** muss er nicht kennen. Ein
// Lauf, der die Sperre prüft, braucht damit keine Datenbank — und einer, der
// sie umgeht, fällt auf.
type Dienstwahl interface {
	IstAn(ctx context.Context, dienst domain.Dienst) (bool, error)
}

// Einstellungswahl liest die Schalter aus den Einstellungen.
type Einstellungswahl struct {
	leser Einstellungsleser

	// riegel fragt, ob der Akteur dieses Aufrufs nach draußen darf —
	// Werkstück O1 von Phase 5.2 (#1425). `nil` heißt: niemand fragt; die
	// Wurzeln setzen ihn, und `internal/app/leseraussen_test.go` belegt es.
	riegel service.Pruefer
}

// MitRiegel setzt den Prüfer, den jede Anfrage nach draußen fragt.
func (w *Einstellungswahl) MitRiegel(riegel service.Pruefer) *Einstellungswahl {
	w.riegel = riegel

	return w
}

// NachAussen sagt, ob der Akteur dieses Aufrufs Daten nach draußen schicken
// darf — `externSenden` verlangt `mitarbeiter`, und ein Token reicht nie
// dorthin ([0078](../../docs/entscheidungen/0078-der-wirksame-akteur.md)).
func (w *Einstellungswahl) NachAussen(ctx context.Context) error {
	if w.riegel == nil {
		return nil
	}

	return w.riegel.Pruefe(ctx, domain.WirkungExternSenden)
}

// Riegeltraeger ist eine Dienstwahl, die auch weiß, ob der Akteur nach
// draußen darf. **Eine zweite Schnittstelle und keine zweite Methode an
// [Dienstwahl]**: die Frage „ist der Dienst an?" gehört zu den
// Einstellungen, die Frage „darf dieser Mensch senden?" zum Akteur — und
// eine Wahl in einem Test, die nur die erste beantwortet, soll das dürfen.
type Riegeltraeger interface {
	NachAussen(ctx context.Context) error
}

// Einstellungsleser ist der Teil des Einstellungsspeichers, den die Wahl
// braucht.
type Einstellungsleser interface {
	Alle(ctx context.Context) (map[string]string, error)
}

// Netzvorgaben sagt, wie dieses Programm nach außen geht (N5, #1571).
//
// **Hier und nicht in einer zweiten Naht**: diese Wahl liest die
// Einstellungen ohnehin, und der Proxy steht in derselben Tabelle wie die
// Dienstschalter. Eine eigene Naht daneben wäre ein zweiter Leser über
// derselben Quelle.
func (w *Einstellungswahl) Netzvorgaben(ctx context.Context) Netzvorgaben {
	werte, fehler := w.leser.Alle(ctx)
	if fehler != nil {
		return Netzvorgaben{Fristen: netz.UeblicheFristen()}
	}

	return NetzvorgabenAus(werte)
}

// NeueEinstellungswahl baut die Wahl über einem Einstellungsspeicher.
func NeueEinstellungswahl(leser Einstellungsleser) *Einstellungswahl {
	return &Einstellungswahl{leser: leser}
}

// IstAn sagt, ob dieser Dienst gefragt werden darf.
//
// **Ohne Eintrag gilt die Vorgabe** — und die ist für die drei neuen Dienste
// „aus" (0063, Verworfen: „Ein Dienst, dessen Bedingungen jemand nicht gelesen
// hat, soll nicht in seinem Namen fragen").
func (w *Einstellungswahl) IstAn(ctx context.Context, dienst domain.Dienst) (
	bool, error,
) {
	alle, err := w.leser.Alle(ctx)
	if err != nil {
		return false, fmt.Errorf("die Einstellungen lesen: %w", err)
	}

	schluessel := domain.DienstEinstellung(dienst)

	wert, gesetzt := alle[schluessel]
	if !gesetzt || wert == "" {
		wert = domain.EinstellungsVorgabe(schluessel)
	}

	return wert == domain.DienstAn, nil
}

// Wert gibt den Wert einer Einstellung — oder ihre Vorgabe.
//
// **Dieselbe Stelle wie [Einstellungswahl.IstAn]** (T11, #1476): beide lesen
// dieselbe Tabelle, und eine zweite Naht daneben wäre ein zweiter Leser über
// derselben Zeile. Was sie unterscheidet, ist, dass jene einen Schalter
// deutet und diese einen Wert gibt.
func (w *Einstellungswahl) Wert(ctx context.Context, schluessel string) (
	string, error,
) {
	alle, err := w.leser.Alle(ctx)
	if err != nil {
		return "", fmt.Errorf("die Einstellungen lesen: %w", err)
	}

	wert, gesetzt := alle[schluessel]
	if !gesetzt || wert == "" {
		return domain.EinstellungsVorgabe(schluessel), nil
	}

	return wert, nil
}

// ImmerAn ist die Wahl für Läufe, die keinen Bestand haben.
//
// **Sie ist keine Bequemlichkeit, sondern eine benannte Stelle.** Wer sie
// benutzt, sagt damit: hier gibt es keine Einstellung, die gelesen werden
// könnte. Ein `nil` an derselben Stelle sähe aus wie ein Versehen.
type ImmerAn struct{}

// IstAn sagt immer ja.
func (ImmerAn) IstAn(context.Context, domain.Dienst) (bool, error) {
	return true, nil
}

// pruefeDienst gibt einen Fehler, wenn der Dienst nicht gefragt werden darf.
//
// **Ein Lesefehler heißt nicht »an«.** Wer den Schalter nicht lesen kann, weiß
// nicht, ob er fragen darf — und fragt dann nicht. Die Umkehrung wäre die
// schlimmere: ein Dienst, der nach einem Lesefehler doch hinausgeht.
func pruefeDienst(
	ctx context.Context, wahl Dienstwahl, dienst domain.Dienst,
) error {
	an, err := wahl.IstAn(ctx, dienst)
	if err != nil {
		return err
	}

	if !an {
		return domain.DienstIstAus(dienst)
	}

	return nil
}

// MitDienstwahl setzt die Wahl an der Anwendung.
//
// **Gesetzt wird sie beim Start**, wo der Bestand offen ist. Ein Weg, der sie
// braucht, nimmt sie von hier statt sie selbst zu bauen — sonst läse jeder
// Weg seinen eigenen Schalter, und einer läse irgendwann einen anderen.
// pruefeNachAussen fragt beides: ist der Dienst an, und darf der Akteur
// senden. **Vor jeder Anfrage**, denn `netz.Klient` ruft es so.
func pruefeNachAussen(
	ctx context.Context, wahl Dienstwahl, dienst domain.Dienst,
) error {
	if err := pruefeDienst(ctx, wahl, dienst); err != nil {
		return err
	}

	if traeger, hatRiegel := wahl.(Riegeltraeger); hatRiegel {
		return traeger.NachAussen(ctx)
	}

	return nil
}

func (a *App) MitDienstwahl(wahl Dienstwahl) { a.dienstwahl = wahl }

// wahl gibt die Dienstwahl der Anwendung.
//
// # Warum ohne sie »an« gilt und nicht »aus«
//
// **Weil eine Anwendung ohne Bestand keinen Schalter hat**, den sie lesen
// könnte — und ein `nil` heißt hier nicht „abgeschaltet", sondern „danach
// wurde nie gefragt". Wer daraus „aus" machte, ließe jeden Lauf ohne
// Datenbank ins Leere greifen.
//
// **Was das kostet, hält ein Wächter**: `TestDerStartSetztDieDienstwahl`
// fährt den echten Startweg und sieht nach, dass sie gesetzt ist. Ohne ihn
// wäre ein vergessener Aufruf im Start eine stille Abschaltung der Sperre —
// und die sähe von außen aus wie eine funktionierende.
func (a *App) wahl() Dienstwahl {
	if a.dienstwahl == nil {
		return ImmerAn{}
	}

	return a.dienstwahl
}

// HatDienstwahl sagt, ob eine Wahl gesetzt ist.
//
// **Nur für den Wächter über den Startweg.** Ohne ihn wäre ein vergessener
// [App.MitDienstwahl] im Start eine stille Abschaltung der Sperre, und die
// sähe von außen aus wie eine funktionierende (§10).
func (a *App) HatDienstwahl() bool { return a.dienstwahl != nil }

// Fragewahl sagt, welcher Dienst eine Frage beantwortet.
//
// **Neben [Dienstwahl] und nicht statt ihrer** (Vermerk 0063, Entscheidung
// 1): „darf gefragt werden" und „beantwortet diese Frage" sind zwei
// Aussagen, und wer sie zusammenlegte, könnte die zweite nicht stellen.
type Fragewahl interface {
	FuerFrage(ctx context.Context, frage domain.Frage) (domain.Dienst, error)
}

// FuerFrage gibt den Dienst, der für eine Frage gewählt ist.
//
// **Ohne Eintrag gilt die Vorgabe** — und die ist der erste Dienst, der die
// Frage beantworten kann.
func (w *Einstellungswahl) FuerFrage(ctx context.Context, frage domain.Frage) (
	domain.Dienst, error,
) {
	alle, err := w.leser.Alle(ctx)
	if err != nil {
		return "", fmt.Errorf("die Einstellungen lesen: %w", err)
	}

	schluessel := domain.FrageEinstellung(frage)

	wert, gesetzt := alle[schluessel]
	if !gesetzt || wert == "" {
		wert = domain.EinstellungsVorgabe(schluessel)
	}

	return domain.Dienst(wert), nil
}

// FuerFrage gibt bei [ImmerAn] den ersten Dienst, der die Frage beantwortet.
//
// **Sie ist die Wahl für Läufe ohne Bestand** — dort gibt es keine
// Einstellung, die gelesen werden könnte.
func (ImmerAn) FuerFrage(_ context.Context, frage domain.Frage) (
	domain.Dienst, error,
) {
	dienste := domain.DiensteFuer(frage)
	if len(dienste) == 0 {
		return "", fmt.Errorf(
			"%w: für %q gibt es keinen Dienst", domain.ErrValidation, frage)
	}

	return dienste[0], nil
}
