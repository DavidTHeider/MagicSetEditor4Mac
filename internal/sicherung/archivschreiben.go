package sicherung

// **Die Schreibseite eines Bündels** — aus `archiv.go` herausgelöst, als die
// Datei mit [PackenAls] (Werkstück O10, #1434) über die Zeilengrenze aus §2
// wuchs.
//
// **Der Schnitt folgt der Frage**: dort steht, *was* ein Bündel ist und wie
// es entsteht und veröffentlicht wird; hier, *wie* seine Teile und sein
// Manifest hineinkommen.

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// hineinschreiben legt die Teile und danach das Manifest ins Archiv.
//
// **Das Manifest zuletzt**, weil es die Summen der Teile trägt: es entsteht
// beim Schreiben, nicht davor. Ein zweiter Lauf über die Quellen wäre die
// Alternative — und läse jede Datei zweimal.
func hineinschreiben(nach io.Writer, manifestname string, teile []Teil, angaben Angaben) error {
	archiv := zip.NewWriter(nach)

	beschreibung := Beschreibung{Angaben: angaben, Teile: make([]Teilangabe, 0, len(teile))}

	for _, teil := range teile {
		angabe, fehler := teilSchreiben(archiv, teil)
		if fehler != nil {
			return fehler
		}

		beschreibung.Teile = append(beschreibung.Teile, angabe)
	}

	if fehler := manifestSchreiben(archiv, manifestname, beschreibung); fehler != nil {
		return fehler
	}

	if fehler := archiv.Close(); fehler != nil {
		return fmt.Errorf("das Archiv schließen: %w", fehler)
	}

	return nil
}

// teilSchreiben legt eine Datei ins Archiv und gibt ihre Angabe.
func teilSchreiben(archiv *zip.Writer, teil Teil) (Teilangabe, error) {
	quelle, err := quelleVon(teil)
	if err != nil {
		return Teilangabe{}, fmt.Errorf("%s lesen: %w", teil.Name, err)
	}

	defer func() { _ = quelle.Close() }() //nolint:errcheck // nur gelesen

	// **Anlegen und Schreiben auf einem Weg.** `zip.Writer.Create` gibt
	// zurück, was beim Abschluss des **vorigen** Eintrags schiefging — beim
	// ersten Teil also nie. Zwei Zweige daraus zu machen hieße, einen zu
	// schreiben, den kein Test betritt (§3).
	rechner := sha256.New()

	var geschrieben int64

	ziel, err := archiv.Create(teil.Name)
	if err == nil {
		geschrieben, err = io.Copy(io.MultiWriter(ziel, rechner), quelle)
	}

	if err != nil {
		return Teilangabe{}, fmt.Errorf("%s ins Archiv legen: %w", teil.Name, err)
	}

	return Teilangabe{
		Name:    teil.Name,
		Groesse: geschrieben,
		Summe:   hex.EncodeToString(rechner.Sum(nil)),
	}, nil
}

// quelleVon gibt den Leser eines Teils — die Datei oder den Inhalt.
//
// **Der Inhalt schließt sich mit**, damit der Aufrufer beide Fälle gleich
// behandelt: `io.NopCloser` ist genau dafür da.
func quelleVon(teil Teil) (io.ReadCloser, error) {
	if teil.Quelle == "" {
		return io.NopCloser(bytes.NewReader(teil.Inhalt)), nil
	}

	return os.Open(teil.Quelle) // #nosec G304 -- der Aufrufer nennt die Teile
}

// manifestSchreiben legt das Manifest ins Archiv.
func manifestSchreiben(archiv *zip.Writer, manifestname string, beschreibung Beschreibung) error {
	ziel, err := archiv.Create(manifestname)
	if err == nil {
		entwerfer := json.NewEncoder(ziel)
		entwerfer.SetIndent("", "  ")

		err = entwerfer.Encode(beschreibung)
	}

	if err != nil {
		return fmt.Errorf("das Manifest schreiben: %w", err)
	}

	return nil
}
