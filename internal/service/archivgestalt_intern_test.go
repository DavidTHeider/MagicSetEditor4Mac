// **Was ein Archiv sein muss, damit es zurückgespielt wird** — die Befunde 3
// und 4 des Zweitblicks über
// #1162.
//
// Beide haben dieselbe Gestalt: das Manifest sagt etwas, und niemand hört zu.
// Einmal die Fassung, einmal die Art des Eintrags `beiwerk`.
package service

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sicherung"
)

// **Eine Sicherung aus einer neueren Fassung wird abgewiesen — an ihrer
// eigenen Angabe.**
//
// # Der Befund
//
// `Angaben.Fassung` steht im Manifest, und der Kommentar daneben sagt
// ausdrücklich: „damit ein Rückspielweg entscheiden kann, ob er eine Sicherung
// aus einer **neueren** Fassung annimmt". Entschieden hat das niemand — der
// Aufrufer verwarf die Beschreibung mit `_`.
//
// Geprüft wurden allein die Migrationen im Bestand. Eine Sicherung einer
// künftigen Fassung, deren Schema sich **nicht** geändert hat, deren
// Archivgestalt aber schon, kam damit durch.
func TestEineSicherungAusEinerNeuerenFassungWirdAbgewiesen(t *testing.T) {
	weg, ort, _ := mitBilderort(t)

	ziel := gesichert(t, weg, filepath.Join(ort, "kuenftig"+sicherung.Endung))

	umschreiben(t, ziel, "99.0.0")

	_, rueckfehler := weg.Zurueckspielen(t.Context(), ziel)
	if rueckfehler == nil {
		t.Fatal("die Sicherung aus der Zukunft wurde angenommen")
	}

	if !strings.Contains(rueckfehler.Error(), "99.0.0") {
		t.Errorf("die Meldung nennt die Fassung nicht: %q", rueckfehler.Error())
	}

	pruefeFassungsfehler(t, rueckfehler)

	// **Und die Vorschau ebenso** — sie soll es sagen, bevor jemand
	// bestätigt, nicht danach.
	_, schaufehler := weg.Beschauen(t.Context(), ziel)
	if schaufehler == nil {
		t.Fatal("die Vorschau nahm die Sicherung aus der Zukunft an")
	}
}

