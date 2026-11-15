package app

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Bindungen ist die Oberfläche, die Wails dem Frontend hinhält.
//
// # Warum es diesen Typ gibt, und warum er teuer bezahlt ist
//
// Wails bindet **jede exportierte Methode** des übergebenen Verbunds. Bis zu
// diesem Punkt war das `*App` — und damit standen `Starten`, `Beenden` und die
// sieben `Mit…`-Nähte im Fenster: `window.go.app.App.MitBestand(null)` war von
// der Oberfläche aus erreichbar und hätte den Bestand ausgehängt.
//
// Schwerer wog das zweite: **Wails reicht keinen `context.Context` hinein.**
// `ParseArgs` verlangt genau so viele Argumente, wie die Go-Signatur führt, und
// der Kontext zählt mit. Jeder Aufruf aus dem Frontend kam deshalb mit einem
// Argument zu wenig an — und blieb **ohne Antwort**: die Zusage wurde nie
// erfüllt und nie abgelehnt, die Seite stand für immer bei „lädt …".
//
// Nachgemessen in der laufenden Anwendung über `wails dev`:
//
//	ERR | error parsing arguments: received 0 arguments to method 'app.App.Sets', expected 1
//
// Gefunden hat es der Zweitblick über Phase 5. Kein Test konnte es finden: die
// Go-Tests rufen `App` unmittelbar, und die Frontend-Vorrichtung legt eine
// eigene Fassade unter `globalThis.go` — beide Seiten waren grün, und dazwischen
// lag die Naht, die niemand betreten hat (AGENTS.md §3).
//
// # Was dieser Typ deshalb tut
//
// Er nennt die gebundene Oberfläche **ausdrücklich**. **Und seit BK-01 nimmt
// jede Methode den Kontext ihres Rufs** (#2109,
// `docs/entscheidungen/0170-der-rufkontext-endet-mit-dem-ruf.md`): Wails 3
// reicht ihn als ersten Parameter hinein (`needsContext`) und bricht ihn ab,
// wenn die Oberfläche den Ruf aufgibt oder das Fenster sich schließt. Bis
// dahin kam der Kontext aus dem Fenster (`Bereit`) — und ein abgebrochener
// Ruf wartete weiter am Pool. Profil und Vorgang hängt [App.imProfil] an,
// wie zuvor.
//
// # Und sie kommt an die Anwendung nur durch das Tor der Sitzung
//
// **Das Feld heißt nicht `app`**, damit keine Methode es erreicht, ohne
// [Bindungen.app] zu rufen — und das hält gesperrt an (ID-02, Nachlauf,
// Befund 1). Bis dahin stand das Tor in [App.kontext]; wer den Kontext nicht
// brauchte oder vorher etwas tat, ging vorbei.
type Bindungen struct {
	anwendung *App
}

// NeueBindungen baut die gebundene Oberfläche über einer Anwendung.
func NeueBindungen(anwendung *App) *Bindungen { return &Bindungen{anwendung: anwendung} }

// app ist der Weg jeder gebundenen Methode zur Anwendung — **durch das Tor**.
func (b *Bindungen) app(ctx context.Context) *App {
	b.anwendung.tore(ctx)

	return b.anwendung
}

// tor hält eine gebundene Methode an, die die Anwendung nicht braucht — eine
// Rechnung über ihre Eingabe. **Auch sie hält**: gesperrt ist gesperrt, und
// eine Ausnahme je Art wäre eine Liste, die niemand pflegt.
func (b *Bindungen) tor(ctx context.Context) { b.anwendung.tore(ctx) }

// frei ist der Weg der Sperranzeige — **ohne Sitzungs- und Schaltertor**,
// und nur für die Methoden in [sitzungsfrei]. `sitzungstor_intern_test.go`
// hält beides zusammen. **Das Lebenstor gilt auch hier** (A02): ohne
// Sitzung heißt nicht ohne Bestand.
func (b *Bindungen) frei() *App {
	b.anwendung.lebenstor(context.Background())

	return b.anwendung
}

