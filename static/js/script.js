let app;

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

window.onload = () => {
    app = new Vue({
        el: "#app",
        data: {
            comments: [],
            os: getOS(),
            form: {
                name: "",
                content: ""
            }
        },
        methods: {
            getData() {
                axios.get("/comments").then((res) => (app.comments = res.data));
            },
            send() {
                if (
                    app.form.name &&
                    app.form.content &&
                    app.form.name.length <= 10 &&
                    app.form.content.length <= 50
                ) {
                    axios
                        .post("/comments", {
                            name: app.form.name,
                            content: app.form.content,
                            os: app.os
                        })
                        .then((res) => {
                            app.comments = res.data;
                            app.form.content = "";
                        });
                }
            }
        },
        created() {
            this.getData();
        }
    });
};
