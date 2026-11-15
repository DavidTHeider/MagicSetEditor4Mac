package app

import (
	"context"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Die Bindungen des Paketverwalters.
//
// **Eine eigene Datei**, weil `bindungen.go` an der Zeilengrenze steht (§2) —
// und weil der Schnitt einer Frage folgt: hier steht, was mit installierten
// Paketen zu tun ist.

// Pakete gibt den Katalog der installierten Pakete.
func (b *Bindungen) Pakete(ctx context.Context) PaketeAntwort {
	return b.app(ctx).Pakete(b.ruf(ctx))
}

// PaketPlanWaehlen fragt nach einem Verzeichnis und plant, es einzubinden —
// **erst der Plan, dann die Wahl** (DOM-01.AK-02, #1696): was es braucht, wen
// es ersetzt und welche Sets es trifft, steht da, bevor etwas geschieht.
//
// **Ein Verzeichnisdialog**, denn ein MSE-Vorlagenpaket ist im Regelfall ein
// Verzeichnis (`magic.mse-game/`). Ein Dateidialog zeigte es unter macOS als
// undurchdringlichen Ordner an, in den man nur hineingehen kann.
//
// **Der Dialog gehört hierher und nicht ins Frontend** — derselbe Grund wie
// beim Öffnen eines Sets: ein Fenster bekäme einen Namen und keinen Pfad.
func (b *Bindungen) PaketPlanWaehlen(ctx context.Context) PaketplanAntwort {
	ctx = b.ruf(ctx)
	texte := DialogtexteFuer(b.app(ctx).sprache)

	pfad, err := b.app(ctx).waehleDatei(ctx, Dateiauswahl{
		Titel:       texte.PaketWaehlen,
		Verzeichnis: true,
	})
	if err != nil {
		return PaketplanAntwort{Fehler: Uebersetze(ctx, err)}
	}

	// **Ein weggeklickter Dialog ist ein Abbruch** und kein Fehler.
	if pfad == "" {
		return PaketplanAntwort{Abgebrochen: true}
	}

	return b.app(ctx).PaketPlan(ctx, string(service.PaketvorgangEinbinden), "", "", pfad)
}

// PaketPlan plant einen Vorgang an einem Paket — DOM-01.AK-02 (#1696).
func (b *Bindungen) PaketPlan(ctx context.Context, vorgang, art, name, quelle string) PaketplanAntwort {
	return b.app(ctx).PaketPlan(b.ruf(ctx), vorgang, art, name, quelle)
}

// PaketEinbindenVon bindet ein Paket aus dem Ort ein, den der Plan nannte.
func (b *Bindungen) PaketEinbindenVon(ctx context.Context, quelle string) PaketAntwort {
	return b.app(ctx).PaketInstallieren(b.ruf(ctx), quelle)
}

// PaketErsetzen legt eine neue Fassung ein — DOM-01.AK-02/AK-03.
func (b *Bindungen) PaketErsetzen(ctx context.Context, quelle string, bestaetigt bool) PaketAntwort {
	return b.app(ctx).PaketErsetzen(b.ruf(ctx), quelle, bestaetigt)
}

// PaketZurueck holt die vorige Fassung zurück.
func (b *Bindungen) PaketZurueck(ctx context.Context, art, name string) PaketAntwort {
	return b.app(ctx).PaketZurueck(b.ruf(ctx), art, name)
}

// PaketAnlegen legt ein neues Paket aus einem Gerüst an (#1350).
//
// **Kein Dialog davor**, anders als bei [Bindungen.PaketPlanWaehlen]: dort
// wird ein Verzeichnis **gesucht**, hier eines **angelegt** — und wo es
// hinkommt, entscheidet nicht die Person, sondern §11: ins verwaltete
// Verzeichnis, wie ein bezogenes Paket.
func (b *Bindungen) PaketAnlegen(ctx context.Context, art, name, vollname, spiel string) PaketAntwort {
	return b.app(ctx).PaketAnlegen(b.ruf(ctx), art, name, vollname, spiel)
}

// PaketKopieren legt eine eigene Kopie eines Pakets an — FMT-02.AK-04
// (#1691). Ohne `neu` heißt sie wie das Original mit `-kopie`.
func (b *Bindungen) PaketKopieren(ctx context.Context, art, name, neu string) PaketAntwort {
	return b.app(ctx).PaketKopieren(b.ruf(ctx), art, name, neu)
}

// PaketArten nennt, was sich anlegen lässt.
//
// **Aus der Anwendung und nicht aus dem Frontend**: eine Liste im TypeScript
// wäre eine zweite Wahrheit, und die Prüfung dahinter steht in `service`.
func (b *Bindungen) PaketArten(ctx context.Context) []Anlegbarkeit {
	b.tor(ctx)

	return Anlegbar()
}

// PaketEntfernen nimmt ein installiertes Paket wieder heraus.
func (b *Bindungen) PaketEntfernen(ctx context.Context, art, name string) SchreibAntwort {
	return b.app(ctx).PaketEntfernen(b.ruf(ctx), art, name)
}

// Abhaengigkeiten prüft die Abhängigkeiten der installierten Pakete.
func (b *Bindungen) Abhaengigkeiten(ctx context.Context) AbhaengigkeitenAntwort {
	return b.app(ctx).Abhaengigkeiten(b.ruf(ctx))
}

// Paketbedarf nennt, was einem Set an Paketen fehlt — DOM-01.AK-04 (#1696).
func (b *Bindungen) Paketbedarf(ctx context.Context, setID string) AbhaengigkeitenAntwort {
	return b.app(ctx).Paketbedarf(b.ruf(ctx), setID)
}

// MSE2Vorschlaege nennt die üblichen Orte einer MSE2-Installation.
//
// **Vorgeschlagen, nicht ungefragt gelesen** (#55): was in einem dieser
// Verzeichnisse liegt, erfährt die Anwendung erst, wenn jemand es wählt.
func (b *Bindungen) MSE2Vorschlaege(ctx context.Context) OrteAntwort {
	return b.app(ctx).MSE2Vorschlaege()
}

// VorlagenUebernehmen fragt nach einem Verzeichnis und holt seine Vorlagen.
func (b *Bindungen) VorlagenUebernehmen(ctx context.Context, vorschlag string) UebernahmeAntwort {
	ctx = b.ruf(ctx)

	if strings.TrimSpace(vorschlag) != "" {
		return b.app(ctx).VorlagenUebernehmen(ctx, vorschlag)
	}

	texte := DialogtexteFuer(b.app(ctx).sprache)

	pfad, err := b.app(ctx).waehleDatei(ctx, Dateiauswahl{
		Titel:       texte.MSE2Waehlen,
		Verzeichnis: true,
	})
	if err != nil {
		return UebernahmeAntwort{
			Uebernommen: []PaketDTO{},
			Uebergangen: []string{},
			Befunde:     []string{},
			Fehler:      Uebersetze(ctx, err),
		}
	}

	return b.app(ctx).VorlagenUebernehmen(ctx, pfad)
}

// Stile nennt die Stile, die für ein Set zur Wahl stehen.
func (b *Bindungen) Stile(ctx context.Context, setID string) StileAntwort {
	return b.app(ctx).Stile(b.ruf(ctx), setID)
}

// Stilfolgen sagt, was ein Wechsel kostet — bevor er stattfindet.
func (b *Bindungen) Stilfolgen(ctx context.Context, setID, paket string) FolgenAntwort {
	return b.app(ctx).Stilfolgen(b.ruf(ctx), setID, paket)
}

// StilWechseln setzt den Stil eines Sets.
func (b *Bindungen) StilWechseln(ctx context.Context, setID, paket string) SchreibAntwort {
	return b.app(ctx).StilWechseln(b.ruf(ctx), setID, paket)
}

// Stiloptionen gibt die Optionen des gewählten Stils.
func (b *Bindungen) Stiloptionen(ctx context.Context, setID string) StiloptionenAntwort {
	return b.app(ctx).Stiloptionen(b.ruf(ctx), setID)
}

// Stilwahlbilder gibt zu einer Stiloption je Wahl ihr Bild
// (#1225).
func (b *Bindungen) Stilwahlbilder(ctx context.Context, setID, option string) WahlbilderAntwort {
	return b.app(ctx).Stilwahlbilder(b.ruf(ctx), setID, option)
}

// StiloptionSetzen schreibt den Wert einer Stiloption.
func (b *Bindungen) StiloptionSetzen(ctx context.Context, setID, feld, wert string) SchreibAntwort {
	return b.app(ctx).StiloptionSetzen(b.ruf(ctx), setID, feld, wert)
}

// StiloptionZuruecksetzen stellt eine Stiloption auf ihre Vorgabe zurück.
func (b *Bindungen) StiloptionZuruecksetzen(ctx context.Context, setID, feld string) SchreibAntwort {
	return b.app(ctx).StiloptionZuruecksetzen(b.ruf(ctx), setID, feld)
}

// Bezugsquellen gibt die Quellen, die die Anwendung anbietet (#263).
func (b *Bindungen) Bezugsquellen(ctx context.Context) BezugsquellenAntwort {
	return b.app(ctx).Bezugsquellen()
}
