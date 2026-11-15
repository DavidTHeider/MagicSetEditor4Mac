package rest

// **Dateien hin und her über den Dienst** — FMT-03.AK-03
// (#1692).
//
// # Warum es das braucht
//
// **Ein Pfad ist eine Auskunft über den Rechner, auf dem der Dienst läuft.**
// Auf der Schleife ist das der Rechner des Aufrufers; seit N6 kann der Dienst
// jenseits davon horchen (mit TLS und Zugang), und dort hat der Aufrufer
// keinen Pfad auf diesem Rechner. Er schickt deshalb die **Datei selbst** —
// und holt eine Ausgabe ab, statt sie irgendwohin schreiben zu lassen.
//
// # Hin: die Datei im Rumpf
//
// `POST /sets`, `POST /sets/vorschau` und `POST /pakete/eingelesen` nehmen mit
// `Content-Type: application/zip` (oder `application/octet-stream`) die Datei
// selbst. Die Ablage legt sie **nach ihrem Inhalt benannt** in einen Eingang
// neben dem Bestand — dauerhaft, denn ein Set liest seine Bilder später aus
// seiner Quelle; und derselbe Inhalt bekommt denselben Ort, damit „eine
// Quelle, ein Set" auch für eine zweite Übertragung gilt. Danach geht die
// Anfrage mit diesem Pfad weiter, **durch den Pfadvertrag** wie jede andere.
//
// # Zurück: der Ausgang
//
// Die fünf Ausgabewege ohne `ziel` schreiben in einen Ausgang und antworten
// mit `Location: /ausgang/{name}`; `GET` holt die Datei ab, `DELETE` vergisst
// sie. **Als Auftrag** geschieht das im Lauf selbst — eine Wiederholung
// bekommt einen frischen Ort, statt an dem der vorigen zu scheitern.
//
// # Was dieses Blatt nicht tut
//
// **Es fasst keine Datei an.** Wo Eingang und Ausgang liegen und wie lange
// eine Ausgabe bleibt, weiß die Naht ([Dateiablage]); hier steht nur, welche
// Wege es gibt und wie eine Anfrage umgeschrieben wird.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"mime"
	"net/http"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Hochladegrenze ist, wie groß eine übertragene Datei sein darf. **Größer als
// eine JSON-Anfrage**, denn ein Set mit seinen Bildern wiegt schnell hundert
// Megabyte — und die Ablage schreibt sie auf die Platte, statt sie im
// Speicher zu halten.
const Hochladegrenze int64 = 1 << 30

// Hochladefrist und Abholfrist sind, wie lange eine Übertragung dauern darf —
// und die Hochladefrist auch, wie lange danach die Antwort.
//
// **Gemessen und nicht angenommen** (`internal/flaechen/uebertragung_test.go`):
// der Dienst liest eine Anfrage in zehn Sekunden (`cmd/mse-rest`), und bis
// FMT-03 galt das für den ganzen Rumpf — ein Set über eine langsame Leitung
// brach mitten im Lesen ab und kam als `500` zurück. **Die Frist für den Kopf
// bleibt**: verlängert wird erst hinter der Marke, für einen Aufrufer, der
// sich ausgewiesen hat, und nur für diese eine Anfrage.
const (
	Hochladefrist = 30 * time.Minute
	Abholfrist    = 30 * time.Minute
)

// Dateiablage hält übertragene Dateien und Ausgaben zum Abholen.
type Dateiablage interface {
	// Ablegen legt einen Inhalt ab und gibt seinen Ort — derselbe Inhalt,
	// derselbe Ort.
	Ablegen(ctx context.Context, endung string, inhalt io.Reader) (string, error)
	// Zwischenlegen legt einen Inhalt für **eine** Anfrage ab — die Vorschau;
	// `weg` nimmt ihn wieder weg (#1944).
	Zwischenlegen(ctx context.Context, endung string, inhalt io.Reader) (pfad string, weg func(), err error)
	// Ausgangsort gibt einen freien Ort für eine Ausgabe.
	Ausgangsort(ctx context.Context, endung string) (name, pfad string, err error)
	// Abholen öffnet eine Ausgabe; Vergessen nimmt sie weg.
	Abholen(ctx context.Context, name string) (io.ReadCloser, error)
	Vergessen(ctx context.Context, name string) error
}

