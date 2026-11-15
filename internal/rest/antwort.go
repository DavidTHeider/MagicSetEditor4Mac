package rest

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Vorgangskopf ist die Kopfzeile, in der **jede** Antwort ihre
// Vorgangskennung trägt — Werkstück O9 von Phase 5.2
// (#1433).
//
// # Im Kopf und nicht im Rumpf
//
// **Weil sie an jede Antwort gehört, auch an die geglückte.** Der Rumpf einer
// geglückten Antwort ist die Nutzlast; ein Feld daneben änderte die Gestalt
// jeder einzelnen. Wer eine langsame Anfrage nachsehen will, braucht den
// Griff genauso wie bei einer gescheiterten.
//
// # Gesetzt wird sie an der Wurzel
//
// `cmd/mse-rest` würfelt die Kennung je Anfrage und setzt diese Zeile, bevor
// der erste Weg läuft — ein Kopf nach dem Statuscode kommt nicht mehr an. Der
// **Name** steht hier, weil er zur Fläche gehört und nicht zu ihrer Wurzel:
// eine zweite Zeichenkette in einem Test liefe auseinander.
//
// **`X-` und kein registrierter Name**: dies ist keine Kopfzeile der
// IANA-Liste, und ein Name ohne Präfix behauptete, eine zu sein.
const Vorgangskopf = "X-Vorgang"

// Befund ist die Gestalt jeder Fehlerantwort.
//
// **Eine Gestalt für alle**, damit ein Skript sie einmal lesen lernt. `Code`
// ist maschinenlesbar und bleibt; `Meldung` ist für Menschen und darf sich
// ändern.
type Befund struct {
	Code    string `json:"code"`
	Meldung string `json:"meldung"`

	// Teilstand ist, was **schon geschrieben** wurde, als es scheiterte
	// (#1041).
	//
	// **Wahlfrei, und deshalb `omitempty`**: ein `teilstand: null` in jeder
	// Fehlerantwort ließe jede Seite zweimal prüfen. Wo nichts geschrieben
	// wurde, steht das Feld nicht da.
	//
	// **Erweiternd im Sinne von 0033**: ein Feld kommt hinzu, Vorhandenes
	// bleibt. Ein älterer Klient liest die Antwort weiter und sieht es nicht.
	Teilstand any `json:"teilstand,omitempty"`
}

// nein schreibt eine Fehlerantwort.
func nein(schreiben http.ResponseWriter, status int, code, meldung string) {
	neinMit(schreiben, status, code, meldung, nil)
}

// neinMit schreibt eine Fehlerantwort samt dem, was schon geschrieben wurde.
func neinMit(schreiben http.ResponseWriter, status int, code, meldung string, teilstand any) {
	schreiben.Header().Set("Content-Type", "application/json; charset=utf-8")
	schreiben.WriteHeader(status)

	// **Ein Schreibfehler bleibt hier still.** Die Antwort ist schon
	// angefangen; ein zweiter Statuscode ginge nicht mehr, und ein Protokoll
	// gehört nicht in ein Blatt.
	//
	//nolint:errcheck,errchkjson // siehe oben
	_ = json.NewEncoder(schreiben).Encode(
		Befund{Code: code, Meldung: meldung, Teilstand: teilstand})
}

// ja schreibt eine geglückte Antwort.
func ja(schreiben http.ResponseWriter, status int, rumpf any) {
	schreiben.Header().Set("Content-Type", "application/json; charset=utf-8")
	schreiben.WriteHeader(status)

	//nolint:errcheck,errchkjson // wie bei [nein]: die Antwort ist angefangen
	_ = json.NewEncoder(schreiben).Encode(rumpf)
}

