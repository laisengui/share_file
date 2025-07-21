// 下载功能
const downloadSubmit = document.getElementById('download-submit');
const downloadCode = document.getElementById('download-code');
const downloadMessage = document.getElementById('download-message');

downloadSubmit.addEventListener('click', () => {
    const code = downloadCode.value.trim();

    if ( !/^[0-9a-z]+$/.test(code)) {
        showMessage(downloadMessage, '请输入下载码', 'error');
        return;
    }

    // 尝试下载
    window.location.href = `/download/${code}`;

    // 检查下载是否成功
    setTimeout(() => {
        if (!downloadMessage.textContent.includes('成功')) {
            showMessage(downloadMessage, '下载失败: 文件不存在或已过期', 'error');
        }
    }, 2000);
});

function resetDownloadForm() {
    downloadCode.value = '';
    downloadMessage.style.display = 'none';
}

function showMessage(element, message, type) {
    element.innerHTML = message;
    element.className = `message ${type}`;
    element.style.display = 'block';
}