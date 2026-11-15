// **Die lokale Sitzung** — ID-02
// (#1701).
//
// # Was hier entschieden wird
//
// **Wann eine Sitzung sperren, sich lösen, enden und wiederkommen darf** —
// die Übergänge selbst stehen in [domain.Sitzungszustand]; hier steht, was
// jeder von ihnen **vorher** prüft. Alles bleibt auf diesem Rechner: keine
// zentrale Registrierung, keine Bestätigung über eine Adresse (AK-01).
//
// # Was hier **nicht** steht
//
// **Keine Rolle und kein Peer.** Eine Wiederherstellung setzt das Kennwort und
// sonst nichts; wer danach mehr darf als vorher, hätte es über einen Zettel
// bekommen, und das wäre der Fehler, den AK-04 ausschließt.

package service

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Sitzungsablage ist die Naht zum Bestand.
//
// **Der Zustand wird festgehalten**, damit ⌘Q eine Sperre nicht aufhebt;
// **wer** gesperrt ist, sagt der Zeiger auf den arbeitenden Nutzer.
type Sitzungsablage interface {
	Nutzer(ctx context.Context) ([]domain.Nutzer, error)
	KennwortSetzen(ctx context.Context, nutzerID int64, gehuetet string) error
	SitzungLesen(ctx context.Context) (domain.Sitzungszustand, error)
	SitzungFesthalten(ctx context.Context, zustand domain.Sitzungszustand) error

	// KennwortMitCodeSetzen verbraucht den Code vom Zettel, setzt das
	// Kennwort und hält den Zustand `danach` fest — **in einem Geschäft**,
	// alles oder nichts (Zweitblick über ID-02, Befund 5; Nachlauf, Befund 2).
	KennwortMitCodeSetzen(ctx context.Context, nutzerID, codeID int64, gehuetet string, danach domain.Sitzungszustand) error

	// AnmeldungFesthalten stellt beide Zeiger und hält die Sitzung als
	// angemeldet fest — **in einem Geschäft** (Nachlauf über ID-02, Befund 2).
	AnmeldungFesthalten(ctx context.Context, nutzerID, profilID int64) error
}

// Anmeldeziel sagt, wohin eine Anmeldung zeigt.
//
// **Zwei Hälften, weil nur eine davon scheitern darf**: das Profil wird vor
// dem Geschäft bestimmt und nur gelesen; was danach im Prozess umzustellen
// ist, geschieht, wenn der Bestand es schon festhält — und kann nicht mehr
// scheitern. Vorher war es **ein** Schritt, der die Zeiger schrieb, bevor der
// Zustand festgehalten war.
type Anmeldeziel struct {
	// Profil nennt das Profil des Nutzers, auf das die Anmeldung zeigt.
	Profil func(ctx context.Context, nutzerID int64) (int64, error)

	// Nachziehen stellt im Prozess auf dieses Profil um; nil lässt es aus.
	Nachziehen func(ctx context.Context, profilID int64)
}

// Sitzungsfaktor ist die Naht zu Sensor und Zettel — im Betrieb
// [Zweiterfaktor].
type Sitzungsfaktor interface {
	SensorMoeglich(ctx context.Context, nutzerID int64) (bool, error)
	SensorPruefen(ctx context.Context, nutzerID int64) error
	CodeFinden(ctx context.Context, nutzerID int64, code string) (int64, error)

	// Pruefe ist die **strenge** Prüfung beim Anmelden — der Nutzer kann sie
	// nicht abschalten, anders als die beim Profilwechsel.
	Pruefe(ctx context.Context, nutzerID int64, nachweis string) error
}

// Sitzung hält, ob und wer an diesem Fenster angemeldet ist.
//
// **Ein Schloss über jedem Übergang** (§7, Regel 4): ein zweites Entsperren,
// während das erste noch rechnet, ginge gegen einen Stand, den es nicht mehr
// gibt.
type Sitzung struct {
	ablage Sitzungsablage
	faktor Sitzungsfaktor

	schloss sync.Mutex
	zustand domain.Sitzungszustand
	nutzer  int64

	// sicht ist der Stand **zum Lesen ohne Schloss** — geschrieben nach jedem
	// Übergang, unter dem Schloss. **Nicht das Schloss selbst**: der
	// Nutzerwechsel beim Anmelden läuft unter ihm und fragt über die
	// Profilnaht nach dem Stand; ein `sync.Mutex` ist nicht
	// wiedereintrittsfähig, und die Frage klemmte für immer.
	sicht atomic.Pointer[sitzungssicht]
}

