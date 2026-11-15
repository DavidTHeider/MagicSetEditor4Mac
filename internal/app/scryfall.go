package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Scryfallverzeichnis ist der Ort des Zwischenspeichers.
//
// **Neben dem Bestand und den Paketen**, damit alles, was diese Anwendung
// anlegt, an einer Stelle liegt — und wer aufräumen will, ein Verzeichnis
// löscht statt drei.
//
// **Am Vorgabeort, auch bei umgelenktem Bestand** — QA-10.NAT-01 (#1741). Der
// Zwischenspeicher gehört dem Rechner wie die Pakete: er hält öffentliche
// Daten, und ein Wegwerfbestand, der ihn neu lüde, holte 25 MB je Lauf. Wer
// ihn trennen will, trennt das Zuhause (#888 zeigt, was geschieht, wenn nicht).
//
// **Und unter `Caches`, nicht neben dem Bestand** — OPS-08.AK-03 (#1725).
// Hier stand `…/Application Support/MagicSetEditor4Mac/scryfall`, neben
// `bestand.db` und `pakete/`: wer den Ordner ansah, konnte löschbare Antworten
// nicht von Daten unterscheiden. `Library/Caches` ist der Ort, den macOS für
// Wiederholbares vorsieht; eine Sicherung nimmt ihn nicht mit, und der
// Speicher selbst sagt nun, wie groß er ist, und lässt sich leeren.
//
// **Ein alter Speicher zieht einmal um**, damit Antworten und Bilder, die schon
// da sind, nicht verloren gehen — ohne Netz wären sie sonst fort. Gelingt der
// Umzug nicht (ein anderes Laufwerk), beginnt der neue leer.
//
// **Aber nicht aus einem umgelenkten Bestand heraus** (`MSE_BESTAND`): Läufe
// und Proben arbeiten so, und sie sollen den Speicher einer installierten
// Anwendung nicht wegtragen, die ihn noch am alten Ort sucht.
func Scryfallverzeichnis() (string, error) {
	return scryfallverzeichnis(umzugErlaubt)
}

// umzugErlaubt sagt, ob ein alter Speicher umziehen darf — **nie aus einem
// Testlauf**, erkannt wie beim Anmeldebund (`internal/schluesselbund`) an
// `os.Args[0]`. Gemessen am 2026-09-26: ein Lauf ohne umgelenktes Zuhause
// trug den echten Speicher dieses Rechners nach `Caches`, und die installierte
// Anwendung hätte ihn nicht mehr gefunden. Zurückgestellt wurde von Hand.
var umzugErlaubt = !strings.HasSuffix(os.Args[0], ".test")

// scryfallverzeichnis ist [Scryfallverzeichnis] mit ausdrücklicher Erlaubnis
// zum Umzug — die Läufe des Umzugs setzen sie selbst.
func scryfallverzeichnis(umziehen bool) (string, error) {
	// **Beide auf einmal, ein Zweig** — sie scheitern an derselben Ursache,
	// dem fehlenden Benutzerverzeichnis (dieselbe Überlegung wie in
	// [ScryfallBildnaht]).
	pfad, bestandsfehler := Datenbankpfad()
	zwischen, zwischenfehler := zwischenspeicherort()

	if err := errors.Join(bestandsfehler, zwischenfehler); err != nil {
		return "", err
	}

	// **Zusammengeführt und nicht nur umbenannt** — Zweitblick über Phase
	// 6.4, Befund 13. Hier zog der alte Speicher nur in einen **fehlenden**
	// Ort um; ein einziger Lauf mit eigenem Bestand legte ihn an, und danach
	// fand der Umzug nie mehr statt.
	ort := filepath.Join(zwischen, "MagicSetEditor4Mac", "scryfall")
	if umziehen && os.Getenv(Bestandsumgebung) == "" {
		zusammenfuehren(filepath.Join(filepath.Dir(pfad), "scryfall"), ort)
	}

	if err := os.MkdirAll(ort, 0o750); err != nil {
		return "", fmt.Errorf("das Scryfall-Verzeichnis anlegen: %w", err)
	}

	return ort, nil
}

