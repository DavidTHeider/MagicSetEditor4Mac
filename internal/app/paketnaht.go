// **Der Paketkatalog und der Paketbezug für die Flächen ohne Fenster**
// (#1107).
//
// # Warum jede Fläche auffrischt
//
// Der Katalog steht im Bestand, gefüllt wird er von der Platte. Das Fenster
// frischt beim Start auf (`app.go`); eine Kommandozeile hat keinen Start, und
// ohne diesen Zug gäbe sie auf einem frischen Rechner **keinen einzigen
// Stil** zurück — dieselbe Frage, zwei Antworten, und genau das verbietet §16.
//
// **Befunde statt Abbruch** (§8): ein `.mse-style` ohne lesbare Hauptdatei
// liegt in jeder gewachsenen Sammlung. Wer daran abbräche, hätte für **kein**
// Paket einen Katalog.

package app

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// Paketweg liest den Katalog und bezieht Pakete.
type Paketweg struct {
	verwalter     Paketfuehrer
	bezieher      Paketbezieher
	aktualisierer Paketaktualisierer
}

// MitVerwalter setzt die Naht zum Paketkatalog — für Tests.
func (w *Paketweg) MitVerwalter(verwalter Paketfuehrer) *Paketweg {
	w.verwalter = verwalter

	return w
}

// MitBezieher setzt die Naht zum Paketbezug — für Tests.
//
// **Wie [App.MitPaketbezieher]**, und aus demselben Grund: ein Bezug lädt aus
// dem Netz, und ein Test, der das täte, prüfte die Leitung (§3).
func (w *Paketweg) MitBezieher(bezieher Paketbezieher) *Paketweg {
	w.bezieher = bezieher

	return w
}

// PaketnahtUeber baut den Weg zu den Paketen über einer Datenbank.
func PaketnahtUeber(db *sql.DB) (*Paketweg, error) {
	pakete, err := Paketverzeichnis()
	if err != nil {
		return nil, err
	}

	return paketwegMit(db, pakete), nil
}

// paketwegMit baut ihn über einem schon aufgelösten Paketverzeichnis.
func paketwegMit(db *sql.DB, pakete string) *Paketweg {
	wurzeln := paketwurzelnFuer(pakete)

	verwalter := service.NeuerPaketverwalter(storage.NeuerPaketkatalog(db), pakete, wurzeln...).
		MitSetkoepfen(storage.NeuerSetLeser(db))
	bezieher := service.NeuerPaketbezieher(pakete, Kenner())

	return &Paketweg{
		verwalter: verwalter, bezieher: bezieher,
		aktualisierer: service.NeuePaketaktualisierung(verwalter, bezieher, bezieher),
	}
}

// Liste frischt den Katalog auf und gibt ihn.
//
// **Beides in einem Zug**, wie die Paketseite des Fensters: wer ein Paket von
// Hand hineinlegt, sieht es beim nächsten Aufruf — ohne einen zweiten Befehl.
//
// **Mit der Übersicht** — Herkunftsart, lokale Änderung, wer es braucht und
// wie viele Sets es benutzen (DOM-01.AK-01, #1696): dieselbe Auskunft wie im
// Fenster (§16).
func (w *Paketweg) Liste(ctx context.Context) (pakete []Paketzeile, befunde []domain.Befund,
	fehler error,
) {
	befunde, err := w.verwalter.Auffrischen(ctx)
	if err != nil {
		return nil, befunde, err
	}

	liste, lesefehler := w.verwalter.Uebersicht(ctx)
	if lesefehler != nil {
		return nil, befunde, lesefehler
	}

	zeilen := make([]Paketzeile, 0, len(liste))
	for _, zeile := range liste {
		zeilen = append(zeilen, Paketzeile{
			Paket: zeile.Paket, Herkunftsart: string(zeile.Herkunftsart),
			GebrauchtVon: append([]string{}, zeile.GebrauchtVon...), Benutzt: zeile.Benutzt,
		})
	}

	return zeilen, befunde, nil
}

// Paketzeile ist ein Paket samt Übersicht — **ohne `service`-Typen**, damit
// eine Fläche ohne Fenster sie lesen kann, ohne `internal/service` zu kennen
// (§2), wie [Paketuebernahme]. DOM-01.AK-01 (#1696).
type Paketzeile struct {
	Paket        domain.Paket
	Herkunftsart string
	GebrauchtVon []string
	Benutzt      int
}

// Pakete frischt den Katalog auf und gibt die Pakete **ohne Übersicht** — für
// eine Auswahlliste wie die der Vorgaben, die keine Nutzung zählt und keine
// Setköpfe lesen soll.
func (w *Paketweg) Pakete(ctx context.Context) ([]domain.Paket, error) {
	if _, err := w.verwalter.Auffrischen(ctx); err != nil {
		return nil, err
	}

	return w.verwalter.Liste(ctx)
}

