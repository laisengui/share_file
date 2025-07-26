class I18n {
    constructor() {
        this.locale = 'zh'; // 默认中文
        this.getLang()
        this.translations = {};
        this.loadTranslations();
        this.initLanguageSelector();
    }

    /**
     * 加载资源包
     * @returns {Promise<void>}
     */
    async loadTranslations() {
        try {
            const response = await fetch(`static/locales/${this.locale}.json`);
            this.translations[this.locale] = await response.json();
            this.translatePage()
        } catch (error) {
            console.error('Failed to load translations:', error);
        }
    }

    /**
     * 渲染页面
     */
    translatePage() {
        const elements = document.querySelectorAll('[t]');

        elements.forEach(element => {
            const key = element.getAttribute('t');
            const keys = key.split('.');
            let translation = this.translations[this.locale];

            try {
                for (const k of keys) {
                    translation = translation[k];
                }
                if (translation) {
                    element.textContent = translation;
                }
            } catch (e) {
                console.warn(`Translation not found for key: ${key}`);
            }
        });
    }

    /**
     * 初始化按钮
     */
    initLanguageSelector() {
        const selector = document.createElement('div');
        selector.className = 'language-selector';
        selector.innerHTML = `
      <button class="lang-btn ${this.locale === 'zh' ? 'active' : ''}" data-lang="zh">中文</button>
      <button class="lang-btn ${this.locale === 'en' ? 'active' : ''}" data-lang="en">English</button>
    `;

        document.querySelector('header').appendChild(selector);

        selector.querySelectorAll('.lang-btn').forEach(btn => {
            btn.addEventListener('click', () => {
                this.setLang(btn.dataset.lang)
                selector.querySelectorAll('.lang-btn').forEach(b => b.classList.remove('active'));
                btn.classList.add('active');
            });
        });
    }

    /**
     * 设置语言
     * @param lang
     */
    setLang(lang) {
        const name = "lang"
        const days = 365
        let expires = "";
        if (days) {
            const date = new Date();
            date.setTime(date.getTime() + (days * 24 * 60 * 60 * 1000));
            expires = "; expires=" + date.toUTCString();
        }
        document.cookie = name + "=" + (lang || "") + expires + "; path=/";

        this.locale=lang
        this.loadTranslations()
    }

    /**
     * 拿到语言
     * @returns {string}
     */
    getLang() {
        const nameEQ = "lang=";
        const ca = document.cookie.split(';');
        for (let i = 0; i < ca.length; i++) {
            let c = ca[i];
            while (c.charAt(0) === ' ') c = c.substring(1, c.length);
            if (c.indexOf(nameEQ) === 0) return c.substring(nameEQ.length, c.length);
        }
        return this.locale;
    }
}

// 初始化国际化
const i18n = new I18n();

// 暴露翻译函数到全局
function t(key, params) {
    return i18n.trans(key, params);
}
