// **Die Fehlerwege der Rückspielmarke** (#1162, Werkstück 1).
//
// Sie brauchen ein Verzeichnis, das sich zusperren lässt — und deshalb einen
// **Abstand** zwischen Bestand und Bildern: die Marke liegt beim Bestand, der
// Tausch geschieht bei den Bildern. Ohne den Abstand träfe ein Riegel beides,
// und der Lauf prüfte den falschen Schritt (§10).
package service

import (
	"os"
	"path/filepath"
	"testing"
)

// **Lässt sich die Marke nicht schreiben, wird nicht getauscht.**
//
// Sie ist die Zusage, dass ein Abbruch nachgeholt wird. Ohne sie zu tauschen
// hieße, die Zusage stillschweigend fallen zu lassen.
func TestOhneMarkeWirdNichtGetauscht(t *testing.T) {
	bau, unter, _ := beiwerkAnwendung(t)
	weg := bau.weg

	ziel := gesichert(t, weg, filepath.Join(unter, "gut.msesicherung"))

	// **Ein Verzeichnis am Ort der Marke.** Ein zugesperrtes Elternverzeichnis
	// träfe den Zwischenplatz des Auspackens mit — der liegt daneben —, und
	// dann scheiterte der Lauf am falschen Schritt und wäre trotzdem grün
	// (§10).
	if err := os.Mkdir(weg.markenpfad(), 0o700); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	_, rueckfehler := weg.Zurueckspielen(t.Context(), ziel)
	if rueckfehler == nil {
		t.Fatal("es wurde getauscht, ohne die Marke schreiben zu können")
	}

	// **Und die Anwendung läuft weiter.** Der Weg hat sie geschlossen und
	// wieder geöffnet — genau das zeigt der Zähler, den die Fassade sonst im
	// Erfolg eines Aufrufs verbirgt.
	if bau.leben.geoeffnet != bau.leben.geschlossen {
		t.Errorf("geschlossen %d-mal, geöffnet %d-mal — die Anwendung steht ohne Datenbank da",
			bau.leben.geschlossen, bau.leben.geoeffnet)
	}
}

