// **Den Bestand sichern, hineinsehen und zurückspielen über den Dienst** —
// #1169,
// Werkstück 2.
//
// # Zwei Wege und nicht drei
//
// **Zurückgespielt wird hier nicht, und der Grund ist technisch und nicht
// fachlich.** Der Dienst hält **eine** Verbindung über seine ganze Laufzeit,
// und jede seiner rund fünfundzwanzig Nähte hat sie sich beim Aufbau
// abgeholt. Ein Rückspielvorgang schließt sie, tauscht die Datei und öffnet
// eine **neue** — und danach hielten alle Nähte einen Griff auf eine Datei,
// die es nicht mehr gibt. Der nächste Aufruf meldete etwas ganz anderes als
// den eigentlichen Grund.
//
// **Richtig ginge es**, indem jede Naht den Zeiger auf den Zeiger bekäme.
// Das ist ein Umbau von fünfundzwanzig Stellen für einen Endpunkt, und er
// gehört entschieden und nicht nebenbei gemacht.
//
// **Dasselbe gilt für den Neuanfang** (G09,
// #2273):
// `POST /ruecksetzung/anwendung` tauschte bis dahin trotzdem, und danach
// antwortete der Dienst auf jede Anfrage mit einem Fehler. Die Rücksetzung des
// Dienstes nennt jetzt das Hindernis `ohneTausch`, und ein Wächter hält jeden
// Tauschweg aus den langlebigen Flächen heraus.
//
// **Die Kommandozeile hat das Problem nicht**: sie öffnet, arbeitet, endet.
// Deshalb steht `mse bestand-zurueckspielen` dort und hier nicht — und das
// ist der Grund, den §16 verlangt, wenn eine Fähigkeit an einer Fläche fehlt.
//
// # Was hier steht
//
// Sichern und Hineinsehen. Beides lässt den laufenden Bestand unberührt: das
// eine legt daneben, das andere liest.

package rest

import (
	"context"
	"net/http"
	"strings"
)

// Bestandssicherung ist, was die drei Endpunkte dieser Datei brauchen.
type Bestandssicherung interface {
	// BestandSichern schreibt die Sicherung — mit Kennwort, wenn eines
	// kommt, und ausdrücklich im Klartext nur mit `klartext` (OPS-06).
	BestandSichern(ctx context.Context, ziel, kennwort string, klartext bool) error
	SicherungSchauen(ctx context.Context, pfad, kennwort string) (Sicherungsangaben, error)

	// Verschluesselungsstand sagt, ob der Bestand verschlüsselt auf der Platte
	// liegt — Werkstück 14 von Phase 4.12.
	//
	// **Nur das Lesen und nicht das Umschalten**, und der Grund ist derselbe,
	// aus dem hier auch kein Zurückspielen steht: das Umschalten **tauscht die
	// Bestandsdatei**. Dieser Prozess hält eine Verbindung über seine ganze
	// Laufzeit, und jedes Repository hängt daran; nach einem Tausch hielten
	// sie alle einen Griff auf eine Datei, die es nicht mehr gibt.
	//
	// Geschaltet wird am Fenster und mit `mse bestand-verschluesselung` —
	// beides sind Stellen, die ihren Lebenszyklus danach neu aufbauen.
	Verschluesselungsstand() (bool, error)

	// Sicherungsstand nennt das Alter der letzten Sicherung und den Plan —
	// OPS01 (#2347), in der Gestalt aus `internal/app`, an jeder Fläche
	// dieselbe.
	Sicherungsstand(ctx context.Context) (any, error)
}

// Sicherungsangaben sind, was eine Sicherung über sich hergibt.
type Sicherungsangaben struct {
	AngelegtAm  string `json:"angelegtAm"`
	Schemastand string `json:"schemastand"`
	Sets        int    `json:"sets"`
	Karten      int    `json:"karten"`
	Bilder      string `json:"bilder"`

	Verschluesselt bool `json:"verschluesselt"`

	// PaketeEnthalten und PaketeFehlen: welche eigenen Pakete mitkommen und
	// welche bezogenen hier neu zu beziehen sind (OPS-06.AK-03).
	PaketeEnthalten []string `json:"paketeEnthalten"`
	PaketeFehlen    []string `json:"paketeFehlen"`
}

// Kennwortkopf trägt das Kennwort einer verschlüsselten Sicherung beim
// Lesen — OPS-06.AK-02
// (#1723).
//
// **Ein Kopf und keine Abfrage**: eine Adresse steht in Protokollen,
// Verläufen und Zwischenspeichern; ein Kopf nicht. Beim Sichern steht das
// Kennwort im Rumpf, wie jede andere Angabe eines `POST`.
const Kennwortkopf = "X-Sicherungskennwort"

// pflichttext verlangt eine nicht leere Angabe.
//
// **Neben [ausRumpf] und nicht statt ihr**: die dort verlangt eine
// **Kennung**, hier steht ein Pfad. Ein Pfad, der als Zahl gelesen würde, wäre
// eine Meldung über die falsche Sache.
func pflichttext(schreiben http.ResponseWriter, name, roh string) (string, bool) {
	if strings.TrimSpace(roh) == "" {
		nein(schreiben, http.StatusBadRequest, "validierung",
			name+" darf nicht leer sein")

		return "", false
	}

	return roh, true
}

// bestandSichern schreibt den ganzen Bestand an einen Ort.
func (d *Dienst) bestandSichern(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf struct {
		Ziel     string `json:"ziel"`
		Kennwort string `json:"kennwort"`
		Klartext bool   `json:"klartext"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	ziel, gut := pflichttext(schreiben, "ziel", rumpf.Ziel)
	if !gut {
		return
	}

	if fehler := d.naehte.Sicherung.BestandSichern(anfrage.Context(), ziel,
		rumpf.Kennwort, rumpf.Klartext); fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusCreated, map[string]any{"ziel": ziel})
}

// sicherungsstandSchauen nennt das Alter der letzten Sicherung.
func (d *Dienst) sicherungsstandSchauen(schreiben http.ResponseWriter, anfrage *http.Request) {
	stand, fehler := d.naehte.Sicherung.Sicherungsstand(anfrage.Context())
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, stand)
}

// verschluesselungSchauen sagt, ob der Bestand verschlüsselt daliegt.
//
// **Nur lesend** — der Grund steht bei [Bestandssicherung].
func (d *Dienst) verschluesselungSchauen(
	schreiben http.ResponseWriter, _ *http.Request,
) {
	an, fehler := d.naehte.Sicherung.Verschluesselungsstand()
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"verschluesselt": an})
}

// sicherungSchauen sagt, was in einer Sicherung liegt — **ohne** sie zu nehmen.
func (d *Dienst) sicherungSchauen(schreiben http.ResponseWriter, anfrage *http.Request) {
	pfad, gut := pflichttext(schreiben, "pfad", anfrage.URL.Query().Get("pfad"))
	if !gut {
		return
	}

	angaben, fehler := d.naehte.Sicherung.SicherungSchauen(anfrage.Context(), pfad,
		anfrage.Header.Get(Kennwortkopf))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, angaben)
}
