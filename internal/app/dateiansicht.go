package app

// **Eine Datei ansehen und an macOS geben** — UX-13
// (#1685).
//
// Was eine Datei zeigt, entscheidet [service.DateiAnsehen]; hier wird es zur
// Gestalt des Fensters, und das Fenster bekommt als einzige Fläche die
// Bilddaten — die anderen haben die Datei selbst. Die Übergabe an macOS —
// im Finder zeigen, mit dem Standardprogramm öffnen — steht in
// `dateiaufmacher.go`, weil ihr Ruf einen echten Finder öffnet.

import (
	"context"
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// DateiansichtDTO ist eine Datei, wie Fenster, Kommandozeile, Dienst und
// Werkzeugserver sie zeigen.
type DateiansichtDTO struct {
	Name        string `json:"name"`
	Pfad        string `json:"pfad"`
	Art         string `json:"art"`
	Vorschau    string `json:"vorschau"`
	Groesse     int64  `json:"groesse"`
	Geaendert   string `json:"geaendert"`
	Verzeichnis bool   `json:"verzeichnis"`
	Oeffnen     bool   `json:"oeffnen"`

	Text     string `json:"text"`
	Gekuerzt bool   `json:"gekuerzt"`

	// Bild ist eine Datenadresse — nur am Fenster.
	Bild   string `json:"bild"`
	Breite int    `json:"breite"`
	Hoehe  int    `json:"hoehe"`
	Seiten int    `json:"seiten"`

	Eintraege []DateieintragDTO `json:"eintraege"`
	Unsicher  int               `json:"unsicher"`

	Verschluesselt bool `json:"verschluesselt"`
	ZuGross        bool `json:"zuGross"`
	Unlesbar       bool `json:"unlesbar"`

	Fehler *Fehler `json:"fehler"`
}

// DateieintragDTO ist ein Eintrag eines Pakets oder Archivs.
type DateieintragDTO struct {
	Name        string `json:"name"`
	Groesse     int64  `json:"groesse"`
	Verzeichnis bool   `json:"verzeichnis"`
	Unsicher    bool   `json:"unsicher"`
}

// DateiAnsehen zeigt eine Datei — mit ihrem Bild, wenn sie eines ist.
func (a *App) DateiAnsehen(ctx context.Context, pfad string) DateiansichtDTO {
	ctx = a.imProfil(ctx)

	ansicht, fehler := service.DateiAnsehen(zuhauseAus(pfad))
	if fehler != nil {
		return DateiansichtDTO{Eintraege: []DateieintragDTO{}, Fehler: Uebersetze(ctx, fehler)}
	}

	dto := alsDateiansicht(ansicht)
	if len(ansicht.Bild) > 0 {
		dto.Bild = "data:" + ansicht.Bildtyp + ";base64," + base64.StdEncoding.EncodeToString(ansicht.Bild)
	}

	return dto
}

// DateiansichtAus ist die Ansicht für die Flächen ohne Fenster — ohne
// Bilddaten. Ein Fehler steht darin wie an der Bindung.
func DateiansichtAus(pfad string) DateiansichtDTO {
	ansicht, fehler := service.DateiAnsehen(pfad)
	if fehler != nil {
		return DateiansichtDTO{Eintraege: []DateieintragDTO{}, Fehler: Uebersetze(context.Background(), fehler)}
	}

	return alsDateiansicht(ansicht)
}

func alsDateiansicht(ansicht service.Dateiansicht) DateiansichtDTO {
	eintraege := make([]DateieintragDTO, 0, len(ansicht.Eintraege))
	for _, eintrag := range ansicht.Eintraege {
		eintraege = append(eintraege, DateieintragDTO(eintrag))
	}

	return DateiansichtDTO{
		Name: ansicht.Name, Pfad: ansicht.Pfad, Art: string(ansicht.Regel.Art),
		Vorschau: string(ansicht.Regel.Vorschau), Groesse: ansicht.Groesse,
		Geaendert: alsZeitpunkt(ansicht.Geaendert), Verzeichnis: ansicht.Verzeichnis,
		Oeffnen: ansicht.Regel.Oeffnen && !ansicht.Verzeichnis,
		Text:    ansicht.Text, Gekuerzt: ansicht.Gekuerzt,
		Breite: ansicht.Breite, Hoehe: ansicht.Hoehe, Seiten: ansicht.Seiten,
		Eintraege: eintraege, Unsicher: ansicht.Unsicher,
		Verschluesselt: ansicht.Verschluesselt, ZuGross: ansicht.ZuGross, Unlesbar: ansicht.Unlesbar,
	}
}

// Dateiaufmacher gibt eine Datei an macOS — die Argumente von `open`.
type Dateiaufmacher func(ctx context.Context, argumente ...string) error

// dateiZeigen zeigt eine Datei im Finder.
func (a *App) dateiZeigen(ctx context.Context, pfad string, aufmachen Dateiaufmacher) SchreibAntwort {
	return a.dateiUebergeben(ctx, pfad, false, aufmachen, "-R", pfad)
}

// dateiOeffnen gibt eine Datei dem Standardprogramm — wenn ihre Regel es
// erlaubt.
func (a *App) dateiOeffnen(ctx context.Context, pfad string, aufmachen Dateiaufmacher) SchreibAntwort {
	return a.dateiUebergeben(ctx, pfad, true, aufmachen, pfad)
}

func (a *App) dateiUebergeben(
	ctx context.Context, pfad string, oeffnen bool, aufmachen Dateiaufmacher, argumente ...string,
) SchreibAntwort {
	pfad = zuhauseAus(pfad)
	for stelle, argument := range argumente {
		argumente[stelle] = zuhauseAus(argument)
	}

	if fehler := service.DateiUebergabePruefen(pfad, oeffnen); fehler != nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	if fehler := aufmachen(ctx, argumente...); fehler != nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, domain.MitGrund(domain.ErrValidation, fehler,
			"der Finder nimmt die Datei nicht an"))}
	}

	return SchreibAntwort{}
}

// DateiAnsehen ist die Bindung.
func (b *Bindungen) DateiAnsehen(ctx context.Context, pfad string) DateiansichtDTO {
	return b.app(ctx).DateiAnsehen(b.ruf(ctx), pfad)
}

// zuhauseAus löst eine Tilde am Anfang zum Zuhause auf — **nur am Fenster**.
//
// Die Orte der Hilfe stehen als Tilde da (`schwaerze`), und wer daneben
// „Ansehen" drückt, meint die Datei. **Der Dienst tut das nicht**: er prüft
// den Pfad vor dem Ansehen gegen den Pfadvertrag, und eine Tilde danach
// aufzulösen hieße, einen anderen Pfad zu prüfen als den gelesenen.
func zuhauseAus(pfad string) string {
	rest, mitTilde := strings.CutPrefix(pfad, "~/")
	if !mitTilde {
		return pfad
	}

	zuhause, fehler := os.UserHomeDir()
	if fehler != nil {
		return pfad
	}

	return filepath.Join(zuhause, rest)
}
