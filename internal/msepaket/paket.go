package msepaket

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// ErrAusbruch meldet einen Verweis, der aus dem Paket herausführt.
//
// Ein Paket kommt von außen — heruntergeladen, weitergereicht, aus einer
// fremden Sammlung. Ein Eintrag `../../../.ssh/id_ed25519` darin ist kein
// theoretischer Fall, sondern der Grund für diese Prüfung.
var ErrAusbruch = errors.New("der Verweis führt aus dem Paket heraus")

// Befund ist ein benannter Vorbehalt, kein Fehler (AGENTS.md §8).
//
// Ein Paket mit acht lesbaren und einer unlesbaren Abhängigkeit wird mit acht
// Abhängigkeiten und einem Befund geöffnet, nicht abgewiesen. Fremde Daten sind
// der Normalfall.
type Befund struct {
	Stelle  Stelle
	Meldung string
}

func (b Befund) String() string {
	return fmt.Sprintf("%s:%d:%d: %s", b.Stelle.Datei, b.Stelle.Zeile, b.Stelle.Spalte, b.Meldung)
}

// Abhaengigkeit ist ein Paket, das dieses Paket voraussetzt.
type Abhaengigkeit struct {
	Paket   string
	Version Datum
	Stelle  Stelle
}

// Paket ist ein geöffnetes MSE-Paket — ein Verzeichnis oder ein ZIP-Archiv.
type Paket struct {
	Pfad            string
	Art             Art
	Spiel           string
	KurzerName      string
	VollerName      string
	Version         Datum
	Abhaengigkeiten []Abhaengigkeit

	// Gruppe ist `installer group:` — der Pfad, unter dem MSE2 dieses Paket in
	// seinem Paketbaum zeigt, mit `/` als Trenner.
	//
	// **Leer ist erlaubt und heißt: keine Gruppe.** MSE setzt dann „custom"
	// (`package_update_list.cpp`, Zeile 205); das ist ein Oberflächentext und
	// gehört nicht hierher, sondern in die Sprachkataloge (§9).
	Gruppe string

	// Rang ist `position hint:` — die Ordnung innerhalb einer Gruppe.
	//
	// Ohne Angabe [RangOhneHinweis].
	Rang int

	// Befunde sind die Vorbehalte, unter denen dieses Paket gelesen wurde.
	Befunde []Befund

	// KopfRoh sind die Bytes der Hauptdatei, wie sie dastanden.
	//
	// **Damit sie niemand zweimal lesen muss.** Der Import führt sie mit
	// ([0011](../../docs/entscheidungen/0011-die-quelldatei-wird-mitgefuehrt.md)),
	// und ein zweiter Griff in dasselbe Archiv wäre ein zweiter Fehlerweg für
	// eine Datei, die gerade gelesen wurde — also ein Zweig ohne Fall (§3).
	KopfRoh []byte

	// Kopf ist der Datenbaum der Hauptdatei.
	Kopf *Baum

	dateien fs.FS

	// schliesser gibt das Archiv frei; bei einem Verzeichnis ist er nil.
	//
	// Als `io.Closer` und nicht als `*zip.ReadCloser`: so lässt sich der
	// Fehlerpfad darunter prüfen. Ein Archiv, das sich nicht schließen lässt,
	// kommt auf einer gesunden Platte nicht zustande — im Betrieb aber schon,
	// und dann entscheidet dieser Zweig, ob die Anwendung es meldet.
	schliesser io.Closer
}

