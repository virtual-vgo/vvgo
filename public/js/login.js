$(document).ready(function () {
    $('#signinForm').submit(function (event) {
        const user = $('#inputUser').val();
        const pass = $('#inputPassword').val();
        $.post("/login/password", {"user": user, "pass": pass}, function () {
            document.location.href = "/login/success"
        }).fail(function () {
            $('#invalidAuth').removeClass('d-none')
        })
        event.preventDefault()
    })
})
