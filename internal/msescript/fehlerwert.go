package msescript

import (
	"errors"
	"fmt"
)

// **Ein fehlgeschlagener Zugriff ist ein Wert, kein Abbruch.**
//
//	ScriptValueP ScriptValue::getMember(const String& name) const {
//	  return delay_error(ScriptErrorNoMember(typeName(), name));
//	}
//
// `src/script/value.cpp`, Zeilen 67–74, Fassung f9d9356 vom 2021-07-06. Der
// zurückgegebene Wert ist ein `ScriptDelayedError`: **jede** Umwandlung wirft,
// aber ein weiterer Zugriff reicht ihn durch (Zeilen 140–167).
//
// # Wozu MSE das braucht
//
//	case I_OR_ELSE:
//	  if (at == SCRIPT_ERROR) a = b;
//
// `src/script/context.cpp`, Zeilen 535–537. **`or else` fängt genau diesen
// Wert** — und nicht `nil`. Die Vorlagen rechnen damit:
//
//	uninitialized := (style.layout.blocks[1].bottom or else 0) <= 0
//
// `magic.mse-game/script`, `flavor_bar_top`. **406 Befunde** über die 190 Stile
// von `M15-Magic-Pack` lauteten auf diesen einen Zugriff
// (#345).

// ArtFehlerwert ist die Art eines verzögerten Fehlers.
const ArtFehlerwert WertArt = "Fehler"

// ErrVerzoegert ist der Fehler, den ein benutzter Fehlerwert wirft.
var ErrVerzoegert = errors.New("dieser Zugriff ging ins Leere")

// Fehlerwert baut einen verzögerten Fehler.
//
// **Die Meldung wird mitgetragen**, nicht ersetzt: wer ihn benutzt, soll
// erfahren, **welcher** Zugriff ins Leere ging — „ein Wert ist nicht da" wäre
// eine Zeile im Protokoll und keine Auskunft (§8).
func Fehlerwert(meldung string) Wert {
	return Wert{Art: ArtFehlerwert, Text: meldung}
}

// IstFehler sagt, ob dieser Wert ein verzögerter Fehler ist.
func (w Wert) IstFehler() bool { return w.Art == ArtFehlerwert }

// alsFehler gibt den Fehler, den dieser Wert beim Benutzen wirft.
func (w Wert) alsFehler() error {
	return fmt.Errorf("%w: %s", ErrVerzoegert, w.Text)
}
