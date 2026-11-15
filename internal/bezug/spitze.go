package bezug

// **Wie weit eine Quelle ist** — G04
// (#2251).
//
// Bezogen wird an einem **Commit** (§11), und so ein Stand veraltet nicht von
// selbst. Was fehlte, ist die Auskunft, dass es einen neueren gibt — MSE2
// gibt sie über eine Fassungsliste (`gui/update_checker.cpp`, Zeilen 27–80,
// Fassung f9d9356 vom 2021-07-06), dieses Programm über die Quelle selbst.
//
// # Was hier nicht geschieht
//
// **Kein Zweig wird nachgeführt.** [Spitze] löst den Hauptzweig zu **einem
// Commit** auf, und bezogen wird danach dieser Commit — wer ihn nicht
// bestätigt, bekommt nichts. **Und nichts geschieht von selbst**: gefragt
// wird, wenn jemand fragt (§16, Regel 5 — ohne Netz arbeitet die Anwendung
// weiter).
//
// # Die Gestalt der Antworten, nachgelesen
//
// `GET /repos/{owner}/{repo}/commits/{ref}` gibt `sha` und
// `commit.committer.date`; `GET /repos/{owner}/{repo}/compare/{basis}...{kopf}`
// gibt `status` (`diverged`, `ahead`, `behind`, `identical`), `ahead_by`,
// `behind_by` und `files` mit `filename` und `status` (`added`, `removed`,
// `modified`, `renamed`, `copied`, `changed`, `unchanged`). **Mehr als 300
// Dateien nennt er nicht** — für die ganze Spanne; weitere Seiten blättern
// die Commits, und die Dateien stehen nur auf der ersten. Nachgelesen am
// 2026-10-11 in GitHubs REST-Dokumentation („Commits", API-Fassung
// 2022-11-28) und gemessen am 2026-10-10: `M15-Magic-Pack` steht vom
// bezogenen Commit `3f7f8ce` aus 63 Commits hinter seiner Spitze, und der
// Vergleich nennt genau 300 Dateien. **Hier stand bis zum 2026-10-11**, dass
// weitere Seiten Dateien bis 3000 nennen — erinnert statt gelesen; was es
// darüber hinaus braucht, steht in `baum.go`.

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// Spitzenstand ist der Commit, auf den der Hauptzweig gerade zeigt.
type Spitzenstand struct {
	Commit string
	Zeit   time.Time
}

// Aenderung ist eine Datei, die sich zwischen zwei Commits ändert.
type Aenderung struct {
	Datei string

	// Art ist GitHubs Wort dafür: `added`, `removed`, `modified`, `renamed`,
	// `copied`, `changed`, `unchanged`.
	Art string

	// Vorher ist der vorige Name einer umbenannten Datei (`previous_filename`)
	// — leer bei jeder anderen Art. Ohne ihn verschwand ein Paket, aus dem
	// eine Datei in ein anderes wanderte, aus der Vorschau (Befund 3 des
	// Zweitblicks über Phase G, zweiter Teil).
	Vorher string
}

// Vergleich sagt, was zwischen zwei Commits liegt.
type Vergleich struct {
	// Lage ist `ahead`, `behind`, `identical` oder `diverged` — vom ersten
	// Commit aus gesehen.
	Lage       string
	Vorsprung  int
	Rueckstand int
	Dateien    []Aenderung

	// Gekuerzt sagt, dass es mehr Dateien gibt, als hier stehen — die
	// Gegenstelle nennt höchstens 300 ([Baum] nennt alle).
	Gekuerzt bool
}

// vergleichsgrenze ist die Zahl der Dateien, die ein Vergleich höchstens
// nennt.
const vergleichsgrenze = 300

// auskunftsgrenze begrenzt eine Antwort. **32 MB**: ein Vergleich trägt die
// Unterschiede jeder Datei mit, und bei 300 Dateien eines Vorlagenpacks ist
// das mehr als eine Seite Text — aber kein Gigabyte.
const auskunftsgrenze = 32 << 20