// ruf gibt den Kontext eines Rufs — **hinter dem Tor**, und mit seinem
// Ende.
//
// **Der Kontext kommt von Wails und nicht vom Fenster** (BK-01, #2109). Bis
// dahin trug jeder Ruf den Kontext aus [App.Bereit]: Wails brach einen Ruf
// ab, wenn das Fenster es verlangte oder sich schloss, und die Goroutine
// darunter wartete weiter — am Pool, der **eine** Verbindung hat, oder am
// Tor, das Einstellungen liest. Jetzt endet sie mit dem Ruf.
//
// **Was länger leben soll als die Antwort, löst sich selbst** — Partie,
// Horchen, Fernhorchen und die zwei Dienste nehmen
// `context.WithoutCancel` und gehören ihrem eigenen Beenden
// (`docs/entscheidungen/0170-der-rufkontext-endet-mit-dem-ruf.md`).
func (b *Bindungen) ruf(ctx context.Context) context.Context {
	return b.app(ctx).rufkontext(ctx)
}

// fensterkontext gibt den Kontext des Fensters — oder einen leeren.
//
// Vor dem ersten `OnStartup` gibt es kein Fenster. Ein Aufruf davor kann nicht
// aus dem Frontend kommen (es gibt noch keins), wohl aber aus einem Test.
func (a *App) fensterkontext() context.Context {
	// **Und hier entsteht die Vorgangskennung** — Werkstück O9 von Phase 5.2
	// (#1433). Dies ist die zweite Hälfte der Fenstergrenze: was über
	// [Bindungen] hereinkommt, geht durch diese Zeile; was Wails direkt an
	// eine Methode mit Kontext reicht, durch [App.imProfil]. Beide Stellen
	// würfeln nur, wenn noch keine da ist — zusammen ergeben sie **eine**
	// Kennung je Ruf und nicht zwei.
	if a.fenster != nil {
		return MitNeuemVorgang(a.fenster)
	}

	return MitNeuemVorgang(context.Background())
}

// Ab hier: die Oberfläche. Jede Methode reicht durch und tut sonst nichts —
// eine Entscheidung an dieser Stelle wäre eine, die kein Test sieht.

// **Ein Set öffnen geht am Fenster über die Vorschau** —
// [Bindungen.SetVorschauWaehlen] fragt nach der Datei, [Bindungen.SetAusDatei]
// liest sie ein (FMT-03, #1692). Der Dialog gehört hierher und nicht ins
// Frontend: ein Fenster bekäme einen Namen und keinen Pfad.

// Beispielsets nennt die Sets, die das Bündel mitbringt (#702).
func (b *Bindungen) Beispielsets(ctx context.Context) BeispielsetsAntwort {
	return b.app(ctx).Beispielsets(b.ruf(ctx))
}

// BeispielsetOeffnen schlägt ein mitgeliefertes Set auf — **ohne Dialog**.
//
// **Der Pfad kommt aus [Bindungen.Beispielsets] und nicht vom Fenster.** Wer
// hier einen beliebigen entgegennähme, hätte einen Weg, über den ein Skript
// jede Datei dieses Rechners in den Bestand zöge; die Liste ist der Filter.
func (b *Bindungen) BeispielsetOeffnen(ctx context.Context, pfad string) OeffnenAntwort {
	for _, satz := range b.app(ctx).Beispielsets(b.ruf(ctx)).Sets {
		if satz.Pfad == pfad {
			return b.app(ctx).SetOeffnen(b.ruf(ctx), pfad)
		}
	}

	return OeffnenAntwort{
		Befunde:  []string{},
		Angelegt: []string{},
		Fehler: &Fehler{
			Code:       CodeNichtGefunden,
			Schluessel: domain.SchluesselNichtGefunden,
			Meldung:    rueckfallmeldung,
		},
	}
}

