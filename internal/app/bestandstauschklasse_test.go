// **Der Tausch des Bestands als Fehlerklasse** — A05
// (#2219).
//
// # Wofür
//
// A02 hat den Erfolgsweg geheilt: nach einem Tausch geht das Tor wieder auf.
// Ein Tausch hat aber mehr Ausgänge als den Erfolg, und **jeder** muss auf
// einem vollständigen Stand enden — dem alten oder dem neuen, nie einem
// gemischten — mit offenem Tor und einer Meldung, die sagt, welcher es ist.
//
// # Was jeder Fall prüft
//
//  1. Der Ausgang ist der erwartete (Erfolg oder ein benannter Fehler).
//  2. **Danach geht das Tor**: lesen und schreiben über [app.Bindungen].
//  3. Der Stand ist **ganz** der alte oder **ganz** der neue — an einer
//     **Kohorte** aus drei Teilen, die beide verschieden tragen: einer
//     Einstellung, einem Kontakt im Profil und einem Bild im Beiwerk. Ein
//     Mischstand aus neuen Daten und altem Beiwerk fiele daran auf (Befund
//     5 des Zweitblicks über Phase A — bis dahin trug nur die Einstellung).
//  4. Die Datei ist heil: `PRAGMA quick_check` und `foreign_key_check` —
//     **Integrität, kein Inhaltsvergleich**; den macht Punkt 3.
//  5. Der Ausgang nennt seinen Grund: der erwartete Fehlercode, beim Erfolg
//     der Ort, an dem der alte Bestand liegt.
//
// # Was er nicht prüft
//
// **Einen Abbruch mitten im Umbenennen** — dafür gibt es die Marke und ihre
// Läufe in `internal/service` (`rueckspielmarke*_test.go`): sie halten den
// Prozesstod zwischen Daten- und Bildertausch, den ein Lauf in einem Prozess
// nicht herstellen kann.

package app_test

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// tauschausgang ist ein Fall der Matrix.
type tauschausgang struct {
	name string

	// vorbereiten richtet die Lage her, nachdem gesichert ist; was sie
	// zurückgibt, räumt danach auf.
	vorbereiten func(t *testing.T, anwendung *app.App, ort, bestand, sicherung string) (loslassen func())

	// kontext ist der Kontext des Rückspielens.
	kontext func(t *testing.T) context.Context

	// schluessel ist der Schlüssel der Meldung, wenn es scheitert.
	schluessel string

	// gelingt sagt, ob getauscht werden soll — und damit, welcher Stand
	// danach erwartet ist: die neue Kohorte aus der Sicherung oder die alte.
	gelingt bool

	// code ist der erwartete Fehlercode, wenn es nicht gelingt.
	code app.Fehlercode
}

// kohorte ist ein Stand, an drei Stellen abgelesen.
type kohorte struct {
	Sprache  string
	Kontakte []string
	Bild     string
}

// kohorteSetzen legt einen Stand an: Sprache, genau einen Kontakt und ein
// Bild im Beiwerk — alle drei mit dem Namen der Kohorte.
func kohorteSetzen(t *testing.T, anwendung *app.App, bilder, name, sprache string) {
	t.Helper()

	if antwort := anwendung.EinstellungSetzen(t.Context(), "sprache", sprache); antwort.Fehler != nil {
		t.Fatalf("setzen: %+v", antwort.Fehler)
	}

	for _, vorhanden := range anwendung.Kontakte(t.Context()).Kontakte {
		if antwort := anwendung.KontaktLoeschen(t.Context(), vorhanden.ID, vorhanden.Fassung); antwort.Fehler != nil {
			t.Fatalf("Kontakt löschen: %+v", antwort.Fehler)
		}
	}

	if antwort := anwendung.KontakteUebernehmen(t.Context(), []app.KontaktDTO{{Name: "Kohorte " + name}}); antwort.Fehler != nil {
		t.Fatalf("Kontakt anlegen: %+v", antwort.Fehler)
	}

	if fehler := os.MkdirAll(bilder, 0o700); fehler != nil {
		t.Fatalf("Beiwerk anlegen: %v", fehler)
	}

	if fehler := os.WriteFile(filepath.Join(bilder, "kohorte.png"), []byte(name), 0o600); fehler != nil {
		t.Fatalf("Bild schreiben: %v", fehler)
	}
}