// Oeffne öffnet ein Paket, gleich ob Verzeichnis oder ZIP-Archiv.
//
// Beides ist derselbe Gegenstand: ein `.mse-set` ist im Betrieb meist ein
// Archiv, ein Vorlagenpaket meist ein Verzeichnis. Wer sie verschieden
// behandelt, schreibt jede Auswertung zweimal — und die seltener benutzte
// Fassung fällt zuerst um.
func Oeffne(pfad string) (*Paket, error) {
	paket := &Paket{Pfad: pfad, Art: ArtAusName(pfad)}

	dateien, schliesser, err := paketbaum(pfad)
	if err != nil {
		return nil, err
	}

	paket.dateien = dateien
	paket.schliesser = schliesser

	if fehler := paket.kopfLesen(); fehler != nil {
		// Der Schließfehler wird angehängt, nicht verworfen: der eigentliche
		// Befund ist der beim Lesen, aber ein Archiv, das sich auch nicht
		// schließen lässt, ist die zweite Hälfte der Nachricht.
		return nil, errors.Join(fehler, paket.Schliessen())
	}

	return paket, nil
}

// paketbaum öffnet ein Paket als Dateibaum — Verzeichnis wie Archiv.
//
// **Eine Stelle für beide Formen**, denn jeder Leser dieses Pakets braucht sie:
// [Oeffne] liest daraus den Kopf, [Bildquelle] die Bilddateien. Zwei Kopien
// dieser Fallunterscheidung liefen an dem Tag auseinander, an dem eine von
// ihnen eine dritte Form lernte.
//
// **Ein Verzeichnis wird als [os.Root] geöffnet**, nicht über `os.DirFS` —
// FMT-06.AK-01 (#1695).
// `os.DirFS` hält einen Symlink nach draußen nicht auf, so steht es in seiner
// eigenen Dokumentation; eine Wurzel tut es, und ein Symlink, der im Paket
// bleibt, bleibt lesbar. Die Wurzel hält das Verzeichnis offen und ist deshalb
// der Schließer, wie das Archiv.
func paketbaum(pfad string) (fs.FS, io.Closer, error) {
	angaben, err := os.Stat(pfad)
	if err != nil {
		// **Eine fehlende Datei ist „nicht gefunden"**, und das gehört hierher
		// eingeordnet (#997).
		//
		// Ohne diese Zeile war ein Pfad, den es nicht gibt, im Fachkern **gar
		// nicht eingeordnet** — nur ein `*fs.PathError`. Die Flächen, die den
		// rohen Text zeigen, haben es verdeckt; der Dienst auf der
		// Rückschleife darf das nicht (§8) und meldete deshalb ein 500 ohne
		// Grund für etwas, das der Aufrufer sofort beheben könnte.
		//
		// **Nur `ErrNotExist`**, nicht jeder Statfehler: eine Datei ohne
		// Leserecht gibt es, und sie als „nicht gefunden" zu melden schickte
		// jemanden auf die Suche nach etwas, das dasteht.
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil, domain.MitGrund(domain.ErrNotFound, err,
				"das Paket %q ist nicht zu finden", pfad)
		}

		return nil, nil, fmt.Errorf("Paket %q öffnen: %w", pfad, err)
	}

	if angaben.IsDir() {
		wurzel, fehler := os.OpenRoot(pfad)
		if fehler != nil {
			return nil, nil, fmt.Errorf("Paket %q öffnen: %w", pfad, fehler)
		}

		baum := wurzel.FS()

		return baum, wurzel, nil
	}

	// **Nach der Ressourcenregel** (#1886): ein Archiv über der Grenze wird
	// abgewiesen, bevor sein Verzeichnis gelesen ist.
	archiv, schliesser, fehler := oeffneArchiv(pfad)
	if fehler != nil {
		return nil, nil, fmt.Errorf("Paket %q als Archiv öffnen: %w", pfad, fehler)
	}

	return archiv, schliesser, nil
}

// Schliessen gibt das Archiv frei. Bei einem Verzeichnis tut es nichts.
func (p *Paket) Schliessen() error {
	if p.schliesser == nil {
		return nil
	}

	archiv := p.schliesser
	p.schliesser = nil

	if err := archiv.Close(); err != nil {
		return fmt.Errorf("Paket %q schließen: %w", p.Pfad, err)
	}

	return nil
}

