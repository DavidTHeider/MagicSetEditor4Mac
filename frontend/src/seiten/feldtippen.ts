// **Tippen in ein Kartenfeld, samt Ersetzung** — aus `Feldeingabe.tsx`
// herausgelöst, als #2290 die Leserrolle dazubrachte und die Datei an die
// Grenze von 400 Zeilen stieß (§2). Der Schnitt folgt der Frage: dort steht,
// **welche Form** ein Feld hat; hier, **was beim Tippen geschieht**.
import type { ChangeEvent } from 'react'

// Ersetzer wendet die `auto replace:`-Regeln an (`Feldeingabe`).
export type Ersetzer = (
  text: string,
  stelle: number,
) => Promise<{ text: string; stelle: number }>

// tippen nimmt eine Eingabe entgegen und ersetzt, wenn getippt wurde.
//
// **Die Schreibmarke kommt vom Feld und geht ans Feld zurück.** Ohne das
// zweite spränge sie nach jeder Ersetzung ans Ende — und wer mitten in einem
// Satz `(C)` tippt, schriebe danach hinten weiter.
export function tippen(
  ereignis: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>,
  nachher: (wert: string) => void,
  autoErsetzen: Ersetzer | undefined,
) {
  const feldknoten = ereignis.target
  const art = (ereignis.nativeEvent as InputEvent).inputType

  if (!autoErsetzen || art !== 'insertText') {
    nachher(feldknoten.value)

    return
  }

  const vorher = feldknoten.value
  const marke = feldknoten.selectionStart ?? vorher.length

  // **Das Getippte ist sofort ein Entwurf** — nicht erst mit der Antwort
  // der Ersetzung (Zweitblick über C und D, Befund 3). Käme dazwischen ein
  // fremder Wert für dieses Feld, sähe der Entwurfshaken sonst keinen
  // Entwurf und baute die Eingabe neu: Text und Fokus wären fort, bevor
  // ein Konflikt greifen kann. **Und der Text steht im Feld, ob die
  // Ersetzung glückt oder nicht** (#918) — sie ist eine Zugabe, nicht die
  // Bedingung des Tippens.
  nachher(vorher)

  const lauf = autoErsetzen(vorher, marke).then((ersetzt) => {
    // **Eine Antwort gehört zu dem Stand, aus dem sie stammt** (§7 Regel 2).
    //
    // Die Ersetzung ist ein asynchroner Aufruf **während des Tippens**: wer
    // schnell tippt, hat zwei davon unterwegs. Käme die erste nach der
    // zweiten zurück und schriebe ihr Ergebnis ins Feld, verlöre der
    // Tippende sein zuletzt getipptes Zeichen — nachgestellt in
    // `autoersetzung.test.tsx`, und es ist wirklich passiert.
    //
    // **Der Stand ist Text und Marke**, nicht ein Zähler daneben: beides ist
    // genau das, was die Antwort verändern will, und beides kann sich nicht
    // von sich selbst unterscheiden.
    //
    // **Der Text allein genügt nicht**: der Aufruf schickt beides, und die
    // Antwort schreibt beides zurück — wer die Marke bewegt, während die
    // Ersetzung läuft, sähe sie danach an die alte Stelle springen.
    if (!derselbeStand(feldknoten, vorher, marke)) {
      return
    }

    if (ersetzt.text !== vorher) {
      feldknoten.value = ersetzt.text
      feldknoten.setSelectionRange(ersetzt.stelle, ersetzt.stelle)
      nachher(ersetzt.text)
    }
  })

  // **Eine abgelehnte Zusage ist kein Fehler des Tippens** (#918): das
  // Getippte ist oben schon gemeldet, und es bleibt, wie es ist.
  lauf.catch(() => undefined)
}

// derselbeStand sagt, ob Feld und Marke noch dort stehen, wo der Aufruf begann.
//
// **Beides zusammen** (§7 Regel 2): eine Antwort schreibt Text **und** Marke
// zurück, also gehört auch beides zu dem Stand, aus dem sie stammt.
function derselbeStand(
  feld: HTMLInputElement | HTMLTextAreaElement,
  text: string,
  marke: number,
): boolean {
  return feld.value === text && (feld.selectionStart ?? text.length) === marke
}
