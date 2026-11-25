// Perfila Console — comportamento opcional do painel. Tudo funciona sem este arquivo: links e
// formulários comuns (PRG) e páginas de confirmação. Com ele: diálogos nativos, cópia de ID,
// toasts que somem, linha de tabela clicável e menu abaixo de 960px. Sem eval, sem inline (CSP).
(function () {
  "use strict";

  // Diálogo de confirmação: o gatilho é um link para a página de confirmação; com JS, abre o
  // <dialog> com showModal() (foco preso e Esc pelo navegador) e devolve o foco ao fechar.
  document.addEventListener("click", function (ev) {
    var gatilho = ev.target.closest("[data-dialogo]");
    if (gatilho) {
      var dlg = document.getElementById(gatilho.getAttribute("data-dialogo"));
      if (dlg && typeof dlg.showModal === "function") {
        ev.preventDefault();
        dlg.showModal();
        var foco = dlg.querySelector("[data-foco-inicial]");
        if (foco) { foco.focus(); }
        dlg.addEventListener("close", function () { gatilho.focus(); }, { once: true });
      }
      return;
    }
    // Critério por linhas (v1.6): sem JS, "Adicionar condição" volta ao servidor com mais uma
    // linha; com JS, a linha é clonada na hora, vazia.
    var adicionar = ev.target.closest("[data-adicionar-condicao]");
    if (adicionar) {
      var lista = adicionar.form && adicionar.form.querySelector("[data-condicoes]");
      var ultima = lista && lista.lastElementChild;
      if (ultima) {
        ev.preventDefault();
        var n = lista.children.length;
        var nova = ultima.cloneNode(true);
        nova.setAttribute("data-teste", "condicao-vazia");
        nova.querySelectorAll("select, input").forEach(function (c) { c.value = ""; c.id = c.name + "-" + n; });
        nova.querySelectorAll("label").forEach(function (l) {
          l.htmlFor = l.htmlFor.replace(/-\d+$/, "-" + n);
          l.textContent = l.textContent.replace(/\d+$/, String(n + 1));
        });
        var legenda = nova.querySelector("legend");
        if (legenda) { legenda.textContent = "Condição " + (n + 1); }
        if (!nova.querySelector(".pf-crit__and")) {
          var e = document.createElement("span");
          e.className = "pf-crit__and";
          e.setAttribute("aria-hidden", "true");
          e.textContent = "E";
          nova.insertBefore(e, nova.querySelector("label"));
        }
        lista.appendChild(nova);
        nova.querySelector("select").focus();
      }
      return;
    }
    var fechar = ev.target.closest("[data-fechar-dialogo]");
    if (fechar) {
      ev.preventDefault();
      fechar.closest("dialog").close();
      return;
    }
    var copiar = ev.target.closest("[data-copiar]");
    if (copiar && navigator.clipboard) {
      navigator.clipboard.writeText(copiar.getAttribute("data-copiar")).then(function () { toast("ID copiado"); });
      return;
    }
    var fecharToast = ev.target.closest(".pf-toast__close");
    if (fecharToast) {
      fecharToast.closest(".pf-toast").remove();
      return;
    }
    var menu = ev.target.closest("[data-menu]");
    if (menu) {
      var app = document.getElementById("app");
      var aberto = app.classList.toggle("pf-app--menu-aberto");
      menu.setAttribute("aria-expanded", aberto ? "true" : "false");
      return;
    }
    // Fora do menu aberto (ou num link dele), a gaveta fecha.
    var gaveta = document.getElementById("app");
    if (gaveta && gaveta.classList.contains("pf-app--menu-aberto") && !ev.target.closest(".pf-nav") ) {
      gaveta.classList.remove("pf-app--menu-aberto");
    }
    // Linha inteira clicável; a primeira célula tem um link real para teclado e leitor de tela.
    var linha = ev.target.closest("tr[data-href]");
    if (linha && !ev.target.closest("a, button, input, select")) {
      var link = linha.querySelector("a[href]");
      if (link) { link.click(); }
    }
  });

  document.addEventListener("keydown", function (ev) {
    var app = document.getElementById("app");
    if (ev.key === "Escape" && app && app.classList.contains("pf-app--menu-aberto")) {
      app.classList.remove("pf-app--menu-aberto");
      var b = document.querySelector("[data-menu]");
      if (b) { b.setAttribute("aria-expanded", "false"); b.focus(); }
    }
  });

  // Filtro de consentimento: aplica ao escolher, sem precisar clicar em Buscar (o botão continua
  // valendo sem JavaScript). requestSubmit passa pelo htmx, que troca só a área do painel.
  document.addEventListener("change", function (ev) {
    var campo = ev.target.closest("[data-filtrar]");
    if (campo && campo.form) {
      if (typeof campo.form.requestSubmit === "function") { campo.form.requestSubmit(); } else { campo.form.submit(); }
    }
  });

  function toast(msg) {
    var area = document.querySelector(".pf-toasts");
    if (!area) { return; }
    var t = document.createElement("div");
    t.className = "pf-toast";
    var s = document.createElement("span");
    s.textContent = msg;
    t.appendChild(s);
    area.appendChild(t);
    agendar(t);
  }

  // Some em 5 s; fica enquanto o ponteiro ou o foco estiverem nele.
  function agendar(t) {
    var timer;
    var iniciar = function () { timer = setTimeout(function () { t.remove(); }, 5000); };
    var parar = function () { clearTimeout(timer); };
    t.addEventListener("mouseenter", parar);
    t.addEventListener("focusin", parar);
    t.addEventListener("mouseleave", iniciar);
    t.addEventListener("focusout", iniciar);
    iniciar();
  }

  // Confete: ao voltar do processamento (aviso "Processamento concluído"), um estouro de pontos sai
  // do ingresso da recomendação. Só Web Animations e CSSOM (compatível com a CSP); nada com movimento
  // reduzido.
  function rodarFaixa() {
    var t = Array.prototype.filter.call(document.querySelectorAll(".pf-toast"), function (x) {
      return /^\s*Processamento concluído/.test(x.textContent || "");
    })[0];
    var faixa = document.querySelector(".pf-stage");
    if (t && faixa && !faixa.classList.contains("pf-stage--rodando")) { faixa.classList.add("pf-stage--rodando"); }
  }

  function confete(raiz) {
    if (window.matchMedia && window.matchMedia("(prefers-reduced-motion: reduce)").matches) { return; }
    var t = Array.prototype.filter.call(raiz.querySelectorAll(".pf-toast"), function (x) {
      return !x.hasAttribute("data-confete") && /^\s*Processamento concluído/.test(x.textContent || "");
    })[0];
    var alvo = document.querySelector(".pf-ticket");
    if (!t || !alvo || !alvo.animate) { return; }
    t.setAttribute("data-confete", "1");
    var r = alvo.getBoundingClientRect(), cores = ["var(--brand)", "var(--accent)", "var(--ok)", "var(--stage-lit)"];
    for (var i = 0; i < 22; i++) {
      var p = document.createElement("i");
      p.className = "pf-confete";
      p.style.left = (r.left + r.width / 2) + "px";
      p.style.top = (r.top + r.height / 2) + "px";
      p.style.background = cores[i % 4];
      document.body.appendChild(p);
      var a = Math.random() * Math.PI * 2, d = 70 + Math.random() * 130;
      p.animate([
        { transform: "translate(0,0) scale(1)", opacity: 1 },
        { transform: "translate(" + (Math.cos(a) * d) + "px," + (Math.sin(a) * d + 40) + "px) scale(.2)", opacity: 0 }
      ], { duration: 700 + Math.random() * 400, easing: "cubic-bezier(.2,.8,.2,1)" }).onfinish = (function (el) { return function () { el.remove(); }; })(p);
    }
  }

  function preparar(raiz) {
    raiz.querySelectorAll(".pf-toast").forEach(agendar);
    rodarFaixa();
    confete(document);
    // O botão de copiar já chega visível (não pisca); só some se o navegador não puder copiar.
    if (!navigator.clipboard) {
      raiz.querySelectorAll("[data-copiar]").forEach(function (b) { b.hidden = true; });
    }
    // Remove ?aviso= da barra de endereço para o toast não voltar ao recarregar.
    if (window.history && location.search.indexOf("aviso=") >= 0) {
      var u = new URL(location.href);
      u.searchParams.delete("aviso");
      u.searchParams.delete("n");
      history.replaceState(history.state, "", u.pathname + (u.search ? u.search : "") + u.hash);
    }
  }

  document.addEventListener("DOMContentLoaded", function () { preparar(document); });
  // Depois de uma troca do htmx: prepara o conteúdo novo e leva o foco ao título (leitor de tela)
  // ou ao foco inicial da confirmação.
  var caminho = location.pathname;
  document.addEventListener("htmx:afterSettle", function (ev) {
    preparar(ev.detail.elt || document);
    if (location.pathname === caminho) { return; }
    caminho = location.pathname;
    // Uma página de confirmação pede o foco inicial (autofocus, por exemplo Cancelar); senão, o título.
    var alvo = document.querySelector("#principal [autofocus]");
    if (alvo) { alvo.focus({ preventScroll: true }); return; }
    var h1 = document.querySelector("#principal h1");
    if (h1) { h1.setAttribute("tabindex", "-1"); h1.focus({ preventScroll: true }); }
  });

  // Movimento da identidade Manhã, só com ponteiro e sem alterar o conteúdo: as marcações da régua
  // do processamento perto do cursor crescem, e as formas de fundo acompanham o cursor de leve.
  // Desligado com prefers-reduced-motion (lido a cada evento, então vale se o sistema mudar).
  // Só CSSOM (style.setProperty/transform), compatível com a CSP. Um requestAnimationFrame por quadro.
  var reduz = window.matchMedia ? window.matchMedia("(prefers-reduced-motion: reduce)") : null;
  var base = { "is-major": 34, "is-mid": 20, "is-minor": 10 };
  var quadro = 0, ultimo = null;
  document.addEventListener("pointermove", function (ev) {
    if (reduz && reduz.matches) { return; }
    ultimo = ev;
    if (quadro) { return; }
    quadro = requestAnimationFrame(function () {
      quadro = 0;
      var e = ultimo;
      var formas = document.querySelectorAll(".pf-bg i");
      if (formas.length === 2) {
        var x = e.clientX / window.innerWidth - 0.5, y = e.clientY / window.innerHeight - 0.5;
        formas[0].style.transform = "translate(" + (x * -40) + "px," + (y * -30) + "px)";
        formas[1].style.transform = "translate(" + (x * 50) + "px," + (y * 40) + "px)";
      }
      var regua = e.target && e.target.closest && e.target.closest(".pf-rule");
      if (!regua) { return; }
      regua.querySelectorAll(".pf-ticks__camada i").forEach(function (t) {
        var r = t.getBoundingClientRect(), d = Math.abs(e.clientX - (r.left + 1.5));
        var cls = t.classList.contains("is-major") ? "is-major" : t.classList.contains("is-mid") ? "is-mid" : "is-minor";
        t.style.setProperty("--h", (base[cls] + Math.max(0, 26 * (1 - d / 90))) + "px");
      });
    });
  });
  document.addEventListener("pointerout", function (ev) {
    if (ev.target.closest && ev.target.closest(".pf-rule") && !(ev.relatedTarget && ev.relatedTarget.closest && ev.relatedTarget.closest(".pf-rule"))) {
      document.querySelectorAll(".pf-rule .pf-ticks i").forEach(function (t) { t.style.removeProperty("--h"); });
    }
  });
})();
