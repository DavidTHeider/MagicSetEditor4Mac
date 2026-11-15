// **Kennwort wechseln und wiederherstellen** — ID-02
// (#1701).
//
// **Eine eigene Datei neben `sitzung.go`**, damit jene unter der Zeilengrenze
// aus §2 bleibt: dort stehen die Übergänge der Sitzung, hier die zwei Wege,
// die ein Kennwort setzen.

package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/kennwort"
)

// KennwortWechseln setzt ein neues Kennwort — mit dem alten, wenn es eines
// gibt.
func (s *Sitzung) KennwortWechseln(ctx context.Context, alt, neu string) error {
	s.schloss.Lock()
	defer s.schloss.Unlock()

	if _, fehler := s.zustand.Nach(domain.SitzungKennwortWechseln); fehler != nil {
		return fehler
	}

	einer, fehler := s.wer(ctx)
	if fehler != nil {
		return fehler
	}

	if einer.Kennwort != "" {
		if falsch := stimmt(einer.Kennwort, alt); falsch != nil {
			return falsch
		}
	}

	gehuetet, fehler := neuesKennwort(neu)
	if fehler != nil {
		return fehler
	}

	return s.ablage.KennwortSetzen(ctx, einer.ID, gehuetet)
}

// Wiederherstellen setzt ein vergessenes Kennwort über einen Code vom Zettel.
//
// **Gesperrt stellt nur der Gesperrte wieder her**: sonst wäre der Zettel
// eines anderen Nutzers ein Weg an dieser Sperre vorbei.
func (s *Sitzung) Wiederherstellen(ctx context.Context, nutzerID int64, code, neu string) error {
	s.schloss.Lock()
	defer s.schloss.Unlock()

	nach, fehler := s.zustand.Nach(domain.SitzungWiederherstellen)
	if fehler != nil {
		return fehler
	}

	if s.zustand == domain.SitzungGesperrt && nutzerID != s.nutzer {
		return fmt.Errorf("%w: gesperrt stellt nur der gesperrte Nutzer wieder her",
			domain.ErrConflict)
	}

	// **Erst hüten, dann den Code suchen**: ein zu kurzes Kennwort kostete
	// sonst einen Code und setzte nichts.
	gehuetet, fehler := neuesKennwort(neu)
	if fehler != nil {
		return fehler
	}

	codeID, fehler := s.faktor.CodeFinden(ctx, nutzerID, code)
	if fehler != nil {
		return fehler
	}

	// **Code, Kennwort und Zustand in einem Geschäft** (Zweitblick über
	// ID-02, Befund 5; Nachlauf, Befund 2) — scheitert es, ist der Code nicht
	// verbraucht, und eine Sperre bleibt eine. **Abgemeldet bleibt
	// abgemeldet** (Befund 3): wer danach arbeitet, meldet sich an — mit
	// Kennwort und zweitem Faktor. Das sagt der Übergang selbst.
	if gescheitert := s.ablage.KennwortMitCodeSetzen(ctx, nutzerID, codeID, gehuetet, nach); gescheitert != nil {
		return gescheitert
	}

	s.zustand = nach
	s.zeigen()

	return nil
}

// stimmt hält ein Kennwort gegen die gehütete Form.
//
// **Leer ist falsch**: ein Nutzer ohne Kennwort entsperrt nicht mit einem —
// dieselbe Regel wie beim Anmelden ([Kontoanmeldung]).
func stimmt(gehuetet, klartext string) error {
	falsch := domain.NewFieldError("kennwort", domain.SchluesselKennwortFalsch,
		"Dieses Kennwort stimmt nicht.")

	if gehuetet == "" {
		return falsch
	}

	if fehler := kennwort.Stimmt(gehuetet, klartext); fehler != nil {
		if errors.Is(fehler, kennwort.ErrStimmtNicht) {
			return falsch
		}

		return fehler
	}

	return nil
}

// neuesKennwort hütet ein neues und nennt das Feld, wenn es nicht taugt.
func neuesKennwort(klartext string) (string, error) {
	gehuetet, fehler := kennwort.Huete(klartext)
	if fehler != nil {
		return "", domain.NewFieldError("neu", domain.SchluesselKennwortLaenge,
			"Ein Kennwort hat mindestens %d und höchstens %d Zeichen.",
			kennwort.Mindestlaenge, kennwort.Hoechstlaenge).
			MitWerten(map[string]string{
				"mindestens": strconv.Itoa(kennwort.Mindestlaenge),
				"hoechstens": strconv.Itoa(kennwort.Hoechstlaenge),
			})
	}

	return gehuetet, nil
}
