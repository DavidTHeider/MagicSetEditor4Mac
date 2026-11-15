package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// sets gibt die offenen Sets.
func (d *Dienst) sets(schreiben http.ResponseWriter, anfrage *http.Request) {
	koepfe, fehler := d.naehte.Leser.Setkoepfe(anfrage.Context())
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"sets": alsSetzeilen(koepfe)})
}

// karten gibt die Karten eines Sets.
func (d *Dienst) karten(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gut := kennung(schreiben, anfrage, "setID")
	if !gut {
		return
	}

	zeilen, fehler := d.naehte.Leser.Karten(anfrage.Context(), setID)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	// **Ein Set ohne Karten und ein Set, das es nicht gibt, sind zweierlei.**
	// Der Leser gibt für beides eine leere Liste; ohne diese Unterscheidung
	// bekäme ein Aufrufer auf eine erfundene Kennung ein fröhliches 200.
	//
	// **Und es sind drei Zustände, nicht zwei**: da, nicht da, und **nicht zu
	// sagen**. Der dritte ging bis zum 2026-08-26 als „nicht da" hinaus, und
	// damit konnte ein Aufrufer „der Datensatz fehlt" nicht von „der Bestand
	// ist nicht lesbar" unterscheiden — die Auskunft, für die es §8 gibt
	// (#1155).
	if len(zeilen) == 0 {
		da, nachsehfehler := d.setGibtEs(anfrage, setID)
		if nachsehfehler != nil {
			fehlerhaft(schreiben, nachsehfehler)

			return
		}

		if !da {
			fehlerhaft(schreiben, fmt.Errorf("%w: das Set %d", domain.ErrNotFound, setID))

			return
		}
	}

	ja(schreiben, http.StatusOK, map[string]any{"karten": alsKartenzeilen(zeilen)})
}

