let modalAction = {
  url: null,
  method: 'GET',
  body: null,
  wsconnection: WebSocket,
};

function openModal({ url, method = 'GET', body = null, message = 'Are you sure?',wsconnection =WebSocket }) {
  modalAction = { url, method, body,wsconnection };
  document.getElementById('modal-message').textContent = message;
  document.getElementById('modal-backdrop').classList.remove('hidden');
}

function closeModal() {
  document.getElementById('modal-backdrop').classList.add('hidden');
 // modalAction = { url: null, method: 'GET', body: null };
}

function confirmYes() {
  closeModal();
  const { url, method, body,wsconnection } = modalAction;
  if (!url || !method) return;

const form = document.createElement('form');
form.method = method;
form.action = url;

if (method === 'POST' && body) {
    Object.entries(body).forEach(([key, value]) => {
        const input = document.createElement('input');
        input.type = 'hidden';
        input.name = key;
        input.value = value;
        form.appendChild(input);
    });
}
else if (method === 'GET' && body) {
    const searchParams = new URLSearchParams();
    Object.entries(body).forEach(([key, value]) => {
        
    form.action += `${value.toString()}/`;
    });
}
else if (method =='ws' && body && wsconnection){ // websocket connection
  wsconnection.send(body)
  return 
}

document.body.appendChild(form);
form.submit();
}

