package service

// **Die lokale Verbuchung eines Handels** — Werkstück `T6` von Phase 5.6
// (#1471).
//
// # Warum hier so viel geprüft wird, bevor irgendetwas geschieht
//
// **Weil danach nichts mehr zu berichtigen ist.** Ein verbuchter Handel
// bewegt echtes Papier; wer ihn zurücknehmen will, braucht eine
// Gegenbewegung, eine Erklärung und die Zustimmung des anderen. Die billigste
// Stelle, an der ein Fehler auffallen kann, ist diese hier.
//
// # Und warum die Unterschrift vor dem Bestand entsteht
//
// **Weil sie aus dem Schlüsselbund kommt und der zurückfragt.** Eine
// SQLite-Transaktion, die auf einen Dialog wartet, hält die ganze Datenbank —
// und ein Mensch, der den Dialog wegklickt, entscheidet damit über einen
// Bestand, den er nicht sieht. Der Bestand bekommt die fertige Unterschrift.

import (
	"context"
	"encoding/hex"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Besitzquelle ist, was die Verbuchung über eine Besitzzeile wissen muss.
type Besitzquelle interface {
	Einzeln(ctx context.Context, id int64) (*domain.Exemplar, error)
}

// Zusagenquelle gibt die Zusagen eines Vorgangs.
//
// **Je Vorgang und nicht je Exemplar**: die Verbuchung fragt „was habe ich für
// diesen Handel festgehalten", und eine Schleife über alle Besitzzeilen wäre
// dieselbe Frage von hinten.
type Zusagenquelle interface {
	FuerVorgang(ctx context.Context, vorgang string) ([]domain.Reservierung, error)
}

// Verbuchungsablage ist, was die Verbuchung schreibt und nachschlägt.
type Verbuchungsablage interface {
	// Abschliessen führt alles in **einer** Transaktion aus.
	Abschliessen(ctx context.Context, abschluss domain.Handelsabschluss) error

	// Beleg gibt den abgelegten Beleg — oder `nil`.
	Beleg(ctx context.Context, vorgang string) (*domain.Belegstand, error)

	// Gegenzeichnen legt den **geprüften** Beleg der Gegenseite ab — und
	// setzt damit `beidseitig`
	// (#1650).
	//
	// **Gerufen wird es von der Post und nicht von hier.** Wer
	// gegenzeichnete, weil jemand am Fenster für beide bestätigt hat,
	// machte aus einer Behauptung eine Tatsache.
	Gegenzeichnen(ctx context.Context, fremd domain.Fremdbeleg) error

	// Journal gibt die Bewegungen zu einem Vorgang.
	Journal(ctx context.Context, vorgang string) ([]domain.Journalzeile, error)
}

// Verbuchungsbestand fasst zusammen, was die Stelle vom Bestand braucht.
type Verbuchungsbestand interface {
	Besitzquelle
	Zusagenquelle
	Verbuchungsablage
}

// Verbuchungsstelle verbucht Handel lokal und schließt sie ab.
type Verbuchungsstelle struct {
	stand   Handelsablage
	bestand Verbuchungsbestand
	ausweis Ausweisstelle
	drucke  Druckbezeichnungen
	wache   Verbuchungswache
}

// Beleg gibt den abgelegten Beleg eines Vorgangs — oder `nil`.
func (v *Verbuchungsstelle) Beleg(
	ctx context.Context, vorgang string,
) (*domain.Belegstand, error) {
	return v.bestand.Beleg(ctx, vorgang)
}

// Verbuchen bucht einen übergebenen Handel im eigenen Bestand.
//
// # Warum ein zweiter Ruf nichts bewegt
//
// **Weil der Beleg vor dem Schritt nachgeschlagen wird.** Ein Neustart, ein
// wiederhergestelltes Backup und eine zweite Nachricht führen alle hierher;
// steht der Beleg da, ist verbucht, und der Stand kommt zurück, ohne dass
// noch einmal gebucht wird. Das Issue verlangt es wörtlich: „Neustart/Restore
// erzeugt niemals zwei Bestandsbewegungen."
// # Und warum hier ein Nachweis verlangt wird
//
// **Weil danach echtes Papier den Besitzer gewechselt hat** (T10, #1475).
// Eine Rolle sagt, dass jemand grundsätzlich handeln darf; ein Nachweis sagt,
// dass ein Mensch **diesen** Handel in **diesem** Stand gewollt hat. Ein Feld
// `bestaetigt`, das der Aufrufer selbst setzt, wäre keiner.
//
// **Damit ist MCP und REST dieser Schritt verschlossen**, und zwar nicht
// durch eine Aufzählung, sondern durch die Reichweite eines Tokens: sie endet
// bei [domain.WirkungSchreiben], und dies hier ist [domain.WirkungVerbindlich].
func (v *Verbuchungsstelle) Verbuchen(
	ctx context.Context, vorgang, nachweis string,
) (domain.Handelsangebot, error) {
	// **Der Vorgang zuerst und der Ausweis danach.** Der Ausweis kommt aus
	// dem Schlüsselbund, und der fragt zurück — für einen Handel, den es
	// nicht gibt, klickte sonst jemand einen Dialog weg.
	angebot, fehler := v.stand.Vorgang(ctx, vorgang)
	if fehler != nil {
		return domain.Handelsangebot{}, fehler
	}

	// **Nur ein Tausch bewegt Bestand** (TRADE-03.AK-02, #1717).
	if !angebot.Art.BewegtBestand() {
		return domain.Handelsangebot{}, domain.MitGrund(domain.ErrValidation, nil,
			"Ein Handel der Art %q bewegt keinen Bestand und wird nicht verbucht.", angebot.Art)
	}

	beleg, fehler := v.bestand.Beleg(ctx, vorgang)
	if fehler != nil {
		return domain.Handelsangebot{}, fehler
	}

	// **Der zweite Ruf verbraucht keinen Nachweis.** Er bucht nicht, und ein
	// Mensch, dessen Freigabe an einer Wiederholung verfiele, müsste für
	// nichts ein zweites Mal zustimmen.
	if beleg != nil {
		return angebot, nil
	}

	if verwehrt := v.darf(ctx, domain.WirkungVerbindlich, nachweis,
		HandelsverbuchungZweck, vorgang, angebot.Revision); verwehrt != nil {
		return domain.Handelsangebot{}, verwehrt
	}

	ich, fehler := v.ausweis.Ausweis(ctx)
	if fehler != nil {
		return domain.Handelsangebot{}, fehler
	}

	verbucht, fehler := angebot.Schreite(
		domain.SchrittVerbuchen, string(ich.Kennung()))
	if fehler != nil {
		return domain.Handelsangebot{}, fehler
	}

	zeilen, fehler := v.zeilenFuer(ctx, verbucht, string(ich.Kennung()))
	if fehler != nil {
		return domain.Handelsangebot{}, fehler
	}

	abschluss := domain.Handelsabschluss{
		Vorher: angebot, Angebot: verbucht,
		Beleg: domain.Handelsbeleg{
			Vorgang: vorgang, Abdruck: verbucht.Abdruck(),
			Parteien: verbucht.Parteien, Zeilen: zeilen,
		},
	}

	// **Hier und nicht im Bestand** — der Grund steht im Kopf.
	abschluss.Unterschrift = hex.EncodeToString(
		ich.Signiere(abschluss.Beleg.Unterschriftsform()))

	if fehler := v.bestand.Abschliessen(ctx, abschluss); fehler != nil {
		return domain.Handelsangebot{}, fehler
	}

	return verbucht, nil
}

// Abgleichen meldet, dass die Gegenseite noch fehlt.
//
// **Es bucht nichts und gibt nichts frei.** Es sagt nur, dass hier fertig ist
// und dort noch nicht — und genau das ist die Auskunft, die ein Mensch
// braucht, um nachzufragen.
// **Und deshalb braucht es dafür keinen Nachweis** (T10, #1475): gefragt wird
// nach [domain.WirkungSchreiben], und das ist, was ein Mitarbeiter ohnehin
// darf. Einen Menschen für eine Meldung zu fragen, die nichts bindet, wäre
// ein Dialog ohne Entscheidung — und wer oft genug ohne Grund gefragt wird,
// klickt irgendwann alles weg.
func (v *Verbuchungsstelle) Abgleichen(
	ctx context.Context, vorgang string,
) (domain.Handelsangebot, error) {
	if fehler := v.darf(ctx, domain.WirkungSchreiben, "",
		HandelsverbuchungZweck, vorgang, 0); fehler != nil {
		return domain.Handelsangebot{}, fehler
	}

	return v.schreiten(ctx, vorgang, domain.SchrittAbgleichen, "")
}

// Bestaetigen nimmt eine Stimme für den beidseitigen Abschluss entgegen.
//
// # Warum erst die zweite gegenzeichnet
//
// **Weil zwei Unterschriften zwei sind.** Der Übergang verlangt **alle**
// Parteien; solange eine fehlt, bleibt der Vorgang im Abgleich, und der Beleg
// bleibt einseitig. Wer hier schon gegenzeichnete, täuschte eine zweite
// Unterschrift vor, die es nicht gibt — das Issue verbietet es wörtlich.
//
// **`wer` ist die Partei, deren Stimme ankommt** — die eigene oder die des
// Gegenübers. Welche es sein darf, entscheidet der Vertrag und nicht diese
// Stelle.
// # Und warum auch das einen Nachweis braucht
//
// **Weil `wer` frei gewählt ist** (T10, #1475). Wer diesen Ruf ohne Riegel
// hätte, könnte die Stimme des **Gegenübers** einsetzen und aus einem
// einseitigen Beleg einen beidseitigen machen — eine zweite Unterschrift,
// die es nicht gibt. Das Issue verbietet genau das: ein Modell bestätigt
// keine physische Übergabe autonom.
func (v *Verbuchungsstelle) Bestaetigen(
	ctx context.Context, vorgang, wer, nachweis string,
) (domain.Handelsangebot, error) {
	angebot, fehler := v.stand.Vorgang(ctx, vorgang)
	if fehler != nil {
		return domain.Handelsangebot{}, fehler
	}

	// **Der Nachweis hängt an der Partei, deren Stimme gesetzt wird**
	// (Befund 5 des Zweitblicks über Phase 5.6). Einer, der nur den Vorgang
	// nennte, gälte für **jede** Stimme — und ein Mensch, der einmal
	// freigibt, setzte damit auch die Unterschrift des Gegenübers.
	if verwehrt := v.darf(ctx, domain.WirkungVerbindlich, nachweis,
		HandelsabschlussZweck, Handelsfreigabeziel(vorgang, wer),
		angebot.Revision); verwehrt != nil {
		return domain.Handelsangebot{}, verwehrt
	}

	gestimmt, fehler := v.schreiten(ctx, vorgang, domain.SchrittBestaetigen, wer)
	if fehler != nil {
		return domain.Handelsangebot{}, fehler
	}

	// **Und hier wird nicht gegengezeichnet**
	// (#1650).
	//
	// **Bis zum 2026-09-19 stand hier `Gegenzeichnen`**, sobald der Zustand
	// `beidseitigVerbucht` erreicht war — und der Weg dorthin ist, dass ein
	// Mensch an **diesem** Fenster die Stimme des Gegenübers einsetzt. Damit
	// sagte `beidseitig` „auch die Gegenseite hat unterschrieben", wo jemand
	// bloß behauptet hatte, sie habe es. Das schließt der eigene Kommentar
	// an [domain.Belegstand.Beidseitig] wörtlich aus.
	//
	// **Gegengezeichnet wird jetzt, wenn ein geprüfter Beleg ankommt** —
	// [Handelspost] tut es, und nur sie.
	return gestimmt, nil
}

// schreiten geht einen Schritt, der den Bestand nicht bewegt.
func (v *Verbuchungsstelle) schreiten(
	ctx context.Context, vorgang string, schritt domain.Handelsschritt,
	wer string,
) (domain.Handelsangebot, error) {
	angebot, fehler := v.stand.Vorgang(ctx, vorgang)
	if fehler != nil {
		return domain.Handelsangebot{}, fehler
	}

	// **Auch hier der Vorgang zuerst** — derselbe Grund wie bei
	// [Verbuchungsstelle.Verbuchen].
	if wer == "" {
		ich, ausweisfehler := v.ausweis.Ausweis(ctx)
		if ausweisfehler != nil {
			return domain.Handelsangebot{}, ausweisfehler
		}

		wer = string(ich.Kennung())
	}

	geschritten, fehler := angebot.Schreite(schritt, wer)
	if fehler != nil {
		return domain.Handelsangebot{}, fehler
	}

	// **Der Bestand bewegt sich hier nicht — der Vorgang schon.** Ein
	// Schritt, der nur zurückgegeben und nicht abgelegt wird, ist nach dem
	// nächsten Start nicht geschehen.
	if fehler := v.stand.Fortschreiben(ctx, angebot, geschritten); fehler != nil {
		return domain.Handelsangebot{}, fehler
	}

	return geschritten, nil
}

// zeilenFuer macht aus einem Angebot die Zeilen des Belegs.
//
// **Abgänge kommen aus den Zusagen, Zugänge aus dem Angebot.** Das ist kein
// Zufall: was ich weggebe, muss ich festgehalten haben — was ich bekomme,
// habe ich noch nicht, und es kann keine Besitzzeile nennen.
func (v *Verbuchungsstelle) zeilenFuer(
	ctx context.Context, angebot domain.Handelsangebot, ich string,
) ([]domain.Journalzeile, error) {
	abgaenge, fehler := v.abgaengeFuer(ctx, angebot, ich)
	if fehler != nil {
		return nil, fehler
	}

	zeilen := abgaenge

	for _, stelle := range angebot.Positionen {
		if stelle.Von == ich {
			continue
		}

		// **Dieselbe Zeile, die der Schritt zur Einigung schon geprüft hat**
		// ([domain.Handelsangebot.PruefeVerbuchbar], #2495).
		zeile := stelle.Zugangszeile(angebot.Vorgang)

		// **Was die Gegenseite schickt, wird hier geprüft und nicht in der
		// Transaktion** (§8: fremde Daten sind der Normalfall). Wer es
		// durchreichte, bekäme die Beanstandung aus der Tiefe eines offenen
		// Zuges — mit einem Satz über eine Spalte statt über den Handel, und
		// nachdem der Abgang schon gerechnet war.
		if befunde := zeile.AlsExemplar().Validate(); len(befunde) > 0 {
			return nil, befunde[0]
		}

		zeilen = append(zeilen, zeile)
	}

	return zeilen, nil
}

// abgaengeFuer macht aus den bestätigten Zusagen die Abgangszeilen.
func (v *Verbuchungsstelle) abgaengeFuer(
	ctx context.Context, angebot domain.Handelsangebot, ich string,
) ([]domain.Journalzeile, error) {
	zusagen, fehler := v.bestand.FuerVorgang(ctx, angebot.Vorgang)
	if fehler != nil {
		return nil, fehler
	}

	// **Was ich weggebe, steht im Angebot** — die Zusagen sagen, woher.
	offen := 0

	for _, stelle := range angebot.Positionen {
		if stelle.Von == ich {
			offen += stelle.Menge
		}
	}

	zeilen := []domain.Journalzeile{}

	for _, zusage := range zusagen {
		if !zusage.Uebergeben {
			continue
		}

		zeile, fehler := v.zeileAus(ctx, angebot.Vorgang, zusage)
		if fehler != nil {
			return nil, fehler
		}

		zeilen = append(zeilen, zeile)
		offen -= zusage.Menge
	}

	// **Ohne vollständige Deckung wird gar nicht gebucht.** Eine halbe
	// Übergabe ist keine; was fehlt, klärt ein Mensch.
	if offen > 0 {
		return nil, domain.MitGrund(domain.ErrConflict, nil,
			"es fehlen %d Stück an bestätigter Übergabe — der Handel wird "+
				"nicht halb verbucht", offen)
	}

	return zeilen, nil
}

// zeileAus macht aus einer Zusage die Abgangszeile ihrer Besitzzeile.
func (v *Verbuchungsstelle) zeileAus(
	ctx context.Context, vorgang string, zusage domain.Reservierung,
) (domain.Journalzeile, error) {
	stueck, fehler := v.bestand.Einzeln(ctx, zusage.ExemplarID)
	if fehler != nil {
		return domain.Journalzeile{}, fehler
	}

	// **Die Fassung wird hier schon geprüft und im Bestand noch einmal**
	// (§7, Regel 2). Hier, damit ein Mensch eine Auskunft bekommt, statt
	// einen Transaktionsabbruch zu lesen; dort, weil zwischen beiden
	// Prüfungen ein zweites Fenster liegen kann.
	if stueck.Fassung != zusage.Fassung {
		return domain.Journalzeile{}, domain.MitGrund(domain.ErrConflict, nil,
			"die Besitzzeile %d steht in Fassung %d, zugesagt wurde gegen %d",
			stueck.ID, stueck.Fassung, zusage.Fassung)
	}

	return domain.Journalzeile{
		Art: domain.JournalAbgang, Anlass: domain.AnlassTausch,
		Vorgang: vorgang, Printing: stueck.Printing, Menge: zusage.Menge,
		Zustand: stueck.Zustand, Sprache: stueck.Sprache,
		Finish: stueck.Ausfuehrung(), ExemplarID: stueck.ID,
		Lagerort: stueck.Lagerort, Notiz: stueck.Notiz,
	}, nil
}
