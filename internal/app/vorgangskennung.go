// **Die Vorgangskennung an der Fenstergrenze** — Werkstück O9 von Phase 5.2
// (#1433).
//
// # Vier Grenzen, eine Kennung
//
// Ein Ruf kommt über eine von vier Flächen herein. Jede von ihnen würfelt beim
// Eintritt **eine** Kennung und hängt sie an den Kontext; von da an trägt sie
// jede Protokollzeile dieses Rufs und seine Antwort. Wo die Grenze liegt:
//
// | Fläche | Stelle |
// | --- | --- |
// | Fenster | [App.imProfil] — die erste Zeile jeder Bindung, und ein Wächter hält das |
// | Kommandozeile | `cmd/mse/bestandsgriff.go` |
// | MCP | `cmd/mse-mcp/server.go`, in der Ausstattung je Ruf |
// | Dienst | `cmd/mse-rest/dienst.go`, in `imProfil` je Anfrage |
//
// **Und nur dort.** Eine zweite Stelle, die eine zweite würfelte, gäbe
// demselben Vorgang zwei Namen — genau das, was das Issue mit „keine drei
// bedeutungsgleichen Pflicht-IDs" verbietet.

package app

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// MitNeuemVorgang würfelt eine Kennung und hängt sie an — **es sei denn, es
// steht schon eine da**.
//
// **Warum die Rückfrage**: die Flächen ohne Fenster bauen ihren Kontext in
// Schichten (Profil, Modus, Token), und `internal/app` wird dabei mehrfach
// gerufen. Eine zweite Kennung mitten im Vorgang wäre schlimmer als keine —
// die Protokollzeilen davor und danach fänden sich nicht mehr.
func MitNeuemVorgang(ctx context.Context) context.Context {
	if service.VorgangAus(ctx) != "" {
		return ctx
	}

	return service.MitVorgang(ctx, domain.NeueVorgangskennung())
}

// MitVorgang trägt eine **gegebene** Kennung ein — für die Wurzeln ohne
// Fenster, die `internal/service` nicht kennen dürfen (§2), und für Tests, die
// eine bekannte Kennung brauchen.
func MitVorgang(ctx context.Context, kennung domain.Vorgangskennung) context.Context {
	return service.MitVorgang(ctx, kennung)
}

// VorgangAus liest die Kennung — leer, wo keine Grenze war.
func VorgangAus(ctx context.Context) domain.Vorgangskennung {
	return service.VorgangAus(ctx)
}

// Uebersetze bringt einen Fehler in die Form, die das Frontend kennt, und
// **schreibt die Kennung dieses Vorgangs hinein**.
//
// # Warum hier und nicht in [uebersetze]
//
// Weil die Kennung nichts mit der Einordnung des Fehlers zu tun hat: die eine
// Frage ist „was ist das für ein Fehler", die andere „in welchem Ruf ist er
// entstanden". Getrennt bleibt die Einordnung ohne Kontext prüfbar.
//
// # Was der Aufrufer davon hat
//
// Denselben Griff, den er im Protokoll findet. Der Satz, den er sieht, ist
// für Menschen (§8); der **Grund** steht unter derselben Kennung in
// `mse.log` — das ist die „korrelierbare Diagnose" aus dem Issue, und sie ist
// der einzige Weg dorthin, weil der Grund die Oberfläche absichtlich nie
// erreicht.
func Uebersetze(ctx context.Context, fehler error) *Fehler {
	uebersetzt := uebersetze(ctx, fehler)
	if uebersetzt == nil {
		return nil
	}

	uebersetzt.Vorgang = string(service.VorgangAus(ctx))

	return mitGriffInsProtokoll(uebersetzt)
}

// mitGriffInsProtokoll gibt dem Rückfallsatz die Kennung, auf die er verweist
// — #1601.
//
// # Warum hier und nicht in [uebersetze]
//
// **Weil erst hier die Kennung dasteht.** Die Einordnung weiß, dass sie den
// Fehler nicht kennt; sie weiß nicht, in welchem Ruf er entstand.
//
// # Warum nicht jeder Satz sie bekommt
//
// **Weil nur dieser eine ins Leere zeigt.** „Was genau, steht im Protokoll"
// ist ohne den Griff ein Hinweis auf eine Datei und nicht auf einen Eintrag.
// Ein Satz, der selbst sagt, was zu tun ist, braucht ihn nicht — und trüge
// dann eine Hexzahl, die niemanden angeht.
//
// **Mit einer Ausnahme: [CodeNetz]** (D02, #2233). Sein Satz nennt einen
// Schritt, aber der Grund — welche Adresse, welcher Proxy, was der Transport
// sagte — steht absichtlich nur im Protokoll. Wer nach dem Rat noch fragt,
// findet ihn dort, und der Satz nennt die Kennung: die Seiten machen aus
// einer Ursache sofort Text, und „Diagnose öffnen" steht nur an drei
// Stellen. Der Schlüssel bleibt der eigene.
//
// **Am Code und nicht am Schlüssel festgemacht**: [CodeAbgebrochen] trägt
// denselben Schlüssel, ist aber kein unbekannter Fehler — und die Oberfläche
// zeigt ihn gar nicht.
func mitGriffInsProtokoll(befund *Fehler) *Fehler {
	if befund.Vorgang == "" {
		return befund
	}

	switch befund.Code {
	case CodeUnbekannt:
		befund.Schluessel = domain.SchluesselUnbekanntMitVorgang
		befund.Werte = map[string]string{"vorgang": befund.Vorgang}
	case CodeNetz:
		befund.Werte["vorgang"] = befund.Vorgang
	}

	return befund
}

// Nutzerfehler ist die **zweite** Tür — für einen Fehler, dessen ganze
// Erklärung in seinem Satz steht.
//
// # Warum es sie gibt, und warum sie nicht bequemer ist
//
// Eine Vorgangskennung ist ein **Griff ins Protokoll**: sie verbindet den
// Satz, den jemand sieht, mit den Zeilen, die den Grund nennen. Eine
// Eingabeprüfung — „die Kennung %q ist keine Zahl", „diese Naht ist nicht
// eingerichtet" — schreibt **keine** solche Zeile: sie hat keinen verdeckten
// Grund und keinen Unterbau, aus dem etwas käme. Eine Kennung daran wäre ein
// Griff ins Leere.
//
// Deshalb steht sie hier und nicht als Bequemlichkeit für Stellen, die keinen
// Kontext zur Hand haben: sie ist die Aussage *„hier gibt es nichts zu
// korrelieren"*, und wer sie für einen Fehler aus dem Unterbau nimmt, nimmt
// dem Aufrufer den einzigen Weg zum Grund.
//
// **Ein Wächter hält die eine Hälfte** (`internal/werkzeuge/vorgangskennung_test.go`):
// sie darf nur dort stehen, wo **gar kein** `context.Context` in Reichweite
// ist. Damit ist die Gefahr abgestellt, um die es geht — die kürzere Tür zu
// nehmen, obwohl die längere offenstand.
//
// **Was er nicht hält**: ob eine kontextlose Funktion wirklich nur Eingaben
// prüft. Das ist eine Frage ans Review (§10), und der Kopf des Wächters nennt
// die eine Stelle, an der sie sich heute stellt.
func Nutzerfehler(fehler error) *Fehler {
	return uebersetze(context.Background(), fehler)
}
