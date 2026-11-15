// **Die Startseite** — herausgelöst, weil `seiten.ts` mit ihr über 400 Zeilen
// ginge (§2).
//
// **Ein Katalog darf sich auf mehrere Dateien verteilen**, solange er als
// **einer** ankommt — so wie `seitenkarten.ts`. Was gleich bleibt, ist die
// Zahl der Kataloge; an ihr hängt der Vertragstest gegen die Go-Seite (§9).
export const start = {
  ueberblick: {
    titel: 'Твой обзор',
    mitProblemen: 'с проблемами: {anzahl}',
    geprueftVon: 'проверено {geprueft} из {gesamt}',
    mitBefund: 'с замечаниями: {anzahl}',
    nichtGeprueft: 'проверено не полностью',
    problemeGeprueft: 'Проблемы проверены {wann} · {geprueft} из {gesamt} карт',
    problemeOffen: 'Проблемы: не проверено {offen} из {gesamt} наборов',
    pruefen: 'Проверить сейчас',
    setsMitProblemen: '{sets} с проблемами',
    kartenMitBefund: '{karten} с замечаниями в {sets}',
    aeltesterPreis: 'самая старая цена {wann}',
    ohnePreis: 'без цены: {anzahl}',
    mitFehlenden: 'не хватает карт: {anzahl}',
    decksMitFehlenden: '{decks}, в которых не хватает карт',
    alter: {
      gerade: 'только что',
      minuten: '{minuten} мин назад',
      stunden: '{stunden} ч назад',
      tage: '{tage} дн. назад',
      tag: '1 дн. назад',
    },
  },
  // Resume — work item U11 of phase 5.1 (#1422).
  // Attention and background tasks (#1422).
  aufmerksamkeit: {
    titel: 'Требует внимания',
    nichts: 'Ничего — проверено: {sets}, {decks} и {pakete}.',
    durchgesehen: 'Проверено: {sets}, {decks} и {pakete}.',
    arten: {
      setVorbehalte: 'не истолковано: {anzahl}',
      deckFehlend: 'нет в коллекции: {anzahl}',
      paketBefund: 'в скриптах: {anzahl}',
    },
    paketeMitBefund: '{pakete} с замечаниями в скрипте',
    gruendeZeigen: 'Показать причины',
    gruendeVerbergen: 'Скрыть причины',
    setBefunde: '{karten} с замечаниями',
    befunde: 'Не проверено:',
    alle: 'Показать все ({anzahl})',
    weniger: 'Показать меньше',
    gruende: {
      vorbehalt: '{0}, строка {1}',
    },
    amOrt: '{ort}: {satz}',
    // The operations band (OPS02, #2348).
    betrieb: {
      arten: {
        platzKnapp: 'Рядом с библиотекой свободно лишь {frei}',
        sicherungAlt: 'Последняя резервная копия — {tage} дн. назад',
        dienstBeendet: 'Служба при входе завершилась с кодом {ausgang}',
        webprozessNeu: 'Веб-представление перезагружено {anzahl} раз(а)',
      },
      rat: {
        platzKnapp:
          'Если свободно меньше примерно 2 ГБ или 5 %, места для копии мало — освободи его в разделе «Данные и резервные копии».',
        sicherungAlt:
          'Старше 14 дней — создай копию библиотеки в разделе «Данные и резервные копии».',
        dienstBeendet:
          'После ошибки launchd перезапускает её — причина завершения указана в диагностике.',
        webprozessNeu:
          'WebKit завершил свой процесс; несохранённое в открытой форме могло пропасть — проверь это в диагностике.',
      },
    },
  },
  auftraege: {
    titel: 'Фоновые задачи',
    staende: {
      laeuft: 'выполняется',
      fertig: 'готово',
      gescheitert: 'сбой',
      abgebrochen: 'отменено',
    },
    arten: {
      preise: 'Загрузка цен для коллекции',
      sammeldaten: 'Загрузка массива данных Scryfall',
      fassung: 'Загрузка новой версии',
      bogen: 'Размещение листа печати',
      set: 'Чтение набора',
      mtgjson: 'Чтение MTG JSON',
      vorlage: 'Экспорт через шаблон',
      installer: 'Чтение установщика',
      weitergabe: 'Передача в виде установщика',
      probleme: 'Проверить наборы на проблемы',
    },
    abbrechen: 'Отменить: {art}',
    dauer: '{sekunden} с',
    laufend: {
      eine: 'задача выполняется',
      wenige: 'задачи выполняются',
      viele: 'задач выполняется',
    },
  },
  weiterarbeiten: {
    titel: 'Продолжить',
    weiterAn: 'Продолжить работу: {art} «{name}»',
    zuletztWann: 'Последний раз {vorgang}: {zeit}',
    angeheftet: 'закреплено',
    angeheftetLeer:
      'Ничего не закреплено — закрепи строку, чтобы она оставалась здесь.',
    leer: 'Здесь будет твоя недавняя работа. Пока пусто — первый шаг: набор или пакет шаблонов.',
    handlungSets: 'Создать или открыть набор',
    handlungPakete: 'Загрузить пакет',
    anheften: 'Закрепить',
    loesen: 'Открепить',
    anheftenVon: 'Закрепить «{name}»',
    loesenVon: 'Открепить «{name}»',
    arten: { set: 'Набор', deck: 'Колода', paket: 'Пакет' },
    vorgaenge: {
      geoeffnet: 'открыто',
      eingefuehrt: 'импортировано',
      ausgegeben: 'экспортировано',
    },
  },
  titel: 'Главная',
  ueberschrift: 'MagicSetEditor4Mac',
  was: 'Создавай, проверяй и печатай карты для Magic и других игр — с теми же шаблонами, которые использует Magic Set Editor 2.',
  ersteSchritte: 'Первые шаги',
  einfuehrung: {
    willkommen:
      'Три шага — и первая карта готова. Вернуть это окно можно в любой момент в настройках.',
    wiederzeigen: 'Показывать снова при запуске',
    schliessen: 'Начнём',
    bezeichnung: 'Введение',
  },
  schritte: {
    setOeffnen: {
      titel: 'Открыть набор',
      erklaerung:
        'Открой существующий файл .mse-set — или импортируй набор из MTG JSON.',
      handlung: 'К наборам',
    },
    kartenBearbeiten: {
      titel: 'Редактировать карты',
      erklaerung:
        'Просматривай и изменяй карты открытого набора; предпросмотр сразу показывает каждое изменение.',
      handlung: 'К картам',
      ersteinSet: 'Для этого сначала нужен открытый набор.',
      stattdessen: 'Сначала открой набор',
    },
    paketeHolen: {
      titel: 'Получить шаблоны',
      erklaerung:
        'Восемь пакетов шаблонов уже здесь. Нужно больше игр или стилей? Загрузи их здесь.',
      handlung: 'К пакетам',
    },
  },
}
