package rest

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Grenzen dieser Fläche: Kosten, Principals, Rumpf und Antwort** —
// Werkstück O5 von Phase 5.2 (#1429).
//
// `GET /grenzen` sagt, was gilt — maschinenlesbar, nach Vermerk 0033 (Regel 9):
// wer die Grenzen wissen will, fragt, statt sie aus einem 429 zu erraten.

// Kontingentnaht bucht das Guthaben eines Principals — die Gestalt von
// `service.Kontingente`, hier als Schnittstelle, denn `rest` kennt nur
// `domain` (§2); der Wurzelknoten reicht die Umsetzung herein.
type Kontingentnaht interface {
	Abbuchen(principal string, klasse domain.Kostenklasse, fenster bool) (bool, time.Duration)
}

// Sperrnaht führt die Sperre je vorgezeigtem Nachweis — dito.
type Sperrnaht interface {
	Gesperrt(nachweis string) (bool, time.Duration)
	Fehlgeschlagen(nachweis string)
	Gelungen(nachweis string)
}

// principalMarke ist der Principal aller Rufe mit der Prozessmarke: wer die
// Marke hat, hat die Datei — das ist **ein** Vertrauensraum, nicht viele.
const principalMarke = "marke"

// principalFuer ist der Principal eines Tokens.
//
// **Die Maschinenidentität, wo es sie gibt** — Befund 5 des Zweitblicks über
// Phase 5.2. Hier stand eine Prüfsumme über den vorgezeigten Klartext, und
// sie hatte zwei Folgen, die beide der Naht zwischen O2 und O5 widersprechen:
// eine **Rotation** gab frisches Guthaben (das Geheimnis war ein anderes), und
// zwei Token derselben Maschine hatten zwei Töpfe. Der Principal **ist** die
// Maschine (0080), und die Wache hängt ihn an den Kontext.
//
// **Die Prüfsumme bleibt als Rückfall** — für den Weg, auf dem die Wache gar
// nicht erst zum Zuge kam: ein abgewiesenes Token hat keinen Kontext, und
// seine Fehlversuche zählen trotzdem. Ein Klartext-Token hat in keiner
// Tabelle etwas verloren.
func principalFuer(identitaet, nachweis string) string {
	if identitaet != "" {
		return "token:" + identitaet
	}

	summe := sha256.Sum256([]byte(nachweis))

	return "token:" + hex.EncodeToString(summe[:8])
}

// KostenklasseFuer sagt, was ein Ruf dieser Fläche kostet — nach der
// Kostentafel und dem Muster, das der Pfad trifft (#2003).
//
// **Hier stand bis zum 2026-10-02 eine eigene, kurze Liste** — drei teure
// Wege, alles andere nach der Methode —, und MCP buchte nach Wörtern im
// Werkzeugnamen. Eine Karte zeichnen kostete deshalb an beiden 2 statt 5,
// ein Bogen hier 2 und dort 8. Jetzt fragen beide [domain.KostenklasseFuer].
//
// **Nach der Methode kostet nur noch, was die Tafel nicht nennt** — und sie
// nennt auch die lesenden `POST`-Wege (#2020): eine Vorschau mit Rumpf kostet
// Lesen, wie ihr Werkzeug.
//
// **Der Pfad darf eine Abfrage tragen** — `/sammlung/ausgeben?format=csv`
// trifft dasselbe Muster wie ohne.
func KostenklasseFuer(methode, pfad string) domain.Kostenklasse {
	ziel, fehler := url.ParseRequestURI(pfad)
	if fehler != nil {
		ziel = &url.URL{Path: pfad}
	}

	return kostenklasseAus(wegweiser(), &http.Request{Method: methode, URL: ziel})
}

// kostenklasseAus fragt die Tafel mit dem Muster, das die Anfrage trifft —
// ein Weg, den es nicht gibt, kostet nach seiner Methode.
//
// **Und mit dem Pfad, den sie trägt** (#2020): die Rechnungen des Symboleditors
// sind ein Muster, `POST /symbol/{rechnung}`, und die Tafel nennt jede mit
// ihrem Wert — so, wie `FAEHIGKEITEN.md` es tut. Ein Pfad mit Kennung
// (`/sets/12/bogen`) trifft keinen Eintrag; er hat das Muster.
func kostenklasseAus(weg *http.ServeMux, anfrage *http.Request) domain.Kostenklasse {
	_, muster := weg.Handler(anfrage)

	return domain.KostenklasseFuerSchreibweisen(domain.FlaecheREST, schreibend(anfrage),
		domain.RESTWeg(muster), anfrage.Method+" "+anfrage.URL.Path)
}

// wegweiser kennt jeden Weg des Dienstes — nur, um zu einem Pfad sein Muster
// zu finden. Bedient wird über ihn nichts.
var wegweiser = sync.OnceValue(func() *http.ServeMux {
	weg := http.NewServeMux()
	(&Dienst{}).wegeEintragen(weg)

	return weg
})

// ErlaubteRufe sind, wie viele Rufe eines Weges ein voller Principal am
// Stück schafft — für die Prüfung der Grenze je Endpunkt. **`-1` heißt:
// unbegrenzt**, ein Weg, der nichts kostet (ein Abbruch, #2003).
func ErlaubteRufe(grenzen domain.Haushaltsgrenzen, methode, weg string) int {
	kosten := KostenklasseFuer(methode, weg).Kosten()
	if kosten <= 0 {
		return -1
	}

	return grenzen.Principal.Burst / kosten
}