// sitzungssicht ist ein Stand, wie [Sitzung.Stand] ihn nennt.
type sitzungssicht struct {
	zustand domain.Sitzungszustand
	nutzer  int64
}

// NeueSitzung baut sie abgemeldet.
func NeueSitzung(ablage Sitzungsablage, faktor Sitzungsfaktor) *Sitzung {
	sitzung := &Sitzung{ablage: ablage, faktor: faktor, zustand: domain.SitzungAbgemeldet}
	sitzung.zeigen()

	return sitzung
}

// zeigen macht den Stand lesbar; das Schloss hält der Rufer.
func (s *Sitzung) zeigen() {
	s.sicht.Store(&sitzungssicht{zustand: s.zustand, nutzer: s.nutzer})
}

// Beginnen nimmt den festgehaltenen Zustand auf — für den Nutzer, mit dem
// das Programm startet.
//
// **Kein Übergang, sondern der Anfang**: wer die Anwendung öffnet, hat sich
// an diesem Rechner schon angemeldet — beim Betriebssystem. **Außer**, er hat
// gesperrt oder sich abgemeldet, bevor er sie schloss.
func (s *Sitzung) Beginnen(ctx context.Context, nutzerID int64) error {
	zustand, fehler := s.ablage.SitzungLesen(ctx)
	if fehler != nil {
		return fehler
	}

	s.schloss.Lock()
	defer s.schloss.Unlock()

	s.zustand, s.nutzer = zustand, nutzerID
	if zustand == domain.SitzungAbgemeldet {
		s.nutzer = 0
	}

	s.zeigen()

	return nil
}

// Stand sagt Zustand und Nutzer; abgemeldet ist der Nutzer 0. **Mitten in
// einem Übergang** sagt er den Stand davor.
func (s *Sitzung) Stand() (zustand domain.Sitzungszustand, nutzerID int64) {
	sicht := s.sicht.Load()

	return sicht.zustand, sicht.nutzer
}

// Anmelden meldet einen Nutzer dieses Rechners an — mit seinem Kennwort,
// wenn er eines hat, und **immer** mit seinem zweiten Faktor, wenn er einen
// hat.
//
// **Die Faktorprüfung ist die strenge** ([Sitzungsfaktor.Pruefe]): die beim
// Profilwechsel kann der Nutzer abschalten, die beim Anmelden nicht — und
// sie gilt auch für den, der vor dem Abmelden schon arbeitete (Zweitblick
// über ID-02, Befund 2). **Zeiger und Zustand gehen in einem Geschäft**
// (Nachlauf, Befund 2); erst danach zieht der Prozess nach, und erst danach
// ist die Sitzung offen. Unter demselben Schloss, damit kein zweiter Ruf
// dazwischen einen halben Stand sieht (§7, Regel 4). **Keine Adresse und
// keine Bestätigung** — ein lokales Konto braucht keine Registrierung
// (AK-01).
func (s *Sitzung) Anmelden(
	ctx context.Context, nutzerID int64, klartext, nachweis string, ziel Anmeldeziel,
) error {
	s.schloss.Lock()
	defer s.schloss.Unlock()

	nach, fehler := s.zustand.Nach(domain.SitzungAnmelden)
	if fehler != nil {
		return fehler
	}

	einer, fehler := s.nutzerMit(ctx, nutzerID)
	if fehler != nil {
		return fehler
	}

	if einer.Kennwort != "" {
		if falsch := stimmt(einer.Kennwort, klartext); falsch != nil {
			return falsch
		}
	}

	if fehlt := s.faktor.Pruefe(ctx, nutzerID, nachweis); fehlt != nil {
		return fehlt
	}

	profilID, fehler := ziel.Profil(ctx, nutzerID)
	if fehler != nil {
		return fehler
	}

	if gescheitert := s.ablage.AnmeldungFesthalten(ctx, nutzerID, profilID); gescheitert != nil {
		return gescheitert
	}

	if ziel.Nachziehen != nil {
		ziel.Nachziehen(ctx, profilID)
	}

	s.zustand, s.nutzer = nach, nutzerID
	s.zeigen()

	return nil
}

