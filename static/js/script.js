let app;

const TOKEN_KEY = "freechat_token";
const USER_KEY = "freechat_user";

function getOS() {
    let info = navigator.userAgent.toLowerCase();
    let windowsMatch = /windows/;
    let macMatch = /mac/;
    let linuxMatch = /linux/;
    let iphoneMatch = /iphone/;
    let androidMatch = /android/;

    if (info.match(windowsMatch)) {
        return "Windows";
    } else if (info.match(iphoneMatch)) {
        return "IPhone";
    } else if (info.match(androidMatch)) {
        return "Android";
    } else if (info.match(macMatch)) {
        return "Mac OS";
    } else if (info.match(linuxMatch)) {
        return "Linux";
    } else {
        return "Unknown";
    }
}

function loadStoredUser() {
    try {
        const raw = localStorage.getItem(USER_KEY);
        return raw ? JSON.parse(raw) : null;
    } catch (err) {
        return null;
    }
}

function storeSession(token, username) {
    localStorage.setItem(TOKEN_KEY, token);
    localStorage.setItem(USER_KEY, JSON.stringify({ username: username }));
}

function clearSession() {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(USER_KEY);
}

// 所有请求自动带上 JWT
axios.interceptors.request.use((config) => {
    const token = localStorage.getItem(TOKEN_KEY);
    if (token) {
        config.headers.Authorization = "Bearer " + token;
    }
    return config;
});

// token 失效时清掉本地会话并弹出登录框（登录/注册接口自身的 401 交给表单处理）
axios.interceptors.response.use(
    (res) => res,
    (error) => {
        const status = error.response && error.response.status;
        const url = (error.config && error.config.url) || "";
        if (status === 401 && app && url.indexOf("/auth/") !== 0) {
            clearSession();
            app.user = null;
            if (!app.authModal.open) {
                app.openAuth("login");
                app.authModal.error = "登录状态已失效，请重新登录";
            }
        }
        return Promise.reject(error);
    }
);

function apiErrorMessage(error, fallback) {
    if (error.response && error.response.data && error.response.data.message) {
        return error.response.data.message;
    }
    return fallback;
}

window.onload = () => {
    app = new Vue({
        el: "#app",
        data: {
            comments: [],
            os: getOS(),
            user: loadStoredUser(),
            form: {
                content: ""
            },
            authForm: {
                username: "",
                password: ""
            },
            authModal: {
                open: false,
                mode: "login",
                loading: false,
                error: ""
            }
        },
        methods: {
            getData() {
                axios.get("/comments").then((res) => (app.comments = res.data));
            },
            send() {
                const content = app.form.content.trim();
                if (!app.user || !content || content.length > 50) {
                    return;
                }
                axios
                    .post("/comments", {
                        content: content,
                        os: app.os
                    })
                    .then((res) => {
                        app.comments = res.data;
                        app.form.content = "";
                    })
                    .catch((error) => {
                        alert(apiErrorMessage(error, "发送失败，请稍后重试"));
                    });
            },
            openAuth(mode) {
                app.authModal.mode = mode || "login";
                app.authModal.error = "";
                app.authForm.username = "";
                app.authForm.password = "";
                app.authModal.open = true;
            },
            closeAuth() {
                app.authModal.open = false;
                app.authModal.error = "";
                app.authForm.password = "";
            },
            switchMode(mode) {
                app.authModal.mode = mode;
                app.authModal.error = "";
            },
            submitAuth() {
                const username = app.authForm.username.trim();
                const password = app.authForm.password;
                const isLogin = app.authModal.mode === "login";

                if (!username || !password) {
                    app.authModal.error = "请填写用户名和密码";
                    return;
                }
                if (!isLogin && (username.length < 3 || username.length > 20)) {
                    app.authModal.error = "用户名长度需为 3-20 个字符";
                    return;
                }
                if (!isLogin && password.length < 6) {
                    app.authModal.error = "密码至少 6 个字符";
                    return;
                }

                app.authModal.loading = true;
                app.authModal.error = "";

                axios
                    .post(isLogin ? "/auth/login" : "/auth/register", {
                        username: username,
                        password: password
                    })
                    .then((res) => {
                        storeSession(res.data.token, res.data.username);
                        app.user = { username: res.data.username };
                        app.authModal.loading = false;
                        app.closeAuth();
                        app.getData();
                    })
                    .catch((error) => {
                        app.authModal.loading = false;
                        app.authModal.error = apiErrorMessage(
                            error,
                            isLogin ? "登录失败，请稍后重试" : "注册失败，请稍后重试"
                        );
                    });
            },
            logout() {
                clearSession();
                app.user = null;
                app.form.content = "";
                app.getData();
            }
        },
        created() {
            this.getData();
            // 用本地 token 换一次身份，确认它还没过期
            if (this.user) {
                axios.get("/auth/me").catch(() => {
                    clearSession();
                    app.user = null;
                });
            }
        }
    });
};
