package rest

// **Die Wege zu den Aufträgen** — FMT-03.AK-02
// (#1692).
//
// # Die Bitte und was aus ihr wird
//
// Wer einen der [auftragswege] mit `Prefer: respond-async` ruft (RFC 7240,
// Abschnitt 4.1), bekommt `202 Accepted`, `Preference-Applied` und den Ort
// des Auftrags. **Der Weg selbst bleibt derselbe**: er läuft mit derselben
// Anfrage hinter denselben Wachen, nur nebenher, und seine Antwort wird
// aufgenommen statt geschickt. **Ohne die Bitte antwortet jeder Weg wie
// bisher**, und ein Weg, der keiner der langen ist, übergeht sie — eine
// Bitte ist kein Befehl. Was das Buch hält, steht in `auftraege.go`.

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Auftragsnaht führt die Aufträge des Dienstes — [Auftragsbuch] erfüllt sie.
type Auftragsnaht interface {
	Annehmen(ctx context.Context, eingabe Auftragseingabe, lauf Auftragslauf) (AuftragDTO, error)
	Wiederholen(ctx context.Context, id string, lauf Auftragslauf) (AuftragDTO, error)
	Auftraege(ctx context.Context) ([]AuftragDTO, error)
	Auftrag(ctx context.Context, id string) (AuftragDTO, error)
	Ergebnis(ctx context.Context, id string) (Auftragsergebnis, error)
	Abbrechen(ctx context.Context, id string) (lief bool, fehler error)
}

// auftragswege sind die Wege, die als Auftrag laufen dürfen — die langen
// Datei-Läufe: einlesen, sichern, zeichnen, legen, ausführen, weitergeben —
// und **die Prüfung eines Sets** (PERF02, #2352; Zweitblick über Phase 3,
// Befund 11): sie zeichnet jede Karte und kann die Schreibfrist überdauern.
var auftragswege = []string{
	"POST /sets/{setID}/probleme",
	"POST /sets",
	"POST /sets/{setID}/sicherung",
	"POST /karten/{karteID}/bild",
	"POST /sets/{setID}/bogen",
	"POST /sets/{setID}/vorlagen",
	"POST /mtgjson",
	"POST /sets/{setID}/mtgjson",
	"POST /pakete/weitergabe",
	"POST /pakete/eingelesen",
}

// Auftragswege nennt die Wege, die als Auftrag laufen dürfen — eine Kopie.
func Auftragswege() []string { return slices.Clone(auftragswege) }

// MitFortschrittsnaht sagt dem Dienst, wie ein Melder in den Kontext kommt —
// ohne sie steht an einem Auftrag kein Fortschritt.
func (d *Dienst) MitFortschrittsnaht(
	naht func(context.Context, domain.Fortschrittsmelder) context.Context,
) *Dienst {
	d.fortschritt = naht

	return d
}

// auftragswegeEintragen trägt die fünf Wege ein.
func (d *Dienst) auftragswegeEintragen(weg *http.ServeMux) {
	weg.HandleFunc("GET /auftraege", d.auftraegeZeigen)
	weg.HandleFunc("GET /auftraege/{auftragID}", d.auftragZeigen)
	weg.HandleFunc("GET /auftraege/{auftragID}/ergebnis", d.auftragsergebnis)
	weg.HandleFunc("DELETE /auftraege/{auftragID}", d.auftragAbbrechen)
	weg.HandleFunc("POST /auftraege/{auftragID}/wiederholung", func(schreiben http.ResponseWriter, anfrage *http.Request) {
		d.angenommen(schreiben)(d.naehte.Auftraege.Wiederholen(
			anfrage.Context(), anfrage.PathValue("auftragID"), d.auftragslauf(d.mitAusgang(weg))))
	})
}

// asynchronErbeten sagt, ob die Anfrage um `respond-async` bittet.
func asynchronErbeten(anfrage *http.Request) bool {
	for _, kopf := range anfrage.Header.Values("Prefer") {
		for wunsch := range strings.SplitSeq(kopf, ",") {
			name, _, _ := strings.Cut(wunsch, ";")
			if strings.EqualFold(strings.TrimSpace(name), "respond-async") {
				return true
			}
		}
	}

	return false
}

