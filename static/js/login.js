// 登录相关功能
document.addEventListener('DOMContentLoaded', function () {
    // 检查登录状态
    checkLoginStatus();

    // 登录按钮点击事件
    document.getElementById('login-btn').addEventListener('click', function () {
        document.getElementById('login-modal').style.display = 'flex';
    });

    // 关闭登录模态框
    document.getElementById('close-login').addEventListener('click', function () {
        document.getElementById('login-modal').style.display = 'none';
    });

    // 登录表单提交
    document.getElementById('login-form').addEventListener('submit', function (e) {
        e.preventDefault();
        const username = document.getElementById('username').value;
        const password = document.getElementById('password').value;
        const formData = new FormData();
        formData.append('username', username);
        formData.append('password', password);
        const xhr = new XMLHttpRequest();
        xhr.open('POST', '/login');
        xhr.onload = () => {
            const response = JSON.parse(xhr.responseText);
            if (xhr.status === 200) {
                // 登录成功
                updateLoginUI(true, username);
                document.getElementById('login-modal').style.display = 'none';
                document.getElementById('login-message').style.display = 'none';
            } else {
                // 登录失败
                const message = document.getElementById('login-message');
                showMessage(data.message || '登录失败', "error");
                message.className = 'message error';
                message.style.display = 'block';
            }
        };

        xhr.onerror = () => {
            showMessage(data.message || '登录请求失败', "error");
        };
        xhr.send(formData);
    });

    // 退出登录
    document.getElementById('logout-btn').addEventListener('click', function () {
        fetch('/logout', {
            method: 'POST'
        })
            .then(response => response.json())
            .then(data => {
                Location.reload()
            })
            .catch(error => {
                console.error('退出登录错误:', error);
            });
    });
});

// 检查登录状态
function checkLoginStatus() {
    fetch('/login/status')
        .then(response => response.json())
        .then(data => {
            if (data.loggedIn && data.username) {
                updateLoginUI(true, data.username);
            } else {
                updateLoginUI(false);
            }
        })
        .catch(error => {
            console.error('检查登录状态错误:', error);
            updateLoginUI(false);
        });
}

// 更新登录UI状态
function updateLoginUI(isLoggedIn, username = '') {
    if (isLoggedIn) {
        document.getElementById('login-container').style.display = 'flex';
        document.getElementById('logout-container').style.display = 'none';
        document.getElementById('username-display').textContent = username;
    } else {
        document.getElementById('login-container').style.display = 'none';
        document.getElementById('logout-container').style.display = 'flex';
    }
}