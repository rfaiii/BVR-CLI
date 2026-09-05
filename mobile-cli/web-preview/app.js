document.addEventListener('DOMContentLoaded', () => {
    // Model Accordion
    const modelToggle = document.getElementById('model-toggle');
    const modelList = document.getElementById('model-list');
    
    modelToggle.addEventListener('click', () => {
        modelList.classList.toggle('open');
        const icon = modelToggle.querySelector('i');
        icon.classList.toggle('ph-caret-down');
        icon.classList.toggle('ph-caret-up');
    });

    // Mock Commands
    const commands = [
        { cmd: 'themes', desc: 'Change application theme', icon: 'ph-paint-brush' },
        { cmd: 'ollama', desc: 'Manage local Ollama models', icon: 'ph-cpu' },
        { cmd: 'node', desc: 'Manage desktop NODE connections', icon: 'ph-network' },
        { cmd: 'finder', desc: 'Open file browser', icon: 'ph-folder' },
        { cmd: 'cd', desc: 'Change working directory', icon: 'ph-arrow-u-up-left' },
        { cmd: 'mcp', desc: 'Manage MCP servers', icon: 'ph-server' }
    ];

    // Command Palette Logic
    const chatInput = document.getElementById('chat-input');
    const commandPalette = document.getElementById('command-palette');
    const paletteList = document.getElementById('palette-list');
    const chatView = document.getElementById('chat-view');

    let paletteState = 'root';

    const themes = [
        { cmd: 'material', desc: 'Default dark aesthetics', icon: 'ph-moon' },
        { cmd: 'charmtone', desc: 'Charm CLI pink/purple', icon: 'ph-terminal' },
        { cmd: 'light', desc: 'Blinding light mode', icon: 'ph-sun' }
    ];

    function renderCommands(filterText) {
        paletteList.innerHTML = '';
        
        let currentList = paletteState === 'root' ? commands : themes;
        let filtered = currentList;

        if (filterText && filterText !== '/') {
            const search = filterText.substring(1).toLowerCase();
            filtered = currentList.filter(c => 
                c.cmd.toLowerCase().includes(search) || 
                c.desc.toLowerCase().includes(search)
            );
        }

        if (filtered.length === 0) {
            commandPalette.classList.add('hidden');
            return;
        }

        filtered.forEach(c => {
            const div = document.createElement('div');
            div.className = 'cmd-item';
            div.innerHTML = `
                <i class="ph-fill ${c.icon} cmd-icon"></i>
                <div class="cmd-info">
                    <span class="cmd-name">${paletteState === 'root' ? '/' : ''}${c.cmd}</span>
                    <span class="cmd-desc">${c.desc}</span>
                </div>
            `;
            div.addEventListener('click', () => {
                executeCommand(c.cmd);
            });
            paletteList.appendChild(div);
        });

        commandPalette.classList.remove('hidden');
    }

    chatInput.addEventListener('input', (e) => {
        const text = e.target.value;
        if (!text) {
            paletteState = 'root';
        }
        if (text.startsWith('/')) {
            renderCommands(text);
        } else {
            commandPalette.classList.add('hidden');
        }
    });

    // Chat Execution
    function appendMessage(text, isUser = true) {
        const msgDiv = document.createElement('div');
        msgDiv.className = `message ${isUser ? 'user' : 'system'}`;
        msgDiv.innerHTML = `<div class="bubble">${text}</div>`;
        chatView.appendChild(msgDiv);
        chatView.scrollTop = chatView.scrollHeight;
    }

    function executeCommand(cmdName) {
        if (paletteState === 'root') {
            if (cmdName === 'themes') {
                paletteState = 'themes';
                chatInput.value = '/';
                renderCommands('/');
            } else {
                chatInput.value = '';
                commandPalette.classList.add('hidden');
                appendMessage(`Executed command: /${cmdName}\n\n(This is a mock response from the web prototype)`, false);
            }
        } else if (paletteState === 'themes') {
            paletteState = 'root';
            chatInput.value = '';
            commandPalette.classList.add('hidden');
            appendMessage(`Switched theme to: ${cmdName}`, false);
        }
    }

    chatInput.addEventListener('keydown', (e) => {
        if (e.key === 'Enter') {
            const text = chatInput.value.trim();
            if (!text) return;

            if (text.startsWith('/')) {
                const cmd = text.substring(1).trim();
                executeCommand(cmd);
            } else {
                appendMessage(text, true);
                chatInput.value = '';
                // Mock response
                setTimeout(() => {
                    appendMessage("I received your message on the web prototype!", false);
                }, 500);
            }
        }
    });
});
