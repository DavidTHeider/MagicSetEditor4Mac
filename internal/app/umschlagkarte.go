package app

// **Karte und Vorlage aus einem Umschlag annehmen** —
// #1639.
//
// # Warum das Ziel im Ruf steht und nicht im Umschlag
//
// **Weil der Umschlag es nicht nennt.** Eine Karte gehört in ein **Set**,
// eine Vorlage in ein **Paket**; welches, entscheidet ein Mensch. Ein Weg,
// der es riete, legte Karten in irgendein Set — und ein Weg, der ohne Ziel
// abbräche, wäre eine Bindung, die nie glückt.
//
// **Die Kennung aus dem Umschlag ist hier keine.** Sie stammt aus einem
// fremden Bestand; wer sie übernähme, schriebe über eine Karte, von der der
// Absender nichts weiß. Was entsteht, ist **neu**, und die Antwort nennt es.
//
// # Und warum ein vorhandenes Paket abgewiesen wird
//
// **Weil den Namen der Mensch gewählt hat, der gerade annimmt** — anders als
// bei einem Deck, dessen Name aus dem Umschlag kommt und deshalb einen
// sichtbaren Zusatz bekommt ([App.deckAnnehmen]). Wer einen Namen wählt, den
// es schon gibt, hat sich vertan; ihm einen anderen unterzuschieben wäre
// dieselbe stille Umbenennung, die dort ausdrücklich abgelehnt ist.

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/austausch"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// UmschlagKarte legt eine empfangene Karte in ein **gewähltes** Set.
//
// **Eine Karte, die nicht ganz passt, ist ein Befund und kein Grund, sie zu
// verwerfen** (§8): ein Feld, das dieses Spiel nicht kennt, endet mit einem
// Satz, der es nennt — die Karte steht dann da, und ihre Kennung kommt mit
// zurück. Dieselbe Gestalt wie bei einem Deckplatz, den niemand legen kann.
func (a *App) UmschlagKarte(
	ctx context.Context, pfad, setID string,
) AnlegeAntwort {
	ctx = a.imProfil(ctx)

	sendung, fehler := a.umschlagLesen(pfad)
	if fehler != nil {
		return AnlegeAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	defer func() { _ = os.RemoveAll(sendung.ort) }() //nolint:errcheck // Wegwerfort

	roh, fehler := lesenAusSendung(sendung, austausch.ArtKarte, Kartendatei)
	if fehler != nil {
		return AnlegeAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	var gelesen KarteDTO

	if fehler := json.Unmarshal(roh, &gelesen); fehler != nil {
		return AnlegeAntwort{Fehler: Uebersetze(ctx, domain.MitGrund(
			domain.ErrValidation, fehler, "die Karte lesen"))}
	}

	return a.karteAnnehmen(ctx, setID, gelesen)
}

// karteAnnehmen legt die gelesene Karte an und setzt ihre Werte.
func (a *App) karteAnnehmen(
	ctx context.Context, setID string, gelesen KarteDTO,
) AnlegeAntwort {
	angelegt := a.KarteAnlegen(ctx, setID)
	if angelegt.Fehler != nil {
		return angelegt
	}

	// **Die Werte gehen einzeln hinein** — dieselbe Gestalt wie bei den
	// Plätzen eines angenommenen Decks.
	for _, wert := range gelesen.Werte {
		geschrieben := a.FeldSetzen(ctx, setID, angelegt.KarteID,
			wert.Feld, wert.Wert)
		if geschrieben.Fehler != nil {
			return AnlegeAntwort{
				KarteID: angelegt.KarteID, Fehler: geschrieben.Fehler,
			}
		}
	}

	return angelegt
}

// UmschlagVorlage installiert eine empfangene Vorlage als **gewähltes** Paket.
//
// **`name` trägt seine Endung** — `meinspiel.mse-game`. Sie sagt, was das
// Paket ist; der Umschlag sagt es nicht, denn er trägt die **Dateien** und
// nicht den Namen, aus dem sie stammen.
//
// # Warum installiert und nicht Datei für Datei geschrieben
//
// **Weil eine Vorlage ein Paket ist und kein Stapel Dateien.** Der Weg über
// die Paketwerkstatt könnte nur **ersetzen**, was schon dasteht
// (`msepaket.DateiSchreiben` liest die Datei, bevor es schreibt) — von einem
// empfangenen Paket käme damit an, was das leere Gerüst zufällig schon
// hatte. [App.PaketInstallieren] nimmt ein ganzes Verzeichnis, und **ein
// vorhandenes Paket ist ihm ein Konflikt und kein Überschreiben** — genau
// die Zusage, die dieses Issue verlangt.
func (a *App) UmschlagVorlage(
	ctx context.Context, pfad, name string,
) PaketAntwort {
	ctx = a.imProfil(ctx)

	sendung, fehler := a.umschlagLesen(pfad)
	if fehler != nil {
		return PaketAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	defer func() { _ = os.RemoveAll(sendung.ort) }() //nolint:errcheck // Wegwerfort

	dateien, fehler := vorlagenAusSendung(sendung)
	if fehler != nil {
		return PaketAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	gebaut, fehler := vorlagePacken(sendung, name, dateien)
	if fehler != nil {
		return PaketAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	return a.PaketInstallieren(ctx, gebaut)
}

// Vorlagendatei ist eine Datei aus einem Umschlag.
//
// **Eine Scheibe und keine Abbildung**: die Reihenfolge entscheidet, was
// zuerst geschrieben wird, und eine Abbildung hätte bei jedem Lauf eine
// andere — ein Fehler bei der dritten Datei träfe mal die eine, mal die
// andere.
type Vorlagendatei struct {
	Name   string
	Inhalt string
}

// vorlagePacken legt die Dateien als Paketverzeichnis in den Wegwerfort.
//
// **Neben dem entpackten Umschlag und nicht darin**: was dort liegt, ist
// durch das Manifest gedeckt; was hier entsteht, ist ein Paket, und die
// Installation kopiert es weiter.
//
// **Der Name darf kein Verzeichnis sein.** Er kommt von einem Menschen, und
// ein `../` darin legte ein Paket irgendwohin.
func vorlagePacken(
	sendung geleseneSendung, name string, dateien []Vorlagendatei,
) (string, error) {
	if name == "" || name == "." || name == ".." ||
		name != path.Base(name) ||
		strings.ContainsRune(name, os.PathSeparator) {
		return "", fmt.Errorf(
			"%w: %q ist kein Paketname — er nennt kein Verzeichnis und trägt "+
				"seine Endung, etwa %q",
			domain.ErrValidation, name, "meinspiel.mse-game")
	}

	// **Die Wurzel liegt am Wegwerfort und nicht am neuen Paket.** Sie hält
	// jeden Namen drin, der von hier aus geschrieben wird — auch den des
	// Pakets selbst.
	wurzel, fehler := os.OpenRoot(sendung.ort)
	if fehler != nil {
		return "", domain.MitGrund(domain.ErrValidation, fehler,
			"den Umschlag öffnen")
	}

	defer func() { _ = wurzel.Close() }() //nolint:errcheck // nur geschrieben

	if fehler := wurzel.MkdirAll(name, 0o750); fehler != nil {
		return "", domain.MitGrund(domain.ErrValidation, fehler,
			"das Paket %q anlegen", name)
	}

	for _, datei := range dateien {
		if fehler := wurzel.WriteFile(
			path.Join(name, datei.Name), []byte(datei.Inhalt),
			0o600); fehler != nil {
			return "", domain.MitGrund(domain.ErrValidation, fehler,
				"%s in das Paket legen", datei.Name)
		}
	}

	return filepath.Join(sendung.ort, name), nil
}

// vorlagenAusSendung liest die Vorlagendateien einer Sendung.
//
// **Über das Manifest und nicht über das Verzeichnis**: was im Umschlag liegt
// und nicht im Manifest steht, ist nicht unterschrieben — und was nicht
// unterschrieben ist, geht hier nicht in ein Paket (N4).
func vorlagenAusSendung(sendung geleseneSendung) ([]Vorlagendatei, error) {
	dateien := []Vorlagendatei{}

	for _, objekt := range sendung.Manifest.Objekte {
		if objekt.Art != austausch.ArtVorlage {
			continue
		}

		inhalt, fehler := datauslesen(sendung, objekt.Name)
		if fehler != nil {
			return nil, fehler
		}

		dateien = append(dateien, Vorlagendatei{
			Name: path.Base(objekt.Name), Inhalt: inhalt,
		})
	}

	if len(dateien) == 0 {
		return nil, fmt.Errorf("%w: dieser Umschlag enthält keine %s",
			domain.ErrValidation, austausch.ArtVorlage)
	}

	return dateien, nil
}

// datauslesen holt eine Datei aus dem Wegwerfverzeichnis der Sendung.
//
// **Über `os.Root` und nicht über einen gefügten Pfad**: der Name kommt aus
// einem fremden Manifest, und ein `../` darin führte aus dem Verzeichnis
// hinaus. Die Wurzel hält ihn drin.
func datauslesen(sendung geleseneSendung, name string) (string, error) {
	wurzel, fehler := os.OpenRoot(sendung.Verzeichnis)
	if fehler != nil {
		return "", domain.MitGrund(domain.ErrValidation, fehler,
			"den Umschlag öffnen")
	}

	defer func() { _ = wurzel.Close() }() //nolint:errcheck // nur gelesen

	roh, fehler := wurzel.ReadFile(strings.TrimPrefix(name, "./"))
	if fehler != nil {
		return "", domain.MitGrund(domain.ErrValidation, fehler,
			"%s aus dem Umschlag lesen", name)
	}

	return string(roh), nil
}