// kohorteLesen liest den Stand an denselben drei Stellen.
func kohorteLesen(t *testing.T, anwendung *app.App, bilder string) kohorte {
	t.Helper()

	gelesen := kohorte{Sprache: merkmal(t, anwendung)}
	for _, kontakt := range anwendung.Kontakte(t.Context()).Kontakte {
		gelesen.Kontakte = append(gelesen.Kontakte, kontakt.Name)
	}

	roh, _ := os.ReadFile(filepath.Join(bilder, "kohorte.png")) //nolint:errcheck,gosec // aus t.TempDir(); fehlt es, steht "" da

	gelesen.Bild = string(roh)

	return gelesen
}

func TestJederAusgangDesTauschsEndetAufEinemGanzenStand(t *testing.T) {
	alt := kohorte{Sprache: "en", Kontakte: []string{"Kohorte alt"}, Bild: "alt"}
	neu := kohorte{Sprache: "de", Kontakte: []string{"Kohorte neu"}, Bild: "neu"}

	ohneAufraeumen := func() {}
	gesichert := func(*testing.T, *app.App, string, string, string) func() { return ohneAufraeumen }

	faelle := []tauschausgang{
		{name: "Erfolg", vorbereiten: gesichert, kontext: probekontext, gelingt: true},
		{
			name: "ein zweiter Prozess hält den Bestand",
			vorbereiten: func(t *testing.T, _ *app.App, _, bestand, _ string) func() {
				t.Helper()

				return zweiterHalter(t, bestand)
			},
			kontext: probekontext, code: app.CodeKonflikt, schluessel: "konfliktgrund",
		},
		{
			name: "die Sicherung ist abgeschnitten",
			vorbereiten: func(t *testing.T, _ *app.App, _, _, sicherung string) func() {
				t.Helper()
				abschneiden(t, sicherung)

				return ohneAufraeumen
			},
			kontext: probekontext, code: app.CodeValidierung, schluessel: "validierungsgrund",
		},
		{
			// **Scheitert vor dem Tausch** — am Zwischenplatz neben dem
			// Bestand. Wie es nach dem Tausch scheitert, prüft
			// `TestEinScheiterndesOeffnenNachDemTausch…` (intern).
			name: "der Ort des Bestands ohne Schreibrecht",
			vorbereiten: func(t *testing.T, _ *app.App, ort, _, _ string) func() {
				t.Helper()

				//nolint:gosec // G302: ein Verzeichnis braucht das Ausführungsrecht, um es zu betreten
				if fehler := os.Chmod(ort, 0o500); fehler != nil {
					t.Fatalf("Rechte nehmen: %v", fehler)
				}

				//nolint:errcheck,gosec // t.TempDir räumt sonst nicht; ein Verzeichnis braucht x
				return func() { _ = os.Chmod(ort, 0o700) }
			},
			kontext: probekontext, code: app.CodeValidierung, schluessel: "gesperrt",
		},
		{
			name: "der Aufruf ist schon abgebrochen", vorbereiten: gesichert,
			kontext: func(t *testing.T) context.Context {
				t.Helper()

				ctx, abbrechen := context.WithCancel(probekontext(t))
				abbrechen()

				return ctx
			},
			code: app.CodeAbgebrochen, schluessel: "vorgangAbgebrochen",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			ort := t.TempDir()
			bestand := filepath.Join(ort, "bestand.db")
			bilder := filepath.Join(ort, "beiwerk")
			anwendung := laufendeAnwendung(t, bestand)
			anwendung.MitBilderort(func() (string, error) { return bilder, nil })
			bindungen := app.NeueBindungen(anwendung)

			kohorteSetzen(t, anwendung, bilder, "neu", "de")

			sicherung := filepath.Join(ort, "gut.msesicherung")
			anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) { return sicherung, nil })

			if antwort := anwendung.BestandSichern(t.Context(), "", false); antwort.Fehler != nil {
				t.Fatalf("sichern: %+v", antwort.Fehler)
			}

			kohorteSetzen(t, anwendung, bilder, "alt", "en")

			loslassen := fall.vorbereiten(t, anwendung, ort, bestand, sicherung)
			antwort := bindungen.SicherungZurueckspielen(fall.kontext(t), sicherung, "")
			loslassen()

			if fall.gelingt != (antwort.Fehler == nil && antwort.Ersetzt) {
				t.Fatalf("Ausgang %+v, erwartet war gelingt=%v", antwort, fall.gelingt)
			}

			erwartet := alt
			if fall.gelingt {
				erwartet = neu

				if antwort.Beiseite == "" {
					t.Error("der Erfolg sagt nicht, wo der alte Bestand liegt")
				}
			} else {
				if antwort.Fehler.Code != fall.code || string(antwort.Fehler.Schluessel) != fall.schluessel {
					t.Errorf("Fehler %+v, erwartet war %q/%q", antwort.Fehler, fall.code, fall.schluessel)
				}

				// **Die ausdrückliche Rückkehrmeldung** (A05, Nachlauf über
				// Phase A): ein Fehler sagt, ob der bisherige Stand wieder
				// ganz dasteht — die Kohorte darunter prüft, dass es stimmt.
				if !antwort.Unveraendert {
					t.Error("der Fehler sagt nicht, dass der bisherige Stand unverändert ist")
				}
			}

			if gelesen := kohorteLesen(t, anwendung, bilder); !reflect.DeepEqual(gelesen, erwartet) {
				t.Errorf("danach steht %+v, erwartet war die ganze Kohorte %+v", gelesen, erwartet)
			}

			durchsTor(t, bindungen, "nach dem Ausgang "+fall.name)
			heil(t, bestand)
		})
	}
}

