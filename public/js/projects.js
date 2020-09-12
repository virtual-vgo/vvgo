$(document).ready(function () {
    $('#projects').DataTable({
        paging: false,
        order: [[1, 'asc'], [0, 'asc']],
        columnDefs: [
            {
                targets: [0],
                visible: false,
            },
            {
                targets: [1],
                visible: false,
            },
            {
                targets: [2, 3],
                orderable: false,
            },
            {
                className: "text-left",
                targets: "_all"
            }
        ]
    });
});
