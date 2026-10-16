(function () {
  "use strict";

  var log = document.getElementById("chat-log");
  var form = document.getElementById("chat-form");
  var status = document.getElementById("chat-status");
  if (!log || !form || !log.dataset.events || !window.EventSource || !window.fetch) return;

  var box = form.querySelector("textarea");
  var send = form.querySelector("button");
  var end = document.getElementById("latest");
  var errorLine = null;

  function nearBottom() {
    return log.scrollHeight - log.scrollTop - log.clientHeight < 80;
  }

  function toBottom() {
    log.scrollTop = log.scrollHeight;
  }

  function setStatus(text, state) {
    status.textContent = text;
    status.dataset.state = state;
  }

  function sysLine(text, cls) {
    var p = document.createElement("p");
    p.className = "irc-sys" + (cls ? " " + cls : "");
    p.textContent = text;
    return p;
  }

  function lastDay() {
    var lines = log.querySelectorAll(".irc[data-day]");
    return lines.length ? lines[lines.length - 1].dataset.day : "";
  }

  function showError(text) {
    clearError();
    errorLine = sysLine("-!- " + text, "irc-err");
    log.insertBefore(errorLine, end);
    toBottom();
  }

  function clearError() {
    if (errorLine) {
      errorLine.remove();
      errorLine = null;
    }
  }

  var source = new EventSource(log.dataset.events + "?after=" + encodeURIComponent(log.dataset.last || "0"));
  source.onopen = function () { setStatus("[live]", "live"); };
  source.onerror = function () {
    if (source.readyState === EventSource.CLOSED) {
      setStatus("[disconnected]", "down");
      return;
    }
    setStatus("[reconnecting]", "down");
  };
  source.addEventListener("msg", function (e) {
    if (document.getElementById("m" + e.lastEventId)) return;
    var stick = nearBottom();
    var tpl = document.createElement("template");
    tpl.innerHTML = e.data.trim();
    var line = tpl.content.firstElementChild;
    if (!line) return;
    var empty = log.querySelector(".irc-empty");
    if (empty) empty.remove();
    var prev = lastDay();
    if (prev && line.dataset.day !== prev) {
      log.insertBefore(sysLine("-- day changed to " + line.dataset.day + " --"), end);
    }
    log.insertBefore(line, end);
    if (stick || line.classList.contains("mine")) toBottom();
    if (line.classList.contains("sys") && /has (joined|left)/.test(line.textContent)) {
      var names = document.querySelector(".chat-names .names-head");
      if (names) names.textContent = names.textContent.replace(/^\d+/, function (n) {
        return String(Math.max(0, parseInt(n, 10) + (/joined/.test(line.textContent) ? 1 : -1)));
      });
    }
  });

  function submit() {
    if (!box.value.trim() || send.disabled) return;
    send.disabled = true;
    fetch(form.action, {
      method: "POST",
      body: new FormData(form),
      headers: { "X-Requested-With": "fetch" },
      credentials: "same-origin"
    })
      .then(function (r) {
        if (r.ok) {
          box.value = "";
          clearError();
          return;
        }
        return r.text().then(function (t) {
          showError(r.status === 422 ? t : "couldn't send (" + r.status + ")");
        });
      })
      .catch(function () { showError("couldn't send. check your connection."); })
      .then(function () {
        send.disabled = false;
        box.focus();
      });
  }

  form.addEventListener("submit", function (e) {
    e.preventDefault();
    submit();
  });

  box.addEventListener("keydown", function (e) {
    if (e.key === "Enter" && !e.shiftKey && !e.isComposing) {
      e.preventDefault();
      submit();
    }
  });

  document.documentElement.classList.add("chat-js");
  toBottom();
  box.focus();
})();
