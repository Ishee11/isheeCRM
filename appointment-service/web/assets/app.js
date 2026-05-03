const state = {
  view: "dashboard",
  visits: [],
  services: [],
  membershipTypes: [],
};

const titles = {
  dashboard: ["CRM", "Обзор"],
  schedule: ["Записи", "Расписание"],
  clients: ["Карточки", "Клиенты"],
  catalog: ["Каталог", "Услуги"],
  memberships: ["Продажи", "Абонементы"],
};

const money = new Intl.NumberFormat("ru-RU", {
  style: "currency",
  currency: "RUB",
  maximumFractionDigits: 0,
});

const dateTime = new Intl.DateTimeFormat("ru-RU", {
  day: "2-digit",
  month: "short",
  hour: "2-digit",
  minute: "2-digit",
});

function qs(selector) {
  return document.querySelector(selector);
}

function qsa(selector) {
  return Array.from(document.querySelectorAll(selector));
}

function text(value, fallback = "-") {
  if (value === null || value === undefined || value === "") return fallback;
  return String(value);
}

function escapeHtml(value) {
  return text(value, "").replace(/[&<>"']/g, (char) => ({
    "&": "&amp;",
    "<": "&lt;",
    ">": "&gt;",
    '"': "&quot;",
    "'": "&#39;",
  })[char]);
}

function formatDate(value) {
  if (!value) return "-";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return text(value);
  return dateTime.format(date);
}

function formatMoney(value) {
  const number = Number(value || 0);
  return money.format(number);
}

function statusClass(value) {
  const normalized = text(value, "").toLowerCase();
  if (["done", "paid", "success", "completed"].includes(normalized)) return "ok";
  if (["cancelled", "canceled", "failed", "unpaid"].includes(normalized)) return "bad";
  return "warn";
}

async function api(path, options = {}) {
  const response = await fetch(path, {
    headers: {
      "Content-Type": "application/json",
      ...(options.headers || {}),
    },
    ...options,
  });

  const contentType = response.headers.get("content-type") || "";
  const payload = contentType.includes("application/json")
    ? await response.json().catch(() => ({}))
    : await response.text();

  if (!response.ok) {
    const message = typeof payload === "string"
      ? payload
      : payload.details || payload.error || response.statusText;
    throw new Error(message);
  }

  return payload;
}

function toast(message, type = "info") {
  const region = qs("#toastRegion");
  const item = document.createElement("div");
  item.className = `toast ${type === "error" ? "error" : ""}`;
  item.textContent = message;
  region.appendChild(item);
  setTimeout(() => item.remove(), 4200);
}

function setView(view) {
  state.view = view;
  qsa(".nav-item").forEach((item) => item.classList.toggle("active", item.dataset.view === view));
  qsa(".view").forEach((item) => item.classList.toggle("active", item.id === `view-${view}`));
  const [eyebrow, title] = titles[view] || titles.dashboard;
  qs("#currentEyebrow").textContent = eyebrow;
  qs("#currentTitle").textContent = title;
  window.location.hash = view;
  refreshCurrentView();
}

async function checkVersion() {
  const dot = qs("#serviceDot");
  const status = qs("#serviceStatus");
  const build = qs("#serviceBuild");

  try {
    const version = await api("/version");
    dot.className = "status-dot ok";
    status.textContent = "Онлайн";
    build.textContent = version.image_tag ? `build ${version.image_tag}` : "build unknown";
  } catch (error) {
    dot.className = "status-dot fail";
    status.textContent = "Недоступен";
    build.textContent = error.message;
  }
}

async function loadStatistics() {
  try {
    const stats = await api("/statistics/current-month");
    qs("#metricVisits").textContent = text(stats.total_visits, "0");
    qs("#metricRevenue").textContent = formatMoney(stats.total_earnings);
    qs("#metricServices").textContent = text(stats.total_services, "0");
    qs("#metricSubscriptions").textContent = text(stats.total_subscriptions, "0");
  } catch (error) {
    toast(`Статистика: ${error.message}`, "error");
  }
}

function visitQuery() {
  const params = new URLSearchParams();
  const from = qs("#visitFrom").value;
  const to = qs("#visitTo").value;
  const status = qs("#visitStatus").value;
  const clientId = qs("#visitClientId").value.trim();
  if (from) params.set("from", from);
  if (to) params.set("to", to);
  if (status) params.set("status", status);
  if (clientId) params.set("client_id", clientId);
  return params.toString();
}

async function loadVisits() {
  const tbody = qs("#visitsTable");
  if (tbody) {
    tbody.innerHTML = `<tr><td colspan="6">Загрузка...</td></tr>`;
  }

  try {
    const query = visitQuery();
    const data = await api(`/visits/${query ? `?${query}` : ""}`);
    state.visits = Array.isArray(data) ? data : data.items || [];
    renderVisits();
    renderDashboardVisits();
  } catch (error) {
    if (tbody) {
      tbody.innerHTML = `<tr><td colspan="6">Ошибка: ${escapeHtml(error.message)}</td></tr>`;
    }
    toast(`Записи: ${error.message}`, "error");
  }
}

