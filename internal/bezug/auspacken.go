// Das sichere Auspacken eines fremden Archivs.
//
// **Eine eigene Datei, weil sie eine eigene Frage beantwortet** (§2): nicht
// *woher* ein Paket kommt, sondern *wie* es auf die Platte gelangt, ohne dass
// der Absender bestimmt, wohin. `bezug.go` ging über die Zeilengrenze, als die
// Größenprüfung dazukam
// (#898).
//
// Hier stehen deshalb die Entscheidungen, die ein **fremdes** Archiv betreffen:
// wohin ein Eintrag darf, was ein Link darf, was ein Ausführungsbit bedeutet.

package bezug

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// auspacken schreibt den Inhalt eines Archivs nach `ziel`.
//
// Gibt die Namen der Dateien zurück, die dabei übergangen wurden.
// **`leser` und nicht `[]byte`** (#1163, Befund Z7-01): ein Archiv wird
// geströmt, nicht gesammelt. Der Vorgänger nahm bis zu zwei Gigabyte entgegen
// — und der Kommentar an dieser Stelle handelte davon, wie man diese zwei
// Gigabyte **nicht noch einmal** kopiert.
func auspacken(leser io.Reader, unterverzeichnis, ziel string) ([]string, error) {
	entpacker, err := gzip.NewReader(leser)
	if err != nil {
		return nil, fmt.Errorf("Archiv auspacken: %w", err)
	}

	defer func() { _ = entpacker.Close() }() //nolint:errcheck // dito

	ausgepackt := tar.NewReader(entpacker)

	var (
		uebergangen []string
		gesamt      int64
		gezaehlt    int
	)

	for {
		kopf, err := ausgepackt.Next()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("Archiv auspacken: %w", err)
		}

		// **Gezählt wird **jeder** Eintrag, nicht jeder ausgepackte.**
		//
		// Der erste Anlauf zählte hinter der Auswahl: Verzeichnisse,
		// Sondertypen, Einträge außerhalb des gewählten Unterverzeichnisses
		// und übergangene Programme kamen gar nicht bis zum Zähler. Ein Tar
		// mit Millionen Verzeichnissen ging damit vollständig an der Grenze
		// vorbei — und kostete trotzdem Laufzeit für jeden davon
		// (Zweitblick über Phase 3.2).
		gezaehlt++
		if gezaehlt > hoechstensEintraege {
			return nil, domain.NewFieldError("archiv", domain.SchluesselValidierung,
				"Dieses Archiv führt mehr als %d Einträge. So viele hat kein "+
					"Vorlagenpaket.", hoechstensEintraege)
		}

		name, urteil, err := zielname(kopf, unterverzeichnis, ziel)
		if err != nil {
			return nil, err
		}

		switch urteil {
		case uebergehen:
			// **Nur was zum Paket gehört, wird gemeldet.** Ein Programm neben
			// dem Paketverzeichnis ist nicht übergangen worden — es war nie
			// Teil dessen, was jemand angefordert hat.
			//
			// **Mit seinem Pfad im Paket** (#2049): der Echtlauf meldete
			// fünfmal „build.bat", ohne zu sagen, welche. Abgeleitet aus dem
			// **geprüften** Weg, nicht aus dem Archivkopf.
			uebergangen = append(uebergangen, filepath.ToSlash(
				strings.TrimPrefix(name, filepath.Clean(ziel)+string(filepath.Separator))))

			continue

		case nichtsDamit:
			continue
		}

		// **Der Deckel ist, was vom Budget übrig ist** — und nicht die
		// Einzelgrenze.
		//
		// Der erste Anlauf gab jedem Eintrag die vollen zwei Gigabyte und
		// zählte **danach**: bei knapp vier Gigabyte auf der Platte durfte
		// der nächste noch zwei schreiben, bevor die Abweisung kam. Die
		// zugesagten vier wurden damit fast sechs (Zweitblick über Phase 3.2).
		uebrig := hoechstensGesamt - gesamt

		geschrieben, err := schreiben(ausgepackt, kopf, name, min(hoechstensEintrag, uebrig))
		if err != nil {
			return nil, err
		}

		// **Gezählt wird, was **geschrieben** wurde, nicht was der Kopf
		// ansagt.** Ein Archiv, das eine kleine Größe angibt und mehr liefert,
		// ginge sonst an dieser Grenze vorbei — und genau das ist die Gestalt
		// einer Kompressionsbombe.
		gesamt += geschrieben

		// **Und die Meldung nennt die Grenze, die wirklich griff.** Sie steht
		// hier und nicht in `schreiben`, weil erst hier beide bekannt sind:
		// „der Eintrag ist zu groß" und „zusammen ist es zu viel" sind zwei
		// verschiedene Auskünfte, und die falsche schickt jemanden auf die
		// Suche nach der falschen Datei.
		if geschrieben > uebrig {
			return nil, domain.NewFieldError("archiv", domain.SchluesselValidierung,
				"Die Einträge dieses Archivs wiegen zusammen über %d Byte. "+
					"So viel hat kein Vorlagenpaket.", hoechstensGesamt)
		}

		if geschrieben > hoechstensEintrag {
			return nil, domain.NewFieldError("archiv", domain.SchluesselValidierung,
				"Der Eintrag %q ist über %d Byte zu groß. Er wird nicht "+
					"abgeschnitten: ein halbes Paket sieht aus wie ein ganzes.",
				kopf.Name, hoechstensEintrag)
		}
	}

	return uebergangen, nil
}

