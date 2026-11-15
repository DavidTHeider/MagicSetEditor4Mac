package flaechen_test

// **Ein Handel von der Eröffnung bis zum beidseitigen Abschluss** —
// #1637.
//
// # Was hier gemessen wird
//
// **Der geglückte Weg, abwechselnd über beide Flächen** — eröffnen,
// anbieten, zustimmen, vorbereiten, übergeben, verbuchen, abgleichen,
// bestätigen. Nach **jedem** Schritt wird der Stand über die **andere**
// Fläche gelesen: zwei Flächen, die verschiedene Stände zeigten, wären zwei
// Anwendungen.
//
// **Das war bis zu diesem Issue nicht möglich**, und beide Läufe von `T6` und
// `T10` sagten es in ihrem Kopf: `service.Handelsleitung` hatte keinen
// Aufrufer, und ohne sie kam kein Vorgang bis zur bestätigten Übergabe. Die
// Absätze „Was er **nicht** misst" dort sind mit diesem Lauf weggefallen.
//
// # Und wer die Freigabe erteilt
//
// **Das Fenster, und nur das Fenster** (T10,
// #1475).
// Dieser Lauf öffnet dafür denselben Bestand, den die Kindprozesse benutzen,
// und stellt den Nachweis über `internal/app` aus — ohne Dienstmodus, also
// als Fenster. **Genau das ist die Zusage**: Kommandozeile und Dienst können
// ihn **einlösen** und nicht ausstellen, und beides steht hier nebeneinander.

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schluesselbund"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// derHandel ist der Vorgang, den dieser Lauf führt.
const derHandel = "v-abnahme"

// dasGegenueberImHandel ist die zweite Partei.
//
// **Eine erfundene Peerkennung und kein zweiter Rechner**: dieser Lauf misst
// die **Flächen**, nicht die Leitung — die steht als `T12`
// (#1477).
// Was hier über die Kennung des Gegenübers hereinkommt, ist seine Stimme, und
// sie kommt so an, wie sie über eine Leitung ankäme.
const dasGegenueberImHandel = "bgegenueber"

// diePositionen sind die Zeilen des Angebots.
//
// **Alles kommt von drüben, nichts geht von hier weg.** Das ist Absicht: was
// weggeht, muss festgehalten sein, und **welche** Besitzzeile weggeht, sagt
// ein Angebot nicht — es nennt einen Druck. Das wählt ein Mensch, und dafür
// steht #1643.
const diePositionen = `[{"von":"` + dasGegenueberImHandel +
	`","printing":"bolt-m10","menge":1,"guete":"nearMint","sprache":"de",` +
	`"finish":"nonfoil"}]`

// derStandweg und derVorschlagsweg sind die Endpunkte, wie sie im Katalog
// stehen: `GET /handel/{vorgang}` und `POST /handel/{vorgang}/vorschlag`.
const (
	derStandweg      = "GET /handel/{vorgang}"
	derVorschlagsweg = "POST /handel/{vorgang}/vorschlag"
	derVersandweg    = "POST /handel/{vorgang}/versand"
	derAufnahmeweg   = "POST /handel/{vorgang}/wiederaufnahme"
	derPostweg       = "POST /handelspost/{nachrichtId}/bestaetigung"
	derBelegweg      = "GET /handel/{vorgang}/beleg"
)

func TestEinHandelKommtUeberBeideFlaechenBisZumAbschluss(t *testing.T) {
	umgebung := append(mitEigenerSaat(t),
		schluesselbund.Bundort+"="+probebund(t))
	an := starteDienst(t, umgebung)

	// **Ohne Identität kein Initiator** — an ihm hängt die Sequenz.
	angelegt := mussKommandozeile(t, umgebung, "peer-anlegen")

	if stand := zeilenwert(t, angelegt, "stand"); stand != "bereit" {
		t.Fatalf("die Identität steht auf %q", stand)
	}

	ich := zeilenwert(t, angelegt, "kennung")

	// **Eröffnet wird an der Kommandozeile, die Positionen von der Eingabe.**
	mitEingabe(t, umgebung, diePositionen,
		"handel", "eroeffnen", derHandel, "tausch", dasGegenueberImHandel)

	standGleich(t, umgebung, an, "entwurf")

	// **Angeboten wird über den Dienst** — und die Kommandozeile sieht es.
	schrittUeberDenDienst(t, an, "anbieten", "")
	standGleich(t, umgebung, an, "angeboten")

	// **Zustimmen braucht beide Stimmen**: die eigene kommt von hier, die
	// des Gegenübers kommt so an, wie sie über eine Leitung ankäme.
	mussKommandozeile(t, umgebung, "handel", "zustimmen", derHandel)
	standGleich(t, umgebung, an, "angeboten")

	schrittUeberDenDienst(t, an, "zustimmen", dasGegenueberImHandel)
	standGleich(t, umgebung, an, "angenommen")

	mussKommandozeile(t, umgebung, "handel", "vorbereiten", derHandel)
	standGleich(t, umgebung, an, "vorbereitet")

	// **Übergeben braucht ebenfalls beide** — ein Zustand, den eine Seite
	// allein herstellte, hieße „übergeben" und meinte „ich sage, dass ich
	// übergeben habe".
	mussKommandozeile(t, umgebung, "handel", "uebergeben", derHandel)
	schrittUeberDenDienst(t, an, "uebergeben", dasGegenueberImHandel)
	standGleich(t, umgebung, an, "uebergabeOffen")

	abschliessen(t, umgebung, an, ich)
}

