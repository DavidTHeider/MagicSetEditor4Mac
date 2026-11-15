// **Eine Exportvorlage ausführen — ohne Fenster** —
// #1291.
//
// # Der Befund
//
// `docs/FAEHIGKEITEN.md` führte die Zeile so:
//
//	| Über eine Exportvorlage ausgeben | `Exportvorlagen`, `VorlageAusfuehren`, `ExportoptionSetzen` | — | — | — | ✋ |
//
// Die erhobene Hand meint einen Speicherdialog, und der steckt allein in
// `Bindungen.VorlageAusfuehren`. `App.VorlageAusfuehren` nimmt einen Pfad
// entgegen und sagt es selbst: *„Der Pfad kommt vom Aufrufer, nicht aus einem
// Dialog."*
//
// **Und es ist die Fähigkeit, für die MSE2 seine Vorlagen mitbringt**: das
// `M15-Magic-Pack` führt zwölf `.mse-export-template` — Cockatrice, MWS,
// LackeyCCG, Planesculptors. Wer ein Set aus einem Skript heraus in eines
// dieser Formate bringen will, hatte bis dahin keinen Weg.

package app

import (
	"context"
	"database/sql"
	"errors"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// Vorlagenweg führt die Exportvorlagen eines Sets aus.
type Vorlagenweg struct{ anwendung *App }

// Exportvorlage ist eine Vorlage, wie eine Fläche ohne Fenster sie sieht.
//
// **Ohne `msepaket`-Typen**, damit sie gelesen werden kann, ohne das
// Paketformat zu kennen (§2).
type Exportvorlage struct {
	// Paket ist der Name, mit dem die Vorlage aufgerufen wird.
	Paket string

	Name string

	// Dateityp ist MSEs `file type:` — er sagt, welche Endung das Ergebnis
	// tragen soll (`*.xml|*.xml|*.*|*.*`).
	Dateityp string

	// Optionen sind die Felder, die die Vorlage einstellbar macht.
	Optionen []Vorlagenoption
}

// Vorlagenoption ist ein einstellbares Feld einer Exportvorlage.
type Vorlagenoption struct {
	Name string
	Art  string

	// Wert ist der behaltene oder der voreingestellte.
	Wert string
}

// Vorlagenlauf ist der Ausgang einer Ausfuhr.
type Vorlagenlauf struct {
	Pfad string

	Befunde []string
}

// VorlagenwegUeber baut den Weg, auf dem eine Exportvorlage ohne Fenster läuft.
//
// **Nicht `Vorlagennaht`**, obwohl der Name naheläge: die gibt es schon, und
// sie beantwortet eine andere Frage — wie eine Karte mit einer
// **ungespeicherten** Stilvorlage aussieht (#1164). Zwei Dinge mit einem Namen
// wären eine Verwechslung, die der Übersetzer nur beim ersten Mal bemerkt.
//
// **Dieselben Teile wie [Fremdbildnaht]**: eine kopflose Anwendung, in die
// genau die Nähte eingesetzt werden, die der Vorlagenweg braucht. Ein zweiter
// Weg durch den Vorlagendienst wäre eine zweite Antwort auf dieselbe Frage —
// und die Ausfuhr hat davon genug (§2).
func VorlagenwegUeber(db *sql.DB, urheber domain.Urheber) (*Vorlagenweg, error) {
	schreiber, schreibfehler := Schreibnaht(db, urheber)

	werk, werkfehler := Zeichennaht(db)

	speicherer, speicherfehler := Setschreibnaht(db)

	// **Und die Stilwahl** — ohne sie ist `set.stylesheet` an Kommandozeile,
	// REST und MCP `nil`, während das Fenster ihn kennt (#2039), und keine
	// Karte rechnet im Stil ihrer Vorlage
	// (#2182).
	stile, stilfehler := stilwaehlerUeber(db)

	if fehler := errors.Join(schreibfehler, werkfehler, speicherfehler, stilfehler); fehler != nil {
		return nil, fehler
	}

	leser := storage.NeuerSetLeser(db)

	anwendung := NeueApp("")
	anwendung.MitEditor(schreiber, Kartenlesernaht(db), werk.Planbauer, werk.Rasterer)
	anwendung.MitBestand(leser)
	anwendung.MitSetwerten(setwertnaht{
		leser:     leser,
		schreiber: storage.NeuerKartenschreiber(db).Als(urheber),
	})
	anwendung.MitVorlagen(werk.Vorlagen)
	anwendung.MitBildausfuhr(werk.Bildausfuhr)
	anwendung.MitSpeicherer(speicherer)
	anwendung.MitSymbolzeichnungen(werk.Bildquelle)
	anwendung.MitExportoptionen(storage.NeueExportoptionen(db))
	anwendung.MitStilwahl(stile)

	return &Vorlagenweg{anwendung: anwendung}, nil
}

// Vorlagen nennt die Exportvorlagen, die zu einem Set passen.
//
// **Mit den behaltenen Optionswerten**, wie im Fenster: wer eine Option
// gesetzt hat, sieht sie beim nächsten Mal wieder.
func (v *Vorlagenweg) Vorlagen(ctx context.Context, setID int64) (
	[]Exportvorlage, []string, error,
) {
	antwort := v.anwendung.Exportvorlagen(ctx, kennungText(setID))
	if antwort.Fehler != nil {
		return nil, antwort.Befunde, AlsFehler(antwort.Fehler)
	}

	vorlagen := make([]Exportvorlage, 0, len(antwort.Vorlagen))

	for _, gelesen := range antwort.Vorlagen {
		optionen := make([]Vorlagenoption, 0, len(gelesen.Optionen))
		for _, option := range gelesen.Optionen {
			optionen = append(optionen, Vorlagenoption{
				Name: option.Name, Art: option.Art, Wert: option.Wert,
			})
		}

		vorlagen = append(vorlagen, Exportvorlage{
			Paket: gelesen.Paket, Name: gelesen.Name,
			Dateityp: gelesen.Dateityp, Optionen: optionen,
		})
	}

	return vorlagen, antwort.Befunde, nil
}

// OptionSetzen behält den Wert einer Vorlagenoption.
func (v *Vorlagenweg) OptionSetzen(ctx context.Context, vorlage, feld, wert string) error {
	return AlsFehler(v.anwendung.ExportoptionSetzen(ctx, vorlage, feld, wert).Fehler)
}

// Ausfuehren führt eine Vorlage aus und schreibt das Ergebnis an einen Pfad.
//
// **Über das ganze Set**: eine Teilmenge stellt das Fenster ein, und ohne
// Fenster wird nichts angehakt (`domain.Kartenauswahl` leer heißt: alle).
func (v *Vorlagenweg) Ausfuehren(
	ctx context.Context, setID int64, paket, ziel string, wahl map[string]string,
) (*Vorlagenlauf, error) {
	antwort := v.anwendung.VorlageAusfuehren(
		ctx, kennungText(setID), paket, ziel, wahl, nil)
	if antwort.Fehler != nil {
		// **Die Befunde reisen mit dem Fehler** (#209): „der Name
		// `write_image_file` ist hier nicht bekannt" ist eine Aussage über
		// **ihre** Vorlage, und ohne sie stünde nur „Die Eingabe ist nicht
		// gültig".
		return &Vorlagenlauf{Befunde: antwort.Befunde}, AlsFehler(antwort.Fehler)
	}

	return &Vorlagenlauf{Pfad: antwort.Pfad, Befunde: antwort.Befunde}, nil
}

// Symbolquelle nennt den Typ, den [App.MitSymbolzeichnungen] erwartet.
//
// **Ein Alias und kein zweiter Typ**: er steht hier, damit ein Aufrufer die
// Naht setzen kann, ohne `internal/service` zu nennen (§2).
type Symbolquelle = service.Symbolquelle

// Vorlagenprobe sagt, was eine Ausfuhr ergäbe — vor dem Schreiben
// (#2015).
type Vorlagenprobe struct {
	// Karten sind die Karten, über die die Vorlage lief.
	Karten int

	// Dateiname ist der der Hauptdatei, Bytes ihre Größe.
	Dateiname string
	Bytes     int

	// Beigaben sind die Dateien daneben, BeigabenBytes die bekannte Größe,
	// Verzeichnis der Ordner dafür — leer, wenn die Vorlage keinen anlegt.
	Beigaben      int
	BeigabenBytes int
	Verzeichnis   string

	// Geschaetzt sagt, dass BeigabenBytes eine Untergrenze ist.
	Geschaetzt bool

	Befunde []string
}

// Probelauf wertet eine Vorlage aus und sagt, was sie ergäbe — und schreibt
// nichts.
//
// **Derselbe Weg wie [Vorlagenweg.Ausfuehren]** (`rechneVorlage`): dieselben
// Fehler, und die Befunde reisen auch hier mit dem Fehler (#209). Der
// Schreibriegel wird nicht gefragt, und ein Verzeichnis für die Beigaben wird
// nicht angelegt.
func (v *Vorlagenweg) Probelauf(
	ctx context.Context, setID int64, paket, ziel string, wahl map[string]string,
) (*Vorlagenprobe, error) {
	antwort := v.anwendung.VorlageProbelauf(ctx, kennungText(setID), paket, ziel, wahl, nil)
	if antwort.Fehler != nil {
		return &Vorlagenprobe{Befunde: antwort.Befunde}, AlsFehler(antwort.Fehler)
	}

	return &Vorlagenprobe{
		Karten: antwort.Karten, Dateiname: antwort.Dateiname, Bytes: antwort.Bytes,
		Beigaben: antwort.Beigaben, BeigabenBytes: antwort.BeigabenBytes,
		Verzeichnis: antwort.Verzeichnis, Geschaetzt: antwort.Geschaetzt,
		Befunde: antwort.Befunde,
	}, nil
}
