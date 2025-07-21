// 模态框控制
const uploadCard = document.getElementById('upload-card');
const downloadCard = document.getElementById('download-card');
const uploadModal = document.getElementById('upload-modal');
const downloadModal = document.getElementById('download-modal');
const closeUpload = document.getElementById('close-upload');
const closeDownload = document.getElementById('close-download');

uploadCard.addEventListener('click', () => {
    uploadModal.style.display = 'flex';
});

downloadCard.addEventListener('click', () => {
    downloadModal.style.display = 'flex';
});

closeUpload.addEventListener('click', () => {
    uploadModal.style.display = 'none';
    resetUploadForm();
});

closeDownload.addEventListener('click', () => {
    downloadModal.style.display = 'none';
    resetDownloadForm();
});

// 点击模态框外部关闭
window.addEventListener('click', (e) => {
    if (e.target === uploadModal) {
        uploadModal.style.display = 'none';
        resetUploadForm();
    }
    if (e.target === downloadModal) {
        downloadModal.style.display = 'none';
        resetDownloadForm();
    }
});