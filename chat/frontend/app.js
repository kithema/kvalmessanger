const API_URL = 'http://localhost:8080/api';

let currentUser = null;
let currentChatId = null;
let currentChatData = null;
let token = null;
let messagePollInterval = null;

// DOM elements
const authPage = document.getElementById('auth-page');
const chatPage = document.getElementById('chat-page');
const loginForm = document.getElementById('login-form');
const registerForm = document.getElementById('register-form');
const chatList = document.getElementById('chat-list');
const messagesContainer = document.getElementById('messages-container');
const messageInput = document.getElementById('message-input');
const sendBtn = document.getElementById('send-btn');
const currentChatName = document.getElementById('current-chat-name');
const chatParticipantsInfo = document.getElementById('chat-participants-info');
const chatAvatarText = document.getElementById('chat-avatar-text');
const userAlias = document.getElementById('user-alias');
const userNameDisplay = document.getElementById('user-name');
const logoutBtn = document.getElementById('logout-btn');
const newChatBtn = document.getElementById('new-chat-btn');
const newChatModal = document.getElementById('new-chat-modal');
const newChatForm = document.getElementById('new-chat-form');
const profileBtn = document.getElementById('profile-btn');
const profileModal = document.getElementById('profile-modal');
const profileForm = document.getElementById('profile-form');
const profileAlias = document.getElementById('profile-alias');
const profileName = document.getElementById('profile-name');
const profilePassword = document.getElementById('profile-password');
const profileError = document.getElementById('profile-error');
const profileSuccess = document.getElementById('profile-success');
const chatInfoBtn = document.getElementById('chat-info-btn');
const chatInfoModal = document.getElementById('chat-info-modal');
const chatInfoTitle = document.getElementById('chat-info-title');
const infoChatName = document.getElementById('info-chat-name');
const infoChatType = document.getElementById('info-chat-type');
const infoParticipantsList = document.getElementById('info-participants-list');

// Tab switching
document.querySelectorAll('.tab-btn').forEach(btn => {
    btn.addEventListener('click', function() {
        document.querySelectorAll('.tab-btn').forEach(b => b.classList.remove('active'));
        this.classList.add('active');
        const tab = this.dataset.tab;
        document.querySelectorAll('.auth-form').forEach(f => f.classList.remove('active'));
        document.getElementById(`${tab}-form`).classList.add('active');
    });
});

// Login
loginForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    const alias = document.getElementById('login-alias').value;
    const password = document.getElementById('login-password').value;
    const errorEl = document.getElementById('login-error');

    try {
        const response = await fetch(`${API_URL}/login`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ alias, password })
        });

        const data = await response.json();
        if (!response.ok) {
            errorEl.textContent = data.error || 'Ошибка входа';
            return;
        }

        token = data.token;
        currentUser = data.user;
        localStorage.setItem('token', token);
        localStorage.setItem('user', JSON.stringify(currentUser));
        errorEl.textContent = '';
        showChatPage();
    } catch (err) {
        errorEl.textContent = 'Ошибка соединения с сервером';
    }
});

// Register
registerForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    const alias = document.getElementById('reg-alias').value;
    const name = document.getElementById('reg-name').value || alias;
    const password = document.getElementById('reg-password').value;
    const errorEl = document.getElementById('reg-error');

    try {
        const response = await fetch(`${API_URL}/register`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ alias, name, password })
        });

        const data = await response.json();
        if (!response.ok) {
            errorEl.textContent = data.error || 'Ошибка регистрации';
            return;
        }

        errorEl.textContent = 'Регистрация успешна! Теперь войдите.';
        errorEl.style.color = '#66bb6a';
        document.querySelector('[data-tab="login"]').click();
        document.getElementById('login-alias').value = alias;
        document.getElementById('login-password').value = password;
    } catch (err) {
        errorEl.textContent = 'Ошибка соединения с сервером';
    }
});

// Logout
logoutBtn.addEventListener('click', () => {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    token = null;
    currentUser = null;
    currentChatId = null;
    currentChatData = null;
    if (messagePollInterval) {
        clearInterval(messagePollInterval);
        messagePollInterval = null;
    }
    showAuthPage();
});

// Profile
profileBtn.addEventListener('click', () => {
    if (currentUser) {
        profileAlias.value = currentUser.alias;
        profileName.value = currentUser.name || '';
        profilePassword.value = '';
        profileError.textContent = '';
        profileSuccess.textContent = '';
        profileModal.classList.add('active');
    }
});

document.getElementById('profile-modal-close').addEventListener('click', () => {
    profileModal.classList.remove('active');
});

profileModal.addEventListener('click', (e) => {
    if (e.target === profileModal) {
        profileModal.classList.remove('active');
    }
});

profileForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    const name = profileName.value.trim();
    const password = profilePassword.value;

    try {
        const response = await fetch(`${API_URL}/profile`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            },
            body: JSON.stringify({ 
                name: name || undefined,
                password: password || undefined
            })
        });

        if (!response.ok) {
            const data = await response.json();
            profileError.textContent = data.error || 'Ошибка обновления профиля';
            profileSuccess.textContent = '';
            return;
        }

        if (name) {
            currentUser.name = name;
            userNameDisplay.textContent = name;
            localStorage.setItem('user', JSON.stringify(currentUser));
        }

        profileSuccess.textContent = 'Профиль успешно обновлен!';
        profileError.textContent = '';
        setTimeout(() => {
            profileModal.classList.remove('active');
        }, 2000);
    } catch (err) {
        profileError.textContent = 'Ошибка соединения с сервером';
        profileSuccess.textContent = '';
    }
});

// Chat Info Modal
chatInfoBtn.addEventListener('click', () => {
    if (currentChatId && currentChatData) {
        showChatInfo(currentChatData);
    }
});

document.getElementById('chat-info-modal-close').addEventListener('click', () => {
    chatInfoModal.classList.remove('active');
});

chatInfoModal.addEventListener('click', (e) => {
    if (e.target === chatInfoModal) {
        chatInfoModal.classList.remove('active');
    }
});

async function showChatInfo(chatData) {
    chatInfoTitle.textContent = `Информация о чате`;
    infoChatName.textContent = chatData.name || 'Без названия';
    infoChatType.textContent = chatData.is_group ? '👥 Групповой' : '👤 Личный';
    
    // Получаем участников
    try {
        const response = await fetch(`${API_URL}/chats/${chatData.id}/members`, {
            headers: { 'Authorization': `Bearer ${token}` }
        });
        if (response.ok) {
            const members = await response.json();
            infoParticipantsList.innerHTML = members.map(m => `
                <div class="participant-item">
                    <span class="participant-alias">@${m.alias}</span>
                    <span class="participant-name">${m.name || ''}</span>
                    ${m.user_id === currentUser.id ? '<span class="participant-owner">Вы</span>' : ''}
                </div>
            `).join('');
        } else {
            infoParticipantsList.innerHTML = '<span style="color:#888;">Не удалось загрузить участников</span>';
        }
    } catch (err) {
        infoParticipantsList.innerHTML = '<span style="color:#888;">Ошибка загрузки участников</span>';
    }
    
    chatInfoModal.classList.add('active');
}

// Show chat page
function showChatPage() {
    authPage.classList.remove('active');
    chatPage.classList.add('active');
    userAlias.textContent = `@${currentUser.alias}`;
    userNameDisplay.textContent = currentUser.name || currentUser.alias;
    loadChats();
    if (messagePollInterval) clearInterval(messagePollInterval);
    messagePollInterval = setInterval(() => {
        loadChats(true);
        if (currentChatId) loadMessages(currentChatId, true);
    }, 5000);
}

function showAuthPage() {
    chatPage.classList.remove('active');
    authPage.classList.add('active');
}

// Load chats
async function loadChats(silent = false) {
    try {
        const response = await fetch(`${API_URL}/chats`, {
            headers: { 'Authorization': `Bearer ${token}` }
        });
        if (!response.ok) throw new Error('Failed to load chats');
        const chats = await response.json();
        renderChats(chats);
    } catch (err) {
        if (!silent) console.error('Error loading chats:', err);
    }
}

function renderChats(chats) {
    chatList.innerHTML = '';
    chats.forEach(chat => {
        const div = document.createElement('div');
        div.className = `chat-item${chat.id === currentChatId ? ' active' : ''}`;
        
        const isGroup = chat.is_group;
        const avatar = isGroup ? '👥' : '👤';
        const name = chat.name || (isGroup ? 'Групповой чат' : 'Личный чат');
        
        div.innerHTML = `
            <div class="chat-item-left">
                <div class="chat-item-avatar">${avatar}</div>
                <div class="chat-item-info">
                    <div class="chat-item-name">${name}</div>
                    <div class="chat-item-last">${isGroup ? 'Групповой чат' : 'Личный чат'}</div>
                </div>
            </div>
            <div class="chat-item-right">
                <div class="chat-item-time">${new Date(chat.updated_at).toLocaleTimeString()}</div>
                ${chat.unread_count > 0 ? `<span class="unread-badge">${chat.unread_count}</span>` : ''}
            </div>
        `;
        
        div.addEventListener('click', () => {
            currentChatId = chat.id;
            currentChatData = chat;
            updateChatHeader(chat);
            loadMessages(chat.id);
            document.querySelectorAll('.chat-item').forEach(el => el.classList.remove('active'));
            div.classList.add('active');
        });
        chatList.appendChild(div);
    });
}

function updateChatHeader(chat) {
    const isGroup = chat.is_group;
    const name = chat.name || (isGroup ? 'Групповой чат' : 'Личный чат');
    currentChatName.textContent = name;
    chatAvatarText.textContent = isGroup ? '👥' : '👤';
    chatParticipantsInfo.textContent = isGroup ? 'Групповой чат' : 'Личный чат';
}