// hochladeweg sagt, welche Endung eine übertragene Datei bekommt, in
// welchem Feld ihr Pfad weitergeht und ob sie nur für diese Anfrage liegt.
type hochladeweg struct {
	endung, feld string
	fluechtig    bool
}

var hochladewege = map[string]hochladeweg{
	"POST /sets": {".mse-set", "pfad", false},
	// **Die Vorschau liest nur** — ihre Datei liegt für diese Anfrage und
	// geht danach, statt den Eingang wachsen zu lassen (#1944).
	"POST /sets/vorschau":     {".mse-set", "pfad", true},
	"POST /pakete/eingelesen": {".mse-installer", "datei", false},
}

// ausgangswege sind die Ausgaben, die ohne `ziel` in den Ausgang gehen.
var ausgangswege = map[string]string{
	"POST /sets/{setID}/sicherung": ".mse-set",
	"POST /karten/{karteID}/bild":  ".png",
	"POST /sets/{setID}/bogen":     ".pdf",
	"POST /sets/{setID}/mtgjson":   ".json",
	"POST /pakete/weitergabe":      ".mse-installer",
}

// Uebertragungswege sind die Wege, die eine Datei annehmen oder in den
// Ausgang schreiben — für den Wächter über `docs/DATEIOPERATIONEN.md`.
func Uebertragungswege() []string {
	wege := slices.Collect(maps.Keys(hochladewege))

	return slices.Sorted(slices.Values(append(wege, slices.Collect(maps.Keys(ausgangswege))...)))
}

// dateiarten nennt die Art einer abgeholten Datei — **eine eigene Liste**,
// denn `mime.TypeByExtension` liest auf manchen Systemen Dateien des
// Rechners, und die Antwort hinge von ihnen ab.
var dateiarten = map[string]string{
	".pdf": "application/pdf", ".png": "image/png", ".json": "application/json",
}

// dateiErbeten sagt, ob der Rumpf eine Datei ist und kein JSON.
func dateiErbeten(anfrage *http.Request) bool {
	art, _, _ := mime.ParseMediaType(anfrage.Header.Get("Content-Type")) //nolint:errcheck // eine leere Art ist keine Datei

	return art == "application/zip" || art == "application/octet-stream"
}

// mitHochladen legt eine übertragene Datei ab und schickt ihren Pfad weiter.
func (d *Dienst) mitHochladen(weg *http.ServeMux, weiter http.Handler) http.Handler {
	return http.HandlerFunc(func(schreiben http.ResponseWriter, anfrage *http.Request) {
		_, muster := weg.Handler(anfrage)

		ziel, gibt := hochladewege[muster]
		if !gibt || !dateiErbeten(anfrage) || d.naehte.Dateiablage == nil {
			weiter.ServeHTTP(schreiben, anfrage)

			return
		}

		fristen := http.NewResponseController(schreiben)
		_ = fristen.SetReadDeadline(time.Now().Add(Hochladefrist)) //nolint:errcheck // ein Schreiber ohne Fristen — ein Test — lässt sie, wie sie ist

		pfad, weg, fehler := d.ablegen(anfrage.Context(), ziel, anfrage.Body)
		defer weg()

		// **Die Schreibfrist zählt ab dem Anfragekopf** (`WriteTimeout`) und
		// wäre nach einer langen Übertragung schon vorbei — dann bekäme der
		// Aufrufer nicht einmal ein `413`. Sie beginnt deshalb neu, wenn die
		// Datei da ist, und reicht so weit wie die Übertragung selbst.
		_ = fristen.SetWriteDeadline(time.Now().Add(Hochladefrist)) //nolint:errcheck // wie darüber

		var zuGross *http.MaxBytesError
		if errors.As(fehler, &zuGross) {
			nein(schreiben, http.StatusRequestEntityTooLarge, "validierung", "die Datei ist zu groß")

			return
		} else if fehler != nil {
			fehlerhaft(schreiben, fehler)

			return
		}

		umgeschrieben(anfrage, map[string]any{ziel.feld: pfad})
		weiter.ServeHTTP(schreiben, anfrage)
	})
}

