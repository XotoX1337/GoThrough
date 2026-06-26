// ===================================================================
// GoThrough HUD — wired to the real Wails backend (overlay.App).
// ===================================================================

// ===================================================================
// Type declarations
// ---------------------------------------------------------------
// These mirror the Go wire structs (overlay/app.go, settings/settings.go,
// configstore/configstore.go); fields arrive as JSON with lower-cased keys.
// esbuild only strips these types (no type-checking) — they exist for editor
// support. Keep them in sync with the Go structs and the devui mock
// (tools/devui/main.go). This file is a global script (no import/export), so a
// plain `interface Window` merges into the DOM lib's Window — no `declare global`.
// ===================================================================

interface Binding {
    mods: string[];
    key?: string;
    button?: string;
}

interface Hotkeys {
    next: Binding;
    prev: Binding;
    toggleHide: Binding;
    focusOverlay: Binding;
    quit: Binding;
}

interface WindowPos {
    x: number;
    y: number;
    set: boolean;
}

interface LastConfig {
    path: string;
    embedded: boolean;
}

interface Settings {
    version: number;
    hotkeys: Hotkeys;
    opacity: number;
    theme: string;
    language: string;
    lastConfig?: LastConfig;
    windowPos?: WindowPos;
}

interface Entry {
    game: string;
    title: string;
    author: string;
    chapter: number;
    path: string;
    hash: string;
}

interface QuestInfo {
    name: string;
    status?: string; // received | completed | ""
    note?: string;
}

interface TaskInfo {
    text: string;
    info?: string;
    warning?: string;
    hint?: string;
}

interface ChoiceOptionInfo {
    value: string;
    label: string;
    description?: string;
}

// StepInfo is one position in the resolved sequence: a normal step, OR a choice
// when isChoice is true (then title is the prompt, choiceKey identifies it, and
// options are the answers). description / task text are Markdown.
interface StepInfo {
    current: number;
    total: number;
    isFirst: boolean;
    isLast: boolean;
    section?: string;
    // Choice (isChoice === true)
    isChoice?: boolean;
    choiceKey?: string;
    selected?: string;
    options?: ChoiceOptionInfo[];
    // Step content (isChoice === false)
    id?: number;
    title: string;
    description?: string;
    tasks?: TaskInfo[];
    optional?: boolean;
    quests?: QuestInfo[];
    hints?: string[];
    warnings?: string[];
    infos?: string[];
}

interface MetaInfo {
    game: string;
    title: string;
    variant?: string;
}

// OverlayApp is the Wails-bound backend (overlay.App). Wails wraps every bound
// method as a Promise; methods returning a Go error reject on failure.
interface OverlayApp {
    IsPicker(): Promise<boolean>;
    ListConfigs(): Promise<Entry[]>;
    DownloadGame(game: string): Promise<void>;
    OpenBrowse(): Promise<string>;
    LoadConfig(path: string, embedded: boolean): Promise<void>;
    UnloadConfig(): Promise<void>;
    ClearChapterProgress(relpath: string): Promise<void>;
    ClearGameProgress(game: string): Promise<void>;
    ClearCache(): Promise<void>;
    ResetSettings(): Promise<Settings>;
    Meta(): Promise<MetaInfo>;
    Steps(): Promise<StepInfo[]>;
    Choose(choiceKey: string, value: string): Promise<StepInfo>;
    NextFile(): Promise<string>;
    LoadNext(): Promise<void>;
    CurrentStep(): Promise<StepInfo>;
    Next(): Promise<StepInfo>;
    Prev(): Promise<StepInfo>;
    Goto(index: number): Promise<StepInfo>;
    Settings(): Promise<Settings>;
    SaveHotkeys(hotkeys: Hotkeys): Promise<Settings>;
    SaveOpacity(opacity: number): Promise<Settings>;
    SaveTheme(theme: string): Promise<Settings>;
    SaveLanguage(lang: string): Promise<Settings>;
    SaveWindowPos(x: number, y: number): Promise<void>;
    FitWindow(width: number, height: number): Promise<void>;
}

// WailsRuntime is the subset of window.runtime the HUD uses.
interface WailsRuntime {
    Quit?(): void;
    WindowSetPosition(x: number, y: number): void;
    EventsOn(event: string, cb: (...data: any[]) => void): void;
}

interface Window {
    go: { overlay: { App: OverlayApp } };
    runtime: WailsRuntime;
}

// State is the single in-memory UI state object the render loop reads from.
interface State {
    view: "picker" | "steps" | "settings";
    configs: Entry[];
    pickerGame: string | null;
    pickerLoading: string;
    pickerError: string;
    meta: MetaInfo;
    steps: StepInfo[];
    current: StepInfo | null;
    activePos: number;
    nextFile: string;
    sectionOverride: Record<string, boolean>;
    lastScrolledPos: number | null;
    settings: Settings | null;
    capturing: any;
    confirm: any;
    settingsError: string;
    settingsFrom: "picker" | "steps";
    collapsed: boolean;
    locked: boolean;
    width: number;
    panelHeight: number;
}

const App = window.go.overlay.App;

