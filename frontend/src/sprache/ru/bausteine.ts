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
      'Тем временем здесь стоит «{wert}» — изменено в другом месте. Твой черновик ещё не сохранён.',
    behalten: 'Сохранить мой черновик',
    nehmen: 'Взять новое значение',
  },
  listenansicht: {
    suche: 'Поиск',
    suchePlatzhalter: 'Название или игра',
    sortierung: 'Порядок',
    ansicht: 'Вид',
    liste: 'Список',
    kacheln: 'Плитки',
    zahl: '{gezeigt} из {gesamt}',
    nichtsPasst: 'Ничего не подходит под «{suche}».',
    sucheLeeren: 'Очистить поиск',
    filterZuruecksetzen: 'Сбросить поиск и фильтры',
    weitere: 'Показать ещё {anzahl}',
    wieImBestand: 'Как сохранено',
    zuletztBenutzt: 'Сначала недавно использованные',
    nachName: 'По названию',
    nachKarten: 'По числу карт',
  },
  speicherstand: {
    schreibt: 'Сохранение …',
    geschrieben: 'Каждое изменение записано в хранилище.',
    fehler: '«{feld}» не сохранено — причина указана у поля.',
  },
  gegenueberTitel: 'Добавить или подтвердить партнёра',
  kontaktwahlEintrag: '{name} · {fingerabdruck}',
  kontaktwahlSuchen: 'Найти партнёра …',
  feldhilfe: 'Пояснение: {feld}',
  mehrZu: 'Подробнее о «{satz}»',

  // **Where something came from** — Werkstück 12 of Phase 4.12
  // (#1351).
  herkunftAusProfil: 'Взято из профиля {quelle}.',
  herkunftAusEinfuhr: 'Импортировано из {quelle}.',

  pflichtfeldLeer: '{feld}: нужно указать значение.',

  waehrung: {
    beispiel: 'напр., EUR',
    keineGestalt:
      '«{wert}» — не код валюты: он состоит из трёх букв, например EUR.',
    grossSchreiben:
      '«{wert}» — код валюты пишется заглавными буквами: {gross}.',
  },

  drucksprachen: {
    beispiel: 'напр., en',
    namen: {
      en: 'Английский',
      es: 'Испанский',
      fr: 'Французский',
      de: 'Немецкий',
      it: 'Итальянский',
      pt: 'Португальский',
      ja: 'Японский',
      ko: 'Корейский',
      ru: 'Русский',
      zhs: 'Китайский (упрощённый)',
      zht: 'Китайский (традиционный)',
      he: 'Иврит',
      la: 'Латинский',
      grc: 'Древнегреческий',
      ar: 'Арабский',
      sa: 'Санскрит',
      ph: 'Фирексийский',
      qya: 'Квенья',
      dw: 'Дварфийский',
      tlh: 'Клингонский',
    },
  },

  // **The password field** — piece 6 of phase 4.12 (#1351).
  kennwort: {
    zeigen: 'Показать пароль открытым текстом',
    zeigenKurz: 'Показать',
    verbergen: 'Снова скрыть пароль',
    verbergenKurz: 'Скрыть',
  },

  // **The field for the second factor** — piece 7 of phase 4.12 (#1351).
  nachweis: {
    beschriftung: 'Код',
    abschicken: 'Подтвердить',
  },

  aktionen: {
    abbrechen: 'Отмена',
    oeffnen: 'Открыть',
    schliessen: 'Закрыть',
    speichern: 'Сохранить',
  },
  stilliste: {
    mitZusatz: '{kennung} · {zusatz}',
    keinerGewaehlt: '— Выбери стиль —',
  },
  druckwahl: {
    drucke: 'Печати',
    hinweis:
      'Этот пир увидит только эти печати — в том количестве, которое у тебя есть.',
    suchen: 'Найти печати',
    gewaehlt: 'Выбрано: {anzahl}',
    nurSuche: 'В твоей коллекции: {anzahl}. Введи название, чтобы выбрать.',
    keiner: 'Нет печати, подходящей под «{suche}».',
    genauer: 'Подходит больше печатей, чем показано. Уточни поиск.',
    leer: 'В твоей коллекции пока нет печатей, которыми можно поделиться. Сначала добавь экземпляры.',
    fehlt: 'Выбери хотя бы одну печать: выбор без печатей ничего не открывает.',
    behalten: 'Экземпляров каждой печати, которые я оставляю',
    behaltenHinweis:
      'Пир увидит то, что сверх этого числа: при 1 и трёх экземплярах — два.',
    detailBehalten: 'ты оставляешь {behalten}',
  },
  suchauswahl: {
    keiner: 'Нет совпадений с «{suche}».',
    treffer: 'Найдено: {anzahl}',
    begrenzt: 'Показано {gezeigt} из {anzahl} — уточни поиск …',
  },
  zeitzonenwahl: {
    system: 'Как в системе',
  },
  aktionskatalog: {
    aktionen: {
      anlegen: { kurz: 'Создать', heisst: 'Создаёт что-то новое.' },
      bearbeiten: { kurz: 'Изменить', heisst: 'Изменяет на месте.' },
      sichern: { kurz: 'Сохранить', heisst: 'Фиксирует текущее состояние.' },
      zuruecknehmen: {
        kurz: 'Отменить',
        heisst: 'Отменяет последний шаг.',
      },
      duplizieren: {
        kurz: 'Дублировать',
        heisst: 'Создаёт вторую копию рядом.',
      },
      kopieren: { kurz: 'Копировать', heisst: 'Помещает в буфер обмена.' },
      entfernen: {
        kurz: 'Убрать',
        heisst: 'Убирает из этого представления; в других местах остаётся.',
      },
      loeschen: {
        kurz: 'Удалить',
        heisst: 'Удаляет безвозвратно и везде.',
      },
    },
  },
  nurlesen: {
    hinweis:
      'Ты просматриваешь этот профиль как читатель: можно всё смотреть и ничего не менять. Перейди в свой профиль, чтобы работать, или попроси того, кто управляет этим профилем, назначить тебе роль «Соавтор».',
    knopf: 'Только чтение — в этом профиле ничего нельзя изменить.',
  },

  // **The context strip** — piece U4 of phase 5.1 (#1415).
  kontext: {
    bezeichnung: 'Рабочий контекст',
    profilVon: 'владелец: {eigner}',
    rollen: {
      leser: 'Читатель',
      mitarbeiter: 'Соавтор',
      verwalter: 'Администратор',
    },
    nurLesen: 'только чтение',
    zuDenNutzern: 'Сменить профиль …',
    zeile: '{name} · {profil}',
  },

  zustaende: {
    laedt: 'Загрузка …',
    leer: 'Здесь пусто',
    fehler: 'Не удалось загрузить.',
    fehlerMitGrund: 'Не удалось загрузить. {grund}',
    wiederholen: 'Загрузить снова',
    neuLaden: 'Перезагрузить',
  },
  // Find and replace **inside the field** (#1164).
  suche: {
    oeffnen: 'Найти в тексте',
    schliessen: 'Закрыть поиск',
    suchen: 'Найти',
    ersetzen: 'Заменить на',
    treffer: '{bei} из {ganz}',
    keinTreffer: 'Совпадений нет',
    weiter: 'Следующее совпадение',
    zurueck: 'Предыдущее совпадение',
    einsetzen: 'Заменить',
    alleEinsetzen: 'Заменить все',
    grossKlein:
      'Регистр учитывается: в MSE-Script «card» и «Card» — это разные вещи.',
  },
  // Completion **inside** the field (#1164).
  // **Welche Karten hinausgehen** (#1167).
  ausgabeauswahl: {
    aufschrift: 'Какие карты',
    ganzesSet: 'Весь сет (карт: {zahl})',
    ganzesSetOhneZahl: 'Весь сет',
    teilmenge: 'Только выбранные ({zahl})',
    ohneNamen: 'без названия',
    eingrenzen: 'Отобрать карты',
  },

  vorschlaege: {
    bezeichnung: 'Предложения',
    uebernehmen: 'Принять клавишей ↹',
  },

  // **The regions of a workspace** — piece U2 of phase 5.1 (#1413).
  arbeitsplatz: {
    breiteVon: 'Ширина «{region}»',
    breiteWert: '{punkte} пт',
    navigatorAuf: 'Показать навигатор',
    navigatorZu: 'Скрыть навигатор',
    inspectorAuf: 'Показать инспектор',
    inspectorZu: 'Скрыть инспектор',
    navigatorEng:
      'Слишком узко для всех трёх колонок — навигатор открывается поверх центральной области',
    inspectorEng:
      'Слишком узко для всех трёх колонок — инспектор открывается поверх центральной области',
  },
  // The command centre (#1423).
  druck: {
    vorlaeufig: 'предварительно',
    vorlaeufigErklaerung:
      'У этой печати пока нет идентификатора Scryfall. Цены и изображения не показываются, пока печать не будет определена.',
  },
  kommandozentrale: {
    titel: 'Командный центр',
    erklaerung:
      'Команда — это инструмент сервера MCP, вызываемый здесь, в окне; ответ тот же, что и через MCP или в командной строке. Ничто не проходит через оболочку: то, что ты вводишь, остаётся данными.',
    lesend: 'читает',
    schreibend: 'изменяет',
    gefaehrlich: 'нелегко отменить',
    faehigkeit: 'Возможность',
    cli: 'Командная строка',
    rest: 'Сервис',
    parameter: 'Параметры',
    keineParameter: 'У этой команды нет параметров.',
    pflicht: 'Обязательно',
    jsonHinweis:
      'В формате JSON: список — в квадратных скобках, объект — в фигурных.',
    befundPflicht: 'В это поле нужно ввести значение.',
    befundZahl: 'Это не число.',
    befundJson: 'Это не JSON.',
    vorschau: 'Что будет отправлено',
    bestaetigen: 'Предпросмотр прочитан, выполнить как есть.',
    ausfuehren: 'Выполнить',
    abbrechen: 'Отмена',
    laeuft: 'Команда выполняется …',
    ergebnis: 'Результат',
    ausgabe: 'Вывод команды',
    alsText: 'Текст',
    alsJson: 'JSON',
    dauer: '{ms} мс',
    verlaufLeer:
      'Пока ничего не выполнено. Секреты — токены, пароли — никогда не показываются здесь открытым текстом.',
    verlauf: 'История этого сеанса',
    wiederholen: 'Загрузить {name} снова',
    gelungen: 'выполнено',
    gescheitert: 'не удалось',
    abgebrochen: 'отменено',
    schliessen: 'Закрыть',
  },
  // „Diagnose öffnen“ an einem Fehler mit Vorgangskennung (NET-02,
  // S13.NAT-04, #1706).
  diagnoseverweis: {
    oeffnen: 'Открыть диагностику',
  },
  // „KI einrichten“ an einer KI-Meldung (AI-01.AK-04, #1742).
  kieinrichtung: {
    oeffnen: 'Настроить ИИ',
  },
  kikasten: {
    titel: 'ИИ для этой карты',
  },
} as const