// **Ein Archiv, dessen `beiwerk` eine Datei ist, wird abgewiesen.**
//
// # Der Befund
//
// Geprüft wurde die bloße **Existenz**. Eine Datei dieses Namens wurde als
// Bilderstand genommen und an den Ort des Bilderverzeichnisses gelegt: das
// Zurückspielen meldete Erfolg, die Kartenbilder fehlten, und jedes spätere
// Einsetzen scheiterte, weil unter einer Datei kein Ordner entsteht.
func TestEinBeiwerkDasEineDateiIstWirdAbgewiesen(t *testing.T) {
	weg, ort, _ := mitBilderort(t)

	inhalt := filepath.Join(ort, "keinordner.txt")
	if err := os.WriteFile(inhalt, []byte("kein Ordner"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	bestand := filepath.Join(ort, "bestandskopie.db")
	if err := weg.sicherer.Sichern(t.Context(), bestand); err != nil {
		t.Fatalf("sichern: %v", err)
	}

	ziel := filepath.Join(ort, "schief"+sicherung.Endung)
	if err := sicherung.Packen(ziel, []sicherung.Teil{
		{Name: Bestandsname, Quelle: bestand},
		{Name: Beiwerkordner, Quelle: inhalt},
	}, sicherung.Angaben{Fassung: Pruefassung}); err != nil {
		t.Fatalf("packen: %v", err)
	}

	_, rueckfehler := weg.Zurueckspielen(t.Context(), ziel)
	if rueckfehler == nil {
		t.Fatal("das Archiv mit einer Datei als beiwerk wurde angenommen")
	}

	if !strings.Contains(rueckfehler.Error(), Beiwerkordner) {
		t.Errorf("die Meldung nennt den Eintrag nicht: %q", rueckfehler.Error())
	}
}

// umschreiben setzt die Fassung im Manifest einer Sicherung neu.
func umschreiben(t *testing.T, pfad, fassung string) {
	t.Helper()

	offen := t.TempDir()

	beschreibung, fehler := sicherung.Entpacken(pfad, offen)
	if fehler != nil {
		t.Fatalf("auspacken: %v", fehler)
	}

	teile := make([]sicherung.Teil, 0, len(beschreibung.Teile))
	for _, teil := range beschreibung.Teile {
		teile = append(teile, sicherung.Teil{
			Name:   teil.Name,
			Quelle: filepath.Join(offen, filepath.FromSlash(teil.Name)),
		})
	}

	angaben := beschreibung.Angaben
	angaben.Fassung = fassung

	neben := pfad + ".neu"
	if fehler := sicherung.Packen(neben, teile, angaben); fehler != nil {
		t.Fatalf("packen: %v", fehler)
	}

	if fehler := os.Rename(neben, pfad); fehler != nil {
		t.Fatalf("an seinen Platz: %v", fehler)
	}
}

// **Eine Sicherung ohne lesbare Fassungsangabe wird abgewiesen.**
//
// # Warum hier nicht gilt, was für die nackte `.db` gilt
//
// Der erste Anlauf ließ sie durch, mit derselben Begründung wie dort: aus
// „weiß nichts darüber" wird kein „stammt aus der Zukunft". Für ein **Archiv**
// stimmt das nicht — dieses Format beginnt mit 0.3.0, und **jede**
// `.msesicherung`, die diese Anwendung je geschrieben hat, nennt ihre Fassung.
//
// Eine leere oder unlesbare Angabe kommt deshalb nicht von einer älteren
// Fassung, sondern von etwas anderem. Sie durchzulassen hieße, die Wache mit
// `fassung: "morgen"` zu umgehen, die `99.0.0` abweist (Befund 5 des Nachlaufs).
func TestEineSicherungOhneLesbareFassungWirdAbgewiesen(t *testing.T) {
	for name, fassung := range map[string]string{
		"ohne Angabe":   "",
		"unlesbar":      "morgen",
		"mit Buchstabe": "0.3.x",
	} {
		t.Run(name, func(t *testing.T) {
			weg, ort, _ := mitBilderort(t)
			weg.MitSicherungspruefer(gepruefteSicherung())

			ziel := gesichert(t, weg, filepath.Join(ort, "ohne"+sicherung.Endung))

			umschreiben(t, ziel, fassung)

			_, rueckfehler := weg.Zurueckspielen(t.Context(), ziel)
			if rueckfehler == nil {
				t.Fatal("die Sicherung ohne lesbare Fassung kam durch")
			}

			pruefeFassungsfehler(t, rueckfehler)
		})
	}
}

// **Die eigene Fassung kommt durch — und eine ältere ebenso.**
//
// Die Gegenprobe: ohne sie sagte der Lauf darüber auch dann ja, wenn **jede**
// Sicherung abgewiesen würde.
func TestDieEigeneUndEineAeltereFassungKommenDurch(t *testing.T) {
	for name, fassung := range map[string]string{
		"die eigene":  Pruefassung,
		"eine ältere": "0.1.0",
	} {
		t.Run(name, func(t *testing.T) {
			weg, ort, _ := mitBilderort(t)
			weg.MitSicherungspruefer(gepruefteSicherung())

			ziel := gesichert(t, weg, filepath.Join(ort, "gut"+sicherung.Endung))

			umschreiben(t, ziel, fassung)

			if _, rueckfehler := weg.Zurueckspielen(t.Context(), ziel); rueckfehler != nil {
				t.Errorf("die Sicherung wurde abgewiesen: %v", rueckfehler)
			}
		})
	}
}

// pruefeFassungsfehler sichert den **Schlüssel** zu, nicht nur den Satz.
//
// # Warum der Schlüssel und nicht die Meldung
//
// Das Fenster übersetzt über den Schlüssel und bevorzugt den Katalog vor der
// mitgereisten Go-Meldung. Ein Test, der nur den Satz liest, bleibt deshalb
// grün, während im Fenster etwas ganz anderes steht — genau das war am
// 2026-08-26 der Fall: der erste Anlauf umschloss den Fehler mit
// `domain.ErrConflict`, und die Oberfläche zeigte „Der Eintrag wurde
// zwischenzeitlich geändert. Bitte neu laden." (§3, „zwischen den Prüfebenen";
// Befund 6 des Nachlaufs).
//
// **Seit dem Umzug in den Dienst wird der Typ geprüft und nicht das DTO**
// (#1175): `app.Uebersetze` macht aus [domain.FremdeArchivfassung] den
// Schlüssel, und dass es ihn macht, prüft `internal/app`. Was hier zählt, ist
// dass dieser Fehler ankommt — und kein anderer.
func pruefeFassungsfehler(t *testing.T, fehler error) {
	t.Helper()

	var fremde *domain.FremdeArchivfassung
	if !errors.As(fehler, &fremde) {
		t.Fatalf("der Fehler ist keine fremde Archivfassung: %v", fehler)
	}

	if fremde.Hier != Pruefassung {
		t.Errorf("die Werte nennen die eigene Fassung nicht: %+v", fremde)
	}
}
