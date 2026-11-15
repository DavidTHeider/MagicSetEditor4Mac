package service

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **Die drei Funktionen, die selbst schreiben** —
// #657.
//
// # Erschlossen, nicht erfunden
//
//	ctx.setVariable(_("copy_file"),        script_copy_file);
//	ctx.setVariable(_("write_text_file"),  script_write_text_file);
//	ctx.setVariable(_("write_image_file"), script_write_image_file);
//	ctx.setVariable(_("write_set_file"),   script_write_set_file);
//
// `src/script/functions/export.cpp`, Zeilen 449–458, Fassung f9d9356 vom
// 2021-07-06. `write_image_file` steht seit
// #216 in
// `exportbild.go`; die drei Geschwister stehen hier.
//
// Über `M15-Magic-Pack` rufen sie vier der neunzehn Vorlagen:
// `magic-spoiler` alle drei, `magic-cockatrice-v2`, `magic-lackeyccg` und
// `magic-planesculptors` je `write_text_file`.
//
// # Was jede tut
//
//	copy_file(input)               → liest `input` aus dem **Vorlagenpaket**
//	write_text_file(text, file: …) → schreibt `text` mit BOM
//	write_set_file(file: …)        → legt das Set daneben
//
// Alle drei geben den **Namen** zurück; die Vorlagen hängen ihn in ein
// `<img src="…">` oder in eine Zeile ihrer Ausgabe.
//
// # Warum eine Naht und nicht drei
//
// **Weil MSE eine Schranke hat und nicht drei**: `guard_export_info` steht vor
// jeder der Funktionen und fragt dasselbe — läuft gerade eine Ausfuhr? Wer
// hier drei Fähigkeiten einzeln vergäbe, könnte zwei davon setzen und die
// dritte vergessen, und niemand sähe es.

// Beigabenschreiber ist die Naht zu den Funktionen, die Dateien ablegen.
//
// **Ohne ihn gibt es die Namen nicht** — dieselbe Regel wie beim
// [Bildschreiber]: eine Vorlage, die `copy_file` ruft, bekommt dann „der Name …
// ist hier nicht bekannt" und nicht eine Datei an einem Ort, den niemand
// gewählt hat.
type Beigabenschreiber interface {
	// AusVorlage gibt eine Datei des Vorlagenpakets.
	AusVorlage(name string) ([]byte, error)

	// Lege schreibt Bytes neben die Ausgabe und gibt den Dateinamen.
	Lege(datei string, inhalt []byte) (string, error)

	// LegeSet schreibt das Set neben die Ausgabe und gibt den Namen.
	LegeSet(ctx context.Context, datei string) (string, error)
}

// dateikopierer ist `copy_file`.
//
//	SCRIPT_FUNCTION(copy_file) {
//	  SCRIPT_PARAM_C(String, input); // file to copy
//	  String out_name = input;
//	  String out_path = get_export_full_path(out_name);
//	  auto in = ei.export_template->openIn(input);
//	  …
//	  SCRIPT_RETURN(out_name);
//	}
//
// `src/script/functions/export.cpp`, Zeilen 369–383, Fassung f9d9356 vom
// 2021-07-06. **Der Name ist zugleich Quelle und Ziel**: gelesen wird aus dem
// Vorlagenpaket, geschrieben unter demselben Namen neben die Ausgabe.
func (l *ausfuhrlauf) dateikopierer() msescript.Eingebaut {
	return func(_ context.Context, ruf msescript.Ruf) (msescript.Wert, error) {
		name, fehler := einDateiname(ruf, "copy_file", "input")
		if fehler != nil {
			return msescript.Nichts, fehler
		}

		inhalt, lesefehler := l.mittel.Beigaben.AusVorlage(name)
		if lesefehler != nil {
			return msescript.Nichts, lesefehler
		}

		gelegt, schreibfehler := l.mittel.Beigaben.Lege(name, inhalt)
		if schreibfehler != nil {
			return msescript.Nichts, schreibfehler
		}

		l.abgelegt++

		return msescript.Text(gelegt), nil
	}
}