// Kontingent ist ein Eimer, wie die Fläche ihn nennt.
type Kontingent struct {
	Burst           int `json:"burst"`
	Nachfuellen     int `json:"nachfuellen"`
	FensterSekunden int `json:"fensterSekunden"`
}

// Grenzen ist die Antwort auf `GET /grenzen`.
type Grenzen struct {
	Parallel     map[string]int `json:"parallel"`
	AntwortBytes int64          `json:"antwortBytes"`
	Principal    Kontingent     `json:"principal"`
	Global       Kontingent     `json:"global"`
	Reserve      int            `json:"reserve"`
	Kosten       map[string]int `json:"kosten"`

	// Kontingente ist `an` oder `aus` — ob Principal und Global gelten
	// (QA-08.AK-03, 0161). **Wer kein 429 bekommt, soll nicht raten müssen,
	// warum.**
	Kontingente string `json:"kontingente"`

	// Verbindungen, Auftraege und Kostentafel — #2003: bis dahin erfuhr ein
	// Aufrufer die vier Aufträge aus dem Satz einer Ablehnung und die
	// sechzehn Verbindungen gar nicht.
	Verbindungen int                     `json:"verbindungen"`
	Auftraege    Auftragsgrenzen         `json:"auftraege"`
	Kostentafel  []domain.Kostenauskunft `json:"kostentafel"`
}

// Auftragsgrenzen sind die Grenzen des Auftragsbuchs, wie die Fläche sie
// nennt.
type Auftragsgrenzen struct {
	Laufend              int `json:"laufend"`
	Behalten             int `json:"behalten"`
	AufbewahrungSekunden int `json:"aufbewahrungSekunden"`
}

// GrenzenVon setzt die Grenzen in die Gestalt der Fläche um.
func GrenzenVon(grenzen domain.Haushaltsgrenzen) Grenzen {
	parallel := make(map[string]int, len(grenzen.Parallel))
	for klasse, anzahl := range grenzen.Parallel {
		parallel[string(klasse)] = anzahl
	}

	kosten := make(map[string]int, len(domain.Kostenklassen()))
	for _, klasse := range domain.Kostenklassen() {
		kosten[string(klasse)] = klasse.Kosten()
	}

	return Grenzen{
		Parallel:     parallel,
		AntwortBytes: grenzen.Antwort,
		Principal:    kontingentVon(grenzen.Principal),
		Global:       kontingentVon(grenzen.Global),
		Reserve:      grenzen.Reserve,
		Kosten:       kosten,
		Kontingente:  grenzen.Kontingentstand(),
		Verbindungen: grenzen.Verbindungen,
		Auftraege: Auftragsgrenzen{
			Laufend: grenzen.Auftraege.Laufend, Behalten: grenzen.Auftraege.Behalten,
			AufbewahrungSekunden: int(grenzen.Auftraege.Aufbewahrung.Seconds()),
		},
		Kostentafel: domain.KostentafelAlsAuskunft(),
	}
}

func kontingentVon(kontingent domain.Kontingent) Kontingent {
	return Kontingent{
		Burst:           kontingent.Burst,
		Nachfuellen:     kontingent.Nachfuellen,
		FensterSekunden: int(kontingent.Fenster.Seconds()),
	}
}

// Grenzenauskunft ist die Naht zur Anwendung: dieselbe Antwort wie am
// Fenster, über MCP und auf der Kommandozeile — die Grenzen des Prozesses.
type Grenzenauskunft interface {
	Grenzen(ctx context.Context) (Grenzen, error)
}

// grenzenAuskunft beantwortet `GET /grenzen`.
func (d *Dienst) grenzenAuskunft(schreiben http.ResponseWriter, anfrage *http.Request) {
	grenzen, fehler := d.naehte.Grenzen.Grenzen(anfrage.Context())
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, grenzen)
}

// begrenzterSchreiber lässt höchstens `rest` Bytes hinaus.
//
// **Dann bricht er ab, statt zu kürzen**: eine gekürzte JSON-Antwort sähe
// aus wie eine ganze, und der Aufrufer läse Unsinn. `http.ErrAbortHandler`
// schließt die Verbindung ohne Stapelspur — so ist es in `net/http` gemeint.
type begrenzterSchreiber struct {
	http.ResponseWriter
	rest int64
}

// Unwrap lässt `http.ResponseController` hindurch — sonst verlängerte keine
// Übertragung ihre Frist (FMT-03.AK-03).
func (b *begrenzterSchreiber) Unwrap() http.ResponseWriter { return b.ResponseWriter }

func begrenzt(schreiben http.ResponseWriter, hoechstens int64) *begrenzterSchreiber {
	return &begrenzterSchreiber{ResponseWriter: schreiben, rest: hoechstens}
}

func (b *begrenzterSchreiber) Write(p []byte) (int, error) {
	if int64(len(p)) > b.rest {
		panic(http.ErrAbortHandler)
	}

	b.rest -= int64(len(p))

	return b.ResponseWriter.Write(p)
}