// mitAuftrag nimmt die erbetenen Aufträge an und reicht alles andere weiter.
func (d *Dienst) mitAuftrag(weg *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(schreiben http.ResponseWriter, anfrage *http.Request) {
		// **Der Ausgang gehört in den Lauf** (`dateiuebertragung.go`): als
		// Auftrag bekommt jede Wiederholung einen frischen Ort.
		innen := d.mitAusgang(weg)

		_, muster := weg.Handler(anfrage)
		if d.naehte.Auftraege == nil || !asynchronErbeten(anfrage) || !slices.Contains(auftragswege, muster) {
			innen.ServeHTTP(schreiben, anfrage)

			return
		}

		// **Dieselbe Grenze wie `gelesen`** — der Weg nähme ohnehin nicht mehr.
		rumpf, fehler := io.ReadAll(io.LimitReader(anfrage.Body, rumpfgrenze+1))
		if fehler != nil || len(rumpf) > rumpfgrenze {
			nein(schreiben, http.StatusRequestEntityTooLarge, "validierung", "der Rumpf ist zu groß")

			return
		}

		d.angenommen(schreiben)(d.naehte.Auftraege.Annehmen(anfrage.Context(), Auftragseingabe{
			Weg: muster, Methode: anfrage.Method, Pfad: anfrage.URL.RequestURI(),
			Art: anfrage.Header.Get("Content-Type"), Rumpf: rumpf,
		}, d.auftragslauf(innen)))
	})
}

// angenommen sagt, wo ein angenommener Auftrag steht — oder warum keiner.
func (d *Dienst) angenommen(schreiben http.ResponseWriter) func(AuftragDTO, error) {
	return func(auftrag AuftragDTO, fehler error) {
		switch {
		case istZuViele(fehler):
			schreiben.Header().Set("Retry-After", "5")
			nein(schreiben, http.StatusTooManyRequests, "auftraege", fehler.Error())

		case fehler != nil:
			fehlerhaft(schreiben, fehler)

		default:
			schreiben.Header().Set("Location", "/auftraege/"+auftrag.ID)
			schreiben.Header().Set("Preference-Applied", "respond-async")
			ja(schreiben, http.StatusAccepted, map[string]any{"auftrag": auftrag})
		}
	}
}

// auftragslauf fährt eine Eingabe durch denselben Weg und nimmt die Antwort auf.
//
// **Der Pfadvertrag urteilt hier ein zweites Mal** — für den ersten Lauf
// eine Wiederholung, für eine Wiederholung die erste Prüfung seit damals: der
// Vertrag kann sich geändert haben.
func (d *Dienst) auftragslauf(weg http.Handler) Auftragslauf {
	return func(ctx context.Context, eingabe Auftragseingabe, melde domain.Fortschrittsmelder) Auftragsergebnis {
		if d.fortschritt != nil {
			ctx = d.fortschritt(ctx, melde)
		}

		antwort := neueAufnahme(d.grenzen.Antwort)

		// **Der Lauf holt seine eigene Freigabe** — die der Anfrage, die ihn
		// annahm, ist längst zurückgegeben (Zweitblick über Phase 2, Befund 2).
		frei, verwehrt := d.freigabeFuerLauf(ctx)
		if verwehrt != nil {
			fehlerhaft(antwort, verwehrt)

			return antwort.ergebnis()
		}

		defer frei()

		// **Ohne Fehlerweg**: Methode und Pfad kamen aus einer Anfrage, die
		// der Server schon zerlegt hat.
		anfrage, _ := http.NewRequestWithContext(ctx, eingabe.Methode, eingabe.Pfad, //nolint:errcheck // siehe darüber
			bytes.NewReader(eingabe.Rumpf))
		anfrage.Header.Set("Content-Type", eingabe.Art)

		if fehler := d.pfadeVerwehrt(anfrage, eingabe.Weg); fehler != nil {
			fehlerhaft(antwort, fehler)
		} else {
			weg.ServeHTTP(antwort, anfrage)
		}

		return antwort.ergebnis()
	}
}

// AuftragDTO ist ein Auftrag, wie der Dienst ihn zeigt.
type AuftragDTO struct {
	ID              string         `json:"id"`
	Weg             string         `json:"weg"`
	Pfad            string         `json:"pfad"`
	Stand           string         `json:"stand"`
	Fortschritt     FortschrittDTO `json:"fortschritt"`
	Begonnen        string         `json:"begonnen"`
	Beendet         string         `json:"beendet"`
	Status          int            `json:"status"`
	WiederholungVon string         `json:"wiederholungVon"`
}

// FortschrittDTO ist ein Zwischenstand samt seinem Satz.
type FortschrittDTO struct {
	Erledigt int    `json:"erledigt"`
	Gesamt   int    `json:"gesamt"`
	Einheit  string `json:"einheit"`
	Satz     string `json:"satz"`
}

