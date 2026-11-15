//go:build abnahme && darwin

package fenster

// **Die Leiste** — lesen, auslösen und aufklappen, wie ein Mensch es mit der
// Maus täte.

import (
	"fmt"
	"time"

	"github.com/tmc/apple/objc"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// echteLeiste liest die Leiste, **nachdem** sie sich aktualisiert hat
// (`update`): erst dann sagt `isEnabled`, was ein Klick jetzt täte.
func echteLeiste() []Menuepunkt {
	return punkteAus(objc.Send[objc.ID](nsApp(), objc.Sel("mainMenu")))
}

func punkteAus(menue objc.ID) []Menuepunkt {
	if menue == 0 {
		return nil
	}

	objc.Send[struct{}](menue, objc.Sel("update"))

	dienste := objc.Send[objc.ID](nsApp(), objc.Sel("servicesMenu"))
	anzahl := objc.Send[int](menue, objc.Sel("numberOfItems"))
	punkte := make([]Menuepunkt, 0, anzahl)

	for stelle := range anzahl {
		eintrag := objc.Send[objc.ID](menue, objc.Sel("itemAtIndex:"), stelle)
		unter := objc.Send[objc.ID](eintrag, objc.Sel("submenu"))

		punkte = append(punkte, Menuepunkt{
			Titel: text(eintrag, "title"),
			Kuerzel: kuerzelText(text(eintrag, "keyEquivalent"),
				objc.Send[uint64](eintrag, objc.Sel("keyEquivalentModifierMask"))),
			Aktiv:     objc.Send[bool](eintrag, objc.Sel("isEnabled")),
			Trenner:   objc.Send[bool](eintrag, objc.Sel("isSeparatorItem")),
			Verborgen: objc.Send[bool](eintrag, objc.Sel("isHidden")),
			Aktion:    aktionsname(objc.Send[objc.SEL](eintrag, objc.Sel("action"))),
			Dienste:   unter != 0 && unter == dienste,
			Unter:     punkteAus(unter),
		})
	}

	return punkte
}

// bekannteAktionen sind die Selektoren, die eine Leiste von macOS trägt —
// alles andere ist ein Eintrag mit eigener Handlung (Wails).
var bekannteAktionen = []string{
	"undo:", "redo:", "cut:", "copy:", "paste:", "pasteAsPlainText:",
	"delete:", "selectAll:", "hide:", "hideOtherApplications:",
	"unhideAllApplications:", "terminate:", "arrangeInFront:",
	"performMiniaturize:", "performZoom:", "toggleFullScreen:",
	"startSpeaking:", "stopSpeaking:", "orderFrontStandardAboutPanel:",
	"orderFrontCharacterPalette:", "startDictation:", "performClose:",
	"showHelp:",
}

func aktionsname(selektor objc.SEL) string {
	if selektor == 0 {
		return ""
	}

	for _, name := range bekannteAktionen {
		if objc.Sel(name) == selektor {
			return name
		}
	}

	return "eigene"
}

// klicken löst einen Eintrag aus, **wie ein Klick** — und nur, wenn er aktiv
// ist: einen grauen Eintrag klickt auch ein Mensch nicht.
//
// **Mit einem Fenster** geht eine Handlung ohne eigenes Ziel an dessen
// Ersten Antwortenden — so, wie AppKit sie schickte, wäre das Fenster das
// Schlüsselfenster. Das braucht es, weil eine Anwendung, die aus einem
// Hintergrundprozess startet, seit macOS 14 nicht nach vorn kommt und damit
// kein Schlüsselfenster hat: Widerrufen, Kopieren und Einsetzen liefen sonst
// ins Leere, und ein Mensch hätte vorher ins Fenster geklickt. **Versucht und
// verworfen** am 2026-09-26: `keyWindow` der Anwendung umzubiegen — danach
// kam gar keine Taste mehr an.
func klicken(pfad []string, imFenster objc.ID) error {
	leiste := echteLeiste()

	stellen, fehler := pfadIn(leiste, pfad)
	if fehler != nil {
		return fehler
	}

	menue := objc.Send[objc.ID](nsApp(), objc.Sel("mainMenu"))
	ebene := leiste

	for tiefe, stelle := range stellen {
		if tiefe == len(stellen)-1 {
			if imFenster != 0 {
				return anDenErsten(menue, stelle, imFenster)
			}

			if !ebene[stelle].Aktiv {
				return fmt.Errorf("%w: »%s« ist nicht aktiv", errKeinEintrag, ebene[stelle].Titel)
			}

			objc.Send[struct{}](menue, objc.Sel("performActionForItemAtIndex:"), stelle)

			return nil
		}

		menue = objc.Send[objc.ID](objc.Send[objc.ID](menue, objc.Sel("itemAtIndex:"), stelle), objc.Sel("submenu"))
		ebene = ebene[stelle].Unter
	}

	return nil
}

// anDenErsten schickt die Handlung eines Eintrags an den Ersten Antwortenden
// des Fensters, wenn der Eintrag kein eigenes Ziel hat.
func anDenErsten(menue objc.ID, stelle int, fenster objc.ID) error {
	eintrag := objc.Send[objc.ID](menue, objc.Sel("itemAtIndex:"), stelle)
	handlung := objc.Send[objc.SEL](eintrag, objc.Sel("action"))
	if handlung == 0 {
		return fmt.Errorf("%w: der Eintrag hat keine Handlung", errKeinEintrag)
	}

	// **Die Kette entlang, wie AppKit sie ginge**: vom Ersten Antwortenden
	// über `nextResponder` bis zum Fenster. Direkt an den Ersten geschickt,
	// warf WKWebView bei `undo:` eine Ausnahme — sie kennt die Handlung nicht,
	// ihr Fenster schon.
	ziel := objc.Send[objc.ID](eintrag, objc.Sel("target"))
	for antwortend := objc.Send[objc.ID](fenster, objc.Sel("firstResponder")); ziel == 0 && antwortend != 0; antwortend = objc.Send[objc.ID](antwortend, objc.Sel("nextResponder")) {
		if objc.RespondsToSelector(antwortend, handlung) {
			ziel = antwortend
		}
	}

	if ziel == 0 {
		return fmt.Errorf("%w: niemand in der Kette nimmt %s an", errKeinEintrag, aktionsname(handlung))
	}

	objc.Send[bool](nsApp(), objc.Sel("sendAction:to:from:"), handlung, ziel, eintrag)

	return nil
}

// aufklappen öffnet ein Untermenü der Leiste als Aufklappmenü über dem
// Fenster, liest, was darin steht, **während es offen ist**, und schließt es
// wieder. So sieht man, was macOS erst beim Öffnen einträgt — die Dienste.
//
// Das Öffnen hält den Hauptfaden in der Menüschleife fest; gelesen und
// geschlossen wird aus dieser Schleife heraus (`InvokeSync` läuft dort mit).
func aufklappen(fenster *application.WebviewWindow, pfad []string) ([]string, error) {
	var menue objc.ID

	fehler := application.InvokeSyncWithError(func() error {
		stellen, fehler := pfadIn(echteLeiste(), pfad)
		if fehler != nil {
			return fehler
		}

		menue = objc.Send[objc.ID](nsApp(), objc.Sel("mainMenu"))
		for _, stelle := range stellen {
			menue = objc.Send[objc.ID](objc.Send[objc.ID](menue, objc.Sel("itemAtIndex:"), stelle), objc.Sel("submenu"))
		}

		if menue == 0 {
			return fmt.Errorf("%w: »%s« hat kein Untermenü", errKeinEintrag, pfad[len(pfad)-1])
		}

		return nil
	})
	if fehler != nil {
		return nil, fehler
	}

	application.InvokeAsync(func() {
		objc.Send[bool](menue, objc.Sel("popUpMenuPositioningItem:atLocation:inView:"),
			objc.ID(0), punkt{X: 40, Y: 40}, webansicht(fenster))
	})

	time.Sleep(1500 * time.Millisecond)

	return application.InvokeSyncWithResult(func() []string {
		anzahl := objc.Send[int](menue, objc.Sel("numberOfItems"))
		gelesen := make([]string, 0, anzahl)

		for stelle := range anzahl {
			eintrag := objc.Send[objc.ID](menue, objc.Sel("itemAtIndex:"), stelle)
			if !objc.Send[bool](eintrag, objc.Sel("isSeparatorItem")) {
				gelesen = append(gelesen, text(eintrag, "title"))
			}
		}

		objc.Send[struct{}](menue, objc.Sel("cancelTracking"))

		return gelesen
	}), nil
}