// zusammenfuehren zieht einen alten Speicher in den neuen — Eintrag für
// Eintrag.
//
// **Was an beiden Orten liegt, hat im neuen den jüngeren Stand**, und die
// alte Fassung fällt: es ist ein Zwischenspeicher, und zwei Fassungen
// derselben Antwort wären eine zu viel. Was sich nicht bewegen lässt, bleibt
// stehen und wird beim nächsten Start noch einmal versucht.
func zusammenfuehren(alt, neu string) {
	eintraege, fehler := os.ReadDir(alt)
	if fehler != nil {
		return
	}

	_ = os.MkdirAll(neu, 0o750) //nolint:errcheck // scheitert es, scheitert unten jedes Umbenennen, und alles bleibt

	for _, eintrag := range eintraege {
		von, nach := filepath.Join(alt, eintrag.Name()), filepath.Join(neu, eintrag.Name())

		switch _, dort := os.Lstat(nach); {
		case eintrag.IsDir():
			zusammenfuehren(von, nach)
		case os.IsNotExist(dort):
			_ = os.Rename(von, nach) //nolint:errcheck // siehe oben
		default:
			_ = os.Remove(von) //nolint:errcheck // siehe oben
		}
	}

	_ = os.Remove(alt) //nolint:errcheck // nur, wenn er leer ist
}

// Scryfallquelle ist, was die Fassade vom Abgleich braucht.
//
// # Warum eine Schnittstelle und kein fertiger Abgleich
//
// Damit dieser Weg **ohne Netz** prüfbar ist. Der erste Anlauf baute den
// Abgleich im Verfahren — und damit brauchte jeder Test ein Zuhause, ein
// Verzeichnis und eine Gegenstelle. Was er prüfen sollte, war aber die
// Fassade: Kennungen, Fehlerübersetzung, DTO.
type Scryfallquelle interface {
	// VorschlagFuerKarte folgt dem Verweis der Karte, wenn sie einen hat
	// (#1167).
	//
	// **Und nicht `Vorschlag`**: welcher Druck gemeint ist, ist eine
	// fachliche Frage, und §2 verbietet sie hier. Die Fassade reicht die
	// Kartenkennung durch und entscheidet nichts.
	VorschlagFuerKarte(ctx context.Context, spiel string, jetzige map[string]string,
		karteID int64, name string) (*service.Vorschlag, error)
	Uebernehmen(ctx context.Context, setID, karteID int64,
		vorschlag *service.Vorschlag, felder []string) ([]service.Feldvorschlag, error)
}

// Scryfallumgebung ist die Umgebungsvariable, die die Grundadresse umlenkt.
//
// # Wofür sie da ist
//
// **Für die Flächenläufe** (`internal/flaechen`,
// #1028). Sie
// starten `mse`, `mse-mcp` und `mse-rest` als **eigene Prozesse**; eine
// Veränderliche im selben Programm erreichen sie nicht. Ohne diese Naht ließe
// sich „dieselbe Fähigkeit, überall dieselbe Antwort" (§16) nur mit echtem
// Netz prüfen — und ein Lauf, der wirklich 25 MB lädt, ist in dieser Phase
// zweimal vorgekommen (§3).
//
// **Im Betrieb bleibt sie ungesetzt**, und dann steht hier die Adresse des
// Dienstes. Wer sie setzt, schickt die Anwendung woandershin; das ist dieselbe
// Art von Eingriff wie `MSE_BESTAND` auf eine andere Datenbank.
const Scryfallumgebung = "MSE_SCRYFALL_ADRESSE"

// scryfalladresse ist die Grundadresse des Dienstes.
//
// **Eine Veränderliche, damit ein Test sie umlenken kann** — dieselbe Naht wie
// die Fristen in `internal/bezug`. Im Betrieb ändert sie sich nie; ohne sie
// ließe sich der Weg vom Fenster bis zur Gegenstelle nur mit echtem Netz
// prüfen, und was das wert ist, sagt §3.
var scryfalladresse = umgelenkteAdresse()

