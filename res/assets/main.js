window.aa_build_storage_user = null;
window.apiFrame = null;

function loginWithDiscordStage2() {
    popupCenterScreen('/login?in_popup', 'Login', 800, 800, false);
}

function logoutOfDiscord() {
    fetch("/logout", {method: "POST"}).then(function (resp) {
        if (resp.ok) {
            window.aa_build_storage_user = null;
            check_user();
        }
    });
}

function triggerUploadOnServer(event) {
    let btn = event.target;
    let p = location.pathname;
    let data = new FormData();
    data.set("file", btn.parentElement.parentElement.getAttribute("data-card-name"));
    btn.classList.add("is-publishing");
    fetch(p + "/publish", {
        method: "POST",
        body: data,
    }).then(function (resp) {
        if (resp.ok) {
            btn.classList.remove("is-publishing");
            btn.classList.add("is-published");
        } else {
            btn.classList.remove("is-publishing");
            btn.classList.add("is-error");
            resp.text().then(function (text) {
                alert("Error message: " + text);
            });
        }
    });
}

window.addEventListener("load", function () {
    check_id_domain();
    check_user();
});

function reloadPublishClickEvents() {
    document.querySelectorAll(".card.card-can-upload .publish-btn").forEach(function (el) {
        el.removeEventListener("click", triggerUploadOnServer);
        el.addEventListener("click", triggerUploadOnServer);
    });
}

function check_id_domain() {
    const f = document.createElement("iframe");
    f.src = "/check";
    f.style.display = "none";
    document.body.appendChild(f);
    window.apiFrame = f;
}

function check_user() {
    let is_logged_in = window.aa_build_storage_user !== null;
    reloadPublishClickEvents();
    showOrHideWithBool("loginBtn", !is_logged_in);
    showOrHideWithBool("loginMenu", is_logged_in);
    showOrHideBoolFunc(is_logged_in, function () {
        document.body.classList.add("logged-in");
    }, function () {
        document.body.classList.remove("logged-in");
    });
    showOrHideBoolFunc(is_logged_in && window.aa_build_storage_user.admin, function () {
        document.body.classList.add("can-publish");
    }, function () {
        document.body.classList.remove("can-publish");
    })

    if (window.aa_build_storage_user !== null) {
        document.getElementById("loginMenuName").textContent = window.aa_build_storage_user.name;
        document.getElementById("loginMenuAvatar").src = window.aa_build_storage_user.picture;
    } else {
        document.getElementById("loginMenuName").textContent = "Wumpus";
        document.getElementById("loginMenuAvatar").src = "";
    }
}

function showOrHideWithBool(id, v) {
    const el = document.getElementById(id);
    showOrHideBoolFunc(v, function () {
        el.style.display = "inline-block";
    }, function () {
        el.style.display = "none";
    });
}

function showOrHideBoolFunc(v, f1, f2) {
    if (v) f1();
    else f2();
}

window.onmessage = function (event) {
    if (event.origin !== location.origin) return;
    if (isObject(event.data)) {
        console.log(event.data);
        if (isObject(event.data.user)) {
            let d = Object.assign({sub: null, login: null, name: null, picture: null, admin: false}, event.data.user);
            if (d.sub === null || d.login === null || d.name === null || d.picture === null) {
                alert("Failed to log user in: the login data is structured correctly but probably corrupted");
                return;
            }
            window.aa_build_storage_user = d;
            check_user();
            return;
        }
    }
    alert("Failed to log user in: the login data was probably corrupted");
}


function isObject(obj) {
    return obj != null && obj.constructor.name === "Object"
}


function popupCenterScreen(url, title, w, h, focus) {
    const top = (screen.availHeight - h) / 4, left = (screen.availWidth - w) / 2;
    const popup = openWindow(url, title, `scrollbars=yes,width=${w},height=${h},top=${top},left=${left}`);
    if (focus === true && window.focus) popup.focus();
    return popup;
}

function openWindow(url, winnm, options) {
    const wTop = firstAvailableValue([window.screen.availTop, window.screenY, window.screenTop, 0]);
    const wLeft = firstAvailableValue([window.screen.availLeft, window.screenX, window.screenLeft, 0]);
    let top = 0, left = 0;
    let result;
    if ((result = /top=(\d+)/g.exec(options))) top = parseInt(result[1]);
    if ((result = /left=(\d+)/g.exec(options))) left = parseInt(result[1]);
    if (options) {
        options = options.replace("top=" + top, "top=" + (parseInt(top) + wTop));
        options = options.replace("left=" + left, "left=" + (parseInt(left) + wLeft));
        w = window.open(url, winnm, options);
    } else w = window.open(url, winnm);
    return w;
}

function firstAvailableValue(arr) {
    for (let i = 0; i < arr.length; i++)
        if (typeof arr[i] != 'undefined')
            return arr[i];
}
