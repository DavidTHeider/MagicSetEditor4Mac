package schriften

import (
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// MitPaketen erweitert den Suchweg um die Verzeichnisse bezogener Pakete.
//
// # Warum nach den Systemorten und nicht davor
//
// **Was jemand selbst installiert hat, ist die Schrift, die er meint** (#140).
// Eine Paketschrift gleichen Namens ersetzte sie sonst, ohne dass jemand es
// angeordnet hätte — und der Unterschied fiele erst an einer Karte auf, die
// anders aussieht als vorher.
//
// Die Reihenfolge steht nicht in einem Kommentar, sondern in der Bauweise:
// `datei` geht zuerst die Orte durch und erst danach die Pakete.
//
// # Warum der ganze Baum und nicht ein festes Unterverzeichnis
//
// `M15-Magic-Pack` führt seine Schriften unter `Magic - Fonts/` — **neben** den
// Paketen, nicht in ihnen. Ein anderes Paket legt sie woandershin. Ein fester
// Name wäre eine Annahme über fremde Sammlungen; gesucht wird deshalb, wo sie
// liegen.
//
// # Was ausdrücklich **nicht** geschieht
//
// Die Schriften werden **nicht ins System kopiert**. Das wäre eine Installation,
// die niemand angeordnet hat, und sie überlebte die Deinstallation der
// Anwendung.
//
// # Was unter einer früheren Wurzel liegt, wird nicht noch einmal gegangen
//
// Der Gang über eine Wurzel kommt an allem vorbei, was darunter liegt; eine
// zweite Wurzel dort fände nichts, was die erste nicht schon fand, und auch
// nichts früher — sie steht ja dahinter. **Gemessen am 2026-10-02**
// (#2005):
// das Zeichenwerk gab `pakete`, `pakete` und `pakete/…/data`, und jede
// Schrift, die nicht unter ihrem Dateinamen liegt, kostete drei Gänge über
// das Pack statt einem.
//
// **Ihr Stand zählt trotzdem** ([Suche.paketstand]): ein Paket, das jemand
// unmittelbar in die tiefere Wurzel legt, ändert deren Zeit und nicht die
// der Wurzel darüber.
func (s *Suche) MitPaketen(wurzeln ...string) *Suche {
	s.paketwurzeln = s.paketwurzeln[:0]
	s.standwurzeln = s.standwurzeln[:0]

	for _, wurzel := range wurzeln {
		if wurzel == "" || slices.Contains(s.standwurzeln, wurzel) {
			continue
		}

		s.standwurzeln = append(s.standwurzeln, wurzel)

		if !s.schonGegangen(wurzel) {
			s.paketwurzeln = append(s.paketwurzeln, wurzel)
		}
	}

	return s
}

// schonGegangen sagt, ob der Gang über eine frühere Wurzel an `ort`
// vorbeikommt.
//
// **Nicht über einen Verweis hinweg**: `filepath.WalkDir` folgt keinem, und
// eine Wurzel, die dem Namen nach darunter liegt, aber hinter einem Verweis,
// erreichte der Gang darüber nie.
func (s *Suche) schonGegangen(ort string) bool {
	for _, wurzel := range s.paketwurzeln {
		rest, err := filepath.Rel(wurzel, ort)
		if err != nil || rest == ".." || strings.HasPrefix(rest, ".."+string(filepath.Separator)) {
			continue
		}

		if ohneVerweis(wurzel, rest) {
			return true
		}
	}

	return false
}

// ohneVerweis sagt, ob jede Stufe von `wurzel` nach `rest` ein Verzeichnis ist.
func ohneVerweis(wurzel, rest string) bool {
	stufe := wurzel

	for _, teil := range strings.Split(rest, string(filepath.Separator)) {
		stufe = filepath.Join(stufe, teil)

		angaben, err := os.Lstat(stufe)
		if err != nil || !angaben.IsDir() {
			return false
		}
	}

	return true
}

// paketdatei sucht eine Schrift in den bezogenen Paketen.
//
// **Bei jedem Fehlschlag neu gegangen, nicht einmal beim Start.** Ein Paket, das
// jemand während der Sitzung bezieht, wäre sonst bis zum Neustart unsichtbar —
// und genau darauf folgt in der Regel der erste Blick auf eine Karte.
func (s *Suche) paketdatei(gesucht string) (ort, datei string, gefunden bool) {
	for _, wurzel := range s.paketwurzeln {
		if ort, datei, gefunden = s.paketdateiUnter(wurzel, gesucht); gefunden {
			return ort, datei, true
		}
	}

	return "", "", false
}

// paketdateiUnter sucht sie unter **einer** Wurzel.
func (s *Suche) paketdateiUnter(wurzel, gesucht string) (ort, datei string, gefunden bool) {
	//nolint:errcheck // der Rückruf gibt nur nil und fs.SkipAll — ein Fehler kann nicht herauskommen
	_ = filepath.WalkDir(wurzel, func(pfad string, eintrag fs.DirEntry, err error) error {
		if err != nil || eintrag.IsDir() || gefunden {
			// Ein Verzeichnis, das sich nicht lesen lässt, ist kein Fehler: die
			// Wurzel gehört den Nutzenden, und ein Bezug kann halb daliegen.
			return nil //nolint:nilerr // siehe darüber
		}

		if !passendeEndung(strings.ToLower(filepath.Ext(eintrag.Name()))) {
			return nil
		}

		if strings.ToLower(ohneEndung(eintrag.Name())) != gesucht {
			return nil
		}

		ort, datei, gefunden = filepath.Dir(pfad), eintrag.Name(), true

		return fs.SkipAll
	})

	return ort, datei, gefunden
}

// paketnamensindex baut die Zuordnung „eigener Name → Datei" über die Pakete.
//
// **Getrennt von [Suche.namensindex]**, weil die Paketwurzel ein ganzer Baum
// ist und die Orte eine Liste von Verzeichnissen. Der Preis ist derselbe: jede
// Schriftdatei wird einmal geöffnet, und deshalb geschieht es erst, wenn beide
// Dateinamenswege nichts gefunden haben.
//
// **Einmal je Stand der Wurzel**, nicht einmal je Sitzung. Ein Lauf öffnet
// jede Schriftdatei; ihn bei jedem Fehlschlag zu wiederholen kostete über 189
// Dateien spürbar. Gemerkt wird er deshalb — aber **an den Stand gebunden**.
//
// # Warum nicht einmal je Sitzung
//
// Hier stand: „Ein Paket, das während der Sitzung dazukommt, wird über den
// Dateinamensweg trotzdem sofort gefunden." **Das gilt nur, wenn der
// Schriftname der Dateiname ist.** `magic-m15.mse-style` verlangt
// `Beleren Bold`, und die Datei heißt `beleren-bold_P1.01.ttf` — genau der
// Fall aus #502.
//
// Damit stand ein **leerer** Index für die ganze Sitzung fest, sobald jemand
// vor dem Bezug einmal eine Karte geöffnet hatte. Seit
// #263 ist
// genau das der vorgesehene Weg: starten, hinsehen, beziehen, wieder
// hinsehen — und die Karte blieb bis zum Neustart unbeschriftet. Der
// Zweitblick über Phase G hat es gefunden.
//
// # Woran der Stand hängt
//
// An der **Änderungszeit der Wurzel**. Sie trägt diese Zusage nur, weil beide
// Schreibwege ein Paket **fertig** an seinen Platz bringen: `Beziehen` und
// `kopierePaket` bauen es in einem vorläufigen Verzeichnis und benennen es um.
// Damit ändert sich die Wurzel **zuletzt**.
//
// Der Nachlauf über Phase G hat gefunden, dass das für die Übernahme nicht
// galt: `os.CopyFS` legte das Zielverzeichnis zuerst an — die Wurzel änderte
// sich also **vor** den Dateien darunter —, und eine Schriftsuche im Fenster
// dazwischen merkte sich einen halben Bestand für die ganze Sitzung. Genau der
// Weg, gegen den dieser Stand geschrieben wurde.
//
// **Was er weiterhin nicht sieht**: eine Änderung **innerhalb** eines schon
// vorhandenen Pakets, von jemand anderem gemacht — eine Schriftdatei, die
// während der Sitzung im Finder dorthin gelegt wird. Das ist die Grenze dieser
// Zusage, und sie steht hier, statt unausgesprochen zu bleiben. Sie zu
// schließen kostete einen Lauf über die ganze Wurzel je Fehlschlag: **93 ms**
// über die 30.220 Dateien des installierten Packs, gemessen am 2026-08-19.
func (s *Suche) paketnamensindex() map[string]string {
	stand := s.paketstand()

	if s.paketnamen != nil && s.paketzeit == stand {
		return s.paketnamen
	}

	s.paketnamen = map[string]string{}
	s.paketzeit = stand

	for _, wurzel := range s.paketwurzeln {
		//nolint:errcheck // der Rückruf gibt nur nil — ein Fehler kann nicht herauskommen
		_ = filepath.WalkDir(wurzel, func(pfad string, eintrag fs.DirEntry, err error) error {
			if err != nil || eintrag.IsDir() {
				return nil //nolint:nilerr // ein unlesbares Verzeichnis ist kein Fehler
			}

			if passendeEndung(strings.ToLower(filepath.Ext(eintrag.Name()))) {
				s.merkeInPaketen(pfad)
			}

			return nil
		})
	}

	return s.paketnamen
}

// merkeInPaketen trägt eine Paketschrift unter ihrem eigenen Namen ein.
func (s *Suche) merkeInPaketen(pfad string) {
	// Der Pfad stammt aus dem Durchlauf der Paketwurzel, nicht aus einer
	// Eingabe — dieselbe Lage wie in [Suche.Suche].
	s.eintragen(s.paketnamen, pfad)
}

// paketstand gibt die Änderungszeiten der Paketwurzeln als Zeichenkette.
//
// **Leer heißt „unbekannt" und nicht „unverändert".** Ein Verzeichnis, das
// sich nicht lesen lässt, wird bei jedem Fehlschlag neu gelaufen — teurer,
// aber nie falsch. Der umgekehrte Fehler kostete die Schrift bis zum Neustart.
//
// **Alle Wurzeln zusammen**: eine allein sagte nichts über eine zweite, die
// sich geändert hat.
func (s *Suche) paketstand() string {
	var stand strings.Builder

	for _, wurzel := range s.standwurzeln {
		angaben, err := os.Stat(wurzel)
		if err != nil {
			return ""
		}

		stand.WriteString(angaben.ModTime().String())
	}

	return stand.String()
}
