/*
  VectorChat Plex Widget (standalone)
  - Monospaced, dashboard-inspired style (IBM Plex Mono vibe)
  - No external deps; Shadow DOM styling; responsive.
  - Use via: <script src="/3/vectorchat-plex-widget.js"></script>
*/
(function(){
  if (window.__vectorchat_plex_widget_loaded__) return; 
  window.__vectorchat_plex_widget_loaded__ = true;

  let isRecording=false; let message=''; let isTyping=false; let isSending=false; let sessionId=undefined;
  let initialAssistantMessage="Hello! I'm your VectorChat assistant. Ask me anything.";
  /** @type {{id:string,text:string,isUser:boolean,timestamp:Date}[]} */ let messages=[];

  const scriptEl = document.currentScript;
  const scriptMeta = (() => {
    const attrSrc = scriptEl?.getAttribute('src') || '';
    let resolvedSrc = '';
    try { resolvedSrc = scriptEl?.src || attrSrc; } catch { resolvedSrc = attrSrc; }

    let parsedUrl = null;
    if (resolvedSrc) {
      try {
        parsedUrl = new URL(resolvedSrc);
      } catch (_) {
        try {
          parsedUrl = new URL(
            resolvedSrc,
            typeof window !== 'undefined' ? window.location.origin : undefined,
          );
        } catch {
          parsedUrl = null;
        }
      }
    }

    const origin = parsedUrl && parsedUrl.origin !== 'null' ? parsedUrl.origin : '';
    const pathname = parsedUrl?.pathname || '';
    const apiBaseAttr = (scriptEl?.getAttribute('data-api-base') || '').trim();
    const chatIdAttr = scriptEl?.getAttribute('data-chat-id') || '';
    return { origin, pathname, apiBaseAttr, chatIdAttr };
  })();

  const apiBase = (() => {
    const candidate = scriptMeta.apiBaseAttr || scriptMeta.origin;
    return candidate ? candidate.replace(/\/$/, '') : '';
  })();

  const buildApiUrl = (path) => apiBase ? `${apiBase}${path}` : path;

  const chatId = (() => {
    if (scriptMeta.chatIdAttr) {
      try { return decodeURIComponent(scriptMeta.chatIdAttr); } catch { return scriptMeta.chatIdAttr; }
    }
    const segments = scriptMeta.pathname.split('/').filter(Boolean);
    if (segments.length >= 2) {
      const raw = segments[segments.length - 2] || '';
      try { return decodeURIComponent(raw); } catch { return raw; }
    }
    return '';
  })();

  const chatMessageEndpoint = chatId ? buildApiUrl(`/api/chatbot/${encodeURIComponent(chatId)}/message`) : '';
  const chatbotInfoEndpoint = chatId ? buildApiUrl(`/api/chatbot/${encodeURIComponent(chatId)}`) : '';

  const host = document.createElement('div');
  host.style.position='fixed'; host.style.left='50%'; host.style.bottom='24px'; host.style.transform='translateX(-50%)'; host.style.zIndex='2147483647';
  const shadow = host.attachShadow({mode:'open'});

  const style = document.createElement('style');
  style.textContent = `
    :host{all:initial}
    *,*::before,*::after{box-sizing:border-box}
    .root{
      --bg:#0f1216;            /* canvas */
      --panel:#11151a;         /* panel base */
      --panel-2:#0d1014;       /* deeper */
      --border:#222830;        /* outlines */
      --muted:#8c96a5;         /* secondary text */
      --fg:#e6eaee;            /* primary text */
      --cyan:#2eccb6;          /* accent up */
      --red:#ff6b6b;           /* accent down */
      --amber:#f0b44f;         /* subtle accent */
      --shadow: 0 18px 48px rgba(0,0,0,.55);
      --ease: cubic-bezier(.22,.61,.36,1);
      --dur: 260ms; --fast: 170ms;
      font-family: "IBM Plex Mono", ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
      letter-spacing:.01em;
      -webkit-font-smoothing: antialiased; -moz-osx-font-smoothing: grayscale;
    }
    .hidden{display:none}

    /* Collapsed input (console-like) */
    .pill{ min-width:420px; max-width:560px; padding:12px 14px; border-radius:0;
      background: linear-gradient(180deg, var(--panel), var(--panel-2));
      border:1px solid var(--border); box-shadow: var(--shadow); color:var(--fg); cursor:pointer;
      transition: transform var(--fast) var(--ease), opacity var(--fast) var(--ease), box-shadow var(--fast) var(--ease), background var(--fast) var(--ease);
      opacity:0; transform: translateY(14px) scale(.995);
    }
    .pill.in{ opacity:1; transform: translateY(0) scale(1); }
    .pill:hover{ box-shadow: 0 22px 60px rgba(0,0,0,.6); }
    .muted{ color:var(--muted); font-size:.9rem; }
    .row{ display:flex; align-items:center; }
    .row>*+*{ margin-left:10px }
    .icon-btn{ border:none; background:transparent; padding:4px; border-radius:0; }
    .icon{ width:16px; height:16px; color:#cfd6df }

    /* Expanded panel */
    .card{ width:420px; height:62px; opacity:0; transform: translateY(14px) scale(.995); border-radius:0; overflow:hidden; position:relative;
      background: linear-gradient(180deg, var(--panel), var(--panel-2)); border:1px solid var(--border); box-shadow: var(--shadow); color:var(--fg);
      transition: height var(--dur) var(--ease), transform var(--dur) var(--ease), opacity var(--dur) var(--ease);
    }
    .card.open{ height:500px; opacity:1; transform: translateY(0) scale(1); }

    /* Title bar with subtle grid */
    .card::before{ content:""; position:absolute; inset:0; background-image:
      linear-gradient(var(--border) 1px, transparent 1px),
      linear-gradient(90deg, var(--border) 1px, transparent 1px);
      background-size: 28px 28px; opacity:.08; pointer-events:none; }

    .close{ position:absolute; top:10px; right:10px; border-radius:0; z-index:3; }
    .close:hover{ background: rgba(255,255,255,.05); }

    .messages{ position:relative; z-index:2; height:416px; padding:16px; padding-top:52px; overflow-y:auto; display:flex; flex-direction:column; gap:10px; scroll-behavior:smooth; }
    .msg-row{ display:flex }
    .left{ justify-content:flex-start } .right{ justify-content:flex-end }
    .bubble{ max-width:320px; padding:10px 12px; border-radius:0; font-size:.92rem; line-height:1.5; border:1px solid var(--border); background: #0f1318; }
    .ai{ background: #0e1318; }
    .me{ background: #131922; border-color:#2a3340; }

    @keyframes in { from{opacity:0; transform: translateY(8px)} to{opacity:1; transform: translateY(0)} }
    .enter{ animation: in var(--fast) var(--ease) both }

    /* Footer */
    .footer{ position:relative; z-index:2; padding:12px 14px; border-top:1px solid var(--border); background: #0e1216; }
    .input{ width:100%; background:transparent; border:none; outline:none; color:var(--fg); font-size:.95rem; }
    .input::placeholder{ color:var(--muted); font-size:.88rem }

    .typing{ display:inline-grid; grid-auto-flow:column; gap:6px; }
    .sq{ width:6px; height:6px; background:var(--muted); animation: blink 1.2s infinite var(--ease); }
    .sq:nth-child(2){ animation-delay:.2s } .sq:nth-child(3){ animation-delay:.4s }
    @keyframes blink { 0%,100%{opacity:.4} 40%{opacity:1} }

    /* Badges */
    .badge{ font-size:.7rem; color:#0b0; background:rgba(46,204,182,.1); border:1px solid rgba(46,204,182,.35); padding:2px 6px; border-radius:6px; }

    /* Responsive */
    @media (max-width:480px){
      .pill{ min-width:0; width: calc(100vw - 32px); }
      .card{ width: calc(100vw - 32px); }
      .card.open{ height: min(70vh, 520px); }
      .messages{ height: calc(100% - 80px); padding:12px; padding-top:48px; }
      .footer{ padding:10px 12px }
    }
  `;

  const icons={
    mic:`<svg viewBox="0 0 24 24" class="icon" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 1a3 3 0 0 0-3 3v6a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3Z"/><path d="M19 10v2a7 7 0 0 1-14 0v-2"/><line x1="12" y1="19" x2="12" y2="23"/><line x1="8" y1="23" x2="16" y2="23"/></svg>`,
    barchart:`<svg viewBox="0 0 24 24" class="icon" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="3" y1="3" x2="3" y2="21"/><line x1="9" y1="9" x2="9" y2="21"/><line x1="15" y1="5" x2="15" y2="21"/><line x1="21" y1="13" x2="21" y2="21"/></svg>`,
    send:`<svg viewBox="0 0 24 24" class="icon" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="22" y1="2" x2="11" y2="13"/><polygon points="22 2 15 22 11 13 2 9 22 2"/></svg>`,
    x:`<svg viewBox="0 0 24 24" class="icon" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>`,
    check:`<svg viewBox="0 0 24 24" class="icon" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 6 9 17l-5-5"/></svg>`,
  };

  const root = document.createElement('div'); root.className='root';

  const pill = document.createElement('div');
  pill.className='pill';
  pill.innerHTML = `
    <div class="row" style="justify-content:space-between;">
      <span class="muted">Ask anything...</span>
      <div class="row">
        <button class="icon-btn" data-action="mic">${icons.mic}</button>
        ${icons.barchart}
      </div>
    </div>`;

  const card = document.createElement('div');
  card.className='card';
  card.innerHTML = `
    <button class="icon-btn close" data-action="close">${icons.x}</button>
    <div class="messages" data-ref="messages"></div>
    <div class="footer" data-ref="footer"></div>`;

  const messagesEl = () => card.querySelector('[data-ref="messages"]');
  const footerEl = () => card.querySelector('[data-ref="footer"]');

  function renderFooter(){
    const f = footerEl(); if(!f) return;
    if(isRecording){
      f.innerHTML = `
        <div class="row" style="justify-content:space-between;">
          <button class="icon-btn" data-action="cancel">${icons.x}</button>
          <div style="flex:1; display:flex; align-items:center; justify-content:center; gap:2px; height:24px;">
            ${Array.from({length:32}).map(()=>'<div class="bar"></div>').join('')}
          </div>
          <button class="icon-btn" data-action="accept">${icons.check}</button>
        </div>`;
      f.querySelectorAll('.bar').forEach((el)=>{ el.style.width='2px'; el.style.background='#cfd6df'; el.style.animation=`w ${0.8+Math.random()*0.5}s var(--ease) ${Math.random()*0.4}s infinite`; el.style.borderRadius='1px'; });
      const s = document.createElement('style'); s.textContent='@keyframes w{0%{height:4px}50%{height:20px}100%{height:4px}}'; shadow.appendChild(s);
    } else {
      const has=(message||'').trim().length>0;
      const sendDisabledAttr = isSending ? 'disabled aria-disabled="true"' : '';
      f.innerHTML = `
        <div class="row" style="justify-content:space-between;">
          <input class="input" data-ref="input" placeholder="Type your message..." ${isSending ? 'disabled' : ''} />
          <div class="row">
            <button class="icon-btn" data-action="mic">${icons.mic}</button>
            ${has? `<button class="icon-btn" data-action="send" ${sendDisabledAttr}>${icons.send}</button>` : icons.barchart}
          </div>
        </div>`;
      const input = f.querySelector('[data-ref="input"]');
      if(input){ input.value=message; input.setAttribute('aria-disabled', isSending ? 'true':'false');
        setTimeout(()=>{ if(!isSending){ input.focus(); input.setSelectionRange(input.value.length,input.value.length); } },0);
        input.addEventListener('input', e=>{ message=e.target.value; renderFooter(); });
        input.addEventListener('keydown', e=>{ if(e.key==='Enter'&&!e.shiftKey){ e.preventDefault(); send(); } }); }
    }
    f.querySelectorAll('[data-action]').forEach(btn=>{
      const a=btn.getAttribute('data-action');
      btn.addEventListener('click', ev=>{ ev.stopPropagation();
        if(btn.hasAttribute('disabled')) return;
        if(a==='mic'){ isRecording=true; renderFooter(); }
        else if(a==='send'){ send(); }
        else if(a==='cancel'){ isRecording=false; renderFooter(); }
        else if(a==='accept'){ message='This is a simulated voice message converted to text.'; isRecording=false; renderFooter(); }
        else if(a==='close'){ close(); }
      });
    });
  }

  function renderMessages(){
    const m = messagesEl(); if(!m) return; m.innerHTML='';
    if(messages.length===0){
      const empty=document.createElement('div'); empty.style.textAlign='center'; empty.style.color='var(--muted)'; empty.style.marginTop='24px'; empty.style.fontSize='.88rem'; empty.textContent='Start a conversation'; m.appendChild(empty);
    } else {
      for(const msg of messages){
        const row=document.createElement('div'); row.className='msg-row '+(msg.isUser?'right':'left')+' enter';
        const b=document.createElement('div'); b.className='bubble '+(msg.isUser?'me':'ai'); b.textContent=msg.text; row.appendChild(b); m.appendChild(row);
      }
    }
    if(isTyping){
      const row=document.createElement('div'); row.className='msg-row left enter';
      const b=document.createElement('div'); b.className='bubble ai'; b.innerHTML='<span class="typing"><span class="sq"></span><span class="sq"></span><span class="sq"></span></span>';
      row.appendChild(b); m.appendChild(row);
    }
    try{ m.scrollTo({top:m.scrollHeight, behavior:'smooth'}); }catch{ m.scrollTop=m.scrollHeight; }
  }

  function pushAssistantMessage(text){
    messages.push({id:String(Date.now()+Math.random()), text, isUser:false, timestamp:new Date()});
  }

  async function requestChatbotResponse(userText){
    const payload={ query:userText, session_id:sessionId };
    const endpoints=[];
    if(chatMessageEndpoint){ endpoints.push(chatMessageEndpoint); }
    if(chatId){
      const directEndpoint = buildApiUrl(`/chat/${encodeURIComponent(chatId)}/message`);
      if(directEndpoint && !endpoints.includes(directEndpoint)){ endpoints.push(directEndpoint); }
    }
    if(!endpoints.length) throw new Error('Missing chat endpoint');

    let lastError=null;
    for(const url of endpoints){
      try{
        const response = await fetch(url, {
          method:'POST',
          headers:{ 'Content-Type':'application/json' },
          credentials:'include',
          body: JSON.stringify(payload)
        });
        if(!response.ok){
          const errText = await response.text().catch(()=> '');
          throw new Error(`Chat request failed (${response.status}): ${errText || response.statusText}`);
        }
        const data = await response.json().catch(()=> ({}));
        if(data && typeof data.session_id === 'string'){ sessionId = data.session_id; }
        return data?.response || data?.message || "I'm sorry, but I couldn't generate a response right now.";
      }catch(err){
        lastError = err;
      }
    }
    throw lastError || new Error('Unable to reach chatbot endpoint.');
  }

  async function initializeWidget(){
    if(!chatId){
      initialAssistantMessage = "VectorChat widget misconfigured: missing chatbot ID.";
      return;
    }
    const infoEndpoints=[];
    if(chatbotInfoEndpoint){ infoEndpoints.push(chatbotInfoEndpoint); }
    const directInfo = buildApiUrl(`/chat/chatbot/${encodeURIComponent(chatId)}`);
    if(directInfo && !infoEndpoints.includes(directInfo)){ infoEndpoints.push(directInfo); }
    for(const url of infoEndpoints){
      try{
        const response = await fetch(url, { credentials:'include' });
        if(!response.ok) continue;
        const data = await response.json().catch(()=> ({}));
        const greeting = data?.welcome_message || data?.greeting || data?.intro_message || data?.description;
        if(greeting && typeof greeting === 'string'){
          initialAssistantMessage = greeting;
          if(messages.length===1 && !messages[0].isUser){
            messages[0].text = greeting;
            renderMessages();
          }
          break;
        }
      }catch(err){
        console.warn('[VectorChat] Failed to preload chatbot info:', err);
      }
    }
  }

  function send(){
    const text=(message||'').trim();
    if(!text || isSending) return;
    if(!chatId){
      pushAssistantMessage("This widget is not configured with a chatbot ID. Please reload and try again.");
      message=''; renderFooter(); renderMessages();
      return;
    }
    messages.push({id:String(Date.now()), text, isUser:true, timestamp:new Date()});
    message='';
    isSending=true; isTyping=true;
    renderFooter(); renderMessages();
    requestChatbotResponse(text).then(resText=>{
      pushAssistantMessage(String(resText||'').trim()||'...');
    }).catch(err=>{
      console.error('[VectorChat] Widget error:', err);
      pushAssistantMessage("I'm having trouble responding right now. Please try again later.");
    }).finally(()=>{
      isSending=false; isTyping=false;
      renderFooter(); renderMessages();
    });
  }

  function open(){ card.classList.add('open'); pill.classList.add('hidden'); if(messages.length===0 && initialAssistantMessage){ messages.push({id:String(Date.now()), text:initialAssistantMessage, isUser:false, timestamp:new Date()}); } renderFooter(); renderMessages(); }
  function close(){ card.classList.remove('open'); setTimeout(()=>{ pill.classList.remove('hidden'); pill.classList.add('in'); },160); }

  pill.addEventListener('click', open);
  pill.querySelector('[data-action="mic"]').addEventListener('click', e=>{ e.stopPropagation(); open(); isRecording=true; renderFooter(); });
  card.querySelector('[data-action="close"]').addEventListener('click', e=>{ e.stopPropagation(); close(); });

  const rootEl = document.createElement('div'); rootEl.className='root'; rootEl.appendChild(pill); rootEl.appendChild(card);
  shadow.appendChild(style); shadow.appendChild(rootEl); document.body.appendChild(host);
  initializeWidget();
  requestAnimationFrame(()=>{ pill.classList.add('in'); });

  try { const mq=matchMedia('(max-width:480px)'); const apply=()=>{ host.style.bottom = mq.matches ? '16px' : '24px'; }; mq.addEventListener?mq.addEventListener('change',apply):mq.addListener(apply); apply(); } catch {}
})();
