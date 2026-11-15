// **Die vier Naht-Gruppen rund um andere Rechner** — die eigene Identität
// (#1454),
// die bekannten Gegenüber, das Erreichbarsein
// (#1455,
// #1456)
// und der portable Umschlag
// (#1457).
//
// **Eine eigene Datei**, weil `api.ts` mit ihnen über die Zeilengrenze aus §2
// ginge. **An der Form ändert sich nichts**: eine Schnittstelle, die vier
// erbt, ist genau die Summe der vier.
import type { Peerapi } from './apipeer'
import type { Horchapi, Kontaktapi } from './apipeerkontakt'
import type { Einladungsapi } from './apikontakteinladung'
import type { Sichtfreigabeapi } from './apisichtfreigabe'
import type { Handelspostapi } from './apihandelspost'
import type { Handelspartnerapi } from './apihandelspartner'
import type { Marktlinkapi } from './apimarktlinke'
import type { Marktkontoapi } from './apimarktkonto'
import type { Handelsverbuchungapi } from './apihandelsverbuchung'
import type { Handelsleitungapi } from './apihandelsleitung'
import type { Handelsverzeichnisapi } from './apihandelsverzeichnis'
import type { Tauschapi } from './apitauschabgleich'
import type { Umschlagapi } from './apiumschlag'
import type { Netzwegapi } from './apinetz'
import type { Diensteapi } from './apidienste'
import type { Teilenapi } from './apiteilen'

// Netzapi ist, was dieses Fenster über andere Rechner kann.
export interface Netzapi
  extends
    Peerapi,
    Kontaktapi,
    Einladungsapi,
    Sichtfreigabeapi,
    Tauschapi,
    Handelspostapi,
    Handelsverbuchungapi,
    Handelsleitungapi,
    Handelsverzeichnisapi,
    Handelspartnerapi,
    Marktlinkapi,
    Marktkontoapi,
    Horchapi,
    Umschlagapi,
    Netzwegapi,
    Diensteapi,
    Teilenapi {}