// Setbedarf nennt, was einem Set an Paketen fehlt — DOM-01.AK-04 (#1696),
// als DTO, damit eine Fläche ohne Fenster `internal/service` nicht kennt.
func (w *Paketweg) Setbedarf(ctx context.Context, setID int64) ([]AbhaengigkeitDTO, error) {
	befunde, err := w.verwalter.Setbedarf(ctx, setID)
	if err != nil {
		return nil, err
	}

	return alsAbhaengigkeiten(befunde), nil
}

// Plan plant einen Vorgang — DOM-01.AK-02 (#1696), derselbe Weg wie im
// Fenster ([App.PaketPlan]).
func (w *Paketweg) Plan(ctx context.Context, vorgang, art, name, quelle string) (PaketplanDTO, error) {
	return planen(ctx, w.verwalter, vorgang, art, name, quelle)
}

// Ersetzen legt eine neue Fassung ein — DOM-01.AK-02/AK-03.
func (w *Paketweg) Ersetzen(ctx context.Context, quelle string, bestaetigt bool) (domain.Paket, error) {
	return w.verwalter.Ersetzen(ctx, quelle, bestaetigt)
}

// Zurueck holt die vorige Fassung zurück.
func (w *Paketweg) Zurueck(ctx context.Context, art, name string) (domain.Paket, error) {
	return w.verwalter.Zurueck(ctx, art, name)
}

// Auffrischen gleicht den Katalog mit der Platte ab.
//
// **Die Befunde werden zurückgegeben und nicht verschluckt** — ein Paket, das
// nicht gelesen werden konnte, ist der häufigste Grund dafür, dass ein Stil
// fehlt.
func (w *Paketweg) Auffrischen(ctx context.Context) ([]domain.Befund, error) {
	return w.verwalter.Auffrischen(ctx)
}

// Abhaengigkeiten nennt, was ein installiertes Paket vermisst.
func (w *Paketweg) Abhaengigkeiten(
	ctx context.Context,
) ([]service.Abhaengigkeitsbefund, error) {
	return w.verwalter.Abhaengigkeiten(ctx)
}

// Installieren bindet ein Paket ein, das schon auf der Platte liegt (#1291).
//
// **Nicht dasselbe wie [Paketweg.Beziehen]**: jenes lädt aus dem Netz. Wer
// einen Rechner aus einem Skript einrichtet, hat oft beides — ein Archiv, das
// er selbst mitbringt, und eines, das er holt.
func (w *Paketweg) Installieren(ctx context.Context, quelle string) (domain.Paket, error) {
	if strings.TrimSpace(quelle) == "" {
		return domain.Paket{}, fmt.Errorf(
			"%w: ohne Quelle lässt sich kein Paket einbinden", domain.ErrValidation)
	}

	return w.verwalter.Installieren(ctx, quelle)
}

// Anlegen schreibt ein neues Paket aus einem Gerüst (#1350).
//
// **Ohne Auffrischen davor**, anders als [Paketweg.Liste]: das Anlegen endet
// in `Installieren`, und das trägt selbst ein. Ein Abgleich davor läse die
// Platte für nichts.
//
// **Vier Zeichenketten und kein `service.Neuanlage`**, und das ist keine
// Bequemlichkeit: `cmd/mse`, `cmd/mse-mcp` und `cmd/mse-rest` dürfen
// `internal/service` nicht kennen (§2), und der Schichtenwächter hat es
// angestrichen. Wer den Verbund hier durchreichte, zwänge jede Fläche, ihn zu
// bauen — und damit die Schicht zu importieren, die sie nicht kennen soll.
//
// **Der Vollname fällt hier auf den Namen zurück**, an **einer** Stelle für
// alle vier Flächen: sonst täte es jede für sich, und drei Kopien derselben
// Zeile laufen auseinander.
func (w *Paketweg) Anlegen(
	ctx context.Context, art, name, vollname, spiel string,
) (domain.Paket, error) {
	if strings.TrimSpace(vollname) == "" {
		vollname = name
	}

	return w.verwalter.Anlegen(ctx, service.Neuanlage{
		Art: art, Name: name, Vollname: vollname, Spiel: spiel,
	})
}

// Kopieren legt eine eigene Kopie eines Pakets an — FMT-02.AK-04 (#1691).
func (w *Paketweg) Kopieren(ctx context.Context, art, name, neu string) (domain.Paket, error) {
	return w.verwalter.Kopieren(ctx, art, name, neu)
}

// Weitergeben schreibt Pakete als `.mse-installer` — FMT-02.AK-04 (#1691).
func (w *Paketweg) Weitergeben(ctx context.Context, namen []string, ziel string) (string, error) {
	return w.verwalter.Weitergeben(ctx, namen, ziel)
}

// Installerziel sagt, welche Datei [Paketweg.Weitergeben] schreiben wird —
// für MCP und REST, die sie vorher reservieren (0044).
func Installerziel(namen []string, ziel string) (string, error) {
	return service.Installerziel(namen, ziel)
}