// umgelenkteAdresse liest [Scryfallumgebung] — oder gibt die echte Adresse.
func umgelenkteAdresse() string {
	if ort := os.Getenv(Scryfallumgebung); ort != "" {
		return ort
	}

	return scryfall.Adresse
}

// MitScryfalladresse lenkt die Grundadresse um — **nur für Tests**.
//
// Sie gibt zurück, was vorher dastand, damit ein Aufrufer sie zurückstellen
// kann.
func MitScryfalladresse(neue string) (vorher string) {
	vorher, scryfalladresse = scryfalladresse, neue

	return vorher
}

// ScryfallvorschlagAntwort trägt einen Vorschlag oder einen Fehler.
//
// **Ein Umschlag und kein zweiter Rückgabewert**: Wails nimmt bei zwei
// Rückgabewerten den zweiten nur dann als Fehler, wenn er `error` erfüllt —
// ein `*Fehler` täte es nicht und wäre stillschweigend verworfen.
type ScryfallvorschlagAntwort struct {
	Vorschlag *ScryfallvorschlagDTO `json:"vorschlag"`
	Fehler    *Fehler               `json:"fehler"`
}

// ScryfallVorschlag holt einen Vorschlag für eine Karte — und schreibt nichts.
//
// **§16, Regel 6**: „Nichts wird still übernommen. Was von außen in ein
// Kartenfeld will, zeigt vorher, was sich ändert, und trägt danach seine
// Herkunft." Was sich ändern würde, steht in der Antwort; geschrieben wird
// erst mit [App.ScryfallUebernehmen].
func (a *App) ScryfallVorschlag(
	ctx context.Context, setID, karteID, name string,
) ScryfallvorschlagAntwort {
	ctx = a.imProfil(ctx)

	setkennung, befund := benannteKennungAus("setID", setID)
	if befund != nil {
		return ScryfallvorschlagAntwort{Fehler: befund}
	}

	kennung, befund := kennungAus(karteID)
	if befund != nil {
		return ScryfallvorschlagAntwort{Fehler: befund}
	}

	if a.scryfall == nil {
		return ScryfallvorschlagAntwort{Fehler: Uebersetze(ctx, fmt.Errorf(
			"%w: der Zugang zu Scryfall ist nicht eingerichtet", domain.ErrValidation))}
	}

	spiel, jetzige, fehler := a.kartenstand(ctx, setkennung, kennung)
	if fehler != nil {
		return ScryfallvorschlagAntwort{Fehler: fehler}
	}

	vorschlag, holfehler := a.scryfall.VorschlagFuerKarte(ctx, spiel, jetzige, kennung, name)
	if holfehler != nil {
		return ScryfallvorschlagAntwort{Fehler: Uebersetze(ctx, holfehler)}
	}

	als := alsVorschlagDTO(vorschlag)

	return ScryfallvorschlagAntwort{Vorschlag: &als}
}