func auftragAlsDTO(auftrag restauftrag) AuftragDTO {
	dto := AuftragDTO{
		ID: auftrag.id, Weg: auftrag.eingabe.Weg, Pfad: auftrag.eingabe.Pfad, Stand: auftrag.stand,
		Fortschritt: FortschrittDTO{
			Erledigt: auftrag.fortschritt.Erledigt, Gesamt: auftrag.fortschritt.Gesamt,
			Einheit: auftrag.fortschritt.Einheit, Satz: auftrag.fortschritt.Satz(),
		},
		Begonnen:        auftrag.begonnen.UTC().Format(time.RFC3339),
		WiederholungVon: auftrag.wiederholungVon,
	}

	if auftrag.ergebnis != nil {
		dto.Beendet = auftrag.beendet.UTC().Format(time.RFC3339)
		dto.Status = auftrag.ergebnis.Status
	}

	return dto
}

func (d *Dienst) auftraegeZeigen(schreiben http.ResponseWriter, anfrage *http.Request) {
	liste, fehler := d.naehte.Auftraege.Auftraege(anfrage.Context())
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"auftraege": liste})
}

func (d *Dienst) auftragZeigen(schreiben http.ResponseWriter, anfrage *http.Request) {
	auftrag, fehler := d.naehte.Auftraege.Auftrag(anfrage.Context(), anfrage.PathValue("auftragID"))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"auftrag": auftrag})
}

// auftragsergebnis gibt die Antwort, die der Weg gleich gegeben hätte.
func (d *Dienst) auftragsergebnis(schreiben http.ResponseWriter, anfrage *http.Request) {
	ergebnis, fehler := d.naehte.Auftraege.Ergebnis(anfrage.Context(), anfrage.PathValue("auftragID"))
	if errors.Is(fehler, domain.ErrConflict) {
		nein(schreiben, http.StatusConflict, "laeuft", "der Auftrag läuft noch")

		return
	} else if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	schreiben.Header().Set("Content-Type", ergebnis.Art)

	if ergebnis.Ort != "" {
		schreiben.Header().Set("Location", ergebnis.Ort)
	}

	schreiben.WriteHeader(ergebnis.Status)

	_, _ = schreiben.Write(ergebnis.Rumpf) //nolint:errcheck // die Verbindung ist fort
}

// auftragAbbrechen hält einen laufenden Auftrag an oder vergisst einen beendeten.
func (d *Dienst) auftragAbbrechen(schreiben http.ResponseWriter, anfrage *http.Request) {
	lief, fehler := d.naehte.Auftraege.Abbrechen(anfrage.Context(), anfrage.PathValue("auftragID"))

	switch {
	case fehler != nil:
		fehlerhaft(schreiben, fehler)
	case lief:
		schreiben.WriteHeader(http.StatusAccepted)
	default:
		schreiben.WriteHeader(http.StatusNoContent)
	}
}

// aufnahme nimmt eine Antwort auf, statt sie zu schicken — mit derselben
// Grenze wie eine geschickte.
type aufnahme struct {
	kopf   http.Header
	status int
	rumpf  bytes.Buffer
	rest   int64
}

func neueAufnahme(grenze int64) *aufnahme {
	return &aufnahme{kopf: http.Header{}, status: http.StatusOK, rest: grenze}
}

func (a *aufnahme) Header() http.Header { return a.kopf }

func (a *aufnahme) WriteHeader(status int) { a.status = status }

// Write nimmt auf — und bricht wie [begrenzt] ab, wenn die Grenze reißt.
func (a *aufnahme) Write(teil []byte) (int, error) {
	if int64(len(teil)) > a.rest {
		panic(http.ErrAbortHandler)
	}

	a.rest -= int64(len(teil))

	return a.rumpf.Write(teil)
}

func (a *aufnahme) ergebnis() Auftragsergebnis {
	return Auftragsergebnis{
		Status: a.status, Art: a.kopf.Get("Content-Type"), Ort: a.kopf.Get("Location"), Rumpf: a.rumpf.Bytes(),
	}
}

type principalschluessel struct{}

// mitPrincipal legt den Principal eines Rufs in seinen Kontext.
func mitPrincipal(ctx context.Context, principal string) context.Context {
	return context.WithValue(ctx, principalschluessel{}, principal)
}

// principalAus liest ihn — leer, wo keine Marke geprüft hat.
func principalAus(ctx context.Context) string {
	if principal, da := ctx.Value(principalschluessel{}).(string); da {
		return principal
	}

	return ""
}
