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
    // Linha inteira clicável; a primeira célula tem um link real para teclado e leitor de tela.
    var linha = ev.target.closest("tr[data-href]");
    if (linha && !ev.target.closest("a, button, input, select")) {
      var link = linha.querySelector("a[href]");
      if (link) { link.click(); }
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

  function preparar(raiz) {
    raiz.querySelectorAll(".pf-toast").forEach(agendar);
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
  // Depois de uma troca do htmx: prepara o conteúdo novo e leva o foco ao título (leitor de tela).
  var caminho = location.pathname;
  document.addEventListener("htmx:afterSettle", function (ev) {
    preparar(ev.detail.elt || document);
    if (location.pathname === caminho) { return; }
    caminho = location.pathname;
    var h1 = document.querySelector("#principal h1");
    if (h1) { h1.setAttribute("tabindex", "-1"); h1.focus({ preventScroll: true }); }
  });
})();