// Load messages
async function loadMessages(chatId, silent = false) {
    try {
        const response = await fetch(`${API_URL}/messages?chat_id=${chatId}`, {
            headers: { 'Authorization': `Bearer ${token}` }
        });
        if (!response.ok) throw new Error('Failed to load messages');
        const messages = await response.json();
        renderMessages(messages);
    } catch (err) {
        if (!silent) console.error('Error loading messages:', err);
    }
}

function renderMessages(messages) {
    messagesContainer.innerHTML = '';
    if (messages.length === 0) {
        messagesContainer.innerHTML = `
            <div class="empty-chat-message">
                <span>💬</span>
                <p>Нет сообщений. Начните общение!</p>
            </div>
        `;
        return;
    }
    
    messages.reverse().forEach(msg => {
        const div = document.createElement('div');
        const isMine = msg.sender_id === currentUser.id;
        div.className = `message ${isMine ? 'message-mine' : 'message-other'}`;
        
        let content = '';
        if (msg.is_deleted) {
            content = '<div class="message-text message-deleted">Сообщение удалено</div>';
        } else {
            if (!isMine) {
                const displayName = msg.sender_name || msg.sender_alias || 'Пользователь';
                content += `<div class="message-username">${displayName}</div>`;
            }
            content += `<div class="message-text">${msg.text || ''}</div>`;
            content += `<div class="message-time">${new Date(msg.sent_at).toLocaleTimeString()}</div>`;
        }
        div.innerHTML = content;
        messagesContainer.appendChild(div);
    });
    messagesContainer.scrollTop = messagesContainer.scrollHeight;
}

// Send message
sendBtn.addEventListener('click', sendMessage);
messageInput.addEventListener('keydown', (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault();
        sendMessage();
    }
});

async function sendMessage() {
    if (!currentChatId) {
        alert('Выберите чат');
        return;
    }
    const text = messageInput.value.trim();
    if (!text) return;

    try {
        const response = await fetch(`${API_URL}/messages`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            },
            body: JSON.stringify({ chat_id: currentChatId, text })
        });
        if (!response.ok) throw new Error('Failed to send message');
        messageInput.value = '';
        loadMessages(currentChatId);
        loadChats();
    } catch (err) {
        console.error('Error sending message:', err);
        alert('Не удалось отправить сообщение');
    }
}

// New chat modal
newChatBtn.addEventListener('click', () => {
    newChatModal.classList.add('active');
});

document.querySelector('.modal-close').addEventListener('click', () => {
    newChatModal.classList.remove('active');
});

newChatModal.addEventListener('click', (e) => {
    if (e.target === newChatModal) {
        newChatModal.classList.remove('active');
    }
});

newChatForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    const name = document.getElementById('chat-name').value.trim();
    const isGroup = document.getElementById('chat-is-group').checked;
    const participantsInput = document.getElementById('chat-participants').value.trim();
    
    let userAliases = [];
    if (participantsInput) {
        userAliases = participantsInput.split(',').map(a => a.trim()).filter(a => a.length > 0);
    }
    
    if (!isGroup && userAliases.length !== 1) {
        alert('Для личного чата укажите 1 участника (alias)');
        return;
    }
    if (isGroup && userAliases.length < 1) {
        alert('Для группового чата укажите хотя бы 1 участника (alias)');
        return;
    }

    try {
        const response = await fetch(`${API_URL}/chats`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            },
            body: JSON.stringify({ 
                name, 
                is_group: isGroup, 
                user_aliases: userAliases 
            })
        });

        const responseText = await response.text();
        let data;
        try {
            data = JSON.parse(responseText);
        } catch (e) {
            alert('Ошибка сервера: ' + responseText);
            return;
        }

        if (!response.ok) {
            alert(data.error || 'Не удалось создать чат');
            return;
        }

        newChatModal.classList.remove('active');
        newChatForm.reset();
        loadChats();
        if (data.chat_id) {
            currentChatId = data.chat_id;
            loadMessages(data.chat_id);
        }
    } catch (err) {
        console.error('Error creating chat:', err);
        alert('Ошибка соединения с сервером');
    }
});

document.getElementById('chat-is-group').addEventListener('change', function() {
    const container = document.getElementById('participants-container');
    const label = container.querySelector('label');
    const hint = container.querySelector('.hint');
    if (this.checked) {
        label.textContent = 'Участники (alias через запятую):';
        hint.textContent = 'Введите alias участников через запятую';
    } else {
        label.textContent = 'Собеседник (alias):';
        hint.textContent = 'Введите alias собеседника';
    }
});

// Initialization
document.addEventListener('DOMContentLoaded', () => {
    const savedToken = localStorage.getItem('token');
    const savedUser = localStorage.getItem('user');
    if (savedToken && savedUser) {
        token = savedToken;
        currentUser = JSON.parse(savedUser);
        showChatPage();
    }
});

document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') {
        newChatModal.classList.remove('active');
        profileModal.classList.remove('active');
        chatInfoModal.classList.remove('active');
    }
});