// fehlerhaft übersetzt einen Fehler aus dem Fachkern auf HTTP.
//
// # Die Zuordnung, und warum sie so lautet
//
// | Fehler | Status | Warum |
// | --- | --- | --- |
// | [domain.ErrValidation], [domain.ErrGesperrt] | 400 | die Eingabe stimmt nicht — oder ihr Ort ist gesperrt |
// | [domain.ErrNotFound] | 404 | den Eintrag gibt es nicht |
// | [domain.ErrConflict] | 409 | überholt, doppelt oder gegen eine Regel des Bestands — die Meldung sagt, welches |
// | [domain.ErrErlaubnis] | 403 | du bist es, du darfst nur nicht |
// | [domain.ErrNurLesend] | 403 | dasselbe, mit Code `nurLesend` |
// | [domain.ErrBelegt] | **503** | ein anderer schreibt gerade; **gleich nochmal** |
// | [domain.ErrVoll], `ENOSPC` | **507** | kein Platz mehr; erst Platz schaffen (#1812) |
// | [domain.ErrKontingent] | **429** | über dem eigenen Budget; `Retry-After` nennt, wann wieder |
// | Abbruch | 499 | der Aufrufer ist gegangen |
// | sonst | 500 | |
//
// **Der belegte Bestand ist ausdrücklich kein 409.** Der heißt hier „dein Stand
// ist überholt", und die richtige Antwort darauf ist Neulesen — was in dieselbe
// Sperre liefe. Bei 503 mit `Retry-After` heißt die Antwort: gleich nochmal,
// unverändert. Das ist der Unterschied, um den es
// [0045](../../docs/entscheidungen/0045-zwei-programme-an-einem-bestand.md)
// geht.
//
// **403 und nicht 401** — Werkstück 11 von Phase 4.12 (#1351). Ein 401 hieße
// „weise dich aus", und genau das hat der Aufrufer getan; sein Nachweis reicht
// nur nicht so weit. Dieselbe Unterscheidung trifft schon die Tokenwache, und
// sie steht dort ausgeschrieben.
//
// **Vor dem 2026-09-06 landete er im 500er-Zweig**, weil ihn nur die
// Tokenwache erzeugte und die ihn selbst beantwortet. Seit der Profilwechsel
// eine Rolle verlangt, kommt er aus dem Dienst — und „der Aufruf ist
// gescheitert" war die falsche Auskunft für „das darfst du nicht".
//
// **499 ist kein Standard**, sondern die Zahl, die nginx dafür eingeführt hat
// und die verbreitet gelesen wird. Der Aufrufer sieht sie ohnehin nicht — er
// ist weg; sie steht im Protokoll und unterscheidet dort einen Abbruch von
// einem Fehler.
func fehlerhaft(schreiben http.ResponseWriter, fehler error) {
	fehlerhaftMit(schreiben, fehler, nil)
}