const STORE_KEY = "gt-overlay-v7"; // v7: narrower default width + lower min width
const saved = (() => { try { return JSON.parse(localStorage.getItem(STORE_KEY) || "{}"); } catch { return {}; } })();
// Size the HUD against the real desktop. screen.availHeight is the work area
// (taskbar excluded) in CSS px — the same unit Wails' WindowSetSize uses, so
// it's DPI-correct. The now-card is a FIXED golden-ratio (≈38%) scroll region;
// the window defaults to that plus room for the list, and can be dragged up to
// the full available height.
const DESKTOP_H = (window.screen && window.screen.availHeight) || 920;
// Fixed now-card height as a fraction of desktop. 0.382 = the golden-ratio
// minor part (1/φ²) — sits between 1/3 (too little) and 1/2 (too much).
const NOWCARD_H = Math.round(DESKTOP_H * 0.382);
const DEFAULT_W = 380, DEFAULT_H = Math.min(DESKTOP_H, NOWCARD_H + 300);
const state: State = {
    view: "picker",           // "picker" | "steps" | "settings"
    // --- picker ---
    configs: [],              // []configstore.Entry catalog from App.ListConfigs()
    pickerGame: null,         // selected game in the two-level picker (null = game list)
    pickerLoading: "",        // game currently being downloaded ("" = none)
    pickerError: "",
    // --- steps ---
    meta: { game: "", title: "" },
    steps: [],
    current: null,            // live StepInfo of the active item (now-card source of truth)
    activePos: 1,
    nextFile: "",             // `next:` reference for end-of-walkthrough hand-off
    sectionOverride: {},      // section name → user's explicit open(true)/closed(false)
    lastScrolledPos: null,    // position auto-scroll last targeted (only scroll on change)
    // --- settings ---
    settings: null,           // {version, hotkeys, opacity, theme} from App.Settings()
    capturing: null,
    confirm: null,            // {message, confirmLabel, onConfirm} for the modal (null = closed)
    settingsError: "",
    settingsFrom: "picker",   // view to return to when settings is closed ("picker" | "steps")
    // --- layout ---
    collapsed: false,
    locked: true,             // always start fixed; never persisted
    width: saved.width || DEFAULT_W,
    panelHeight: saved.panelHeight || DEFAULT_H, // fixed window height when expanded
};

const MIN_W = 240, MAX_W = 680;
const MIN_H = 240, MAX_H = DESKTOP_H;

const mover = document.getElementById("mover");
const panel = document.getElementById("panel");
const resizeHandle = document.getElementById("resize");
const modalHost = document.getElementById("modal");
mover.style.setProperty("--gt-nowh", NOWCARD_H + "px"); // fixed now-card height (≈38% desktop, golden ratio)

// Chevron icon (▾), rotated via CSS when collapsed.
const CHEVRON_SVG = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="6 9 12 15 18 9"></polyline></svg>';
// Header icons, shared across the picker / steps / settings headers.
const HOME_SVG = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"></path><polyline points="9 22 9 12 15 12 15 22"></polyline></svg>';
const GEAR_SVG = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="3"></circle><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"></path></svg>';

function save() {
    try {
        localStorage.setItem(STORE_KEY, JSON.stringify({
            width: state.width, panelHeight: state.panelHeight,
            // Cached so the theme can be applied synchronously on next launch,
            // before the async Settings() round-trip — avoids a wrong-colour
            // flash of the themed UI (e.g. the progress bar) on restart.
            theme: state.settings ? state.settings.theme : saved.theme,
        }));
    } catch {}
}

