const state = {
  apps: null,
};

const renderToDOM = () => {
  const appContainer = document.querySelector("#app");
  appContainer.innerHTML = '';

  if (state.apps === null) {
    const loading = document.createElement("h3");
    loading.textContent = "Loading...";
    appContainer.append(loading);
    return;
  }

  appContainer.append(
    ...state.apps.map(status => {
      const div = document.createElement("div");
      div.classList.add("status");

      const p = document.createElement("p");

      p.append(status.app);

      const statusText = document.createElement("span");
      statusText.classList.add(status.isOk ? "green" : "red");
      statusText.textContent = status.isOk ? '(RUNNING)' : '(DOWN)';

      const statusIcon = document.createElement("span");
      statusIcon.textContent = status.isOk ? "🚀" : "❌";
      statusIcon.classList.add("status-icon");

      div.append(p);
      div.append(statusText);
      div.append(statusIcon);

      return div;
    })
  );
};

const isHttps = location.protocol === "https";
const wsUrl = `${isHttps ? 'wss' : 'ws'}://${location.host}/ws`;
const socket = new WebSocket(wsUrl);

socket.onopen = ev => {
  console.log("Connected to the pinger successfully");
  console.log("Initial connection event", ev);
};

socket.onmessage = ev => {
  const statusData = JSON.parse(ev.data);

  console.log(statusData);

  if (Array.isArray(statusData)) {
    state.apps = statusData;
    renderToDOM();
    return;
  } 

  const existsInApps = !!state.apps?.find(status => status.app === statusData.app);

  if (existsInApps) {
    state.apps = state.apps.map(a => a.app === statusData.app ? statusData : a);
  } else {
    state.apps.push(statusData);
  }

  renderToDOM();
};

document.addEventListener("beforeunload", () => {
  try {
    socket.close();
  } catch (err) {}
})
