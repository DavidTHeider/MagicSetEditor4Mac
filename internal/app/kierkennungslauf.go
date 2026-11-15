package app

// **Wie ein Kartenbild an ein Modell kommt** — Werkstück A13.3 von Phase 5.3
// (#1452).
//
// # Warum das eine eigene Datei ist
//
// **Weil es eine andere Frage beantwortet.** Nebenan steht, *was* eine
// Erkennung ist — Bezug, Antwort, Bindung, und was davon nach draußen geht.
// Hier steht, *wie* sie zustande kommt: welche Datei gelesen wird, in welche
// Form die Antwort gezwungen wird und was von ihr übrig bleibt.
//
// Der Schnitt folgt der Frage und nicht der Zeilenzahl (§2) — die 400 waren
// der Anlass, nicht der Grund.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"path/filepath"
	"strings"

	// Für [image.DecodeConfig]: dieselben drei Formate, die auch das Bildfeld
	// annimmt (`bildsetzen.go`) — BMP und TIFF kennt die Standardbibliothek
	// nicht, und kein KI-Anbieter nimmt sie entgegen.
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// erkennungslauf ist, was ein Lauf mitbringt.
type erkennungslauf struct {
	profil    int64
	bezug     KIErkennungsbezug
	bild      []byte
	medientyp string
}

// sichtbildLesen holt die Datei und sagt, was für eine es ist.
//
// **Weder `bildLesen` noch `bilddateiLesen`**: den ersten Namen trägt das
// Nachzeichnen eines Symbols, den zweiten das Setzen eines Kartenbilds, und
// beide geben etwas anderes zurück. Drei Bedeutungen unter zwei Namen wären
// eine Falle für den nächsten Leser.
//
// # Warum der Kopf entschlüsselt und nicht beschnuppert wird
//
// **Der erste Anlauf nahm `http.DetectContentType`** — und `internal/app`
// darf `net/http` nicht kennen (`internal/schichten/netz_test.go`): die
// Zusage „diese Anwendung lädt von sich aus nichts" hängt an einer
// Aufzählung, und `app` in sie aufzunehmen machte sie wertlos. Die Funktion
// lädt nichts, aber die Grenze läuft am **Paket** und nicht an der Absicht.
//
// **Und [image.DecodeConfig] ist ohnehin die bessere Auskunft**: sie liest den
// Kopf wirklich, statt die ersten Bytes zu vergleichen. Eine Datei, die mit
// der PNG-Signatur beginnt und danach Unsinn trägt, kommt hier nicht durch —
// beim Anbieter käme sie bis zur Abrechnung.
//
// **Der Medientyp kommt damit aus dem Inhalt und nicht aus der Endung**: eine
// Datei, die `.png` heißt und ein JPEG ist, bekäme sonst eine Kopfzeile, die
// der Anbieter zu Recht abweist.
func sichtbildLesen(pfad string) (bild []byte, medientyp string, fehler error) {
	if strings.TrimSpace(pfad) == "" {
		return nil, "", fmt.Errorf(
			"%w: es wurde keine Bilddatei genannt", domain.ErrValidation)
	}

	// **Mit dem Budget einer Paketdatei** — Befund 2 des Nachlaufs über Phase
	// 6.1: den Pfad nennen auch MCP und der Dienst, und eine Datei über jeder
	// Grenze belegte sonst den Speicher, bevor irgendetwas geprüft war.
	roh, lesefehler := msepaket.LeseDateiBegrenzt(filepath.Clean(pfad))
	if errors.Is(lesefehler, msepaket.ErrZuGross) {
		return nil, "", domain.MitGrund(domain.ErrValidation, lesefehler,
			"%s ist größer, als ein Bild hier sein darf", filepath.Base(pfad))
	}

	if lesefehler != nil {
		// **Der Satz nennt die Datei, der Grund geht ins Protokoll** (§8,
		// [0075](../../docs/entscheidungen/0075-der-satz-und-sein-grund.md)):
		// „permission denied" hilft niemandem vor dem Fenster, und der volle
		// Pfad trüge ein Stück Verzeichnisbaum in die Oberfläche.
		return nil, "", domain.MitGrund(domain.ErrNotFound, lesefehler,
			"%s ließ sich nicht lesen", filepath.Base(pfad))
	}

	_, gestalt, deutefehler := image.DecodeConfig(bytes.NewReader(roh))
	if deutefehler != nil {
		return nil, "", domain.MitGrund(domain.ErrValidation, deutefehler,
			"%s ist kein lesbares Bild — erkannt werden PNG, JPEG und GIF",
			filepath.Base(pfad))
	}

	return roh, "image/" + gestalt, nil
}

// erkennungsschema zwingt die Antwort in die Form der Beobachtungen.
func erkennungsschema() map[string]any {
	arten := make([]string, 0, len(domain.Evidenzarten()))
	for _, art := range domain.Evidenzarten() {
		arten = append(arten, string(art))
	}

	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"beobachtungen"},
		"properties": map[string]any{
			"beobachtungen": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type":                 "object",
					"additionalProperties": false,
					"required":             []string{"art", "lesbar"},
					"properties": map[string]any{
						"art":    map[string]any{"type": "string", "enum": arten},
						"wert":   map[string]any{"type": "string"},
						"lesbar": map[string]any{"type": "boolean"},
					},
				},
			},
		},
	}
}