// MTGJSONOeffnen fragt nach einer MTG-JSON-Datei und liest sie ein.
func (b *Bindungen) MTGJSONOeffnen(ctx context.Context, spiel string) OeffnenAntwort {
	texte := DialogtexteFuer(b.app(ctx).sprache)

	return b.waehlenUndOeffnen(ctx, Dateiauswahl{
		Titel:  texte.MTGJSONDatei,
		Filter: texte.MTGJSONFilter,
		Muster: "*.json",
	}, func(ctx context.Context, pfad string) OeffnenAntwort {
		return b.app(ctx).MTGJSONOeffnen(ctx, pfad, spiel)
	})
}

// MTGJSONVerzeichnisOeffnen fragt nach einem Verzeichnis einzelner Sets.
//
// **Ein eigener Weg, weil macOS zwei Dialoge kennt.** Ein Dateidialog lässt
// keinen Ordner wählen; wer nur ihn anbietet, kann die halbe Gestalt von MTG
// JSON nicht öffnen.
func (b *Bindungen) MTGJSONVerzeichnisOeffnen(ctx context.Context, spiel string) OeffnenAntwort {
	texte := DialogtexteFuer(b.app(ctx).sprache)

	return b.waehlenUndOeffnen(ctx, Dateiauswahl{
		Titel:       texte.MTGJSONVerzeichnis,
		Verzeichnis: true,
	}, func(ctx context.Context, pfad string) OeffnenAntwort {
		return b.app(ctx).MTGJSONOeffnen(ctx, pfad, spiel)
	})
}

// MTGJSONSpiele nennt die Spiele, für die es eine Zuordnung gibt.
func (b *Bindungen) MTGJSONSpiele(ctx context.Context) []string {
	return b.app(ctx).MTGJSONSpiele()
}

// PaketHerkunft sagt, wessen Inhalte ein Bezug holen würde.
func (b *Bindungen) PaketHerkunft(ctx context.Context, adresse, stand string) HerkunftAntwort {
	return b.app(ctx).PaketHerkunft(adresse, stand)
}

// PaketBeziehen holt ein Paket auf diesen Rechner.
//
// **Nur auf ausdrückliche Anweisung** (#56): die Oberfläche zeigt zuerst
// `PaketHerkunft` und ruft dies erst, wenn jemand zugestimmt hat. Ein Programm,
// das ungefragt über ein Gigabyte lädt, überrascht seine Nutzenden.
func (b *Bindungen) PaketBeziehen(ctx context.Context, adresse, stand, unterverzeichnis, name string) BezugAntwort {
	return b.app(ctx).PaketBeziehen(b.ruf(ctx), adresse, stand, unterverzeichnis, name)
}

// MTGJSONSchreiben fragt nach einem Ort und schreibt ein Set als MTG JSON.
//
// **Ein Speicherdialog, kein Öffnendialog.** Wer eine Datei anlegt, sucht keinen
// vorhandenen Namen aus, sondern gibt einen neuen — und macOS kennt dafür einen
// eigenen Dialog samt Warnung vor dem Überschreiben.
func (b *Bindungen) MTGJSONSchreiben(ctx context.Context, setID, spiel string) AusfuhrAntwort {
	ctx = b.ruf(ctx)
	texte := DialogtexteFuer(b.app(ctx).sprache)

	pfad, err := b.app(ctx).waehleDatei(ctx, Dateiauswahl{
		Titel:     texte.MTGJSONSchreiben,
		Filter:    texte.MTGJSONFilter,
		Muster:    "*.json",
		Vorschlag: texte.MTGJSONVorschlag,
		Speichern: true,
	})
	if err != nil {
		return AusfuhrAntwort{Befunde: []string{}, Fehler: Uebersetze(ctx, err)}
	}

	return b.app(ctx).MTGJSONSchreiben(ctx, setID, spiel, pfad)
}

// Exportvorlagen nennt die Vorlagen, die zum Spiel eines Sets gehören.
func (b *Bindungen) Exportvorlagen(ctx context.Context, setID string) VorlagenAntwort {
	return b.app(ctx).Exportvorlagen(b.ruf(ctx), setID)
}