function renderVisits() {
  const tbody = qs("#visitsTable");
  if (!tbody) return;
  qs("#visitsCaption").textContent = `${state.visits.length} записей`;

  if (!state.visits.length) {
    tbody.innerHTML = `<tr><td colspan="6">Записей нет</td></tr>`;
    return;
  }

  tbody.innerHTML = state.visits.map((visit) => `
    <tr>
      <td>${escapeHtml(visit.id)}</td>
      <td>${escapeHtml(visit.client_name || visit.client_id)}</td>
      <td>${escapeHtml(visit.service_name || visit.service_id)}</td>
      <td>${escapeHtml(formatDate(visit.start_time))}</td>
      <td><span class="pill ${statusClass(visit.appointment_status)}">${escapeHtml(visit.appointment_status || "scheduled")}</span></td>
      <td><span class="pill ${statusClass(visit.payment_status)}">${escapeHtml(visit.payment_status || "unknown")}</span></td>
    </tr>
  `).join("");
}

function renderDashboardVisits() {
  const target = qs("#dashboardVisits");
  if (!target) return;

  const items = state.visits.slice(0, 6);
  if (!items.length) {
    target.innerHTML = `<div class="empty-state">Записей нет</div>`;
    return;
  }

  target.innerHTML = items.map((visit) => `
    <article class="list-row">
      <div>
        <strong>${escapeHtml(visit.client_name || `Клиент ${visit.client_id || ""}`)}</strong>
        <span>${escapeHtml(visit.service_name || "Услуга")} · ${escapeHtml(formatDate(visit.start_time))}</span>
      </div>
      <span class="pill ${statusClass(visit.appointment_status)}">${escapeHtml(visit.appointment_status || "scheduled")}</span>
    </article>
  `).join("");
}

async function loadServices() {
  try {
    const data = await api("/services/");
    state.services = Array.isArray(data) ? data : data.items || [];
    renderServices();
    renderDashboardServices();
  } catch (error) {
    toast(`Услуги: ${error.message}`, "error");
    qs("#servicesGrid").innerHTML = `<div class="empty-state">Каталог недоступен</div>`;
  }
}

function renderServices() {
  const target = qs("#servicesGrid");
  if (!target) return;
  qs("#servicesCaption").textContent = `${state.services.length} услуг`;

  if (!state.services.length) {
    target.innerHTML = `<div class="empty-state">Услуги не найдены</div>`;
    return;
  }

  target.innerHTML = state.services.map((service) => `
    <article class="service-card">
      <strong>${escapeHtml(service.name)}</strong>
      <span>${escapeHtml(service.duration)} мин</span>
      <div class="kv-grid">
        <div class="kv">
          <span>ID</span>
          <strong>${escapeHtml(service.service_id || service.id)}</strong>
        </div>
        <div class="kv">
          <span>Цена</span>
          <strong>${escapeHtml(formatMoney(service.price))}</strong>
        </div>
      </div>
    </article>
  `).join("");
}

function renderDashboardServices() {
  const target = qs("#dashboardServices");
  if (!target) return;
  const items = state.services.slice(0, 5);
  if (!items.length) {
    target.innerHTML = `<div class="empty-state">Каталог пуст</div>`;
    return;
  }

  target.innerHTML = items.map((service) => `
    <article class="list-row">
      <div>
        <strong>${escapeHtml(service.name)}</strong>
        <span>${escapeHtml(service.duration)} мин</span>
      </div>
      <strong>${escapeHtml(formatMoney(service.price))}</strong>
    </article>
  `).join("");
}

async function createService(event) {
  event.preventDefault();
  const form = event.currentTarget;
  const body = {
    name: form.name.value.trim(),
    duration: Number(form.duration.value),
    price: Number(form.price.value),
  };

  try {
    await api("/services/add", {
      method: "POST",
      body: JSON.stringify(body),
    });
    form.reset();
    toast("Услуга добавлена");
    await loadServices();
  } catch (error) {
    toast(`Не удалось добавить услугу: ${error.message}`, "error");
  }
}

async function findClient() {
  const phone = qs("#clientPhone").value.trim();
  if (!phone) {
    toast("Введите телефон", "error");
    return;
  }

  try {
    const result = await api(`/clients/find?phone=${encodeURIComponent(phone)}`);
    qs("#clientId").value = result.client_id || "";
    await loadClient(result.client_id);
  } catch (error) {
    toast(`Клиент: ${error.message}`, "error");
  }
}

async function loadClient(idFromSearch) {
  const id = idFromSearch || qs("#clientId").value.trim();
  if (!id) {
    toast("Введите ID клиента", "error");
    return;
  }

  try {
    const client = await api(`/clients/info?client_id=${encodeURIComponent(id)}`);
    renderClient(client, id);
  } catch (error) {
    toast(`Карточка: ${error.message}`, "error");
  }
}