// ablegen legt eine übertragene Datei ab — dauerhaft in den Eingang oder,
// für eine Vorschau, nur zwischen. `weg` räumt eine zwischengelegte nach der
// Anfrage wieder fort und tut sonst nichts.
func (d *Dienst) ablegen(ctx context.Context, ziel hochladeweg, inhalt io.Reader) (pfad string, weg func(), fehler error) {
	if !ziel.fluechtig {
		pfad, fehler = d.naehte.Dateiablage.Ablegen(ctx, ziel.endung, inhalt)

		return pfad, func() {}, fehler
	}

	pfad, weg, fehler = d.naehte.Dateiablage.Zwischenlegen(ctx, ziel.endung, inhalt)
	if fehler != nil {
		return "", func() {}, fehler
	}

	return pfad, weg, nil
}

// mitAusgang gibt einer Ausgabe ohne `ziel` einen Ort im Ausgang.
func (d *Dienst) mitAusgang(weg *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(schreiben http.ResponseWriter, anfrage *http.Request) {
		_, muster := weg.Handler(anfrage)

		endung, gibt := ausgangswege[muster]
		felder, lesbar := rumpffelder(anfrage)

		if !gibt || !lesbar || d.naehte.Dateiablage == nil || zielGenannt(felder) {
			weg.ServeHTTP(schreiben, anfrage)

			return
		}

		name, pfad, fehler := d.naehte.Dateiablage.Ausgangsort(anfrage.Context(), endung)
		if fehler != nil {
			fehlerhaft(schreiben, fehler)

			return
		}

		// **Vermerkt, bevor der Weg schreibt** (#2002): wer den Namen aus
		// der Antwort kennt, soll ihn nicht früher abholen können als der,
		// dem sie gehört.
		d.ausgaenge.vermerke(name, principalAus(anfrage.Context()))

		felder["ziel"] = pfad
		umgeschrieben(anfrage, felder)
		weg.ServeHTTP(&ausgangsschreiber{ResponseWriter: schreiben, ort: "/ausgang/" + name}, anfrage)
	})
}

// rumpffelder liest den Rumpf als Objekt und legt ihn zurück. Der zweite
// Wert ist falsch, wenn er keines ist — dann sagt der Weg, was nicht stimmt.
func rumpffelder(anfrage *http.Request) (map[string]any, bool) {
	roh, fehler := io.ReadAll(io.LimitReader(anfrage.Body, rumpfgrenze+1))
	anfrage.Body = io.NopCloser(io.MultiReader(bytes.NewReader(roh), anfrage.Body))

	var felder map[string]any
	if fehler != nil || json.Unmarshal(roh, &felder) != nil || felder == nil {
		return nil, false
	}

	return felder, true
}

// zielGenannt sagt, ob der Aufrufer selbst ein Ziel nennt — **in jeder
// Schreibweise, die der Dekodierer dem Feld `ziel` zuordnet** (#2073). Sonst
// legte der Dienst seinen Ausgang neben ein `"Ziel"`, und der Weg schriebe
// still dorthin.
func zielGenannt(felder map[string]any) bool {
	for name, wert := range felder {
		if ziel, ist := wert.(string); ist && ziel != "" && strings.EqualFold(name, "ziel") {
			return true
		}
	}

	return false
}

