package app

// **Ein Mensch gibt den Neuanfang frei** — G12
// (#2282).
//
// # Der Befund
//
// **Die ganze Anwendung ließ sich nirgends zurücksetzen.** Die Wache verlangt
// einen Nachweis mit dem Zweck [domain.RuecksetzungsZweck]
// (`service/ruecksetzungswache.go`), und das Fenster fragte nach seiner
// Kennung — ausgestellt hat ihn keine Bindung. Dieselbe Lage wie bei der
// KI-Freigabe vor #1539 (`kifreigabe.go`): der Mechanismus war fertig, nur
// nicht angeschlossen.
//
// # Warum es sie über MCP, Kommandozeile und Dienst nicht gibt
//
// **Weil `Nachweise.Ausstellen` jeden Modus außer `fenster` abweist**
// ([0078](../../docs/entscheidungen/0078-der-wirksame-akteur.md)). Wer an der
// Kommandozeile neu anfangen will, holt sich die Freigabe am Fenster und
// zeigt ihre Kennung dort vor (`mse zuruecksetzen anwendung <name> <nachweis>`).

import (
	"context"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// RuecksetzungsFreigabefrist ist, wie lange eine Freigabe für den Neuanfang
// gilt.
//
// **Fünf Minuten**, dieselbe Zahl wie [KIFreigabefrist] und aus demselben
// Grund: sie deckt den Weg vom Knopf bis zur Ausführung, nicht den
// Nachmittag. Wer länger braucht, gibt sie neu.
const RuecksetzungsFreigabefrist = 5 * time.Minute

// RuecksetzungFreigabeErteilen stellt den Nachweis für **einen** Neuanfang
// aus.
//
// **Ohne Revision**: die Wache bindet den Nachweis an Zweck und Art
// (`Ruecksetzung.erlaubt`), und an die gesehene Vorschau bindet die Revision
// der Vorschau selbst — sie geht beim Ausführen eigens mit.
func (a *App) RuecksetzungFreigabeErteilen(ctx context.Context) NachweisAntwort {
	ctx = a.imProfil(ctx)

	if a.autorisierung == nil {
		return NachweisAntwort{Fehler: Uebersetze(ctx, fehlenderBestand())}
	}

	nachweis, fehler := a.autorisierung.Ausstellen(ctx, service.Nachweisantrag{
		// **Derselbe Zweck und dasselbe Ziel, die die Wache einlöst.** Ein
		// anderer wäre ein Nachweis, der zu nichts passt.
		Zweck: domain.RuecksetzungsZweck,
		Ziel:  string(domain.RuecksetzungAnwendung),
		Frist: RuecksetzungsFreigabefrist,
	})
	if fehler != nil {
		return NachweisAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	return NachweisAntwort{
		Kennung:  nachweis.Kennung,
		Laeuftab: nachweis.Ablauf.UTC().Format(time.RFC3339),
	}
}

// RuecksetzungFreigabeErteilen ist die Fensterbindung für die Freigabe.
func (b *Bindungen) RuecksetzungFreigabeErteilen(ctx context.Context) NachweisAntwort {
	return b.app(ctx).RuecksetzungFreigabeErteilen(b.ruf(ctx))
}