function esc(s) {
    return String(s ?? "").replace(/[&<>"]/g,
        c => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;" }[c]));
}

// --- Markdown (tiny, escape-first) ------------------------------------
// Walkthrough text (incl. remote configs) is untrusted, so HTML is escaped
// BEFORE any markdown is applied — the markers (** * -) are matched against
// already-escaped text, which can never reintroduce tags.
function mdInline(s) {
    return esc(s)
        .replace(/\*\*([^*]+)\*\*/g, "<strong>$1</strong>")
        .replace(/\*([^*]+)\*/g, "<em>$1</em>");
}
function md(s) {
    const lines = String(s ?? "").split("\n");
    const out = [];
    let inList = false;
    for (const raw of lines) {
        const m = raw.match(/^\s*-\s+(.*)$/);
        if (m) {
            if (!inList) { out.push("<ul>"); inList = true; }
            out.push("<li>" + mdInline(m[1]) + "</li>");
        } else {
            if (inList) { out.push("</ul>"); inList = false; }
            if (raw.trim()) out.push("<p>" + mdInline(raw) + "</p>");
        }
    }
    if (inList) out.push("</ul>");
    return out.join("");
}

// --- i18n -------------------------------------------------------------
// UI strings keyed by language. English is the default; German is selectable
// in settings (state.settings.language). Values are either plain strings or
// functions for the few parametric/word-order-sensitive ones (resolved via
// tf()). Walkthrough content itself is authored in the config, not here.
const I18N = {
    en: {
        adjust: "Adjust", lock: "Lock", close: "Close",
        openFile: "Open file…", chooseGame: "Choose game", backToGames: "← Games",
        chapterLabel: n => "Chapter " + n,
        chaptersCount: n => n + (n === 1 ? " chapter" : " chapters"),
        by: a => "by " + a,
        noChapters: "No chapters found.",
        noConfigs: "No configs found.",
        downloading: g => "Downloading " + g + "…",
        clearProgress: "Reset progress",
        clearCache: "Clear cache",
        resetSettings: "Reset to defaults",
        confirmClearCache: "Delete all downloaded configs from the cache?",
        confirmClearGame: g => "Reset all saved progress for " + g + "?",
        confirmTitle: "Are you sure?", cancel: "Cancel", confirmDelete: "Delete", confirmReset: "Reset",
        switchWalkthrough: "Switch walkthrough", settingsTitle: "Settings", closeOverlay: "Close overlay",
        steps: "Steps", noSteps: "No steps loaded", now: "Now", decision: "Decision",
        stepCount: (a, b) => "Step " + a + " / " + b,
        chosen: "chosen", nextFile: f => "Continue ➜ " + f, optional: "optional",
        quests: "Quests", completed: "✓ completed", received: "received",
        warning: "Caution", info: "Info", hints: "Hints",
        shortcuts: "Shortcuts", pressKey: "Press key/mouse…",
        opacity: "Opacity", design: "Theme", language: "Language",
        settingsHint: "Click a shortcut, then press a key or mouse combo. Esc cancels.",
        hkNext: "Next step", hkPrev: "Previous step", hkToggle: "Toggle overlay",
        hkFocus: "Mouse into overlay", hkQuit: "Close overlay",
        themeDark: "Dark", themeLight: "Light", themeContrast: "Contrast",
        mouseL: "Mouse L", mouseM: "Mouse M", mouseR: "Mouse R", mouse4: "Mouse 4", mouse5: "Mouse 5",
    },
    de: {
        adjust: "Anpassen", lock: "Fixieren", close: "Schließen",
        openFile: "Datei öffnen…", chooseGame: "Spiel wählen", backToGames: "← Spiele",
        chapterLabel: n => "Kapitel " + n,
        chaptersCount: n => n + (n === 1 ? " Kapitel" : " Kapitel"),
        by: a => "von " + a,
        noChapters: "Keine Kapitel gefunden.",
        noConfigs: "Keine Configs gefunden.",
        downloading: g => "Lade " + g + "…",
        clearProgress: "Fortschritt zurücksetzen",
        clearCache: "Cache leeren",
        resetSettings: "Auf Standard zurücksetzen",
        confirmClearCache: "Alle heruntergeladenen Configs aus dem Cache löschen?",
        confirmClearGame: g => "Allen gespeicherten Fortschritt für " + g + " zurücksetzen?",
        confirmTitle: "Bist du sicher?", cancel: "Abbrechen", confirmDelete: "Löschen", confirmReset: "Zurücksetzen",
        switchWalkthrough: "Anderes Walkthrough", settingsTitle: "Einstellungen", closeOverlay: "Overlay schließen",
        steps: "Schritte", noSteps: "Keine Schritte geladen", now: "Jetzt", decision: "Entscheidung",
        stepCount: (a, b) => "Schritt " + a + " / " + b,
        chosen: "gewählt", nextFile: f => "Weiter ➜ " + f, optional: "optional",
        quests: "Quests", completed: "✓ abgeschlossen", received: "erhalten",
        warning: "Achtung", info: "Info", hints: "Hinweise",
        shortcuts: "Tastenkürzel", pressKey: "Taste/Maus drücken…",
        opacity: "Transparenz", design: "Design", language: "Sprache",
        settingsHint: "Klick auf ein Kürzel, dann Tasten- oder Maustasten-Kombination drücken. Esc bricht ab.",
        hkNext: "Nächster Schritt", hkPrev: "Vorheriger Schritt", hkToggle: "Overlay ein/aus",
        hkFocus: "Maus ins Overlay", hkQuit: "Overlay schließen",
        themeDark: "Dunkel", themeLight: "Hell", themeContrast: "Kontrast",
        mouseL: "Maus L", mouseM: "Maus M", mouseR: "Maus R", mouse4: "Maus 4", mouse5: "Maus 5",
    },
};
const LANGUAGES = [
    { key: "en", label: "English" },
    { key: "de", label: "Deutsch" },
];
function lang() {
    const l = state.settings && state.settings.language;
    return I18N[l] ? l : "en";
}
function t(key) {
    const v = I18N[lang()][key];
    return v != null ? v : (I18N.en[key] != null ? I18N.en[key] : key);
}
// tf resolves a (possibly parametric) string: passes args through if the
// entry is a function, otherwise returns the plain string.
function tf(key, ...args) {
    const v = t(key);
    return typeof v === "function" ? v(...args) : v;
}

function renderQuests(quests) {
    if (!quests || !quests.length) return "";
    const rows = quests.map(q => {
        const status = q.status === "completed"
            ? `<span class="quest-status done">${t("completed")}</span>`
            : (q.status === "received" ? `<span class="quest-status">${t("received")}</span>` : "");
        const note = q.note ? `<div class="quest-note">${mdInline(q.note)}</div>` : "";
        return `<div class="quest ${esc(q.status || "")}">
                    <div class="quest-line"><span class="quest-name">${esc(q.name)}</span> ${status}</div>${note}
                </div>`;
    }).join("");
    return `<div class="info-block"><div class="info-head mono">${t("quests")}</div>${rows}</div>`;
}
function renderList(items, head, cls) {
    if (!items || !items.length) return "";
    const lis = items.map(t => `<li>${mdInline(t)}</li>`).join("");
    return `<div class="info-block ${cls || ""}"><div class="info-head mono">${head}</div><ul class="info-list">${lis}</ul></div>`;
}
// renderTasks renders a step's actionable sub-steps. Each task is a bullet;
// an optional per-task info/warning/hint renders as a small callout beneath it.
function taskCallout(kind, text) {
    if (!text) return "";
    const tag = kind === "hint" ? "" : `<span class="task-callout-tag mono">${t(kind)}</span>`;
    return `<span class="task-callout ${kind}">${tag}${mdInline(text)}</span>`;
}
function renderTasks(tasks) {
    if (!tasks || !tasks.length) return "";
    const items = tasks.map(t => `<li class="task">${mdInline(t.text)}${
        taskCallout("warning", t.warning)}${taskCallout("info", t.info)}${taskCallout("hint", t.hint)}</li>`).join("");
    return `<ul class="task-list">${items}</ul>`;
}

// --- Opacity ----------------------------------------------------------
function applyOpacity(v) {
    // Drives the panel's background alpha (not element opacity), so 1.0
    // is a fully solid, full-colour overlay and lower values go glassy.
    mover.style.setProperty("--gt-opacity", (v != null && v > 0) ? v : 1);
}

// --- Theme ------------------------------------------------------------
const THEMES = [
    { key: "dark",     labelKey: "themeDark" },
    { key: "light",    labelKey: "themeLight" },
    { key: "contrast", labelKey: "themeContrast" },
];
function applyTheme(t) {
    // Sets html[data-theme]; the stylesheet's token blocks do the rest.
    document.documentElement.dataset.theme = THEMES.some(x => x.key === t) ? t : "dark";
}

// --- Hotkeys ----------------------------------------------------------
const HOTKEY_ACTIONS = [
    { key: "next",         labelKey: "hkNext" },
    { key: "prev",         labelKey: "hkPrev" },
    { key: "toggleHide",   labelKey: "hkToggle" },
    { key: "focusOverlay", labelKey: "hkFocus" },
    { key: "quit",         labelKey: "hkQuit" },
];

const MOD_LABEL = { ctrl: "Ctrl", alt: "Alt", shift: "Shift", win: "Win" };
const KEY_LABEL = { left: "←", right: "→", up: "↑", down: "↓", space: "Space", return: "Enter", escape: "Esc", delete: "Del", tab: "Tab" };
// Mouse-button labels are localised ("Mouse L" / "Maus L"); the keys map a
// binding's button name to an I18N key.
const BUTTON_KEY = { left: "mouseL", middle: "mouseM", right: "mouseR", back: "mouse4", x1: "mouse4", forward: "mouse5", x2: "mouse5" };
function formatBinding(b) {
    if (!b) return "—";
    let trigger;
    if (b.button) trigger = BUTTON_KEY[b.button] ? t(BUTTON_KEY[b.button]) : b.button;
    else if (b.key) trigger = KEY_LABEL[b.key] || b.key.toUpperCase();
    else return "—";
    const parts = (b.mods || []).map(m => MOD_LABEL[m] || m);
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
    else key = ({
        ArrowLeft: "left", ArrowRight: "right", ArrowUp: "up", ArrowDown: "down",
        Space: "space", Enter: "return", Escape: "escape", Delete: "delete", Tab: "tab",
    })[c] || null;
    if (!key) return null;
    return { mods: modsFromEvent(e), key };
}

function bindingFromMouseEvent(e) {
    const button = ({ 0: "left", 1: "middle", 2: "right", 3: "back", 4: "forward" })[e.button];
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

// --- Derived ----------------------------------------------------------
function derive() {
    const total = state.steps.length;
    const doneCount = Math.max(0, state.activePos - 1);
    return { total, doneCount, pct: total ? Math.round(doneCount / total * 100) : 0 };
}

// --- Render -----------------------------------------------------------
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

    // The panel is a fixed-height flex column only in the expanded steps
    // view; the "steps-expanded" class arms the flex chain (see CSS).
    panel.classList.toggle("steps-expanded", state.view === "steps" && !state.collapsed);

    applyLayout();
    wireEvents();
    renderModal();
    scrollToCurrent();
}

// openConfirm shows the glass confirmation modal. onConfirm runs on accept;
// danger styles the primary button red (all current uses are destructive).
function openConfirm({ message, confirmLabel, onConfirm }) {
    state.confirm = { message, confirmLabel, onConfirm };
    render();
}
function closeConfirm() {
    if (!state.confirm) return;
    state.confirm = null;
    render();
}

// renderModal paints (or clears) the persistent #modal overlay from
// state.confirm and wires its buttons + backdrop. Empty host = no modal.
function renderModal() {
    const c = state.confirm;
    if (!c) { modalHost.innerHTML = ""; return; }
    modalHost.innerHTML = `
        <div class="modal-card" id="modalCard">
            <div class="modal-title">${t("confirmTitle")}</div>
            <div class="modal-msg">${esc(c.message)}</div>
            <div class="modal-actions">
                <div class="modal-btn ghost" id="modalCancel">${t("cancel")}</div>
                <div class="modal-btn danger" id="modalConfirm">${esc(c.confirmLabel)}</div>
            </div>
        </div>`;
    // Swallow mousedown so clicks don't start a window drag underneath.
    modalHost.addEventListener("mousedown", stopProp);
    // Backdrop click (outside the card) cancels.
    modalHost.onclick = e => { if (e.target === modalHost) closeConfirm(); };
    document.getElementById("modalCancel").addEventListener("click", closeConfirm);
    document.getElementById("modalConfirm").addEventListener("click", () => {
        const fn = c.onConfirm;
        closeConfirm();
        if (fn) fn();
    });
}

function renderPicker() {
    // Two levels: pick a game, then a chapter. The folder structure already
    // groups configs by game (configs/<game>/...), surfaced here via c.game.
    let kicker, rowsHTML;
    if (state.pickerGame) {
        const chapters = state.configs
            .filter(c => c.game === state.pickerGame)
            .sort((a, b) => (a.chapter || 0) - (b.chapter || 0));
        kicker = esc(state.pickerGame);
        const rows = chapters.map(c => `
            <div class="picker-row" data-path="${esc(c.path)}" data-embedded="true">
                <div class="picker-row-main">
                    ${c.chapter ? `<div class="picker-game mono">${esc(tf("chapterLabel", c.chapter))}</div>` : ''}
                    <div class="picker-title">${esc(c.title)}</div>
                    ${c.author ? `<div class="picker-meta mono">${esc(tf("by", c.author))}</div>` : ''}
                </div>
                <div class="picker-clear" data-clear-chapter="${esc(c.path)}" title="${t("clearProgress")}">⟲</div>
            </div>`).join('');
        rowsHTML = `<div class="picker-back mono" id="pickerBack">${t("backToGames")}</div>` +
            (rows || `<div class="picker-empty">${t("noChapters")}</div>`);
    } else {
        kicker = t("chooseGame");
        const games = [...new Set(state.configs.map(c => c.game))].sort();
        rowsHTML = games.length
            ? games.map(g => {
                const n = state.configs.filter(c => c.game === g).length;
                return `<div class="picker-row" data-game="${esc(g)}">
                    <div class="picker-row-main">
                        <div class="picker-title">${esc(g)}</div>
                        <div class="picker-meta mono">${esc(tf("chaptersCount", n))}</div>
                    </div>
                    <div class="picker-clear" data-clear-game="${esc(g)}" title="${t("clearProgress")}">⟲</div>
                </div>`;
            }).join('')
            : `<div class="picker-empty">${t("noConfigs")}</div>`;
    }

    const loadingBadge = state.pickerLoading
        ? `<div class="picker-loading mono">${esc(tf("downloading", state.pickerLoading))}</div>`
        : '';

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
            ${state.pickerError ? `<div class="picker-error">${esc(state.pickerError)}</div>` : ''}
            ${loadingBadge}
        </div>`;
}

// rowHTML renders one checklist item (step or branch decision).
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

// buildChecklist groups consecutive steps by section. Named sections become
// collapsible groups; a fully-done section gets a green "abgeschlossen" badge
// and collapses by default (until the user toggles it). Unsectioned steps
// render as loose rows. Open state defaults to "open while in progress,
// closed once done", overridable per section via state.sectionOverride.
function buildChecklist() {
    const groups = [];
    let cur = null;
    state.steps.forEach((s, i) => {
        const sec = s.section || "";
        if (!cur || cur.section !== sec) { cur = { section: sec, items: [] }; groups.push(cur); }
        cur.items.push({ s, i });
    });
    return groups.map(g => {
        const rows = g.items.map(({ s, i }) => rowHTML(s, i)).join("");
        if (!g.section) return rows; // unsectioned steps render loose
        const allDone = g.items.every(({ i }) => (i + 1) < state.activePos);
        const ov = state.sectionOverride[g.section];
        const open = ov === undefined ? !allDone : ov;
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
    // The now-card renders from the LIVE current item (state.current), not the
    // steps array — the array's per-item isLast/Done flags are only accurate
    // for the engine's position at fetch time, which breaks the hand-off button.
    const current = state.current || state.steps[state.activePos - 1] || null;

    let nowHTML;
    if (!current) {
        nowHTML = `<div class="done-card">${t("noSteps")}</div>`;
    } else if (current.isChoice) {
        const opts = (current.options || []).map(o => {
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
        const handoff = (current.isLast && state.nextFile)
            ? `<button class="next-file-btn" id="nextFileBtn">${esc(tf("nextFile", state.nextFile))}</button>` : "";
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

// settingsBodyHTML builds the settings panel body (hotkey rebinding, opacity,
// theme, language). Independent of any loaded walkthrough, so it serves the
// settings view whether opened from the picker or from an active walkthrough.
function settingsBodyHTML() {
    return `
        <div class="settings-view">
            <div class="settings-kicker mono">${t("shortcuts")}</div>
            ${HOTKEY_ACTIONS.map(a => {
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
                    value="${Math.round((state.settings ? (state.settings.opacity ?? 1) : 1) * 100)}">
            </div>
            <div class="setting-row">
                <span class="setting-name">${t("design")}</span>
                <span class="theme-switch">${THEMES.map(th => {
                    const active = (state.settings && state.settings.theme || "dark") === th.key;
                    return `<button class="theme-opt${active ? " active" : ""}" data-theme="${th.key}">${esc(t(th.labelKey))}</button>`;
                }).join("")}</span>
            </div>
            <div class="setting-row">
                <span class="setting-name">${t("language")}</span>
                <span class="theme-switch">${LANGUAGES.map(l => {
                    const active = lang() === l.key;
                    return `<button class="lang-opt${active ? " active" : ""}" data-lang="${l.key}">${esc(l.label)}</button>`;
                }).join("")}</span>
            </div>
            <div class="settings-error">${esc(state.settingsError)}</div>
            <div class="settings-hint">${t("settingsHint")}</div>
            <button class="settings-reset" id="resetSettings">${t("resetSettings")}</button>
        </div>`;
}

// renderSettings draws the settings view with its own header. It can be reached
// from the picker (no walkthrough loaded) or from an active walkthrough — the
// header shows the walkthrough meta + a home button only when one is loaded.
function renderSettings() {
    const loaded = state.settingsFrom === "steps"; // opened from an active walkthrough
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
    // Fixed panel height ONLY in the expanded steps view: the now-card keeps
    // its natural size and the scrollable list flexes to fill the rest, so
    // the window stays a constant size across steps (no jumping). Picker,
    // settings and the collapsed view shrink-wrap to their content (auto).
    const expanded = state.view === "steps" && !state.collapsed;
    panel.style.height = expanded ? state.panelHeight + "px" : "auto";
}

// scrollToCurrent brings the active step to the top of the list when the
// position actually changes (advancing/jumping), so done steps scroll off the
// top. Manual scrolling between navigations is left alone — render() isn't
// called then, and the guard avoids re-yanking on cosmetic re-renders (theme,
// collapse). The current row is never inside a collapsed (done) section.
function scrollToCurrent() {
    if (state.view !== "steps" || state.collapsed) return;
    if (state.lastScrolledPos === state.activePos) return;
    const cl = panel.querySelector(".checklist");
    const row = cl && cl.querySelector(".row.current");
    if (cl && row) {
        // Leave a little room above the current step so the most recently
        // checked-off step (or its section header) stays in view — feels
        // calmer than slamming the current row to the very top.
        const PEEK = 40;
        const r = row.getBoundingClientRect(), c = cl.getBoundingClientRect();
        cl.scrollTop = Math.max(0, cl.scrollTop + (r.top - c.top) - PEEK);
        state.lastScrolledPos = state.activePos;
    }
}

// --- Interactions -----------------------------------------------------
function gotoIndex(index) {
    App.Goto(index).then(info => { state.activePos = info.current; state.current = info; render(); });
}

// downloadGame fetches all chapters of a game into the cache, then drills
// into its chapter list. On failure it still drills in (already-cached
// chapters remain playable) and surfaces the error.
function downloadGame(game) {
    state.pickerLoading = game;
    state.pickerError = "";
    render();
    App.DownloadGame(game).then(() => {
        state.pickerLoading = "";
        state.pickerGame = game;
        render();
    }).catch(err => {
        state.pickerLoading = "";
        state.pickerGame = game;
        state.pickerError = String(err && err.message ? err.message : err);
        render();
    });
}

// clearGameProgress resets every chapter's saved progress for a game (after
// confirmation). The picker shows no progress state, so there's nothing to
// re-render on success beyond clearing any prior error.
function clearGameProgress(game) {
    openConfirm({
        message: tf("confirmClearGame", game),
        confirmLabel: t("confirmReset"),
        onConfirm: () => {
            state.pickerError = "";
            App.ClearGameProgress(game).then(() => render()).catch(err => {
                state.pickerError = String(err && err.message ? err.message : err);
                render();
            });
        },
    });
}

// clearChapterProgress resets one chapter's saved progress (no confirmation —
// it's a single, easily-redone reset).
function clearChapterProgress(path) {
    state.pickerError = "";
    App.ClearChapterProgress(path).then(() => render()).catch(err => {
        state.pickerError = String(err && err.message ? err.message : err);
        render();
    });
}

// clearCache empties the on-disk config cache (after confirmation) and returns
// to the game list, where games are re-downloaded on demand.
function clearCache() {
    openConfirm({
        message: t("confirmClearCache"),
        confirmLabel: t("confirmDelete"),
        onConfirm: () => {
            state.pickerError = "";
            App.ClearCache().then(() => {
                state.pickerGame = null;
                render();
            }).catch(err => {
                state.pickerError = String(err && err.message ? err.message : err);
                render();
            });
        },
    });
}

// resetSettings restores all settings to their defaults and re-renders so
// theme/language/opacity reflect the reset immediately. Per spec this needs
// no confirmation.
function resetSettings() {
    App.ResetSettings().then(s => {
        state.settings = s;
        applyOpacity(s.opacity ?? 1);
        applyTheme(s.theme);
        save();
        state.settingsError = "";
        render();
    }).catch(err => {
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
        state.sectionOverride = {}; // fresh walkthrough → fresh section state
        state.lastScrolledPos = null;
        state.view = "steps";
        render();
    }).catch(err => {
        state.pickerError = String(err && err.message ? err.message : err);
        render();
    });
}

// reloadSteps re-fetches the sequence after it changes shape (a choice
// reveals/hides gated steps; a `next:` hand-off swaps the whole walkthrough).
function reloadSteps() {
    return Promise.all([App.Steps(), App.CurrentStep(), App.NextFile()])
        .then(([steps, current, nextFile]) => {
            state.steps = steps;
            state.current = current;
            state.activePos = current.current;
            state.nextFile = nextFile || "";
            state.lastScrolledPos = null; // re-target auto-scroll after reshape
            render();
        });
}

function chooseOption(key, value) {
    App.Choose(key, value).then(() => reloadSteps());
}

// refreshAfterSwap re-fetches everything after the active walkthrough is
// replaced (a `next:` hand-off swaps meta, steps and current all at once).
function refreshAfterSwap() {
    state.sectionOverride = {}; // hand-off to a new file → fresh section state
    return App.Meta().then(meta => { state.meta = meta; return reloadSteps(); });
}

function loadNext() {
    App.LoadNext().then(() => refreshAfterSwap())
        .catch(err => { console.error("[GoThrough] next hand-off failed:", err); });
}

function goHome() {
    App.UnloadConfig().then(() => {
        state.view = "picker";
        state.pickerGame = null;
        state.configs = [];
        return App.ListConfigs();
    }).then(configs => {
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
    closeBtn.addEventListener("click", () => { window.runtime?.Quit?.(); });

    const pickerSettingsBtn = state.view === "picker" && document.getElementById("settingsBtn");
    if (pickerSettingsBtn) {
        pickerSettingsBtn.addEventListener("mousedown", stopProp);
        pickerSettingsBtn.addEventListener("click", toggleSettings);
    }

    if (state.view === "picker") {
        panel.querySelectorAll(".picker-row").forEach(row => {
            row.addEventListener("click", () => {
                if (row.dataset.game) {        // level 1: download the game, then drill in
                    downloadGame(row.dataset.game);
                } else if (row.dataset.path) { // level 2: load the chapter (from cache)
                    loadConfig(row.dataset.path, row.dataset.embedded === "true");
                }
            });
        });
        panel.querySelectorAll(".picker-clear").forEach(btn => {
            btn.addEventListener("mousedown", stopProp);
            btn.addEventListener("click", e => {
                e.stopPropagation(); // don't trigger the row's load/download
                if (btn.dataset.clearGame) clearGameProgress(btn.dataset.clearGame);
                else if (btn.dataset.clearChapter) clearChapterProgress(btn.dataset.clearChapter);
            });
        });
        const back = document.getElementById("pickerBack");
        if (back) {
            back.addEventListener("mousedown", stopProp);
            back.addEventListener("click", () => { state.pickerGame = null; render(); });
        }
        const browse = document.getElementById("pickerBrowse");
        if (browse) {
            browse.addEventListener("mousedown", stopProp);
            browse.addEventListener("click", () => {
                App.OpenBrowse().then(path => {
                    if (path) loadConfig(path, false);
                });
            });
        }
        const clearCacheBtn = document.getElementById("pickerClearCache");
        if (clearCacheBtn) {
            clearCacheBtn.addEventListener("mousedown", stopProp);
            clearCacheBtn.addEventListener("click", clearCache);
        }
        return; // no further event wiring needed in picker
    }

    // steps / settings view
    const stepsBar = document.getElementById("stepsBar");
    if (stepsBar) {
        stepsBar.addEventListener("click", () => { state.collapsed = !state.collapsed; render(); });
    }

    panel.querySelectorAll(".section-head").forEach(head => {
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

    panel.querySelectorAll(".row").forEach(row =>
        row.addEventListener("click", () => gotoIndex(Number(row.dataset.index))));

    panel.querySelectorAll(".choice-opt").forEach(btn => {
        btn.addEventListener("mousedown", stopProp);
        btn.addEventListener("click", () => chooseOption(btn.dataset.key, btn.dataset.value));
    });

    const nextBtn = document.getElementById("nextFileBtn");
    if (nextBtn) {
        nextBtn.addEventListener("mousedown", stopProp);
        nextBtn.addEventListener("click", loadNext);
    }

    panel.querySelectorAll(".combo").forEach(combo => {
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
            App.SaveOpacity(v).then(s => { state.settings = s; });
        });
    }

    panel.querySelectorAll(".theme-opt").forEach(btn => {
        btn.addEventListener("mousedown", stopProp);
        btn.addEventListener("click", () => selectTheme(btn.dataset.theme));
    });

    panel.querySelectorAll(".lang-opt").forEach(btn => {
        btn.addEventListener("mousedown", stopProp);
        btn.addEventListener("click", () => selectLanguage(btn.dataset.lang));
    });

    const resetBtn = document.getElementById("resetSettings");
    if (resetBtn) {
        resetBtn.addEventListener("mousedown", stopProp);
        resetBtn.addEventListener("click", resetSettings);
    }
}

// toggleSection flips a named section's open/closed state, recording an
// explicit override (so a done section the user opened stays open, and vice
// versa). The effective open state is recomputed exactly as buildChecklist
// does — open while in progress, closed once done, unless overridden.
function toggleSection(name) {
    if (!name) return;
    const items = state.steps.filter(s => (s.section || "") === name);
    const allDone = items.length > 0 &&
        state.steps.every((s, i) => (s.section || "") !== name || (i + 1) < state.activePos);
    const ov = state.sectionOverride[name];
    const open = ov === undefined ? !allDone : ov;
    state.sectionOverride[name] = !open;
    render();
}

// selectLanguage persists the interface language and re-renders (English is
// the default; German is the alternative). On a save error the message is
// surfaced and the previous language stays in effect.
function selectLanguage(l) {
    if (l === lang()) return;
    App.SaveLanguage(l).then(s => { state.settings = s; render(); })
        .catch(err => {
            state.settingsError = String(err && err.message ? err.message : err);
            render();
        });
}

// selectTheme applies a theme immediately (so it feels instant) and persists
// it; on a save error the previous theme is restored.
function selectTheme(theme) {
    const prev = state.settings ? state.settings.theme : "dark";
    applyTheme(theme);
    App.SaveTheme(theme).then(s => { state.settings = s; save(); render(); })
        .catch(err => {
            applyTheme(prev);
            state.settingsError = String(err && err.message ? err.message : err);
            render();
        });
}

// --- Settings panel + hotkey rebinding --------------------------------
function toggleSettings() {
    if (state.view === "settings") {
        state.view = state.settingsFrom || "steps";
    } else {
        state.settingsFrom = state.view; // "picker" or "steps" — return here on close
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
    window.go.overlay.App.SaveHotkeys(hotkeys).then(saved => {
        state.settings = saved;
        state.capturing = null;
        state.settingsError = "";
        render();
    }).catch(err => {
        state.capturing = null;
        state.settingsError = String(err && err.message ? err.message : err);
        render();
    });
}

function onCaptureKeydown(e) {
    if (state.confirm) {
        if (e.key === "Escape") { e.preventDefault(); e.stopPropagation(); closeConfirm(); }
        else if (e.key === "Enter") {
            e.preventDefault(); e.stopPropagation();
            const fn = state.confirm.onConfirm;
            closeConfirm();
            if (fn) fn();
        }
        return;
    }
    if (!state.capturing) return;
    e.preventDefault();
    e.stopPropagation();
    if (e.key === "Escape") { state.capturing = null; render(); return; }
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
window.addEventListener("contextmenu", e => { if (state.capturing) e.preventDefault(); }, true);

function toggleLock() {
    state.locked = !state.locked;
    save();
    render();
}

function stopProp(e) { e.stopPropagation(); }

// --- Resize -----------------------------------------------------------
let drag = null;

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
    if (drag) { drag = null; mover.classList.remove("dragging"); save(); }
}

resizeHandle.addEventListener("mousedown", startResize);
window.addEventListener("mousemove", onMove);
window.addEventListener("mouseup", onUp);

// --- Window drag ------------------------------------------------------
let winDrag = null;

async function startWindowDrag(e) {
    if (state.locked || e.button !== 0) return;
    if (e.target.closest(".lock-btn, .icon-btn, .close-btn, .picker-browse, .picker-clearcache, .picker-row, .picker-back")) return;
    const rt = window.runtime;
    if (!rt || !rt.WindowGetPosition) return;
    e.preventDefault();
    const [pos, size, screens] = await Promise.all([
        rt.WindowGetPosition(), rt.WindowGetSize(), rt.ScreenGetAll(),
    ]);
    const scr = (screens || []).find(s => s.isPrimary) || (screens || [])[0];
    const sw = scr ? (scr.size?.width || scr.width) : Infinity;
    const sh = scr ? (scr.size?.height || scr.height) : Infinity;
    winDrag = {
        sx: e.screenX, sy: e.screenY, ox: pos.x, oy: pos.y,
        maxX: isFinite(sw) ? Math.max(0, sw - size.w) : Infinity,
        maxY: isFinite(sh) ? Math.max(0, sh - size.h) : Infinity,
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

// Persist the position only once, at the end of a drag (not on every
// mousemove), so the overlay reopens where the user left it. lastX is unset
// if the window never actually moved, so a plain click won't write.
function onWinUp() {
    if (winDrag && winDrag.lastX != null) {
        App.SaveWindowPos?.(winDrag.lastX, winDrag.lastY);
    }
    winDrag = null;
}

window.addEventListener("mousemove", onWinMove);
window.addEventListener("mouseup", onWinUp);

// --- FitWindow --------------------------------------------------------
// In the expanded steps view the window size is pinned to our own fixed
// dimensions rather than the measured panel — this stops the OS window from
// jittering as the now-card grows/shrinks per step (sub-pixel reflows,
// appearing scrollbars). Picker/settings/collapsed shrink-wrap to content.
function fitWindow() {
    const expanded = state.view === "steps" && !state.collapsed;
    let w, h;
    if (expanded) {
        w = state.width;
        h = state.panelHeight;
    } else {
        const r = panel.getBoundingClientRect();
        w = Math.ceil(r.width); h = Math.ceil(r.height);
    }
    if (w > 0 && h > 0) window.go?.overlay?.App?.FitWindow?.(w, h);
}
new ResizeObserver(fitWindow).observe(panel);

// --- Boot -------------------------------------------------------------
async function init() {
    // Apply the cached theme synchronously first, so the very first paint of
    // the HUD already uses the right colours (Settings() below is async).
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
                App.Meta(), App.Steps(), App.CurrentStep(), App.NextFile(),
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
        window.runtime.EventsOn("step:changed", info => {
            state.activePos = info.current;
            state.current = info;
            render();
        });
        // A hotkey hand-off swapped the whole walkthrough — reload everything.
        window.runtime.EventsOn("config:changed", () => { refreshAfterSwap(); });
        // The background catalog refresh finished — adopt the fresh list so
        // the picker reflects any newly-added games/chapters.
        window.runtime.EventsOn("configs:remote", entries => {
            if (Array.isArray(entries) && entries.length) {
                state.configs = entries;
                if (state.view === "picker") render();
            }
        });
    }
}
init();
