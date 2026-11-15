// **Was es nicht gibt, ist an jeder Fläche nicht gefunden** — C04
// (#2229).
//
// Die Flächenparität (`docs/FLAECHENPARITAET.md`) zählte bei 13 von 32
// Klassen einen Lauf über die Fehlerklasse. Dieser legt für die übrigen den
// einfachsten Fehlerfall nebeneinander: eine Kennung, die es nicht gibt, an
// Kommandozeile, MCP, Dienst und Fenster — dieselbe Klasse, derselbe Kern.
// **Wo eine Fläche die Fähigkeit nicht hat**, nennt der Fall den Grund aus
// `FAEHIGKEITEN.md`; wo die einfachste Abweisung keine Kennung ist (eine
// Kostengrenze für einen Zeitraum, den es nicht gibt), nennt er seine Klasse.
package flaechen_test

import (
	"net/http"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

const fremdeKennung = "4711"

func TestWasEsNichtGibtIstAnJederFlaecheNichtGefunden(t *testing.T) {
	// **Mit umgelenktem Scryfall**: ein Preis für einen Druck, den es nicht
	// gibt, fragt sonst den echten Dienst.
	umgebung := append(mitEigenerSaat(t), app.Scryfallumgebung+"="+scryfallgegenstelle(t).URL)
	setID := eingefahren(t, umgebung)
	an := starteDienst(t, umgebung)
	bild, sammlung := probebild(t), filepath.Join(t.TempDir(), "gibtsnicht.json")

	faelle := []struct {
		name, kern, klasse  string
		cli                 []string
		werkzeug            string
		eingabe             map[string]any
		ohneMCP             string
		methode, weg, rumpf string
		fenster             func(*app.Bindungen) *app.Fehler
	}{
		{
			name: "ein Format, das es nicht gibt", kern: "dieses Format gibt es nicht",
			cli:      []string{"format", "loeschen", fremdeKennung},
			werkzeug: "schreiben.format.loeschen", eingabe: map[string]any{"formatID": 4711},
			methode: http.MethodDelete, weg: "/formate/" + fremdeKennung,
			fenster: func(b *app.Bindungen) *app.Fehler { return b.FormatLoeschen(t.Context(), fremdeKennung).Fehler },
		},
		{
			name: "ein Turnier, das es nicht gibt", kern: "ein Turnier mit dieser Kennung gibt es nicht",
			cli:      []string{"turnier", "loeschen", fremdeKennung},
			werkzeug: "schreiben.turnier.loeschen", eingabe: map[string]any{"turnierID": 4711},
			methode: http.MethodDelete, weg: "/turniere/" + fremdeKennung,
			fenster: func(b *app.Bindungen) *app.Fehler { return b.TurnierLoeschen(t.Context(), fremdeKennung).Fehler },
		},
		{
			name: "ein Token, das es nicht gibt", kern: fremdeKennung,
			cli:      []string{"token-loeschen", fremdeKennung},
			werkzeug: "schreiben.token.loeschen", eingabe: map[string]any{"tokenID": fremdeKennung},
			methode: http.MethodDelete, weg: "/token/" + fremdeKennung,
			fenster: func(b *app.Bindungen) *app.Fehler { return b.TokenLoeschen(t.Context(), fremdeKennung).Fehler },
		},
		{
			name: "ein Sprachpaket, das es nicht gibt", kern: "gibtsnicht",
			cli:      []string{"sprachpaket", "loeschen", "gibtsnicht"},
			werkzeug: "schreiben.sprachpaket.loeschen", eingabe: map[string]any{"kennung": "gibtsnicht"},
			methode: http.MethodDelete, weg: "/sprachpakete/gibtsnicht",
			fenster: func(b *app.Bindungen) *app.Fehler { return b.SprachpaketLoeschen(t.Context(), "gibtsnicht").Fehler },
		},
		{
			name: "ein Symbol der Bibliothek, das es nicht gibt", kern: fremdeKennung,
			cli:      []string{"symbolbibliothek", "loeschen", fremdeKennung},
			werkzeug: "schreiben.symbolbibliothek.loeschen", eingabe: map[string]any{"id": 4711, "bestaetigt": true},
			methode: http.MethodDelete, weg: "/symbole/" + fremdeKennung + "?bestaetigt=true",
			fenster: func(b *app.Bindungen) *app.Fehler {
				return b.BibliothekssymbolLoeschen(t.Context(), fremdeKennung, true).Fehler
			},
		},
		{
			name: "ein Schlüsselwort, das es nicht gibt", kern: fremdeKennung,
			cli:      []string{"schluesselwort", setID, "loeschen", fremdeKennung},
			werkzeug: "schreiben.schluesselwort.loeschen", eingabe: map[string]any{"setID": zahl(t, setID), "wortID": 4711},
			methode: http.MethodDelete, weg: "/sets/" + setID + "/schluesselwoerter/" + fremdeKennung,
			fenster: func(b *app.Bindungen) *app.Fehler {
				return b.SchluesselwortLoeschen(t.Context(), setID, fremdeKennung).Fehler
			},
		},
		{
			name: "eine Partie, die es nicht gibt", kern: "4711",
			cli:      []string{"partie-loeschen", "00000000000000000000000000004711"},
			werkzeug: "schreiben.partie.loeschen", eingabe: map[string]any{"partieID": "00000000000000000000000000004711"},
			methode: http.MethodDelete, weg: "/partien/00000000000000000000000000004711",
			fenster: func(b *app.Bindungen) *app.Fehler {
				return b.PartieLoeschen(t.Context(), "00000000000000000000000000004711").Fehler
			},
		},
		{
			name: "ein Nutzer, den es nicht gibt", kern: fremdeKennung,
			cli:      []string{"nutzer-loeschen", fremdeKennung, "Niemand"},
			werkzeug: "schreiben.nutzer.loeschen", eingabe: map[string]any{"nutzerID": fremdeKennung, "name": "Niemand"},
			methode: http.MethodDelete, weg: "/nutzer/" + fremdeKennung + "?name=Niemand",
			fenster: func(b *app.Bindungen) *app.Fehler {
				return b.NutzerLoeschen(t.Context(), fremdeKennung, "Niemand").Fehler
			},
		},
		{
			name: "ein Profil, das es nicht gibt", kern: fremdeKennung,
			cli:      []string{"profil-loeschen", fremdeKennung, "Niemand"},
			werkzeug: "schreiben.profil.loeschen", eingabe: map[string]any{"profilID": fremdeKennung, "name": "Niemand"},
			methode: http.MethodDelete, weg: "/profile/" + fremdeKennung + "?name=Niemand",
			fenster: func(b *app.Bindungen) *app.Fehler {
				return b.ProfilLoeschen(t.Context(), fremdeKennung, "Niemand").Fehler
			},
		},
		{
			name: "eine Autoersetzung, die es nicht gibt", kern: "gibtsnicht",
			cli:      []string{"autoersetzung", setID, "loeschen", "gibtsnicht"},
			werkzeug: "schreiben.autoersetzung.loeschen", eingabe: map[string]any{"setID": zahl(t, setID), "muster": "gibtsnicht"},
			methode: http.MethodDelete, weg: "/sets/" + setID + "/autoersetzungen/gibtsnicht",
			fenster: func(b *app.Bindungen) *app.Fehler {
				return b.AutoersetzungLoeschen(t.Context(), setID, "gibtsnicht").Fehler
			},
		},
		{
			name: "eine Rücknahme in einem Set, das es nicht gibt", kern: fremdeKennung,
			cli:      []string{"zuruecknehmen", fremdeKennung},
			werkzeug: "schreiben.zuruecknehmen", eingabe: map[string]any{"setID": 4711},
			methode: http.MethodPost, weg: "/sets/" + fremdeKennung + "/zuruecknehmen",
			fenster: func(b *app.Bindungen) *app.Fehler { return b.Zuruecknehmen(t.Context(), fremdeKennung, "").Fehler },
		},
		{
			name: "ein Kontakt, den es nicht gibt", kern: "gibtsnicht",
			cli:      []string{"kontakt-sperren", "gibtsnicht"},
			werkzeug: "schreiben.peerkontakt.sperren", eingabe: map[string]any{"kennung": "gibtsnicht"},
			methode: http.MethodPut, weg: "/peer/kontakte/gibtsnicht/sperre",
			fenster: func(b *app.Bindungen) *app.Fehler { return b.PeerkontaktSperren(t.Context(), "gibtsnicht").Fehler },
		},
		{
			name: "eine Marke an einer Karte, die es nicht gibt", kern: fremdeKennung,
			cli:      []string{"marke", setID, fremdeKennung, "Fertig"},
			werkzeug: "schreiben.marke", eingabe: map[string]any{"setID": zahl(t, setID), "karteID": 4711, "marke": "Fertig"},
			methode: http.MethodPut, weg: "/sets/" + setID + "/karten/" + fremdeKennung + "/marken/Fertig",
			fenster: func(b *app.Bindungen) *app.Fehler {
				return b.MarkeSetzen(t.Context(), setID, fremdeKennung, "Fertig").Fehler
			},
		},
		{
			name: "ein Kartenbild für eine Karte, die es nicht gibt", kern: fremdeKennung,
			cli:      []string{"bild-setzen", setID, fremdeKennung, "kunst", bild},
			werkzeug: "schreiben.karte.bild", eingabe: map[string]any{"setID": zahl(t, setID), "karteID": 4711, "feld": "kunst", "pfad": bild},
			methode: http.MethodPut, weg: "/sets/" + setID + "/karten/" + fremdeKennung + "/felder/kunst/bild",
			rumpf: `{"pfad":` + alsJSONText(t, bild) + `}`,
			fenster: func(b *app.Bindungen) *app.Fehler {
				return b.BildEinsetzen(t.Context(), setID, fremdeKennung, "kunst", bild, app.AuswahlDTO{}).Fehler
			},
		},
		{
			name: "ein Handel, den es nicht gibt", kern: "den Handelsvorgang gibt es nicht",
			cli:     []string{"handel", "verwerfen", "gibtsnicht"},
			ohneMCP: "MCP schreibt keinen Handel (FAEHIGKEITEN.md)",
			methode: http.MethodDelete, weg: "/handel/gibtsnicht/entwurf",
			fenster: func(b *app.Bindungen) *app.Fehler { return b.HandelEntwurfLoeschen(t.Context(), "gibtsnicht").Fehler },
		},
		{
			name: "ein Preis für einen Druck, den es nicht gibt", kern: "Scryfall kennt diesen Druck nicht",
			cli:      []string{"preise-holen", "gibtsnicht"},
			werkzeug: "schreiben.preiseholen", eingabe: map[string]any{"printing": "gibtsnicht"},
			methode: http.MethodPost, weg: "/drucke/gibtsnicht/preise",
			fenster: func(b *app.Bindungen) *app.Fehler { return b.PreiseHolen(t.Context(), "gibtsnicht").Fehler },
		},
		{
			// **Die Laufzeit vor der Datei**: jede Fläche fragt zuerst, ob
			// überhaupt eingebettet werden darf — und sagt dasselbe.
			name: "ein Index, solange die Laufzeit aus ist", kern: "die KI-Laufzeit ist abgeschaltet", klasse: klasseErlaubnis,
			cli:     []string{"indexaufbau", probeeinbetter, sammlung, "prf-1", "2026-01-01"},
			ohneMCP: "MCP baut keinen Index (FAEHIGKEITEN.md)",
			methode: http.MethodPost, weg: "/index/aufbau",
			rumpf: string(mussJSON(t, map[string]string{
				"modell": probeeinbetter, "quelle": sammlung, "revision": "prf-1", "quelldatum": "2026-01-01",
			})),
			fenster: func(b *app.Bindungen) *app.Fehler {
				return b.IndexAufbauen(t.Context(), probeeinbetter, sammlung, "prf-1", "2026-01-01", false).Fehler
			},
		},
		{
			name: "ein Faktor eines Nutzers, den es nicht gibt", kern: "dieser Nutzer hat keinen Faktor der Art",
			cli:      []string{"faktor-loeschen", fremdeKennung, "totp"},
			werkzeug: "schreiben.faktor.loeschen", eingabe: map[string]any{"nutzerID": fremdeKennung, "art": "totp"},
			methode: http.MethodDelete, weg: "/nutzer/" + fremdeKennung + "/faktoren/totp",
			fenster: func(b *app.Bindungen) *app.Fehler {
				return b.FaktorEntfernen(t.Context(), fremdeKennung, "totp").Fehler
			},
		},
		{
			name: "eine Kostengrenze für einen Zeitraum, den es nicht gibt", kern: "jahrzehnt", klasse: klasseValidierung,
			cli:     []string{"kostenbudget", "anbieter", "jahrzehnt", "50000", "EUR"},
			ohneMCP: "MCP setzt keine Grenze — die Grenze ist die Freigabe (FAEHIGKEITEN.md)",
			methode: http.MethodPut, weg: "/kosten/anbieter/budget",
			rumpf: `{"zeitraum":"jahrzehnt","wert":50000,"waehrung":"EUR"}`,
			fenster: func(b *app.Bindungen) *app.Fehler {
				return b.KostenbudgetSetzen(t.Context(), "anbieter", "jahrzehnt", 50000, "EUR").Fehler
			},
		},
	}

	for _, fall := range faelle {
		klasse := fall.klasse
		if klasse == "" {
			klasse = klasseNichtGefunden
		}

		var mcp func() (string, bool)
		if fall.werkzeug != "" {
			mcp = func() (string, bool) { return mcpRuf(t, umgebung, fall.werkzeug, fall.eingabe) }
		}

		abweisungGleich(t, fall.name, abweisung{
			klasse: klasse,
			kern:   fall.kern,
			cli: func() (string, int) {
				aus, fehler, code := kommandozeile(t, umgebung, fall.cli...)

				return aus + fehler, code
			},
			mcp:     mcp,
			ohneMCP: fall.ohneMCP,
			rest:    func() (int, string) { return an.ruf(t, fall.methode, fall.weg, fall.rumpf) },
			fenster: func() *app.Fehler { return fall.fenster(fensterAus(t, umgebung)) },
		})
	}
}