// urteil sagt, was mit einem Eintrag geschieht.
type urteil int

const (
	// auspacken: der Eintrag gehört ins Ziel.
	einpacken urteil = iota

	// uebergehen: er gehört zum Paket, kommt aber nicht mit — und das wird
	// gesagt.
	uebergehen

	// nichtsDamit: er gehört nicht zum angeforderten Paket. Kein Befund: er war
	// nie Teil dessen, was jemand wollte.
	nichtsDamit
)

// zielname sagt, wohin ein Eintrag gehört und ob er mitkommt.
func zielname(kopf *tar.Header, unterverzeichnis, ziel string) (string, urteil, error) {
	// **Das Wurzelverzeichnis des Archivs fällt weg.** GitHub packt alles unter
	// ein Verzeichnis, dessen Name den Stand trägt — wer es behielte, hätte den
	// Stand für immer im Pfad.
	_, rest, gefunden := strings.Cut(kopf.Name, "/")
	if !gefunden {
		return "", nichtsDamit, nil
	}

	if unterverzeichnis != "" {
		innen, darunter := strings.CutPrefix(rest, unterverzeichnis+"/")
		if !darunter {
			return "", nichtsDamit, nil
		}

		rest = innen
	}

	// **Ein leerer Rest ist das Verzeichnis selbst, kein Ausbruch.**
	//
	// GitHub schreibt in sein Archiv zuerst einen Verzeichniseintrag für die
	// Wurzel — `M15-Magic-Pack-3f7f8ce…/`, mit Schrägstrich und ohne Inhalt.
	// Hinter dem ersten Schrägstrich steht dann nichts mehr, und die Wegprüfung
	// darunter verglich das Ziel mit sich selbst: `ziel` beginnt nicht mit
	// `ziel/`, also „zeigt hinaus". Damit scheiterte der Bezug an **jedem**
	// GitHub-Archiv — an genau dem Fall, für den er gebaut ist.
	//
	// Dasselbe gilt für den Eintrag des angeforderten Unterverzeichnisses.
	// Anzulegen ist hier nichts: `schreiben` legt die Verzeichnisse einer Datei
	// ohnehin an.
	if rest == "" {
		return "", nichtsDamit, nil
	}

	// **Der Weg wird geprüft, bevor er benutzt wird** (#56). Ein `..` im Archiv
	// oder ein symbolischer Verweis nach oben schriebe sonst dorthin, wo
	// niemand es angeordnet hat.
	sauber := filepath.Clean(filepath.Join(ziel, filepath.FromSlash(rest)))
	if !strings.HasPrefix(sauber, filepath.Clean(ziel)+string(filepath.Separator)) {
		return "", nichtsDamit, domain.NewFieldError("archiv", domain.SchluesselValidierung,
			"Der Eintrag %q zeigt aus dem Zielverzeichnis hinaus.", kopf.Name)
	}

	switch {
	case kopf.Typeflag == tar.TypeSymlink || kopf.Typeflag == tar.TypeLink:
		return "", nichtsDamit, domain.NewFieldError("archiv", domain.SchluesselValidierung,
			"Der Eintrag %q ist ein Verweis. Verweise werden nicht ausgepackt: "+
				"wohin sie zeigen, entscheidet das Archiv und nicht wir.", kopf.Name)

	case kopf.Typeflag != tar.TypeReg:
		return "", nichtsDamit, nil

	case ausfuehrbar(kopf):
		return sauber, uebergehen, nil
	}

	return sauber, einpacken, nil
}