// abschliessen fährt die drei Schritte, die eine Freigabe verlangen.
func abschliessen(
	t *testing.T, umgebung []string, an dienst, ich string,
) {
	t.Helper()

	mussKommandozeile(t, umgebung, "handel", "verbuchen", derHandel,
		eineFreigabe(t, umgebung, "verbuchen"))
	standGleich(t, umgebung, an, "lokalVerbucht")

	mussKommandozeile(t, umgebung, "handel", "abgleichen", derHandel)
	standGleich(t, umgebung, an, "abgleichOffen")

	// **Zwei Stimmen, zwei Freigaben** — ein Nachweis wird **einmal**
	// eingelöst, und wer zwei Unterschriften mit einer erteilte, hätte die
	// zweite nicht gewollt.
	mussKommandozeile(t, umgebung, "handel", "bestaetigen", derHandel, ich,
		eineFreigabeFuer(t, umgebung, derHandel, "bestaetigen", ich))
	standGleich(t, umgebung, an, "abgleichOffen")

	mussKommandozeile(t, umgebung, "handel", "bestaetigen", derHandel,
		dasGegenueberImHandel, eineFreigabeFuer(t, umgebung, derHandel,
			"bestaetigen", dasGegenueberImHandel))
	standGleich(t, umgebung, an, "beidseitigVerbucht")

	belegGleich(t, umgebung, an)
	befundGleich(t, umgebung, an) // #2500

	// **Und der Beleg ist beidseitig** — der Zustand ist eine Marke, der
	// Beleg das Papier dazu.
	aus := mussREST(t, an, http.MethodGet,
		strings.ReplaceAll(pfadAus(derStandweg), "{vorgang}", derHandel), "")
	if !strings.Contains(aus, `"ende":true`) {
		t.Errorf("der abgeschlossene Vorgang gilt nicht als Ende: %s", aus)
	}
}

// belegGleich liest den Beleg über beide Flächen, **ohne** einen Schritt —
// wer unterschrieben hat, sagte bis zur Hüllenprobe `nat04handel` (#1672)
// nur die Antwort auf den eigenen.
func belegGleich(t *testing.T, umgebung []string, an dienst) {
	t.Helper()

	type beleg struct {
		Vorgang, Zustand, Abdruck, Unterschrift, Gebucht string
		Beidseitig                                       bool
	}

	var ueberCLI, ueberREST beleg

	for ziel, roh := range map[*beleg]string{
		&ueberCLI: mussKommandozeile(t, umgebung,
			"--json", "handelsbeleg", derHandel),
		&ueberREST: mussREST(t, an, http.MethodGet,
			strings.ReplaceAll(pfadAus(derBelegweg), "{vorgang}", derHandel), ""),
	} {
		if fehler := json.Unmarshal([]byte(roh), ziel); fehler != nil {
			t.Fatalf("kein Beleg: %v\n%s", fehler, roh)
		}
	}

	if ueberCLI != ueberREST || ueberCLI.Unterschrift == "" ||
		ueberCLI.Zustand != "beidseitigVerbucht" {
		t.Errorf("die Kommandozeile liest %+v, der Dienst %+v", ueberCLI, ueberREST)
	}
}

// standGleich liest den Stand über beide Flächen und vergleicht ihn.
//
// **Beide und nicht eine.** Der Satz aus §16 lautet „dieselbe Fähigkeit,
// überall dieselbe Antwort" — und eine Fläche, die ihn allein bestätigte,
// bestätigte sich selbst.
func standGleich(
	t *testing.T, umgebung []string, an dienst, erwartet string,
) {
	t.Helper()

	ausCLI := mussKommandozeile(t, umgebung, "handel", "stand", derHandel)
	if !strings.Contains(ausCLI, erwartet) {
		t.Fatalf("die Kommandozeile sagt %q statt %q", ausCLI, erwartet)
	}

	ausREST := mussREST(t, an, http.MethodGet,
		strings.ReplaceAll(pfadAus(derStandweg), "{vorgang}", derHandel), "")
	if !strings.Contains(ausREST, `"zustand":"`+erwartet+`"`) {
		t.Fatalf("der Dienst sagt %q statt %q", ausREST, erwartet)
	}
}

