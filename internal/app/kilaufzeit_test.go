package app_test

// **Die Einrichtung einer lokalen KI-Laufzeit** — Werkstück A7 von Phase 5.3
// (#1446).

import (
	"errors"
	"sync"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// einModell baut ein gültiges Modell mit einer belegten Fähigkeit.
//
// **Immer lokal**, und das ist der Punkt: die Umstufung auf `cloud` kommt aus
// der **Adresse** und nicht aus dem Modell ([domain.KIOrtFuerAdresse]). Ein
// Parameter dafür verführte dazu, sie am falschen Ende zu prüfen.
func einModell(alias string) domain.KIModell {
	return domain.KIModell{
		Alias: alias, Anbieter: "ollama", Laufzeit: "ollama 0.12.3",
		Ort: domain.KIOrtLokal,
		Geprueft: []domain.KIGeprueft{{
			Faehigkeit: domain.KITextgenerierung, Datum: "2026-09-11",
			Beleg: "Ollama meldet completion",
		}},
		Herkunft: domain.KIHerkunft{Glieder: []domain.KIGlied{{
			Art: domain.KIArtefakt, Kennung: alias, Quelle: "ollama:" + alias,
			Revision: "sha256:abc", Lizenz: "Apache-2.0",
			Dateien: []domain.KIDateihash{{Name: alias, Hash: "sha256:abc"}},
		}}},
	}
}

// **Ohne Schalter wird nicht gefragt.**
//
// # Warum das der erste Lauf ist
//
// **Weil eine Anwendung, die beim Start an einen Port klopft, etwas tut, um
// das niemand gebeten hat.** Die Vorgabe von `ki.laufzeit` ist `aus`, und
// „aus" heißt hier dasselbe wie bei einem Dienstschalter: **kein einziger
// Aufruf geht hinaus**.
func TestOhneSchalterWirdDieLaufzeitNichtGefragt(t *testing.T) {
	t.Parallel()

	laufzeit := &laufzeitdoppel{fassung: "0.12.3"}

	antwort := kianwendung(t, map[string]string{}, laufzeit).
		KIModelle(probekontext(t))
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if laufzeit.gefragt != 0 {
		t.Errorf("die Laufzeit wurde %d-mal gefragt", laufzeit.gefragt)
	}

	if antwort.Einrichtung != app.KIEinrichtungAus {
		t.Errorf("die Einrichtung ist %q", antwort.Einrichtung)
	}
}

// **Mit Schalter kommen die Modelle der Laufzeit.**
func TestMitSchalterKommenDieModelleDerLaufzeit(t *testing.T) {
	t.Parallel()

	laufzeit := &laufzeitdoppel{
		fassung: "0.12.3",
		modelle: []domain.KIModell{einModell("gemma4:8b")},
	}

	antwort := kianwendung(t, map[string]string{
		domain.KILaufzeit: domain.KILaufzeitAn,
	}, laufzeit).KIModelle(probekontext(t))
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(antwort.Modelle) != 1 || antwort.Modelle[0].Alias != "gemma4:8b" {
		t.Fatalf("die Modelle sind %+v", antwort.Modelle)
	}

	if antwort.Einrichtung != app.KIEinrichtungBereit {
		t.Errorf("die Einrichtung ist %q", antwort.Einrichtung)
	}
}

// **Eine Laufzeit, die nicht antwortet, ist kein Fehler der Anwendung.**
//
// **§16, Satz 5**: „Ohne Netz arbeitet die Anwendung weiter." Wer hier einen
// Fehler zurückgäbe, machte eine abgeschaltete Ollama-Installation zu einem
// Ausfall der Einstellungsseite.
func TestEineStummeLaufzeitIstKeinFehlerDerAnwendung(t *testing.T) {
	t.Parallel()

	antwort := kianwendung(t, map[string]string{
		domain.KILaufzeit: domain.KILaufzeitAn,
	}, &laufzeitdoppel{fehler: errors.New("niemand da")}).KIModelle(probekontext(t))

	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if antwort.Einrichtung != app.KIEinrichtungUnerreichbar {
		t.Errorf("die Einrichtung ist %q", antwort.Einrichtung)
	}

	// **Und der Grund steht dabei** — eine leere Seite ohne Erklärung ist
	// das, was §6 verbietet.
	//
	// **Er steht seit A8 (#1447) beim Anbieter** und nicht mehr über allem:
	// bei zwei Anbietern spräche ein gemeinsamer Befund von einem und sähe
	// aus, als spräche er von beiden.
	if len(befundeVon(t, antwort, "ollama")) == 0 {
		t.Error("die stumme Laufzeit kam ohne Befund")
	}
}

// **Eine Laufzeit ohne Modelle ist etwas anderes als keine Laufzeit.**
func TestEineLaufzeitOhneModelleSagtEsSo(t *testing.T) {
	t.Parallel()

	antwort := kianwendung(t, map[string]string{
		domain.KILaufzeit: domain.KILaufzeitAn,
	}, &laufzeitdoppel{fassung: "0.12.3"}).KIModelle(probekontext(t))

	if antwort.Einrichtung != app.KIEinrichtungOhneModell {
		t.Errorf("die Einrichtung ist %q", antwort.Einrichtung)
	}
}

// **Und ein Modell ohne belegte Fähigkeit wieder etwas anderes.**
//
// **Das ist der Satz „eine Modellliste garantiert keine Capability"** an der
// Fläche: installiert heißt nicht brauchbar, und die Seite sagt, welches von
// beidem fehlt.
func TestEinModellOhneFaehigkeitSagtEsSo(t *testing.T) {
	t.Parallel()

	ohne := einModell("einbetter:300m")
	ohne.Geprueft = nil

	antwort := kianwendung(t, map[string]string{
		domain.KILaufzeit: domain.KILaufzeitAn,
	}, &laufzeitdoppel{
		fassung: "0.12.3", modelle: []domain.KIModell{ohne},
	}).KIModelle(probekontext(t))

	if antwort.Einrichtung != app.KIEinrichtungOhneFaehigkeit {
		t.Errorf("die Einrichtung ist %q", antwort.Einrichtung)
	}
}

// **Eine fremde Adresse stuft jedes Modell dahinter als nicht-lokal ein.**
//
// # Warum das die harte Seite der Zusage ist
//
// **Weil die Laufzeit sich selbst als lokal meldet.** Ollama nennt ein Modell
// ohne `-cloud`-Suffix lokal, gleich wo der Dienst steht. Wer das übernähme,
// führte unter „nur lokal" ein Modell, das auf einem fremden Rechner rechnet
// — das Issue verlangt „eindeutige Einstufung".
func TestEineFremdeAdresseStuftJedesModellUm(t *testing.T) {
	t.Parallel()

	antwort := kianwendung(t, map[string]string{
		domain.KILaufzeit: domain.KILaufzeitAn,
		domain.KIAdresse:  "http://ki.example.org:11434",
	}, &laufzeitdoppel{
		fassung: "0.12.3",
		modelle: []domain.KIModell{einModell("gemma4:8b")},
	}).KIModelle(probekontext(t))

	if len(antwort.Modelle) != 1 {
		t.Fatalf("die Modelle sind %+v", antwort.Modelle)
	}

	if antwort.Modelle[0].Ort != string(domain.KIOrtCloud) {
		t.Errorf("der Ort ist %q", antwort.Modelle[0].Ort)
	}
}

// **Und unter »nur lokal« steht die Regel in der Antwort.**
//
// # Warum hier kein Befund steht
//
// **Weil `ki.laufzeit` kein Dienstschalter ist.** `richtlinie.nurlokal`
// schaltet **jeden fremden Dienst** ab (O3, #1427) — die lokale Laufzeit ist
// keiner, und ihr Schalter bleibt wirksam. Was die Regel hier tut, steht an
// den **Modellen**: eines unter einer fremden Adresse ist `cloud`, und
// `KIOrtsregel` weist es beim Vorschlag ab.
//
// **Der Befund, den es dafür gibt, gehört Anthropic** — siehe
// `kianthropic_test.go`.
func TestUnterNurLokalStehtDieRegelInDerAntwort(t *testing.T) {
	t.Parallel()

	antwort := kianwendung(t, map[string]string{
		domain.KILaufzeit:         domain.KILaufzeitAn,
		domain.KIAdresse:          "http://ki.example.org:11434",
		domain.RichtlinieNurLokal: domain.NurLokalAn,
	}, &laufzeitdoppel{
		fassung: "0.12.3",
		modelle: []domain.KIModell{einModell("gemma4:8b")},
	}).KIModelle(probekontext(t))

	if !antwort.NurLokal {
		t.Fatal("die Regel steht nicht in der Antwort")
	}

	// **Und das Modell ist umgestuft** — daran greift sie.
	if len(antwort.Modelle) != 1 ||
		antwort.Modelle[0].Ort != string(domain.KIOrtCloud) {
		t.Errorf("die Modelle sind %+v", antwort.Modelle)
	}
}

// kianwendung baut eine Anwendung mit gesetzten Einstellungen und Laufzeit.
func kianwendung(
	t *testing.T, werte map[string]string, laufzeit app.KILaufzeitquelle,
) *app.App {
	t.Helper()

	db := frischerBestand(t)

	for schluessel, wert := range werte {
		if fehler := storage.NeueEinstellungen(db).Setzen(
			probekontext(t), schluessel, wert); fehler != nil {
			t.Fatalf("%s setzen: %v", schluessel, fehler)
		}
	}

	// **Ein eigenes Zuhause je Lauf** (§3): ohne es läse der Cloudstand die
	// Serverdatei dieses Rechners, und der Lauf wäre auf dem nächsten ein
	// anderer.
	return app.KIAnwendungUeber(db).MitKIZuhause(t.TempDir()).
		MitKILaufzeit(func(string) app.KILaufzeitquelle { return laufzeit })
}

// befundeVon gibt die Befunde eines Anbieters.
func befundeVon(t *testing.T, antwort app.KIKatalogantwort, name string) []string {
	t.Helper()

	for _, anbieter := range antwort.Anbieter {
		if anbieter.Name == name {
			return anbieter.Befunde
		}
	}

	t.Fatalf("den Anbieter %q gibt es nicht: %+v", name, antwort.Anbieter)

	return nil
}

// einrichtungVon gibt den Einrichtungsstand eines Anbieters.
func einrichtungVon(t *testing.T, antwort app.KIKatalogantwort, name string) string {
	t.Helper()

	for _, anbieter := range antwort.Anbieter {
		if anbieter.Name == name {
			return anbieter.Einrichtung
		}
	}

	t.Fatalf("den Anbieter %q gibt es nicht: %+v", name, antwort.Anbieter)

	return ""
}

// gleichzeitig zählt den **Höchststand** gleichzeitiger Rufe.
//
// **Der Höchststand und nicht die Summe** (§10): wie viele Anfragen insgesamt
// hinausgingen, sagt über Parallelität nichts. Gemessen wird, wie viele
// zugleich unterwegs waren.
type gleichzeitig struct {
	schloss sync.Mutex
	jetzt   int
	hoechst int
}

func (g *gleichzeitig) an() {
	g.schloss.Lock()
	defer g.schloss.Unlock()

	g.jetzt++

	if g.jetzt > g.hoechst {
		g.hoechst = g.jetzt
	}
}

func (g *gleichzeitig) aus() {
	g.schloss.Lock()
	defer g.schloss.Unlock()

	g.jetzt--
}

func (g *gleichzeitig) hoechststand() int {
	g.schloss.Lock()
	defer g.schloss.Unlock()

	return g.hoechst
}
