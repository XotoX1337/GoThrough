(()=>{var d=window.go.overlay.App,N="gt-overlay-v7",B=(()=>{try{return JSON.parse(localStorage.getItem(N)||"{}")}catch{return{}}})(),G=window.screen&&window.screen.availHeight||920,R=Math.round(G*.382),Z=380,ee=Math.min(G,R+300),t={view:"picker",configs:[],pickerGame:null,pickerLoading:"",pickerError:"",meta:{game:"",title:""},steps:[],current:null,activePos:1,nextFile:"",sectionOverride:{},lastScrolledPos:null,settings:null,capturing:null,confirm:null,settingsError:"",settingsFrom:"picker",collapsed:!1,locked:!0,width:B.width||Z,panelHeight:B.panelHeight||ee},te=240,ne=680,se=240,ie=G,w=document.getElementById("mover"),h=document.getElementById("panel"),oe=document.getElementById("resize"),S=document.getElementById("modal");w.style.setProperty("--gt-nowh",R+"px");var j='<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="6 9 12 15 18 9"></polyline></svg>',q='<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"></path><polyline points="9 22 9 12 15 12 15 22"></polyline></svg>',F='<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="3"></circle><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"></path></svg>';function M(){try{localStorage.setItem(N,JSON.stringify({width:t.width,panelHeight:t.panelHeight,theme:t.settings?t.settings.theme:B.theme}))}catch{}}function r(e){return String(e??"").replace(/[&<>"]/g,n=>({"&":"&amp;","<":"&lt;",">":"&gt;",'"':"&quot;"})[n])}function E(e){return r(e).replace(/\*\*([^*]+)\*\*/g,"<strong>$1</strong>").replace(/\*([^*]+)\*/g,"<em>$1</em>")}function ae(e){let n=String(e??"").split(`
`),s=[],i=!1;for(let a of n){let c=a.match(/^\s*-\s+(.*)$/);c?(i||(s.push("<ul>"),i=!0),s.push("<li>"+E(c[1])+"</li>")):(i&&(s.push("</ul>"),i=!1),a.trim()&&s.push("<p>"+E(a)+"</p>"))}return i&&s.push("</ul>"),s.join("")}var x={en:{adjust:"Adjust",lock:"Lock",close:"Close",openFile:"Open file…",chooseGame:"Choose game",backToGames:"← Games",chapterLabel:e=>"Chapter "+e,chaptersCount:e=>e+(e===1?" chapter":" chapters"),by:e=>"by "+e,noChapters:"No chapters found.",noConfigs:"No configs found.",downloading:e=>"Downloading "+e+"…",clearProgress:"Reset progress",clearCache:"Clear cache",resetSettings:"Reset to defaults",confirmClearCache:"Delete all downloaded configs from the cache?",confirmClearGame:e=>"Reset all saved progress for "+e+"?",confirmTitle:"Are you sure?",cancel:"Cancel",confirmDelete:"Delete",confirmReset:"Reset",switchWalkthrough:"Switch walkthrough",settingsTitle:"Settings",closeOverlay:"Close overlay",steps:"Steps",noSteps:"No steps loaded",now:"Now",decision:"Decision",stepCount:(e,n)=>"Step "+e+" / "+n,chosen:"chosen",nextFile:e=>"Continue ➜ "+e,optional:"optional",quests:"Quests",completed:"✓ completed",received:"received",warning:"Caution",info:"Info",hints:"Hints",shortcuts:"Shortcuts",pressKey:"Press key/mouse…",opacity:"Opacity",design:"Theme",language:"Language",settingsHint:"Click a shortcut, then press a key or mouse combo. Esc cancels.",hkNext:"Next step",hkPrev:"Previous step",hkToggle:"Toggle overlay",hkFocus:"Mouse into overlay",hkQuit:"Close overlay",themeDark:"Dark",themeLight:"Light",themeContrast:"Contrast",mouseL:"Mouse L",mouseM:"Mouse M",mouseR:"Mouse R",mouse4:"Mouse 4",mouse5:"Mouse 5"},de:{adjust:"Anpassen",lock:"Fixieren",close:"Schließen",openFile:"Datei öffnen…",chooseGame:"Spiel wählen",backToGames:"← Spiele",chapterLabel:e=>"Kapitel "+e,chaptersCount:e=>e+" Kapitel",by:e=>"von "+e,noChapters:"Keine Kapitel gefunden.",noConfigs:"Keine Configs gefunden.",downloading:e=>"Lade "+e+"…",clearProgress:"Fortschritt zurücksetzen",clearCache:"Cache leeren",resetSettings:"Auf Standard zurücksetzen",confirmClearCache:"Alle heruntergeladenen Configs aus dem Cache löschen?",confirmClearGame:e=>"Allen gespeicherten Fortschritt für "+e+" zurücksetzen?",confirmTitle:"Bist du sicher?",cancel:"Abbrechen",confirmDelete:"Löschen",confirmReset:"Zurücksetzen",switchWalkthrough:"Anderes Walkthrough",settingsTitle:"Einstellungen",closeOverlay:"Overlay schließen",steps:"Schritte",noSteps:"Keine Schritte geladen",now:"Jetzt",decision:"Entscheidung",stepCount:(e,n)=>"Schritt "+e+" / "+n,chosen:"gewählt",nextFile:e=>"Weiter ➜ "+e,optional:"optional",quests:"Quests",completed:"✓ abgeschlossen",received:"erhalten",warning:"Achtung",info:"Info",hints:"Hinweise",shortcuts:"Tastenkürzel",pressKey:"Taste/Maus drücken…",opacity:"Transparenz",design:"Design",language:"Sprache",settingsHint:"Klick auf ein Kürzel, dann Tasten- oder Maustasten-Kombination drücken. Esc bricht ab.",hkNext:"Nächster Schritt",hkPrev:"Vorheriger Schritt",hkToggle:"Overlay ein/aus",hkFocus:"Maus ins Overlay",hkQuit:"Overlay schließen",themeDark:"Dunkel",themeLight:"Hell",themeContrast:"Kontrast",mouseL:"Maus L",mouseM:"Maus M",mouseR:"Maus R",mouse4:"Maus 4",mouse5:"Maus 5"}},re=[{key:"en",label:"English"},{key:"de",label:"Deutsch"}];function T(){let e=t.settings&&t.settings.language;return x[e]?e:"en"}function o(e){let n=x[T()][e];return n??(x.en[e]!=null?x.en[e]:e)}function $(e,...n){let s=o(e);return typeof s=="function"?s(...n):s}function ce(e){if(!e||!e.length)return"";let n=e.map(s=>{let i=s.status==="completed"?`<span class="quest-status done">${o("completed")}</span>`:s.status==="received"?`<span class="quest-status">${o("received")}</span>`:"",a=s.note?`<div class="quest-note">${E(s.note)}</div>`:"";return`<div class="quest ${r(s.status||"")}">
                    <div class="quest-line"><span class="quest-name">${r(s.name)}</span> ${i}</div>${a}
                </div>`}).join("");return`<div class="info-block"><div class="info-head mono">${o("quests")}</div>${n}</div>`}function I(e,n,s){if(!e||!e.length)return"";let i=e.map(a=>`<li>${E(a)}</li>`).join("");return`<div class="info-block ${s||""}"><div class="info-head mono">${n}</div><ul class="info-list">${i}</ul></div>`}function O(e,n){if(!n)return"";let s=e==="hint"?"":`<span class="task-callout-tag mono">${o(e)}</span>`;return`<span class="task-callout ${e}">${s}${E(n)}</span>`}function le(e){return!e||!e.length?"":`<ul class="task-list">${e.map(s=>`<li class="task">${E(s.text)}${O("warning",s.warning)}${O("info",s.info)}${O("hint",s.hint)}</li>`).join("")}</ul>`}function A(e){w.style.setProperty("--gt-opacity",e!=null&&e>0?e:1)}var z=[{key:"dark",labelKey:"themeDark"},{key:"light",labelKey:"themeLight"},{key:"contrast",labelKey:"themeContrast"}];function P(e){document.documentElement.dataset.theme=z.some(n=>n.key===e)?e:"dark"}var de=[{key:"next",labelKey:"hkNext"},{key:"prev",labelKey:"hkPrev"},{key:"toggleHide",labelKey:"hkToggle"},{key:"focusOverlay",labelKey:"hkFocus"},{key:"quit",labelKey:"hkQuit"}],ue={ctrl:"Ctrl",alt:"Alt",shift:"Shift",win:"Win"},pe={left:"←",right:"→",up:"↑",down:"↓",space:"Space",return:"Enter",escape:"Esc",delete:"Del",tab:"Tab"},K={left:"mouseL",middle:"mouseM",right:"mouseR",back:"mouse4",x1:"mouse4",forward:"mouse5",x2:"mouse5"};function ge(e){if(!e)return"—";let n;if(e.button)n=K[e.button]?o(K[e.button]):e.button;else if(e.key)n=pe[e.key]||e.key.toUpperCase();else return"—";let s=(e.mods||[]).map(i=>ue[i]||i);return s.push(n),s.join(" + ")}function me(e){if(["Control","Alt","Shift","Meta"].includes(e.key))return null;let n=null,s=e.code;return/^Key[A-Z]$/.test(s)?n=s.slice(3).toLowerCase():/^Digit[0-9]$/.test(s)?n=s.slice(5):/^F([1-9]|1[0-2])$/.test(s)?n=s.toLowerCase():n={ArrowLeft:"left",ArrowRight:"right",ArrowUp:"up",ArrowDown:"down",Space:"space",Enter:"return",Escape:"escape",Delete:"delete",Tab:"tab"}[s]||null,n?{mods:_(e),key:n}:null}function ve(e){let n={0:"left",1:"middle",2:"right",3:"back",4:"forward"}[e.button];return n?{mods:_(e),button:n}:null}function _(e){let n=[];return e.ctrlKey&&n.push("ctrl"),e.altKey&&n.push("alt"),e.shiftKey&&n.push("shift"),e.metaKey&&n.push("win"),n}function he(){let e=t.steps.length,n=Math.max(0,t.activePos-1);return{total:e,doneCount:n,pct:e?Math.round(n/e*100):0}}function l(){w.classList.toggle("editing",!t.locked),w.classList.toggle("collapsed",t.collapsed&&t.view==="steps"),A(t.settings?t.settings.opacity:1),document.documentElement.lang=T(),t.view==="picker"?ke():t.view==="settings"?be():$e(),h.classList.toggle("steps-expanded",t.view==="steps"&&!t.collapsed),U(),Ie(),fe(),Ce()}function Q({message:e,confirmLabel:n,onConfirm:s}){t.confirm={message:e,confirmLabel:n,onConfirm:s},l()}function L(){t.confirm&&(t.confirm=null,l())}function fe(){let e=t.confirm;if(!e){S.innerHTML="";return}S.innerHTML=`
        <div class="modal-card" id="modalCard">
            <div class="modal-title">${o("confirmTitle")}</div>
            <div class="modal-msg">${r(e.message)}</div>
            <div class="modal-actions">
                <div class="modal-btn ghost" id="modalCancel">${o("cancel")}</div>
                <div class="modal-btn danger" id="modalConfirm">${r(e.confirmLabel)}</div>
            </div>
        </div>`,S.addEventListener("mousedown",g),S.onclick=n=>{n.target===S&&L()},document.getElementById("modalCancel").addEventListener("click",L),document.getElementById("modalConfirm").addEventListener("click",()=>{let n=e.onConfirm;L(),n&&n()})}function ke(){let e,n;if(t.pickerGame){let i=t.configs.filter(c=>c.game===t.pickerGame).sort((c,p)=>(c.chapter||0)-(p.chapter||0));e=r(t.pickerGame);let a=i.map(c=>`
            <div class="picker-row" data-path="${r(c.path)}" data-embedded="true">
                <div class="picker-row-main">
                    ${c.chapter?`<div class="picker-game mono">${r($("chapterLabel",c.chapter))}</div>`:""}
                    <div class="picker-title">${r(c.title)}</div>
                    ${c.author?`<div class="picker-meta mono">${r($("by",c.author))}</div>`:""}
                </div>
                <div class="picker-clear" data-clear-chapter="${r(c.path)}" title="${o("clearProgress")}">⟲</div>
            </div>`).join("");n=`<div class="picker-back mono" id="pickerBack">${o("backToGames")}</div>`+(a||`<div class="picker-empty">${o("noChapters")}</div>`)}else{e=o("chooseGame");let i=[...new Set(t.configs.map(a=>a.game))].sort();n=i.length?i.map(a=>{let c=t.configs.filter(p=>p.game===a).length;return`<div class="picker-row" data-game="${r(a)}">
                    <div class="picker-row-main">
                        <div class="picker-title">${r(a)}</div>
                        <div class="picker-meta mono">${r($("chaptersCount",c))}</div>
                    </div>
                    <div class="picker-clear" data-clear-game="${r(a)}" title="${o("clearProgress")}">⟲</div>
                </div>`}).join(""):`<div class="picker-empty">${o("noConfigs")}</div>`}let s=t.pickerLoading?`<div class="picker-loading mono">${r($("downloading",t.pickerLoading))}</div>`:"";h.innerHTML=`
        <div class="header" id="header">
            <div class="header-left">
                <div>
                    <div class="kicker mono">GoThrough</div>
                </div>
            </div>
            <div class="header-right">
                <div class="lock-btn mono" id="lockBtn">${t.locked?o("adjust"):o("lock")}</div>
                <div class="icon-btn" id="settingsBtn" title="${o("settingsTitle")}">${F}</div>
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
            ${t.pickerError?`<div class="picker-error">${r(t.pickerError)}</div>`:""}
            ${s}
        </div>`}function we(e,n){let s=n+1,i=s<t.activePos,a=s===t.activePos,c=e.isChoice?`⑂ ${r(e.title)}`:r(e.title),p=e.optional?`<span class="row-opt">${o("optional")}</span>`:"";return`
        <div class="row${i?" done":""}${a?" current":""}${e.isChoice?" is-choice":""}" data-index="${n}">
            <div class="mark">${i?'<span class="check">✓</span>':""}</div>
            <div class="row-main">
                <div class="row-label">${c}${p}</div>
            </div>
        </div>`}function ye(){let e=[],n=null;return t.steps.forEach((s,i)=>{let a=s.section||"";(!n||n.section!==a)&&(n={section:a,items:[]},e.push(n)),n.items.push({s,i})}),e.map(s=>{let i=s.items.map(({s:k,i:u})=>we(k,u)).join("");if(!s.section)return i;let a=s.items.every(({i:k})=>k+1<t.activePos),c=t.sectionOverride[s.section],p=c===void 0?!a:c,m=a?`<span class="section-done mono">${o("completed")}</span>`:"";return`<div class="section-group${p?"":" sec-collapsed"}">
            <div class="section-head mono" data-section="${r(s.section)}">
                <span class="section-chevron">${j}</span>
                <span>${r(s.section)}</span>
                ${m}
            </div>
            <div class="section-rows">${i}</div>
        </div>`}).join("")}function $e(){let{total:e,doneCount:n,pct:s}=he(),i=t.current||t.steps[t.activePos-1]||null,a;if(!i)a=`<div class="done-card">${o("noSteps")}</div>`;else if(i.isChoice){let p=(i.options||[]).map(m=>{let k=i.selected&&m.value===i.selected;return`
            <button class="choice-opt${k?" chosen":""}" data-key="${r(i.choiceKey)}" data-value="${r(m.value)}">
                <span class="choice-opt-label">${r(m.label)}${k?` <span class="choice-opt-tag">${o("chosen")}</span>`:""}</span>
                ${m.description?`<span class="choice-opt-desc">${E(m.description)}</span>`:""}
            </button>`}).join("");a=`<div class="now-card choice-card">
               <div class="now-label mono">${o("decision")}</div>
               <div class="now-step">${r(i.title)}</div>
               <div class="now-meta mono">${r($("stepCount",t.activePos,e))}${i.section?" · "+r(i.section):""}</div>
               <div class="choice-options">${p}</div>
           </div>`}else{let p=i.optional?`<span class="opt-badge">${o("optional")}</span>`:"",m=i.isLast&&t.nextFile?`<button class="next-file-btn" id="nextFileBtn">${r($("nextFile",t.nextFile))}</button>`:"";a=`<div class="now-card">
               <div class="now-label mono">${o("now")}</div>
               <div class="now-step">${r(i.title)} ${p}</div>
               <div class="now-meta mono">${r($("stepCount",t.activePos,e))}${i.section?" · "+r(i.section):""}</div>
               ${i.description?`<div class="now-desc">${ae(i.description)}</div>`:""}
               ${le(i.tasks)}
               ${I(i.warnings,o("warning"),"warnings")}
               ${I(i.infos,o("info"),"infos")}
               ${ce(i.quests)}
               ${I(i.hints,o("hints"),"hints")}
               ${m}
           </div>`}let c=ye();h.innerHTML=`
        <div class="header" id="header">
            <div class="header-left">
                <div>
                    <div class="kicker mono">${r(t.meta.game)}</div>
                    <div class="quest-title" id="questTitle">${r(t.meta.title)}</div>
                </div>
            </div>
            <div class="header-right">
                <div class="lock-btn mono" id="lockBtn">${t.locked?o("adjust"):o("lock")}</div>
                <div class="icon-btn" id="homeBtn" title="${o("switchWalkthrough")}">${q}</div>
                <div class="icon-btn" id="settingsBtn" title="${o("settingsTitle")}">${F}</div>
                <div class="close-btn" id="closeBtn" title="${o("closeOverlay")}">✕</div>
            </div>
        </div>
        ${a}
        <div class="steps-section">
            <div class="steps-bar" id="stepsBar">
                <span class="steps-bar-left">
                    <span class="steps-chevron">${j}</span>
                    <span class="progress-label mono">${o("steps")}</span>
                </span>
                <div class="progress-track"><div class="progress-fill" style="width:${s}%"></div></div>
                <span class="progress-count mono">${n} / ${e}</span>
            </div>
            <div class="collapsible">
                <div class="checklist">${c}</div>
            </div>
        </div>`}function Ee(){return`
        <div class="settings-view">
            <div class="settings-kicker mono">${o("shortcuts")}</div>
            ${de.map(e=>{let n=t.settings&&t.settings.hotkeys[e.key],s=t.capturing===e.key;return`<div class="setting-row">
                    <span class="setting-name">${r(o(e.labelKey))}</span>
                    <span class="combo mono${s?" capturing":""}" data-action="${e.key}">${s?o("pressKey"):r(ge(n))}</span>
                </div>`}).join("")}
            <div class="setting-row">
                <span class="setting-name">${o("opacity")}</span>
                <input type="range" class="opacity-slider" id="opacitySlider"
                    min="10" max="100" step="5"
                    value="${Math.round((t.settings?t.settings.opacity??1:1)*100)}">
            </div>
            <div class="setting-row">
                <span class="setting-name">${o("design")}</span>
                <span class="theme-switch">${z.map(e=>`<button class="theme-opt${(t.settings&&t.settings.theme||"dark")===e.key?" active":""}" data-theme="${e.key}">${r(o(e.labelKey))}</button>`).join("")}</span>
            </div>
            <div class="setting-row">
                <span class="setting-name">${o("language")}</span>
                <span class="theme-switch">${re.map(e=>`<button class="lang-opt${T()===e.key?" active":""}" data-lang="${e.key}">${r(e.label)}</button>`).join("")}</span>
            </div>
            <div class="settings-error">${r(t.settingsError)}</div>
            <div class="settings-hint">${o("settingsHint")}</div>
            <button class="settings-reset" id="resetSettings">${o("resetSettings")}</button>
        </div>`}function be(){let e=t.settingsFrom==="steps";h.innerHTML=`
        <div class="header" id="header">
            <div class="header-left">
                <div>
                    <div class="kicker mono">${e?r(t.meta.game):"GoThrough"}</div>
                    ${e?`<div class="quest-title">${r(t.meta.title)}</div>`:""}
                </div>
            </div>
            <div class="header-right">
                <div class="lock-btn mono" id="lockBtn">${t.locked?o("adjust"):o("lock")}</div>
                ${e?`<div class="icon-btn" id="homeBtn" title="${o("switchWalkthrough")}">${q}</div>`:""}
                <div class="icon-btn active" id="settingsBtn" title="${o("settingsTitle")}">${F}</div>
                <div class="close-btn" id="closeBtn" title="${o(e?"closeOverlay":"close")}">✕</div>
            </div>
        </div>
        ${Ee()}`}function U(){w.style.left="0px",w.style.top="0px",w.style.width=t.width+"px";let e=t.view==="steps"&&!t.collapsed;h.style.height=e?t.panelHeight+"px":"auto"}function Ce(){if(t.view!=="steps"||t.collapsed||t.lastScrolledPos===t.activePos)return;let e=h.querySelector(".checklist"),n=e&&e.querySelector(".row.current");if(e&&n){let i=n.getBoundingClientRect(),a=e.getBoundingClientRect();e.scrollTop=Math.max(0,e.scrollTop+(i.top-a.top)-40),t.lastScrolledPos=t.activePos}}function Se(e){d.Goto(e).then(n=>{t.activePos=n.current,t.current=n,l()})}function Le(e){t.pickerLoading=e,t.pickerError="",l(),d.DownloadGame(e).then(()=>{t.pickerLoading="",t.pickerGame=e,l()}).catch(n=>{t.pickerLoading="",t.pickerGame=e,t.pickerError=String(n&&n.message?n.message:n),l()})}function Pe(e){Q({message:$("confirmClearGame",e),confirmLabel:o("confirmReset"),onConfirm:()=>{t.pickerError="",d.ClearGameProgress(e).then(()=>l()).catch(n=>{t.pickerError=String(n&&n.message?n.message:n),l()})}})}function xe(e){t.pickerError="",d.ClearChapterProgress(e).then(()=>l()).catch(n=>{t.pickerError=String(n&&n.message?n.message:n),l()})}function Be(){Q({message:o("confirmClearCache"),confirmLabel:o("confirmDelete"),onConfirm:()=>{t.pickerError="",d.ClearCache().then(()=>{t.pickerGame=null,l()}).catch(e=>{t.pickerError=String(e&&e.message?e.message:e),l()})}})}function Me(){d.ResetSettings().then(e=>{t.settings=e,A(e.opacity??1),P(e.theme),M(),t.settingsError="",l()}).catch(e=>{t.settingsError=String(e&&e.message?e.message:e),l()})}function D(e,n){t.pickerError="",d.LoadConfig(e,n).then(()=>Promise.all([d.Meta(),d.Steps(),d.CurrentStep(),d.NextFile()])).then(([s,i,a,c])=>{t.meta=s,t.steps=i,t.current=a,t.activePos=a.current,t.nextFile=c||"",t.sectionOverride={},t.lastScrolledPos=null,t.view="steps",l()}).catch(s=>{t.pickerError=String(s&&s.message?s.message:s),l()})}function Y(){return Promise.all([d.Steps(),d.CurrentStep(),d.NextFile()]).then(([e,n,s])=>{t.steps=e,t.current=n,t.activePos=n.current,t.nextFile=s||"",t.lastScrolledPos=null,l()})}function Te(e,n){d.Choose(e,n).then(()=>Y())}function V(){return t.sectionOverride={},d.Meta().then(e=>(t.meta=e,Y()))}function Ae(){d.LoadNext().then(()=>V()).catch(e=>{console.error("[GoThrough] next hand-off failed:",e)})}function He(){d.UnloadConfig().then(()=>(t.view="picker",t.pickerGame=null,t.configs=[],d.ListConfigs())).then(e=>{t.configs=e,l()})}function Ie(){document.getElementById("header").addEventListener("mousedown",ze);let e=document.getElementById("lockBtn");e.addEventListener("mousedown",g),e.addEventListener("click",Ne);let n=document.getElementById("closeBtn");n.addEventListener("mousedown",g),n.addEventListener("click",()=>{window.runtime?.Quit?.()});let s=t.view==="picker"&&document.getElementById("settingsBtn");if(s&&(s.addEventListener("mousedown",g),s.addEventListener("click",W)),t.view==="picker"){h.querySelectorAll(".picker-row").forEach(v=>{v.addEventListener("click",()=>{v.dataset.game?Le(v.dataset.game):v.dataset.path&&D(v.dataset.path,v.dataset.embedded==="true")})}),h.querySelectorAll(".picker-clear").forEach(v=>{v.addEventListener("mousedown",g),v.addEventListener("click",J=>{J.stopPropagation(),v.dataset.clearGame?Pe(v.dataset.clearGame):v.dataset.clearChapter&&xe(v.dataset.clearChapter)})});let u=document.getElementById("pickerBack");u&&(u.addEventListener("mousedown",g),u.addEventListener("click",()=>{t.pickerGame=null,l()}));let C=document.getElementById("pickerBrowse");C&&(C.addEventListener("mousedown",g),C.addEventListener("click",()=>{d.OpenBrowse().then(v=>{v&&D(v,!1)})}));let H=document.getElementById("pickerClearCache");H&&(H.addEventListener("mousedown",g),H.addEventListener("click",Be));return}let i=document.getElementById("stepsBar");i&&i.addEventListener("click",()=>{t.collapsed=!t.collapsed,l()}),h.querySelectorAll(".section-head").forEach(u=>{u.addEventListener("click",()=>Oe(u.dataset.section))});let a=document.getElementById("homeBtn");a&&(a.addEventListener("mousedown",g),a.addEventListener("click",He));let c=document.getElementById("settingsBtn");c&&(c.addEventListener("mousedown",g),c.addEventListener("click",W)),h.querySelectorAll(".row").forEach(u=>u.addEventListener("click",()=>Se(Number(u.dataset.index)))),h.querySelectorAll(".choice-opt").forEach(u=>{u.addEventListener("mousedown",g),u.addEventListener("click",()=>Te(u.dataset.key,u.dataset.value))});let p=document.getElementById("nextFileBtn");p&&(p.addEventListener("mousedown",g),p.addEventListener("click",Ae)),h.querySelectorAll(".combo").forEach(u=>{u.addEventListener("mousedown",g),u.addEventListener("click",()=>Ke(u.dataset.action))});let m=document.getElementById("opacitySlider");m&&(m.addEventListener("mousedown",g),m.addEventListener("input",()=>{let u=Number(m.value)/100;A(u)}),m.addEventListener("change",()=>{let u=Number(m.value)/100;d.SaveOpacity(u).then(C=>{t.settings=C})})),h.querySelectorAll(".theme-opt").forEach(u=>{u.addEventListener("mousedown",g),u.addEventListener("click",()=>Fe(u.dataset.theme))}),h.querySelectorAll(".lang-opt").forEach(u=>{u.addEventListener("mousedown",g),u.addEventListener("click",()=>Ge(u.dataset.lang))});let k=document.getElementById("resetSettings");k&&(k.addEventListener("mousedown",g),k.addEventListener("click",Me))}function Oe(e){if(!e)return;let s=t.steps.filter(c=>(c.section||"")===e).length>0&&t.steps.every((c,p)=>(c.section||"")!==e||p+1<t.activePos),i=t.sectionOverride[e],a=i===void 0?!s:i;t.sectionOverride[e]=!a,l()}function Ge(e){e!==T()&&d.SaveLanguage(e).then(n=>{t.settings=n,l()}).catch(n=>{t.settingsError=String(n&&n.message?n.message:n),l()})}function Fe(e){let n=t.settings?t.settings.theme:"dark";P(e),d.SaveTheme(e).then(s=>{t.settings=s,M(),l()}).catch(s=>{P(n),t.settingsError=String(s&&s.message?s.message:s),l()})}function W(){t.view==="settings"?t.view=t.settingsFrom||"steps":(t.settingsFrom=t.view,t.view="settings"),t.capturing=null,t.settingsError="",l()}function Ke(e){t.capturing=e,t.settingsError="",l()}function X(e){let n=t.capturing,s={...t.settings.hotkeys,[n]:e};window.go.overlay.App.SaveHotkeys(s).then(i=>{t.settings=i,t.capturing=null,t.settingsError="",l()}).catch(i=>{t.capturing=null,t.settingsError=String(i&&i.message?i.message:i),l()})}function De(e){if(t.confirm){if(e.key==="Escape")e.preventDefault(),e.stopPropagation(),L();else if(e.key==="Enter"){e.preventDefault(),e.stopPropagation();let s=t.confirm.onConfirm;L(),s&&s()}return}if(!t.capturing)return;if(e.preventDefault(),e.stopPropagation(),e.key==="Escape"){t.capturing=null,l();return}let n=me(e);n&&X(n)}function We(e){if(!t.capturing)return;e.preventDefault(),e.stopPropagation();let n=ve(e);n&&X(n)}window.addEventListener("keydown",De,!0);window.addEventListener("mousedown",We,!0);window.addEventListener("contextmenu",e=>{t.capturing&&e.preventDefault()},!0);function Ne(){t.locked=!t.locked,M(),l()}function g(e){e.stopPropagation()}var y=null;function Re(e){t.locked||(e.preventDefault(),e.stopPropagation(),y={sx:e.screenX,sy:e.screenY,ow:t.width,oh:t.panelHeight},w.classList.add("dragging"))}function je(e){y&&(t.width=Math.max(te,Math.min(y.ow-(e.screenX-y.sx),ne)),t.panelHeight=Math.max(se,Math.min(y.oh+(e.screenY-y.sy),ie)),U())}function qe(){y&&(y=null,w.classList.remove("dragging"),M())}oe.addEventListener("mousedown",Re);window.addEventListener("mousemove",je);window.addEventListener("mouseup",qe);var f=null,b=0;async function ze(e){if(t.locked||e.button!==0||e.target.closest(".lock-btn, .icon-btn, .close-btn, .picker-browse, .picker-clearcache, .picker-row, .picker-back"))return;let n=window.runtime;if(!n||!n.WindowGetPosition)return;e.preventDefault();let s=await n.WindowGetPosition();f={sx:e.screenX,sy:e.screenY,ox:s.x,oy:s.y,tx:s.x,ty:s.y}}function _e(e){f&&(f.tx=Math.round(f.ox+(e.screenX-f.sx)),f.ty=Math.round(f.oy+(e.screenY-f.sy)),b||(b=requestAnimationFrame(Qe)))}async function Qe(){if(b=0,!f)return;let e=await d.MoveWindow?.(f.tx,f.ty);e&&f&&(f.lastX=e.x,f.lastY=e.y)}async function Ue(){b&&(cancelAnimationFrame(b),b=0);let e=f;if(f=null,!e||!(e.tx!==e.ox||e.ty!==e.oy))return;let s=await d.MoveWindow?.(e.tx,e.ty),i=s?s.x:e.lastX,a=s?s.y:e.lastY;i!=null&&d.SaveWindowPos?.(i,a)}window.addEventListener("mousemove",_e);window.addEventListener("mouseup",Ue);function Ye(){let e=t.view==="steps"&&!t.collapsed,n,s;if(e)n=t.width,s=t.panelHeight;else{let i=h.getBoundingClientRect();n=Math.ceil(i.width),s=Math.ceil(i.height)}n>0&&s>0&&window.go?.overlay?.App?.FitWindow?.(n,s)}new ResizeObserver(Ye).observe(h);async function Ve(){P(B.theme);try{let e=await d.Settings();if(t.settings=e,A(e.opacity??1),P(e.theme),await d.IsPicker())t.view="picker",t.configs=await d.ListConfigs();else{let[s,i,a,c]=await Promise.all([d.Meta(),d.Steps(),d.CurrentStep(),d.NextFile()]);t.meta=s,t.steps=i,t.current=a,t.activePos=a.current,t.nextFile=c||"",t.view="steps"}}catch(e){console.error("[GoThrough] init failed:",e)}l(),window.runtime&&window.runtime.EventsOn&&(window.runtime.EventsOn("step:changed",e=>{t.activePos=e.current,t.current=e,l()}),window.runtime.EventsOn("config:changed",()=>{V()}),window.runtime.EventsOn("configs:remote",e=>{Array.isArray(e)&&e.length&&(t.configs=e,t.view==="picker"&&l())}))}Ve();})();
