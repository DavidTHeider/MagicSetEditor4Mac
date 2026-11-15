// The texts of the **Play** area — work item G1 of phase 5.5
// (#1458).
import { abzugworte } from './spielenabzug'
import { zugworte } from './spielenzug'
import { menueworte } from './spielenmenue'
import { vorhangworte } from './spielenvorhang'
import { partienwoerter } from './spielenpartien'
import { regelwoerter } from './spielenregeln'
import { anfangworte } from './spielenanfang'
import { beginnworte } from './spielenbeginn'
import { gastworte } from './spielengast'
import { netzworte } from './spielennetz'
import { leihewoerter } from './spielenleihe'

export const spielen = {
  ...abzugworte,
  ...zugworte,
  ...menueworte,
  ...vorhangworte,
  ...partienwoerter,
  ...regelwoerter,
  ...anfangworte,
  ...beginnworte,
  ...gastworte,
  ...netzworte,
  ...leihewoerter,

  titel: '对战',
  erklaerung:
    '一局对局是一张摆着冻结套牌的牌桌。这里发生的一切，都是一串固定下来的事件——每个事件都挂在前一个事件上，没有任何事件会被覆盖。你移动一张卡牌，就能立刻看到其他人看到的是什么。',
  anfangTitel: '开始一局对局',
  anfangErklaerung:
    '每个座位带一副套牌，并且它会被冻结：之后再修改你的套牌，不会改变这局对局。游戏取自第一副所选的套牌。',
  leerDecksTitel: '还没有套牌',
  leerDecks: '对战需要一副套牌。请在“套牌”下创建一副。',
  sitz: '座位',
  sitzName: '牌桌上的名字',
  sitzDeck: '套牌',
  sitzLeihe: '借自 …{von} · {karten} 张',
  deckWaehlen: '选择套牌',
  anschlussWartet: '“{name}”已一并带来，坐在第一个位置。第二个位置空着。',
  sitzHinzu: '添加座位',
  sitzWeg: '移除此座位',
  sitzGruppe: '座位 {nummer}',
  spiel: '游戏',
  anfangen: '开始对局',
  tischTitel: '牌桌上',
  partieGegen: '对阵 {namen} 的对局',
  partieZwischen: '{namen} 之间的对局',
  partieVon: '单人对局：{namen}',
  fassung: '修订版',
  wirt: '主持方',
  wirtHinweis:
    '主持方持有这局对局的完整状态，因此看到的比它自己的视角所显示的更多。这里写明，是因为这需要在加入之前知道，而不是之后。',
  tischAufHaufen: '{zahl} 张叠成一堆',
  tischMarkeEine: '{zahl} 个 {art} 指示物',
  tischMarkeViele: '{zahl} 个 {art} 指示物',
  tischArt: {
    spielstein: '衍生物',
    emblem: '徽记',
    faehigkeit: '堆叠上的异能',
  },

  sicht: '视角属于',
  sichtZuschauer: '旁观者（无座位）',
  sichtWechseln: '切换座位',
  sichtLaedt: '正在加载所选视角……',
  verlassen: '关闭对局',
  zonen: '什么在哪里',
  zoneBibliothek: '牌库',
  zoneHand: '手牌',
  zoneTisch: '战场',
  zoneFriedhof: '坟场',
  zoneExil: '放逐区',
  zoneStapel: '堆叠',
  zoneKommando: '指挥官区域',
  zonekurzBibliothek: '牌库',
  zonekurzFriedhof: '坟场',
  zonekurzExil: '放逐',
  zonekurzKommando: '指挥区',
  zoneLeer: '空',

  unbekannterTyp: '未知类别：{woerter}',
  geheim: '面朝下',
  getappt: '已横置',
  befehle: '操作',
  regelnTitel: '强制执行的内容',
  regelwerk: '规则集',
  wirksamAb: '生效自',
  hoechsteStrenge: '最高严格程度',
  strenge: '严格程度',
  andereSeite: '另一面',
  grossAnsehen: '放大查看',
  ersteSeite: '第一面',
  stelle: '放到位置',
  stelleOben: '顶部',
  stelleUnten: '底部',
  stelleZufaellig: '随机插入',
  mischpflichtTitel: '待处理的洗牌义务',

  legendeTitel: '待做选择：传奇规则',
  legendeSatz: '{wer} 必须选择保留哪个名为“{name}”的传奇永久物（CR 704.5j）。',
  legendeBehalten: '保留这个',

  reihungTitel: '待做选择：异能的顺序',
  reihungSatz:
    '{wer} 操控着多个同时触发的异能，并决定它们放上堆叠的顺序（CR 603.3b）。最先放上的会位于最下方，最后结算。',
  reihungLegen: '下一个放置：',
  stand: '状态',
  standLaeuft: '进行中',
  standPausiert: '已暂停',
  standBeendet: '已结束',
  pausieren: '暂停',
  fortsetzen: '继续',
  beenden: '结束',
  beendenFrage: '要结束这局对局吗？结束是最终的——已结束的对局无法继续。',
  beendenJa: '结束',
  beendenNein: '取消',
  ereignisse: '刚刚发生的事',
  ereignisseLeer: '暂无。',
  ereignisseLuecke:
    '此处可能缺少部分走法：页面隐藏期间收到的消息超出了可补回的数量。',
  ereignisPrivat: '仅你可见',
  ereignisSatz: '{sitz}：{handlung} — {objekt}',
  ereignisSatzOhneKarte: '{sitz}：{handlung}',

  // **Die Rückschau** (G5, #1462).
  verlauf: '至今经过',
  verlaufAnsehen: '显示历史',
  verlaufNachladen: '重新加载',
  verlaufLaeuft: '正在获取……',
  verlaufUngeholt:
    '尚未获取历史。它会显示这局对局的每一步操作，范围为此座位可以看到的部分。',
  verlaufLeer: '这局对局中还没有发生任何事。',
  verlaufGrenze: '其他座位仅自己看到的内容不在此处——事后也不会出现。',

  // **Was von einer Partie übrig bleibt** (G5, #1462).
  protokoll: '这局对局留下什么',
  protokollErklaerung:
    '这局对局的每一步操作都存放在这台电脑上，每个座位只能看到属于自己的部分。你可以把自己的部分公开给以后查看——公开给某一位对手，或公开给所有人。',
  protokollAnsehen: '查看公开情况',
  protokollNiemand: '没有人公开任何内容。',
  protokollFuer: '公开给',
  protokollAlle: '所有人',
  protokollSitz: '座位',
  protokollEmpfaenger: '接收者，用逗号分隔——或用 * 表示所有人',
  protokollFreigeben: '公开',
  protokollZuruecknehmen: '撤回',
  protokollNurSpaeter:
    '公开只在对局结束后生效。在牌桌上，每个人看到的仍然只是属于自己的部分。',
  protokollKeineRueckholung: '撤回只是关上门——它无法收回别人已经读到的内容。',
  protokollAufbewahrung:
    '已结束的记录在 90 天后会出现在数据整理中，等待你决定。只有你说要清除时才会清除。',
  protokollLoeschen: '删除记录',
  protokollLoeschenErklaerung:
    '这会把这局对局的状态、历史和公开设置从这台电脑上移除。这并不涉及别处的副本。',
  protokollGeloescht: '记录已删除。',
  protokollLoeschenFrage: '要删除这局对局的记录吗？',
  protokollLoeschenJa: '删除记录',

  // **Die Diagnose der Leitung** (G5, #1462).
  diagnose: '连接状态',
  diagnoseErklaerung:
    '这里显示此牌桌通过什么通信，以及有多少内容在等待。其中没有任何内容涉及卡牌——无论是你的还是别人的。',
  diagnoseAnsehen: '检查连接',
  diagnoseStill: '当前没有牌桌在运行。',
  diagnoseAlsWirt: '这台电脑主持牌桌。',
  diagnoseAlsGast: '这台电脑坐在别人的牌桌上。',
  diagnosePausiert: '连接已断开——座位仍然保留。',
  diagnoseWeg: '路径',
  diagnoseGegenueber: '对等端',
  diagnoseSeit: '已持续',
  diagnoseWartend: '等待中',
  diagnoseVerworfen: '已丢弃',
  diagnoseStand: '序列位置',
  diagnoseRunde: '最近一次往返',
  diagnoseBildzeit: '此窗口中的帧间隔',
  diagnoseUnbekannt: '未知',
  diagnoseOhneLeitung: '没有建立任何连接。',

  // **Die Freigabe für Gäste** (G5, #1462).
  freigabe: '允许其他人加入这张牌桌',
  freigabeErklaerung:
    '可被访问和允许他人参与是两个决定。打开牌桌，就是允许对方在这局对局中行动——座位由你分配，而不是由对方分配。',
  freigabeOffen: '这张牌桌已开放。',
  freigabeZu: '这张牌桌未开放。',
  freigabeStarten: '开放牌桌',
  freigabeBeenden: '关闭牌桌',
  freigabeNachsehen: '重新检查',
  freigabeSitze: '空闲给',
  freigabeNiemand: '还没有人入座。',
  freigabePartner: '这是',
  freigabePartnerFremd: '这位对方不在你的联系人中。',
  freigabeBesetzt: '已入座：',
  freigabeOhneHorchposten:
    '要让对方能够到达，这台电脑还必须可被访问——这在“设置”中设定。',

  // **Der gezeichnete Tisch** — Werkstück G3 von Phase 5.5
  // (#1460).
  tischBild: '牌桌',
  tischListe: '以列表显示',
  tischGrafisch: '以牌桌显示',
  tischDarstellung: '显示方式',
  tischEinpassen: '显示全部',
  tischNaeher: '放大',
  tischWeiter: '缩小',
  tischDrehen: '旋转牌桌',
  tischAuswahl: '前往所选项',
  tischEigenerSitz: '前往我的座位',
  tischBlick: '桌面视图',
  tischZiel: '跳转到',
  tischOhneWebgl:
    '这台电脑不提供加速图形。牌桌会以列表形式显示在下方，绘制牌桌上能做的一切，在列表中也都能做。',
  tischOhneRenderer:
    '无法绘制牌桌。下方的列表包含每张卡牌，其余一切也都可以在那里进行。',
  tischVerloren:
    '图形连接已断开。重新打开窗口后牌桌会恢复；在此之前，下方的列表包含每张卡牌。',
  tischHeimatlos: '这些对象位于本程序不认识的区域，因此不会出现在牌桌上：',
  tischVerdeckt: '面朝下',
  tischGetappt: '已横置',
  tischHaengtAn: '结附于',
  tischbild: '背景',
  tischbildHolz: '木纹',
  tischbildDunklesHolz: '深色木纹',
  tischbildPicknick: '野餐',
  tischbildDungeon: '地城',
  tischbildTaverne: '酒馆',
  tischbildTurniertisch: '比赛桌',
  tischEigenes: '自定图像……',
  tischEigenesWeg: '移除自定图像',
  gespraech: '牌桌交流',
  gespraechVerlauf: '牌桌上说过的话',
  gespraechEingabe: '说点什么',
  gespraechSenden: '发言',
  gespraechLeer: '还没有人说话。',
  gespraechNiemandZeigt: '目前没有人在指向任何东西。',
  gespraechZeigt: '{sitz}：{tut}',
  gespraechZeigtAuf: '{sitz}：{tut} {objekt}',
  zeigerZeiger: '指向',
  zeigerGreift: '拖动',
  zeigerSchwebt: '查看',
  zeigerPing: '呼叫',
  spielart: '模式',
  spielartGoldfish: '独自练习',
  spielartSelfplay: '多个座位',
  format: '赛制',
  fehltDeck: '请为每个座位选择一副套牌——对局的游戏取决于第一副。',
  fehltEigenesDeck:
    '至少一个座位需要自己的套牌——借来的套牌不带游戏，游戏取自第一副自己的套牌。',
  fehltName: '每个座位都需要一个名字；在牌桌上，它用来区分玩家。',

  // **Die digitale Deckleihe** (G6, #1463).
  leihe: '出借或借用套牌',
  leiheOeffnen: '出借或借用 …',
  leiheVerleihenTitel: '出借套牌',
  leiheAnnehmenTitel: '接受一笔借用',
  leiheListeTitel: '你的借用',
  leiheErklaerung:
    '出借，就是允许对手使用你套牌的一个冻结快照进行对战。套牌仍留在这里，你收藏中的任何一张卡牌都不会改变。',
  leiheKeinEigentum:
    '这是一项使用授权，而不是转让。无论是数字上还是实物上，都没有任何东西易主。',
  leiheKeinKopierschutz:
    '而且这不是复制保护：看过卡牌的人，就是看过了。借用所规定的，是本程序会用什么来开始一局对局。',
  leiheAnsehen: '查看借用',
  leiheKeine: '没有借用记录。',
  leiheEntleiher: '出借给谁？',
  leiheEntleiherLeer:
    '还没有已确认的对方。请在下方的“添加或确认对方”框中添加一位。',
  leiheDeck: '哪副套牌',
  leiheDeckWaehlen: '选择一副套牌……',
  leiheKeinDeck: '还没有套牌——请在“套牌”下创建。',
  leiheFachleute: '仅为某一局特定的对局出借',
  leiheFachleuteErklaerung:
    '不填写时，借用适用于对方的下一局对局。若要绑定到恰好某一局，请填入该局的 ID——对方可以在其对局列表中找到。',
  leiheWoran: '对局 ID——留空表示下一局',
  leiheStunden: '有效多少小时',
  leiheZiel: '信封存放到哪里',
  leiheAnbieten: '出借并保存',
  leiheAbgelegt: '信封在这里：',
  leiheUmschlag: '借用的信封',
  leiheAnnehmen: '接受借用',
  leiheWiderrufen: '撤销',
  leiheZurueckgeben: '归还',
  leiheVerliehen: '出借给',
  leiheGeliehen: '借自',
  leiheBis: '有效至',
  leiheKarten: '张卡牌',
  leiheVerbraucht: '已兑现于',
  leiheOffen: '尚未兑现',
  leiheZurueckgenommen: '已撤销',
  leiheModusZuende: '正在进行的对局会打完。',
  leiheModusAbbruch: '正在进行的对局会被中止。',
  leiheRuecknahmeGrenze:
    '撤销会阻止新的开局。它不会从正在进行的对局中移除任何内容，也不会删除对方那里的任何副本。',

  // **Wie die Partie ausgegangen ist** (M17, #1618).
  ausgangTitel: '结果',
  ausgangSieger: '获胜者是',
  ausgangUnentschieden: '平局——所有仍在对局中的人同时落败。',
  grundAufgabe: '已认输',
  standLeiste: '玩家状况',
  standDu: '你',
  standLeben: '生命',
  standGift: '中毒',
  standMana: '法术力',
  standStrahlung: '辐射',
  standSteuer: '税',
  standCommanderschaden: '{name}：{schaden}',
  standCommanderTod: '达到 21 点，对局结束',
  standLeergezogen: '从空牌库抓牌',
  grundLeben: '生命为零或以下',
  grundBibliothek: '需要从空牌库抓牌',
  grundGift: '拥有十个中毒指示物',
  aufgeben: '认输',
  aufgebenFrage: '认输吗？该座位的对局会立即结束，其牌也随之离场。',
  aufgebenJa: '确认认输',
  aufgebenNein: '继续对局',
  aufgebenHinweis:
    '你可以随时认输，且立即生效。你会离开对局，你的卡牌也随你离开。',

  // **Der Weg zurück in die eigene Partie** (#1600).

  // **Ein Sitz an einem Profil** (G8, #1465).
  sitzprofil: '配置文件',
  sitzprofilKeines: '无配置文件',
  sitzprofilErklaerung:
    '如果你有两个配置文件，可以给每个座位指定一个。对局随后会从属于你当前配置文件的座位视角开始；其他视角可通过牌桌上的视角选择器切换。另一个视角所属的内容此后不再显示。如果没有给任何座位指定配置文件，你当前的配置文件将坐在第一个座位。',
  sitzprofilBefunde: '相关提示：',
}