// schrittUeberDenDienst geht einen Schritt über `POST /handel/…/schritt`.
func schrittUeberDenDienst(t *testing.T, an dienst, schritt, wer string) {
	t.Helper()

	schrittAmVorgang(t, an, derHandel, schritt, wer)
}

// schrittAmVorgang tut dasselbe für einen benannten Vorgang.
//
// **Getrennt, seit ein zweiter Lauf einen zweiten Vorgang führt** (#1643):
// die feste Kennung darüber war bis dahin die einzige.
func schrittAmVorgang(
	t *testing.T, an dienst, vorgang, schritt, wer string,
) {
	t.Helper()

	mussREST(t, an, http.MethodPost,
		strings.ReplaceAll(pfadAus(derSchrittweg), "{vorgang}", vorgang),
		`{"schritt":"`+schritt+`","wer":"`+wer+`"}`)
}

// eineFreigabe stellt den Nachweis aus — **am Fenster**.
//
// **Über denselben Bestand, den die Kindprozesse benutzen**, und ohne
// Dienstmodus: `Nachweise.Ausstellen` weist jeden anderen ab. Damit steht
// hier nebeneinander, was die Zusage aus T10 ausmacht — ausgestellt wird am
// Fenster, eingelöst an der Kommandozeile.
func eineFreigabe(t *testing.T, umgebung []string, schritt string) string {
	t.Helper()

	return eineFreigabeFuer(t, umgebung, derHandel, schritt, "")
}

// eineFreigabeFuer tut dasselbe für einen benannten Vorgang (#1643) und eine
// benannte Partei (Befund 5 des Zweitblicks über Phase 5.6).
func eineFreigabeFuer(
	t *testing.T, umgebung []string, vorgang, schritt, wer string,
) string {
	t.Helper()

	db, fehler := storage.Oeffnen(t.Context(), bestandsortAus(t, umgebung))
	if fehler != nil {
		t.Fatalf("den Bestand öffnen: %v", fehler)
	}

	defer func() { _ = db.Close() }() //nolint:errcheck // nur zum Aufräumen

	// **Das gezeigte Profil und keine geratene Kennung** — jede
	// Wurzelabfrage klemmt darauf
	// ([0066](../../docs/entscheidungen/0066-nutzer-profile-und-die-vier-entscheidungen.md)),
	// und die Kindprozesse arbeiten in demselben.
	ctx, profilfehler := storage.MitGezeigtemProfil(t.Context(), db)
	if profilfehler != nil {
		t.Fatalf("das Profil lesen: %v", profilfehler)
	}

	antwort := app.PeerflaecheUeber(db).HandelFreigabeErteilen(
		ctx, vorgang, schritt, wer)
	if antwort.Fehler != nil {
		t.Fatalf("die Freigabe für %q: %v", schritt, antwort.Fehler.AlsFehler())
	}

	return antwort.Kennung
}

// **Ein Vorschlag und der Ausgang, über beide Flächen.**
//
// # Was hier gemessen wird
//
//   - **Dass der Initiator aus `angeboten` nicht vorschlägt** (#1975) — und
//     beide Flächen dieselbe Abweisung geben. Bis dahin hielt das nur das
//     Fenster, und hier stand ein Vorschlag des Initiators, der die zweite
//     Revision anlegte. Dass ein Vorschlag der Gegenseite die nächste Revision
//     anlegt, zeigt der Lauf über zwei Rechner (`handelsabnahme_test.go`).
//   - **Dass eine Nachricht im Ausgang liegen bleibt, bis jemand sie
//     bestätigt** (`T5`): die Wiederaufnahme nennt sie, die Bestätigung räumt
//     sie weg, und danach nennt die Wiederaufnahme sie nicht mehr.
//   - **Und dass beide Flächen denselben Ausgang sehen.** Zwei Ausgänge
//     nebeneinander wären zwei Programme.
func TestDerAusgangIstUeberBeideFlaechenDerselbe(t *testing.T) {
	umgebung := append(mitEigenerSaat(t),
		schluesselbund.Bundort+"="+probebund(t))
	an := starteDienst(t, umgebung)

	mussKommandozeile(t, umgebung, "peer-anlegen")

	mitEingabe(t, umgebung, diePositionen,
		"handel", "eroeffnen", derHandel, "tausch", dasGegenueberImHandel)
	mussKommandozeile(t, umgebung, "handel", "anbieten", derHandel)

	// **Der Vorschlag des Initiators wird über den Dienst abgewiesen** …
	status, rumpf := an.ruf(t, http.MethodPost,
		strings.ReplaceAll(pfadAus(derVorschlagsweg), "{vorgang}", derHandel),
		`{"positionen":`+diePositionen+`}`)
	if status != http.StatusBadRequest || !strings.Contains(rumpf, "nicht vorschlagen") {
		t.Errorf("der Dienst nahm den Vorschlag des Initiators: %d %s", status, rumpf)
	}

	// **… und an der Kommandozeile mit demselben Satz.**
	_, klage, code := mitEingabeUndCode(t, umgebung, diePositionen,
		"handel", "vorschlagen", derHandel)
	if code == 0 || !strings.Contains(klage, "nicht vorschlagen") {
		t.Errorf("die Kommandozeile nahm den Vorschlag des Initiators: %d %s", code, klage)
	}

	standGleich(t, umgebung, an, "angeboten")

	if aus := mussKommandozeile(t, umgebung, "handel", "stand",
		derHandel); !strings.Contains(aus, "\t1\t") {
		t.Errorf("die Kommandozeile nennt nicht mehr die erste Revision: %s", aus)
	}

	ausgangGleich(t, umgebung, an)
}