// Sperren sperrt — aber nur, wenn es einen Weg zurück gibt.
func (s *Sitzung) Sperren(ctx context.Context) error {
	s.schloss.Lock()
	defer s.schloss.Unlock()

	if _, fehler := s.zustand.Nach(domain.SitzungSperren); fehler != nil {
		return fehler
	}

	einer, fehler := s.wer(ctx)
	if fehler != nil {
		return fehler
	}

	// **Ohne Kennwort und ohne Sensor ließe sich die Sperre nicht lösen** —
	// außer über einen Zettel, und den hat nicht jeder. Eine Sperre, aus der
	// nur ein Neustart führt, wäre keine Sperre, sondern ein Absturz.
	if einer.Kennwort == "" {
		moeglich, fehler := s.faktor.SensorMoeglich(ctx, einer.ID)
		if fehler != nil {
			return fehler
		}

		if !moeglich {
			return domain.NewFieldError("kennwort", domain.SchluesselSperreOhneRueckweg,
				"Ohne Kennwort und ohne lokale Entsperrung ließe sich die Sperre nicht lösen.")
		}
	}

	return s.gehe(ctx, domain.SitzungSperren, s.nutzer)
}

// Entsperren löst die Sperre mit dem Kennwort.
func (s *Sitzung) Entsperren(ctx context.Context, klartext string) error {
	s.schloss.Lock()
	defer s.schloss.Unlock()

	if _, fehler := s.zustand.Nach(domain.SitzungEntsperren); fehler != nil {
		return fehler
	}

	einer, fehler := s.wer(ctx)
	if fehler != nil {
		return fehler
	}

	if fehler := stimmt(einer.Kennwort, klartext); fehler != nil {
		return fehler
	}

	return s.gehe(ctx, domain.SitzungEntsperren, s.nutzer)
}

// EntsperrenUeberDenSensor löst die Sperre über Touch ID — **lokale
// Entsperrung**, kein Passkey (AK-02).
func (s *Sitzung) EntsperrenUeberDenSensor(ctx context.Context) error {
	s.schloss.Lock()
	defer s.schloss.Unlock()

	if _, fehler := s.zustand.Nach(domain.SitzungEntsperren); fehler != nil {
		return fehler
	}

	if fehler := s.faktor.SensorPruefen(ctx, s.nutzer); fehler != nil {
		return fehler
	}

	return s.gehe(ctx, domain.SitzungEntsperren, s.nutzer)
}

// Abmelden beendet die Sitzung.
func (s *Sitzung) Abmelden(ctx context.Context) error {
	s.schloss.Lock()
	defer s.schloss.Unlock()

	return s.gehe(ctx, domain.SitzungAbmelden, 0)
}

// gehe führt einen Übergang aus und hält ihn fest; das Schloss hält der
// Rufer.
//
// **Erst festhalten, dann umstellen**: was sich nicht festhalten lässt,
// geschieht nicht — sonst zeigte das Fenster gesperrt, und der nächste Start
// begänne angemeldet.
func (s *Sitzung) gehe(ctx context.Context, ereignis domain.Sitzungsereignis, nutzerID int64) error {
	nach, fehler := s.zustand.Nach(ereignis)
	if fehler != nil {
		return fehler
	}

	if nach != s.zustand {
		if fehler := s.ablage.SitzungFesthalten(ctx, nach); fehler != nil {
			return fehler
		}
	}

	s.zustand, s.nutzer = nach, nutzerID
	s.zeigen()

	return nil
}

// wer liest den Nutzer dieser Sitzung aus dem Bestand.
func (s *Sitzung) wer(ctx context.Context) (domain.Nutzer, error) {
	return s.nutzerMit(ctx, s.nutzer)
}

// nutzerMit liest einen Nutzer aus dem Bestand.
func (s *Sitzung) nutzerMit(ctx context.Context, nutzerID int64) (domain.Nutzer, error) {
	alle, fehler := s.ablage.Nutzer(ctx)
	if fehler != nil {
		return domain.Nutzer{}, fehler
	}

	for _, einer := range alle {
		if einer.ID == nutzerID {
			return einer, nil
		}
	}

	return domain.Nutzer{}, fmt.Errorf("%w: den Nutzer %d gibt es nicht", domain.ErrNotFound, nutzerID)
}