// VorlageAusfuehren fragt nach einem Ziel und schreibt die Ausgabe dorthin.
//
// **Der Filter kommt aus der Vorlage** (`file type:`), und das ist der Punkt
// aus #209:
// eine Vorlage, die XML schreibt, soll nicht `*.txt` vorschlagen. MSE schreibt
// ihn als `*.xml|*.xml|*.*|*.*`; genommen wird das **erste** Muster.
func (b *Bindungen) VorlageAusfuehren(
	ctx context.Context,
	setID, paket, dateityp string, wahl map[string]string, auswahl []string,
) VorlagenausfuhrAntwort {
	ctx = b.ruf(ctx)
	texte := DialogtexteFuer(b.app(ctx).sprache)

	muster := erstesMuster(dateityp)

	pfad, err := b.app(ctx).waehleDatei(ctx, Dateiauswahl{
		Titel:     texte.VorlageSchreiben,
		Filter:    texte.VorlageFilter + " (" + muster + ")",
		Muster:    muster,
		Vorschlag: vorlagenvorschlag(dateityp),
		Speichern: true,
	})
	if err != nil {
		return VorlagenausfuhrAntwort{Befunde: []string{}, Fehler: Uebersetze(ctx, err)}
	}

	if pfad == "" {
		return b.app(ctx).VorlageAusfuehren(ctx, setID, paket, pfad, wahl, auswahl)
	}

	return imDateilauf(b.app(ctx), ctx, "vorlage", pfad, func(lauf context.Context) VorlagenausfuhrAntwort {
		return b.app(ctx).VorlageAusfuehren(lauf, setID, paket, pfad, wahl, auswahl)
	}, func(antwort VorlagenausfuhrAntwort) *Fehler { return antwort.Fehler })
}

// Befehlszeile richtet `mse` und `mse-mcp` in `~/.local/bin` ein oder nimmt
// sie heraus — OPS-03; leer liest sie nur.
func (b *Bindungen) Befehlszeile(ctx context.Context, handlung string) Befehlsvorschau {
	return b.app(ctx).Befehlszeile(b.ruf(ctx), handlung)
}

// Startbefund sagt, ob der Bestand offen ist — und wenn nicht, warum.
//
// **Den Kontext braucht sie nicht**, weil sie nichts fragt und nichts schreibt: sie liest, was
// beim Start dastand. Und weil sie gerade dann gerufen wird, wenn nichts
// aufging (#962).
func (b *Bindungen) Startbefund(_ context.Context) Startbefund {
	return b.frei().Startbefund()
}

// OrdnerAntwort trägt ein gewähltes Verzeichnis oder einen Fehler.
type OrdnerAntwort struct {
	Ordner string `json:"ordner"`

	// Abgebrochen sagt, dass niemand einen Ordner gewählt hat.
	Abgebrochen bool `json:"abgebrochen"`

	Fehler *Fehler `json:"fehler"`
}

// Bildordner fragt nach dem Verzeichnis, in das die Bilder kommen.
//
// **Ein Ordner, keine Datei.** Ein Stapellauf schreibt viele Dateien; die
// Namen bildet `Bildnamen`. Er beginnt beim gemerkten Ordner ([AblageBilder]).
func (b *Bindungen) Bildordner(ctx context.Context) OrdnerAntwort {
	texte := DialogtexteFuer(b.app(ctx).sprache)
	ctx = b.ruf(ctx)

	ordner, err := b.app(ctx).waehleDatei(ctx, Dateiauswahl{
		Titel: texte.Bildordner, Verzeichnis: true, Ablage: AblageBilder,
	})
	if err != nil {
		return OrdnerAntwort{Fehler: Uebersetze(ctx, err)}
	}

	if ordner == "" {
		return OrdnerAntwort{Abgebrochen: true}
	}

	return OrdnerAntwort{Ordner: ordner}
}

// Bildnamen gibt die Dateinamen, unter denen ein Set herauskäme.
func (b *Bindungen) Bildnamen(
	ctx context.Context,
	setID, vorlage, format, verhalten string, auswahl []string,
) BildnamenAntwort {
	return b.app(ctx).Bildnamen(b.ruf(ctx), setID, vorlage, format, verhalten, auswahl)
}

