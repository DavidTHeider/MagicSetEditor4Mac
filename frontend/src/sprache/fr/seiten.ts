// The texts of the individual pages.
import { bestandspflege } from './seitenbestand'
import { datenpflege } from './seitenpflege'
import { kosten } from './seitenkosten'
import { ruecksetzung } from './seitenruecksetzung'
import { ruecksprung } from './seitenruecksprung'
import { einstellungen } from './seiteneinstellungen'
import { nutzer } from './seitennutzer'
import { karten } from './seitenkarten'
import { sprachstaende } from './seitensprachstaende'
import { handeln } from './seitenhandeln'
import { technik } from './seitentechnik'
import { recherche } from './seitenrecherche'
import { decks } from './seitendecks'
import { turniere } from './seitenturniere'
import { handlungen } from './seitenhandlungen'
import { spielen } from './seitenspielen'
import { sammlung } from './seitensammlung'
import { adressbuch } from './seitenadressbuch'
import { vorlagen } from './seitenvorlagen'
import { sets } from './seitensets'
import { ausgeben } from './seitenausgeben'
import { werkzeugseiten } from './seitenwerkzeuge'
import { start } from './seitenstart'
import { paketplan } from './seitenpaketplan'
import { bezuege } from './seitenbezuege'

export const seiten = {
  start,
  sets,
  setangaben: {
    kopf: 'Informations du fichier',
    kopfHinweis:
      'Ce que Magic Set Editor a noté sur ce set lors de l’enregistrement. Ces informations sont conservées telles quelles et ne sont pas modifiées.',
    mseFassung: 'Écrit avec MSE',
    spielfassung: 'Version du paquet de jeu',
    stilfassung: 'Version de la feuille de style',
    versionskontrolle: 'Gestion de versions',
    apprenticecode: 'Code Apprentice',
    titel: 'Informations du set',
    leerTitel: 'Aucun champ d’information',
    leer: 'Le paquet de jeu de ce set ne déclare aucun bloc `set field:`. Choisis un jeu qui en fournit.',
  },
  karten,
  sprachstaende,
  decks,
  turniere,
  spielen,
  handlungen,
  sammlung,
  adressbuch,
  vorlagen,
  ...werkzeugseiten,
  bilder: {
    titel: 'Images',
    erklaerung:
      'Chaque carte est dessinée et écrite séparément. L’exécution se déroule ici, dans la fenêtre : elle indique où elle en est et peut être arrêtée à tout moment.',
    ordnerWaehlen: 'Choisir un dossier …',
    keinOrdner: 'Aucun dossier choisi pour l’instant.',
    gemerkt: 'Dernier choisi — le même dossier pour tous les sets.',
    nichtGemerkt:
      'Le dossier vaut pour cette sortie, mais il n’a pas été mémorisé :',
    vorlage: 'Modèle de nom',
    konflikt: 'Si un fichier existe déjà',
    konfliktHilfe:
      'Deux cartes de même nom donnent le même nom de fichier. « Renuméroter » est le réglage par défaut de Magic Set Editor 2 : il compte en avant au sein d’une exécution et écrase ce qu’une exécution précédente a laissé.',

    // Die Schlüssel kommen aus `internal/domain/einstellung.go`.
    konflikte: {
      altbehalten: 'Conserver l’existant',
      ueberschreiben: 'Écraser',
      nummerieren: 'Numéroter',
      nummerierenneu: 'Renuméroter',
    },
    vorlageHilfe:
      '{nummer} devient le numéro de carte à trois chiffres, {name} le nom de la carte. Les noms identiques reçoivent un numéro à la suite.',
    format: 'Format',
    formatPNG: 'PNG',
    formatJPG: 'JPG',
    formatBMP: 'BMP',
    formatTIFF: 'TIFF',
    formatSVG: 'SVG',
    guete: 'Qualité',
    gueteFast: 'Rapide — pour un aperçu rapide',
    gueteMedium: 'Selon la règle de Magic Set Editor 2',
    gueteHigh: 'Élevée — sans scintillement sur les lignes fines',
    gueteUltra: 'Maximale — pour l’impression en pleine résolution',
    breite: 'Largeur en pixels',
    breiteHilfe:
      'Pas de dpi : aucun modèle n’indique la largeur d’une carte en millimètres. 750 pixels pour une carte de 63 mm de large, c’est 300 dpi.',
    ausgeben: 'Exporter les images',
    abbrechen: 'Arrêter',
    laeuft: 'écrites sur',
    fertig: 'images écrites dans',
    abgebrochen: 'Arrêté.',
    uebersprungen:
      '{anzahl} cartes ignorées — le fichier existait déjà, et « Conserver l’existant » est sélectionné :',
    befunde:
      'Dessin incomplet. Les images existent ; ces éléments y manquent :',
  },
  druckbogen: {
    titel: 'Feuille d’impression',
    erklaerung:
      'Les cartes conservent leurs dimensions : 63 × 88 mm, exactement comme elles sont découpées. Les traits de coupe se trouvent au bord du papier, pas sur la carte.',
    papier: 'Papier',
    papierA4: 'A4',
    papierLetter: 'Letter',
    anschnitt: 'Fond perdu en millimètres',
    anschnittHilfe:
      'Le fond perdu est la partie de l’image qui dépasse la ligne de coupe — il évite qu’une carte découpée présente un bord blanc. Les cartes sont espacées du double du fond perdu.',
    aufloesung: 'Résolution',
    dpi150: '150 dpi',
    dpi300: '300 dpi',
    dpi600: '600 dpi',
    schnittmarken: 'Traits de coupe',
    legen: 'Composer la feuille d’impression …',
    fertig: 'sur',
    seiten: 'pages :',
    befunde:
      'Tout n’a pas tenu. La feuille existe ; ces endroits sont à vérifier :',
  },
  // The **technical area** (#1164) — commands to copy.
  handeln,
  technik,

  // **Die Recherche** (A16, #1515) — sie steht im Sucharbeitsplatz und
  // im Karteneditor, nicht auf einer eigenen Seite.
  recherche,
  pakete: {
    plan: paketplan,
    bezuege,
    sammeldaten: {
      ueberschrift: 'Données en masse de Scryfall',
      erklaerung:
        'Un fichier contenant toutes les données des cartes. Une fois présent ici, les fonctions Scryfall marchent aussi sans réseau — avec des données vieilles d’un jour au plus.',
      fehlt: 'Pas encore récupérées.',
      vorhanden: 'Présentes, {mb} Mo.',
      angeboten: 'Scryfall propose {mb} Mo, état du {stand}.',
      geholt: 'Récupéré le {zeit}.',
      holen: 'Récupérer les données en masse',
      laedt: 'Récupération en cours …',
    },

    titel: 'Paquets',
    bezugUeberschrift: 'Récupérer un modèle',
    bezugErklaerung:
      'Les modèles MSE se trouvent dans des dépôts. Ce qui est récupéré est un état — une version publiée ou un commit —, jamais une branche : une branche évolue, et la même indication donnerait demain un autre paquet.',
    bezugAdresse: 'Adresse du dépôt',
    bezugStand: 'État (version publiée ou commit)',
    bezugUnterverzeichnis: 'Sous-dossier (racine du paquet)',
    bezugName: 'Nom dans le dossier des paquets',
    bezugPruefen: 'Afficher la provenance …',
    bezugStarten: 'Récupérer',
    bezugAbbrechen: 'Ne pas récupérer',
    bezugGeholt: 'Récupéré dans',
    bezugPruefsumme: 'Somme de contrôle',
    bezugUebergangen:
      'Non décompressés — les fichiers exécutables ne sont pas inclus :',
    bezugUnvollstaendig:
      'L’adresse et l’état sont manquants. Sans les deux, il n’y a rien à récupérer.',
    bezugHerkunft: 'Va récupérer {quelle}, état {stand}.',
    bezugLizenz:
      'Ces contenus ne proviennent pas de MagicSetEditor4Mac et ne sont pas sous notre licence. Les cadres de cartes, symboles et polices peuvent être protégés ; leur utilisation relève de ta responsabilité. Les fichiers exécutables ne sont pas décompressés.',
    mse2Ueberschrift: 'Reprendre depuis une installation de MSE2',
    mse2Erklaerung:
      'Quiconque utilise MSE2 a déjà ses modèles sur son Mac. Les emplacements habituels sont indiqués ci-dessous — aucun n’est lu tant que tu ne l’as pas choisi.',
    mse2Uebernehmen: 'Reprendre d’ici',
    mse2VonHier: 'Reprendre d’ici — {ort}',
    mse2Waehlen: 'Choisir un dossier …',
    mse2KeinVorschlag:
      'Aucune installation de MSE2 trouvée aux emplacements habituels. Choisis toi-même le dossier contenant les modèles — dans MSE2, il s’appelle « data ».',
    installerErklaerung:
      'Ou un fichier .mse-installer — c’est ainsi que MSE2 transmet des modèles. Ce qui est déjà installé reste tel quel.',
    installerWaehlen: 'Lire un installateur …',
    mse2Ergebnis: '{anzahl} modèles repris.',
    mse2Uebergangen: 'Déjà présents et donc ignorés — rien n’a été écrasé :',
    abhaengigUeberschrift: 'Dépendances',
    abhaengigAllesGut: 'Toutes les dépendances sont satisfaites.',
    abhaengigFehlt:
      '{paket} nécessite {braucht} — non installé. Récupère-le ou reprends-le depuis une installation de MSE2.',
    abhaengigZuAlt:
      '{paket} nécessite {braucht} en version {verlangt} ; la version installée est {da}. Sans la plus récente, le modèle dessine silencieusement autre chose.',
    abhaengigZyklus:
      '{paket} et {braucht} dépendent l’un de l’autre en boucle.',
    abhaengigUnlesbar:
      '{paket} nécessite {braucht} : les versions ne peuvent pas être comparées — requise {verlangt}, installée {da}.',
    quellenErklaerung:
      'Quatre collections, à un clic chacune — l’adresse et l’état sont renseignés ; rien n’est récupéré sans ton accord. Les polices des modèles sont incluses ; sans elles, aucune carte ne s’affiche correctement. Aucune des quatre ne porte de licence : les contenus appartiennent à leurs ayants droit, pas à nous.',
    quellenGroesse: 'environ {megabyte} Mo',

    // Die Schlüssel kommen aus `internal/bezug/quellen.go`. Ein Wächter
    // in `internal/werkzeuge` hält beide Seiten zusammen: eine Quelle
    // ohne Satz hier macht die Pipeline rot.
    quellen: {
      basicM15:
        'Les cadres M15 habituels. Le chemin le plus rapide vers une carte dessinée.',
      fullNonMagic: 'Des modèles pour d’autres jeux que Magic.',
      m15: 'Les modèles M15 publics — 324 paquets. Le bon départ pour la plupart des gens.',
      fullMagic:
        'Presque tous les modèles Magic publics — 526 paquets. Cette collection est l’étalon de ce projet.',
    },
    katalogUeberschrift: 'Paquets installés',
    katalogLeer:
      'Aucun paquet installé. Sans jeu ni style, l’aperçu se rabat sur un plan de remplacement.',
    katalogName: 'Nom',
    katalogArt: 'Type',
    katalogVersion: 'Version',
    katalogHandlung: 'Action',
    katalogBefunde: 'Lors de la lecture des dossiers de paquets :',
    anlegen: 'Créer un paquet',
    anlegenArt: 'Type du nouveau paquet',
    anlegenArtSpiel: 'Jeu (.mse-game)',
    anlegenArtStil: 'Style (.mse-style)',
    anlegenArtInclude: 'Include (.mse-include)',
    anlegenArtExport: 'Modèle d’export (.mse-export-template)',
    anlegenName: 'Nom court',
    anlegenNameHinweis:
      'Lettres, chiffres, traits d’union et tirets bas — il devient le nom du dossier',
    anlegenVollname: 'Nom complet',
    anlegenVollnameHinweis: 'Vide signifie : identique au nom court',
    anlegenSpiel: 'Jeu du nouveau paquet',
    anlegenSpielHinweis:
      'Un style met en page les champs de son jeu, un modèle d’export les écrit',
    anlegenSpielLeer: 'Choisis',
    anlegenKnopf: 'Créer',
    anlegenGelungen: 'Créé : {name}',
    loeschen: 'Supprimer',
    loeschenAn: 'Supprimer — {name}',
    handlungenAn: 'Actions — {name}',
    handlungenKurz: '…',
    katalogSuche: 'Rechercher',
    auswahlLoesen: 'Afficher tous les paquets',
    leiste: {
      bezeichnung: 'Chiffres des paquets',
      installiert: 'Installés',
      befund: 'Avec constats',
      nichtGelesen: 'non lu',
      auswahl: '{pakete} avec constats',
    },
    katalogSucheHinweis: 'Nom, type, provenance ou groupe',
    katalogArtAlle: 'Tous les types',
    katalogOrdnung: 'Trier par',
    katalogOrdnungGruppen: 'Groupes',
    katalogOrdnungName: 'Nom',
    katalogOrdnungVersion: 'Version',
    katalogOrdnungHerkunft: 'Provenance',
    katalogGefunden: '{gefunden} paquets sur {gesamt}',
    katalogNichtsGefunden:
      'Aucun paquet ne correspond à cette recherche. Essaie une autre orthographe, ou réinitialise le filtre de type.',
    katalogLeerTitel: 'Aucun paquet pour l’instant',
    katalogNichtsTitel: 'Aucun résultat',
    katalogOhneGruppe: 'Sans groupe',
    katalogGruppeAn: 'Groupe {name}',
    katalogGruppe: 'Groupe',
    katalogOhneVersion: 'sans version',
    katalogZustandFehlt: 'Dépendance manquante',
    katalogZustandZuAlt: 'Dépendance trop ancienne',
    katalogZustandUnlesbar: 'Versions non comparables',
    katalogZustandZyklus: 'Cycle de dépendances',
    katalogEinVorbehalt: 'une réserve',
    katalogVorbehalte: '{anzahl} réserves',
    umschlagTitel: 'Un modèle depuis une enveloppe',
    umschlagErklaerung:
      'Une enveloppe transporte les **fichiers** d’un modèle ; tu choisis le nom avec son extension, par exemple `monjeu.mse-game` ou `monstyle.mse-style`. Le modèle se trouve ensuite chez toi sous ce nom. Si le paquet existe déjà, rien n’est écrasé : tu reçois un refus et choisis un autre nom.',
    umschlagName: 'Comment le paquet doit s’appeler',
    umschlagOhneNamen: 'Indique d’abord comment le paquet doit s’appeler.',
    umschlagAnnehmen: 'Accepter le modèle',
    umschlagAngenommen:
      'Le modèle se trouve maintenant dans ton dossier de paquets.',
  },
  // **Suchen und Ersetzen** (#1168).
  ersetzen: {
    titel: 'Remplacer',
    erklaerung:
      'Remplace du texte dans les champs de texte de ce set. D’abord l’aperçu, puis l’application — et une seule annulation défait toute l’opération.',
    suchtext: 'Rechercher',
    ersatz: 'Remplacer par',
    grossKlein: 'Respecter la casse',
    ganzesWort: 'Mots entiers uniquement',
    vorschauen: 'Afficher ce qui change',
    ersetzen: 'Remplacer maintenant ({treffer})',
    nichtsGefunden: 'Rien trouvé. Essayer un autre mot ?',
    zusammenfassung: 'Occurrences : {treffer} · cartes concernées : {karten}',
    getan: 'Remplacé — occurrences : {treffer} · cartes : {karten}',
    ruecknahme: 'Appuie sur ⌘Z pour annuler toute l’opération.',
    stellen: 'Endroits concernés',
    kartenfeld: 'Carte {nummer} · {feld}',
    weitere:
      'Les autres endroits ne sont pas listés — ils sont remplacés malgré tout.',
  },

  statistik: {
    worueber: 'Compter par',
    belegungWahl: 'Remplissage par champ',
    gruppeSpielpaket: 'Selon le paquet de jeu',
    gruppeFelder: 'Par valeur de champ',
    verteilungVon:
      'Répartition de « {feld} » — part des cartes qui le renseignent',
    nachDimension: 'Statistiques selon « {dimension} »',
    titel: 'Statistiques',
    karten: 'cartes',
    felder: 'champs distincts',
    leere: 'cartes sans aucune valeur',
    belegung: 'Remplissage par champ — part de toutes les cartes du set',
    belegungLeer: 'Ce set ne contient encore aucune valeur de champ.',
    verteilungLeer: 'Ce champ n’a aucune valeur dans ce set.',
    anteil: 'Part',
    dimensionenErklaerung:
      'Un paquet de jeu indique lui-même ce qui peut être compté — catégorie ' +
      'de couleur, valeur de mana, rareté. Chaque dimension calcule sa valeur ' +
      'pour chaque carte et demande donc un instant.',
    dimensionLeer: 'Aucune carte n’a de valeur selon cette dimension.',
    gruppeHinweis:
      'Choisis une barre pour voir ses cartes dans la liste des cartes.',
    gruppeGewaehlt: 'Choisi : « {gruppe} » · cartes : {karten}',
    gruppeZeigen: 'Afficher les cartes',
    gruppeAufheben: 'Annuler le choix',
    gruppeLeer: 'Aucune carte n’a cette valeur.',
    filterBezeichnung: 'Filtre issu des statistiques',
    filterKopf: 'Issu des statistiques : « {quelle} » = « {gruppe} »',
    filterZurStatistik: 'Retour aux statistiques',
    filterAufheben: 'Retirer le filtre',
  },
  // Die Lasche „Style“: Stilpaket und Stiloptionen des Sets (#288).
  stil: {
    titel: 'Style',
    vorschauTitel: 'Aperçu',
    vorschauLeer:
      'Ce set ne contient pas encore de carte — crées-en une pour voir le style.',
    eigeneFarbe: 'couleur personnalisée',
    vorgabe: 'Par défaut',
  },
  suche: {
    titel: 'Recherche',
    eingabe: 'Que cherches-tu ?',
    recherche: 'Recherche …',
    feld: 'Uniquement dans ce champ',
    alleFelder: 'Tous les champs',
    set: 'Uniquement dans ce set',
    alleSets: 'Tous les sets',
    leer: 'Rien n’a encore été recherché. Saisis quelque chose pour trouver des cartes.',
    nichts: 'Rien trouvé. Essayer un autre mot ?',
    gekuerzt:
      'Il y a plus de résultats que ceux affichés ici. Affine la recherche pour les voir.',
    treffer: 'Résultats',
  },
  ausgeben,
  hilfe: {
    titel: 'Aide',
    erklaerung:
      'Ce que signifient les termes utilisés par cette application — et ce qu’il faut faire au premier lancement.',
    glossar: 'Termes',

    // **Die Beschriftung der Laschenleiste** — seit U11 sind es Laschen
    // und keine Sprungmarken mehr; der Name ist geblieben, weil er sagt,
    // was die Leiste tut.
    aufDieserSeite: 'Sur cette page',
    seitenuebersicht: 'Ce que fait chaque page',
  },

  bestandspflege,
  datenpflege,
  kosten,
  ruecksetzung,
  ruecksprung,

  einstellungen,

  nutzer,
} as const