// **Eine unlesbare Marke hält den Start nicht auf.**
//
// Sie ist ein Befund über eine Nebendatei. Die Anwendung deswegen nicht
// starten zu lassen nähme den Nutzenden auch den Weg, es von Hand zu richten.
func TestEineUnlesbareMarkeHaeltDenStartNichtAuf(t *testing.T) {
	bau, _, bilder := beiwerkAnwendung(t)
	weg := bau.weg

	if err := os.WriteFile(weg.markenpfad(),
		[]byte("kein JSON"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	// **Ein frischer Weg**, wie ihn der Start nach dem Kill baut.
	//
	// **Dass der Start ihn wirklich ruft, prüft `internal/app`**
	// (#1175) — hier steht, was er dann tut.
	frisch := wegAuf(t, weg.orte.Datenbank)
	frisch.orte.Bilder = func() (string, error) { return bilder, nil }

	frisch.RueckspielVollenden()

	// **Und sie bleibt liegen** — was man nicht deuten kann, wirft man nicht
	// weg.
	if _, err := os.Stat(weg.markenpfad()); err != nil {
		t.Errorf("die unlesbare Marke wurde entfernt: %v", err)
	}
}

// **Kann die Vollendung nicht arbeiten, hält sie den Start ebenso wenig auf.**
func TestEineGescheiterteVollendungHaeltDenStartNichtAuf(t *testing.T) {
	bau, _, _ := beiwerkAnwendung(t)
	weg := bau.weg

	if fehler := weg.markeSchreiben(rueckspielstand{
		Bilder: BilderGeleert,
	}); fehler != nil {
		t.Fatalf("Marke schreiben: %v", fehler)
	}

	// **Ein frischer Weg**, wie ihn der Start nach dem Kill baut (#1175).
	frisch := wegAuf(t, weg.orte.Datenbank)
	frisch.orte.Bilder = func() (string, error) {
		return "", os.ErrPermission
	}

	frisch.RueckspielVollenden()

	// **Die Marke bleibt liegen** — die Arbeit steht noch aus.
	if _, err := os.Stat(weg.markenpfad()); err != nil {
		t.Errorf("die Marke wurde entfernt, obwohl nichts geschah: %v", err)
	}
}

// **Ein abgebrochenes Leeren, das schon geschehen war, räumt nur die Marke
// weg.**
//
// Der Gegenfall zum Ersetzen: dort sagt der fehlende Auszug, dass es getan
// ist; hier sagt es das fehlende Bilderverzeichnis.
func TestEinAbgebrochenesLeerenNachDemTauschRaeumtNurDieMarkeWeg(t *testing.T) {
	bau, unter, bilder := beiwerkAnwendung(t)
	weg := bau.weg

	if err := os.RemoveAll(bilder); err != nil {
		t.Fatalf("wegräumen: %v", err)
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

	if _, err := os.Stat(weg.markenpfad()); !os.IsNotExist(err) {
		t.Error("die Marke liegt noch da")
	}

	// **Und es wurde nichts beiseitegelegt** — es war ja nichts da.
	if _, err := os.Stat(filepath.Join(unter, Beiwerkbeiseite)); !os.IsNotExist(err) {
		t.Error("es wurde etwas beiseitegelegt, obwohl nichts dastand")
	}
}

// **Bleibt die Marke nach der Vollendung liegen, steht es im Protokoll.**
//
// Der Fall: die Arbeit ist getan, das Wegräumen scheitert. Er darf den Start
// nicht aufhalten — und beim nächsten Mal findet die Vollendung „schon
// geschehen" vor und tut nichts.
func TestEineLiegengebliebeneMarkeHaeltDenStartNichtAuf(t *testing.T) {
	bau, _, bilder := beiwerkAnwendung(t)
	weg := bau.weg

	if fehler := weg.markeSchreiben(rueckspielstand{
		Bilder: BilderGeleert,
	}); fehler != nil {
		t.Fatalf("Marke schreiben: %v", fehler)
	}

	// **Zugesperrt wird das Verzeichnis der Marke**, nicht das der Bilder:
	// die Vollendung soll arbeiten können und erst am Wegräumen scheitern.
	zusperren(t, filepath.Dir(weg.orte.Datenbank))

	frisch := wegAuf(t, weg.orte.Datenbank)
	frisch.orte.Bilder = func() (string, error) { return bilder, nil }

	// **Die Vollendung läuft durch** — das ist die Zusage: ein Befund über
	// eine Nebendatei hält den Start nicht auf. Dass der Start sie ruft,
	// prüft `internal/app` (#1175).
	frisch.RueckspielVollenden()

	// **Die Arbeit ist getan** — sie geschieht im Bilderverzeichnis, und das
	// ist offen.
	if _, err := os.Stat(bilder); !os.IsNotExist(err) {
		t.Errorf("die Vollendung hat nicht gearbeitet: %v", err)
	}

	// **Und die Marke liegt noch da**, weil das Wegräumen scheiterte. Beim
	// nächsten Start findet die Vollendung „schon geschehen" vor und tut
	// nichts — das ist der Grund, warum ein Liegenbleiben tragbar ist.
	if _, err := os.Stat(weg.markenpfad()); err != nil {
		t.Errorf("die Marke wurde entfernt, obwohl das Verzeichnis zu war: %v", err)
	}
}

// **Und bleibt sie nach einem Zurückspielen liegen, ebenso.**
func TestEineLiegengebliebeneMarkeNachDemRueckspielen(t *testing.T) {
	bau, unter, bilder := beiwerkAnwendung(t)
	weg := bau.weg

	ziel := gesichert(t, weg, filepath.Join(unter, "gut.msesicherung"))

	// **Zugesperrt wird im Augenblick des Bildertauschs** — davor muss die
	// Marke entstehen können, danach nicht mehr weggehen.
	weg.orte.Bilder = func() (string, error) {
		zusperren(t, filepath.Dir(weg.orte.Datenbank))

		return bilder, nil
	}

	_, rueckfehler := weg.Zurueckspielen(t.Context(), ziel)
	if rueckfehler != nil {
		t.Fatalf("zurückspielen: %v", rueckfehler)
	}

	if _, err := os.Stat(weg.markenpfad()); err != nil {
		t.Errorf("die Marke wurde entfernt, obwohl das Verzeichnis zu war: %v", err)
	}
}