// ausgangGleich legt eine Nachricht in den Ausgang und räumt sie wieder weg.
func ausgangGleich(t *testing.T, umgebung []string, an dienst) {
	t.Helper()

	// **Geschickt wird über den Dienst** — er nennt die Kennung zurück.
	geschickt := mussREST(t, an, http.MethodPost,
		strings.ReplaceAll(pfadAus(derVersandweg), "{vorgang}", derHandel),
		`{"empfaenger":"`+dasGegenueberImHandel+`"}`)

	nachricht := zwischen(t, geschickt, `"nachrichtId":"`, `"`)

	// **Und die Kommandozeile sieht dieselbe Nachricht liegen.**
	offen := mussKommandozeile(t, umgebung, "handelspost-wiederaufnahme",
		derHandel, dasGegenueberImHandel, "0", "", "0")
	if !strings.Contains(offen, nachricht) {
		t.Fatalf("die Kommandozeile kennt %q nicht: %s", nachricht, offen)
	}

	if !strings.Contains(offen, "klaerung\tfalse") {
		t.Errorf("die Wiederaufnahme meldet eine Klärung: %s", offen)
	}

	// **Eine zweite Nachricht von der Kommandozeile liegt daneben.**
	//
	// **Ohne `--json`**, weil dieser Befehl eine Kennung und sonst nichts
	// ausgibt — eine Zeile, die man lesen kann, ist hier die ganze Antwort.
	zweite := strings.TrimSpace(mussKommandozeile(t, umgebung,
		"handelspost-schicken", derHandel, dasGegenueberImHandel))
	if zweite == nachricht {
		t.Fatalf("beide Wege gaben dieselbe Kennung %q", zweite)
	}

	// **Bestätigt wird über die Kommandozeile, nachgesehen über den Dienst.**
	mussKommandozeile(t, umgebung, "handelspost-bestaetigen", nachricht)

	danach := mussREST(t, an, http.MethodPost,
		strings.ReplaceAll(pfadAus(derAufnahmeweg), "{vorgang}", derHandel),
		`{"empfaenger":"`+dasGegenueberImHandel+
			`","revision":0,"abdruck":"","folge":0}`)
	if strings.Contains(danach, nachricht) {
		t.Errorf("die bestätigte Nachricht liegt noch im Ausgang: %s", danach)
	}

	if !strings.Contains(danach, zweite) {
		t.Errorf("die zweite Nachricht fehlt im Ausgang: %s", danach)
	}

	// **Und über den Dienst geht die zweite ebenso weg.**
	mussREST(t, an, http.MethodPost,
		strings.ReplaceAll(pfadAus(derPostweg), "{nachrichtId}", zweite), `{}`)

	leer := mussKommandozeile(t, umgebung, "handelspost-wiederaufnahme",
		derHandel, dasGegenueberImHandel, "0", "", "0")
	if strings.Contains(leer, zweite) {
		t.Errorf("die zweite Nachricht liegt noch im Ausgang: %s", leer)
	}
}

// zwischen schneidet aus einer Ausgabe, was zwischen zwei Marken steht.
func zwischen(t *testing.T, aus, vor, nach string) string {
	t.Helper()

	_, dahinter, gefunden := strings.Cut(aus, vor)
	if !gefunden {
		t.Fatalf("%q steht nicht in %s", vor, aus)
	}

	wert, _, gefunden := strings.Cut(dahinter, nach)
	if !gefunden {
		t.Fatalf("%q endet nicht mit %q: %s", dahinter, nach, aus)
	}

	return wert
}