// BildSchreiben legt eine Karte als Bild in ein Verzeichnis.
//
// **Eine Karte je Aufruf, und der Stapel läuft in der Oberfläche.** Nur so
// schreitet er sichtbar fort und lässt sich abbrechen (#49): ein Lauf im
// Hintergrund müsste dem Fenster sagen können, wo er steht, und diese Naht gibt
// es nicht (#162).
func (b *Bindungen) BildSchreiben(
	ctx context.Context,
	karteID, verzeichnis, datei, format string, zielbreite int,
	verhalten, guete string,
) BildAntwort {
	return b.app(ctx).BildSchreiben(b.ruf(ctx), karteID, verzeichnis, datei,
		format, zielbreite, verhalten, guete)
}

// BogenLegen fragt nach einem Ort und schreibt einen Druckbogen dorthin.
//
// **`auswahl` leer heißt: das ganze Set** (#1167).
func (b *Bindungen) BogenLegen(
	ctx context.Context,
	setID, papier string, anschnitt float64, dpi int, marken bool,
	auswahl []string,
) BogenAntwort {
	ctx = b.ruf(ctx)
	texte := DialogtexteFuer(b.app(ctx).sprache)

	pfad, err := b.app(ctx).waehleDatei(ctx, Dateiauswahl{
		Titel:     texte.BogenSchreiben,
		Filter:    texte.PDFFilter,
		Muster:    "*.pdf",
		Vorschlag: texte.BogenVorschlag,
		Speichern: true,
	})
	if err != nil {
		return BogenAntwort{Befunde: []string{}, Fehler: Uebersetze(ctx, err)}
	}

	if pfad == "" {
		return BogenAntwort{Befunde: []string{}, Abgebrochen: true}
	}

	return imDateilauf(b.app(ctx), ctx, "bogen", pfad, func(lauf context.Context) BogenAntwort {
		return b.app(ctx).BogenLegen(lauf, setID, pfad, papier, anschnitt, dpi, marken, auswahl)
	}, func(antwort BogenAntwort) *Fehler { return antwort.Fehler })
}

// waehlenUndOeffnen fragt nach einem Ort und reicht ihn weiter.
func (b *Bindungen) waehlenUndOeffnen(
	ctx context.Context,
	auswahl Dateiauswahl, oeffnen func(context.Context, string) OeffnenAntwort,
) OeffnenAntwort {
	ctx = b.ruf(ctx)

	pfad, err := b.app(ctx).waehleDatei(ctx, auswahl)
	if err != nil {
		return OeffnenAntwort{
			Befunde: []string{}, Angelegt: []string{}, Fehler: Uebersetze(ctx, err),
		}
	}

	if pfad == "" {
		return oeffnen(ctx, pfad)
	}

	return imDateilauf(b.app(ctx), ctx, "mtgjson", pfad, func(lauf context.Context) OeffnenAntwort {
		return oeffnen(lauf, pfad)
	}, func(antwort OeffnenAntwort) *Fehler { return antwort.Fehler })
}

// Einstellungen gibt die Anwendungseinstellungen samt ihrer möglichen Werte.
func (b *Bindungen) Einstellungen(ctx context.Context) EinstellungenAntwort {
	return b.frei().Einstellungen(b.frei().rufkontext(ctx))
}

// EinstellungSetzen schreibt eine Einstellung fort.
func (b *Bindungen) EinstellungSetzen(ctx context.Context, schluessel, wert string) SchreibAntwort {
	return b.app(ctx).EinstellungSetzen(b.ruf(ctx), schluessel, wert)
}

// ExportoptionSetzen behält den Wert einer Option einer Exportvorlage (#492).
func (b *Bindungen) ExportoptionSetzen(ctx context.Context, vorlage, feld, wert string) SchreibAntwort {
	return b.app(ctx).ExportoptionSetzen(b.ruf(ctx), vorlage, feld, wert)
}