// Datei liest eine Datei aus dem Paket.
//
// Der Name ist ein **Paketpfad**: mit Schrägstrichen getrennt, gleich auf
// welchem Betriebssystem. Er wird über `path` behandelt, nie über `filepath` —
// unter macOS fällt der Unterschied nicht auf, unter Windows sofort
// (`docs/PORTIERUNG.md`).
func (p *Paket) Datei(name string) ([]byte, error) {
	sauber, err := paketpfad(name)
	if err != nil {
		return nil, err
	}

	inhalt, fehler := leseBegrenzt(p.dateien, sauber)
	if fehler != nil {
		return nil, fmt.Errorf("%s in %q lesen: %w", sauber, p.Pfad, fehler)
	}

	return inhalt, nil
}

// Dateien nennt die Dateien im Paket, ohne Verzeichnisse.
//
// **Für den Rundlauf**: er prüft jede Datei, nicht die, an die jemand gedacht
// hat. Die Reihenfolge ist die des Archivs beziehungsweise des Verzeichnisses.
//
// Ein Fehler beim Gehen wird **zurückgegeben**, nicht übergangen. Der erste
// Anlauf verschluckte ihn mit der Begründung, eine Teilliste sei brauchbarer
// als nichts — das stimmt für den Aufrufer, aber er soll wissen, dass sie eine
// Teilliste ist.
func (p *Paket) Dateien() ([]string, error) {
	var namen []string

	fehler := fs.WalkDir(p.dateien, ".", func(pfad string, eintrag fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !eintrag.IsDir() {
			namen = append(namen, pfad)
		}

		return nil
	})
	if fehler != nil {
		return namen, fmt.Errorf("die Dateien von %s aufzählen: %w", p.KurzerName, fehler)
	}

	return namen, nil
}

// paketpfad prüft einen Verweis und gibt ihn in der Form, die `fs.FS` erwartet.
//
// `fs.ValidPath` allein genügt nicht als Erklärung, deshalb steht sie hier
// ausgeschrieben: sie weist einen führenden Schrägstrich, `.` und `..` ab —
// also genau die drei Wege, auf denen ein Eintrag aus dem Paket herausführt.
func paketpfad(name string) (string, error) {
	sauber := path.Clean(strings.TrimSpace(name))

	if !fs.ValidPath(sauber) || sauber == "." {
		return "", fmt.Errorf("%w: %q", ErrAusbruch, name)
	}

	return sauber, nil
}

// kopfLesen liest die Hauptdatei und füllt die Angaben des Pakets.
func (p *Paket) kopfLesen() error {
	haupt := p.Art.Hauptdatei()
	if haupt == "" {
		return fmt.Errorf("%q trägt keine bekannte Paketendung — die Art ist unbestimmt", p.Pfad)
	}

	inhalt, err := p.Datei(haupt)
	if err != nil {
		return fmt.Errorf("Hauptdatei %q fehlt: %w", haupt, err)
	}

	baum, fehler := Zerteile(filepath.Join(filepath.Base(p.Pfad), haupt), inhalt)
	if fehler != nil {
		return fehler
	}

	p.KopfRoh = inhalt
	p.Kopf = baum
	p.Spiel = wert(baum.Ersten("game"))
	// **Die Namen auch als Block** (C02, #2227): freier Text, den [Schreibe]
	// als Block setzt, wo er eine Klammer offen lässt. `game:` nicht — in
	// einer `.mse-locale` ist es ein Abschnitt.
	p.KurzerName = WertOderBlock(baum.Ersten("short name"))
	p.VollerName = WertOderBlock(baum.Ersten("full name"))
	p.Version = p.datumOderBefund(baum.Ersten("version"))
	p.einordnungLesen(baum)
	p.abhaengigkeitenLesen(baum)

	return nil
}

