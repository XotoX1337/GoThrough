(()=>{var u=window.go.overlay.App,N="gt-overlay-v7",B=(()=>{try{return JSON.parse(localStorage.getItem(N)||"{}")}catch{return{}}})(),O=window.screen&&window.screen.availHeight||920,W=Math.round(O*.382),J=380,Z=Math.min(O,W+300),t={view:"picker",configs:[],pickerGame:null,pickerLoading:"",pickerError:"",meta:{game:"",title:""},steps:[],current:null,activePos:1,nextFile:"",sectionOverride:{},lastScrolledPos:null,settings:null,capturing:null,confirm:null,settingsError:"",settingsFrom:"picker",collapsed:!1,locked:!0,width:B.width||J,panelHeight:B.panelHeight||Z},ee=240,te=680,ne=240,se=O,w=document.getElementById("mover"),f=document.getElementById("panel"),ie=document.getElementById("resize"),b=document.getElementById("modal");w.style.setProperty("--gt-nowh",W+"px");var R='<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="6 9 12 15 18 9"></polyline></svg>',j='<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"></path><polyline points="9 22 9 12 15 12 15 22"></polyline></svg>',G='<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="3"></circle><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"></path></svg>';function x(){try{localStorage.setItem(N,JSON.stringify({width:t.width,panelHeight:t.panelHeight,theme:t.settings?t.settings.theme:B.theme}))}catch{}}function c(e){return String(e??"").replace(/[&<>"]/g,n=>({"&":"&amp;","<":"&lt;",">":"&gt;",'"':"&quot;"})[n])}function E(e){return c(e).replace(/\*\*([^*]+)\*\*/g,"<strong>$1</strong>").replace(/\*([^*]+)\*/g,"<em>$1</em>")}function oe(e){let n=String(e??"").split(`
`),s=[],i=!1;for(let a of n){let r=a.match(/^\s*-\s+(.*)$/);r?(i||(s.push("<ul>"),i=!0),s.push("<li>"+E(r[1])+"</li>")):(i&&(s.push("</ul>"),i=!1),a.trim()&&s.push("<p>"+E(a)+"</p>"))}return i&&s.push("</ul>"),s.join("")}var P={en:{adjust:"Adjust",lock:"Lock",close:"Close",openFile:"Open file…",chooseGame:"Choose game",backToGames:"← Games",chapterLabel:e=>"Chapter "+e,chaptersCount:e=>e+(e===1?" chapter":" chapters"),by:e=>"by "+e,noChapters:"No chapters found.",noConfigs:"No configs found.",downloading:e=>"Downloading "+e+"…",clearProgress:"Reset progress",clearCache:"Clear cache",resetSettings:"Reset to defaults",confirmClearCache:"Delete all downloaded configs from the cache?",confirmClearGame:e=>"Reset all saved progress for "+e+"?",confirmTitle:"Are you sure?",cancel:"Cancel",confirmDelete:"Delete",confirmReset:"Reset",switchWalkthrough:"Switch walkthrough",settingsTitle:"Settings",closeOverlay:"Close overlay",steps:"Steps",noSteps:"No steps loaded",now:"Now",decision:"Decision",stepCount:(e,n)=>"Step "+e+" / "+n,chosen:"chosen",nextFile:e=>"Continue ➜ "+e,optional:"optional",quests:"Quests",completed:"✓ completed",received:"received",warning:"Caution",info:"Info",hints:"Hints",shortcuts:"Shortcuts",pressKey:"Press key/mouse…",opacity:"Opacity",design:"Theme",language:"Language",settingsHint:"Click a shortcut, then press a key or mouse combo. Esc cancels.",hkNext:"Next step",hkPrev:"Previous step",hkToggle:"Toggle overlay",hkFocus:"Mouse into overlay",hkQuit:"Close overlay",themeDark:"Dark",themeLight:"Light",themeContrast:"Contrast",mouseL:"Mouse L",mouseM:"Mouse M",mouseR:"Mouse R",mouse4:"Mouse 4",mouse5:"Mouse 5"},de:{adjust:"Anpassen",lock:"Fixieren",close:"Schließen",openFile:"Datei öffnen…",chooseGame:"Spiel wählen",backToGames:"← Spiele",chapterLabel:e=>"Kapitel "+e,chaptersCount:e=>e+" Kapitel",by:e=>"von "+e,noChapters:"Keine Kapitel gefunden.",noConfigs:"Keine Configs gefunden.",downloading:e=>"Lade "+e+"…",clearProgress:"Fortschritt zurücksetzen",clearCache:"Cache leeren",resetSettings:"Auf Standard zurücksetzen",confirmClearCache:"Alle heruntergeladenen Configs aus dem Cache löschen?",confirmClearGame:e=>"Allen gespeicherten Fortschritt für "+e+" zurücksetzen?",confirmTitle:"Bist du sicher?",cancel:"Abbrechen",confirmDelete:"Löschen",confirmReset:"Zurücksetzen",switchWalkthrough:"Anderes Walkthrough",settingsTitle:"Einstellungen",closeOverlay:"Overlay schließen",steps:"Schritte",noSteps:"Keine Schritte geladen",now:"Jetzt",decision:"Entscheidung",stepCount:(e,n)=>"Schritt "+e+" / "+n,chosen:"gewählt",nextFile:e=>"Weiter ➜ "+e,optional:"optional",quests:"Quests",completed:"✓ abgeschlossen",received:"erhalten",warning:"Achtung",info:"Info",hints:"Hinweise",shortcuts:"Tastenkürzel",pressKey:"Taste/Maus drücken…",opacity:"Transparenz",design:"Design",language:"Sprache",settingsHint:"Klick auf ein Kürzel, dann Tasten- oder Maustasten-Kombination drücken. Esc bricht ab.",hkNext:"Nächster Schritt",hkPrev:"Vorheriger Schritt",hkToggle:"Overlay ein/aus",hkFocus:"Maus ins Overlay",hkQuit:"Overlay schließen",themeDark:"Dunkel",themeLight:"Hell",themeContrast:"Kontrast",mouseL:"Maus L",mouseM:"Maus M",mouseR:"Maus R",mouse4:"Maus 4",mouse5:"Maus 5"}},ae=[{key:"en",label:"English"},{key:"de",label:"Deutsch"}];function M(){let e=t.settings&&t.settings.language;return P[e]?e:"en"}function o(e){let n=P[M()][e];return n??(P.en[e]!=null?P.en[e]:e)}function $(e,...n){let s=o(e);return typeof s=="function"?s(...n):s}function re(e){if(!e||!e.length)return"";let n=e.map(s=>{let i=s.status==="completed"?`<span class="quest-status done">${o("completed")}</span>`:s.status==="received"?`<span class="quest-status">${o("received")}</span>`:"",a=s.note?`<div class="quest-note">${E(s.note)}</div>`:"";return`<div class="quest ${c(s.status||"")}">
                    <div class="quest-line"><span class="quest-name">${c(s.name)}</span> ${i}</div>${a}
                </div>`}).join("");return`<div class="info-block"><div class="info-head mono">${o("quests")}</div>${n}</div>`}function H(e,n,s){if(!e||!e.length)return"";let i=e.map(a=>`<li>${E(a)}</li>`).join("");return`<div class="info-block ${s||""}"><div class="info-head mono">${n}</div><ul class="info-list">${i}</ul></div>`}function A(e,n){if(!n)return"";let s=e==="hint"?"":`<span class="task-callout-tag mono">${o(e)}</span>`;return`<span class="task-callout ${e}">${s}${E(n)}</span>`}function ce(e){return!e||!e.length?"":`<ul class="task-list">${e.map(s=>`<li class="task">${E(s.text)}${A("warning",s.warning)}${A("info",s.info)}${A("hint",s.hint)}</li>`).join("")}</ul>`}function T(e){w.style.setProperty("--gt-opacity",e!=null&&e>0?e:1)}var z=[{key:"dark",labelKey:"themeDark"},{key:"light",labelKey:"themeLight"},{key:"contrast",labelKey:"themeContrast"}];function L(e){document.documentElement.dataset.theme=z.some(n=>n.key===e)?e:"dark"}var le=[{key:"next",labelKey:"hkNext"},{key:"prev",labelKey:"hkPrev"},{key:"toggleHide",labelKey:"hkToggle"},{key:"focusOverlay",labelKey:"hkFocus"},{key:"quit",labelKey:"hkQuit"}],de={ctrl:"Ctrl",alt:"Alt",shift:"Shift",win:"Win"},ue={left:"←",right:"→",up:"↑",down:"↓",space:"Space",return:"Enter",escape:"Esc",delete:"Del",tab:"Tab"},F={left:"mouseL",middle:"mouseM",right:"mouseR",back:"mouse4",x1:"mouse4",forward:"mouse5",x2:"mouse5"};function pe(e){if(!e)return"—";let n;if(e.button)n=F[e.button]?o(F[e.button]):e.button;else if(e.key)n=ue[e.key]||e.key.toUpperCase();else return"—";let s=(e.mods||[]).map(i=>de[i]||i);return s.push(n),s.join(" + ")}function ge(e){if(["Control","Alt","Shift","Meta"].includes(e.key))return null;let n=null,s=e.code;return/^Key[A-Z]$/.test(s)?n=s.slice(3).toLowerCase():/^Digit[0-9]$/.test(s)?n=s.slice(5):/^F([1-9]|1[0-2])$/.test(s)?n=s.toLowerCase():n={ArrowLeft:"left",ArrowRight:"right",ArrowUp:"up",ArrowDown:"down",Space:"space",Enter:"return",Escape:"escape",Delete:"delete",Tab:"tab"}[s]||null,n?{mods:q(e),key:n}:null}function me(e){let n={0:"left",1:"middle",2:"right",3:"back",4:"forward"}[e.button];return n?{mods:q(e),button:n}:null}function q(e){let n=[];return e.ctrlKey&&n.push("ctrl"),e.altKey&&n.push("alt"),e.shiftKey&&n.push("shift"),e.metaKey&&n.push("win"),n}function ve(){let e=t.steps.length,n=Math.max(0,t.activePos-1);return{total:e,doneCount:n,pct:e?Math.round(n/e*100):0}}function l(){w.classList.toggle("editing",!t.locked),w.classList.toggle("collapsed",t.collapsed&&t.view==="steps"),T(t.settings?t.settings.opacity:1),document.documentElement.lang=M(),t.view==="picker"?fe():t.view==="settings"?Ee():ye(),f.classList.toggle("steps-expanded",t.view==="steps"&&!t.collapsed),Y(),He(),he(),Se()}function _({message:e,confirmLabel:n,onConfirm:s}){t.confirm={message:e,confirmLabel:n,onConfirm:s},l()}function C(){t.confirm&&(t.confirm=null,l())}function he(){let e=t.confirm;if(!e){b.innerHTML="";return}b.innerHTML=`
        <div class="modal-card" id="modalCard">
            <div class="modal-title">${o("confirmTitle")}</div>
            <div class="modal-msg">${c(e.message)}</div>
            <div class="modal-actions">
                <div class="modal-btn ghost" id="modalCancel">${o("cancel")}</div>
                <div class="modal-btn danger" id="modalConfirm">${c(e.confirmLabel)}</div>
            </div>
        </div>`,b.addEventListener("mousedown",m),b.onclick=n=>{n.target===b&&C()},document.getElementById("modalCancel").addEventListener("click",C),document.getElementById("modalConfirm").addEventListener("click",()=>{let n=e.onConfirm;C(),n&&n()})}function fe(){let e,n;if(t.pickerGame){let i=t.configs.filter(r=>r.game===t.pickerGame).sort((r,p)=>(r.chapter||0)-(p.chapter||0));e=c(t.pickerGame);let a=i.map(r=>`
            <div class="picker-row" data-path="${c(r.path)}" data-embedded="true">
                <div class="picker-row-main">
                    ${r.chapter?`<div class="picker-game mono">${c($("chapterLabel",r.chapter))}</div>`:""}
                    <div class="picker-title">${c(r.title)}</div>
                    ${r.author?`<div class="picker-meta mono">${c($("by",r.author))}</div>`:""}
                </div>
                <div class="picker-clear" data-clear-chapter="${c(r.path)}" title="${o("clearProgress")}">⟲</div>
            </div>`).join("");n=`<div class="picker-back mono" id="pickerBack">${o("backToGames")}</div>`+(a||`<div class="picker-empty">${o("noChapters")}</div>`)}else{e=o("chooseGame");let i=[...new Set(t.configs.map(a=>a.game))].sort();n=i.length?i.map(a=>{let r=t.configs.filter(p=>p.game===a).length;return`<div class="picker-row" data-game="${c(a)}">
                    <div class="picker-row-main">
                        <div class="picker-title">${c(a)}</div>
                        <div class="picker-meta mono">${c($("chaptersCount",r))}</div>
                    </div>
                    <div class="picker-clear" data-clear-game="${c(a)}" title="${o("clearProgress")}">⟲</div>
                </div>`}).join(""):`<div class="picker-empty">${o("noConfigs")}</div>`}let s=t.pickerLoading?`<div class="picker-loading mono">${c($("downloading",t.pickerLoading))}</div>`:"";f.innerHTML=`
        <div class="header" id="header">
            <div class="header-left">
                <div>
                    <div class="kicker mono">GoThrough</div>
                </div>
            </div>
            <div class="header-right">
                <div class="lock-btn mono" id="lockBtn">${t.locked?o("adjust"):o("lock")}</div>
                <div class="icon-btn" id="settingsBtn" title="${o("settingsTitle")}">${G}</div>
                <div class="close-btn" id="closeBtn" title="${o("close")}">✕</div>
            </div>
        </div>
        <div class="picker-view">
            <div class="picker-kicker mono">${e}</div>
            <div class="picker-list" id="pickerList">${n}</div>
            <div class="picker-actions">
                <div class="picker-browse mono" id="pickerBrowse">${o("openFile")}</div>
                <div class="picker-clearcache mono" id="pickerClearCache">${o("clearCache")}</div>
            </div>
            ${t.pickerError?`<div class="picker-error">${c(t.pickerError)}</div>`:""}
            ${s}
        </div>`}function ke(e,n){let s=n+1,i=s<t.activePos,a=s===t.activePos,r=e.isChoice?`⑂ ${c(e.title)}`:c(e.title),p=e.optional?`<span class="row-opt">${o("optional")}</span>`:"";return`
        <div class="row${i?" done":""}${a?" current":""}${e.isChoice?" is-choice":""}" data-index="${n}">
            <div class="mark">${i?'<span class="check">✓</span>':""}</div>
            <div class="row-main">
                <div class="row-label">${r}${p}</div>
            </div>
        </div>`}function we(){let e=[],n=null;return t.steps.forEach((s,i)=>{let a=s.section||"";(!n||n.section!==a)&&(n={section:a,items:[]},e.push(n)),n.items.push({s,i})}),e.map(s=>{let i=s.items.map(({s:k,i:d})=>ke(k,d)).join("");if(!s.section)return i;let a=s.items.every(({i:k})=>k+1<t.activePos),r=t.sectionOverride[s.section],p=r===void 0?!a:r,g=a?`<span class="section-done mono">${o("completed")}</span>`:"";return`<div class="section-group${p?"":" sec-collapsed"}">
            <div class="section-head mono" data-section="${c(s.section)}">
                <span class="section-chevron">${R}</span>
                <span>${c(s.section)}</span>
                ${g}
            </div>
            <div class="section-rows">${i}</div>
        </div>`}).join("")}function ye(){let{total:e,doneCount:n,pct:s}=ve(),i=t.current||t.steps[t.activePos-1]||null,a;if(!i)a=`<div class="done-card">${o("noSteps")}</div>`;else if(i.isChoice){let p=(i.options||[]).map(g=>{let k=i.selected&&g.value===i.selected;return`
            <button class="choice-opt${k?" chosen":""}" data-key="${c(i.choiceKey)}" data-value="${c(g.value)}">
                <span class="choice-opt-label">${c(g.label)}${k?` <span class="choice-opt-tag">${o("chosen")}</span>`:""}</span>
                ${g.description?`<span class="choice-opt-desc">${E(g.description)}</span>`:""}
            </button>`}).join("");a=`<div class="now-card choice-card">
               <div class="now-label mono">${o("decision")}</div>
               <div class="now-step">${c(i.title)}</div>
               <div class="now-meta mono">${c($("stepCount",t.activePos,e))}${i.section?" · "+c(i.section):""}</div>
               <div class="choice-options">${p}</div>
           </div>`}else{let p=i.optional?`<span class="opt-badge">${o("optional")}</span>`:"",g=i.isLast&&t.nextFile?`<button class="next-file-btn" id="nextFileBtn">${c($("nextFile",t.nextFile))}</button>`:"";a=`<div class="now-card">
               <div class="now-label mono">${o("now")}</div>
               <div class="now-step">${c(i.title)} ${p}</div>
               <div class="now-meta mono">${c($("stepCount",t.activePos,e))}${i.section?" · "+c(i.section):""}</div>
               ${i.description?`<div class="now-desc">${oe(i.description)}</div>`:""}
               ${ce(i.tasks)}
               ${H(i.warnings,o("warning"),"warnings")}
               ${H(i.infos,o("info"),"infos")}
               ${re(i.quests)}
               ${H(i.hints,o("hints"),"hints")}
               ${g}
           </div>`}let r=we();f.innerHTML=`
        <div class="header" id="header">
            <div class="header-left">
                <div>
                    <div class="kicker mono">${c(t.meta.game)}</div>
                    <div class="quest-title" id="questTitle">${c(t.meta.title)}</div>
                </div>
            </div>
            <div class="header-right">
                <div class="lock-btn mono" id="lockBtn">${t.locked?o("adjust"):o("lock")}</div>
                <div class="icon-btn" id="homeBtn" title="${o("switchWalkthrough")}">${j}</div>
                <div class="icon-btn" id="settingsBtn" title="${o("settingsTitle")}">${G}</div>
                <div class="close-btn" id="closeBtn" title="${o("closeOverlay")}">✕</div>
            </div>
        </div>
        ${a}
        <div class="steps-section">
            <div class="steps-bar" id="stepsBar">
                <span class="steps-bar-left">
                    <span class="steps-chevron">${R}</span>
                    <span class="progress-label mono">${o("steps")}</span>
                </span>
                <div class="progress-track"><div class="progress-fill" style="width:${s}%"></div></div>
                <span class="progress-count mono">${n} / ${e}</span>
            </div>
            <div class="collapsible">
                <div class="checklist">${r}</div>
            </div>
        </div>`}function $e(){return`
        <div class="settings-view">
            <div class="settings-kicker mono">${o("shortcuts")}</div>
            ${le.map(e=>{let n=t.settings&&t.settings.hotkeys[e.key],s=t.capturing===e.key;return`<div class="setting-row">
                    <span class="setting-name">${c(o(e.labelKey))}</span>
                    <span class="combo mono${s?" capturing":""}" data-action="${e.key}">${s?o("pressKey"):c(pe(n))}</span>
                </div>`}).join("")}
            <div class="setting-row">
                <span class="setting-name">${o("opacity")}</span>
                <input type="range" class="opacity-slider" id="opacitySlider"
                    min="10" max="100" step="5"
                    value="${Math.round((t.settings?t.settings.opacity??1:1)*100)}">
            </div>
            <div class="setting-row">
                <span class="setting-name">${o("design")}</span>
                <span class="theme-switch">${z.map(e=>`<button class="theme-opt${(t.settings&&t.settings.theme||"dark")===e.key?" active":""}" data-theme="${e.key}">${c(o(e.labelKey))}</button>`).join("")}</span>
            </div>
            <div class="setting-row">
                <span class="setting-name">${o("language")}</span>
                <span class="theme-switch">${ae.map(e=>`<button class="lang-opt${M()===e.key?" active":""}" data-lang="${e.key}">${c(e.label)}</button>`).join("")}</span>
            </div>
            <div class="settings-error">${c(t.settingsError)}</div>
            <div class="settings-hint">${o("settingsHint")}</div>
            <button class="settings-reset" id="resetSettings">${o("resetSettings")}</button>
        </div>`}function Ee(){let e=t.settingsFrom==="steps";f.innerHTML=`
        <div class="header" id="header">
            <div class="header-left">
                <div>
                    <div class="kicker mono">${e?c(t.meta.game):"GoThrough"}</div>
                    ${e?`<div class="quest-title">${c(t.meta.title)}</div>`:""}
                </div>
            </div>
            <div class="header-right">
                <div class="lock-btn mono" id="lockBtn">${t.locked?o("adjust"):o("lock")}</div>
                ${e?`<div class="icon-btn" id="homeBtn" title="${o("switchWalkthrough")}">${j}</div>`:""}
                <div class="icon-btn active" id="settingsBtn" title="${o("settingsTitle")}">${G}</div>
                <div class="close-btn" id="closeBtn" title="${o(e?"closeOverlay":"close")}">✕</div>
            </div>
        </div>
        ${$e()}`}function Y(){w.style.left="0px",w.style.top="0px",w.style.width=t.width+"px";let e=t.view==="steps"&&!t.collapsed;f.style.height=e?t.panelHeight+"px":"auto"}function Se(){if(t.view!=="steps"||t.collapsed||t.lastScrolledPos===t.activePos)return;let e=f.querySelector(".checklist"),n=e&&e.querySelector(".row.current");if(e&&n){let i=n.getBoundingClientRect(),a=e.getBoundingClientRect();e.scrollTop=Math.max(0,e.scrollTop+(i.top-a.top)-40),t.lastScrolledPos=t.activePos}}function be(e){u.Goto(e).then(n=>{t.activePos=n.current,t.current=n,l()})}function Ce(e){t.pickerLoading=e,t.pickerError="",l(),u.DownloadGame(e).then(()=>{t.pickerLoading="",t.pickerGame=e,l()}).catch(n=>{t.pickerLoading="",t.pickerGame=e,t.pickerError=String(n&&n.message?n.message:n),l()})}function Le(e){_({message:$("confirmClearGame",e),confirmLabel:o("confirmReset"),onConfirm:()=>{t.pickerError="",u.ClearGameProgress(e).then(()=>l()).catch(n=>{t.pickerError=String(n&&n.message?n.message:n),l()})}})}function Pe(e){t.pickerError="",u.ClearChapterProgress(e).then(()=>l()).catch(n=>{t.pickerError=String(n&&n.message?n.message:n),l()})}function Be(){_({message:o("confirmClearCache"),confirmLabel:o("confirmDelete"),onConfirm:()=>{t.pickerError="",u.ClearCache().then(()=>{t.pickerGame=null,l()}).catch(e=>{t.pickerError=String(e&&e.message?e.message:e),l()})}})}function xe(){u.ResetSettings().then(e=>{t.settings=e,T(e.opacity??1),L(e.theme),x(),t.settingsError="",l()}).catch(e=>{t.settingsError=String(e&&e.message?e.message:e),l()})}function K(e,n){t.pickerError="",u.LoadConfig(e,n).then(()=>Promise.all([u.Meta(),u.Steps(),u.CurrentStep(),u.NextFile()])).then(([s,i,a,r])=>{t.meta=s,t.steps=i,t.current=a,t.activePos=a.current,t.nextFile=r||"",t.sectionOverride={},t.lastScrolledPos=null,t.view="steps",l()}).catch(s=>{t.pickerError=String(s&&s.message?s.message:s),l()})}function X(){return Promise.all([u.Steps(),u.CurrentStep(),u.NextFile()]).then(([e,n,s])=>{t.steps=e,t.current=n,t.activePos=n.current,t.nextFile=s||"",t.lastScrolledPos=null,l()})}function Me(e,n){u.Choose(e,n).then(()=>X())}function Q(){return t.sectionOverride={},u.Meta().then(e=>(t.meta=e,X()))}function Te(){u.LoadNext().then(()=>Q()).catch(e=>{console.error("[GoThrough] next hand-off failed:",e)})}function Ie(){u.UnloadConfig().then(()=>(t.view="picker",t.pickerGame=null,t.configs=[],u.ListConfigs())).then(e=>{t.configs=e,l()})}function He(){document.getElementById("header").addEventListener("mousedown",ze);let e=document.getElementById("lockBtn");e.addEventListener("mousedown",m),e.addEventListener("click",Ne);let n=document.getElementById("closeBtn");n.addEventListener("mousedown",m),n.addEventListener("click",()=>{window.runtime?.Quit?.()});let s=t.view==="picker"&&document.getElementById("settingsBtn");if(s&&(s.addEventListener("mousedown",m),s.addEventListener("click",D)),t.view==="picker"){f.querySelectorAll(".picker-row").forEach(h=>{h.addEventListener("click",()=>{h.dataset.game?Ce(h.dataset.game):h.dataset.path&&K(h.dataset.path,h.dataset.embedded==="true")})}),f.querySelectorAll(".picker-clear").forEach(h=>{h.addEventListener("mousedown",m),h.addEventListener("click",V=>{V.stopPropagation(),h.dataset.clearGame?Le(h.dataset.clearGame):h.dataset.clearChapter&&Pe(h.dataset.clearChapter)})});let d=document.getElementById("pickerBack");d&&(d.addEventListener("mousedown",m),d.addEventListener("click",()=>{t.pickerGame=null,l()}));let S=document.getElementById("pickerBrowse");S&&(S.addEventListener("mousedown",m),S.addEventListener("click",()=>{u.OpenBrowse().then(h=>{h&&K(h,!1)})}));let I=document.getElementById("pickerClearCache");I&&(I.addEventListener("mousedown",m),I.addEventListener("click",Be));return}let i=document.getElementById("stepsBar");i&&i.addEventListener("click",()=>{t.collapsed=!t.collapsed,l()}),f.querySelectorAll(".section-head").forEach(d=>{d.addEventListener("click",()=>Ae(d.dataset.section))});let a=document.getElementById("homeBtn");a&&(a.addEventListener("mousedown",m),a.addEventListener("click",Ie));let r=document.getElementById("settingsBtn");r&&(r.addEventListener("mousedown",m),r.addEventListener("click",D)),f.querySelectorAll(".row").forEach(d=>d.addEventListener("click",()=>be(Number(d.dataset.index)))),f.querySelectorAll(".choice-opt").forEach(d=>{d.addEventListener("mousedown",m),d.addEventListener("click",()=>Me(d.dataset.key,d.dataset.value))});let p=document.getElementById("nextFileBtn");p&&(p.addEventListener("mousedown",m),p.addEventListener("click",Te)),f.querySelectorAll(".combo").forEach(d=>{d.addEventListener("mousedown",m),d.addEventListener("click",()=>Fe(d.dataset.action))});let g=document.getElementById("opacitySlider");g&&(g.addEventListener("mousedown",m),g.addEventListener("input",()=>{let d=Number(g.value)/100;T(d)}),g.addEventListener("change",()=>{let d=Number(g.value)/100;u.SaveOpacity(d).then(S=>{t.settings=S})})),f.querySelectorAll(".theme-opt").forEach(d=>{d.addEventListener("mousedown",m),d.addEventListener("click",()=>Ge(d.dataset.theme))}),f.querySelectorAll(".lang-opt").forEach(d=>{d.addEventListener("mousedown",m),d.addEventListener("click",()=>Oe(d.dataset.lang))});let k=document.getElementById("resetSettings");k&&(k.addEventListener("mousedown",m),k.addEventListener("click",xe))}function Ae(e){if(!e)return;let s=t.steps.filter(r=>(r.section||"")===e).length>0&&t.steps.every((r,p)=>(r.section||"")!==e||p+1<t.activePos),i=t.sectionOverride[e],a=i===void 0?!s:i;t.sectionOverride[e]=!a,l()}function Oe(e){e!==M()&&u.SaveLanguage(e).then(n=>{t.settings=n,l()}).catch(n=>{t.settingsError=String(n&&n.message?n.message:n),l()})}function Ge(e){let n=t.settings?t.settings.theme:"dark";L(e),u.SaveTheme(e).then(s=>{t.settings=s,x(),l()}).catch(s=>{L(n),t.settingsError=String(s&&s.message?s.message:s),l()})}function D(){t.view==="settings"?t.view=t.settingsFrom||"steps":(t.settingsFrom=t.view,t.view="settings"),t.capturing=null,t.settingsError="",l()}function Fe(e){t.capturing=e,t.settingsError="",l()}function U(e){let n=t.capturing,s={...t.settings.hotkeys,[n]:e};window.go.overlay.App.SaveHotkeys(s).then(i=>{t.settings=i,t.capturing=null,t.settingsError="",l()}).catch(i=>{t.capturing=null,t.settingsError=String(i&&i.message?i.message:i),l()})}function Ke(e){if(t.confirm){if(e.key==="Escape")e.preventDefault(),e.stopPropagation(),C();else if(e.key==="Enter"){e.preventDefault(),e.stopPropagation();let s=t.confirm.onConfirm;C(),s&&s()}return}if(!t.capturing)return;if(e.preventDefault(),e.stopPropagation(),e.key==="Escape"){t.capturing=null,l();return}let n=ge(e);n&&U(n)}function De(e){if(!t.capturing)return;e.preventDefault(),e.stopPropagation();let n=me(e);n&&U(n)}window.addEventListener("keydown",Ke,!0);window.addEventListener("mousedown",De,!0);window.addEventListener("contextmenu",e=>{t.capturing&&e.preventDefault()},!0);function Ne(){t.locked=!t.locked,x(),l()}function m(e){e.stopPropagation()}var y=null;function We(e){t.locked||(e.preventDefault(),e.stopPropagation(),y={sx:e.screenX,sy:e.screenY,ow:t.width,oh:t.panelHeight},w.classList.add("dragging"))}function Re(e){y&&(t.width=Math.max(ee,Math.min(y.ow-(e.screenX-y.sx),te)),t.panelHeight=Math.max(ne,Math.min(y.oh+(e.screenY-y.sy),se)),Y())}function je(){y&&(y=null,w.classList.remove("dragging"),x())}ie.addEventListener("mousedown",We);window.addEventListener("mousemove",Re);window.addEventListener("mouseup",je);var v=null;async function ze(e){if(t.locked||e.button!==0||e.target.closest(".lock-btn, .icon-btn, .close-btn, .picker-browse, .picker-clearcache, .picker-row, .picker-back"))return;let n=window.runtime;if(!n||!n.WindowGetPosition)return;e.preventDefault();let[s,i,a]=await Promise.all([n.WindowGetPosition(),n.WindowGetSize(),n.ScreenGetAll()]),r=(a||[]).find(k=>k.isPrimary)||(a||[])[0],p=r?r.size?.width||r.width:1/0,g=r?r.size?.height||r.height:1/0;v={sx:e.screenX,sy:e.screenY,ox:s.x,oy:s.y,maxX:isFinite(p)?Math.max(0,p-i.w):1/0,maxY:isFinite(g)?Math.max(0,g-i.h):1/0}}function qe(e){if(!v)return;let n=Math.max(0,Math.min(v.ox+(e.screenX-v.sx),v.maxX)),s=Math.max(0,Math.min(v.oy+(e.screenY-v.sy),v.maxY));v.lastX=Math.round(n),v.lastY=Math.round(s),window.runtime.WindowSetPosition(v.lastX,v.lastY)}function _e(){v&&v.lastX!=null&&u.SaveWindowPos?.(v.lastX,v.lastY),v=null}window.addEventListener("mousemove",qe);window.addEventListener("mouseup",_e);function Ye(){let e=t.view==="steps"&&!t.collapsed,n,s;if(e)n=t.width,s=t.panelHeight;else{let i=f.getBoundingClientRect();n=Math.ceil(i.width),s=Math.ceil(i.height)}n>0&&s>0&&window.go?.overlay?.App?.FitWindow?.(n,s)}new ResizeObserver(Ye).observe(f);async function Xe(){L(B.theme);try{let e=await u.Settings();if(t.settings=e,T(e.opacity??1),L(e.theme),await u.IsPicker())t.view="picker",t.configs=await u.ListConfigs();else{let[s,i,a,r]=await Promise.all([u.Meta(),u.Steps(),u.CurrentStep(),u.NextFile()]);t.meta=s,t.steps=i,t.current=a,t.activePos=a.current,t.nextFile=r||"",t.view="steps"}}catch(e){console.error("[GoThrough] init failed:",e)}l(),window.runtime&&window.runtime.EventsOn&&(window.runtime.EventsOn("step:changed",e=>{t.activePos=e.current,t.current=e,l()}),window.runtime.EventsOn("config:changed",()=>{Q()}),window.runtime.EventsOn("configs:remote",e=>{Array.isArray(e)&&e.length&&(t.configs=e,t.view==="picker"&&l())}))}Xe();})();
