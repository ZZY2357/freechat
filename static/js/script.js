let app;

// 凭证放在 HttpOnly Cookie 里，JS 读不到，也不再需要 localStorage。
// 因此登录态只能在启动时问服务端要，不能同步读出来。
axios.defaults.withCredentials = true;

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

// Cookie 过期或失效时清掉内存里的登录态并弹出登录框
// （登录/注册接口自身的 401 交给表单处理）
axios.interceptors.response.use(
    (res) => res,
    (error) => {
        const status = error.response && error.response.status;
        const url = (error.config && error.config.url) || "";
        if (status === 401 && app && url.indexOf("/auth/") !== 0) {
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
            user: null,
            // 登录态要等 /auth/me 回来才知道，先用它压住头部避免闪一下“未登录”
            authReady: false,
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
            // Cookie 是 HttpOnly 的，前端无法判断自己是否已登录，只能问服务端
            refreshSession() {
                return axios
                    .get("/auth/me")
                    .then((res) => {
                        app.user = { username: res.data.username };
                    })
                    .catch(() => {
                        app.user = null;
                    })
                    .then(() => {
                        app.authReady = true;
                    });
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

                // 凭证由服务端通过 Set-Cookie 下发，响应体里没有 token
                axios
                    .post(isLogin ? "/auth/login" : "/auth/register", {
                        username: username,
                        password: password
                    })
                    .then((res) => {
                        app.user = { username: res.data.username };
                        app.authReady = true;
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
                // HttpOnly Cookie 前端删不掉，必须请服务端清除
                axios
                    .post("/auth/logout")
                    .catch(() => {})
                    .then(() => {
                        app.user = null;
                        app.form.content = "";
                        app.getData();
                    });
            }
        },
        created() {
            this.getData();
            this.refreshSession();
        }
    });
};