function renderClient(client, id) {
  qs("#clientCardCaption").textContent = `ID ${id}`;
  qs("#clientCard").className = "client-card";
  qs("#clientCard").innerHTML = `
    <h3 class="client-name">${escapeHtml(client.name)}</h3>
    <div>${escapeHtml(client.phone || "")}</div>
    <div class="kv-grid">
      <div class="kv"><span>Визиты</span><strong>${escapeHtml(client.visit_count || 0)}</strong></div>
      <div class="kv"><span>Потрачено</span><strong>${escapeHtml(formatMoney(client.spent))}</strong></div>
      <div class="kv"><span>Оплачено</span><strong>${escapeHtml(formatMoney(client.paid))}</strong></div>
      <div class="kv"><span>Скидка</span><strong>${escapeHtml(client.discount || 0)}%</strong></div>
      <div class="kv"><span>Первый визит</span><strong>${escapeHtml(formatDate(client.first_visit))}</strong></div>
      <div class="kv"><span>Последний визит</span><strong>${escapeHtml(formatDate(client.last_visit))}</strong></div>
    </div>
  `;
}

async function loadMembershipTypes() {
  const target = qs("#membershipTypes");
  if (target) target.innerHTML = `<div class="empty-state">Загрузка...</div>`;

  try {
    const data = await api("/subscriptions/types");
    state.membershipTypes = Array.isArray(data) ? data : data.items || [];
    renderMembershipTypes();
  } catch (error) {
    toast(`Абонементы: ${error.message}`, "error");
    if (target) target.innerHTML = `<div class="empty-state">Каталог недоступен</div>`;
  }
}

function renderMembershipTypes() {
  const target = qs("#membershipTypes");
  if (!target) return;
  qs("#membershipCaption").textContent = `${state.membershipTypes.length} типов`;

  if (!state.membershipTypes.length) {
    target.innerHTML = `<div class="empty-state">Типов нет</div>`;
    return;
  }

  target.innerHTML = state.membershipTypes.map((item) => `
    <article class="membership-card">
      <strong>${escapeHtml(item.name)}</strong>
      <span>${escapeHtml(item.sessions_count)} занятий</span>
      <div class="kv-grid">
        <div class="kv"><span>ID</span><strong>${escapeHtml(item.subscription_types_id || item.id)}</strong></div>
        <div class="kv"><span>Цена</span><strong>${escapeHtml(formatMoney(item.cost))}</strong></div>
      </div>
    </article>
  `).join("");
}

async function loadClientMemberships() {
  const id = qs("#membershipClientId").value.trim();
  if (!id) {
    toast("Введите ID клиента", "error");
    return;
  }

  const target = qs("#clientMemberships");
  target.innerHTML = `<div class="empty-state">Загрузка...</div>`;

  try {
    const data = await api(`/subscriptions/client?client_id=${encodeURIComponent(id)}`);
    const items = data.subscriptions || [];
    if (!items.length) {
      target.innerHTML = `<div class="empty-state">Абонементов нет</div>`;
      return;
    }
    target.innerHTML = items.map((item) => `
      <article class="list-row">
        <div>
          <strong>Абонемент ${escapeHtml(item.subscription_id)}</strong>
          <span>Остаток занятий</span>
        </div>
        <strong>${escapeHtml(item.current_balance)}</strong>
      </article>
    `).join("");
  } catch (error) {
    toast(`Абонементы клиента: ${error.message}`, "error");
    target.innerHTML = `<div class="empty-state">Ошибка загрузки</div>`;
  }
}

function refreshCurrentView() {
  if (state.view === "dashboard") {
    loadStatistics();
    loadVisits();
    loadServices();
  }
  if (state.view === "schedule") loadVisits();
  if (state.view === "catalog") loadServices();
  if (state.view === "memberships") loadMembershipTypes();
}

function bindEvents() {
  qsa(".nav-item").forEach((item) => {
    item.addEventListener("click", () => setView(item.dataset.view));
  });
  qsa("[data-view-jump]").forEach((item) => {
    item.addEventListener("click", () => setView(item.dataset.viewJump));
  });
  qs("#refreshButton").addEventListener("click", refreshCurrentView);
  qs("#loadVisitsButton").addEventListener("click", loadVisits);
  qs("#reloadServicesButton").addEventListener("click", loadServices);
  qs("#serviceForm").addEventListener("submit", createService);
  qs("#findClientButton").addEventListener("click", findClient);
  qs("#loadClientButton").addEventListener("click", () => loadClient());
  qs("#reloadMembershipsButton").addEventListener("click", loadMembershipTypes);
  qs("#loadMembershipClientButton").addEventListener("click", loadClientMemberships);
}

function initDates() {
  const now = new Date();
  const start = new Date(now);
  start.setDate(now.getDate() - 7);
  const end = new Date(now);
  end.setDate(now.getDate() + 14);
  qs("#visitFrom").value = start.toISOString().slice(0, 10);
  qs("#visitTo").value = end.toISOString().slice(0, 10);
}

function init() {
  bindEvents();
  initDates();
  checkVersion();
  const initial = window.location.hash.replace("#", "") || "dashboard";
  setView(titles[initial] ? initial : "dashboard");
}

document.addEventListener("DOMContentLoaded", init);
