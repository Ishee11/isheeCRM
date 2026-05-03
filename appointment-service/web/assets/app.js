const state = {
  view: "schedule",
  range: "today",
  visits: [],
  services: [],
  membershipTypes: [],
  selectedVisit: null,
  selectedClient: null,
};

const titles = {
  schedule: ["Рабочий день", "Журнал записи"],
  clients: ["Карточки", "Клиенты"],
  sales: ["Абонементы", "Продажи"],
  services: ["Каталог", "Услуги"],
  reports: ["Показатели", "Статистика"],
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

const timeOnly = new Intl.DateTimeFormat("ru-RU", {
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

function formatTime(value) {
  if (!value) return "-";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "-";
  return timeOnly.format(date);
}

function appointmentHour(value) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return 9;
  return date.getHours();
}

function toDateInputValue(date) {
  return [
    date.getFullYear(),
    String(date.getMonth() + 1).padStart(2, "0"),
    String(date.getDate()).padStart(2, "0"),
  ].join("-");
}

function toDateTimeInputValue(date) {
  const local = new Date(date.getTime() - date.getTimezoneOffset() * 60000);
  return local.toISOString().slice(0, 16);
}

function formatTimezoneOffset(minutes) {
  const sign = minutes >= 0 ? "+" : "-";
  const absolute = Math.abs(minutes);
  const hours = Math.floor(absolute / 60);
  const rest = absolute % 60;
  return rest ? `${sign}${hours}:${String(rest).padStart(2, "0")}` : `${sign}${hours}`;
}

function currentTimezoneLabel() {
  const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone || "локальный часовой пояс";
  const utcOffsetMinutes = -new Date().getTimezoneOffset();
  const moscowOffsetMinutes = utcOffsetMinutes - 180;
  return `${timezone}, UTC${formatTimezoneOffset(utcOffsetMinutes)} / МСК${formatTimezoneOffset(moscowOffsetMinutes)}`;
}

function formatMoney(value) {
  return money.format(Number(value || 0));
}

function localDateTimeToISO(value) {
  if (!value) return "";
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? value : date.toISOString();
}

function dateToISO(value) {
  if (!value) return "";
  const date = new Date(`${value}T12:00:00`);
  return Number.isNaN(date.getTime()) ? value : date.toISOString();
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
  const [eyebrow, title] = titles[view] || titles.schedule;
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

function scheduleQuery() {
  const params = new URLSearchParams();
  params.set("format", "list");
  const selected = qs("#scheduleDate").value;
  const status = qs("#visitStatus").value;

  if (state.range === "today" && selected) {
    params.set("from", selected);
    params.set("to", selected);
  }

  if (state.range === "week" && selected) {
    const from = new Date(`${selected}T00:00:00`);
    const to = new Date(from);
    to.setDate(from.getDate() + 6);
    params.set("from", toDateInputValue(from));
    params.set("to", toDateInputValue(to));
  }

  if (status) params.set("status", status);
  return params.toString();
}

async function loadVisits() {
  const timeline = qs("#visitTimeline");
  timeline.innerHTML = `<div class="empty-state">Загрузка расписания</div>`;

  try {
    const query = scheduleQuery();
    const data = await api(`/visits/?${query}`);
    state.visits = (Array.isArray(data) ? data : data.items || []).sort((a, b) => {
      return new Date(a.start_time || 0) - new Date(b.start_time || 0);
    });
    renderVisits();
    if (state.selectedVisit) {
      const fresh = state.visits.find((visit) => String(visit.id) === String(state.selectedVisit.id));
      state.selectedVisit = fresh || null;
      renderAppointmentContext();
    }
  } catch (error) {
    timeline.innerHTML = `<div class="empty-state">Расписание недоступно</div>`;
    toast(`Записи: ${error.message}`, "error");
  }
}

function renderVisits() {
  const timeline = qs("#visitTimeline");
  qs("#visitsCaption").textContent = `${state.visits.length} записей`;
  renderJournalTotals();

  const hours = journalHours();
  timeline.innerHTML = `
    <div class="journal-grid">
      <div class="journal-head journal-time-head">Время</div>
      <div class="journal-head">Все записи</div>
      ${hours.map((hour) => renderJournalHour(hour)).join("")}
    </div>
  `;

  qsa("[data-visit-id]").forEach((item) => {
    item.addEventListener("click", () => selectVisit(item.dataset.visitId));
  });
  qsa("[data-slot-hour]").forEach((item) => {
    item.addEventListener("click", () => openAppointmentDialog(Number(item.dataset.slotHour)));
  });
}

function journalHours() {
  const businessStart = 8;
  const businessEnd = 21;
  const visitHours = state.visits
    .map((visit) => appointmentHour(visit.start_time))
    .filter((hour) => hour >= 0 && hour <= 23);
  const start = Math.min(businessStart, ...visitHours);
  const end = Math.max(businessEnd, ...visitHours);
  return Array.from({ length: end - start + 1 }, (_, index) => start + index);
}

function renderJournalTotals() {
  const target = qs("#journalTotals");
  const unpaid = state.visits.filter((visit) => (visit.payment_status || "unpaid") === "unpaid").length;
  const paid = state.visits.filter((visit) => (visit.payment_status || "") === "paid").length;
  target.innerHTML = `
    <span class="mini-stat"><strong>${state.visits.length}</strong> записей</span>
    <span class="mini-stat"><strong>${paid}</strong> оплачено</span>
    <span class="mini-stat"><strong>${unpaid}</strong> без оплаты</span>
  `;
}

function renderJournalHour(hour) {
  const hourVisits = state.visits.filter((visit) => appointmentHour(visit.start_time) === hour);
  const label = `${String(hour).padStart(2, "0")}:00`;

  return `
    <div class="journal-time">${label}</div>
    <div class="journal-cell">
      ${hourVisits.length ? hourVisits.map(renderJournalVisit).join("") : `<button class="empty-slot" data-slot-hour="${hour}" type="button">Свободно</button>`}
    </div>
  `;
}

function renderJournalVisit(visit) {
  const active = state.selectedVisit && String(state.selectedVisit.id) === String(visit.id);
  return `
    <button class="journal-visit ${active ? "active" : ""}" data-visit-id="${escapeHtml(visit.id)}" type="button">
      <span class="visit-time">${escapeHtml(formatTime(visit.start_time))}</span>
      <span class="visit-title">
        <strong>${escapeHtml(visit.client_name || `Клиент #${visit.client_id || ""}`)}</strong>
        <small>${escapeHtml(visit.service_name || `Услуга #${visit.service_id || ""}`)}</small>
      </span>
      <span class="pill ${statusClass(visit.payment_status)}">${escapeHtml(visit.payment_status || "unpaid")}</span>
    </button>
  `;
}

function selectVisit(id) {
  state.selectedVisit = state.visits.find((visit) => String(visit.id) === String(id)) || null;
  renderVisits();
  renderAppointmentContext();
}

function renderAppointmentContext() {
  const target = qs("#appointmentContext");
  const caption = qs("#appointmentContextCaption");
  const visit = state.selectedVisit;

  if (!visit) {
    caption.textContent = "Выберите запись в расписании";
    target.className = "empty-state";
    target.textContent = "Запись не выбрана";
    return;
  }

  caption.textContent = `Запись #${visit.id}`;
  target.className = "context-body";
  target.innerHTML = `
    <div class="context-title">
      <strong>${escapeHtml(visit.client_name || `Клиент #${visit.client_id || ""}`)}</strong>
      <span>${escapeHtml(visit.service_name || `Услуга #${visit.service_id || ""}`)}</span>
    </div>
    <div class="visit-tabs">
      <span class="active">Визит</span>
      <span>Клиент</span>
      <span>Оплата</span>
      <span>Абонемент</span>
    </div>
    <div class="kv-grid">
      <div class="kv"><span>Время</span><strong>${escapeHtml(formatDate(visit.start_time))}</strong></div>
      <div class="kv"><span>Оплата</span><strong>${escapeHtml(visit.payment_status || "unpaid")}</strong></div>
      <div class="kv"><span>Статус</span><strong>${escapeHtml(visit.appointment_status || "scheduled")}</strong></div>
      <div class="kv"><span>ID клиента</span><strong>${escapeHtml(visit.client_id || "-")}</strong></div>
    </div>
    <form class="form-stack" id="contextPaymentForm">
      <div class="two-col">
        <label>
          <span>Статус оплаты</span>
          <select name="payment_status">
            <option value="paid">paid</option>
            <option value="partially_paid">partially_paid</option>
            <option value="unpaid">unpaid</option>
          </select>
        </label>
        <label>
          <span>Сумма</span>
          <input name="amount" type="number" min="0" step="1" value="${escapeHtml(visit.amount || 0)}">
        </label>
      </div>
      <button class="primary-button" type="submit">Провести оплату</button>
    </form>
    <form class="form-stack" id="contextSubscriptionVisitForm">
      <div class="two-col">
        <label>
          <span>ID абонемента</span>
          <input name="subscription_id" inputmode="numeric" placeholder="subscription_id">
        </label>
        <label>
          <span>Дата</span>
          <input name="visit_date" type="date" value="${new Date().toISOString().slice(0, 10)}">
        </label>
      </div>
      <button class="secondary-button" type="submit">Списать по абонементу</button>
    </form>
    <div class="action-grid">
      <button class="secondary-button" id="contextOpenClient" type="button">Карточка клиента</button>
      <button class="ghost-button" id="contextRefresh" type="button">Обновить</button>
    </div>
  `;

  qs("#contextPaymentForm").payment_status.value = visit.payment_status || "paid";
  qs("#contextPaymentForm").addEventListener("submit", updatePaymentFromContext);
  qs("#contextSubscriptionVisitForm").addEventListener("submit", addSubscriptionVisitFromContext);
  qs("#contextOpenClient").addEventListener("click", () => openClientFromVisit(visit));
  qs("#contextRefresh").addEventListener("click", loadVisits);
}

async function updatePaymentFromContext(event) {
  event.preventDefault();
  const visit = state.selectedVisit;
  if (!visit) return;
  const form = event.currentTarget;
  const body = {
    client_id: Number(visit.client_id),
    payment_status: form.payment_status.value,
    amount: Number(form.amount.value || 0),
  };

  try {
    await api(`/payments/visits/${encodeURIComponent(visit.id)}`, {
      method: "PUT",
      body: JSON.stringify(body),
    });
    toast("Оплата обновлена");
    await loadVisits();
    await loadStatistics();
  } catch (error) {
    toast(`Оплата не проведена: ${error.message}`, "error");
  }
}

async function addSubscriptionVisitFromContext(event) {
  event.preventDefault();
  const visit = state.selectedVisit;
  if (!visit) return;
  const form = event.currentTarget;
  const body = {
    subscription_id: Number(form.subscription_id.value),
    appointment_id: Number(visit.id),
    visit_date: dateToISO(form.visit_date.value),
  };

  try {
    await api("/payments/subscription", {
      method: "POST",
      body: JSON.stringify(body),
    });
    toast("Занятие списано по абонементу");
    await loadVisits();
  } catch (error) {
    toast(`Списание не выполнено: ${error.message}`, "error");
  }
}

async function loadServices() {
  try {
    const data = await api("/services/?format=list");
    state.services = Array.isArray(data) ? data : data.items || [];
    renderServices();
    renderServiceOptions();
  } catch (error) {
    toast(`Услуги: ${error.message}`, "error");
    qs("#servicesGrid").innerHTML = `<div class="empty-state">Каталог недоступен</div>`;
  }
}

function renderServiceOptions() {
  const select = qs("#appointmentServiceSelect");
  if (!select) return;
  const current = select.value;
  select.innerHTML = `<option value="">Выберите услугу</option>` + state.services.map((service) => {
    const id = service.service_id || service.id;
    return `<option value="${escapeHtml(id)}">${escapeHtml(service.name)} · ${escapeHtml(formatMoney(service.price))}</option>`;
  }).join("");
  if (current) select.value = current;
  renderPopularServices();
}

function renderPopularServices() {
  const target = qs("#popularServices");
  if (!target) return;
  const popular = state.services.slice(0, 6);
  if (!popular.length) {
    target.innerHTML = "";
    return;
  }
  target.innerHTML = `
    <span>Популярные услуги</span>
    <div>
      ${popular.map((service) => {
        const id = service.service_id || service.id;
        return `<button type="button" data-service-pick="${escapeHtml(id)}">${escapeHtml(service.name)}</button>`;
      }).join("")}
    </div>
  `;
  qsa("[data-service-pick]").forEach((button) => {
    button.addEventListener("click", () => {
      qs("#appointmentServiceSelect").value = button.dataset.servicePick;
    });
  });
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
        <div class="kv"><span>ID</span><strong>${escapeHtml(service.service_id || service.id)}</strong></div>
        <div class="kv"><span>Цена</span><strong>${escapeHtml(formatMoney(service.price))}</strong></div>
      </div>
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

async function createClientFromForm(event) {
  event.preventDefault();
  const form = event.currentTarget;
  await createClient({
    name: form.name.value.trim(),
    phone: form.phone.value.trim(),
  });
  form.reset();
}

async function createAppointmentClient(event) {
  event.preventDefault();
  const form = event.currentTarget;
  await createClient({
    name: form.name.value.trim(),
    phone: form.phone.value.trim(),
  });
  form.reset();
}

async function createClient(body) {
  try {
    const result = await api("/clients/", {
      method: "POST",
      body: JSON.stringify(body),
    });
    const client = {
      id: result.client_id,
      name: body.name,
      phone: result.phone || body.phone,
    };
    await selectClient(client);
    toast(`Клиент создан: #${result.client_id}`);
  } catch (error) {
    toast(`Клиент не создан: ${error.message}`, "error");
  }
}

async function findClientByPhone(phone) {
  const result = await api(`/clients/find?phone=${encodeURIComponent(phone)}`);
  const info = await api(`/clients/info?client_id=${encodeURIComponent(result.client_id)}`);
  return {
    id: result.client_id,
    phone: result.phone,
    ...info,
  };
}

async function searchClient(event) {
  event.preventDefault();
  const phone = qs("#clientPhone").value.trim();
  if (!phone) {
    toast("Введите телефон", "error");
    return;
  }

  try {
    await selectClient(await findClientByPhone(phone));
  } catch (error) {
    toast(`Клиент не найден: ${error.message}`, "error");
  }
}

async function searchAppointmentClient(event) {
  event.preventDefault();
  const phone = qs("#appointmentClientPhone").value.trim();
  if (!phone) {
    toast("Введите телефон", "error");
    return;
  }

  try {
    await selectClient(await findClientByPhone(phone));
  } catch (error) {
    toast(`Клиент не найден: ${error.message}`, "error");
  }
}

async function selectClient(client) {
  state.selectedClient = client;
  qs("#appointmentForm").client_id.value = client.id || "";
  qs("#selectedClientBox").innerHTML = `
    <strong>${escapeHtml(client.name || `Клиент #${client.id}`)}</strong><br>
    <span>${escapeHtml(client.phone || "")}</span>
  `;
  renderClient(client);
  renderClientSnapshot(client);
  fillSalesClient();
  await loadClientMemberships();
}

function renderClientSnapshot(client) {
  const target = qs("#clientSnapshot");
  if (!target) return;
  if (!client) {
    target.innerHTML = `<div class="empty-state">История появится после выбора клиента</div>`;
    return;
  }
  target.innerHTML = `
    <div class="kv-grid">
      <div class="kv"><span>Визиты</span><strong>${escapeHtml(client.visit_count || 0)}</strong></div>
      <div class="kv"><span>Потрачено</span><strong>${escapeHtml(formatMoney(client.spent))}</strong></div>
      <div class="kv"><span>Последний визит</span><strong>${escapeHtml(formatDate(client.last_visit))}</strong></div>
      <div class="kv"><span>Скидка</span><strong>${escapeHtml(client.discount || 0)}%</strong></div>
    </div>
  `;
}

function renderClient(client) {
  const target = qs("#clientCard");
  qs("#clientCardCaption").textContent = `ID ${client.id || "-"}`;
  target.className = "client-card";
  target.innerHTML = `
    <div class="client-title">
      <strong>${escapeHtml(client.name || `Клиент #${client.id || ""}`)}</strong>
      <span>${escapeHtml(client.phone || "")}</span>
    </div>
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

async function openClientFromVisit(visit) {
  if (!visit.client_id) return;
  try {
    const info = await api(`/clients/info?client_id=${encodeURIComponent(visit.client_id)}`);
    await selectClient({ id: visit.client_id, ...info });
    setView("clients");
  } catch (error) {
    toast(`Карточка клиента недоступна: ${error.message}`, "error");
  }
}

async function loadMembershipTypes() {
  try {
    const data = await api("/subscriptions/types?format=list");
    state.membershipTypes = Array.isArray(data) ? data : data.items || [];
    renderMembershipOptions();
  } catch (error) {
    toast(`Типы абонементов: ${error.message}`, "error");
  }
}

function renderMembershipOptions() {
  const select = qs("#sellMembershipTypeSelect");
  if (!select) return;
  const current = select.value;
  select.innerHTML = `<option value="">Выберите тип</option>` + state.membershipTypes.map((item) => {
    const id = item.subscription_types_id || item.id;
    return `<option value="${escapeHtml(id)}">${escapeHtml(item.name)} · ${escapeHtml(item.sessions_count)} занятий</option>`;
  }).join("");
  if (current) select.value = current;
}

function fillMembershipSaleFields() {
  const select = qs("#sellMembershipTypeSelect");
  const form = qs("#sellMembershipForm");
  const item = state.membershipTypes.find((membership) => {
    return String(membership.subscription_types_id || membership.id) === select.value;
  });
  if (!item) return;
  form.cost.value = item.cost || "";
  form.sessions_count.value = item.sessions_count || "";
}

function fillSalesClient() {
  const form = qs("#sellMembershipForm");
  const client = state.selectedClient;
  if (!form || !client) return;
  form.client_id.value = client.id || "";
  form.client_label.value = `${client.name || `Клиент #${client.id}`} ${client.phone ? `· ${client.phone}` : ""}`;
  qs("#saleClientCaption").textContent = `Выбран ${client.name || `клиент #${client.id}`}`;
}

async function loadClientMemberships() {
  const target = qs("#clientMemberships");
  if (!target) return;
  const client = state.selectedClient;
  if (!client || !client.id) {
    target.innerHTML = `<div class="empty-state">Клиент не выбран</div>`;
    return;
  }

  target.innerHTML = `<div class="empty-state">Загрузка абонементов</div>`;
  try {
    const data = await api(`/subscriptions/client?client_id=${encodeURIComponent(client.id)}`);
    const items = data.subscriptions || [];
    if (!items.length) {
      target.innerHTML = `<div class="empty-state">Активных абонементов нет</div>`;
      return;
    }
    target.innerHTML = items.map((item) => `
      <article class="list-row">
        <div>
          <strong>Абонемент #${escapeHtml(item.subscription_id)}</strong>
          <span>Остаток занятий</span>
        </div>
        <strong>${escapeHtml(item.current_balance)}</strong>
      </article>
    `).join("");
  } catch (error) {
    target.innerHTML = `<div class="empty-state">Не удалось загрузить абонементы</div>`;
    toast(`Абонементы клиента: ${error.message}`, "error");
  }
}

async function sellMembership(event) {
  event.preventDefault();
  const form = event.currentTarget;
  const body = {
    client_id: Number(form.client_id.value),
    subscription_types_id: Number(form.subscription_types_id.value),
    cost: Number(form.cost.value),
    sessions_count: Number(form.sessions_count.value),
  };

  if (!body.client_id) {
    toast("Сначала выберите клиента", "error");
    return;
  }

  try {
    await api("/subscriptions/sell", {
      method: "POST",
      body: JSON.stringify(body),
    });
    toast("Абонемент продан");
    await loadClientMemberships();
    await loadStatistics();
  } catch (error) {
    toast(`Абонемент не продан: ${error.message}`, "error");
  }
}

async function createAppointment(event) {
  event.preventDefault();
  const form = event.currentTarget;
  const body = {
    client_id: Number(form.client_id.value),
    service_id: Number(form.service_id.value),
    start_time: localDateTimeToISO(form.start_time.value),
  };

  if (!body.client_id) {
    toast("Сначала выберите клиента", "error");
    return;
  }

  try {
    const created = await api("/visits/", {
      method: "POST",
      body: JSON.stringify(body),
    });
    qs("#appointmentDialog").close();
    form.reset();
    toast(`Запись создана: #${created.id || ""}`);
    await loadVisits();
    setView("schedule");
  } catch (error) {
    toast(`Не удалось создать запись: ${error.message}`, "error");
  }
}

function openAppointmentDialog(hour) {
  loadServices();
  const form = qs("#appointmentForm");
  const selectedDate = qs("#scheduleDate").value || toDateInputValue(new Date());
  const date = new Date(`${selectedDate}T00:00:00`);
  const targetHour = Number.isFinite(hour) ? hour : new Date().getHours() + 1;
  date.setHours(targetHour, 0, 0, 0);
  form.start_time.value = toDateTimeInputValue(date);
  if (state.selectedClient) {
    form.client_id.value = state.selectedClient.id || "";
    renderClientSnapshot(state.selectedClient);
  } else {
    qs("#selectedClientBox").textContent = "Клиент не выбран";
    renderClientSnapshot(null);
  }
  qs("#appointmentDialog").showModal();
}

function refreshCurrentView() {
  if (state.view === "schedule") {
    loadVisits();
    loadServices();
  }
  if (state.view === "clients") {
    if (state.selectedClient) renderClient(state.selectedClient);
  }
  if (state.view === "sales") {
    loadMembershipTypes();
    fillSalesClient();
    loadClientMemberships();
  }
  if (state.view === "services") loadServices();
  if (state.view === "reports") loadStatistics();
}

function bindEvents() {
  qsa(".nav-item").forEach((item) => {
    item.addEventListener("click", () => setView(item.dataset.view));
  });

  qsa("[data-range]").forEach((item) => {
    item.addEventListener("click", () => {
      state.range = item.dataset.range;
      qsa("[data-range]").forEach((button) => button.classList.toggle("active", button === item));
      loadVisits();
    });
  });

  qs("#refreshButton").addEventListener("click", refreshCurrentView);
  qs("#openAppointmentButton").addEventListener("click", openAppointmentDialog);
  qs("#openClientButton").addEventListener("click", () => setView("clients"));
  qs("#clientBookButton").addEventListener("click", openAppointmentDialog);
  qs("#scheduleDate").addEventListener("change", loadVisits);
  qs("#visitStatus").addEventListener("change", loadVisits);
  qs("#clientSearchForm").addEventListener("submit", searchClient);
  qs("#clientForm").addEventListener("submit", createClientFromForm);
  qs("#appointmentClientSearchForm").addEventListener("submit", searchAppointmentClient);
  qs("#appointmentClientCreateForm").addEventListener("submit", createAppointmentClient);
  qs("#appointmentForm").addEventListener("submit", createAppointment);
  qs("#serviceForm").addEventListener("submit", createService);
  qs("#sellMembershipForm").addEventListener("submit", sellMembership);
  qs("#sellMembershipTypeSelect").addEventListener("change", fillMembershipSaleFields);
}

function initDates() {
  const now = new Date();
  qs("#scheduleDate").value = toDateInputValue(now);
}

function initTimezoneLabels() {
  const label = `Время: ${currentTimezoneLabel()}`;
  qs("#timezoneLabel").textContent = label;
  qs("#appointmentTimezoneLabel").textContent = label;
}

function init() {
  bindEvents();
  initDates();
  initTimezoneLabels();
  checkVersion();
  loadServices();
  loadMembershipTypes();
  const initial = window.location.hash.replace("#", "") || "schedule";
  setView(titles[initial] ? initial : "schedule");
}

document.addEventListener("DOMContentLoaded", init);
