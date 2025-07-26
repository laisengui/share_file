// 上传功能
const dropzone = document.getElementById('dropzone');
const fileInput = document.getElementById('file-input');
const selectFileBtn = document.getElementById('select-file-btn');
const fileInfo = document.getElementById('file-info');
const filename = document.getElementById('filename');
const filesize = document.getElementById('filesize');
const progressBar = document.getElementById('progress-bar');
const progress = document.getElementById('progress');
const submitBtn = document.getElementById('submit-btn');
const uploadForm = document.getElementById('upload-form');
const uploadMessage = document.getElementById('upload-message');
const fileResult = document.getElementById('file-result');
const fileCode = document.getElementById('file-code');

// 拖拽上传功能
['dragenter', 'dragover', 'dragleave', 'drop'].forEach(eventName => {
    dropzone.addEventListener(eventName, preventDefaults, false);
});
dropzone.addEventListener('click', () => {
    // 添加动画类
    var className = this.className;
    this.className = className + (className ? ' ' : '') + 'click-animation';
    this.className = this.className.replace(/\bclick-animation\b/g, '');
    selectFileBtn.click()
});

function preventDefaults(e) {
    e.preventDefault();
    e.stopPropagation();
}

['dragenter', 'dragover'].forEach(eventName => {
    dropzone.addEventListener(eventName, highlight, false);
});

['dragleave', 'drop'].forEach(eventName => {
    dropzone.addEventListener(eventName, unhighlight, false);
});

function highlight() {
    dropzone.classList.add('active');
}

function unhighlight() {
    dropzone.classList.remove('active');
}

dropzone.addEventListener('drop', handleDrop, false);

function handleDrop(e) {
    const dt = e.dataTransfer;
    const files = dt.files;

    if (files.length > 0) {
        handleFiles(files);
    }
}

selectFileBtn.addEventListener('click', () => {
    fileInput.click();
});

fileInput.addEventListener('change', () => {
    if (fileInput.files.length > 0) {
        handleFiles(fileInput.files);
    }
});

function handleFiles(files) {
    const file = files[0];

    // 显示文件信息
    filename.textContent = file.name;
    filesize.textContent = formatFileSize(file.size);
    fileInfo.style.display = 'block';

    // 启用提交按钮
    submitBtn.disabled = false;

    // 存储文件引用
    uploadForm.file = file;
}

function formatFileSize(bytes) {
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
    if (bytes < 1024 * 1024 * 1024) return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
    return (bytes / (1024 * 1024 * 1024)).toFixed(1) + ' GB';
}
let formUpload=false
uploadForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    if(formUpload){
        return
    }
    const file = uploadForm.file;
    const expiry = document.getElementById('expiry').value;
    const times = document.getElementById('times').value;

    if (!file) {
        showMessage('请先选择文件', 'error');
        return;
    }

    const formData = new FormData();
    formData.append('file', file);
    formData.append('expiry', expiry);
    formData.append('times', times);

    // 显示进度条
    progressBar.style.display = 'block';

    try {
        const xhr = new XMLHttpRequest();

        xhr.upload.addEventListener('progress', (e) => {
            if (e.lengthComputable) {
                const percent = Math.round((e.loaded / e.total) * 100);
                progress.style.width = percent + '%';
            }
        });

        xhr.open('POST', '/upload');

        xhr.onload = () => {
            if (xhr.status === 200) {
                const response = JSON.parse(xhr.responseText);
                //showMessage(uploadMessage, t('uploadModal.success', { code: response.code, expiry: expiry }), 'success');
                showCode(response.code)
            } else {
                showMessage(xhr.responseText|| '上传失败', 'error');
                formUpload=false
            }
        };

        xhr.onerror = () => {
            showMessage('网络错误，请重试', 'error');
            formUpload=false
        };
        formUpload=true
        xhr.send(formData);

    } catch (error) {
        showMessage(t('errors.uploadFailed') + ': ' + error.message, 'error');
        progressBar.style.display = 'none';
    }
});

function showCode(code) {
    fileResult.style.display = 'block';
    fileCode.innerHTML = code
}

function resetUploadForm() {
    formUpload=false
    fileInput.value = '';
    fileInfo.style.display = 'none';
    progressBar.style.display = 'none';
    progress.style.width = '0%';
    submitBtn.disabled = true;
    uploadMessage.style.display = 'none';
    fileResult.style.display = 'none';
    delete uploadForm.file;
}