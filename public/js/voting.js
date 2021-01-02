$(document).ready(function () {
    let sortable = $("#sortable")
    sortable.sortable();
    sortable.disableSelection();

    $("vote").map(() => $(this).style())
})

function submitVote() {
    let votes = $("vote").map(function() {
        return $(this).text();
    }).get();

    let xhr = new XMLHttpRequest();
    let url = "voting/submit"
    xhr.open("POST", url, true);
    xhr.setRequestHeader("Content-Type", "application/json");
    xhr.send(JSON.stringify(votes));
    console.log("vote submitted", votes);
}
