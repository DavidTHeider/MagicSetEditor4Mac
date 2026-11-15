package app

// **In der Nähe gefunden werden** — Werkstück N2 von Phase 5.4
// (#1455).
//
// # Warum das ein zweiter Schritt ist und nicht die Folge des ersten
//
// **Weil es etwas anderes preisgibt.** Wer horcht, ist erreichbar für den, der
// seine Adresse kennt. Wer sich ankündigt, sagt **jedem im Netz**, dass es ihn
// gibt — samt seiner Peerkennung, also seinem öffentlichen Schlüssel. Das
// Issue verlangt genau diese Trennung: „Nur nach bewusster Auswahl … keine
// unbemerkte Dauerwerbung aller Profile."
//
// # Und warum es das Horchen voraussetzt
//
// **Weil eine Ankündigung einen Port nennt.** Ohne Horchposten wäre sie eine
// Einladung an eine Tür, hinter der niemand steht.

import (
	"context"
	"fmt"
	"os"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/entdeckung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// SichtbarWerden kündigt dieses Profil im selben Netz an.
func (a *App) SichtbarWerden(ctx context.Context, name string) Horchantwort {
	ctx = a.imProfil(ctx)

	stand, fehler := a.sichtbarWerden(ctx, name)
	if fehler != nil {
		return Horchantwort{Stand: ohneListen(), Fehler: Uebersetze(ctx, fehler)}
	}

	return Horchantwort{Stand: stand}
}

// SichtbarkeitBeenden nimmt die Ankündigung zurück.
//
// **Das Horchen läuft weiter.** Wer nicht mehr gefunden werden will, ist
// deshalb nicht unerreichbar — wer seine Adresse hat, kommt weiter durch.
func (a *App) SichtbarkeitBeenden(ctx context.Context) Horchantwort {
	ctx = a.imProfil(ctx)

	_ = ctx

	a.sichtbarkeitBeenden()

	return Horchantwort{Stand: a.horchstand()}
}

// Nachbarn sagt, wer sich im selben Netz meldet — **rein lesend**.
func (a *App) Nachbarn(ctx context.Context) NachbarnAntwort {
	ctx = a.imProfil(ctx)

	// **Die eigene Kennung reist im Kontext mit**, damit der eigene
	// Widerhall nicht in der Liste steht. Ein Profil ohne Identität hat
	// keine — und findet sich folgerichtig auch nicht selbst.
	if ausweis, fehler := a.peerwerk().Ausweis(ctx); fehler == nil {
		ctx = service.MitEigenerKennung(ctx, ausweis.Kennung())
	}

	nachbarn, fehler := a.nachbarwerk().Liste(ctx)
	if fehler != nil {
		return NachbarnAntwort{
			Nachbarn: []Nachbar{}, Fehler: Uebersetze(ctx, fehler),
		}
	}

	liste := make([]Nachbar, 0, len(nachbarn.Nachbarn))
	for _, nachbar := range nachbarn.Nachbarn {
		liste = append(liste, alsNachbar(nachbar))
	}

	return NachbarnAntwort{Nachbarn: liste, Verworfen: nachbarn.Verworfen}
}

// sichtbarWerden baut die Ankündigung über dem laufenden Horchposten.
func (a *App) sichtbarWerden(
	ctx context.Context, name string,
) (Horchzustand, error) {
	// **Die Marke vor dem Start** (Nachlauf über Phase 6.12, N2).
	marke := a.laufmarke(domain.LaufSichtbarkeit)

	// **Die Rolle wird gefragt, bevor der Rechner erreichbar wird** —
	// Werkstück N7 von Phase 5.4
	// (#1568).
	//
	// **Bis dahin fragte hier niemand.** `docs/AUTORISIERUNG.md` nannte für
	// diese Zeile „mitarbeiter" — und kein Ruf prüfte es; ein Leser in einem
	// fremden Profil konnte einen Port öffnen. Das war keine Lücke, die N7
	// geschaffen hat, sondern eine, die es **benannt** hat: eine Tabelle,
	// die eine Mindestrolle nennt, die nichts durchsetzt, verspricht mehr,
	// als ihr Rumpf hält (§10).
	if fehler := a.darf(ctx, domain.WirkungExternEmpfangen); fehler != nil {
		return Horchzustand{}, fehler
	}

	ich, fehler := a.peerwerk().Ausweis(ctx)
	if fehler != nil {
		return Horchzustand{}, fehler
	}

	// **Der Platz wird belegt, bevor er gebraucht wird** — Befund 1 des
	// Nachlaufs zum Zweitblick über Phase 5.4, und dieselbe Gestalt wie bei
	// den vier Posten: prüfen, loslassen, öffnen, eintragen. Zwei Rufe
	// öffneten zwei Ankündigungen, und die zweite überschriebe den einzigen
	// Griff der ersten (§7, Regel 4).
	a.peerlage.horchen.schloss.Lock()
	adresse, laeuftSchon := a.peerlage.horchen.adresse,
		a.peerlage.horchen.anmeldung != nil || a.peerlage.horchen.meldetAn
	stand := a.peerlage.horchen.sichtbarkeit

	if adresse != "" && !laeuftSchon {
		a.peerlage.horchen.meldetAn = true
	}

	a.peerlage.horchen.schloss.Unlock()

	if adresse == "" {
		return Horchzustand{}, fmt.Errorf(
			"%w: dieses Profil horcht nicht — eine Ankündigung ohne Port "+
				"wäre eine Einladung an eine Tür, hinter der niemand steht",
			domain.ErrConflict)
	}

	if laeuftSchon {
		return Horchzustand{}, fmt.Errorf(
			"%w: dieses Profil kündigt sich schon an", domain.ErrConflict)
	}

	// **Und er wird auf jedem Weg wieder frei.** Ein Fehlschlag, der die
	// Belegung stehen ließe, machte aus einem Befund eine Sperre bis zum
	// nächsten Neustart.
	defer func() {
		a.peerlage.horchen.schloss.Lock()
		a.peerlage.horchen.meldetAn = false
		a.peerlage.horchen.schloss.Unlock()
	}()

	port, fehler := portAus(adresse)
	if fehler != nil {
		return Horchzustand{}, fehler
	}

	anmeldung, fehler := a.ankuendiger()(entdeckung.Angebot{
		Kennung: ich.Kennung(), Name: name, Port: port,
		Schnittstelle: a.entdeckungsschnittstelle(ctx),
	})
	if fehler != nil {
		return Horchzustand{}, fehler
	}

	// **Wer währenddessen aufgehört hat, hat aufgehört.** `mdns.Register`
	// dauert; zwischen dem Ruf und dieser Zeile kann ein Mensch „nicht mehr
	// sichtbar" gedrückt oder den Horchposten beendet haben. Eine Anmeldung,
	// die danach einträfe, strahlte einen Port aus, den niemand mehr hält —
	// und niemand sähe, warum.
	a.peerlage.horchen.schloss.Lock()
	nochGewollt := a.peerlage.horchen.sichtbarkeit == stand
	if nochGewollt {
		a.peerlage.horchen.anmeldung = anmeldung
	}
	a.peerlage.horchen.schloss.Unlock()

	if !nochGewollt {
		if fehler := anmeldung.Beende(); fehler != nil {
			meldeHorchbefund(fehler)
		}

		return Horchzustand{}, fmt.Errorf(
			"%w: die Ankündigung wurde zurückgenommen, während sie entstand",
			domain.ErrConflict)
	}

	a.nachDemStart(ctx, domain.LaufSichtbarkeit, marke, func(context.Context) { a.sichtbarkeitBeenden() })

	return a.horchstand(), nil
}

// sichtbarkeitBeenden nimmt eine laufende Ankündigung zurück.
func (a *App) sichtbarkeitBeenden() {
	a.peerlage.horchen.schloss.Lock()
	anmeldung := a.peerlage.horchen.anmeldung
	a.peerlage.horchen.anmeldung = nil

	// **Der Zähler steigt auch, wenn gar nichts lief.** Genau dann zählt
	// er: eine Ankündigung, die gerade entsteht, soll danach nicht mehr
	// eingetragen werden.
	a.peerlage.horchen.sichtbarkeit++

	a.peerlage.horchen.schloss.Unlock()

	if anmeldung == nil {
		return
	}

	if fehler := anmeldung.Beende(); fehler != nil {
		// **Gemeldet und nicht gestoppt** — dieselbe Gestalt wie bei
		// [App.sitzungenPausieren]: wer aufhören will, hört auf.
		meldeHorchbefund(fehler)
	}
}

// portAus liest den Port aus einer Adresse, die schon gebunden ist.
//
// **Sie kommt vom Lauscher und nicht von einem Menschen** — deshalb ist ein
// Fehler hier ein Befund über dieses Programm und keine Eingabeprüfung.
//
// **Über `domain` und nicht über `net.SplitHostPort`**: `internal/app` darf
// das Netz nicht erreichen (§2), und die Zerlegung steht in der Domäne, wo
// auch die Form geprüft wird.
func portAus(adresse string) (int, error) {
	_, port, fehler := domain.ZerlegeAdresse(adresse)
	if fehler != nil {
		return 0, fehler
	}

	return port, nil
}

// alsNachbar macht aus der Entität ihr DTO.
func alsNachbar(nachbar service.Nachbar) Nachbar {
	return Nachbar{
		Kennung:       string(nachbar.Kennung),
		Fingerabdruck: nachbar.Fingerabdruck,
		Name:          nachbar.Name,
		Adresse:       nachbar.Adresse,
		Merkadresse:   nachbar.Merkadresse,
		Bekannt:       nachbar.Bekannt,
		Stand:         string(nachbar.Stand),

		GemerkteAdresse:  nachbar.GemerkteAdresse,
		AdresseGeaendert: nachbar.AdresseGeaendert(),
	}
}

// nachbarnaht hängt `internal/entdeckung` an [service.Nachbarschaft].
type nachbarnaht struct{ app *App }

func (n nachbarnaht) Suche(
	ctx context.Context,
) (service.Nachbarfunde, error) {
	ergebnis, fehler := n.app.nachbarsucher()(ctx, entdeckung.Frage{
		Schnittstelle: n.app.entdeckungsschnittstelle(ctx),
	})
	if fehler != nil {
		return service.Nachbarfunde{}, fehler
	}

	funde := make([]service.Nachbarfund, 0, len(ergebnis.Funde))
	for _, fund := range ergebnis.Funde {
		funde = append(funde, service.Nachbarfund{
			Kennung: fund.Kennung, Name: fund.Name, Adresse: fund.Adresse(),
			Merkadresse: fund.Merkadresse(),
		})
	}

	return service.Nachbarfunde{
		Funde: funde, Verworfen: ergebnis.Verworfen,
	}, nil
}

// nachbarwerk baut das Werk für diesen Ruf.
func (a *App) nachbarwerk() *service.Nachbarwerk {
	return service.NeuesNachbarwerk(nachbarnaht{app: a}, a.kontaktwerk())
}

// Entdeckungsschnittstelle nennt die Umgebungsvariable für die
// Netzwerkschnittstelle der Entdeckung.
//
// # Wofür sie da ist
//
// **Für Läufe und für Rechner mit mehreren Netzen.** mDNS ist Multicast: ohne
// Angabe fragt es über die Schnittstelle, die das Betriebssystem wählt — und
// ein Lauf fragte damit das Netz, in dem dieser Rechner gerade steht (§3).
// Mit `lo0` bleibt er auf der Rückschleife.
//
// **Dieselbe Form wie `MSE_BESTAND` und `MSE_SCHLUESSELBUND`**: leer heißt,
// was ohne sie gälte.
const Entdeckungsschnittstelle = "MSE_ENTDECKUNGSSCHNITTSTELLE"

// entdeckungsschnittstelle liest die eingestellte Schnittstelle.
//
// **Die Einstellung zuerst, die Umgebung danach** (NET-02.AK-03, #1706):
// `netz.schnittstelle` ist die Wahl eines Menschen und gilt an jeder Fläche;
// die Umgebungsvariable bleibt für Läufe und für Rechner, an denen niemand
// etwas eingestellt hat. **Leer ist die Automatik** — jede taugliche.
//
// **Ein unbekannter Name ist kein Fehler, sondern keine Angabe.** Wer ihn
// falsch schreibt, bekommt die Vorgabe — und ein Programm, das deswegen nicht
// suchte, wäre die schlechtere Antwort auf einen Tippfehler.
func (a *App) entdeckungsschnittstelle(ctx context.Context) *netzschnittstelle {
	name := os.Getenv(Entdeckungsschnittstelle)

	// **Ein Lesefehler hält nichts auf** — dieselbe Linie wie bei den
	// Netzvorgaben: ohne die Einstellung gilt, was ohne sie gälte.
	if werte, fehler := a.einstellungswerte(ctx); fehler == nil && werte[domain.NetzSchnittstelle] != "" {
		name = werte[domain.NetzSchnittstelle]
	}

	if name == "" {
		return nil
	}

	schnitt, fehler := entdeckung.SchnittstelleNamens(name)
	if fehler != nil {
		meldeHorchbefund(fehler)

		return nil
	}

	return schnitt
}

// ankuendiger und nachbarsucher geben die eingesetzten Nähte — oder die
// echten.
//
// **`nachbarsucher` und nicht `sucher`**: den Namen trägt schon ein Feld
// dieser Anwendung, und zwei Dinge unter einem Namen sind eines zu viel.
func (a *App) ankuendiger() func(entdeckung.Angebot) (Ankuendigung, error) {
	if a.peerlage.horchen.ankuendigen != nil {
		return a.peerlage.horchen.ankuendigen
	}

	return echteAnkuendigung
}

func (a *App) nachbarsucher() func(
	context.Context, entdeckung.Frage,
) (entdeckung.Ergebnis, error) {
	if a.peerlage.horchen.suchen != nil {
		return a.peerlage.horchen.suchen
	}

	return entdeckung.Suche
}

// MitEntdeckung setzt die Nähte zur Ankündigung und zur Suche — für Läufe.
//
// **Beide zusammen und nicht einzeln**: ein Lauf, der ankündigt und dann
// wirklich sucht, fragte das Netz dieses Rechners — und fände, was zufällig
// dort steht (§3).
func (a *App) MitEntdeckung(
	ankuendigen func(entdeckung.Angebot) (Ankuendigung, error),
	suchen func(context.Context, entdeckung.Frage) (entdeckung.Ergebnis, error),
) *App {
	a.peerlage.horchen.ankuendigen = ankuendigen
	a.peerlage.horchen.suchen = suchen

	return a
}

// echteAnkuendigung meldet wirklich im Netz an.
func echteAnkuendigung(angebot entdeckung.Angebot) (Ankuendigung, error) {
	return entdeckung.Melde(angebot)
}

// netzschnittstelle ist, was `internal/entdeckung` als Schnittstelle nimmt.
//
// **Ein eigener Name für `net.Interface`**, damit diese Datei `net` nicht
// nennen muss — sonst brächte sie den Weg ins Netz in die Bindungsfassade
// (§2). Der Typ wird hier nur durchgereicht.
type netzschnittstelle = entdeckung.Netzschnittstelle

// Ankuendigung ist, was sich zurücknehmen lässt.
//
// **Eine Schnittstelle**, damit ein Lauf ohne Netz auskommt (§3) — und aus
// demselben Grund wie [service.Hereinkommend] daneben.
type Ankuendigung interface{ Beende() error }

// **Damit die Zusicherung nicht erst im Betrieb auffällt.**
var _ Ankuendigung = (*entdeckung.Anmeldung)(nil)
