// 下载功能
const downloadSubmit = document.getElementById('download-submit');
const downloadCode = document.getElementById('download-code');
const downloadMessage = document.getElementById('download-message');

downloadSubmit.addEventListener('click', () => {
    const code = downloadCode.value.trim();

    if ( !/^[0-9a-z]+$/.test(code)) {
        showMessage('请输入下载码', 'error');
        return;
    }

    // // 尝试下载
    // window.location.href = `download/${code}`;
    //
    // // 检查下载是否成功
    // setTimeout(() => {
    //     if (!downloadMessage.textContent.includes('成功')) {
    //         showMessage( '下载失败: 文件不存在或已过期', 'error');
    //     }
    // }, 2000);

    fetch(`download/${code}`, {
        method: 'GET',
        headers: {
            'Content-Type': 'application/octet-stream',
        }
    }).then(response => {
        // 从响应头获取文件名
        const contentDisposition = response.headers.get('content-disposition');
        let filename = 'default.ext';

        if (contentDisposition) {
            // 尝试获取 RFC 5987 编码的文件名 (filename*=UTF-8'')
            const utf8FilenameMatch = contentDisposition.match(/filename\*=(?:UTF-8'')?([^;]+)/i);

            if (utf8FilenameMatch && utf8FilenameMatch[1]) {
                // 解码 RFC 5987 编码的文件名
                filename = decodeURIComponent(utf8FilenameMatch[1]);
            } else {
                // 回退到普通文件名
                const filenameMatch = contentDisposition.match(/filename=["']?([^"';]+)["']?/i);
                if (filenameMatch && filenameMatch[1]) {
                    filename = filenameMatch[1];
                }
            }
        }

        return response.blob().then(blob => ({ blob, filename }));
    })
        .then(({ blob, filename }) => {
            // 创建下载链接
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = filename;
            document.body.appendChild(a);
            a.click();

            // 清理
            window.URL.revokeObjectURL(url);
            document.body.removeChild(a);
        })
        .catch(error =>  showMessage(xhr.responseText|| '下载失败', 'error'));
});

function resetDownloadForm() {
    downloadCode.value = '';
    downloadMessage.style.display = 'none';
}

// function showMessage(element, message, type) {
//     element.innerHTML = message;
//     element.className = `message ${type}`;
//     element.style.display = 'block';
// }