/*
  VectorChat Glass Widget (standalone)
  - Single-file vanilla JS chat widget inspired by 1/src/components/GlassChatWidget.tsx
  - No external CSS/JS dependencies. Uses Shadow DOM to isolate styles.
  - Drop this file on a page via a <script src="..."> tag to append the widget.
*/
(function () {
  if (window.__vectorchat_glass_widget_loaded__) return;
  window.__vectorchat_glass_widget_loaded__ = true;

  // State
  let isOpen = false;
  let isRecording = false;
  let message = "";
  /** @type {{id:string,text:string,isUser:boolean,timestamp:Date}[]} */
  let messages = [];

  // Host container (outside Shadow DOM) for fixed positioning
  const host = document.createElement("div");
  host.setAttribute("aria-live", "polite");
  host.style.position = "fixed";
  host.style.left = "50%";
  host.style.bottom = "24px"; // 6 * 4px (tailwind bottom-6)
  host.style.transform = "translateX(-50%)";
  host.style.zIndex = "2147483647"; // Above most elements

  // Shadow root to avoid CSS conflicts
  const shadow = host.attachShadow({ mode: "open" });

  // CSS: Recreate the glass, spacing, animations, and layout
  const style = document.createElement("style");
  style.textContent = `
    :host { all: initial; }
    *, *::before, *::after { box-sizing: border-box; }
    .vcw-root {
      --vcw-bg: rgba(0,0,0,0.60);
      --vcw-bg-hover: rgba(0,0,0,0.70);
      --vcw-border: rgba(255,255,255,0.10);
      --vcw-text: #e5e7eb; /* gray-200 */
      --vcw-muted: #d1d5db; /* gray-300 */
      --vcw-muted-500: #e5e7eb; /* lighter icons by default */
      --vcw-placeholder: #ffffff; /* make input placeholder white */
      --vcw-ease-out: cubic-bezier(.22, .61, .36, 1);
      --vcw-ease-in-out: cubic-bezier(.4, 0, .2, 1);
      --vcw-dur-fast: 180ms;
      --vcw-dur-med: 280ms;
      font-family: ui-sans-serif, system-ui, -apple-system, Segoe UI, Roboto, Helvetica, Arial, "Apple Color Emoji", "Segoe UI Emoji";
      -webkit-font-smoothing: antialiased;
      -moz-osx-font-smoothing: grayscale;
    }
    .vcw-hidden { display: none; }

    /* Common surfaces */
    .vcw-surface {
      background: rgba(12,14,22,0.52); /* neutral glass, no blue */
      -webkit-backdrop-filter: saturate(140%) blur(22px);
      backdrop-filter: saturate(140%) blur(22px);
      border: 1px solid var(--vcw-border);
      border-radius: 14px;
      color: var(--vcw-text);
      box-shadow: 0 16px 48px rgba(4,9,20,0.55), inset 0 1px 0 rgba(255,255,255,0.06);
      position: relative;
      overflow: hidden;
      transform: translateZ(0);
      will-change: transform, opacity;
      contain: layout paint style;
    }
    /* Subtle highlight ring */
    .vcw-surface::after {
      content: ""; position: absolute; inset: 0; pointer-events: none; border-radius: inherit; z-index: 1;
      background: linear-gradient(to bottom, rgba(255,255,255,0.08), rgba(255,255,255,0.02));
      mix-blend-mode: soft-light; opacity: .45;
    }

    /* When open, apply a subtle neutral spotlight and vignette to the panel */
    .vcw-expanded::before {
      content: ""; position: absolute; inset: 0; pointer-events: none; border-radius: inherit; z-index: 0;
      background:
        radial-gradient(40% 18% at 50% 6%, rgba(255,255,255,0.10), rgba(255,255,255,0) 60%),
        radial-gradient(60% 26% at 50% 100%, rgba(0,0,0,0.22), rgba(0,0,0,0) 60%);
    }
    .vcw-row { display: flex; align-items: center; }
    .vcw-justify-between { justify-content: space-between; }
    .vcw-justify-end { justify-content: flex-end; }
    .vcw-justify-start { justify-content: flex-start; }
    .vcw-col { display: flex; flex-direction: column; }

    /* Collapsed pill */
    .vcw-collapsed {
      min-width: 420px; max-width: 500px;
      padding: 12px 16px; cursor: pointer;
      transition: background var(--vcw-dur-fast) var(--vcw-ease-in-out), transform var(--vcw-dur-fast) var(--vcw-ease-out), opacity var(--vcw-dur-fast) var(--vcw-ease-out), box-shadow var(--vcw-dur-fast) var(--vcw-ease-out);
      opacity: 0; transform: translateY(20px) scale(0.995);
    }
    .vcw-collapsed.vcw-in { opacity: 1; transform: translateY(0) scale(1); }
    .vcw-collapsed:hover { background: rgba(12,14,22,0.62); box-shadow: 0 16px 40px rgba(0,0,0,0.45), inset 0 1px 0 rgba(255,255,255,0.08); }
    .vcw-muted { color: var(--vcw-muted); font-size: 0.875rem; }
    .vcw-icon-btn { padding: 4px; border-radius: 8px; transition: background 150ms ease, color 150ms ease; border: none; background: transparent; }
    .vcw-icon-btn:hover { background: rgba(255,255,255,0.10); }
    .vcw-icon { width: 16px; height: 16px; color: var(--vcw-muted-500); display: block; }
    .vcw-icon-btn:hover .vcw-icon { color: #ffffff; }

    /* Expanded panel */
    .vcw-expanded {
      position: relative; overflow: hidden; width: 420px;
      transition: height var(--vcw-dur-med) var(--vcw-ease-in-out), transform var(--vcw-dur-med) var(--vcw-ease-out), opacity var(--vcw-dur-med) var(--vcw-ease-out), box-shadow var(--vcw-dur-med) var(--vcw-ease-out);
      height: 60px; opacity: 0; transform: translateY(20px) scale(0.995);
      will-change: height, transform, opacity;
    }
    .vcw-expanded.vcw-open { height: 500px; opacity: 1; transform: translateY(0) scale(1); }

    .vcw-close-btn { position: absolute; top: 12px; right: 12px; z-index: 10; padding: 6px; border-radius: 10px; }
    .vcw-close-btn:hover { background: rgba(255,255,255,0.10); }

    /* Messages area */
    .vcw-messages {
      position: relative; z-index: 2;
      height: 420px; overflow-y: auto; padding: 16px; padding-top: 48px;
      display: flex; flex-direction: column; gap: 12px;
      scroll-behavior: smooth;
      overscroll-behavior: contain;
      backface-visibility: hidden;
    }
    /* Subtle, clean scrollbar */
    .vcw-messages { scrollbar-width: thin; scrollbar-color: rgba(255,255,255,.15) transparent; }
    .vcw-messages::-webkit-scrollbar { width: 8px; }
    .vcw-messages::-webkit-scrollbar-thumb { background: rgba(255,255,255,.15); border-radius: 8px; }
    .vcw-messages::-webkit-scrollbar-track { background: transparent; }

    .vcw-msg-row { display: flex; }
    .vcw-msg-user { justify-content: flex-end; }
    .vcw-msg-ai { justify-content: flex-start; }
    .vcw-bubble { max-width: 320px; padding: 10px 14px; border-radius: 14px; font-size: 0.94rem; line-height: 1.5; }
    .vcw-bubble-user { background: rgba(255,255,255,0.20); color: #ffffff; }
    .vcw-bubble-ai { background: rgba(255,255,255,0.10); color: #e8eaed; border: 1px solid rgba(255,255,255,0.09); }

    /* Message enter animation */
    @keyframes vcw-msg-in { from { opacity: 0; transform: translateY(8px); } to { opacity: 1; transform: translateY(0); } }
    .vcw-msg-enter { animation: vcw-msg-in var(--vcw-dur-fast) var(--vcw-ease-out) both; will-change: transform, opacity; }

    /* Typing dots */
    @keyframes vcw-dot { 0%,100%{opacity:.4} 50%{opacity:1} }
    .vcw-typing { display: inline-flex; gap: 4px; }
    .vcw-dot { width: 6px; height: 6px; background: #9ca3af; border-radius: 999px; animation: vcw-dot 1.4s infinite ease-in-out; }
    .vcw-dot:nth-child(2) { animation-delay: .2s; }
    .vcw-dot:nth-child(3) { animation-delay: .4s; }

    /* Input area */
    .vcw-footer { padding: 16px; border-top: 1px solid rgba(255,255,255,0.10); position: relative; z-index: 2; }
    .vcw-input {
      background: transparent;
      color: var(--vcw-text);
      caret-color: var(--vcw-text);
      width: 100%;
      outline: none;
      font-size: 0.95rem;
      letter-spacing: .01em;
      border: none;
      box-shadow: none;
      -webkit-appearance: none;
      appearance: none;
    }
    .vcw-input:focus { outline: none; box-shadow: none; border: none; }
    .vcw-root button { border: none; outline: none; }
    .vcw-input::placeholder { color: var(--vcw-placeholder); opacity: .92; font-size: 0.90rem; }
    .vcw-placeholder { color: var(--vcw-placeholder); }

    /* Waveform */
    .vcw-wave { display: flex; align-items: center; justify-content: center; gap: 2px; height: 24px; }
    @keyframes vcw-wave {
      0% { height: 4px; }
      50% { height: var(--h, 16px); }
      100% { height: 4px; }
    }
    .vcw-bar { width: 2px; background: #d1d5db; border-radius: 1px; animation: vcw-wave var(--dur, .9s) var(--vcw-ease-in-out) var(--delay, 0s) infinite; will-change: height; }

    /* Utility */
    .vcw-space-x > * + * { margin-left: 8px; }
    .vcw-space-x-tight > * + * { margin-left: 6px; }

    /* Responsive adjustments for mobile */
    @media (max-width: 480px) {
      .vcw-collapsed { min-width: 0; width: calc(100vw - 32px); }
      .vcw-expanded { width: calc(100vw - 32px); }
      .vcw-expanded.vcw-open { height: min(70vh, 520px); }
      .vcw-messages { height: calc(100% - 88px); padding: 12px; padding-top: 40px; gap: 10px; }
      .vcw-footer { padding: 12px; }
    }
    /* Motion reduction support */
    @media (prefers-reduced-motion: reduce) {
      .vcw-collapsed, .vcw-expanded { transition: none !important; }
      .vcw-msg-enter { animation: none !important; }
      .vcw-bar { animation: none !important; }
      .vcw-messages { scroll-behavior: auto; }
    }
  `;

  // Inline SVG icons (minimal Lucide-like)
  const icons = {
    mic: `<svg viewBox="0 0 24 24" class="vcw-icon" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 1a3 3 0 0 0-3 3v6a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3Z"/><path d="M19 10v2a7 7 0 0 1-14 0v-2"/><line x1="12" y1="19" x2="12" y2="23"/><line x1="8" y1="23" x2="16" y2="23"/></svg>`,
    barchart: `<svg viewBox="0 0 24 24" class="vcw-icon" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="3" y1="3" x2="3" y2="21"/><line x1="9" y1="9" x2="9" y2="21"/><line x1="15" y1="5" x2="15" y2="21"/><line x1="21" y1="13" x2="21" y2="21"/></svg>`,
    send: `<svg viewBox="0 0 24 24" class="vcw-icon" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="22" y1="2" x2="11" y2="13"/><polygon points="22 2 15 22 11 13 2 9 22 2"/></svg>`,
    x: `<svg viewBox="0 0 24 24" class="vcw-icon" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>`,
    check: `<svg viewBox="0 0 24 24" class="vcw-icon" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 6 9 17l-5-5"/></svg>`,
  };

  // Root structure inside shadow
  const root = document.createElement("div");
  root.className = "vcw-root";

  // Collapsed node
  const collapsed = document.createElement("div");
  collapsed.className = "vcw-surface vcw-collapsed";
  collapsed.innerHTML = `
    <div class="vcw-row vcw-justify-between">
      <span class="vcw-muted">Ask anything...</span>
      <div class="vcw-row vcw-space-x">
        <button class="vcw-icon-btn" data-action="mic-collapsed" aria-label="Record">
          ${icons.mic}
        </button>
        ${icons.barchart}
      </div>
    </div>
  `;

  // Expanded node
  const expanded = document.createElement("div");
  expanded.className = "vcw-surface vcw-expanded";
  expanded.innerHTML = `
    <button class="vcw-close-btn vcw-icon-btn" data-action="close" aria-label="Close">${icons.x}</button>
    <div class="vcw-messages" data-ref="messages"></div>
    <div class="vcw-footer" data-ref="footer"></div>
  `;

  // Messages and footer refs
  const messagesEl = () => expanded.querySelector('[data-ref="messages"]');
  const footerEl = () => expanded.querySelector('[data-ref="footer"]');

  // Render functions
  function renderCollapsedIn() {
    collapsed.classList.add("vcw-in");
  }

  function renderFooter() {
    const f = footerEl();
    if (!f) return;
    if (isRecording) {
      // Recording UI
      f.innerHTML = `
        <div class="vcw-row vcw-space-x-tight">
          <button class="vcw-icon-btn" data-action="cancel-record" aria-label="Cancel">${icons.x}</button>
          <div class="vcw-col" style="flex:1;">
            <div class="vcw-wave" data-ref="wave"></div>
          </div>
          <button class="vcw-icon-btn" data-action="accept-record" aria-label="Accept">${icons.check}</button>
        </div>
      `;
      const wave = f.querySelector('[data-ref="wave"]');
      if (wave) {
        for (let i = 0; i < 40; i++) {
          const bar = document.createElement("div");
          const max = 24 + Math.round(Math.random() * 20); // 24-44px
          const dur = 0.5 + Math.random() * 0.5; // .5 - 1.0s
          const delay = Math.random() * 0.5; // 0 - .5s
          bar.className = "vcw-bar";
          bar.style.setProperty("--h", `${max}px`);
          bar.style.setProperty("--dur", `${dur}s`);
          bar.style.setProperty("--delay", `${delay}s`);
          wave.appendChild(bar);
        }
      }
    } else {
      // Normal input UI
      const hasText = (message || "").trim().length > 0;
      f.innerHTML = `
        <div class="vcw-row vcw-space-x">
          <div style="flex:1; position:relative;">
            <input type="text" class="vcw-input" data-ref="input" placeholder="Type your message..." />
          </div>
          <div class="vcw-row vcw-space-x">
            <button class="vcw-icon-btn" data-action="mic" aria-label="Record">${icons.mic}</button>
            ${hasText ? `<button class="vcw-icon-btn" data-action="send" aria-label="Send">${icons.send}</button>` : icons.barchart}
          </div>
        </div>
      `;
      const input = f.querySelector('[data-ref="input"]');
      if (input) {
        input.value = message;
        // Autofocus input when opening/returning from record mode
        setTimeout(() => {
          input.focus();
          input.setSelectionRange(input.value.length, input.value.length);
        }, 0);
        input.addEventListener("input", (e) => {
          message = e.target.value;
          // Re-render footer to swap send/barchart icon
          renderFooter();
        });
        input.addEventListener("keydown", (e) => {
          if (e.key === "Enter" && !e.shiftKey) {
            e.preventDefault();
            handleSend();
          }
        });
      }
    }

    // Bind footer buttons
    f.querySelectorAll("[data-action]").forEach((btn) => {
      const action = btn.getAttribute("data-action");
      btn.addEventListener("click", (ev) => {
        ev.stopPropagation();
        switch (action) {
          case "mic":
            isRecording = true;
            renderFooter();
            break;
          case "send":
            handleSend();
            break;
          case "cancel-record":
            isRecording = false;
            renderFooter();
            break;
          case "accept-record":
            // Simulate transcription
            message = "This is a simulated voice message converted to text.";
            isRecording = false;
            renderFooter();
            break;
          case "close":
            // handled on close button outside
            break;
        }
      });
    });
  }

  function renderMessages() {
    const m = messagesEl();
    if (!m) return;
    m.innerHTML = "";
    if (messages.length === 0) {
      const empty = document.createElement("div");
      empty.style.textAlign = "center";
      empty.style.color = "#9ca3af"; // gray-400
      empty.style.marginTop = "32px";
      empty.style.fontSize = "0.9rem";
      empty.textContent = "Start a conversation";
      m.appendChild(empty);
    } else {
      for (const msg of messages) {
        const row = document.createElement("div");
        row.className =
          "vcw-msg-row vcw-msg-enter " +
          (msg.isUser ? "vcw-msg-user" : "vcw-msg-ai");
        const bubble = document.createElement("div");
        bubble.className =
          "vcw-bubble " + (msg.isUser ? "vcw-bubble-user" : "vcw-bubble-ai");
        bubble.textContent = msg.text;
        row.appendChild(bubble);
        m.appendChild(row);
      }
    }
    // If typing, append typing indicator
    if (isTyping) {
      const row = document.createElement("div");
      row.className = "vcw-msg-row vcw-msg-ai vcw-msg-enter";
      const wrap = document.createElement("div");
      wrap.className = "vcw-bubble vcw-bubble-ai";
      wrap.innerHTML = `<span class="vcw-typing"><span class="vcw-dot"></span><span class="vcw-dot"></span><span class="vcw-dot"></span></span>`;
      row.appendChild(wrap);
      m.appendChild(row);
    }
    // Scroll to bottom with smooth behavior
    try {
      m.scrollTo({ top: m.scrollHeight, behavior: 'smooth' });
    } catch (_) {
      m.scrollTop = m.scrollHeight;
    }
  }

  // Send flow with simulated AI response
  let isTyping = false;
  function handleSend() {
    const text = (message || "").trim();
    if (!text) return;
    const userMessage = {
      id: String(Date.now()),
      text,
      isUser: true,
      timestamp: new Date(),
    };
    messages.push(userMessage);
    message = "";
    renderFooter();
    isTyping = true;
    renderMessages();
    setTimeout(() => {
      const aiMessage = {
        id: String(Date.now() + 1),
        text: "I'm a demo AI assistant. I can help you with various tasks and questions.",
        isUser: false,
        timestamp: new Date(),
      };
      messages.push(aiMessage);
      isTyping = false;
      renderMessages();
    }, 1200);
  }

  // Toggle open/close
  function openWidget() {
    isOpen = true;
    expanded.classList.add("vcw-open");
    collapsed.classList.add("vcw-hidden");
    // Seed a friendly AI greeting on first open
    if (messages.length === 0) {
      messages.push({
        id: String(Date.now()),
        text: "I'm a demo AI assistant. I can help you with various tasks and questions.",
        isUser: false,
        timestamp: new Date(),
      });
    }
    renderFooter();
    renderMessages();
  }
  function closeWidget() {
    isOpen = false;
    expanded.classList.remove("vcw-open");
    // After transition, show collapsed again
    setTimeout(() => {
      collapsed.classList.remove("vcw-hidden");
      renderCollapsedIn();
    }, 200);
  }

  // Wire events
  collapsed.addEventListener("click", () => {
    openWidget();
  });
  // Collapsed mic should also open and set recording
  collapsed
    .querySelector('[data-action="mic-collapsed"]')
    .addEventListener("click", (e) => {
      e.stopPropagation();
      openWidget();
      isRecording = true;
      renderFooter();
    });
  expanded
    .querySelector('[data-action="close"]')
    .addEventListener("click", (e) => {
      e.stopPropagation();
      closeWidget();
    });

  // Mount structure
  root.appendChild(collapsed);
  root.appendChild(expanded);
  shadow.appendChild(style);
  shadow.appendChild(root);
  document.body.appendChild(host);

  // Initial animations
  requestAnimationFrame(() => {
    renderCollapsedIn();
  });

  // Responsive host bottom spacing (mobile vs desktop)
  try {
    const mq = window.matchMedia('(max-width: 480px)');
    const applyHostOffset = () => {
      host.style.bottom = mq.matches ? '16px' : '24px';
    };
    if (mq.addEventListener) mq.addEventListener('change', applyHostOffset);
    else if (mq.addListener) mq.addListener(applyHostOffset);
    applyHostOffset();
  } catch (_) { /* noop */ }
})();
