package service

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Einstellungsablage ist die Naht zum Bestand — implementiert in `storage`.
type Einstellungsablage interface {
	Alle(ctx context.Context) (map[string]string, error)
	Setzen(ctx context.Context, schluessel, wert string) error
}

// Nutzerwerteleser gibt die Einstellungen des Menschen, dem das gezeigte
// Profil gehört — Werkstück 9 von Phase 4.12
// (#1351),
// Punkt 9.
//
// **Ohne Kennung im Aufruf.** Welches Profil gezeigt wird, steht im Kontext
// (`storage.ProfilAus`), und wem es gehört, weiß der Bestand. Eine Kennung im
// Aufruf hieße, dass der Rufer sie erst holen muss — und dann hätte er zwei
// Stände (§7, Regel 2).
type Nutzerwerteleser interface {
	NutzerwerteZumProfil(ctx context.Context) (map[string]string, error)
}

// Einstellungsverwalter liest und schreibt die Anwendungseinstellungen (#245).
type Einstellungsverwalter struct {
	ablage Einstellungsablage

	// nutzerwerte ist die Naht zum Menschen — `nil` heißt: niemand erbt.
	//
	// **Ein `nil` ist kein Sonderfall im Ergebnis**: `erben` fällt dann auf
	// die Vorgabe des Menschen zurück, und die ist genau das, was vor
	// Werkstück 9 hier stand. Ein Bestand ohne diese Naht verhält sich also,
	// wie er sich immer verhalten hat.
	nutzerwerte Nutzerwerteleser

	// ordnerpruefung fragt die Platte nach einem Ordner — OPS-02 (#1719).
	// `nil` heißt: niemand fragt, und ein Ordner gilt, wie er dasteht.
	ordnerpruefung func(ctx context.Context, angabe string) error

	// verschluesselt sagt, ob der Bestand verschlüsselt ist.
	verschluesselt func(ctx context.Context) (bool, error)

	// riegel fragt, ob der Akteur bestimmen darf, wohin diese Anwendung
	// sendet — SEC-P4 (#2000), siehe [domain.IstNetzweg]. `nil` heißt:
	// niemand fragt; die Wurzeln setzen ihn über `einstellungsverwalterUeber`.
	riegel Pruefer
}

// NeuerEinstellungsverwalter gibt den Verwalter über dieser Ablage.
func NeuerEinstellungsverwalter(ablage Einstellungsablage) *Einstellungsverwalter {
	return &Einstellungsverwalter{ablage: ablage}
}

// MitNutzerwerten setzt die Naht zum Menschen.
func (v *Einstellungsverwalter) MitNutzerwerten(
	leser Nutzerwerteleser,
) *Einstellungsverwalter {
	v.nutzerwerte = leser

	return v
}

// MitOrdnerpruefung setzt die Frage an die Platte — OPS-02.AK-01 und AK-03
// (#1719).
//
// **Hier und nicht an jeder Fläche**: Fenster, Kommandozeile, REST und MCP
// setzen über diesen Verwalter, und ein Ordner, der an einer Fläche geprüft
// würde und an einer anderen nicht, wäre ein zweiter Pfadvertrag (AK-04).
func (v *Einstellungsverwalter) MitOrdnerpruefung(
	pruefung func(ctx context.Context, angabe string) error,
) *Einstellungsverwalter {
	v.ordnerpruefung = pruefung

	return v
}

// MitVerschluesselungsstand setzt die Frage, ob der Bestand verschlüsselt ist
// — für den Plan (Nachlauf über Phase 4, Befund 4).
func (v *Einstellungsverwalter) MitVerschluesselungsstand(stand func(ctx context.Context) (bool, error)) *Einstellungsverwalter {
	v.verschluesselt = stand

	return v
}

// MitRiegel setzt den Prüfer für die Netzwege — SEC-P4
// (#2000).
//
// **Hier und nicht an jeder Fläche**, aus demselben Grund wie die
// Ordnerprüfung darüber: Fenster, Kommandozeile, Dienst und MCP setzen über
// diesen Verwalter, und ein Riegel an einer Fläche ließe drei offen.
func (v *Einstellungsverwalter) MitRiegel(riegel Pruefer) *Einstellungsverwalter {
	v.riegel = riegel

	return v
}

// Wahl ist ein Schlüsselwert in zwei Gestalten — Werkstück 9 (#1351).
//
// **Beide, und der Unterschied ist die ganze Zusage**: die Auswahlliste zeigt
// `Gewaehlt` (dort steht vielleicht „wie der Nutzer"), die Anwendung richtet
// sich nach `Wirksam`. Wer nur eines von beiden hätte, zeigte entweder eine
// Wahl, die niemand getroffen hat, oder eine Anwendung, die anders aussieht,
// als die Liste sagt.
type Wahl struct {
	Gewaehlt string
	Wirksam  string

	// Herkunft sagt, wessen Entscheidung der wirksame Wert war — `profil`,
	// `nutzer` oder `vorgabe`. Werkstück O3 von Phase 5.2
	// (#1427).
	//
	// **Die dritte Gestalt, weil zwei zu wenig waren**: „dunkel" allein sagt
	// nicht, ob es hier gewählt wurde, vom Menschen kommt oder die Vorgabe
	// ist — und damit nicht, was ein Zurücksetzen bewirkte.
	Herkunft string
}

