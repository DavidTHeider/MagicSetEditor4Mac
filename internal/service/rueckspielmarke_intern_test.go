// **Gemeinsam oder gar nicht** — Werkstück 1 von
// #1162,
// angestrichen vom Nachlauf des Zweitblicks als **nicht eingelöst**.
//
// # Was fehlte
//
// Der Tausch hat zwei Schritte: erst `bestand.db`, dann `beiwerk/`. Für
// **gemeldete** Fehler gab es einen Rückweg. Für einen **Abbruch** nicht:
// stirbt der Prozess dazwischen, steht die neue Datenbank neben den alten
// Bildern, und niemand erfährt es.
//
// Die Antwort darauf war zuerst ein Folgeissue. Das war zu wenig — das
// Werkstück steht wörtlich im Issue, und ein Abnahmekriterium nachträglich
// herauszuschneiden ist keine Begründung, sondern eine Ausrede (§10).
//
// # Wie der Abbruch hergestellt wird
//
// Nicht mit einem echten Kill: ein Test, der sich selbst umbringt, prüft die
// Testumgebung. Hergestellt wird der **Zustand**, den ein Kill hinterlässt —
// Marke da, Auszug da, Bilder noch die alten — und danach gestartet.
package service

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sicherung"
)

// **Ein abgebrochenes Zurückspielen wird beim Start zu Ende geführt.**
func TestEinAbgebrochenesRueckspielenWirdBeimStartVollendet(t *testing.T) {
	weg, ort, bilder := mitBilderort(t)

	if err := os.MkdirAll(filepath.Join(bilder, "7"), 0o700); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if err := os.WriteFile(filepath.Join(bilder, "7", "kunst.png"),
		[]byte("der alte Stand"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	ziel := gesichert(t, weg, filepath.Join(ort, "gut"+sicherung.Endung))

	// **Der Zustand nach einem Kill**: der Auszug liegt da, die Datenbank ist
	// schon getauscht, die Bilder sind noch die alten.
	ausgepackt, fehler := weg.auspacken(t.Context(), ziel, weg.orte.Datenbank)
	if fehler != nil {
		t.Fatalf("auspacken: %v", fehler)
	}

	if err := os.WriteFile(filepath.Join(ausgepackt.Beiwerk, "7", "kunst.png"),
		[]byte("der Stand aus der Sicherung"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if fehler := weg.markeSchreiben(rueckspielstand{
		Quelle:  ziel,
		Beiwerk: ausgepackt.Beiwerk,
		Bilder:  BilderErsetzt,
	}); fehler != nil {
		t.Fatalf("Marke schreiben: %v", fehler)
	}

	// **Ein frischer Weg**, wie ihn der Start nach dem Kill baut.
	//
	// **Dass der Start ihn wirklich ruft, prüft `internal/app`**
	// (#1175) — hier steht, was er dann tut.
	frisch := wegAuf(t, weg.orte.Datenbank)
	frisch.orte.Bilder = func() (string, error) { return bilder, nil }

	frisch.RueckspielVollenden()

	roh, err := os.ReadFile(filepath.Join(bilder, "7", "kunst.png")) // #nosec G304 -- aus t.TempDir()
	if err != nil {
		t.Fatalf("das Bild fehlt: %v", err)
	}

	if string(roh) != "der Stand aus der Sicherung" {
		t.Errorf("der Bildertausch wurde nicht vollendet: %q", roh)
	}

	if _, err := os.Stat(filepath.Join(filepath.Dir(weg.orte.Datenbank),
		Rueckspielmarke)); !os.IsNotExist(err) {
		t.Error("die Marke liegt noch da")
	}
}

// **Ein abgebrochenes Leeren wird ebenso zu Ende geführt.**
//
// Eine Sicherung **ohne** Bilder beschreibt einen Stand, in dem keine da
// waren. Ohne die Absicht in der Marke ließe sich „noch nicht geschehen" nicht
// von „schon geschehen" unterscheiden: in beiden Fällen liegt im Auszug kein
// `beiwerk/`.
func TestEinAbgebrochenesLeerenWirdBeimStartVollendet(t *testing.T) {
	weg, _, bilder := mitBilderort(t)

	if err := os.MkdirAll(bilder, 0o700); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if err := os.WriteFile(filepath.Join(bilder, "alt.png"),
		[]byte("der alte Stand"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if fehler := weg.markeSchreiben(rueckspielstand{
		Bilder: BilderGeleert,
	}); fehler != nil {
		t.Fatalf("Marke schreiben: %v", fehler)
	}

	// **Ein frischer Weg**, wie ihn der Start nach dem Kill baut.
	//
	// **Dass der Start ihn wirklich ruft, prüft `internal/app`**
	// (#1175) — hier steht, was er dann tut.
	frisch := wegAuf(t, weg.orte.Datenbank)
	frisch.orte.Bilder = func() (string, error) { return bilder, nil }

	frisch.RueckspielVollenden()

	if _, err := os.Stat(bilder); !os.IsNotExist(err) {
		t.Errorf("das Bilderverzeichnis steht noch da: %v", err)
	}

	// **Und der alte Stand liegt beiseite** — dieselbe Zusage wie sonst.
	if _, err := os.Stat(filepath.Join(filepath.Dir(bilder), Beiwerkbeiseite,
		"alt.png")); err != nil {
		t.Errorf("der alte Bilderstand ist fort: %v", err)
	}
}

// **Ein Abbruch nach dem Bildertausch räumt nur die Marke weg.**
//
// Der Fall entsteht zwischen dem geglückten Tausch und dem Entfernen der
// Marke. Er darf nichts anfassen: die Arbeit ist getan.
func TestEinAbbruchNachDemTauschRaeumtNurDieMarkeWeg(t *testing.T) {
	weg, ort, bilder := mitBilderort(t)

	if err := os.MkdirAll(bilder, 0o700); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if err := os.WriteFile(filepath.Join(bilder, "neu.png"),
		[]byte("schon getauscht"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	// **Der Zwischenplatz steht, sein `beiwerk` ist fort** — genau das, was
	// `os.Rename` hinterlässt, wenn danach nur noch das Wegräumen fehlte.
	//
	// **Hier stand ein Pfad, den es nie gab** (`gibt-es-nicht/beiwerk`), und
	// der Kommentar daneben behauptete trotzdem, `os.Rename` habe ihn
	// weggenommen. Seit dem 2026-08-28 unterscheidet `bildertauschVollenden`
	// beides — und dieser Aufbau muss deshalb der sein, den er beschreibt
	// (#1170, Befund 2).
	platz := filepath.Join(ort, ".rueckspiel-abgebrochen")
	if err := os.MkdirAll(platz, 0o700); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if fehler := weg.markeSchreiben(rueckspielstand{
		Beiwerk: filepath.Join(platz, Beiwerkordner),
		Bilder:  BilderErsetzt,
	}); fehler != nil {
		t.Fatalf("Marke schreiben: %v", fehler)
	}

	// **Ein frischer Weg**, wie ihn der Start nach dem Kill baut.
	//
	// **Dass der Start ihn wirklich ruft, prüft `internal/app`**
	// (#1175) — hier steht, was er dann tut.
	frisch := wegAuf(t, weg.orte.Datenbank)
	frisch.orte.Bilder = func() (string, error) { return bilder, nil }

	frisch.RueckspielVollenden()

	roh, err := os.ReadFile(filepath.Join(bilder, "neu.png")) // #nosec G304 -- aus t.TempDir()
	if err != nil {
		t.Fatalf("der getauschte Stand ist fort: %v", err)
	}

	if string(roh) != "schon getauscht" {
		t.Errorf("der Stand wurde angefasst: %q", roh)
	}

	if _, err := os.Stat(filepath.Join(filepath.Dir(weg.orte.Datenbank),
		Rueckspielmarke)); !os.IsNotExist(err) {
		t.Error("die Marke liegt noch da")
	}
}

// **Ohne Marke geschieht beim Start nichts.**
//
// Die Gegenprobe: ohne sie sagten die Läufe darüber auch dann ja, wenn der
// Start **jedes Mal** am Bilderverzeichnis arbeitete.
func TestOhneMarkeGeschiehtBeimStartNichts(t *testing.T) {
	weg, _, bilder := mitBilderort(t)

	if err := os.MkdirAll(bilder, 0o700); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if err := os.WriteFile(filepath.Join(bilder, "steht.png"),
		[]byte("unberührt"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	// **Ein frischer Weg**, wie ihn der Start nach dem Kill baut.
	//
	// **Dass der Start ihn wirklich ruft, prüft `internal/app`**
	// (#1175) — hier steht, was er dann tut.
	frisch := wegAuf(t, weg.orte.Datenbank)
	frisch.orte.Bilder = func() (string, error) { return bilder, nil }

	frisch.RueckspielVollenden()

	if _, err := os.Stat(filepath.Join(bilder, "steht.png")); err != nil {
		t.Errorf("der Start hat das Bilderverzeichnis angefasst: %v", err)
	}

	if _, err := os.Stat(filepath.Join(filepath.Dir(bilder),
		Beiwerkbeiseite)); !os.IsNotExist(err) {
		t.Error("der Start hat etwas beiseitegelegt")
	}
}

// **Das Zurückspielen selbst legt die Marke an — und räumt sie weg.**
//
// Die Läufe darüber schreiben sie von Hand; ohne diesen hier bliebe offen, ob
// sie im Betrieb je entsteht. Genau diese Lücke hat schon zweimal einen ganzen
// Weg ohne Aufrufer stehen lassen (#852, #1013).
func TestDasZurueckspielenLegtDieMarkeAnUndRaeumtSieWeg(t *testing.T) {
	weg, ort, bilder := mitBilderort(t)

	if err := os.MkdirAll(bilder, 0o700); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if err := os.WriteFile(filepath.Join(bilder, "kunst.png"),
		[]byte("ein Bild"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	ziel := gesichert(t, weg, filepath.Join(ort, "gut"+sicherung.Endung))

	// **Nachgesehen im Augenblick des Bildertauschs.** Da ist die Datenbank
	// schon getauscht, und genau dann muss die Marke liegen — sonst hinge der
	// gemischte Stand in der Luft.
	var lagBeimTausch bool

	weg.orte.Bilder = func() (string, error) {
		if _, err := os.Stat(weg.markenpfad()); err == nil {
			lagBeimTausch = true
		}

		return bilder, nil
	}

	if _, rueckfehler := weg.Zurueckspielen(t.Context(), ziel); rueckfehler != nil {
		t.Fatalf("zurückspielen: %v", rueckfehler)
	}

	if !lagBeimTausch {
		t.Error("beim Bildertausch lag keine Marke — ein Abbruch dort bliebe unbemerkt")
	}

	if _, err := os.Stat(weg.markenpfad()); !os.IsNotExist(err) {
		t.Errorf("die Marke blieb liegen: %v", err)
	}
}

// **Und eine nackte Datenbank legt keine an.**
//
// Sie hat **einen** Schritt, und der ist für sich unteilbar. Eine Marke
// darüber schickte den nächsten Start auf eine Arbeit, die es nicht gibt.
func TestEineNackteDatenbankLegtKeineMarkeAn(t *testing.T) {
	weg, ort, bilder := mitBilderort(t)

	ausgepackt := t.TempDir()
	if _, err := sicherung.Entpacken(
		gesichert(t, weg, filepath.Join(ort, "gut"+sicherung.Endung)),
		ausgepackt); err != nil {
		t.Fatalf("die alte Gestalt herstellen: %v", err)
	}

	var lag bool

	weg.orte.Bilder = func() (string, error) {
		if _, err := os.Stat(weg.markenpfad()); err == nil {
			lag = true
		}

		return bilder, nil
	}

	if _, rueckfehler := weg.Zurueckspielen(t.Context(), filepath.Join(ausgepackt, Bestandsname)); rueckfehler != nil {
		t.Fatalf("zurückspielen: %v", rueckfehler)
	}

	if lag {
		t.Error("eine nackte Datenbank legte eine Marke an")
	}

	if _, err := os.Stat(weg.markenpfad()); !os.IsNotExist(err) {
		t.Errorf("es liegt eine Marke da: %v", err)
	}
}