// Einlesen legt die Pakete eines `.mse-installer` ins Verzeichnis —
// FMT-02.AK-04 (#1691).
//
// **Das Ergebnis kommt auch neben einem Fehler**: was eingelesen ist, bleibt
// eingelesen (§7, Regel 3), und wer die Kommandozeile fragt, soll es lesen.
func (w *Paketweg) Einlesen(ctx context.Context, quelle string) (*Paketuebernahme, error) {
	roh, err := w.verwalter.Einlesen(ctx, quelle)
	if roh == nil {
		return nil, err
	}

	return paketuebernahmeAus(roh), err
}

// Entfernen nimmt ein Paket aus dem Paketverzeichnis (#1291).
//
// **Nur das Bezugsverzeichnis.** Die Saat liegt im Bündel; ein Entfernen dort
// löschte einen Teil der Anwendung (#58) — der Verwalter weist es ab.
//
// **`bestaetigterPfad` ist der Pfad, den der Rufende gesehen hat** — leer
// heißt: nicht bestätigt. Der Vergleich steht im Dienst, hinter derselben
// Lesung, die auch löscht (#1344).
func (w *Paketweg) Entfernen(ctx context.Context, art, name, bestaetigterPfad string) error {
	return w.verwalter.Entfernen(ctx, art, name, bestaetigterPfad)
}

// Paketuebernahme ist der Ausgang einer Übernahme.
//
// **Ohne `service`-Typen**, damit eine Fläche ohne Fenster sie lesen kann,
// ohne `internal/service` zu kennen (§2) — wie [Bildschreibung].
type Paketuebernahme struct {
	Uebernommen []domain.Paket
	Uebergangen []string

	// Befunde ist, was dabei aufgefallen ist — als Sätze.
	Befunde []string
}

// Uebernehmen holt alle Vorlagenpakete aus einem Verzeichnis (#1291).
//
// **Übergangen und nicht überschrieben**: wer eine Vorlage bearbeitet hat,
// verlöre sie sonst, und nichts sagte es ihm.
func (w *Paketweg) Uebernehmen(ctx context.Context, ort string) (*Paketuebernahme, error) {
	if strings.TrimSpace(ort) == "" {
		return nil, fmt.Errorf(
			"%w: ohne Verzeichnis lässt sich nichts übernehmen", domain.ErrValidation)
	}

	roh, err := w.verwalter.Uebernehmen(ctx, ort)
	if err != nil {
		return nil, err
	}

	uebernahme := paketuebernahmeAus(roh)

	// **Nach der Übernahme wird aufgefrischt** — wie nach einem Bezug: ein
	// übernommenes Paket, das im Katalog fehlt, ist für jede Stilwahl
	// unsichtbar.
	if _, auffrischfehler := w.verwalter.Auffrischen(ctx); auffrischfehler != nil {
		return uebernahme, auffrischfehler
	}

	return uebernahme, nil
}

// Beziehen holt ein Vorlagenpaket und trägt es in den Katalog ein.
//
// **Der Stand ist Pflicht und kein Zweig** (§11): ein Zweig zeigt morgen auf
// etwas anderes, und derselbe Aufruf ergäbe ein anderes Paket.
func (w *Paketweg) Beziehen(
	ctx context.Context, adresse, stand, unterverzeichnis, name string,
) (*service.Bezugsergebnis, error) {
	ergebnis, err := w.bezieher.Beziehen(ctx, adresse, stand, unterverzeichnis, name)
	if err != nil {
		return nil, err
	}

	// **Nach dem Bezug wird aufgefrischt und vermerkt** — wie im Fenster
	// ([App.PaketBeziehen]). Ein bezogenes Paket, das im Katalog fehlt, ist
	// für jede Stilwahl unsichtbar. **Bis zum 2026-10-10 frischte dieser Weg
	// nur auf**: Herkunft und Prüfsumme kamen nicht in den Katalog, und der
	// Baum war für jede Frage nach seinem Stand unsichtbar (G04, #2251;
	// gefunden vom Flächenlauf über die Bezugsstände).
	if _, vermerkfehler := w.verwalter.Vermerken(ctx, service.Vermerk{
		Ort: ergebnis.Ort, Herkunft: herkunftstext(adresse, ergebnis.Stand), Pruefsumme: ergebnis.Pruefsumme,
	}); vermerkfehler != nil {
		return ergebnis, vermerkfehler
	}

	return ergebnis, nil
}

// paketuebernahmeAus macht aus dem Ergebnis des Dienstes eines ohne
// `service`-Typen — **keine Liste ist `nil`** (§3).
func paketuebernahmeAus(roh *service.Uebernahme) *Paketuebernahme {
	uebernahme := &Paketuebernahme{
		Uebernommen: roh.Uebernommen,
		Uebergangen: roh.Uebergangen,
		Befunde:     alsMeldungen(roh.Befunde),
	}

	if uebernahme.Uebergangen == nil {
		uebernahme.Uebergangen = []string{}
	}

	return uebernahme
}
