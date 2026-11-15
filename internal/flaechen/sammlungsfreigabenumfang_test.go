package flaechen_test

// **Dieselbe Sicht mit ihren Angaben, an jeder Fläche** — UXR-F18
// (#2033),
// §16.
//
// # Der Befund
//
// Das Fenster bot „Ausgewählte Drucke" an, ohne Drucke wählen zu lassen, und
// schickte bei „Dubletten" immer die Behaltemenge 0. Die Kommandozeile setzt
// beides (`mse kontakt-freigeben <kennung> <zweck> <umfang> [behalten|druck…]`),
// der Dienst über `auswahl` und `behalten` im Rumpf. Die Bindung des Fensters
// trug beides schon — es fehlte die Bedienung.
//
// # Was hier gemessen wird
//
// **Dass eine Auswahl aus zwei Drucken und eine Dublettensicht mit
// Behaltemenge über Fenster, Kommandozeile und Dienst dieselbe Freigabe
// ergeben** — gelesen wird jede von allen drei Flächen —, und dass sie
// **dasselbe zeigen**: die Vorschau („was das Gegenüber sähe") nennt dieselben
// Drucke in denselben Mengen.
//
// **Und dass die Ablehnung dieselbe ist**: eine Auswahl ohne Drucke gibt
// nichts frei, an jeder Fläche mit demselben Satz — am Fenster am Feld
// `auswahl`.
//
// # Wie die Fenster-Fläche hier fährt
//
// **Mit dem Rumpf, den die Oberfläche baut** (`Sichtfreigabe` in
// `frontend/src/apisichtfreigabe.ts`): als JSON in [app.SammlungsfreigabeDTO]
// gelesen, also durch die Struct-Marken, die kein Vitest-Lauf sieht (§3) — und
// dann durch die Methode, auf die die Bindung nur weiterleitet. Die Hülle
// selbst (Wails, `Bindungen`) fährt die Probe am Fenster.
//
// # Was hier **nicht** gemessen wird
//
// **MCP.** Dort gibt es kein Werkzeug, und das ist Absicht: eine Zustimmung,
// wer die Sammlung eines Menschen sieht, delegiert man nicht an ein Modell
// (`docs/FAEHIGKEITEN.md`). `internal/werkzeuge/faehigkeiten_test.go` hält
// beide Richtungen zwischen der Tabelle und dem Quelltext; eine zweite
// Prüfung daneben liefe beim nächsten Werkzeug auseinander (§10).