// Stand gibt je Schlüssel den gewählten und den wirksamen Wert — und seit
// Werkstück O3 von Phase 5.2
// (#1427)
// die **Herkunft** des wirksamen.
//
// **Je ein Lesen und nicht zwei.** Wer die Wahl und das Geltende nacheinander
// holte, läse dieselbe Tabelle zweimal — und bekäme zwei Stände, wenn
// dazwischen jemand schreibt (§7, Regel 2).
func (v *Einstellungsverwalter) Stand(ctx context.Context) (map[string]Wahl, error) {
	gesetzt, err := v.ablage.Alle(ctx)
	if err != nil {
		return nil, err
	}

	vomMenschen, lesefehler := v.vomMenschen(ctx)
	if lesefehler != nil {
		return nil, lesefehler
	}

	// **Die Auflösung steht in `domain`** — Werkstück O3 von Phase 5.2
	// (#1427).
	// Seit es **Richtlinien** gibt, ist sie keine Auswahl mehr, sondern eine
	// Faltung, und die Vorgabe eines Profils darf dabei nicht mitreden.
	//
	// **Die rohen Werte gehen hinein und nicht die aufgefüllten**: eine
	// Vorgabe ist keine Aussage, und als Aussage genommen erstickte sie die
	// Wahl eines Menschen.
	geltung := domain.WirksamerStand(gesetzt, vomMenschen)

	stand := make(map[string]Wahl, len(geltung))

	for schluessel, eine := range geltung {
		stand[schluessel] = Wahl{
			Gewaehlt: eine.Gewaehlt,
			Wirksam:  eine.Wirksam,
			Herkunft: eine.Herkunft,
		}
	}

	return stand, nil
}

// vomMenschen liest die Einstellungen des Menschen — leer ohne Naht.
func (v *Einstellungsverwalter) vomMenschen(ctx context.Context) (
	map[string]string, error,
) {
	if v.nutzerwerte == nil {
		return map[string]string{}, nil
	}

	return v.nutzerwerte.NutzerwerteZumProfil(ctx)
}

// Alle gibt, was **gilt** — nicht, was gewählt ist.
//
// **Seit Werkstück 9 ist das ein Unterschied** (#1351, Punkt 9): steht bei
// einem Profil „wie der Nutzer", gilt der Wert des Menschen. Jeder Aufrufer
// dieser Antwort will wissen, was gilt — die Sprache für die Menüleiste, das
// Erscheinungsbild für die Wurzel, `bildkonflikt` für einen Ausgabelauf.
// **Wer die Wahl selbst braucht, ruft [Einstellungsverwalter.Stand]** — sie
// trägt beides, und das ist genau eine Stelle: die Einstellungsseite.
func (v *Einstellungsverwalter) Alle(ctx context.Context) (map[string]string, error) {
	stand, err := v.Stand(ctx)
	if err != nil {
		return nil, err
	}

	werte := make(map[string]string, len(stand))

	for schluessel, wahl := range stand {
		// **Ein leerer wirksamer Wert kommt nicht in die Abbildung** — aus
		// demselben Grund wie unten bei der Vorgabe: die Sprache hat keine,
		// und eine leere Zeichenkette sähe aus wie eine Wahl.
		if wahl.Wirksam != "" {
			werte[schluessel] = wahl.Wirksam
		}
	}

	return werte, nil
}

// Setzen schreibt eine Einstellung fort — geprüft.
//
// **Wohin gesendet wird, setzt nur, wer senden darf** (#2000): ein Netzweg
// fragt den Riegel nach [domain.WirkungExternSenden], und **das Nein kommt
// vor jeder anderen Antwort** (0078) — auch vor der Prüfung des Werts.
func (v *Einstellungsverwalter) Setzen(ctx context.Context, schluessel, wert string) error {
	if v.riegel != nil && domain.IstNetzweg(schluessel) {
		if err := v.riegel.Pruefe(ctx, domain.WirkungExternSenden); err != nil {
			return err
		}
	}

	if err := domain.PruefeEinstellung(schluessel, wert); err != nil {
		return err
	}

	// **Kein Plan über einem verschlüsselten Bestand** (Nachlauf über Phase
	// 4, Befund 4) — hier und nicht am Ordnerdialog allein, denn jede
	// Fläche setzt über diesen Verwalter. Beenden geht immer.
	if schluessel == domain.SicherungsplanOrdner && wert != "" && v.verschluesselt != nil {
		if verschluesselt, _ := v.verschluesselt(ctx); verschluesselt { //nolint:errcheck // ohne Antwort gilt Klartext, wie beim Lauf
			return planVerschluesselt()
		}
	}

	// **Erst die Gestalt, dann die Platte** — eine Angabe, die die Domäne
	// abweist, braucht keinen Systemruf. Und leer heißt: keine Vorgabe.
	if v.ordnerpruefung != nil && wert != "" && domain.IstOrdnerangabe(schluessel) {
		if err := v.ordnerpruefung(ctx, wert); err != nil {
			return err
		}
	}

	return v.ablage.Setzen(ctx, schluessel, wert)
}
