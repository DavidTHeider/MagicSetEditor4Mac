package app

// **Die digitale Deckleihe am Fenster** — Werkstück G6 von Phase 5.5
// (#1463).

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/austausch"
)

// Deckleihen gibt alles, was dieses Profil an Leihen weiß.
func (b *Bindungen) Deckleihen(ctx context.Context) DeckleiheAntwort {
	return b.app(ctx).Deckleihen(b.ruf(ctx))
}

// DeckleiheAnbieten stellt eine Leihe aus und legt sie als Umschlag ab.
//
// **Ein Verbund und keine acht Argumente.** Wails bildet ihn als Abbildung
// ab; anders als bei `PartieZeigen` sind es hier aber **benannte Felder mit
// verschiedenen Bedeutungen** — acht Stellen hintereinander wären die Naht,
// an der jemand zwei vertauscht, und beide sind Zeichenketten.
func (b *Bindungen) DeckleiheAnbieten(
	ctx context.Context,
	wunsch LeihwunschDTO, ziel string,
) DeckleiheAntwort {
	return b.app(ctx).DeckleiheAnbieten(b.ruf(ctx), wunsch, ziel)
}

// DeckleiheAnnehmen liest einen Umschlag und nimmt die Leihe darin an.
func (b *Bindungen) DeckleiheAnnehmen(ctx context.Context, pfad string) DeckleiheAntwort {
	return b.app(ctx).DeckleiheAnnehmen(b.ruf(ctx), pfad)
}

// DeckleiheWiderrufen nimmt eine ausgestellte Leihe zurück.
//
// **Und holt nichts ein.** Was hinausging, ist draußen; was der Widerruf
// bewirkt, ist, dass hier kein neuer Start mehr damit beginnt — und drüben,
// sobald die Nachricht ankommt.
func (b *Bindungen) DeckleiheWiderrufen(ctx context.Context, kennung string) DeckleiheAntwort {
	return b.app(ctx).DeckleiheWiderrufen(b.ruf(ctx), kennung)
}

// DeckleiheZurueckgeben trägt ein, dass eine geliehene nicht mehr gilt.
func (b *Bindungen) DeckleiheZurueckgeben(ctx context.Context, kennung string) DeckleiheAntwort {
	return b.app(ctx).DeckleiheZurueckgeben(b.ruf(ctx), kennung)
}

// LeihzielAntwort trägt den Ort, an den ein Umschlag gelegt werden soll.
type LeihzielAntwort struct {
	Pfad string `json:"pfad"`

	// Abgebrochen sagt, dass niemand einen Ort gewählt hat — **kein
	// Fehler**, und das Fenster lässt dann den vorigen stehen.
	Abgebrochen bool `json:"abgebrochen"`

	Fehler *Fehler `json:"fehler"`
}

// DeckleiheZielWaehlen fragt nach dem Ort für den Umschlag einer Leihe und
// gibt ihn zurück — UXR-F15
// (#2030).
//
// **Rein lesend**: der Dialog legt nichts ab. Geschrieben wird erst von
// [Bindungen.DeckleiheAnbieten], mit dem Pfad, den das Fenster bis dahin
// zeigt und ändern lässt. Ein Pfad von Hand ist auf macOS kein Bedienelement
// (R7 in `docs/UIREVIEW.md`).
//
// **Ein Sichern-Dialog mit Namen und Endung**, die Endung aus
// [austausch.Endung]. Das Gegenstück zum Annehmen ist
// [Bindungen.UmschlagWaehlen]: dieselbe Endung, derselbe Filter — und kein
// zweiter Dialog für denselben Umschlag.
//
// **Am Fenster und nirgends sonst** (§16): an Kommandozeile und MCP bleibt
// der Pfad ein Parameter, denn einen Dateiwahldialog ohne Fenster gibt es
// nicht.
func (b *Bindungen) DeckleiheZielWaehlen(ctx context.Context) LeihzielAntwort {
	ctx = b.ruf(ctx)
	texte := DialogtexteFuer(b.app(ctx).sprache)

	pfad, fehler := b.app(ctx).waehleDatei(ctx, Dateiauswahl{
		Titel:     texte.LeiheAblegen,
		Filter:    texte.UmschlagFilter,
		Muster:    "*" + austausch.Endung,
		Vorschlag: texte.LeiheVorschlag,
		Speichern: true,
	})
	if fehler != nil {
		return LeihzielAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	// **Ein leerer Pfad ist ein Abbruch** — derselbe Grund wie bei
	// [Bindungen.UmschlagSchnueren].
	return LeihzielAntwort{Pfad: pfad, Abgebrochen: pfad == ""}
}