// textschreiber ist `write_text_file`.
//
//	SCRIPT_PARAM_C(String, input); // text to write
//	SCRIPT_PARAM(String, file);    // file to write to
//	tout.WriteString(BYTE_ORDER_MARK);
//	tout.WriteString(input);
//
// `export.cpp`, Zeilen 385–398. **Mit Byte Order Mark** — MSE schreibt ihn
// ausdrücklich, und die Vorlagen erwarten eine Datei, die Windows-Werkzeuge
// als UTF-8 erkennen.
func (l *ausfuhrlauf) textschreiber() msescript.Eingebaut {
	return func(_ context.Context, ruf msescript.Ruf) (msescript.Wert, error) {
		if len(ruf.Stellung) != 1 {
			return msescript.Nichts, fmt.Errorf(
				"%w: write_text_file braucht genau einen Text", domain.ErrValidation)
		}

		datei, fehler := einDateiname(ruf, "write_text_file", "file")
		if fehler != nil {
			return msescript.Nichts, fehler
		}

		gelegt, schreibfehler := l.mittel.Beigaben.Lege(datei,
			append([]byte(bom), ruf.Stellung[0].AlsText()...))
		if schreibfehler != nil {
			return msescript.Nichts, schreibfehler
		}

		l.abgelegt++

		return msescript.Text(gelegt), nil
	}
}

// bom ist die Byte-Order-Mark, die MSE vor einen geschriebenen Text setzt.
const bom = "\ufeff"

// setschreiber ist `write_set_file`.
//
//	SCRIPT_PARAM(String, file);
//	String out_path = get_export_full_path(file);
//	set->saveCopy(out_path);
//
// `export.cpp`, Zeilen 429–439.
//
// **Als Archiv seit Werkstück C von Phase 4.9**
// (#1324),
// und das ist eine Änderung des Verhaltens (§15, Satz 4):
//
//	void Package::saveCopy(const String& name) {
//	  saveToZipfile(name, true, true);
//	  clearKeepFlag();
//	}
//
// `src/util/io/package.cpp`, Zeilen 112–115, Fassung f9d9356 vom 2021-07-06.
// **MSE2 schreibt ein ZIP.**
//
// **Hier stand »als Verzeichnis und nicht als Archiv«** mit dem Verweis auf
// [0018](../../docs/entscheidungen/0018-ein-archiv-wird-nicht-zurueckgeschrieben.md):
// ein ZIP-Schreiber weiche die Byte-Zusage auf. **Der Vermerk deckt diesen Fall
// nicht** — er heißt „wird nicht **zurück**geschrieben" und wägt den Rundlauf
// über die Herkunft ab. Eine Kopie neben die Ausgabe schreibt nichts zurück.
//
// **⌘S bleibt unverändert**; dort gilt 0018 weiter. Wo das Archiv entsteht,
// steht in `app.beigaben.LegeSet`.
func (l *ausfuhrlauf) setleger() msescript.Eingebaut {
	return func(ctx context.Context, ruf msescript.Ruf) (msescript.Wert, error) {
		datei, fehler := einDateiname(ruf, "write_set_file", "file")
		if fehler != nil {
			return msescript.Nichts, fehler
		}

		gelegt, schreibfehler := l.mittel.Beigaben.LegeSet(ctx, datei)
		if schreibfehler != nil {
			return msescript.Nichts, schreibfehler
		}

		l.abgelegt++

		return msescript.Text(gelegt), nil
	}
}

// einDateiname liest einen Namen aus einem benannten oder gestellten Argument.
//
// **`copy_file` nimmt ihn gestellt, die beiden anderen benannt** — so schreiben
// es die Vorlagen des Packs, und so liest MSE sie (`SCRIPT_PARAM_C(String,
// input)` gegen `SCRIPT_PARAM(String, file)`).
func einDateiname(ruf msescript.Ruf, funktion, name string) (string, error) {
	if wert, benannt := ruf.Benannter(name); benannt {
		return wert.AlsText(), nil
	}

	if name == "input" && len(ruf.Stellung) == 1 {
		return ruf.Stellung[0].AlsText(), nil
	}

	return "", fmt.Errorf("%w: %s braucht ein Argument `%s:`",
		domain.ErrValidation, funktion, name)
}

// halberStand nennt, was schon auf der Platte liegt — oder nichts.
//
// **Ein Hinweis, der immer erscheint, ist keiner**: wer ihn jedes Mal liest,
// sieht ihn beim einen Mal nicht, an dem er zählt. Deshalb steht er nur da,
// wenn wirklich etwas geschrieben wurde
// (#750).
//
// **Die Zahl und das Verzeichnis, nicht die Namen.** Wer nachsehen will, geht
// ohnehin ins Verzeichnis; eine Liste von dreihundert Symbolbildern in einer
// Fehlermeldung liest niemand.
func (l *ausfuhrlauf) halberStand(verzeichnis string) string {
	if l.abgelegt == 0 {
		return ""
	}

	return fmt.Sprintf(" — %d Datei(en) liegen bereits in %q; sie werden nicht "+
		"zurückgenommen", l.abgelegt, verzeichnis)
}