// feld setzt einen Feldwert.
func (d *Dienst) feld(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gutesSet := kennung(schreiben, anfrage, "setID")
	karteID, guteKarte := kennung(schreiben, anfrage, "karteID")

	if !gutesSet || !guteKarte {
		return
	}

	// **Ein fehlender Wert gilt hier als leer** — anders als bei MCP und
	// beim Zusatzwert, die ihn abweisen. So steht der Weg seit 0.2.0 da, und
	// ihn abzuweisen wäre brechend: das braucht eine Hauptfassung
	// ([0033](../../docs/entscheidungen/0033-was-ueber-fassungen-hinweg-zugesagt-wird.md), Regel 7).
	var rumpf struct {
		Wert string `json:"wert"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	vorher, hatte, fehler := d.naehte.Schreiber.FeldSetzenMitVorwert(
		anfrage.Context(), setID, karteID, anfrage.PathValue("feld"), rumpf.Wert)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	// **Der vorige Wert reist mit**, damit sich die Änderung zurücknehmen
	// lässt — dieselbe Zusage wie bei MCP
	// ([0010](../../docs/entscheidungen/0010-mcp-als-eigener-prozess.md)).
	ja(schreiben, http.StatusOK, map[string]any{"vorher": vorher, "hatte": hatte})
}

// einfahren liest ein Setpaket in den Bestand.
func (d *Dienst) einfahren(schreiben http.ResponseWriter, anfrage *http.Request) {
	var rumpf struct {
		Pfad string `json:"pfad"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	setID, karten, befunde, fehler := d.naehte.Ausfuhr.Einfahren(anfrage.Context(), rumpf.Pfad)

	teilstand := map[string]any{
		"setID": strconv.FormatInt(setID, 10), "karten": karten,
		"befunde": ohneLeeretexte(befunde),
	}

	// **Der Teilerfolg geht mit dem Fehler hinaus** — Befund 1 des Nachlaufs
	// über Phase 4.4 (#1291). `Importeur.Set` hängt einen Schließfehler des
	// Archivs an ein **geglücktes** Einfahren an; das Set steht dann im
	// Bestand (§7 Regel 3, #1041).
	if fehler != nil {
		fehlerhaftMit(schreiben, fehler, teilstand)

		return
	}

	ja(schreiben, http.StatusCreated, teilstand)
}

// sichern schreibt ein Set an einen Ort, an dem nichts liegt.
func (d *Dienst) sichern(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gut := kennung(schreiben, anfrage, "setID")
	if !gut {
		return
	}

	var rumpf struct {
		Ziel string `json:"ziel"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	befunde, fehler := d.naehte.Ausfuhr.Sichern(anfrage.Context(), setID, rumpf.Ziel)

	teilstand := map[string]any{"ziel": rumpf.Ziel, "befunde": ohneLeeretexte(befunde)}

	// **Der Teilerfolg geht mit dem Fehler hinaus** (#1291, Nachlauf):
	// scheitert nur das Setzen des Arbeitsorts, **liegt die Datei schon am
	// Ziel**.
	if fehler != nil {
		fehlerhaftMit(schreiben, fehler, teilstand)

		return
	}

	ja(schreiben, http.StatusCreated, teilstand)
}

// bogen legt die Karten eines Sets als Druckbogen (#1291).
//
// **Am Set und nicht an der Karte** — und deshalb `POST /sets/{setID}/bogen`,
// dieselbe Gestalt wie die Sicherung daneben.
//
// **Ohne Papierwahl und ohne Auswahl**: A4, neun Karten je Seite, 300 dpi,
// mit Schnittmarken. Am Fenster stellt die Seite beides; ein Dienst, der über
// die Leitung angesprochen wird, bekommt den Bogen, den man drucken kann.
// Einordnung nach §16: **erweiternd** — ein Weg kommt hinzu, keiner ändert
// seine Bedeutung, und Papier und Auswahl lassen sich später als wahlfreie
// Felder ergänzen ([0033](../../docs/entscheidungen/0033-was-ueber-fassungen-hinweg-zugesagt-wird.md)).
func (d *Dienst) bogen(schreiben http.ResponseWriter, anfrage *http.Request) {
	setID, gut := kennung(schreiben, anfrage, "setID")
	if !gut {
		return
	}

	var rumpf struct {
		Ziel string `json:"ziel"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	seiten, karten, befunde, fehler := d.naehte.Ausfuhr.Bogen(anfrage.Context(), setID, rumpf.Ziel)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusCreated, map[string]any{
		"ziel": rumpf.Ziel, "seiten": seiten, "karten": karten, "befunde": befunde,
	})
}

// zeichnen rastert eine Karte an einen Ort, an dem nichts liegt.
func (d *Dienst) zeichnen(schreiben http.ResponseWriter, anfrage *http.Request) {
	karteID, gut := kennung(schreiben, anfrage, "karteID")
	if !gut {
		return
	}

	var rumpf struct {
		Ziel string `json:"ziel"`

		// Guete ist **wahlfrei**: leer heißt „wie MSE2 zeichnet".
		//
		// **Einordnung nach §16: erweiternd** — ein Aufruf ohne `guete` tut
		// genau das, was er vorher tat.
		Guete string `json:"guete"`
	}

	if !gelesen(schreiben, anfrage, &rumpf) {
		return
	}

	breite, hoehe, befunde, fehler := d.naehte.Ausfuhr.Zeichnen(anfrage.Context(),
		karteID, rumpf.Ziel, rumpf.Guete)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	// **Die Befunde gehören in die Antwort** (#1137): die Datei entsteht auch
	// mit ihnen, und wer sie nicht bekommt, hält sie für vollständig.
	//
	// **Einordnung nach §16: erweiternd** — ein Feld kommt hinzu, kein Aufruf
	// ändert seine Bedeutung.
	ja(schreiben, http.StatusCreated, map[string]any{
		"ziel": rumpf.Ziel, "breite": breite, "hoehe": hoehe,
		"befunde": befunde,
	})
}

// rumpffehler sagt, was am Rumpf nicht stimmt — ohne zu sagen, dass dieses
// Programm in Go geschrieben ist.
//
// # Warum die Meldung nicht durchgereicht wird
//
// Der Entzifferer schreibt bei einem Typfehler `json: cannot unmarshal number
// into Go struct field .pfad of type string`. Das ist die Struktur dieses
// Programms in einer Antwort nach außen — der Zweitblick über Phase U hat es
// angestrichen, und §8 verbietet dieselbe Sorte Durchreichung schon für
// SQLite.
//
// **Was bleibt, ist die Hilfe**: welches Feld nicht stimmt und was dort
// erwartet wird. Beides steht in `*json.UnmarshalTypeError`, in Feldern, die
// nichts über Go verraten — außer dem Typnamen, und der wird übersetzt.
func rumpffehler(fehler error) string {
	var typfehler *json.UnmarshalTypeError
	if errors.As(fehler, &typfehler) {
		return fmt.Sprintf("das Feld %q erwartet %s",
			strings.TrimPrefix(typfehler.Field, "."), alsDeutscherTyp(typfehler.Type.String()))
	}

	var unbekannt *json.SyntaxError
	if errors.As(fehler, &unbekannt) {
		return fmt.Sprintf("der Rumpf ist kein gültiges JSON (Stelle %d)", unbekannt.Offset)
	}

	// **Der Rest bleibt allgemein.** `DisallowUnknownFields` schreibt
	// `json: unknown field "x"` — der Feldname stammt vom Aufrufer, aber das
	// Wort davor nicht.
	if feld := unbekanntesFeld.FindStringSubmatch(fehler.Error()); feld != nil {
		return fmt.Sprintf("das Feld %q kennt dieser Dienst nicht", feld[1])
	}

	return "der Rumpf ist kein gültiges JSON"
}

// unbekanntesFeld liest den Feldnamen aus `json: unknown field "x"`.
var unbekanntesFeld = regexp.MustCompile(`unknown field "([^"]*)"`)

// alsDeutscherTyp übersetzt die Go-Typnamen, die in einem Rumpf vorkommen.
//
// **Genau einer**, denn jedes Rumpffeld dieses Dienstes ist heute eine
// Zeichenkette. Ein Zweig für `int` und `bool` erreichte nichts, und §3 nimmt
// eine unerreichbare Verzweigung nicht als Ausnahme an. Kommt ein Zahlenfeld
// dazu, sieht der Test dafür „einen anderen Wert" und der Zweig entsteht mit
// ihm.
//
// **Ein unbekannter Name wird nicht durchgereicht**, sondern verallgemeinert —
// sonst stünde bei einem künftigen Feld wieder ein Go-Typ in der Antwort.
func alsDeutscherTyp(gonamen string) string {
	if gonamen == "string" {
		return "eine Zeichenkette"
	}

	return "einen anderen Wert"
}

// kennung liest eine Zahl aus dem Weg.
func kennung(schreiben http.ResponseWriter, anfrage *http.Request, name string) (int64, bool) {
	roh := anfrage.PathValue(name)

	zahl, fehler := strconv.ParseInt(roh, 10, 64)
	if fehler != nil || zahl < 1 {
		nein(schreiben, http.StatusBadRequest, "validierung",
			fmt.Sprintf("%s muss eine Kennung sein, nicht %q", name, roh))

		return 0, false
	}

	return zahl, true
}

// abfragezahl liest eine Zahl aus der Abfrage.
//
// **Neben [kennung] und nicht statt ihrer**: die eine liest aus dem **Weg**,
// die andere aus der **Abfrage**, und die Meldungen sagen es. Wer beide
// zusammenlegte, müsste dem Aufrufer erklären, wo er suchen soll.
func abfragezahl(schreiben http.ResponseWriter, anfrage *http.Request, name string) (
	int64, bool,
) {
	roh := anfrage.URL.Query().Get(name)

	wert, fehler := strconv.ParseInt(roh, 10, 64)
	if fehler != nil || wert < 1 {
		nein(schreiben, http.StatusBadRequest, "validierung",
			fmt.Sprintf("%s muss als Abfragewert dabeistehen, nicht %q", name, roh))

		return 0, false
	}

	return wert, true
}

// gelesen liest den Rumpf einer Anfrage.
//
// **Mit Obergrenze.** Ein Rumpf ohne Grenze ist ein Weg, den Arbeitsspeicher
// dieses Prozesses zu füllen — und der Aufrufer ist per Annahme nicht
// vertrauenswürdig, sonst bräuchte es keine Marke.
// rumpfgrenze ist, was ein Rumpf höchstens wiegt — ein Mebibyte; ein Pfad ist
// tausendfach kürzer. Der Pfadvertrag liest bis hierher und eins mehr.
const rumpfgrenze = 1 << 20

func gelesen(schreiben http.ResponseWriter, anfrage *http.Request, ziel any) bool {
	entzifferer := json.NewDecoder(http.MaxBytesReader(schreiben, anfrage.Body, rumpfgrenze))

	// **Unbekannte Felder werden abgewiesen.** Ein Tippfehler im Feldnamen
	// führte sonst zu einem leeren Wert statt zu einer Meldung — und der
	// Aufrufer sähe ein 200 für etwas, das er nicht gemeint hat.
	entzifferer.DisallowUnknownFields()

	if fehler := entzifferer.Decode(ziel); fehler != nil {
		var zuGross *http.MaxBytesError

		if errors.As(fehler, &zuGross) {
			nein(schreiben, http.StatusRequestEntityTooLarge, "validierung",
				"der Rumpf ist zu groß")

			return false
		}

		nein(schreiben, http.StatusBadRequest, "validierung", rumpffehler(fehler))

		return false
	}

	return true
}