// ausfuehrbar sagt, ob ein Eintrag ausführbarer Code ist.
//
// **Zwei Kennzeichen, und beide zählen**: die Endung und das Ausführungsrecht.
// Eine `.exe` trägt unter Unix oft kein Ausführungsrecht, und ein Skript ohne
// Endung trägt es. Wer nur eines prüft, packt die andere Hälfte aus.
func ausfuehrbar(kopf *tar.Header) bool {
	if slices.Contains(nichtAuspacken, strings.ToLower(path.Ext(kopf.Name))) {
		return true
	}

	return kopf.Mode&0o111 != 0
}

// zielOeffnen legt die Datei eines Eintrags an.
//
// **Eine Naht für die Prüfung** (BK-06c, #2132): einen Fehler, den das
// Dateisystem erst beim Schließen meldet, gibt eine gewöhnliche Platte nicht
// her.
//
// **Der Weg ist geprüft, bevor er hierher kommt**: `zielname` hält ihn gegen
// das Zielverzeichnis und weist alles ab, was hinausführt — `..` wie
// symbolische Verweise. Ohne diese Prüfung wäre der Aufruf gefährlich; mit
// ihr ist er der Zweck.
var zielOeffnen = func(name string) (io.WriteCloser, error) {
	return os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600) // #nosec G304 -- siehe darüber
}

// schreiben legt einen Eintrag an und gibt zurück, wie viel er wog.
//
// **`deckel` ist, was noch erlaubt ist** — die Einzelgrenze oder der Rest des
// Gesamtbudgets, je nachdem, was kleiner ist. Ein Byte darüber ist ein Fehler
// und kein Abschneiden.
func schreiben(leser *tar.Reader, kopf *tar.Header, name string, deckel int64) (gewicht int64, fehler error) {
	if err := os.MkdirAll(filepath.Dir(name), 0o750); err != nil {
		return 0, fmt.Errorf("Archiv auspacken: %w", err)
	}

	datei, err := zielOeffnen(name)
	if err != nil {
		return 0, fmt.Errorf("Archiv auspacken: %w", err)
	}

	// **Auch das Schließen kann scheitern, und dann ist nichts installiert**
	// (BK-06c, #2132). Hier stand „der Schreibfehler kommt schon aus
	// io.Copy" — ein Dateisystem darf ihn aber erst beim Schließen melden.
	// Scheitert beides, bleiben beide Gründe stehen.
	defer func() {
		if zu := datei.Close(); zu != nil {
			gewicht = 0
			fehler = errors.Join(fehler, fmt.Errorf("Archiv auspacken (%s) abschließen: %w", kopf.Name, zu))
		}
	}()

	// **Ein Byte mehr als erlaubt wird gelesen** — nur so unterscheidet sich
	// „genau an der Grenze" von „darüber". Ein Archiv kann eine Größe angeben
	// und mehr liefern; gezählt wird deshalb, was ankommt.
	geschrieben, err := io.Copy(datei, io.LimitReader(leser, deckel+1))
	if err != nil {
		return 0, fmt.Errorf("Archiv auspacken (%s): %w", kopf.Name, err)
	}

	// **Abschneiden wäre der stille Fehlschlag.** Bis Phase 3.2 endete `io.Copy`
	// hier ohne Fehler, und auf der Platte lag eine Datei, die aussah wie eine
	// ganze — sichtbar erst, wenn jemand die Karte zeichnet, und dann sucht er
	// den Fehler im Paket (#1163, Befund Z7-01).
	return geschrieben, nil
}
