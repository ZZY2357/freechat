let app;

window.onload = () => {
    app = new Vue({
        el: "#app",
        data: {
            comments: [],
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
                if (app.form.name && app.form.content) {
                    axios
                        .post("/comments", app.form)
                        .then((res) => (app.comments = res.data));
                }
            }
        },
        created() {
            this.getData();
        }
    });
};
