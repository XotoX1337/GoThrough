(() => {
  // frontend/src/app.ts
  var App = window.go.overlay.App;
  var STORE_KEY = "gt-overlay-v7";
  var saved = (() => {
    try {
      return JSON.parse(localStorage.getItem(STORE_KEY) || "{}");
    } catch {
      return {};
    }
  })();
  var DESKTOP_H = window.screen && window.screen.availHeight || 920;
  var NOWCARD_H = Math.round(DESKTOP_H * 0.382);
  var DEFAULT_W = 380;
  var DEFAULT_H = Math.min(DESKTOP_H, NOWCARD_H + 300);
  var state = {
    view: "picker",
    // "picker" | "steps" | "settings"
    // --- picker ---
    configs: [],
    // []configstore.Entry catalog from App.ListConfigs()
    pickerGame: null,
    // selected game in the two-level picker (null = game list)
    pickerLoading: "",
    // game currently being downloaded ("" = none)
    pickerError: "",
    // --- steps ---
    meta: { game: "", title: "" },
    steps: [],
    current: null,
    // live StepInfo of the active item (now-card source of truth)
    activePos: 1,
    nextFile: "",
    // `next:` reference for end-of-walkthrough hand-off
    sectionOverride: {},
    // section name → user's explicit open(true)/closed(false)
    lastScrolledPos: null,
    // position auto-scroll last targeted (only scroll on change)
    // --- settings ---
    settings: null,
    // {version, hotkeys, opacity, theme} from App.Settings()
    capturing: null,
    confirm: null,
    // {message, confirmLabel, onConfirm} for the modal (null = closed)
    settingsError: "",
    settingsFrom: "picker",
    // view to return to when settings is closed ("picker" | "steps")
    // --- layout ---
    collapsed: false,
    locked: true,
    // always start fixed; never persisted
    width: saved.width || DEFAULT_W,
    panelHeight: saved.panelHeight || DEFAULT_H
    // fixed window height when expanded
  };
  var MIN_W = 240;
  var MAX_W = 680;
  var MIN_H = 240;
  var MAX_H = DESKTOP_H;
  var mover = document.getElementById("mover");
  var panel = document.getElementById("panel");
  var resizeHandle = document.getElementById("resize");
  var modalHost = document.getElementById("modal");
  mover.style.setProperty("--gt-nowh", NOWCARD_H + "px");
  var CHEVRON_SVG = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="6 9 12 15 18 9"></polyline></svg>';
  var HOME_SVG = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"></path><polyline points="9 22 9 12 15 12 15 22"></polyline></svg>';
  var GEAR_SVG = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="3"></circle><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"></path></svg>';
  function save() {
    try {
      localStorage.setItem(STORE_KEY, JSON.stringify({
        width: state.width,
        panelHeight: state.panelHeight,
        // Cached so the theme can be applied synchronously on next launch,
        // before the async Settings() round-trip — avoids a wrong-colour
        // flash of the themed UI (e.g. the progress bar) on restart.
        theme: state.settings ? state.settings.theme : saved.theme
      }));
    } catch {
    }
  }
  function esc(s) {
    return String(s ?? "").replace(
      /[&<>"]/g,
      (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;" })[c]
    );
  }
  function mdInline(s) {
    return esc(s).replace(/\*\*([^*]+)\*\*/g, "<strong>$1</strong>").replace(/\*([^*]+)\*/g, "<em>$1</em>");
  }
  function md(s) {
    const lines = String(s ?? "").split("\n");
    const out = [];
    let inList = false;
    for (const raw of lines) {
      const m = raw.match(/^\s*-\s+(.*)$/);
      if (m) {
        if (!inList) {
          out.push("<ul>");
          inList = true;
        }
        out.push("<li>" + mdInline(m[1]) + "</li>");
      } else {
        if (inList) {
          out.push("</ul>");
          inList = false;
        }
        if (raw.trim()) out.push("<p>" + mdInline(raw) + "</p>");
      }
    }
    if (inList) out.push("</ul>");
    return out.join("");
  }
  var I18N = {
    en: {
      adjust: "Adjust",
      lock: "Lock",
      close: "Close",
      openFile: "Open file…",
      chooseGame: "Choose game",
      backToGames: "← Games",
      chapterLabel: (n) => "Chapter " + n,
      chaptersCount: (n) => n + (n === 1 ? " chapter" : " chapters"),
      by: (a) => "by " + a,
      noChapters: "No chapters found.",
      noConfigs: "No configs found.",
      downloading: (g) => "Downloading " + g + "…",
      clearProgress: "Reset progress",
      clearCache: "Clear cache",
      resetSettings: "Reset to defaults",
      confirmClearCache: "Delete all downloaded configs from the cache?",
      confirmClearGame: (g) => "Reset all saved progress for " + g + "?",
      confirmTitle: "Are you sure?",
      cancel: "Cancel",
      confirmDelete: "Delete",
      confirmReset: "Reset",
      switchWalkthrough: "Switch walkthrough",
      settingsTitle: "Settings",
      closeOverlay: "Close overlay",
      steps: "Steps",
      noSteps: "No steps loaded",
      now: "Now",
      decision: "Decision",
      stepCount: (a, b) => "Step " + a + " / " + b,
      chosen: "chosen",
      nextFile: (f) => "Continue ➜ " + f,
      optional: "optional",
      quests: "Quests",
      completed: "✓ completed",
      received: "received",
      warning: "Caution",
      info: "Info",
      hints: "Hints",
      shortcuts: "Shortcuts",
      pressKey: "Press key/mouse…",
      opacity: "Opacity",
      design: "Theme",
      language: "Language",
      settingsHint: "Click a shortcut, then press a key or mouse combo. Esc cancels.",
      hkNext: "Next step",
      hkPrev: "Previous step",
      hkToggle: "Toggle overlay",
      hkFocus: "Mouse into overlay",
      hkQuit: "Close overlay",
      themeDark: "Dark",
      themeLight: "Light",
      themeContrast: "Contrast",
      mouseL: "Mouse L",
      mouseM: "Mouse M",
      mouseR: "Mouse R",
      mouse4: "Mouse 4",
      mouse5: "Mouse 5"
    },
    de: {
      adjust: "Anpassen",
      lock: "Fixieren",
      close: "Schließen",
      openFile: "Datei öffnen…",
      chooseGame: "Spiel wählen",
      backToGames: "← Spiele",
      chapterLabel: (n) => "Kapitel " + n,
      chaptersCount: (n) => n + (n === 1 ? " Kapitel" : " Kapitel"),
      by: (a) => "von " + a,
      noChapters: "Keine Kapitel gefunden.",
      noConfigs: "Keine Configs gefunden.",
      downloading: (g) => "Lade " + g + "…",
      clearProgress: "Fortschritt zurücksetzen",
      clearCache: "Cache leeren",
      resetSettings: "Auf Standard zurücksetzen",
      confirmClearCache: "Alle heruntergeladenen Configs aus dem Cache löschen?",
      confirmClearGame: (g) => "Allen gespeicherten Fortschritt für " + g + " zurücksetzen?",
      confirmTitle: "Bist du sicher?",
      cancel: "Abbrechen",
      confirmDelete: "Löschen",
      confirmReset: "Zurücksetzen",
      switchWalkthrough: "Anderes Walkthrough",
      settingsTitle: "Einstellungen",
      closeOverlay: "Overlay schließen",
      steps: "Schritte",
      noSteps: "Keine Schritte geladen",
      now: "Jetzt",
      decision: "Entscheidung",
      stepCount: (a, b) => "Schritt " + a + " / " + b,
      chosen: "gewählt",
      nextFile: (f) => "Weiter ➜ " + f,
      optional: "optional",
      quests: "Quests",
      completed: "✓ abgeschlossen",
      received: "erhalten",
      warning: "Achtung",
      info: "Info",
      hints: "Hinweise",
      shortcuts: "Tastenkürzel",
      pressKey: "Taste/Maus drücken…",
      opacity: "Transparenz",
      design: "Design",
      language: "Sprache",
      settingsHint: "Klick auf ein Kürzel, dann Tasten- oder Maustasten-Kombination drücken. Esc bricht ab.",
      hkNext: "Nächster Schritt",
      hkPrev: "Vorheriger Schritt",
      hkToggle: "Overlay ein/aus",
      hkFocus: "Maus ins Overlay",
      hkQuit: "Overlay schließen",
      themeDark: "Dunkel",
      themeLight: "Hell",
      themeContrast: "Kontrast",
      mouseL: "Maus L",
      mouseM: "Maus M",
      mouseR: "Maus R",
      mouse4: "Maus 4",
      mouse5: "Maus 5"
    }
  };
  var LANGUAGES = [
    { key: "en", label: "English" },
    { key: "de", label: "Deutsch" }
  ];
  function lang() {
    const l = state.settings && state.settings.language;
    return I18N[l] ? l : "en";
  }
  function t(key) {
    const v = I18N[lang()][key];
    return v != null ? v : I18N.en[key] != null ? I18N.en[key] : key;
  }
  function tf(key, ...args) {
    const v = t(key);
    return typeof v === "function" ? v(...args) : v;
  }
  function renderQuests(quests) {
    if (!quests || !quests.length) return "";
    const rows = quests.map((q) => {
      const status = q.status === "completed" ? `<span class="quest-status done">${t("completed")}</span>` : q.status === "received" ? `<span class="quest-status">${t("received")}</span>` : "";
      const note = q.note ? `<div class="quest-note">${mdInline(q.note)}</div>` : "";
      return `<div class="quest ${esc(q.status || "")}">
                    <div class="quest-line"><span class="quest-name">${esc(q.name)}</span> ${status}</div>${note}
                </div>`;
    }).join("");
    return `<div class="info-block"><div class="info-head mono">${t("quests")}</div>${rows}</div>`;
  }
  function renderList(items, head, cls) {
    if (!items || !items.length) return "";
    const lis = items.map((t2) => `<li>${mdInline(t2)}</li>`).join("");
    return `<div class="info-block ${cls || ""}"><div class="info-head mono">${head}</div><ul class="info-list">${lis}</ul></div>`;
  }
  function taskCallout(kind, text) {
    if (!text) return "";
    const tag = kind === "hint" ? "" : `<span class="task-callout-tag mono">${t(kind)}</span>`;
    return `<span class="task-callout ${kind}">${tag}${mdInline(text)}</span>`;
  }
  function renderTasks(tasks) {
    if (!tasks || !tasks.length) return "";
    const items = tasks.map((t2) => `<li class="task">${mdInline(t2.text)}${taskCallout("warning", t2.warning)}${taskCallout("info", t2.info)}${taskCallout("hint", t2.hint)}</li>`).join("");
    return `<ul class="task-list">${items}</ul>`;
  }
  function applyOpacity(v) {
    mover.style.setProperty("--gt-opacity", v != null && v > 0 ? v : 1);
  }
  var THEMES = [
    { key: "dark", labelKey: "themeDark" },
    { key: "light", labelKey: "themeLight" },
    { key: "contrast", labelKey: "themeContrast" }
  ];
  function applyTheme(t2) {
    document.documentElement.dataset.theme = THEMES.some((x) => x.key === t2) ? t2 : "dark";
  }
  var HOTKEY_ACTIONS = [
    { key: "next", labelKey: "hkNext" },
    { key: "prev", labelKey: "hkPrev" },
    { key: "toggleHide", labelKey: "hkToggle" },
    { key: "focusOverlay", labelKey: "hkFocus" },
    { key: "quit", labelKey: "hkQuit" }
  ];
  var MOD_LABEL = { ctrl: "Ctrl", alt: "Alt", shift: "Shift", win: "Win" };
  var KEY_LABEL = { left: "←", right: "→", up: "↑", down: "↓", space: "Space", return: "Enter", escape: "Esc", delete: "Del", tab: "Tab" };
  var BUTTON_KEY = { left: "mouseL", middle: "mouseM", right: "mouseR", back: "mouse4", x1: "mouse4", forward: "mouse5", x2: "mouse5" };
  function formatBinding(b) {
    if (!b) return "—";
    let trigger;
    if (b.button) trigger = BUTTON_KEY[b.button] ? t(BUTTON_KEY[b.button]) : b.button;
    else if (b.key) trigger = KEY_LABEL[b.key] || b.key.toUpperCase();
    else return "—";
    const parts = (b.mods || []).map((m) => MOD_LABEL[m] || m);
    parts.push(trigger);
    return parts.join(" + ");
  }
  function bindingFromKeyEvent(e) {
    if (["Control", "Alt", "Shift", "Meta"].includes(e.key)) return null;
    let key = null;
    const c = e.code;
    if (/^Key[A-Z]$/.test(c)) key = c.slice(3).toLowerCase();
    else if (/^Digit[0-9]$/.test(c)) key = c.slice(5);
    else if (/^F([1-9]|1[0-2])$/.test(c)) key = c.toLowerCase();
    else key = {
      ArrowLeft: "left",
      ArrowRight: "right",
      ArrowUp: "up",
      ArrowDown: "down",
      Space: "space",
      Enter: "return",
      Escape: "escape",
      Delete: "delete",
      Tab: "tab"
    }[c] || null;
    if (!key) return null;
    return { mods: modsFromEvent(e), key };
  }
  function bindingFromMouseEvent(e) {
    const button = { 0: "left", 1: "middle", 2: "right", 3: "back", 4: "forward" }[e.button];
    if (!button) return null;
    return { mods: modsFromEvent(e), button };
  }
  function modsFromEvent(e) {
    const mods = [];
    if (e.ctrlKey) mods.push("ctrl");
    if (e.altKey) mods.push("alt");
    if (e.shiftKey) mods.push("shift");
    if (e.metaKey) mods.push("win");
    return mods;
  }
  function derive() {
    const total = state.steps.length;
    const doneCount = Math.max(0, state.activePos - 1);
    return { total, doneCount, pct: total ? Math.round(doneCount / total * 100) : 0 };
  }
  function render() {
    mover.classList.toggle("editing", !state.locked);
    mover.classList.toggle("collapsed", state.collapsed && state.view === "steps");
    applyOpacity(state.settings ? state.settings.opacity : 1);
    document.documentElement.lang = lang();
    if (state.view === "picker") {
      renderPicker();
    } else if (state.view === "settings") {
      renderSettings();
    } else {
      renderSteps();
    }
    panel.classList.toggle("steps-expanded", state.view === "steps" && !state.collapsed);
    applyLayout();
    wireEvents();
    renderModal();
    scrollToCurrent();
  }
  function openConfirm({ message, confirmLabel, onConfirm }) {
    state.confirm = { message, confirmLabel, onConfirm };
    render();
  }
  function closeConfirm() {
    if (!state.confirm) return;
    state.confirm = null;
    render();
  }
  function renderModal() {
    const c = state.confirm;
    if (!c) {
      modalHost.innerHTML = "";
      return;
    }
    modalHost.innerHTML = `
        <div class="modal-card" id="modalCard">
            <div class="modal-title">${t("confirmTitle")}</div>
            <div class="modal-msg">${esc(c.message)}</div>
            <div class="modal-actions">
                <div class="modal-btn ghost" id="modalCancel">${t("cancel")}</div>
                <div class="modal-btn danger" id="modalConfirm">${esc(c.confirmLabel)}</div>
            </div>
        </div>`;
    modalHost.addEventListener("mousedown", stopProp);
    modalHost.onclick = (e) => {
      if (e.target === modalHost) closeConfirm();
    };
    document.getElementById("modalCancel").addEventListener("click", closeConfirm);
    document.getElementById("modalConfirm").addEventListener("click", () => {
      const fn = c.onConfirm;
      closeConfirm();
      if (fn) fn();
    });
  }
  function renderPicker() {
    let kicker, rowsHTML;
    if (state.pickerGame) {
      const chapters = state.configs.filter((c) => c.game === state.pickerGame).sort((a, b) => (a.chapter || 0) - (b.chapter || 0));
      kicker = esc(state.pickerGame);
      const rows = chapters.map((c) => `
            <div class="picker-row" data-path="${esc(c.path)}" data-embedded="true">
                <div class="picker-row-main">
                    ${c.chapter ? `<div class="picker-game mono">${esc(tf("chapterLabel", c.chapter))}</div>` : ""}
                    <div class="picker-title">${esc(c.title)}</div>
                    ${c.author ? `<div class="picker-meta mono">${esc(tf("by", c.author))}</div>` : ""}
                </div>
                <div class="picker-clear" data-clear-chapter="${esc(c.path)}" title="${t("clearProgress")}">⟲</div>
            </div>`).join("");
      rowsHTML = `<div class="picker-back mono" id="pickerBack">${t("backToGames")}</div>` + (rows || `<div class="picker-empty">${t("noChapters")}</div>`);
    } else {
      kicker = t("chooseGame");
      const games = [...new Set(state.configs.map((c) => c.game))].sort();
      rowsHTML = games.length ? games.map((g) => {
        const n = state.configs.filter((c) => c.game === g).length;
        return `<div class="picker-row" data-game="${esc(g)}">
                    <div class="picker-row-main">
                        <div class="picker-title">${esc(g)}</div>
                        <div class="picker-meta mono">${esc(tf("chaptersCount", n))}</div>
                    </div>
                    <div class="picker-clear" data-clear-game="${esc(g)}" title="${t("clearProgress")}">⟲</div>
                </div>`;
      }).join("") : `<div class="picker-empty">${t("noConfigs")}</div>`;
    }
    const loadingBadge = state.pickerLoading ? `<div class="picker-loading mono">${esc(tf("downloading", state.pickerLoading))}</div>` : "";
    panel.innerHTML = `
        <div class="header" id="header">
            <div class="header-left">
                <div>
                    <div class="kicker mono">GoThrough</div>
                </div>
            </div>
            <div class="header-right">
                <div class="lock-btn mono" id="lockBtn">${state.locked ? t("adjust") : t("lock")}</div>
                <div class="icon-btn" id="settingsBtn" title="${t("settingsTitle")}">${GEAR_SVG}</div>
                <div class="close-btn" id="closeBtn" title="${t("close")}">✕</div>
            </div>
        </div>
        <div class="picker-view">
            <div class="picker-kicker mono">${kicker}</div>
            <div class="picker-list" id="pickerList">${rowsHTML}</div>
            <div class="picker-actions">
                <div class="picker-browse mono" id="pickerBrowse">${t("openFile")}</div>
                <div class="picker-clearcache mono" id="pickerClearCache">${t("clearCache")}</div>
            </div>
            ${state.pickerError ? `<div class="picker-error">${esc(state.pickerError)}</div>` : ""}
            ${loadingBadge}
        </div>`;
  }
  function rowHTML(s, i) {
    const pos = i + 1;
    const done = pos < state.activePos;
    const isCurrent = pos === state.activePos;
    const label = s.isChoice ? `⑂ ${esc(s.title)}` : esc(s.title);
    const optBadge = s.optional ? `<span class="row-opt">${t("optional")}</span>` : "";
    return `
        <div class="row${done ? " done" : ""}${isCurrent ? " current" : ""}${s.isChoice ? " is-choice" : ""}" data-index="${i}">
            <div class="mark">${done ? '<span class="check">✓</span>' : ""}</div>
            <div class="row-main">
                <div class="row-label">${label}${optBadge}</div>
            </div>
        </div>`;
  }
  function buildChecklist() {
    const groups = [];
    let cur = null;
    state.steps.forEach((s, i) => {
      const sec = s.section || "";
      if (!cur || cur.section !== sec) {
        cur = { section: sec, items: [] };
        groups.push(cur);
      }
      cur.items.push({ s, i });
    });
    return groups.map((g) => {
      const rows = g.items.map(({ s, i }) => rowHTML(s, i)).join("");
      if (!g.section) return rows;
      const allDone = g.items.every(({ i }) => i + 1 < state.activePos);
      const ov = state.sectionOverride[g.section];
      const open = ov === void 0 ? !allDone : ov;
      const badge = allDone ? `<span class="section-done mono">${t("completed")}</span>` : "";
      return `<div class="section-group${open ? "" : " sec-collapsed"}">
            <div class="section-head mono" data-section="${esc(g.section)}">
                <span class="section-chevron">${CHEVRON_SVG}</span>
                <span>${esc(g.section)}</span>
                ${badge}
            </div>
            <div class="section-rows">${rows}</div>
        </div>`;
    }).join("");
  }
  function renderSteps() {
    const { total, doneCount, pct } = derive();
    const current = state.current || state.steps[state.activePos - 1] || null;
    let nowHTML;
    if (!current) {
      nowHTML = `<div class="done-card">${t("noSteps")}</div>`;
    } else if (current.isChoice) {
      const opts = (current.options || []).map((o) => {
        const chosen = current.selected && o.value === current.selected;
        return `
            <button class="choice-opt${chosen ? " chosen" : ""}" data-key="${esc(current.choiceKey)}" data-value="${esc(o.value)}">
                <span class="choice-opt-label">${esc(o.label)}${chosen ? ` <span class="choice-opt-tag">${t("chosen")}</span>` : ""}</span>
                ${o.description ? `<span class="choice-opt-desc">${mdInline(o.description)}</span>` : ""}
            </button>`;
      }).join("");
      nowHTML = `<div class="now-card choice-card">
               <div class="now-label mono">${t("decision")}</div>
               <div class="now-step">${esc(current.title)}</div>
               <div class="now-meta mono">${esc(tf("stepCount", state.activePos, total))}${current.section ? " · " + esc(current.section) : ""}</div>
               <div class="choice-options">${opts}</div>
           </div>`;
    } else {
      const optBadge = current.optional ? `<span class="opt-badge">${t("optional")}</span>` : "";
      const handoff = current.isLast && state.nextFile ? `<button class="next-file-btn" id="nextFileBtn">${esc(tf("nextFile", state.nextFile))}</button>` : "";
      nowHTML = `<div class="now-card">
               <div class="now-label mono">${t("now")}</div>
               <div class="now-step">${esc(current.title)} ${optBadge}</div>
               <div class="now-meta mono">${esc(tf("stepCount", state.activePos, total))}${current.section ? " · " + esc(current.section) : ""}</div>
               ${current.description ? `<div class="now-desc">${md(current.description)}</div>` : ""}
               ${renderTasks(current.tasks)}
               ${renderList(current.warnings, t("warning"), "warnings")}
               ${renderList(current.infos, t("info"), "infos")}
               ${renderQuests(current.quests)}
               ${renderList(current.hints, t("hints"), "hints")}
               ${handoff}
           </div>`;
    }
    const rowsHTML = buildChecklist();
    panel.innerHTML = `
        <div class="header" id="header">
            <div class="header-left">
                <div>
                    <div class="kicker mono">${esc(state.meta.game)}</div>
                    <div class="quest-title" id="questTitle">${esc(state.meta.title)}</div>
                </div>
            </div>
            <div class="header-right">
                <div class="lock-btn mono" id="lockBtn">${state.locked ? t("adjust") : t("lock")}</div>
                <div class="icon-btn" id="homeBtn" title="${t("switchWalkthrough")}">${HOME_SVG}</div>
                <div class="icon-btn" id="settingsBtn" title="${t("settingsTitle")}">${GEAR_SVG}</div>
                <div class="close-btn" id="closeBtn" title="${t("closeOverlay")}">✕</div>
            </div>
        </div>
        ${nowHTML}
        <div class="steps-section">
            <div class="steps-bar" id="stepsBar">
                <span class="steps-bar-left">
                    <span class="steps-chevron">${CHEVRON_SVG}</span>
                    <span class="progress-label mono">${t("steps")}</span>
                </span>
                <div class="progress-track"><div class="progress-fill" style="width:${pct}%"></div></div>
                <span class="progress-count mono">${doneCount} / ${total}</span>
            </div>
            <div class="collapsible">
                <div class="checklist">${rowsHTML}</div>
            </div>
        </div>`;
  }
  function settingsBodyHTML() {
    return `
        <div class="settings-view">
            <div class="settings-kicker mono">${t("shortcuts")}</div>
            ${HOTKEY_ACTIONS.map((a) => {
      const b = state.settings && state.settings.hotkeys[a.key];
      const capturing = state.capturing === a.key;
      return `<div class="setting-row">
                    <span class="setting-name">${esc(t(a.labelKey))}</span>
                    <span class="combo mono${capturing ? " capturing" : ""}" data-action="${a.key}">${capturing ? t("pressKey") : esc(formatBinding(b))}</span>
                </div>`;
    }).join("")}
            <div class="setting-row">
                <span class="setting-name">${t("opacity")}</span>
                <input type="range" class="opacity-slider" id="opacitySlider"
                    min="10" max="100" step="5"
                    value="${Math.round((state.settings ? state.settings.opacity ?? 1 : 1) * 100)}">
            </div>
            <div class="setting-row">
                <span class="setting-name">${t("design")}</span>
                <span class="theme-switch">${THEMES.map((th) => {
      const active = (state.settings && state.settings.theme || "dark") === th.key;
      return `<button class="theme-opt${active ? " active" : ""}" data-theme="${th.key}">${esc(t(th.labelKey))}</button>`;
    }).join("")}</span>
            </div>
            <div class="setting-row">
                <span class="setting-name">${t("language")}</span>
                <span class="theme-switch">${LANGUAGES.map((l) => {
      const active = lang() === l.key;
      return `<button class="lang-opt${active ? " active" : ""}" data-lang="${l.key}">${esc(l.label)}</button>`;
    }).join("")}</span>
            </div>
            <div class="settings-error">${esc(state.settingsError)}</div>
            <div class="settings-hint">${t("settingsHint")}</div>
            <button class="settings-reset" id="resetSettings">${t("resetSettings")}</button>
        </div>`;
  }
  function renderSettings() {
    const loaded = state.settingsFrom === "steps";
    panel.innerHTML = `
        <div class="header" id="header">
            <div class="header-left">
                <div>
                    <div class="kicker mono">${loaded ? esc(state.meta.game) : "GoThrough"}</div>
                    ${loaded ? `<div class="quest-title">${esc(state.meta.title)}</div>` : ""}
                </div>
            </div>
            <div class="header-right">
                <div class="lock-btn mono" id="lockBtn">${state.locked ? t("adjust") : t("lock")}</div>
                ${loaded ? `<div class="icon-btn" id="homeBtn" title="${t("switchWalkthrough")}">${HOME_SVG}</div>` : ""}
                <div class="icon-btn active" id="settingsBtn" title="${t("settingsTitle")}">${GEAR_SVG}</div>
                <div class="close-btn" id="closeBtn" title="${loaded ? t("closeOverlay") : t("close")}">✕</div>
            </div>
        </div>
        ${settingsBodyHTML()}`;
  }
  function applyLayout() {
    mover.style.left = "0px";
    mover.style.top = "0px";
    mover.style.width = state.width + "px";
    const expanded = state.view === "steps" && !state.collapsed;
    panel.style.height = expanded ? state.panelHeight + "px" : "auto";
  }
  function scrollToCurrent() {
    if (state.view !== "steps" || state.collapsed) return;
    if (state.lastScrolledPos === state.activePos) return;
    const cl = panel.querySelector(".checklist");
    const row = cl && cl.querySelector(".row.current");
    if (cl && row) {
      const PEEK = 40;
      const r = row.getBoundingClientRect(), c = cl.getBoundingClientRect();
      cl.scrollTop = Math.max(0, cl.scrollTop + (r.top - c.top) - PEEK);
      state.lastScrolledPos = state.activePos;
    }
  }
  function gotoIndex(index) {
    App.Goto(index).then((info) => {
      state.activePos = info.current;
      state.current = info;
      render();
    });
  }
  function downloadGame(game) {
    state.pickerLoading = game;
    state.pickerError = "";
    render();
    App.DownloadGame(game).then(() => {
      state.pickerLoading = "";
      state.pickerGame = game;
      render();
    }).catch((err) => {
      state.pickerLoading = "";
      state.pickerGame = game;
      state.pickerError = String(err && err.message ? err.message : err);
      render();
    });
  }
  function clearGameProgress(game) {
    openConfirm({
      message: tf("confirmClearGame", game),
      confirmLabel: t("confirmReset"),
      onConfirm: () => {
        state.pickerError = "";
        App.ClearGameProgress(game).then(() => render()).catch((err) => {
          state.pickerError = String(err && err.message ? err.message : err);
          render();
        });
      }
    });
  }
  function clearChapterProgress(path) {
    state.pickerError = "";
    App.ClearChapterProgress(path).then(() => render()).catch((err) => {
      state.pickerError = String(err && err.message ? err.message : err);
      render();
    });
  }
  function clearCache() {
    openConfirm({
      message: t("confirmClearCache"),
      confirmLabel: t("confirmDelete"),
      onConfirm: () => {
        state.pickerError = "";
        App.ClearCache().then(() => {
          state.pickerGame = null;
          render();
        }).catch((err) => {
          state.pickerError = String(err && err.message ? err.message : err);
          render();
        });
      }
    });
  }
  function resetSettings() {
    App.ResetSettings().then((s) => {
      state.settings = s;
      applyOpacity(s.opacity ?? 1);
      applyTheme(s.theme);
      save();
      state.settingsError = "";
      render();
    }).catch((err) => {
      state.settingsError = String(err && err.message ? err.message : err);
      render();
    });
  }
  function loadConfig(path, embedded) {
    state.pickerError = "";
    App.LoadConfig(path, embedded).then(() => {
      return Promise.all([App.Meta(), App.Steps(), App.CurrentStep(), App.NextFile()]);
    }).then(([meta, steps, current, nextFile]) => {
      state.meta = meta;
      state.steps = steps;
      state.current = current;
      state.activePos = current.current;
      state.nextFile = nextFile || "";
      state.sectionOverride = {};
      state.lastScrolledPos = null;
      state.view = "steps";
      render();
    }).catch((err) => {
      state.pickerError = String(err && err.message ? err.message : err);
      render();
    });
  }
  function reloadSteps() {
    return Promise.all([App.Steps(), App.CurrentStep(), App.NextFile()]).then(([steps, current, nextFile]) => {
      state.steps = steps;
      state.current = current;
      state.activePos = current.current;
      state.nextFile = nextFile || "";
      state.lastScrolledPos = null;
      render();
    });
  }
  function chooseOption(key, value) {
    App.Choose(key, value).then(() => reloadSteps());
  }
  function refreshAfterSwap() {
    state.sectionOverride = {};
    return App.Meta().then((meta) => {
      state.meta = meta;
      return reloadSteps();
    });
  }
  function loadNext() {
    App.LoadNext().then(() => refreshAfterSwap()).catch((err) => {
      console.error("[GoThrough] next hand-off failed:", err);
    });
  }
  function goHome() {
    App.UnloadConfig().then(() => {
      state.view = "picker";
      state.pickerGame = null;
      state.configs = [];
      return App.ListConfigs();
    }).then((configs) => {
      state.configs = configs;
      render();
    });
  }
  function wireEvents() {
    document.getElementById("header").addEventListener("mousedown", startWindowDrag);
    const lockBtn = document.getElementById("lockBtn");
    lockBtn.addEventListener("mousedown", stopProp);
    lockBtn.addEventListener("click", toggleLock);
    const closeBtn = document.getElementById("closeBtn");
    closeBtn.addEventListener("mousedown", stopProp);
    closeBtn.addEventListener("click", () => {
      window.runtime?.Quit?.();
    });
    const pickerSettingsBtn = state.view === "picker" && document.getElementById("settingsBtn");
    if (pickerSettingsBtn) {
      pickerSettingsBtn.addEventListener("mousedown", stopProp);
      pickerSettingsBtn.addEventListener("click", toggleSettings);
    }
    if (state.view === "picker") {
      panel.querySelectorAll(".picker-row").forEach((row) => {
        row.addEventListener("click", () => {
          if (row.dataset.game) {
            downloadGame(row.dataset.game);
          } else if (row.dataset.path) {
            loadConfig(row.dataset.path, row.dataset.embedded === "true");
          }
        });
      });
      panel.querySelectorAll(".picker-clear").forEach((btn) => {
        btn.addEventListener("mousedown", stopProp);
        btn.addEventListener("click", (e) => {
          e.stopPropagation();
          if (btn.dataset.clearGame) clearGameProgress(btn.dataset.clearGame);
          else if (btn.dataset.clearChapter) clearChapterProgress(btn.dataset.clearChapter);
        });
      });
      const back = document.getElementById("pickerBack");
      if (back) {
        back.addEventListener("mousedown", stopProp);
        back.addEventListener("click", () => {
          state.pickerGame = null;
          render();
        });
      }
      const browse = document.getElementById("pickerBrowse");
      if (browse) {
        browse.addEventListener("mousedown", stopProp);
        browse.addEventListener("click", () => {
          App.OpenBrowse().then((path) => {
            if (path) loadConfig(path, false);
          });
        });
      }
      const clearCacheBtn = document.getElementById("pickerClearCache");
      if (clearCacheBtn) {
        clearCacheBtn.addEventListener("mousedown", stopProp);
        clearCacheBtn.addEventListener("click", clearCache);
      }
      return;
    }
    const stepsBar = document.getElementById("stepsBar");
    if (stepsBar) {
      stepsBar.addEventListener("click", () => {
        state.collapsed = !state.collapsed;
        render();
      });
    }
    panel.querySelectorAll(".section-head").forEach((head) => {
      head.addEventListener("click", () => toggleSection(head.dataset.section));
    });
    const homeBtn = document.getElementById("homeBtn");
    if (homeBtn) {
      homeBtn.addEventListener("mousedown", stopProp);
      homeBtn.addEventListener("click", goHome);
    }
    const settingsBtn = document.getElementById("settingsBtn");
    if (settingsBtn) {
      settingsBtn.addEventListener("mousedown", stopProp);
      settingsBtn.addEventListener("click", toggleSettings);
    }
    panel.querySelectorAll(".row").forEach((row) => row.addEventListener("click", () => gotoIndex(Number(row.dataset.index))));
    panel.querySelectorAll(".choice-opt").forEach((btn) => {
      btn.addEventListener("mousedown", stopProp);
      btn.addEventListener("click", () => chooseOption(btn.dataset.key, btn.dataset.value));
    });
    const nextBtn = document.getElementById("nextFileBtn");
    if (nextBtn) {
      nextBtn.addEventListener("mousedown", stopProp);
      nextBtn.addEventListener("click", loadNext);
    }
    panel.querySelectorAll(".combo").forEach((combo) => {
      combo.addEventListener("mousedown", stopProp);
      combo.addEventListener("click", () => startCapture(combo.dataset.action));
    });
    const opacitySlider = document.getElementById("opacitySlider");
    if (opacitySlider) {
      opacitySlider.addEventListener("mousedown", stopProp);
      opacitySlider.addEventListener("input", () => {
        const v = Number(opacitySlider.value) / 100;
        applyOpacity(v);
      });
      opacitySlider.addEventListener("change", () => {
        const v = Number(opacitySlider.value) / 100;
        App.SaveOpacity(v).then((s) => {
          state.settings = s;
        });
      });
    }
    panel.querySelectorAll(".theme-opt").forEach((btn) => {
      btn.addEventListener("mousedown", stopProp);
      btn.addEventListener("click", () => selectTheme(btn.dataset.theme));
    });
    panel.querySelectorAll(".lang-opt").forEach((btn) => {
      btn.addEventListener("mousedown", stopProp);
      btn.addEventListener("click", () => selectLanguage(btn.dataset.lang));
    });
    const resetBtn = document.getElementById("resetSettings");
    if (resetBtn) {
      resetBtn.addEventListener("mousedown", stopProp);
      resetBtn.addEventListener("click", resetSettings);
    }
  }
  function toggleSection(name) {
    if (!name) return;
    const items = state.steps.filter((s) => (s.section || "") === name);
    const allDone = items.length > 0 && state.steps.every((s, i) => (s.section || "") !== name || i + 1 < state.activePos);
    const ov = state.sectionOverride[name];
    const open = ov === void 0 ? !allDone : ov;
    state.sectionOverride[name] = !open;
    render();
  }
  function selectLanguage(l) {
    if (l === lang()) return;
    App.SaveLanguage(l).then((s) => {
      state.settings = s;
      render();
    }).catch((err) => {
      state.settingsError = String(err && err.message ? err.message : err);
      render();
    });
  }
  function selectTheme(theme) {
    const prev = state.settings ? state.settings.theme : "dark";
    applyTheme(theme);
    App.SaveTheme(theme).then((s) => {
      state.settings = s;
      save();
      render();
    }).catch((err) => {
      applyTheme(prev);
      state.settingsError = String(err && err.message ? err.message : err);
      render();
    });
  }
  function toggleSettings() {
    if (state.view === "settings") {
      state.view = state.settingsFrom || "steps";
    } else {
      state.settingsFrom = state.view;
      state.view = "settings";
    }
    state.capturing = null;
    state.settingsError = "";
    render();
  }
  function startCapture(action) {
    state.capturing = action;
    state.settingsError = "";
    render();
  }
  function applyCapturedBinding(binding) {
    const action = state.capturing;
    const hotkeys = { ...state.settings.hotkeys, [action]: binding };
    window.go.overlay.App.SaveHotkeys(hotkeys).then((saved2) => {
      state.settings = saved2;
      state.capturing = null;
      state.settingsError = "";
      render();
    }).catch((err) => {
      state.capturing = null;
      state.settingsError = String(err && err.message ? err.message : err);
      render();
    });
  }
  function onCaptureKeydown(e) {
    if (state.confirm) {
      if (e.key === "Escape") {
        e.preventDefault();
        e.stopPropagation();
        closeConfirm();
      } else if (e.key === "Enter") {
        e.preventDefault();
        e.stopPropagation();
        const fn = state.confirm.onConfirm;
        closeConfirm();
        if (fn) fn();
      }
      return;
    }
    if (!state.capturing) return;
    e.preventDefault();
    e.stopPropagation();
    if (e.key === "Escape") {
      state.capturing = null;
      render();
      return;
    }
    const binding = bindingFromKeyEvent(e);
    if (!binding) return;
    applyCapturedBinding(binding);
  }
  function onCaptureMousedown(e) {
    if (!state.capturing) return;
    e.preventDefault();
    e.stopPropagation();
    const binding = bindingFromMouseEvent(e);
    if (!binding) return;
    applyCapturedBinding(binding);
  }
  window.addEventListener("keydown", onCaptureKeydown, true);
  window.addEventListener("mousedown", onCaptureMousedown, true);
  window.addEventListener("contextmenu", (e) => {
    if (state.capturing) e.preventDefault();
  }, true);
  function toggleLock() {
    state.locked = !state.locked;
    save();
    render();
  }
  function stopProp(e) {
    e.stopPropagation();
  }
  var drag = null;
  function startResize(e) {
    if (state.locked) return;
    e.preventDefault();
    e.stopPropagation();
    drag = { sx: e.screenX, sy: e.screenY, ow: state.width, oh: state.panelHeight };
    mover.classList.add("dragging");
  }
  function onMove(e) {
    if (!drag) return;
    state.width = Math.max(MIN_W, Math.min(drag.ow - (e.screenX - drag.sx), MAX_W));
    state.panelHeight = Math.max(MIN_H, Math.min(drag.oh + (e.screenY - drag.sy), MAX_H));
    applyLayout();
  }
  function onUp() {
    if (drag) {
      drag = null;
      mover.classList.remove("dragging");
      save();
    }
  }
  resizeHandle.addEventListener("mousedown", startResize);
  window.addEventListener("mousemove", onMove);
  window.addEventListener("mouseup", onUp);
  var winDrag = null;
  async function startWindowDrag(e) {
    if (state.locked || e.button !== 0) return;
    if (e.target.closest(".lock-btn, .icon-btn, .close-btn, .picker-browse, .picker-clearcache, .picker-row, .picker-back")) return;
    const rt = window.runtime;
    if (!rt || !rt.WindowGetPosition) return;
    e.preventDefault();
    const [pos, size, screens] = await Promise.all([
      rt.WindowGetPosition(),
      rt.WindowGetSize(),
      rt.ScreenGetAll()
    ]);
    const scr = (screens || []).find((s) => s.isPrimary) || (screens || [])[0];
    const sw = scr ? scr.size?.width || scr.width : Infinity;
    const sh = scr ? scr.size?.height || scr.height : Infinity;
    winDrag = {
      sx: e.screenX,
      sy: e.screenY,
      ox: pos.x,
      oy: pos.y,
      maxX: isFinite(sw) ? Math.max(0, sw - size.w) : Infinity,
      maxY: isFinite(sh) ? Math.max(0, sh - size.h) : Infinity
    };
  }
  function onWinMove(e) {
    if (!winDrag) return;
    const nx = Math.max(0, Math.min(winDrag.ox + (e.screenX - winDrag.sx), winDrag.maxX));
    const ny = Math.max(0, Math.min(winDrag.oy + (e.screenY - winDrag.sy), winDrag.maxY));
    winDrag.lastX = Math.round(nx);
    winDrag.lastY = Math.round(ny);
    window.runtime.WindowSetPosition(winDrag.lastX, winDrag.lastY);
  }
  function onWinUp() {
    if (winDrag && winDrag.lastX != null) {
      App.SaveWindowPos?.(winDrag.lastX, winDrag.lastY);
    }
    winDrag = null;
  }
  window.addEventListener("mousemove", onWinMove);
  window.addEventListener("mouseup", onWinUp);
  function fitWindow() {
    const expanded = state.view === "steps" && !state.collapsed;
    let w, h;
    if (expanded) {
      w = state.width;
      h = state.panelHeight;
    } else {
      const r = panel.getBoundingClientRect();
      w = Math.ceil(r.width);
      h = Math.ceil(r.height);
    }
    if (w > 0 && h > 0) window.go?.overlay?.App?.FitWindow?.(w, h);
  }
  new ResizeObserver(fitWindow).observe(panel);
  async function init() {
    applyTheme(saved.theme);
    try {
      const settings = await App.Settings();
      state.settings = settings;
      applyOpacity(settings.opacity ?? 1);
      applyTheme(settings.theme);
      const isPicker = await App.IsPicker();
      if (isPicker) {
        state.view = "picker";
        state.configs = await App.ListConfigs();
      } else {
        const [meta, steps, current, nextFile] = await Promise.all([
          App.Meta(),
          App.Steps(),
          App.CurrentStep(),
          App.NextFile()
        ]);
        state.meta = meta;
        state.steps = steps;
        state.current = current;
        state.activePos = current.current;
        state.nextFile = nextFile || "";
        state.view = "steps";
      }
    } catch (e) {
      console.error("[GoThrough] init failed:", e);
    }
    render();
    if (window.runtime && window.runtime.EventsOn) {
      window.runtime.EventsOn("step:changed", (info) => {
        state.activePos = info.current;
        state.current = info;
        render();
      });
      window.runtime.EventsOn("config:changed", () => {
        refreshAfterSwap();
      });
      window.runtime.EventsOn("configs:remote", (entries) => {
        if (Array.isArray(entries) && entries.length) {
          state.configs = entries;
          if (state.view === "picker") render();
        }
      });
    }
  }
  init();
})();