// fehlerhaftMit meldet den Fehler **samt dem, was schon geschrieben wurde**
// (#1041).
//
// # Warum das nötig ist
//
// Ein Aufruf kann eine Karte anlegen und danach am Füllen scheitern. Die Karte
// ist dann **da** — und ein Aufrufer, der nur den Fehler bekommt, findet sie
// nicht und kann sie nicht zurücknehmen. §7 Regel 3 deckt den Zustand
// („geschrieben ist geschrieben"); die **Auskunft** darüber fehlte.
func fehlerhaftMit(schreiben http.ResponseWriter, fehler error, teilstand any) {
	// **Eine volle Platte ist ein voller Datenträger, gleich wer schrieb**
	// (#1812) — und ihr Pfad bleibt draußen (§8).
	fehler = domain.Platzmangel(fehler)

	switch {
	case errors.Is(fehler, context.Canceled), errors.Is(fehler, context.DeadlineExceeded):
		neinMit(schreiben, 499, "abgebrochen", "der Aufruf wurde abgebrochen", teilstand)

	// **Ein veralteter Dienst ist belegt, aber mit eigenem Satz** (M02,
	// #2346): er endet gerade, und nach dem Warten antwortet die neue Fassung.
	case istVeraltet(fehler):
		schreiben.Header().Set("Retry-After", "1")
		neinMit(schreiben, http.StatusServiceUnavailable, "belegt", fehler.Error(), teilstand)

	// **Vor den Fachfehlern**, weil ein belegter Bestand deren Meldung
	// umschließt und hinter ihnen einsortiert nie getroffen würde.
	case errors.Is(fehler, domain.ErrBelegt):
		schreiben.Header().Set("Retry-After", "1")
		neinMit(schreiben, http.StatusServiceUnavailable, "belegt",
			"der Bestand ist gerade belegt — bitte gleich noch einmal versuchen", teilstand)

	case errors.Is(fehler, domain.ErrVoll):
		neinMit(schreiben, http.StatusInsufficientStorage, "voll", fehler.Error(), teilstand)

	// **Das eigene Kontingent ist derselbe 429 wie das Budget des Dienstes**
	// (Nachlauf über Phase 1 von 0.9.0, Befund 4): bis dahin kam es aus dem
	// Fachkern als 500 an. Ohne Grenze gilt die kleinste Wartezeit.
	case errors.Is(fehler, domain.ErrKontingent):
		grenze := &domain.Kontingentgrenze{}
		_ = errors.As(fehler, &grenze)

		schreiben.Header().Set("Retry-After", strconv.Itoa(grenze.Sekunden()))
		neinMit(schreiben, http.StatusTooManyRequests, "drossel", fehler.Error(), teilstand)

	// **Ein fremder Dienst dahinter** (D02, #2233): dieser Dienst arbeitet,
	// der fremde nicht — 504, wo eine Frist griff, sonst 502. Der Satz trägt
	// den Wirt und keine Abfrage (`internal/netz`).
	case errors.Is(fehler, domain.ErrKeineAntwort), errors.Is(fehler, domain.ErrAntwortStockt):
		neinMit(schreiben, http.StatusGatewayTimeout, "netz", fehler.Error(), teilstand)
	case errors.Is(fehler, domain.ErrGegenstelle), errors.Is(fehler, domain.ErrGedrosselt):
		neinMit(schreiben, http.StatusBadGateway, "netz", fehler.Error(), teilstand)

	case errors.Is(fehler, domain.ErrNotFound):
		neinMit(schreiben, http.StatusNotFound, "nichtGefunden", fehler.Error(), teilstand)

	case errors.Is(fehler, domain.ErrConflict):
		neinMit(schreiben, http.StatusConflict, "konflikt", fehler.Error(), teilstand)

	// **Vor [domain.ErrValidation]** — aus demselben Grund wie in
	// `internal/app`: eine fehlende Rolle ist keine falsche Eingabe, und der
	// Aufrufer soll nicht dasselbe noch einmal schicken.
	case errors.Is(fehler, domain.ErrErlaubnis):
		neinMit(schreiben, http.StatusForbidden, "erlaubnis", fehler.Error(), teilstand)

	// **Nur lesen ist dasselbe Nein, mit eigenem Code** (ID-03, #1702):
	// ohne diesen Zweig bekäme ein Leser, der schreiben wollte, „500, der
	// Aufruf ist gescheitert" — als wäre etwas kaputt.
	case errors.Is(fehler, domain.ErrNurLesend):
		neinMit(schreiben, http.StatusForbidden, "nurLesend", fehler.Error(), teilstand)

	// **Ein abgeschalteter Bereich ist kein Fehler des Aufrufs** (OPS-10,
	// #1908): dasselbe Nein, mit eigenem Code — einschalten hilft, nicht
	// noch einmal schicken.
	case errors.Is(fehler, domain.ErrAbgeschaltet):
		neinMit(schreiben, http.StatusForbidden, "abgeschaltet", fehler.Error(), teilstand)

	// **Ein gesperrter Ort ist eine Eingabe, die nicht passt** — wie in
	// `internal/app` (FIX-F02): der Rat ist ein anderer oder freigegebener
	// Ort (Nachlauf über Phase 2 von 0.9.0, Befund 7).
	case errors.Is(fehler, domain.ErrValidation), errors.Is(fehler, domain.ErrGesperrt):
		neinMit(schreiben, http.StatusBadRequest, "validierung", fehler.Error(), teilstand)

	default:
		// **Der unbekannte Fehler wird nicht durchgereicht.** Er könnte einen
		// Pfad oder eine SQLite-Meldung tragen, und beides gehört nicht an
		// eine Schnittstelle (§8).
		neinMit(schreiben, http.StatusInternalServerError, "unbekannt",
			"der Aufruf ist gescheitert", teilstand)
	}
}

// istVeraltet sagt, ob dieser Prozess älter ist als der Bestand.
func istVeraltet(fehler error) bool {
	var veraltet *domain.ProzessVeraltet

	return errors.As(fehler, &veraltet)
}
