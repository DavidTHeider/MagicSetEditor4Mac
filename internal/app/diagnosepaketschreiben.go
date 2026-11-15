package app

// **Das Diagnosepaket schreiben und wieder loswerden** — Werkstück O10 von
// Phase 5.2 (#1434).
//
// # Drei Zusagen
//
//  1. **Geschrieben wird, was gezeigt wurde** — die Teile aus
//     [App.Diagnosepaket], nicht neu zusammengestellte.
//  2. **Ein Abbruch hinterlässt kein scheinbar fertiges Paket** — das Archiv
//     entsteht als Nachbardatei und bekommt seinen Namen erst, wenn es
//     vollständig ist (`sicherung.PackenAls`).
//  3. **Nichts geht von selbst hinaus** — es gibt keinen Weg, der es
//     verschickt.

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// DiagnosepaketNach schreibt das zuletzt gezeigte Paket an einen Ort.
//
// **Für die drei Flächen ohne Fenster** — sie nennen das Ziel; am Fenster
// wählt es ein Dialog ([App.DiagnosepaketSichern]).
func (a *App) DiagnosepaketNach(ctx context.Context, ziel string) SchreibAntwort {
	ctx = a.imProfil(ctx)

	return a.paketSchreiben(ctx, ziel, a.letzteVorschau, a.letzteTeile)
}

// paketSchreiben ist der gemeinsame Weg — mit der Vorschau und den Teilen,
// die geschrieben werden, also schon **ohne** die abgewählten.
func (a *App) paketSchreiben(
	ctx context.Context, ziel string, vorschau *DiagnosepaketAntwort, teile []paketteil,
) SchreibAntwort {
	// **Ohne Vorschau gibt es nichts zu schreiben.** Das ist kein Sonderfall
	// des Fensters, sondern die Zusage selbst: „was du siehst, gibst du
	// weiter" hat ohne ein Gesehenes keinen Inhalt.
	if vorschau == nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, fmt.Errorf(
			"%w: das Diagnosepaket ist noch nicht zusammengestellt worden",
			domain.ErrValidation))}
	}

	if !strings.HasSuffix(ziel, domain.Diagnosepaketendung) {
		return SchreibAntwort{Fehler: Uebersetze(ctx, domain.NewFieldError(
			"ziel", domain.SchluesselValidierung,
			"Ein Diagnosepaket endet auf %s.", domain.Diagnosepaketendung))}
	}

	fehler := service.SchreibeDiagnosepaket(
		ziel, Diagnosemanifest, Fassung, alsArchivteile(*vorschau, teile), time.Now())
	if fehler != nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, domain.MitGrund(
			domain.ErrValidation, fehler,
			"das Diagnosepaket ließ sich nicht schreiben — wähle einen anderen Ort"))}
	}

	a.letzterPaketort = ziel

	return SchreibAntwort{}
}

// DiagnosepaketSichern fragt nach einem Ort und schreibt dorthin — ohne die
// Teile, die `ohne` nennt (OPS-04.AK-04, #1721).
func (a *App) DiagnosepaketSichern(ctx context.Context, ohne []string) SchreibAntwort {
	ctx = a.imProfil(ctx)

	// **Vor dem Dialog ausgewählt** — wer einen Ort wählt, soll danach nicht
	// hören, dass seine Auswahl nicht geht.
	vorschau, teile := a.letzteVorschau, a.letzteTeile
	if vorschau != nil {
		gewaehlt, bleiben, fehler := ohneTeile(*vorschau, teile, ohne)
		if fehler != nil {
			return SchreibAntwort{Fehler: Uebersetze(ctx, fehler)}
		}

		vorschau, teile = &gewaehlt, bleiben
	}

	texte := DialogtexteFuer(a.sprache)

	ziel, fehler := a.waehleDatei(ctx, Dateiauswahl{
		Speichern: true,
		Ablage:    AblageDiagnose,
		Titel:     texte.DiagnosepaketSpeichern,
		Vorschlag: texte.DiagnosepaketVorschlag,
		Filter:    texte.DiagnosepaketFilter,
		Muster:    "*" + domain.Diagnosepaketendung,
	})
	if fehler != nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	// **Abgebrochen ist kein Fehler** — dieselbe Antwort wie bei jedem anderen
	// Dialog dieses Fensters.
	if ziel == "" {
		return SchreibAntwort{}
	}

	// **Die Endung wird angehängt, nicht verlangt.** Wer im Dialog einen Namen
	// tippt, tippt selten die Endung mit; ein Fehler dafür wäre eine Hürde
	// ohne Zweck. An den drei Flächen ohne Fenster gilt das nicht — dort ist
	// der Pfad ein Argument, und ein stilles Anhängen schriebe woandershin,
	// als der Aufrufer sagte.
	if !strings.HasSuffix(ziel, domain.Diagnosepaketendung) {
		ziel += domain.Diagnosepaketendung
	}

	return a.paketSchreiben(ctx, ziel, vorschau, teile)
}