// ScryfallUebernehmen schreibt die genannten Felder eines Vorschlags.
//
// # Warum der Vorschlag neu geholt wird
//
// Weil zwischen dem Anzeigen und dem Zustimmen Zeit vergeht — und weil ein
// Vorschlag, den das Fenster zurückschickt, ein Wert wäre, den der Aufrufer
// bestimmt. Dann könnte jemand über die Fläche **jeden** Text in ein Feld
// schreiben und die Herkunft „von Scryfall" daranhängen.
//
// **Was das kostet**: der Zwischenspeicher liefert dieselbe Antwort ohne
// Netzgang, solange sie frisch ist — also nichts, außer in dem Fall, in dem
// sich die Antwort zwischendurch geändert hat. Dann wird das Neue übernommen,
// und das ist die richtige Seite: übernommen wird, was **jetzt** dasteht.
func (a *App) ScryfallUebernehmen(
	ctx context.Context, setID, karteID, name string, felder []string, stand string,
) SchreibAntwort {
	ctx = a.imProfil(ctx)

	setkennung, befund := benannteKennungAus("setID", setID)
	if befund != nil {
		return SchreibAntwort{Fehler: befund}
	}

	kennung, befund := kennungAus(karteID)
	if befund != nil {
		return SchreibAntwort{Fehler: befund}
	}

	if a.scryfall == nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, fmt.Errorf(
			"%w: der Zugang zu Scryfall ist nicht eingerichtet", domain.ErrValidation))}
	}

	spiel, jetzige, fehler := a.kartenstand(ctx, setkennung, kennung)
	if fehler != nil {
		return SchreibAntwort{Fehler: fehler}
	}

	vorschlag, holfehler := a.scryfall.VorschlagFuerKarte(ctx, spiel, jetzige, kennung, name)
	if holfehler != nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, holfehler)}
	}

	// **Der Vorschlag muss noch gelten** (#1241). Dieser Weg liest die Karte
	// neu und rechnet ihn ein zweites Mal; zwischen Zeigen und Übernehmen kann
	// sie sich geändert haben — dann übernähme der Knopf **etwas anderes als
	// das Gezeigte**.
	//
	// **Leer heißt: ohne Prüfung** — dieselbe Regel wie beim Ersetzen. Auf der
	// Kommandozeile, über MCP und über den Dienst gibt es keinen
	// Bestätigungsschritt, den etwas binden könnte.
	if stand != "" && stand != vorschlag.Stand {
		return SchreibAntwort{Fehler: Uebersetze(ctx, fmt.Errorf(
			"%w: der Vorschlag gilt nicht mehr — seit er entstand, hat sich "+
				"an dieser Karte etwas geändert", domain.ErrConflict))}
	}

	_, schreibfehler := a.scryfall.Uebernehmen(ctx, setkennung, kennung, vorschlag, felder)

	return SchreibAntwort{Fehler: Uebersetze(ctx, schreibfehler)}
}

// kartenstand gibt Spiel und heutige Feldwerte einer Karte — über den Leser
// des Fensters.
//
// **Dieselbe Frage wie [Kartenstand], anderer Weg dorthin.** Das Fenster hat
// seine Leser als Schnittstellen (damit ein Test ohne Bestand auskommt); die
// drei anderen Flächen haben eine `*sql.DB`. Was beide teilen, ist die
// **Regel**: die Karte muss zum genannten Set gehören, und ohne Spiel gibt es
// keine Zuordnung.
func (a *App) kartenstand(
	ctx context.Context, setID, karteID int64,
) (spiel string, werte map[string]string, fehler *Fehler) {
	karte, err := a.kartenleser.KarteLesen(ctx, karteID)
	if err != nil {
		return "", nil, Uebersetze(ctx, err)
	}

	if karte.SetID != setID {
		return "", nil, Uebersetze(ctx, fmt.Errorf(
			"%w: die Karte %d gehört nicht zum Set %d", domain.ErrNotFound, karteID, setID))
	}

	koepfe, kopffehler := a.leser().Setkoepfe(ctx)
	if kopffehler != nil {
		return "", nil, Uebersetze(ctx, kopffehler)
	}

	jetzige := make(map[string]string, len(karte.Values))
	for _, wert := range karte.Values {
		jetzige[wert.Field] = wert.Text
	}

	for _, kopf := range koepfe {
		if kopf.ID == setID {
			return kopf.Game, jetzige, nil
		}
	}

	return "", nil, Uebersetze(ctx, fmt.Errorf("%w: das Set %d", domain.ErrNotFound, setID))
}

// MitScryfall setzt den Zugang zum fremden Dienst — die Naht für Tests.
//
// **Und für den Composition Root**: `app.go` baut ihn beim Start. Ohne diese
// Naht ließe sich die Fassade nur mit Netz prüfen, und was sie prüfen soll,
// sind Kennungen, Fehlerübersetzung und DTO.
func (a *App) MitScryfall(quelle Scryfallquelle) { a.scryfall = quelle }
