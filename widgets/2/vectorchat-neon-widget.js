/*
  VectorChat Neon Widget (standalone)
  - Single-file vanilla JS chat widget with vibrant neon/crypto style.
  - Drop this via <script src="/2/vectorchat-neon-widget.js"></script> to auto-append.
*/
(function () {
  if (window.__vectorchat_neon_widget_loaded__) return;
  window.__vectorchat_neon_widget_loaded__ = true;

  // State
  let isRecording = false;
  let message = '';
  let isTyping = false;
  /** @type {{id:string,text:string,isUser:boolean,timestamp:Date}[]} */
  let messages = [];

  // Host
  const host = document.createElement('div');
  host.style.position = 'fixed';
  host.style.left = '50%';
  host.style.bottom = '24px';
  host.style.transform = 'translateX(-50%)';
  host.style.zIndex = '2147483647';
  const shadow = host.attachShadow({ mode: 'open' });

  const style = document.createElement('style');
  style.textContent = `
    :host { all: initial; }
    *,*::before,*::after{ box-sizing: border-box; }
    .vcw-root {
      --bg-0: #0b1026;              /* deep navy */
      --bg-1: #11163b;              /* card base */
      --fg: #f7f8ff;                /* near-white */
      --muted: #cbd5ff;             /* soft white-blue */
      --primary: #6d7cff;           /* neon indigo */
      --primary-2: #3b7cff;         /* bright blue */
      --accent: #9d7bff;            /* purple glow */
      --ring: rgba(113, 148, 255, .55);
      --glass: rgba(255,255,255,0.08);
      --glass-2: rgba(255,255,255,0.10);
      --shadow: 0 18px 60px rgba(15,20,60,.55);
      --ease: cubic-bezier(.22,.61,.36,1);
      --dur-fast: 180ms; --dur: 280ms;
      font-family: ui-sans-serif, system-ui, -apple-system, Segoe UI, Roboto, Inter, Helvetica, Arial;
      -webkit-font-smoothing: antialiased; -moz-osx-font-smoothing: grayscale;
    }
    .vcw-hidden{display:none}

    /* Collapsed pill (neon glass) */
    .pill{
      min-width:420px; max-width:560px; padding:12px 16px; border-radius:16px;
      background: linear-gradient(180deg, rgba(255,255,255,.10), rgba(255,255,255,.06));
      -webkit-backdrop-filter: blur(18px) saturate(140%);
      backdrop-filter: blur(18px) saturate(140%);
      border:1px solid rgba(255,255,255,.14);
      box-shadow: var(--shadow), 0 0 0 1px rgba(255,255,255,.04) inset;
      color: var(--fg); cursor:pointer; position:relative;
      transition: transform var(--dur-fast) var(--ease), box-shadow var(--dur-fast) var(--ease), background var(--dur-fast) var(--ease), opacity var(--dur-fast) var(--ease);
      opacity:0; transform: translateY(16px) scale(.995);
    }
    .pill::before{ content:""; position:absolute; inset:-1px; border-radius:inherit; pointer-events:none;
      background: radial-gradient(40% 60% at 8% 0%, rgba(109,124,255,.35), transparent 55%),
                  radial-gradient(60% 80% at 100% 100%, rgba(157,123,255,.25), transparent 55%);
      filter: blur(12px); opacity:.8;
    }
    .pill.in{ opacity:1; transform: translateY(0) scale(1); }
    .pill:hover{ background: linear-gradient(180deg, rgba(255,255,255,.12), rgba(255,255,255,.08)); box-shadow: 0 20px 70px rgba(20,40,120,.55), 0 0 0 1px rgba(255,255,255,.06) inset; }
    .muted{ color: var(--muted); font-size:.92rem; }
    .row{ display:flex; align-items:center; }
    .row > * + *{ margin-left:10px; }
    .icon-btn{ border:none; background:transparent; padding:6px; border-radius:10px; }
    .icon{ width:16px; height:16px; color:#eef1ff; opacity:.9; }

    /* Expanded panel: neon card with glow ring */
    .card{ width:420px; height:64px; opacity:0; transform: translateY(16px) scale(.995); overflow:hidden; border-radius:22px; position:relative;
      transition: height var(--dur) var(--ease), transform var(--dur) var(--ease), opacity var(--dur) var(--ease); 
      background: linear-gradient(180deg, rgba(255,255,255,.08), rgba(255,255,255,.06));
      -webkit-backdrop-filter: blur(22px) saturate(140%);
      backdrop-filter: blur(22px) saturate(140%);
      border:1px solid rgba(255,255,255,.14); box-shadow: var(--shadow);
      color: var(--fg);
    }
    .card.open{ height:520px; opacity:1; transform: translateY(0) scale(1); }
    .card::before{ content:""; position:absolute; inset:-2px; border-radius:inherit; pointer-events:none;
      background: conic-gradient(from 0deg, rgba(109,124,255,.35), rgba(59,124,255,.45), rgba(157,123,255,.35), rgba(109,124,255,.35));
      filter: blur(24px); opacity:.45; }
    .close{ position:absolute; top:12px; right:12px; z-index:3; }
    .close:hover{ background: rgba(255,255,255,.10); }

    .messages{ position:relative; z-index:2; height:436px; padding:18px; padding-top:56px; overflow-y:auto; display:flex; flex-direction:column; gap:12px; scroll-behavior:smooth; }
    .bubble{ max-width:320px; padding:12px 14px; border-radius:14px; font-size:.94rem; line-height:1.5; box-shadow: 0 6px 20px rgba(10,15,40,.25); }
    .ai{ background: rgba(255,255,255,.10); border:1px solid rgba(255,255,255,.12); color:#ecf0ff; }
    .me{ background: linear-gradient(180deg, rgba(109,124,255,.35), rgba(59,124,255,.35)); border:1px solid rgba(113,148,255,.55); color:white; }
    .msg-row{ display:flex; }
    .left{ justify-content:flex-start; } .right{ justify-content:flex-end; }
    @keyframes in { from{ opacity:0; transform: translateY(8px);} to{opacity:1; transform: translateY(0);} }
    .enter{ animation: in var(--dur-fast) var(--ease) both; }

    .footer{ position:relative; z-index:2; border-top: 1px solid rgba(255,255,255,.10); padding:14px 16px; background: linear-gradient(180deg, rgba(255,255,255,.04), rgba(0,0,0,.08)); }
    .input{ width:100%; background:transparent; border:none; outline:none; color:#fff; font-size:.95rem; letter-spacing:.01em; }
    .input::placeholder{ color:#fff; opacity:.9; font-size:.90rem; }

    /* Small neon counters / dots */
    @keyframes dot { 0%,100%{opacity:.45} 50%{opacity:1} }
    .typing{ display:inline-flex; gap:4px; }
    .d{ width:6px; height:6px; background:#cbd5ff; border-radius:999px; animation: dot 1.4s infinite var(--ease); }
    .d:nth-child(2){ animation-delay:.2s } .d:nth-child(3){ animation-delay:.4s }

    /* Responsive */
    @media (max-width:480px){
      .pill{ min-width:0; width: calc(100vw - 32px); }
      .card{ width: calc(100vw - 32px); }
      .card.open{ height: min(70vh, 540px); }
      .messages{ height: calc(100% - 92px); padding:14px; padding-top:52px; }
      .footer{ padding:12px; }
    }
  `;

  const icons = {
    mic: `<svg viewBox="0 0 24 24" class="icon" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 1a3 3 0 0 0-3 3v6a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3Z"/><path d="M19 10v2a7 7 0 0 1-14 0v-2"/><line x1="12" y1="19" x2="12" y2="23"/><line x1="8" y1="23" x2="16" y2="23"/></svg>`,
    barchart: `<svg viewBox="0 0 24 24" class="icon" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="3" y1="3" x2="3" y2="21"/><line x1="9" y1="9" x2="9" y2="21"/><line x1="15" y1="5" x2="15" y2="21"/><line x1="21" y1="13" x2="21" y2="21"/></svg>`,
    send: `<svg viewBox="0 0 24 24" class="icon" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="22" y1="2" x2="11" y2="13"/><polygon points="22 2 15 22 11 13 2 9 22 2"/></svg>`,
    x: `<svg viewBox="0 0 24 24" class="icon" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>`,
    check: `<svg viewBox="0 0 24 24" class="icon" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 6 9 17l-5-5"/></svg>`,
  };

  const root = document.createElement('div');
  root.className = 'vcw-root';

  const pill = document.createElement('div');
  pill.className = 'pill';
  pill.innerHTML = `
    <div class="row" style="justify-content:space-between;">
      <span class="muted">Ask anything...</span>
      <div class="row">
        <button class="icon-btn" data-action="mic">${icons.mic}</button>
        ${icons.barchart}
      </div>
    </div>
  `;

  const card = document.createElement('div');
  card.className = 'card';
  card.innerHTML = `
    <button class="icon-btn close" data-action="close">${icons.x}</button>
    <div class="messages" data-ref="messages"></div>
    <div class="footer" data-ref="footer"></div>
  `;

  const messagesEl = () => card.querySelector('[data-ref="messages"]');
  const footerEl = () => card.querySelector('[data-ref="footer"]');

  function renderFooter(){
    const f = footerEl(); if(!f) return;
    if(isRecording){
      f.innerHTML = `
        <div class="row">
          <button class="icon-btn" data-action="cancel">${icons.x}</button>
          <div style="flex:1; display:flex; align-items:center; justify-content:center; gap:2px; height:24px;">
            ${Array.from({length:40}).map(()=>'<div class="bar"></div>').join('')}
          </div>
          <button class="icon-btn" data-action="accept">${icons.check}</button>
        </div>
      `;
      // simple bars
      f.querySelectorAll('.bar').forEach((el, i)=>{
        el.style.width = '2px'; el.style.background = '#dbe2ff'; el.style.borderRadius = '1px';
        el.style.animation = `wave ${0.7 + Math.random()*0.6}s var(--ease) ${Math.random()*0.4}s infinite`;
      });
      const waveStyle = document.createElement('style');
      waveStyle.textContent = `@keyframes wave{0%{height:4px}50%{height:${Math.floor(16+Math.random()*20)}px}100%{height:4px}}`;
      shadow.appendChild(waveStyle);
    } else {
      const has = (message||'').trim().length>0;
      f.innerHTML = `
        <div class="row" style="justify-content:space-between;">
          <input class="input" data-ref="input" placeholder="Type your message..." />
          <div class="row">
            <button class="icon-btn" data-action="mic">${icons.mic}</button>
            ${has? `<button class="icon-btn" data-action="send">${icons.send}</button>` : icons.barchart}
          </div>
        </div>
      `;
      const input = f.querySelector('[data-ref="input"]');
      if(input){
        input.value = message;
        setTimeout(()=>{ input.focus(); input.setSelectionRange(input.value.length, input.value.length); },0);
        input.addEventListener('input', (e)=>{ message = e.target.value; renderFooter(); });
        input.addEventListener('keydown', (e)=>{ if(e.key==='Enter' && !e.shiftKey){ e.preventDefault(); send(); }});
      }
    }
    f.querySelectorAll('[data-action]').forEach(btn=>{
      const action = btn.getAttribute('data-action');
      btn.addEventListener('click', (ev)=>{
        ev.stopPropagation();
        if(action==='mic'){ isRecording = true; renderFooter(); }
        else if(action==='send'){ send(); }
        else if(action==='cancel'){ isRecording = false; renderFooter(); }
        else if(action==='accept'){ message = 'This is a simulated voice message converted to text.'; isRecording=false; renderFooter(); }
        else if(action==='close'){ close(); }
      })
    })
  }

  function renderMessages(){
    const m = messagesEl(); if(!m) return; m.innerHTML='';
    if(messages.length===0){
      const empty = document.createElement('div');
      empty.style.textAlign='center'; empty.style.color='var(--muted)'; empty.style.marginTop='28px'; empty.style.fontSize='.92rem';
      empty.textContent='Start a conversation';
      m.appendChild(empty);
    } else {
      for(const msg of messages){
        const row = document.createElement('div'); row.className='msg-row '+(msg.isUser?'right':'left')+' enter';
        const b = document.createElement('div'); b.className='bubble '+(msg.isUser?'me':'ai'); b.textContent=msg.text; row.appendChild(b); m.appendChild(row);
      }
    }
    if(isTyping){
      const row = document.createElement('div'); row.className='msg-row left enter';
      const b = document.createElement('div'); b.className='bubble ai'; b.innerHTML='<span class="typing"><span class="d"></span><span class="d"></span><span class="d"></span></span>';
      row.appendChild(b); m.appendChild(row);
    }
    try{ m.scrollTo({ top: m.scrollHeight, behavior:'smooth' }); } catch{ m.scrollTop = m.scrollHeight; }
  }

  function send(){
    const text = (message||'').trim(); if(!text) return;
    messages.push({ id: String(Date.now()), text, isUser:true, timestamp:new Date() });
    message=''; renderFooter(); isTyping=true; renderMessages();
    setTimeout(()=>{
      messages.push({ id: String(Date.now()+1), text: "I'm a demo AI assistant. I can help you with various tasks and questions.", isUser:false, timestamp:new Date() });
      isTyping=false; renderMessages();
    }, 1100);
  }

  function open(){
    card.classList.add('open'); pill.classList.add('vcw-hidden');
    if(messages.length===0){ messages.push({ id:String(Date.now()), text: "I'm a demo AI assistant. I can help you with various tasks and questions.", isUser:false, timestamp:new Date() }); }
    renderFooter(); renderMessages();
  }
  function close(){
    card.classList.remove('open');
    setTimeout(()=>{ pill.classList.remove('vcw-hidden'); pill.classList.add('in'); }, 180);
  }

  pill.addEventListener('click', open);
  pill.querySelector('[data-action="mic"]').addEventListener('click', (e)=>{ e.stopPropagation(); open(); isRecording=true; renderFooter(); });
  card.querySelector('[data-action="close"]').addEventListener('click', (e)=>{ e.stopPropagation(); close(); });

  root.appendChild(pill); root.appendChild(card);
  shadow.appendChild(style); shadow.appendChild(root);
  document.body.appendChild(host);
  requestAnimationFrame(()=>{ pill.classList.add('in'); });

  // mobile host offset
  try {
    const mq = window.matchMedia('(max-width: 480px)');
    const applyOffset = () => { host.style.bottom = mq.matches ? '16px' : '24px'; };
    mq.addEventListener ? mq.addEventListener('change', applyOffset) : mq.addListener(applyOffset);
    applyOffset();
  } catch {}
})();