// durchsTor liest und schreibt über die Bindungen — **durch das Tor**, wie
// das Fenster — und meldet eine Panik als Befund statt als Absturz.
func durchsTor(t *testing.T, bindungen *app.Bindungen, wann string) {
	t.Helper()

	defer func() {
		if grund := recover(); grund != nil {
			t.Fatalf("%s weist das Tor ab: %v", wann, grund)
		}
	}()

	ctx := t.Context()

	if antwort := bindungen.Einstellungen(ctx); antwort.Fehler != nil {
		t.Fatalf("%s: Einstellungen lesen: %+v", wann, antwort.Fehler)
	}

	if antwort := bindungen.EinstellungSetzen(ctx, "sprache", "en"); antwort.Fehler != nil {
		t.Fatalf("%s: eine Einstellung schreiben: %+v", wann, antwort.Fehler)
	}

	if antwort := bindungen.Datenstand(ctx); antwort.Fehler != nil {
		t.Fatalf("%s: den Datenstand lesen: %+v", wann, antwort.Fehler)
	}

	if antwort := bindungen.Sets(ctx); antwort.Fehler != nil {
		t.Fatalf("%s: die Sets lesen: %+v", wann, antwort.Fehler)
	}
}

// abschneiden kürzt eine Datei auf die Hälfte — ein abgebrochener Download,
// ein voller Datenträger beim Sichern.
func abschneiden(t *testing.T, pfad string) {
	t.Helper()

	angaben, fehler := os.Stat(pfad)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if fehler := os.Truncate(pfad, angaben.Size()/2); fehler != nil {
		t.Fatalf("abschneiden: %v", fehler)
	}
}

// heil prüft die Datei des Bestands: Seiten und Fremdschlüssel.
func heil(t *testing.T, bestand string) {
	t.Helper()

	db, fehler := storage.OeffnenNurLesend(t.Context(), bestand)
	if fehler != nil {
		t.Fatalf("lesend öffnen: %v", fehler)
	}

	defer func() { _ = db.Close() }() //nolint:errcheck // nur gelesen

	var befund string
	if pruefung := db.QueryRowContext(t.Context(), "PRAGMA quick_check").Scan(&befund); pruefung != nil || befund != "ok" {
		t.Errorf("quick_check: %q (%v)", befund, pruefung)
	}

	verwaist, fehler := db.QueryContext(t.Context(), "PRAGMA foreign_key_check")
	if fehler != nil {
		t.Fatalf("foreign_key_check: %v", fehler)
	}

	defer func() { _ = verwaist.Close() }() //nolint:errcheck // nur gelesen

	if verwaist.Next() {
		t.Error("ein Fremdschlüssel zeigt ins Leere")
	}

	if lesefehler := verwaist.Err(); lesefehler != nil {
		t.Errorf("die Fremdschlüssel prüfen: %v", lesefehler)
	}
}