// umgeschrieben ersetzt den Rumpf durch JSON.
func umgeschrieben(anfrage *http.Request, felder map[string]any) {
	// **Ohne Fehlerweg**: Zeichenketten, Listen und Zahlen aus einem Rumpf,
	// der gerade gelesen wurde.
	roh, _ := json.Marshal(felder) //nolint:errcheck,errchkjson // siehe darüber

	anfrage.Body = io.NopCloser(bytes.NewReader(roh))
	anfrage.ContentLength = int64(len(roh))
	anfrage.Header.Set("Content-Type", "application/json")
}

// ausgangsschreiber nennt den Ort einer geglückten Ausgabe.
type ausgangsschreiber struct {
	http.ResponseWriter

	ort string
}

func (a *ausgangsschreiber) WriteHeader(status int) {
	if status < http.StatusMultipleChoices {
		a.Header().Set("Location", a.ort)
	}

	a.ResponseWriter.WriteHeader(status)
}

// rumpfgrenzeFuer ist die Grenze des Rumpfs — die einer Übertragung oder
// die einer Anfrage.
func (d *Dienst) rumpfgrenzeFuer(anfrage *http.Request) int64 {
	if !dateiErbeten(anfrage) {
		return d.grenzen.Antwort
	}

	if d.hochladegrenze > 0 {
		return d.hochladegrenze
	}

	return Hochladegrenze
}

// MitHochladegrenze setzt eine andere Grenze für übertragene Dateien — für
// einen Betrieb, der engere will, und für Tests.
func (d *Dienst) MitHochladegrenze(grenze int64) *Dienst {
	d.hochladegrenze = grenze

	return d
}

// istAbholung sagt, ob eine Anfrage eine Ausgabe abholt — sie ist von der
// Grenze einer Antwort ausgenommen, denn ein Bogen über dreihundert Karten
// ist größer als jede JSON-Antwort.
func istAbholung(anfrage *http.Request) bool {
	verzeichnis, _ := path.Split(anfrage.URL.Path)

	return anfrage.Method == http.MethodGet && verzeichnis == "/ausgang/"
}

// nichtSeine sagt einem anderen Principal dasselbe wie für einen Namen, den
// es nicht gibt (#2002) — ein `403` verriete, dass es ihn gibt.
func (d *Dienst) nichtSeine(anfrage *http.Request, name string) error {
	if d.ausgaenge.gehoert(name, principalAus(anfrage.Context())) {
		return nil
	}

	return fmt.Errorf("%w: die Ausgabe %q gibt es nicht — abgeholt, vergessen oder älter als eine Stunde",
		domain.ErrNotFound, name)
}

func (d *Dienst) ausgangAbholen(schreiben http.ResponseWriter, anfrage *http.Request) {
	name := anfrage.PathValue("name")

	//nolint:errcheck // wie beim Hochladen
	_ = http.NewResponseController(schreiben).SetWriteDeadline(time.Now().Add(Abholfrist))

	if fehler := d.nichtSeine(anfrage, name); fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	inhalt, fehler := d.naehte.Dateiablage.Abholen(anfrage.Context(), name)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	defer func() { _ = inhalt.Close() }() //nolint:errcheck // nur gelesen

	art, bekannt := dateiarten[path.Ext(name)]
	if !bekannt {
		art = "application/octet-stream"
	}

	schreiben.Header().Set("Content-Type", art)
	schreiben.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": name}))
	schreiben.WriteHeader(http.StatusOK)

	_, _ = io.Copy(schreiben, inhalt) //nolint:errcheck // die Verbindung ist fort
}

func (d *Dienst) ausgangVergessen(schreiben http.ResponseWriter, anfrage *http.Request) {
	name := anfrage.PathValue("name")

	if fehler := d.nichtSeine(anfrage, name); fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	if fehler := d.naehte.Dateiablage.Vergessen(anfrage.Context(), name); fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	d.ausgaenge.vergiss(name)
	schreiben.WriteHeader(http.StatusNoContent)
}
