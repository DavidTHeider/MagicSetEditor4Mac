package rest

// **Zugriffstoken statt Prozessmarke** — Werkstück 10 von Phase 4.12
// (#1351),
// Punkt 13.
//
// # Was ein Zugriffstoken kann, was die Marke nicht kann
//
// Die Marke (`marke.go`) ist **eine** je Prozess: sie gilt, solange der Dienst
// läuft, sie gehört niemandem, und sie darf alles. Für einen Dienst, der auf
// Zuruf startet und mit dem Aufrufer endet, ist das richtig.
//
// **Ein Zugriffstoken ist das Gegenteil in jedem Punkt**: es gehört einem
// **Profil**, es hat eine **Erlaubnis** und einen **Ablauf**, und es lässt
// sich zurückziehen, ohne dass jemand den Dienst anfassen muss.
//
// # Wie sie nebeneinander stehen
//
// **In derselben Kopfzeile, unterschieden an der Marke davor.** Ein Wert, der
// mit `mse_` beginnt, ist ein Zugriffstoken; alles andere wird gegen die
// Prozessmarke gehalten wie bisher. Das ist die **erweiternde** Art einer
// Änderung nach [0033](../../docs/entscheidungen/0033-was-ueber-fassungen-hinweg-zugesagt-wird.md):
// wer heute eine Marke schickt, merkt nichts.
//
// # Woher die Erlaubnis kommt, die eine Anfrage braucht
//
// **Aus ihrer HTTP-Methode.** `GET` und `HEAD` verlangen `lesen`, alles andere
// `schreiben`. Das ist die Regel, die nicht auseinanderlaufen kann — jeder Weg
// dieses Dienstes ist mit seiner Methode eingetragen (`"GET /sets"`), und eine
// Liste daneben wäre eine zweite Wahrheit (§10).
//
// **Was sie kostet, gehört dazu**: ein paar `POST`-Wege arbeiten, ohne etwas
// abzulegen — `/paketdateien/pruefung` prüft eine Datei, `/symbol/{rechnung}`
// zeichnet, `/sammlung/liste/deuten` deutet eine Liste. Ein Token, das nur
// lesen darf, kommt dort **nicht** durch, obwohl nichts geschrieben würde.
// Die Abweichung geht in die sichere Richtung, und die Alternative — eine
// gepflegte Liste der harmlosen `POST`s — hat die Eigenschaft, still falsch zu
// werden.
//
// **Für den Preis gilt diese Regel nicht mehr** (#2020): was ein Ruf kostet,
// steht in `domain.Kostentafel`, und die einundzwanzig lesenden `POST`-Wege
// stehen dort mit der Klasse Lesen — wie ihre Werkzeuge über MCP. Die Methode
// entscheidet über den Preis nur noch dort, wo die Tafel schweigt. **Die
// Erlaubnis bleibt bei der Methode**: eine Liste harmloser Wege, die für den
// Preis still falsch würde, machte einen schreibenden Weg etwas billiger; für die
// Erlaubnis machte sie ihn einem lesenden Token zugänglich
// ([0163](../../docs/entscheidungen/0163-zeichnen-belegt-einen-platz-und-lesen-kostet-lesen.md)).

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Tokenvorspann steht vor jeder Kennung eines Zugriffstokens.
//
// **Er ist die Weiche** zwischen den beiden Arten von Nachweis, und er ist
// derselbe Text wie `service.Tokenmarke` samt Trenner. Dass beide
// zusammenpassen, hält `TestDieBeidenMarkenPassenZusammen`.
const Tokenvorspann = "mse_"

// Tokenwache prüft ein Zugriffstoken und sagt, in welchem Profil es arbeitet.
//
// **Sie gibt einen Kontext zurück und keine Kennung.** Wer das Profil an den
// Kontext hängt, ist `internal/storage` — und das kennt dieses Paket nicht
// (§2). Die Naht reicht deshalb das fertige Ergebnis herein.
type Tokenwache interface {
	// Pruefe gibt den ausgestatteten Kontext und die **Maschinenidentität**
	// des Tokens — Profil und Principal, nicht das Geheimnis.
	//
	// **Die Identität kam mit Befund 5 des Zweitblicks über Phase 5.2 dazu.**
	// Der Dienst budgetierte nach dem Hash des vorgezeigten Klartexts: eine
	// Rotation gab damit frisches Guthaben, und zwei Token derselben Maschine
	// hatten zwei Töpfe.
	Pruefe(ctx context.Context, klartext string, schreibend bool) (
		context.Context, string, error)
}

// istZugriffstoken sagt, ob ein Nachweis eines ist.
func istZugriffstoken(nachweis string) bool {
	return strings.HasPrefix(nachweis, Tokenvorspann)
}

// schreibend sagt, ob eine Methode eine Schreiberlaubnis verlangt.
//
// **`GET` und `HEAD` und sonst nichts.** RFC 9110 §9.2.1 nennt `OPTIONS` und
// `TRACE` ebenfalls sicher — dieser Dienst trägt keinen davon ein, und eine
// Erlaubnis für einen Weg, den es nicht gibt, wäre eine Zusage ohne Rumpf.
func schreibend(anfrage *http.Request) bool {
	return anfrage.Method != http.MethodGet && anfrage.Method != http.MethodHead
}

// nachweisAus zieht den Wert hinter `Bearer ` aus einer Anfrage.
func nachweisAus(anfrage *http.Request) (string, bool) {
	return strings.CutPrefix(anfrage.Header.Get("Authorization"), vorspann)
}

// ErlaubnisZuKnapp sagt, ob ein Fehlschlag an der Erlaubnis lag und nicht am
// Nachweis.
//
// **Exportiert, weil die Naht ihn beantworten können muss**: wer eine eigene
// [Tokenwache] baut, sagt darüber, dass seine Ablehnung eine Auskunft trägt.
func ErlaubnisZuKnapp(fehler error) bool {
	return errors.Is(fehler, domain.ErrErlaubnis)
}
