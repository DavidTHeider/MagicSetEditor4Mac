package app

// **Ein Knopf richtet alles ein, einer nimmt es zurück** — G08
// (#2272).
//
// Bis dahin richtete *Werkzeuge ohne Fenster* fünf Ziele **einzeln** ein,
// jedes mit eigenem Knopf (OPS-09). Hier geht ein Ruf über alle: was es auf
// diesem Rechner gibt und noch fehlt (`offen`, `veraltet`), wird
// eingerichtet; was eingerichtet ist, wird zurückgenommen.
//
// # Was dabei nicht geschieht
//
// **Ein fremder Eintrag wird nie überschrieben** — er wird benannt
// (`fremd`), und der Mensch entscheidet. **Ein Teilfehler nimmt die übrigen
// nicht mit**: jedes Ziel sagt, was ihm geschah, und was gelang, bleibt —
// der Fehler der Antwort nennt das erste, das scheiterte. **Ein Abbruch
// fängt kein weiteres Ziel an**; was schon geschrieben ist, bleibt
// geschrieben (§7, Regel 3).
//
// **Nur am Fenster**, aus demselben Grund wie die Einzelziele: es schreibt in
// Dateien, die dem Menschen vor dem Fenster gehören (FAEHIGKEITEN).

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Die Ausgänge eines Ziels beim Einrichten oder Zurücknehmen aller.
const (
	AusgangEingerichtet = "eingerichtet"
	AusgangWiderrufen   = "widerrufen"
	AusgangFremd        = "fremd"
	AusgangUnveraendert = "unveraendert"
	AusgangGescheitert  = "gescheitert"
	AusgangAbgebrochen  = "abgebrochen"
)

// SchnittstellenausgangDTO ist, was einem Ziel beim Einrichten aller geschah.
type SchnittstellenausgangDTO struct {
	Name    string  `json:"name"`
	Ausgang string  `json:"ausgang"`
	Fehler  *Fehler `json:"fehler"`
}

// SchnittstellenAlleAntwort ist der Stand aller Ziele danach samt ihrer
// Ausgänge.
type SchnittstellenAlleAntwort struct {
	Schnittstellen []SchnittstelleDTO         `json:"schnittstellen"`
	Ausgaenge      []SchnittstellenausgangDTO `json:"ausgaenge"`
	Fehler         *Fehler                    `json:"fehler"`
}

// SchnittstellenAlleEinrichten richtet ein, was es gibt und was fehlt.
func (a *App) SchnittstellenAlleEinrichten(ctx context.Context) SchnittstellenAlleAntwort {
	ctx = a.imProfil(ctx)

	return a.anAllenSchnittstellen(ctx, func(stand service.Schnittstellenstand) bool {
		return stand == service.SchnittstelleOffen || stand == service.SchnittstelleVeraltet
	}, service.SchnittstelleEinrichten, AusgangEingerichtet)
}

// SchnittstellenAlleWiderrufen nimmt zurück, was eingerichtet ist.
func (a *App) SchnittstellenAlleWiderrufen(ctx context.Context) SchnittstellenAlleAntwort {
	ctx = a.imProfil(ctx)

	return a.anAllenSchnittstellen(ctx, func(stand service.Schnittstellenstand) bool {
		return stand == service.SchnittstelleEingerichtet || stand == service.SchnittstelleVeraltet
	}, service.SchnittstelleWiderrufen, AusgangWiderrufen)
}

func (a *App) anAllenSchnittstellen(
	ctx context.Context, betrifft func(service.Schnittstellenstand) bool,
	handlung func(context.Context, service.Schnittstellenziel) (service.Schnittstellenausgang, error),
	gelungen string,
) SchnittstellenAlleAntwort {
	ziele, fehler := a.schnittstellenziele(ctx)
	if fehler != nil {
		return SchnittstellenAlleAntwort{
			Schnittstellen: []SchnittstelleDTO{}, Ausgaenge: []SchnittstellenausgangDTO{},
			Fehler: Uebersetze(ctx, fehler),
		}
	}

	antwort := SchnittstellenAlleAntwort{Ausgaenge: make([]SchnittstellenausgangDTO, 0, len(ziele))}

	for _, ziel := range ziele {
		ausgang := SchnittstellenausgangDTO{Name: ziel.Name, Ausgang: AusgangUnveraendert}
		stand := service.SchnittstelleLesen(ziel)

		switch {
		case stand == service.SchnittstelleFremd:
			ausgang.Ausgang = AusgangFremd

		case !betrifft(stand):

		case ctx.Err() != nil:
			ausgang.Ausgang = AusgangAbgebrochen

		default:
			if _, gescheitert := handlung(ctx, ziel); gescheitert != nil {
				ausgang.Ausgang, ausgang.Fehler = AusgangGescheitert, Uebersetze(ctx, gescheitert)

				if antwort.Fehler == nil {
					antwort.Fehler = ausgang.Fehler
				}
			} else {
				ausgang.Ausgang = gelungen
			}
		}

		antwort.Ausgaenge = append(antwort.Ausgaenge, ausgang)
	}

	antwort.Schnittstellen = a.Schnittstellen(ctx).Schnittstellen

	return antwort
}