// erkennungsantwort ist, was das Modell schickt.
type erkennungsantwort struct {
	Beobachtungen []struct {
		Art    string `json:"art"`
		Wert   string `json:"wert"`
		Lesbar bool   `json:"lesbar"`
	} `json:"beobachtungen"`
}

// erkennenMit fährt den Lauf mit einer fertigen Sichtquelle.
func (a *App) erkennenMit(
	ctx context.Context, lauf erkennungslauf,
	werk *service.KIAuftragswerk, sicht KISichtquelle,
) KIErkennungAntwort {
	// **Geprüft wird, ohne einen Auftrag anzulegen.**
	//
	// [service.KIAuftragswerk.Anlegen] friert Revision und Schema **einer
	// Karte** ein — eine Erkennung hat keine. Was sie braucht, ist die Regel
	// dahinter: „vision belegt" und „darf hier rechnen".
	if _, fehler := werk.Pruefen(service.KIWunsch{
		ProfilID: lauf.profil, Faehigkeit: domain.KIVision,
		Anbieter: lauf.bezug.Anbieter, Alias: lauf.bezug.Alias,
	}); fehler != nil {
		return erkennungsfehler(ctx, fehler)
	}

	// **Gebucht ist schon** — [App.KIKarteErkennen] tut es vor dem
	// Werkzeugsatz und schließt den Posten, wenn dieser Lauf zurückkommt.
	roh, fehler := sicht.Sehen(ctx, KISichtwunsch{
		Anweisung: KIErkennungsanweisung, Alias: lauf.bezug.Alias,
		Auftrag: "Lies diese Karte.", Bild: lauf.bild,
		Medientyp: lauf.medientyp, Schema: erkennungsschema(),
	})
	if fehler != nil {
		return erkennungsfehler(ctx, fehler)
	}

	erkennung, fehler := erkennungAus(roh)
	if fehler != nil {
		return erkennungsfehler(ctx, fehler)
	}

	deutung, fehler := a.bilderkenner.Deuten(ctx, erkennung)
	if fehler != nil {
		return erkennungsfehler(ctx, fehler)
	}

	return alsErkennungsantwort(deutung, lauf.bezug.Anbieter, roh.Modell)
}

// erkennungAus liest die Antwort des Modells.
func erkennungAus(roh KIRohantwort) (domain.Bilderkennung, error) {
	if roh.Abgeschnitten {
		return domain.Bilderkennung{}, fmt.Errorf(
			"%w: die Antwort ist abgeschnitten — eine halbe Leseliste sieht "+
				"aus wie eine ganze", domain.ErrValidation)
	}

	var gelesen erkennungsantwort

	if fehler := json.Unmarshal([]byte(roh.Inhalt), &gelesen); fehler != nil {
		// **Dieselbe Trennung**: was an der Stelle 412 stand, ist eine
		// Auskunft übers Protokoll und nicht über die Karte.
		return domain.Bilderkennung{}, domain.MitGrund(
			domain.ErrValidation, fehler,
			"die Antwort des Modells ist kein brauchbares JSON")
	}

	erkennung := domain.Bilderkennung{
		Beobachtungen: make([]domain.Beobachtung, 0, len(gelesen.Beobachtungen)),
		NichtGelesen:  []domain.Evidenzart{},
	}

	genannt := map[domain.Evidenzart]bool{}

	for _, eine := range gelesen.Beobachtungen {
		erkennung.Beobachtungen = append(erkennung.Beobachtungen,
			domain.Beobachtung{
				Art: domain.Evidenzart(eine.Art), Wert: eine.Wert,
				Lesbar: eine.Lesbar,
			})

		genannt[domain.Evidenzart(eine.Art)] = true
	}

	// **Was nicht genannt wurde, steht ausdrücklich da.**
	//
	// Das ist die Zusage aus [domain.Bilderkennung.NichtGelesen]: eine
	// kürzere Liste sieht aus wie ein Versehen, und „zur Sammlernummer steht
	// nichts" ist eine Antwort.
	for _, art := range domain.Evidenzarten() {
		if !genannt[art] {
			erkennung.NichtGelesen = append(erkennung.NichtGelesen, art)
		}
	}

	return erkennung, nil
}

// erkennungBuchen reserviert den Kostenposten dieser Sichtanfrage.
//
// **Derselbe Weg wie beim Batch** ([App.batchBuchen]): die Kostenwache sieht
// eine Anfrage, gleich ob dahinter eine Karte steht — und ein Bild ist dabei
// die größte von allen.
func (a *App) erkennungBuchen(ctx context.Context, bezug KIErkennungsbezug) (
	string, error,
) {
	budget, fehler := a.budgetFuer(ctx, bezug.Anbieter,
		a.adresseVon(ctx, bezug.Anbieter),
		kideckung{nachweis: bezug.Nachweis})
	if fehler != nil {
		return "", fehler
	}

	return budget.Reservieren(ctx,
		fmt.Sprintf("kisicht-%d", a.jetzt().UnixNano()))
}
