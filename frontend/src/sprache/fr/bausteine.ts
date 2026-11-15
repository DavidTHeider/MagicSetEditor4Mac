import { dateiansicht } from './bausteinedatei'
import { ablage } from './bausteineablage'
import { geste } from './bausteinegeste'
import { laufstand } from './bausteinelauf'
import { objektmarken } from './bausteinemarken'
import { teilen } from './bausteineteilen'

export const bausteine = {
  dateiansicht,
  ablage,
  geste,
  laufstand,
  objektmarken,
  teilen,
  entwurf: {
    fremdGeaendert:
      'Entre-temps, ce champ indique « {wert} » — modifié ailleurs. Ton brouillon n’est pas encore enregistré.',
    behalten: 'Enregistrer mon brouillon',
    nehmen: 'Prendre la nouvelle valeur',
  },
  listenansicht: {
    suche: 'Rechercher',
    suchePlatzhalter: 'Nom ou jeu',
    sortierung: 'Trier',
    ansicht: 'Présentation',
    liste: 'Liste',
    kacheln: 'Mosaïque',
    zahl: '{gezeigt} sur {gesamt}',
    nichtsPasst: 'Rien ne correspond à « {suche} ».',
    sucheLeeren: 'Effacer la recherche',
    filterZuruecksetzen: 'Réinitialiser la recherche et les filtres',
    weitere: 'Afficher {anzahl} de plus',
    wieImBestand: 'Tel qu’enregistré',
    zuletztBenutzt: 'Utilisés récemment d’abord',
    nachName: 'Par nom',
    nachKarten: 'Par nombre de cartes',
  },
  speicherstand: {
    schreibt: 'Enregistrement…',
    geschrieben: 'Chaque modification est enregistrée.',
    fehler: '« {feld} » n’est pas enregistré — la raison figure sur le champ.',
  },
  gegenueberTitel: 'Ajouter ou confirmer un correspondant',
  kontaktwahlEintrag: '{name} · {fingerabdruck}',
  kontaktwahlSuchen: 'Rechercher un correspondant …',
  feldhilfe: 'Explication pour {feld}',
  mehrZu: 'En savoir plus sur « {satz} »',

  // **Where something came from** — Werkstück 12 of Phase 4.12
  // (#1351).
  herkunftAusProfil: 'Repris du profil {quelle}.',
  herkunftAusEinfuhr: 'Importé depuis {quelle}.',

  pflichtfeldLeer: '{feld} nécessite une valeur.',

  waehrung: {
    beispiel: 'p. ex. EUR',
    keineGestalt:
      '« {wert} » n’est pas un code de devise — il comporte trois lettres, comme EUR.',
    grossSchreiben:
      '« {wert} » — un code de devise s’écrit en majuscules : {gross}.',
  },

  drucksprachen: {
    beispiel: 'p. ex. en',
    namen: {
      en: 'Anglais',
      es: 'Espagnol',
      fr: 'Français',
      de: 'Allemand',
      it: 'Italien',
      pt: 'Portugais',
      ja: 'Japonais',
      ko: 'Coréen',
      ru: 'Russe',
      zhs: 'Chinois simplifié',
      zht: 'Chinois traditionnel',
      he: 'Hébreu',
      la: 'Latin',
      grc: 'Grec ancien',
      ar: 'Arabe',
      sa: 'Sanskrit',
      ph: 'Phyrexian',
      qya: 'Quenya',
      dw: 'Nain',
      tlh: 'Klingon',
    },
  },

  // **The password field** — piece 6 of phase 4.12 (#1351).
  kennwort: {
    zeigen: 'Afficher le mot de passe en clair',
    zeigenKurz: 'Afficher',
    verbergen: 'Masquer à nouveau le mot de passe',
    verbergenKurz: 'Masquer',
  },

  // **The field for the second factor** — piece 7 of phase 4.12 (#1351).
  nachweis: {
    beschriftung: 'Code',
    abschicken: 'Confirmer',
  },

  aktionen: {
    abbrechen: 'Annuler',
    oeffnen: 'Ouvrir',
    schliessen: 'Fermer',
    speichern: 'Enregistrer',
  },
  stilliste: {
    mitZusatz: '{kennung} · {zusatz}',
    keinerGewaehlt: '— Choisir un style —',
  },
  druckwahl: {
    drucke: 'Impressions',
    hinweis:
      'Ce pair ne voit que ces impressions, avec la quantité que tu possèdes.',
    suchen: 'Rechercher des impressions',
    gewaehlt: 'Sélection : {anzahl}',
    nurSuche: 'Ta collection compte {anzahl}. Cherche un nom pour en choisir.',
    keiner: 'Aucune impression ne correspond à « {suche} ».',
    genauer:
      'Plus d’impressions correspondent que celles affichées. Affine la recherche.',
    leer: 'Ta collection ne contient encore aucune impression à partager. Ajoute d’abord des exemplaires.',
    fehlt:
      'Choisis au moins une impression : une sélection sans impression ne partage rien.',
    behalten: 'Exemplaires de chaque impression que je garde',
    behaltenHinweis:
      'Ce pair voit ce qui dépasse ce nombre : avec 1 et trois exemplaires, deux.',
    detailBehalten: 'tu gardes {behalten}',
  },
  suchauswahl: {
    keiner: 'Rien ne correspond à « {suche} ».',
    treffer: 'Résultats : {anzahl}',
    begrenzt:
      '{gezeigt} résultats sur {anzahl} affichés — affine la recherche …',
  },
  zeitzonenwahl: {
    system: 'Comme le système',
  },
  aktionskatalog: {
    aktionen: {
      anlegen: { kurz: 'Créer', heisst: 'Crée un nouvel élément.' },
      bearbeiten: { kurz: 'Modifier', heisst: 'Le modifie sur place.' },
      sichern: { kurz: 'Enregistrer', heisst: 'Conserve l’état actuel.' },
      zuruecknehmen: { kurz: 'Annuler', heisst: 'Annule la dernière étape.' },
      duplizieren: {
        kurz: 'Dupliquer',
        heisst: 'Place un second exemplaire à côté.',
      },
      kopieren: { kurz: 'Copier', heisst: 'Le place dans le presse-papiers.' },
      entfernen: {
        kurz: 'Retirer',
        heisst: 'Le retire de cette vue ; il reste ailleurs.',
      },
      loeschen: {
        kurz: 'Supprimer',
        heisst: 'Le supprime définitivement, partout.',
      },
    },
  },
  nurlesen: {
    hinweis:
      'Tu consultes ce profil en tant que lecteur : tu peux tout voir et ne rien modifier. Passe à ton propre profil pour travailler — ou demande à la personne qui gère celui-ci le rôle « collaborateur ».',
    knopf: 'Lecture seule — tu ne peux rien modifier dans ce profil.',
  },

  // **The context strip** — piece U4 of phase 5.1 (#1415).
  kontext: {
    bezeichnung: 'Contexte de travail',
    profilVon: 'appartient à {eigner}',
    rollen: {
      leser: 'Lecteur',
      mitarbeiter: 'Collaborateur',
      verwalter: 'Gestionnaire',
    },
    nurLesen: 'lecture seule',
    zuDenNutzern: 'Changer de profil…',
    zeile: '{name} · {profil}',
  },

  zustaende: {
    laedt: 'Chargement…',
    leer: 'Rien ici',
    fehler: 'Impossible de charger ceci.',
    fehlerMitGrund: 'Impossible de charger ceci. {grund}',
    wiederholen: 'Charger à nouveau',
    neuLaden: 'Recharger',
  },
  // Find and replace **inside the field** (#1164).
  suche: {
    oeffnen: 'Rechercher dans le texte',
    schliessen: 'Fermer la recherche',
    suchen: 'Rechercher',
    ersetzen: 'Remplacer par',
    treffer: '{bei} sur {ganz}',
    keinTreffer: 'Aucun résultat',
    weiter: 'Résultat suivant',
    zurueck: 'Résultat précédent',
    einsetzen: 'Remplacer',
    alleEinsetzen: 'Tout remplacer',
    grossKlein:
      'La casse compte — en MSE-Script, « card » et « Card » sont deux choses différentes.',
  },
  // Completion **inside** the field (#1164).
  // **Welche Karten hinausgehen** (#1167).
  ausgabeauswahl: {
    aufschrift: 'Quelles cartes',
    ganzesSet: 'Le set entier ({zahl} cartes)',
    ganzesSetOhneZahl: 'Le set entier',
    teilmenge: 'Uniquement la sélection ({zahl})',
    ohneNamen: 'sans nom',
    eingrenzen: 'Filtrer les cartes',
  },

  vorschlaege: {
    bezeichnung: 'Suggestions',
    uebernehmen: 'Accepter avec ↹',
  },

  // **The regions of a workspace** — piece U2 of phase 5.1 (#1413).
  arbeitsplatz: {
    breiteVon: 'Largeur de « {region} »',
    breiteWert: '{punkte} points',
    navigatorAuf: 'Afficher le panneau de navigation',
    navigatorZu: 'Masquer le panneau de navigation',
    inspectorAuf: 'Afficher l’inspecteur',
    inspectorZu: 'Masquer l’inspecteur',
    navigatorEng:
      'Trop étroit pour les trois colonnes — le panneau de navigation s’ouvre par-dessus le centre',
    inspectorEng:
      'Trop étroit pour les trois colonnes — l’inspecteur s’ouvre par-dessus le centre',
  },
  // The command centre (#1423).
  druck: {
    vorlaeufig: 'provisoire',
    vorlaeufigErklaerung:
      'Cette impression n’a pas encore d’identifiant Scryfall. Les prix et les images restent absents tant qu’elle n’est pas résolue.',
  },
  kommandozentrale: {
    titel: 'Centre de commandes',
    erklaerung:
      'Une commande est un outil du serveur MCP, appelé ici dans la fenêtre — la même réponse que via MCP et en ligne de commande. Rien ne passe par un shell : ce que tu saisis reste des données.',
    lesend: 'lit',
    schreibend: 'modifie',
    gefaehrlich: 'difficile à annuler',
    faehigkeit: 'Capacité',
    cli: 'Ligne de commande',
    rest: 'Service',
    parameter: 'Paramètres',
    keineParameter: 'Cette commande n’a pas de paramètres.',
    pflicht: 'Obligatoire',
    jsonHinweis:
      'En JSON — une liste entre crochets, un objet entre accolades.',
    befundPflicht: 'Ce champ nécessite une valeur.',
    befundZahl: 'Ce n’est pas un nombre.',
    befundJson: 'Ce n’est pas du JSON.',
    vorschau: 'Ce qui sera envoyé',
    bestaetigen: 'J’ai lu l’aperçu et je veux exécuter ceci tel quel.',
    ausfuehren: 'Exécuter',
    abbrechen: 'Annuler',
    laeuft: 'La commande est en cours d’exécution…',
    ergebnis: 'Résultat',
    ausgabe: 'Sortie de la commande',
    alsText: 'Texte',
    alsJson: 'JSON',
    dauer: '{ms} ms',
    verlaufLeer:
      'Rien n’a encore été exécuté. Les secrets — jetons, mots de passe — n’apparaissent jamais ici en clair.',
    verlauf: 'Historique de cette session',
    wiederholen: 'Recharger {name}',
    gelungen: 'réussi',
    gescheitert: 'échoué',
    abgebrochen: 'annulé',
    schliessen: 'Fermer',
  },
  // „Diagnose öffnen“ an einem Fehler mit Vorgangskennung (NET-02,
  // S13.NAT-04, #1706).
  diagnoseverweis: {
    oeffnen: 'Ouvrir le diagnostic',
  },
  // „KI einrichten“ an einer KI-Meldung (AI-01.AK-04, #1742).
  kieinrichtung: {
    oeffnen: 'Configurer l’IA',
  },
  kikasten: {
    titel: 'IA pour cette carte',
  },
} as const
