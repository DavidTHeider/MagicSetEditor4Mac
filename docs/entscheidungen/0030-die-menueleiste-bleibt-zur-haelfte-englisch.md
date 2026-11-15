# 0030 — Die Menüleiste bleibt zur Hälfte englisch

**Stand:** angenommen
**Datum:** 2026-08-19

## Lage

`AGENTS.md` §9 sagt zu: Oberflächentexte sind zweisprachig, und Deutsch ist die
Ausgangssprache. Die native Menüleiste hält das nicht ein —
#255.

**Und es sind drei Menüs, nicht zwei.** Das Issue nannte „Edit" und „Window";
nachgesehen am 2026-08-19 ist auch das **Anwendungsmenü** englisch, mit
„About …", „Hide …", „Hide Others", „Show All" und „Quit …". Es fällt weniger
auf, weil sein Titel der Name der Anwendung ist.

Deutsch stehen also nur die beiden Menüs, die dieses Projekt selbst baut —
„Datei" und „Ansicht".

## Warum, nachgelesen

`internal/app/menue.go` hängt `menu.AppMenu()`, `menu.EditMenu()` und
`menu.WindowMenu()` an. Alle drei geben in Wails **v2.13.0** nur eine Rolle
zurück; gebaut wird das Menü auf der Plattformseite, mit **fest verdrahteten
Zeichenketten in Objective-C**:

```objc
WailsMenu *editMenu = [[[WailsMenu new] initWithNSTitle:@"Edit"] autorelease];
[editMenu addItem:[self newMenuItem:@"Undo" :@selector(undo:) :@"z" :NSEventModifierFlagCommand]];
```

`internal/frontend/desktop/darwin/WailsMenu.m`, Zeilen 89–100 — kein
`NSLocalizedString`, keine Tabelle, keine Naht.

Die Rollen für die **einzelnen** Einträge — `UndoRole`, `CutRole`, `CopyRole`,
`PasteRole`, `SelectAllRole`, `ZoomRole` — stehen in `pkg/menu/menuroles.go`
**auskommentiert** da, samt der Funktionen, die sie zurückgäben. In der
gepinnten Fassung gibt es sie nicht.

## Die vier Wege, und was jeder kostet

| Weg | Preis |
| --- | --- |
| **Das Menü selbst bauen, mit eigenen Rückrufen** | Undo, Ausschneiden, Kopieren, Einfügen sind WebView-Aktionen. Von Go aus hieße das, JavaScript ins Fenster zu schicken (`document.execCommand`) — abgekündigt, und es griffe nicht in jedem Eingabefeld gleich |
| **Das Menü selbst bauen, ohne Rückrufe** | Ein `menu.MenuItem` ohne `Click` ist ein Eintrag, der beim Anklicken nichts tut. Und mit `Accelerator`: macOS schluckt die Tastenkombination am Menü, statt sie an die WebView weiterzureichen — ⌘C bräche. **Schlechter als englisch** |
| **Wails anheben** | Die Rollen sind in v2 seit Jahren auskommentiert. Ob v3 sie hergibt, ist zu prüfen; v3 ist zugleich ein größerer Umbau, und dieses Projekt pinnt (§1) |
| **Ein Beitrag an Wails** | Der ehrlichste Weg, und keiner, der sich hier terminieren lässt |

## Entscheidung

**Die Lücke wird benannt, nicht überdeckt.** Sie steht in `README.md` unter
einer eigenen Überschrift — nicht als Nebensatz, wie #255 ausdrücklich
verlangt.

**Und sie wird nachgezählt.** `internal/werkzeuge/menuerollen_test.go` liest
das gepinnte Wails-Modul: er zerteilt `menuroles.go` und prüft, ob eine der
sieben Rollen inzwischen **erklärt** ist, und er sieht in `WailsMenu.m` nach,
ob die Beschriftungen weiter als Literale dastehen. Hebt jemand Wails an und
die Rollen sind da, wird er rot — und dann gehört die Lücke geschlossen, nicht
diese Notiz fortgeschrieben.

**Was er nicht prüft**, steht neben ihm: ob die Beschriftungen heute noch
englisch sind. Er sucht die Rollen, nicht die Wörter.

## Der Preis

Wer die Anwendung auf Deutsch stellt, bekommt eine Menüleiste, die zur Hälfte
nicht mitgeht. Das ist ein sichtbarer Widerspruch zur eigenen Zusage, und er
bleibt sichtbar, bis Wails ihn hergibt.

**Was er nicht kostet:** die Bedienung. Die englischen Einträge tun, was sie
sagen, und ihre Tastenkombinationen sind die von macOS.

## Zerteilt und nicht gelesen

Der erste Wächter suchte Zeilen ohne `//` davor — und war bei jedem Lauf rot,
weil die Rollen ein **zweites** Mal vorkommen: in den Funktionen `Undo()`,
`Cut()` und so fort, die zusammen in *einem* `/* … */`-Block stehen. Ein
Wächter, der Blockkommentare für Code hält, wird abgeschaltet. Er zerteilt die
Datei jetzt mit `go/parser` und fragt, was **erklärt** ist.

## Nachtrag vom 2026-09-24: Wails 3 gibt die Rollen her

**Der Fall, für den der Wächter da war, ist eingetreten** — nicht durch ein
Anheben von Wails 2, sondern durch den Wechsel auf Wails 3 (ENG-02,
[0132](0132-wails-3-als-huelle.md)). `pkg/application/menuitem.go` führt
`NewRole` und `SetLabel`; ein Eintrag mit Rolle behält seinen Selektor, auch
wenn er eine eigene Beschriftung bekommt (nachgelesen in `menuitem_darwin.go`,
`getSelectorForRole`, v3.0.0-beta.23).

**Der Wächter ist mit Wails 2 gegangen** — er las ein Modul, das es im Baum
nicht mehr gibt. **Die Notiz wird nicht fortgeschrieben, die Lücke wird
geschlossen**, wie es oben steht: in einem eigenen Zug nach dem Wechsel,
damit der Wechsel als Ganzes zurückzunehmen bleibt (§5), unter
#255.

**Geschlossen mit #255.** Anwendungs-, Bearbeiten- und Fenstermenü sind Bäume
aus Einträgen mit Rolle, beschriftet aus dem Menükatalog; „Über" und die
Einstellungen stehen im Anwendungsmenü. **Was dabei nachgelesen wurde und
anders ist als in Wails' eigenen Menüs**: „Show All" und „Bring All to
Front" hängen dort an Rollen ohne Selektor (`UnHide`, `Front`) und tun
nichts — hier an `ShowAll` und `BringAllToFront`; „Paste and Match Style"
hat gar keinen und fehlt deshalb; „Delete" bringt die Rücktaste als Kürzel
mit, und die fällt weg (`menuitem_selectors_darwin.go`, `menuitem_roles.go`,
v3.0.0-beta.23). Ein Wächter hält jede Rolle der Leiste gegen eine, die
wirkt (`internal/fenster/menue_test.go`).

