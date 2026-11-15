// Die Gruppen des Vertragsrundgangs, in der Reihenfolge, in der sie laufen
// (`../vertrag.mjs`). **Wer löscht, steht am Ende.**
import { aBetrieb } from './a-betrieb.mjs'
import { aEinstellungen } from './a-einstellungen.mjs'
import { aKarten } from './a-karten.mjs'
import { aKI } from './a-ki.mjs'
import { aPakete } from './a-pakete.mjs'
import { aSammlung } from './a-sammlung.mjs'
import { aSets } from './a-sets.mjs'
import { aStil } from './a-stil.mjs'
import { aSymbole } from './a-symbole.mjs'
import { aTurniere } from './a-turniere.mjs'
import { bBestand } from './b-bestand.mjs'
import { bDienste } from './b-dienste.mjs'
import { bHandel } from './b-handel.mjs'
import { bNutzer } from './b-nutzer.mjs'
import { bPartien } from './b-partien.mjs'
import { bPeers } from './b-peers.mjs'
import { bSitzung } from './b-sitzung.mjs'
import { cLesend } from './c-lesend.mjs'
import { decks } from './decks.mjs'
import { karten } from './karten.mjs'
import { leereListen } from './leer.mjs'
import { turniere } from './turniere.mjs'
import { werkzeuge } from './werkzeuge.mjs'

// **Die Kohorte der leeren Listen zuerst** — vor jeder Anlage (C01).
export const kohorten = [leereListen]

export const gruppen = [karten, decks, werkzeuge, turniere, aPakete, aStil, aSymbole, aKarten, aEinstellungen, aSammlung, aTurniere, aSets, aBetrieb, aKI, bNutzer, bPeers, bHandel, bPartien, bDienste, cLesend]

// **Was die Sitzung oder den ganzen Bestand umstellt, läuft nach den übrigen**
// — Anmeldung, Faktoren, Kennwörter, Verschlüsselung, Sicherung, Rücksetzung.
// Danach stimmen die Kennungen der Gruppen davor nicht mehr.
export const schluss = [bSitzung, bBestand]