// Spitze fragt, auf welchen Commit der Hauptzweig gerade zeigt.
//
// `adresse` ist die volle Adresse der Auskunft; welche das ist, weiß der
// Aufrufer (`service`), denn er kennt die Gestalt der Quelle.
func Spitze(ctx context.Context, adresse, kenner string) (Spitzenstand, error) {
	var antwort struct {
		SHA    string `json:"sha"`
		Commit struct {
			Committer struct {
				Date time.Time `json:"date"`
			} `json:"committer"`
		} `json:"commit"`
	}

	if fehler := auskunft(ctx, adresse, kenner, &antwort); fehler != nil {
		return Spitzenstand{}, fehler
	}

	// **Pflichtfelder streng** (§16): ohne Commit oder Datum ist die Antwort
	// keine Auskunft über einen Stand — und ein Zweigname an dieser Stelle
	// wäre genau die Nachführung, die hier nicht geschehen soll.
	if !commitform.MatchString(antwort.SHA) || antwort.Commit.Committer.Date.IsZero() {
		return Spitzenstand{}, domain.NewFieldError("adresse", domain.SchluesselValidierung,
			"%s nennt keinen Stand: ein Commit und sein Datum fehlen.", adresse)
	}

	return Spitzenstand{Commit: antwort.SHA, Zeit: antwort.Commit.Committer.Date.UTC()}, nil
}

// Vergleiche fragt, was zwischen zwei Commits liegt.
func Vergleiche(ctx context.Context, adresse, kenner string) (Vergleich, error) {
	var antwort struct {
		Status   string `json:"status"`
		AheadBy  int    `json:"ahead_by"`
		BehindBy int    `json:"behind_by"`
		Files    []struct {
			Filename         string `json:"filename"`
			Status           string `json:"status"`
			PreviousFilename string `json:"previous_filename"`
		} `json:"files"`
	}

	if fehler := auskunft(ctx, adresse, kenner, &antwort); fehler != nil {
		return Vergleich{}, fehler
	}

	if antwort.Status == "" {
		return Vergleich{}, domain.NewFieldError("adresse", domain.SchluesselValidierung,
			"%s nennt keine Lage der beiden Stände.", adresse)
	}

	vergleich := Vergleich{
		Lage: antwort.Status, Vorsprung: antwort.AheadBy, Rueckstand: antwort.BehindBy,
		Dateien:  make([]Aenderung, 0, len(antwort.Files)),
		Gekuerzt: len(antwort.Files) >= vergleichsgrenze,
	}

	for _, datei := range antwort.Files {
		vergleich.Dateien = append(vergleich.Dateien,
			Aenderung{Datei: datei.Filename, Art: datei.Status, Vorher: datei.PreviousFilename})
	}

	return vergleich, nil
}

// auskunft holt eine JSON-Antwort und übersetzt, was scheitern kann, an das
// Feld, das die Quelle genannt hat.
func auskunft(ctx context.Context, adresse, kenner string, ziel any) error {
	antwort, fehler := netz.Hole(ctx, klient(), nil,
		netz.Kenner{Name: kenner, Annahme: "application/vnd.github+json"}, adresse, auskunftsgrenze)
	if fehler != nil {
		if befund := alsZielbefund(fehler, "adresse", adresse); befund != nil {
			return befund
		}

		if befund := alsDrosselbefund(fehler, "adresse", adresse); befund != nil {
			return befund
		}

		return fmt.Errorf("die Quelle fragen: %w", fehler)
	}

	if antwort.Status != netz.StatusGut {
		return domain.NewFieldError("adresse", domain.SchluesselValidierung,
			"%s gibt keine Auskunft (Status %d).", adresse, antwort.Status)
	}

	if fehler := json.Unmarshal(antwort.Rumpf, ziel); fehler != nil {
		return domain.NewFieldError("adresse", domain.SchluesselValidierung,
			"%s antwortet nicht in der erwarteten Gestalt.", adresse).MitGrund(fehler)
	}

	return nil
}