// ohneTeile nimmt die genannten Teile heraus — aus den Teilen und aus der
// Vorschau, die als Übersicht ins Archiv geht.
//
// **Ein unbekannter Name ist ein Fehler**, kein stilles Nichts — wer
// „geheimnisse" herausnehmen will und es gibt keinen Teil dieses Namens,
// soll das hören. **Und ein Paket ohne jeden Teil entsteht nicht.**
func ohneTeile(
	vorschau DiagnosepaketAntwort, teile []paketteil, ohne []string,
) (DiagnosepaketAntwort, []paketteil, error) {
	for _, name := range ohne {
		if !slices.Contains(domain.Diagnoseteile(), domain.Diagnoseteil(name)) {
			return vorschau, nil, domain.NewFieldError("ohne", domain.SchluesselValidierung,
				"»%s« ist kein Teil eines Diagnosepakets.", name)
		}
	}

	heraus := func(art string) bool { return slices.Contains(ohne, art) }

	bleiben := slices.DeleteFunc(slices.Clone(teile), func(teil paketteil) bool {
		return heraus(string(teil.art))
	})
	if len(bleiben) == 0 {
		return vorschau, nil, domain.NewFieldError("ohne", domain.SchluesselValidierung,
			"Ein Diagnosepaket braucht mindestens einen Teil — nimm einen wieder hinein.")
	}

	vorschau.Teile = slices.DeleteFunc(slices.Clone(vorschau.Teile), func(teil PaketteilDTO) bool {
		return heraus(teil.Teil)
	})
	vorschau.Ohne = slices.Sorted(slices.Values(slices.Clone(ohne)))
	vorschau.Bytes = 0

	for _, teil := range vorschau.Teile {
		vorschau.Bytes += teil.Bytes
	}

	if heraus(string(domain.TeilProtokoll)) {
		vorschau.Von, vorschau.Bis = "", ""
	}

	return vorschau, bleiben, nil
}

// DiagnosepaketLoeschen nimmt das zuletzt geschriebene Paket wieder weg.
//
// # Warum **nur** das zuletzt geschriebene
//
// **Weil eine Bindung, die einen beliebigen Pfad löscht, ein Weg ist, mit dem
// ein Skript jede Datei dieses Rechners entfernt.** Das Issue verlangt eine
// „Löschmöglichkeit" für das Paket, das gerade entstanden ist — und genau die
// steht hier. Dieselbe Überlegung wie bei [Bindungen.BeispielsetOeffnen]: die
// Liste ist der Filter, nicht der Aufrufer.
func (a *App) DiagnosepaketLoeschen(ctx context.Context) SchreibAntwort {
	ctx = a.imProfil(ctx)

	if a.letzterPaketort == "" {
		return SchreibAntwort{Fehler: Uebersetze(ctx, fmt.Errorf(
			"%w: es ist kein Diagnosepaket geschrieben worden",
			domain.ErrValidation))}
	}

	if err := os.Remove(a.letzterPaketort); err != nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, domain.MitGrund(
			domain.ErrValidation, err,
			"das Diagnosepaket ließ sich nicht löschen — es steht in %s",
			filepath.Dir(a.letzterPaketort)))}
	}

	a.letzterPaketort = ""

	return SchreibAntwort{}
}

// Diagnosepaket ist die Fensterbindung für die Vorschau.
func (b *Bindungen) Diagnosepaket(ctx context.Context) DiagnosepaketAntwort {
	return b.app(ctx).Diagnosepaket(b.ruf(ctx))
}

// DiagnosepaketSichern ist die Fensterbindung fürs Schreiben.
func (b *Bindungen) DiagnosepaketSichern(ctx context.Context, ohne []string) SchreibAntwort {
	return b.app(ctx).DiagnosepaketSichern(b.ruf(ctx), ohne)
}

// DiagnosepaketLoeschen ist die Fensterbindung fürs Wegnehmen.
func (b *Bindungen) DiagnosepaketLoeschen(ctx context.Context) SchreibAntwort {
	return b.app(ctx).DiagnosepaketLoeschen(b.ruf(ctx))
}
