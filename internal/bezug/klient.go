// Der HTTP-Klient dieses Pakets — Fristen, Selbstnennung, Grenzen der Gegenstelle.
//
// **Eine eigene Datei, weil sie eine eigene Frage beantwortet** (§2): nicht
// *was* bezogen wird, sondern *unter welchen Bedingungen* dieses Projekt ins
// Netz geht. `bezug.go` ging über die Zeilengrenze, als die Leerlauffrist
// dazukam (#916).
//
// **Damit stehen die Zusagen aus §16 an einem Ort**: jede Anfrage hat eine
// Frist und einen Abbruch, die Anwendung nennt sich, und eine Grenze der
// Gegenstelle wird gelesen statt durchbrochen. Der zweite Netzverbraucher, den
// Phase V bringt, findet sie hier.

package bezug

import (
	"errors"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// kopffrist ist die Zeit, die eine Gegenstelle für ihren Antwortkopf hat.
//
// # Warum nicht `http.Client.Timeout`
//
// Der deckt **den ganzen Vorgang** ab, Rumpf eingeschlossen. Ein Paket darf
// zwei Gigabyte wiegen; über eine langsame Leitung ist das legitim und dauert.
// Eine Gesamtfrist, die dafür reicht, fängt keinen hängenden Server mehr —
// und eine, die ihn fängt, bricht den langsamen Bezug ab.
//
// Gewartet wird deshalb auf den **Kopf**: eine Gegenstelle, die nach dreißig
// Sekunden noch keine Statuszeile geschickt hat, antwortet nicht.
//
// # Was sie allein nicht abdeckt
//
// **Einen Rumpf, der nach dem Kopf stehenbleibt.** Dafür steht die
// [leerlauffrist] daneben.
//
// **Eine Veränderliche**, damit ein Test sie herunterdrehen kann — im Betrieb
// ändert sie sich nie.
var kopffrist = 30 * time.Second

// verbindungsfrist ist die Zeit für Verbindungsaufbau und TLS.
var verbindungsfrist = 15 * time.Second

// leerlauffrist ist die Zeit, die der Rumpf ohne Fortschritt haben darf.
//
// # Warum es sie neben der Kopffrist braucht
//
// `ResponseHeaderTimeout` misst bis zur Statuszeile. Wer sie schickt und dann
// **schweigt**, war danach unbeobachtet — und genau das war der Audit-Befund
// E-37: „ein hängender Server hält den Paketbezug beliebig lange fest".
//
// #899 hat die
// halbe Klasse geschlossen und es im Kommentar gesagt. Der Zweitblick über
// Phase R hat auf der anderen Hälfte bestanden: **eine benannte Lücke ist eine
// Lücke** (#916).
//
// # Warum Stillstand und nicht Dauer
//
// Ein Paket darf zwei Gigabyte wiegen; über eine langsame Leitung ist das
// legitim und dauert. Abgebrochen wird deshalb, **wer nichts mehr schickt** —
// nicht, wer langsam schickt. Die Frist läuft nach jedem gelesenen Stück neu.
var leerlauffrist = 60 * time.Second

// klient gibt den HTTP-Klienten dieses Pakets.
//
// **Aus `internal/netz`, seit dem 2026-08-23**
// (#1002).
// Ein eigener Transport hier wäre neben dem zweiten Verbraucher eine Kopie —
// und was daraus wird, steht in `msepaket/paket.go`.
//
// **Bei jedem Aufruf neu.** Er hält keinen Zustand, den ein zweiter Bezug
// bräuchte, und die Fristen sind veränderlich: ein einmal gebauter Klient
// trüge die Frist von damals, und ein Test könnte sie nicht mehr verschieben.
//
// **Und nur ins Öffentliche** — SEC-P4
// (#2000).
// Die Adresse nennt, wer `POST /pakete` ruft, und das darf ein Token
// `schreiben`. Ohne den Bereich bezog die Anwendung für ihn von einem Dienst
// auf der Schleife oder im eigenen Netz, den er selbst nicht erreicht. Ein
// Lauf gegen `httptest` öffnet die Schleife über [netz.SchleifeUmgebung].
func klient() *netz.Klient {
	return netz.NeuerKlientNach(netz.Fristen{
		Verbindung: verbindungsfrist,
		Kopf:       kopffrist,
		Stillstand: leerlauffrist,
	}, nil, netz.ZielOeffentlich)
}

// kennung sagt, womit sich dieser Bezug vorstellt.
//
// **`User-Agent` und `Accept`** (§16). Für den Paketbezug ist die
// Selbstnennung Höflichkeit; für jeden weiteren Dienst ist sie Pflicht — und
// der Vorsatz gehört an die **eine** Stelle, an der dieses Projekt ins Netz
// geht, nicht an jede einzelne.
func kennung(kenner string) netz.Kenner {
	return netz.Kenner{Name: kenner, Annahme: "*/*"}
}

// alsZielbefund übersetzt ein gesperrtes Ziel in einen Feldbefund — oder
// gibt `nil` (#2000).
//
// **Am Feld, das die Adresse genannt hat** (§8), und **ohne die Kette**: sie
// nennt die aufgelöste Adresse eines Namens, und wer Namen durchprobiert,
// läse daraus das eigene Netz ab. Sie reist als Grund ins Protokoll.
func alsZielbefund(fehler error, feld, adresse string) error {
	if !errors.Is(fehler, netz.ErrZielGesperrt) {
		return nil
	}

	return domain.NewFieldError(feld, domain.SchluesselValidierung,
		"Von %s wird nicht bezogen: die Adresse führt auf diesen Rechner, ins "+
			"eigene Netz oder in einen gesperrten Bereich.", adresse).MitGrund(fehler)
}

// alsDrosselbefund übersetzt ein 429 in einen Feldbefund — oder gibt `nil`.
//
// # Warum die Übersetzung hier steht und nicht in `internal/netz`
//
// Weil erst dieses Paket weiß, an **welchem Feld** die Meldung erscheinen soll.
// `netz` kennt keine Formulare; es meldet [netz.ErrGedrosselt], und wer ein
// Feld hat, hängt sie dort auf (#917).
func alsDrosselbefund(fehler error, feld, adresse string) error {
	if !errors.Is(fehler, netz.ErrGedrosselt) {
		return nil
	}

	return domain.NewFieldError(feld, domain.SchluesselValidierung,
		"%s meldet zu viele Anfragen. Bitte später noch einmal versuchen.", adresse)
}