// abhaengigkeitenLesen sammelt die `depends on`-Blöcke.
//
// Über `Alle` und nicht über `Ersten`: derselbe Schlüssel wiederholt sich, und
// wer nur den ersten nimmt, verliert den Rest — auffallen würde das erst, wenn
// eine Vorlage ihre zweite Abhängigkeit nicht mehr findet.
func (p *Paket) abhaengigkeitenLesen(baum *Baum) {
	for _, block := range baum.Alle("depends on") {
		abhaengigkeit, fehlstelle, fassungsfehler := abhaengigkeitAus(block)
		if abhaengigkeit.Paket == "" {
			p.melde(block.Stelle, "eine Abhängigkeit ohne `package` — sie lässt sich nicht auflösen")

			continue
		}

		if fassungsfehler != nil {
			p.melde(fehlstelle, fassungsfehler.Error())
		}

		p.Abhaengigkeiten = append(p.Abhaengigkeiten, abhaengigkeit)
	}
}

// abhaengigkeitAus liest eine Abhängigkeit in **beiden** Formen, die MSE2
// kennt: als Block mit `package:` und `version:` und als eine Zeile
// `paket fassung` (`Reader::handle(PackageDependency&)`,
// `src/util/io/package.cpp`, Zeilen 515–528, Fassung f9d9356 vom
// 2021-07-06). Ein Installer schreibt nur die zweite.
//
// **Ohne Paketnamen ist das Ergebnis leer**; eine unlesbare Fassung kommt als
// Fehler neben einer Abhängigkeit, deren Name gilt — mit der Zeile, in der
// sie steht.
func abhaengigkeitAus(block *Knoten) (Abhaengigkeit, Stelle, error) {
	abhaengigkeit := Abhaengigkeit{Stelle: block.Stelle}

	// **Ein Block ist, was einen Eintrag unter sich hat** — eine Leerzeile
	// oder ein Kommentar dahinter hängt der Zerteiler auch an, und sechs der
	// elf Dateien mit der kurzen Form im Pack stehen so.
	if !slices.ContainsFunc(block.Kinder, func(kind *Knoten) bool { return kind.Zeilenart == ArtEintrag }) {
		// **Die Fassung hinter dem ersten Leerzeichen**, mit beliebigem
		// Abstand davor — MSE2 liest sie mit `sscanf`.
		teile := strings.Fields(block.Wert)
		if len(teile) == 0 {
			return Abhaengigkeit{}, block.Stelle, nil
		}

		abhaengigkeit.Paket = teile[0]
		if len(teile) == 1 {
			return abhaengigkeit, block.Stelle, nil
		}

		fassung, err := DatumLesen(teile[1])
		abhaengigkeit.Version = fassung

		return abhaengigkeit, block.Stelle, err
	}

	name := block.Ersten("package")
	if name == nil || name.Wert == "" {
		return Abhaengigkeit{}, block.Stelle, nil
	}

	abhaengigkeit.Paket = name.Wert

	fassung := block.Ersten("version")
	if fassung == nil || fassung.Wert == "" {
		return abhaengigkeit, block.Stelle, nil
	}

	datum, err := DatumLesen(fassung.Wert)
	abhaengigkeit.Version = datum

	return abhaengigkeit, fassung.Stelle, err
}

// datumOderBefund liest ein Datum und macht aus einem Fehler einen Befund.
func (p *Paket) datumOderBefund(knoten *Knoten) Datum {
	if knoten == nil || knoten.Wert == "" {
		return Datum{}
	}

	datum, err := DatumLesen(knoten.Wert)
	if err != nil {
		p.melde(knoten.Stelle, err.Error())

		return Datum{}
	}

	return datum
}

func (p *Paket) melde(stelle Stelle, meldung string) {
	p.Befunde = append(p.Befunde, Befund{Stelle: stelle, Meldung: meldung})
}

func wert(knoten *Knoten) string {
	if knoten == nil {
		return ""
	}

	return knoten.Wert
}