import (
	"encoding/json"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// Drei Drucke, die jemand hat — mit Mengen, an denen sich die beiden Umfänge
// unterscheiden: von `bolt` hat er drei, von `counter` eines, von `growth`
// zwei.
const (
	druckBolt    = "a1b2c3d4-0000-4000-8000-00000000000a"
	druckCounter = "a1b2c3d4-0000-4000-8000-00000000000b"
	druckGrowth  = "a1b2c3d4-0000-4000-8000-00000000000c"
)

// freigabenlesung ist, was jede Fläche über die erteilten Sichten sagt.
type freigabenlesung struct {
	Freigaben []struct {
		Zweck    string   `json:"zweck"`
		Umfang   string   `json:"umfang"`
		Auswahl  []string `json:"auswahl"`
		Behalten int      `json:"behalten"`
	} `json:"freigaben"`
}

// sichtlesung ist, was jede Fläche über die Vorschau sagt.
type sichtlesung struct {
	Zeilen []struct {
		Printing string `json:"printing"`
		Menge    int    `json:"menge"`
	} `json:"zeilen"`
}

// freigabenAus macht aus der Antwort einer Fläche eine ordentliche Liste.
//
// **Sortiert und als Zeilen**, damit zwei Flächen gleich sind, wenn sie
// dasselbe sagen — und nicht, wenn sie es in derselben Reihenfolge sagen.
func freigabenAus(t *testing.T, roh string) []string {
	t.Helper()

	var gelesen freigabenlesung
	if fehler := json.Unmarshal([]byte(roh), &gelesen); fehler != nil {
		t.Fatalf("keine Freigabenliste: %v\n%s", fehler, roh)
	}

	zeilen := make([]string, 0, len(gelesen.Freigaben))

	for _, freigabe := range gelesen.Freigaben {
		auswahl := slices.Clone(freigabe.Auswahl)
		slices.Sort(auswahl)

		zeilen = append(zeilen, freigabe.Zweck+"/"+freigabe.Umfang+"/"+
			strings.Join(auswahl, ",")+"/"+strconv.Itoa(freigabe.Behalten))
	}

	slices.Sort(zeilen)

	return zeilen
}

// sichtAus macht aus der Vorschau einer Fläche eine ordentliche Liste.
//
// **Ohne die Referenz**: sie ist sitzungsgebunden und je Abfrage eine andere.
func sichtAus(t *testing.T, roh string) []string {
	t.Helper()

	var gelesen sichtlesung
	if fehler := json.Unmarshal([]byte(roh), &gelesen); fehler != nil {
		t.Fatalf("keine Vorschau: %v\n%s", fehler, roh)
	}

	zeilen := make([]string, 0, len(gelesen.Zeilen))

	for _, zeile := range gelesen.Zeilen {
		zeilen = append(zeilen, zeile.Printing+"×"+strconv.Itoa(zeile.Menge))
	}

	slices.Sort(zeilen)

	return zeilen
}

// einGegenueber nimmt ein Gegenüber auf, das handeln und teilen darf.
func einGegenueber(t *testing.T, umgebung []string) string {
	t.Helper()

	fremd, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	kennung := string(fremd.Kennung())

	mussKommandozeile(t, umgebung, "kontakt-aufnehmen", kennung, "Annas Rechner")
	mussKommandozeile(t, umgebung, "kontakt-bestaetigen", kennung,
		fremd.Fingerabdruck())
	mussKommandozeile(t, umgebung, "zweck-erlauben", kennung, "handel")
	mussKommandozeile(t, umgebung, "zweck-erlauben", kennung, "teilen")

	return kennung
}

// fensterRumpf ist der Rumpf, den die Oberfläche für eine Freigabe baut —
// Feld für Feld wie `Sichtfreigabe` in `apisichtfreigabe.ts`.
func fensterRumpf(
	t *testing.T, kennung, zweck, umfang string, auswahl []string, behalten int,
) string {
	t.Helper()

	return string(mussJSON(t, map[string]any{
		"empfaenger": kennung, "umfang": umfang, "zweck": zweck,
		"auswahl": auswahl, "behalten": behalten, "ablauf": "", "erteilt": "",
	}))
}

// fensterFlaeche sind die Bindungen des Fensters, über demselben Bestand wie
// die Prozesse.
//
// **Über `NeueBindungen`, nicht über `app.App`** (Zweitblick über Phase C,
// Befund 3): hier stand `PeerflaecheUeber(db)` mit einem von Hand gesetzten
// Profil — ohne Tore, ohne Rufkontext und ohne den Weg, den Wails nimmt.
type fensterFlaeche struct {
	t *testing.T
	b *app.Bindungen
}

func fensterOeffnen(t *testing.T, umgebung []string) fensterFlaeche {
	t.Helper()

	return fensterFlaeche{t: t, b: fensterAus(t, umgebung)}
}

// erteilen schickt den Rumpf der Oberfläche durch das DTO der Bindung.
func (f fensterFlaeche) erteilen(rumpf string) app.SammlungsfreigabenAntwort {
	f.t.Helper()

	var dto app.SammlungsfreigabeDTO
	if fehler := json.Unmarshal([]byte(rumpf), &dto); fehler != nil {
		f.t.Fatalf("der Rumpf der Oberfläche passt nicht ins DTO: %v\n%s", fehler, rumpf)
	}

	return f.b.SammlungFreigeben(f.t.Context(), dto)
}

// lesen gibt, was die Bindung über ein Gegenüber sagt, als JSON.
func (f fensterFlaeche) lesen(kennung string) string {
	f.t.Helper()

	return string(mussJSON(f.t, f.b.Sammlungsfreigaben(f.t.Context(), kennung)))
}

// sicht gibt die Vorschau der Bindung als JSON.
func (f fensterFlaeche) sicht(kennung, zweck string) string {
	f.t.Helper()

	return string(mussJSON(f.t, f.b.Freigabesicht(f.t.Context(), kennung, zweck)))
}

// **Eine Auswahl aus zwei Drucken und Dubletten mit Behaltemenge — von jeder
// Fläche erteilt, von jeder gelesen, und dasselbe gezeigt.**
func TestDieAngabenEinerSichtGeltenAnJederFlaeche(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	an := starteDienst(t, umgebung)
	fenster := fensterOeffnen(t, umgebung)

	// **Drei Drucke mit Namen**: die Vorschau nennt sie, und ohne Namen
	// bliebe nur die Kennung.
	mussKommandozeile(t, umgebung, "exemplar", druckBolt, "3", "played", "de",
		"nonfoil", "2026-08-01", "4.17", "EUR", "aus dem Tausch",
		"Lightning Bolt", "lea", "161", "Kiste 4")
	mussKommandozeile(t, umgebung, "exemplar", druckCounter, "1", "played",
		"de", "nonfoil", "", "", "", "", "Counterspell", "lea", "54", "")
	mussKommandozeile(t, umgebung, "exemplar", druckGrowth, "2", "played",
		"de", "nonfoil", "", "", "", "", "Giant Growth", "lea", "197", "")

	vomFenster := einGegenueber(t, umgebung)
	vonDerKommandozeile := einGegenueber(t, umgebung)
	vomDienst := einGegenueber(t, umgebung)

	// **Das Fenster.**
	for _, rumpf := range []string{
		fensterRumpf(t, vomFenster, "handel", "auswahl",
			[]string{druckBolt, druckCounter}, 0),
		// **Ohne Behaltemenge**, an jeder Fläche: 0, keine Zahl, kein Feld —
		// und jede liest dieselbe Vorgabe 1 zurück (`MitVorgaben`, #2033).
		fensterRumpf(t, vomFenster, "teilen", "dubletten", []string{}, 0),
	} {
		if antwort := fenster.erteilen(rumpf); antwort.Fehler != nil {
			t.Fatalf("Fenster: %+v", antwort.Fehler)
		}
	}

	// **Die Kommandozeile.**
	mussKommandozeile(t, umgebung, "kontakt-freigeben", vonDerKommandozeile,
		"handel", "auswahl", druckBolt, druckCounter)
	mussKommandozeile(t, umgebung, "kontakt-freigeben", vonDerKommandozeile,
		"teilen", "dubletten")

	// **Der Dienst.**
	mussREST(t, an, http.MethodPost, "/peer/kontakte/"+vomDienst+"/freigaben",
		`{"zweck":"handel","umfang":"auswahl","auswahl":["`+druckBolt+`","`+
			druckCounter+`"]}`)
	mussREST(t, an, http.MethodPost, "/peer/kontakte/"+vomDienst+"/freigaben",
		`{"zweck":"teilen","umfang":"dubletten"}`)

	erwartet := []string{
		"handel/auswahl/" + druckBolt + "," + druckCounter + "/0",
		"teilen/dubletten//1",
	}

	// **Jede Fläche liest jedes Gegenüber** — wer ein Gegenüber von einer
	// anderen Fläche aus liest, sieht, dass die Freigabe nicht an der Fläche
	// hängt, die sie erteilt hat.
	for _, kennung := range []string{vomFenster, vonDerKommandozeile, vomDienst} {
		gelesen := map[string]string{
			"Fenster": fenster.lesen(kennung),
			"Kommandozeile": mussKommandozeile(t, umgebung, "--json",
				"kontakt-freigaben", kennung),
			"Dienst": mussREST(t, an, http.MethodGet,
				"/peer/kontakte/"+kennung+"/freigaben", ""),
		}

		for name, roh := range gelesen {
			if soll, ist := erwartet, freigabenAus(t, roh); !slices.Equal(soll, ist) {
				t.Errorf("%s liest für %s %v, erwartet %v", name, kennung, ist, soll)
			}
		}
	}

	// **Und gezeigt wird dasselbe**: die Auswahl nennt ihre zwei Drucke mit
	// der ganzen Menge, die Dubletten, was über der Behaltemenge liegt —
	// `counter` hat nur eines und fällt heraus.
	gezeigt := map[string][]string{
		"handel": {druckBolt + "×3", druckCounter + "×1"},
		"teilen": {druckBolt + "×2", druckGrowth + "×1"},
	}

	for _, kennung := range []string{vomFenster, vonDerKommandozeile, vomDienst} {
		for zweck, soll := range gezeigt {
			slices.Sort(soll)

			gesehen := map[string]string{
				"Fenster": fenster.sicht(kennung, zweck),
				"Kommandozeile": mussKommandozeile(t, umgebung, "--json",
					"kontakt-sicht", kennung, zweck),
				"Dienst": mussREST(t, an, http.MethodGet,
					"/peer/kontakte/"+kennung+"/sicht?zweck="+zweck, ""),
			}

			for name, roh := range gesehen {
				if ist := sichtAus(t, roh); !slices.Equal(soll, ist) {
					t.Errorf("%s zeigt %s für %s: %v, erwartet %v",
						name, zweck, kennung, ist, soll)
				}
			}
		}
	}
}

// **Eine Auswahl ohne Drucke gibt nichts frei — an jeder Fläche, mit demselben
// Satz.**
func TestEineAuswahlOhneDruckeWirdAnJederFlaecheAbgewiesen(t *testing.T) {
	const satz = "gibt nichts frei"

	umgebung := mitEigenerSaat(t)
	an := starteDienst(t, umgebung)
	fenster := fensterOeffnen(t, umgebung)
	kennung := einGegenueber(t, umgebung)

	// **Das Fenster nennt das Feld**, an dem die Oberfläche die Meldung
	// anzeigt (§6) — `auswahl`, nicht ein anonymer Satz.
	antwort := fenster.erteilen(fensterRumpf(t, kennung, "handel", "auswahl", []string{}, 0))
	if antwort.Fehler == nil || antwort.Fehler.Code != klasseValidierung ||
		antwort.Fehler.Feld != "auswahl" || !strings.Contains(antwort.Fehler.Meldung, satz) {
		t.Errorf("Fenster: %+v", antwort.Fehler)
	}

	_, fehlerstrom, code := kommandozeile(t, umgebung,
		"kontakt-freigeben", kennung, "handel", "auswahl")
	if code == 0 || !strings.Contains(fehlerstrom, satz) {
		t.Errorf("Kommandozeile: %d %q", code, fehlerstrom)
	}

	status, rumpf := an.ruf(t, http.MethodPost,
		"/peer/kontakte/"+kennung+"/freigaben",
		`{"zweck":"handel","umfang":"auswahl"}`)
	if status != http.StatusBadRequest || !strings.Contains(rumpf, satz) {
		t.Errorf("Dienst: %d %s", status, rumpf)
	}

	// **Und es ist nichts erteilt worden** — an keiner Fläche.
	if aus := mussKommandozeile(t, umgebung, "--json", "kontakt-freigaben",
		kennung); len(freigabenAus(t, aus)) != 0 {
		t.Errorf("trotz Abweisung steht etwas da: %s", aus)
	}
}
